package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"autoscolombia/parqueadero/internal/adapters/httpapi"
	"autoscolombia/parqueadero/internal/adapters/postgres"
	"autoscolombia/parqueadero/internal/adapters/seguridad"
	"autoscolombia/parqueadero/internal/application"
	"autoscolombia/parqueadero/internal/config"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := ejecutar(log); err != nil {
		log.Error("la aplicación terminó con error", "error", err)
		os.Exit(1)
	}
}

func ejecutar(log *slog.Logger) error {
	cfg, err := config.Cargar()
	if err != nil {
		return err
	}
	if cfg.APIToken == "" {
		log.Warn("ejecutando SIN autenticación (ALLOW_INSECURE_NO_AUTH): usar solo en desarrollo local")
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		return errors.Join(errors.New("no se pudo conectar a la base de datos"), err)
	}

	repo := postgres.Nuevo(db)
	if err := repo.EsquemaActualizado(context.Background()); err != nil {
		return err
	}
	svc := application.NuevoServicio(repo, time.Now).ConHasheador(seguridad.NuevoPBKDF2(cfg.HashIteraciones))
	if cfg.APIToken != "" {
		if err := prepararAdministrador(log, svc, cfg); err != nil {
			return err
		}
	}
	api := httpapi.Nueva(svc, cfg.APIToken, log)

	raiz := http.NewServeMux()
	raiz.Handle("/api/", api.Rutas())
	raiz.Handle("/healthz", api.Rutas())
	if st, err := os.Stat(cfg.StaticDir); err == nil && st.IsDir() {
		raiz.Handle("/", httpapi.Estaticos(cfg.StaticDir))
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpapi.Envolver(raiz, log),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 16,
	}

	ctx, parar := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer parar()
	errCh := make(chan error, 1)
	go func() {
		log.Info("API escuchando", "addr", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		log.Info("apagando de forma ordenada")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
	return nil
}

// prepararAdministrador crea el primer administrador con ADMIN_USUARIO / ADMIN_CLAVE cuando la base aún no tiene usuarios.
func prepararAdministrador(log *slog.Logger, svc *application.Servicio, cfg config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if cfg.AdminClave == "" {
		if n, err := svc.Usuarios(ctx); err == nil && len(n) == 0 {
			log.Warn("no hay usuarios: defina ADMIN_USUARIO y ADMIN_CLAVE para crear el primer administrador")
		}
		return nil
	}
	creado, err := svc.AsegurarAdministrador(ctx, cfg.AdminUsuario, cfg.AdminNombre, cfg.AdminClave)
	if err != nil {
		return errors.Join(errors.New("no se pudo crear el administrador inicial"), err)
	}
	if creado {
		log.Info("administrador inicial creado", "usuario", cfg.AdminUsuario)
	}
	return nil
}
