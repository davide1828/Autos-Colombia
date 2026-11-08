// Paquete httpapi: adaptador de presentación (API REST JSON).
package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"autoscolombia/parqueadero/internal/application"
	"autoscolombia/parqueadero/internal/domain"
)

type API struct {
	svc   *application.Servicio
	firma *firmador // nil: la API no exige autenticación (solo desarrollo)
	log   *slog.Logger
	lim   *limitador
}

// Nueva crea la API. secreto firma las sesiones; si está vacío no se exige autenticación (solo desarrollo local).
func Nueva(svc *application.Servicio, secreto string, log *slog.Logger) *API {
	if log == nil {
		log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	a := &API{svc: svc, log: log, lim: nuevoLimitador(10, time.Minute)}
	if secreto != "" {
		a.firma = &firmador{clave: []byte(secreto), ttl: DuracionSesion, ahora: time.Now}
	}
	return a
}

// ConReloj permite fijar el reloj de las sesiones (pruebas de vencimiento).
func (a *API) ConReloj(ahora func() time.Time) *API {
	if a.firma != nil {
		a.firma.ahora = ahora
	}
	return a
}

var (
	ambos     = []domain.Rol{domain.RolAdministrador, domain.RolEncargado}
	soloAdmin = []domain.Rol{domain.RolAdministrador}
)

func (a *API) Rutas() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { escribir(w, 200, map[string]string{"estado": "ok"}) })
	mux.HandleFunc("POST /api/login", a.login)
	mux.HandleFunc("GET /api/sesion", a.requiere(ambos, a.sesion))
	mux.HandleFunc("POST /api/sesion/clave", a.requiere(ambos, a.cambiarMiClave))

	// Operación diaria (ambos roles)
	mux.HandleFunc("POST /api/ingresos", a.requiere(ambos, a.ingreso))
	mux.HandleFunc("POST /api/salidas", a.requiere(ambos, a.salida))
	mux.HandleFunc("POST /api/novedades", a.requiere(ambos, a.novedad))
	mux.HandleFunc("GET /api/celdas", a.requiere(ambos, a.celdas))
	mux.HandleFunc("GET /api/registros/activos", a.requiere(ambos, a.activos))

	// Clientes y vehículos (ambos roles)
	mux.HandleFunc("POST /api/vehiculos", a.requiere(ambos, a.vehiculo))
	mux.HandleFunc("PATCH /api/vehiculos/{id}", a.requiere(ambos, a.editarVehiculo))
	mux.HandleFunc("GET /api/clientes", a.requiere(ambos, a.buscarClientes))
	mux.HandleFunc("GET /api/clientes/{id}", a.requiere(ambos, a.cliente))
	mux.HandleFunc("PATCH /api/clientes/{id}", a.requiere(ambos, a.editarCliente))
	mux.HandleFunc("POST /api/clientes/{id}/mensualidades", a.requiere(ambos, a.agregarMensualidad))

	// Administración (solo administrador)
	mux.HandleFunc("POST /api/celdas", a.requiere(soloAdmin, a.crearCelda))
	mux.HandleFunc("PATCH /api/celdas/{id}", a.requiere(soloAdmin, a.editarCelda))
	mux.HandleFunc("PUT /api/celdas/{id}/estado", a.requiere(soloAdmin, a.estadoCelda))
	mux.HandleFunc("GET /api/usuarios", a.requiere(soloAdmin, a.usuarios))
	mux.HandleFunc("POST /api/usuarios", a.requiere(soloAdmin, a.crearUsuario))
	mux.HandleFunc("PATCH /api/usuarios/{id}", a.requiere(soloAdmin, a.editarUsuario))
	mux.HandleFunc("POST /api/usuarios/{id}/clave", a.requiere(soloAdmin, a.restablecerClave))
	mux.HandleFunc("GET /api/auditoria", a.requiere(soloAdmin, a.auditoria))

	// Comodín: distingue "la ruta no existe" (404) de "método no permitido" (405) con respuestas JSON.
	permitidos := []struct {
		ruta    *regexp.Regexp
		metodos string
	}{
		{regexp.MustCompile(`^/api/(ingresos|salidas|novedades|login|sesion/clave)$`), "POST"},
		{regexp.MustCompile(`^/api/(registros/activos|sesion|auditoria)$`), "GET"},
		{regexp.MustCompile(`^/api/(celdas|vehiculos|clientes|usuarios)$`), "GET, POST"},
		{regexp.MustCompile(`^/api/(vehiculos|clientes|usuarios|celdas)/\d+$`), "GET, PATCH"},
		{regexp.MustCompile(`^/api/clientes/\d+/mensualidades$`), "POST"},
		{regexp.MustCompile(`^/api/usuarios/\d+/clave$`), "POST"},
		{regexp.MustCompile(`^/api/celdas/\d+/estado$`), "PUT"},
	}
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		for _, p := range permitidos {
			if p.ruta.MatchString(r.URL.Path) {
				w.Header().Set("Allow", p.metodos)
				escribirError(w, http.StatusMethodNotAllowed, "metodo_no_permitido", "use el método "+p.metodos)
				return
			}
		}
		escribirError(w, http.StatusNotFound, "no_encontrado", "la ruta no existe")
	})
	return mux
}

