package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"autoscolombia/parqueadero/internal/adapters/httpapi"
	"autoscolombia/parqueadero/internal/adapters/postgres"
	"autoscolombia/parqueadero/internal/application"
)

func env(clave, defecto string) string {
	if v := os.Getenv(clave); v != "" {
		return v
	}
	return defecto
}

func main() {
	db, err := sql.Open("postgres", env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/parqueadero?sslmode=disable"))
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("no se pudo conectar a la base de datos: %v", err)
	}
	svc := application.NuevoServicio(postgres.Nuevo(db), time.Now)
	api := httpapi.Nueva(svc, os.Getenv("API_TOKEN"))

	raiz := http.NewServeMux()
	raiz.Handle("/api/", api.Rutas())
	raiz.Handle("/healthz", api.Rutas())
	// Sirve el frontend compilado (React) si existe.
	if dir := env("STATIC_DIR", "./static"); dirExiste(dir) {
		raiz.Handle("/", http.FileServer(http.Dir(dir)))
	}

	addr := ":" + env("PORT", "8080")
	log.Printf("API escuchando en %s", addr)
	log.Fatal(http.ListenAndServe(addr, raiz))
}

func dirExiste(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}
