package postgres_test

// Pruebas de integración contra PostgreSQL real.
// Se omiten si no existe TEST_DATABASE_URL (apuntar a una base DESECHABLE: se recrea el esquema).

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"autoscolombia/parqueadero/internal/adapters/postgres"
	"autoscolombia/parqueadero/internal/application"
	"autoscolombia/parqueadero/internal/domain"
)

var ctx = context.Background()

func baseDeDatos(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL no definida: se omiten las pruebas de integración")
	}
	db, err := sql.Open("postgres", url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	esquema, err := os.ReadFile("../../../db/schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DROP SCHEMA public CASCADE; CREATE SCHEMA public;`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(esquema)); err != nil {
		t.Fatalf("no se pudo aplicar schema.sql: %v", err)
	}
	return db
}

func sembrar(t *testing.T, db *sql.DB, vehiculos int) {
	t.Helper()
	mustExec := func(q string, args ...any) {
		if _, err := db.Exec(q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	mustExec(`INSERT INTO cliente (documento, nombre) VALUES ('1','Carlos Gómez'), ('2','Marta Rojas')`)
	mustExec(`INSERT INTO mensualidad (id_cliente, fecha_inicio, fecha_fin) VALUES (1,'2026-10-01','2026-10-31'), (2,'2026-08-01','2026-08-31')`)
	for i := 0; i < vehiculos; i++ {
		mustExec(`INSERT INTO vehiculo (placa, tipo, id_cliente) VALUES ($1,'CARRO',1)`, fmt.Sprintf("AAA%03d", i))
	}
	mustExec(`INSERT INTO vehiculo (placa, tipo, id_cliente) VALUES ('DEF456','CARRO',2), ('XYZ98A','MOTO',1)`)
	mustExec(`INSERT INTO celda (codigo, tipo) VALUES ('A-01','CARRO'), ('A-02','CARRO'), ('M-01','MOTO')`)
}

func TestFlujoCompletoContraPostgres(t *testing.T) {
	db := baseDeDatos(t)
	sembrar(t, db, 1)
	reloj := time.Date(2026, 10, 20, 15, 0, 0, 0, time.UTC)
	s := application.NuevoServicio(postgres.Nuevo(db), func() time.Time { return reloj })

	res, err := s.RegistrarIngreso(ctx, "AAA000", 1, &application.NovedadEntrada{Tipo: domain.NovedadDanio, Descripcion: "Rayón"})
	if err != nil || res.Celda != "A-01" || res.Registro.ID == 0 {
		t.Fatalf("ingreso: %v %+v", err, res)
	}
	if _, err := s.RegistrarIngreso(ctx, "AAA000", 2, nil); !errors.Is(err, domain.ErrVehiculoYaDentro) {
		t.Fatalf("segundo ingreso del mismo vehículo: %v", err)
	}
	if _, err := s.RegistrarIngreso(ctx, "DEF456", 2, nil); !errors.Is(err, domain.ErrMensualidadVencida) {
		t.Fatalf("mensualidad vencida: %v", err)
	}
	if _, err := s.RegistrarIngreso(ctx, "XYZ98A", 3, &application.NovedadEntrada{Tipo: domain.NovedadOtra, Descripcion: ""}); !errors.Is(err, domain.ErrNovedadInvalida) {
		t.Fatalf("novedad inválida: %v", err)
	}
	if _, err := s.RegistrarNovedad(ctx, "AAA000", application.NovedadEntrada{Tipo: domain.NovedadOtra, Descripcion: "Alarma"}); err != nil {
		t.Fatalf("novedad: %v", err)
	}

	dentro, err := s.VehiculosDentro(ctx)
	if err != nil || len(dentro) != 1 || dentro[0].Placa != "AAA000" || dentro[0].Cliente != "Carlos Gómez" || dentro[0].Celda != "A-01" {
		t.Fatalf("vehículos dentro: %v %+v", err, dentro)
	}
	libres, _ := s.ConsultarCeldas(ctx, "DISPONIBLE")
	if len(libres) != 2 {
		t.Fatalf("esperaba 2 celdas libres, hubo %d", len(libres))
	}

	reloj = reloj.Add(2*time.Hour + 5*time.Minute)
	sal, err := s.RegistrarSalida(ctx, "AAA000", &application.NovedadEntrada{Tipo: domain.NovedadFaltante, Descripcion: "Falta espejo"})
	if err != nil || sal.Minutos != 125 {
		t.Fatalf("salida: %v %+v", err, sal)
	}
	if libres, _ := s.ConsultarCeldas(ctx, "DISPONIBLE"); len(libres) != 3 {
		t.Fatalf("la celda debía quedar libre; libres=%d", len(libres))
	}
	var n int
	_ = db.QueryRow(`SELECT count(*) FROM novedad`).Scan(&n)
	if n != 3 { // entrada(daño) + estancia(alarma) + salida(faltante)
		t.Fatalf("esperaba 3 novedades guardadas, hubo %d", n)
	}
	if _, err := s.RegistrarSalida(ctx, "AAA000", nil); !errors.Is(err, domain.ErrSinIngresoActivo) {
		t.Fatalf("segunda salida: %v", err)
	}
}

// Día calendario de Colombia: el 31-oct a las 20:30 (01:30 UTC del 1-nov) la mensualidad que vence el 31-oct sigue vigente.
func TestMensualidadVigenteEnElBordeDelDia(t *testing.T) {
	db := baseDeDatos(t)
	sembrar(t, db, 0)
	repo := postgres.Nuevo(db)
	instante := time.Date(2026, 11, 1, 1, 30, 0, 0, time.UTC)
	if ok, err := repo.MensualidadVigente(ctx, 1, instante.In(application.ZonaColombia)); err != nil || !ok {
		t.Fatalf("último día debía estar vigente: %v %v", ok, err)
	}
	if ok, _ := repo.MensualidadVigente(ctx, 1, instante.Add(5*time.Hour).In(application.ZonaColombia)); ok {
		t.Fatal("el 1-nov en Colombia debía estar vencida")
	}
}

// Carrera real: muchos encargados intentan ocupar la MISMA celda; la base de datos garantiza un solo ganador.
func TestIngresoConcurrente_UnaSolaCeldaEnPostgres(t *testing.T) {
	db := baseDeDatos(t)
	const vehiculos = 12
	sembrar(t, db, vehiculos)
	s := application.NuevoServicio(postgres.Nuevo(db), func() time.Time { return time.Date(2026, 10, 20, 15, 0, 0, 0, time.UTC) })

	var wg sync.WaitGroup
	var ok, noDisponible atomic.Int32
	for i := 0; i < vehiculos; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := s.RegistrarIngreso(ctx, fmt.Sprintf("AAA%03d", i), 1, nil)
			switch {
			case err == nil:
				ok.Add(1)
			case errors.Is(err, domain.ErrCeldaNoDisponible):
				noDisponible.Add(1)
			default:
				t.Errorf("error inesperado: %v", err)
			}
		}(i)
	}
	wg.Wait()
	if ok.Load() != 1 || noDisponible.Load() != vehiculos-1 {
		t.Fatalf("esperaba 1 éxito y %d rechazos; hubo %d y %d", vehiculos-1, ok.Load(), noDisponible.Load())
	}
	var activos int
	_ = db.QueryRow(`SELECT count(*) FROM registro WHERE fecha_salida IS NULL AND id_celda=1`).Scan(&activos)
	if activos != 1 {
		t.Fatalf("la celda debía tener un único registro activo, tiene %d", activos)
	}
}