// auth valida la sesión y deja al actor en el contexto. Con la sesión vigente consulta al usuario para que
// desactivarlo o cambiarle el rol surta efecto de inmediato (no hay que esperar a que venza el token).
func (a *API) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if a.firma == nil { // desarrollo sin autenticación
			next(w, r.WithContext(application.ConActor(r.Context(), application.Actor{Usuario: "desarrollo", Rol: domain.RolAdministrador})))
			return
		}
		ip := ipDe(r)
		if a.lim.bloqueado(ip) {
			w.Header().Set("Retry-After", "60")
			escribirError(w, http.StatusTooManyRequests, "demasiados_intentos", "demasiados intentos fallidos; espere un minuto")
			return
		}
		actor, err := a.firma.verificar(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if err == nil {
			var u *domain.Usuario
			if u, err = a.svc.Usuario(r.Context(), actor.ID); err == nil && u.Activo {
				actor.Usuario, actor.Rol = u.Usuario, u.Rol
			} else {
				err = errTokenInvalido
			}
		}
		if err != nil {
			a.lim.falla(ip)
			a.log.Warn("autenticación fallida", "ip", ip)
			escribirError(w, 401, "no_autorizado", "sesión inválida o vencida; inicie sesión de nuevo")
			return
		}
		next(w, r.WithContext(application.ConActor(r.Context(), actor)))
	}
}

// requiere exige sesión y que el rol del actor esté entre los permitidos.
func (a *API) requiere(roles []domain.Rol, next http.HandlerFunc) http.HandlerFunc {
	return a.auth(func(w http.ResponseWriter, r *http.Request) {
		actor, _ := application.ActorDe(r.Context())
		for _, rol := range roles {
			if actor.Rol == rol {
				next(w, r)
				return
			}
		}
		a.log.Warn("operación sin permiso", "usuario", actor.Usuario, "rol", actor.Rol, "ruta", r.URL.Path)
		escribirError(w, http.StatusForbidden, "sin_permiso", domain.ErrSinPermiso.Error())
	})
}

type ingresoReq struct {
	Placa   string                      `json:"placa"`
	CeldaID int                         `json:"id_celda"`
	Novedad *application.NovedadEntrada `json:"novedad,omitempty"`
}

func (a *API) ingreso(w http.ResponseWriter, r *http.Request) {
	var in ingresoReq
	if !leer(w, r, &in) {
		return
	}
	res, err := a.svc.RegistrarIngreso(r.Context(), in.Placa, in.CeldaID, in.Novedad)
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 201, map[string]any{
		"id_registro": res.Registro.ID, "placa": res.Placa, "celda": res.Celda, "fecha_entrada": res.Registro.Entrada,
	})
}

type salidaReq struct {
	Placa   string                      `json:"placa"`
	Novedad *application.NovedadEntrada `json:"novedad,omitempty"`
}

func (a *API) salida(w http.ResponseWriter, r *http.Request) {
	var in salidaReq
	if !leer(w, r, &in) {
		return
	}
	res, err := a.svc.RegistrarSalida(r.Context(), in.Placa, in.Novedad)
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 200, map[string]any{
		"id_registro": res.Registro.ID, "placa": res.Placa, "celda": res.Celda,
		"fecha_entrada": res.Registro.Entrada, "fecha_salida": res.Registro.Salida, "minutos": res.Minutos,
	})
}

type novedadReq struct {
	Placa       string             `json:"placa"`
	Tipo        domain.TipoNovedad `json:"tipo"`
	Descripcion string             `json:"descripcion"`
}

func (a *API) novedad(w http.ResponseWriter, r *http.Request) {
	var in novedadReq
	if !leer(w, r, &in) {
		return
	}
	n, err := a.svc.RegistrarNovedad(r.Context(), in.Placa, application.NovedadEntrada{Tipo: in.Tipo, Descripcion: in.Descripcion})
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 201, n)
}