// Las restricciones de la base protegen aunque se salte la capa de aplicación.
func TestRestriccionesDeLaBase(t *testing.T) {
	db := baseDeDatos(t)
	sembrar(t, db, 2)
	exec := func(q string) error { _, err := db.Exec(q); return err }
	if err := exec(`INSERT INTO registro (id_vehiculo,id_celda,fecha_entrada) VALUES (1,1,now())`); err != nil {
		t.Fatal(err)
	}
	if exec(`INSERT INTO registro (id_vehiculo,id_celda,fecha_entrada) VALUES (1,2,now())`) == nil {
		t.Error("un vehículo no debe tener dos registros activos")
	}
	if exec(`INSERT INTO registro (id_vehiculo,id_celda,fecha_entrada) VALUES (2,1,now())`) == nil {
		t.Error("una celda no debe tener dos registros activos")
	}
	if exec(`UPDATE celda SET estado='ROTA' WHERE id_celda=2`) == nil {
		t.Error("el estado de celda debe respetar el CHECK")
	}
	if exec(`INSERT INTO registro (id_vehiculo,id_celda,fecha_entrada,fecha_salida) VALUES (2,2,now(),now() - interval '1 hour')`) == nil {
		t.Error("la salida no puede ser anterior a la entrada")
	}
}

// ---- RF-08: registrar vehículo contra PostgreSQL real ----

// CP-V10 · RF-08 · PostgreSQL: registro con cliente nuevo y existente; el vehículo registrado ingresa.
func TestRegistrarVehiculoContraPostgres(t *testing.T) {
	db := baseDeDatos(t)
	sembrar(t, db, 0)
	reloj := time.Date(2026, 10, 20, 15, 0, 0, 0, time.UTC)
	s := application.NuevoServicio(postgres.Nuevo(db), func() time.Time { return reloj })
	cli := application.ClienteEntrada{Documento: "5005005005", Nombre: "Ana Pérez", Telefono: "3004445566"}
	men := &application.MensualidadEntrada{FechaInicio: "2026-10-01", FechaFin: "2026-10-31"}

	res, err := s.RegistrarVehiculo(ctx, "jkl-456", domain.Carro, cli, men)
	if err != nil || !res.ClienteNuevo || !res.MensualidadRegistrada || res.Vehiculo.ID == 0 || res.Cliente != "Ana Pérez" {
		t.Fatalf("registro: %v %+v", err, res)
	}
	// Segundo vehículo con el mismo documento: reutiliza al cliente (sin duplicarlo) y conserva el nombre.
	res2, err := s.RegistrarVehiculo(ctx, "MNO12P", domain.Moto, application.ClienteEntrada{Documento: "5005005005", Nombre: "Otro nombre"}, nil)
	if err != nil || res2.ClienteNuevo || res2.Cliente != "Ana Pérez" || res2.Vehiculo.ClienteID != res.Vehiculo.ClienteID {
		t.Fatalf("segundo vehículo: %v %+v", err, res2)
	}
	var clientes, mensualidades int
	_ = db.QueryRow(`SELECT count(*) FROM cliente WHERE documento='5005005005'`).Scan(&clientes)
	_ = db.QueryRow(`SELECT count(*) FROM mensualidad WHERE id_cliente=$1`, res.Vehiculo.ClienteID).Scan(&mensualidades)
	if clientes != 1 || mensualidades != 1 {
		t.Fatalf("esperaba 1 cliente y 1 mensualidad, hubo %d y %d", clientes, mensualidades)
	}
	// Los vehículos registrados pueden ingresar (la mensualidad quedó guardada).
	if _, err := s.RegistrarIngreso(ctx, "JKL456", 1, nil); err != nil {
		t.Fatalf("ingreso del vehículo registrado: %v", err)
	}
	if _, err := s.RegistrarIngreso(ctx, "MNO12P", 3, nil); err != nil {
		t.Fatalf("ingreso de la moto: %v", err)
	}
	// Placa repetida.
	if _, err := s.RegistrarVehiculo(ctx, "JKL456", domain.Carro, cli, nil); !errors.Is(err, domain.ErrPlacaYaRegistrada) {
		t.Fatalf("placa repetida: %v", err)
	}
}

// CP-V11 · RF-08 · PostgreSQL: si falla el vehículo, la transacción deshace también al cliente nuevo.
func TestRegistrarVehiculo_PlacaRepetidaDeshaceClienteNuevo(t *testing.T) {
	db := baseDeDatos(t)
	sembrar(t, db, 0)
	repo := postgres.Nuevo(db)
	c := &domain.Cliente{Documento: "7007007007", Nombre: "Nuevo Cliente"}
	v := &domain.Vehiculo{Placa: "DEF456", Tipo: domain.Carro} // ya existe en la siembra
	m, _ := domain.NuevaMensualidad("2026-10-01", "2026-10-31")
	if _, err := repo.GuardarVehiculo(ctx, c, v, m); !errors.Is(err, domain.ErrPlacaYaRegistrada) {
		t.Fatalf("esperaba ErrPlacaYaRegistrada, obtuve %v", err)
	}
	var n int
	_ = db.QueryRow(`SELECT count(*) FROM cliente WHERE documento='7007007007'`).Scan(&n)
	if n != 0 {
		t.Fatal("el cliente nuevo debía deshacerse junto con el vehículo")
	}
}

// CP-V12 · RF-08, RNF-11 · PostgreSQL: misma placa desde 12 conexiones a la vez; la restricción UNIQUE deja pasar una.
func TestRegistrarVehiculo_MismaPlacaSimultanea(t *testing.T) {
	db := baseDeDatos(t)
	sembrar(t, db, 0)
	db.SetMaxOpenConns(12)
	s := application.NuevoServicio(postgres.Nuevo(db), nil)
	var ok, dup int32
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cli := application.ClienteEntrada{Documento: fmt.Sprintf("90%d", i%3), Nombre: "Cliente"} // documentos repetidos entre hilos
			_, err := s.RegistrarVehiculo(ctx, "JKL456", domain.Carro, cli, nil)
			switch {
			case err == nil:
				atomic.AddInt32(&ok, 1)
			case errors.Is(err, domain.ErrPlacaYaRegistrada):
				atomic.AddInt32(&dup, 1)
			default:
				t.Errorf("error inesperado: %v", err)
			}
		}(i)
	}
	wg.Wait()
	var veh int
	_ = db.QueryRow(`SELECT count(*) FROM vehiculo WHERE placa='JKL456'`).Scan(&veh)
	if ok != 1 || dup != 11 || veh != 1 {
		t.Fatalf("esperaba 1 registro y 11 rechazos, hubo %d, %d (filas: %d)", ok, dup, veh)
	}
}