type vehiculoReq struct {
	Placa       string                          `json:"placa"`
	Tipo        domain.TipoVehiculo             `json:"tipo"`
	Cliente     application.ClienteEntrada      `json:"cliente"`
	Mensualidad *application.MensualidadEntrada `json:"mensualidad,omitempty"`
}

func (a *API) vehiculo(w http.ResponseWriter, r *http.Request) {
	var in vehiculoReq
	if !leer(w, r, &in) {
		return
	}
	res, err := a.svc.RegistrarVehiculo(r.Context(), in.Placa, in.Tipo, in.Cliente, in.Mensualidad)
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 201, map[string]any{
		"id_vehiculo": res.Vehiculo.ID, "placa": res.Vehiculo.Placa, "tipo": res.Vehiculo.Tipo,
		"id_cliente": res.Vehiculo.ClienteID, "cliente": res.Cliente, "cliente_nuevo": res.ClienteNuevo,
		"mensualidad_registrada": res.MensualidadRegistrada,
	})
}

func (a *API) celdas(w http.ResponseWriter, r *http.Request) {
	estado := strings.ToUpper(r.URL.Query().Get("estado"))
	if estado != "" && estado != "DISPONIBLE" && estado != "OCUPADA" && estado != "INACTIVA" {
		escribirError(w, 400, "solicitud_invalida", "estado debe ser DISPONIBLE, OCUPADA o INACTIVA")
		return
	}
	cs, err := a.svc.ConsultarCeldas(r.Context(), estado)
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 200, cs)
}

func (a *API) activos(w http.ResponseWriter, r *http.Request) {
	ds, err := a.svc.VehiculosDentro(r.Context())
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 200, ds)
}

// --- utilidades ---

func leer(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<16)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil || dec.More() {
		escribirError(w, 400, "solicitud_invalida", "el cuerpo no es un JSON válido")
		return false
	}
	return true
}

func escribir(w http.ResponseWriter, codigo int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codigo)
	_ = json.NewEncoder(w).Encode(v)
}

func escribirError(w http.ResponseWriter, codigo int, clave, mensaje string) {
	escribir(w, codigo, map[string]string{"error": clave, "mensaje": mensaje})
}

// responderError traduce errores de negocio a códigos HTTP.
func (a *API) responderError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrPlacaInvalida), errors.Is(err, domain.ErrNovedadInvalida), errors.Is(err, domain.ErrTipoVehiculoInvalido),
		errors.Is(err, domain.ErrClienteInvalido), errors.Is(err, domain.ErrMensualidadInvalida), errors.Is(err, domain.ErrUsuarioInvalido),
		errors.Is(err, domain.ErrClaveDebil), errors.Is(err, domain.ErrRolInvalido), errors.Is(err, domain.ErrCeldaInvalida),
		errors.Is(err, domain.ErrEstadoCeldaInvalido), errors.Is(err, domain.ErrClaveActualIncorrecta):
		escribirError(w, 400, "datos_invalidos", err.Error())
	case errors.Is(err, domain.ErrCredenciales):
		escribirError(w, 401, "no_autorizado", err.Error())
	case errors.Is(err, domain.ErrSinPermiso):
		escribirError(w, 403, "sin_permiso", err.Error())
	case errors.Is(err, domain.ErrVehiculoNoRegistrado), errors.Is(err, domain.ErrCeldaNoExiste), errors.Is(err, domain.ErrUsuarioNoExiste),
		errors.Is(err, domain.ErrClienteNoExiste):
		escribirError(w, 404, "no_encontrado", err.Error())
	case errors.Is(err, domain.ErrVehiculoYaDentro), errors.Is(err, domain.ErrCeldaNoDisponible), errors.Is(err, domain.ErrSinIngresoActivo),
		errors.Is(err, domain.ErrPlacaYaRegistrada), errors.Is(err, domain.ErrUsuarioDuplicado), errors.Is(err, domain.ErrUltimoAdministrador),
		errors.Is(err, domain.ErrCeldaDuplicada), errors.Is(err, domain.ErrCeldaOcupada), errors.Is(err, domain.ErrVehiculoDentro),
		errors.Is(err, domain.ErrMensualidadSolapada):
		escribirError(w, 409, "conflicto", err.Error())
	case errors.Is(err, domain.ErrMensualidadVencida), errors.Is(err, domain.ErrCeldaIncompatible):
		escribirError(w, 422, "regla_de_negocio", err.Error())
	default:
		a.log.Error("error interno", "error", err)
		escribirError(w, 500, "error_interno", "ocurrió un error inesperado")
	}
}
