package httpapi

// Iteración 2: endpoints de sesión, usuarios, clientes, vehículos, celdas y auditoría.

import (
	"net/http"
	"strconv"
	"time"

	"autoscolombia/parqueadero/internal/application"
	"autoscolombia/parqueadero/internal/domain"
)

func idDe(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		escribirError(w, 400, "datos_invalidos", "el identificador de la ruta no es válido")
		return 0, false
	}
	return id, true
}

// ---------- sesión ----------

type loginReq struct {
	Usuario string `json:"usuario"`
	Clave   string `json:"clave"`
}

func (a *API) login(w http.ResponseWriter, r *http.Request) {
	if a.firma == nil {
		escribirError(w, 400, "solicitud_invalida", "la API está en modo de desarrollo sin autenticación")
		return
	}
	ip := ipDe(r)
	if a.lim.bloqueado(ip) {
		w.Header().Set("Retry-After", "60")
		escribirError(w, http.StatusTooManyRequests, "demasiados_intentos", "demasiados intentos fallidos; espere un minuto")
		return
	}
	var in loginReq
	if !leer(w, r, &in) {
		return
	}
	u, err := a.svc.Autenticar(r.Context(), in.Usuario, in.Clave)
	if err != nil {
		if err == domain.ErrCredenciales {
			a.lim.falla(ip)
			a.log.Warn("inicio de sesión fallido", "ip", ip)
		}
		a.responderError(w, err)
		return
	}
	token, exp := a.firma.emitir(application.Actor{ID: u.ID, Usuario: u.Usuario, Rol: u.Rol})
	escribir(w, 200, map[string]any{"token": token, "expira": exp, "usuario": u})
}

func (a *API) sesion(w http.ResponseWriter, r *http.Request) {
	actor, _ := application.ActorDe(r.Context())
	nombre := actor.Usuario
	if actor.ID != 0 {
		if u, err := a.svc.Usuario(r.Context(), actor.ID); err == nil {
			nombre = u.Nombre
		}
	}
	escribir(w, 200, map[string]any{"estado": "ok", "usuario": map[string]any{"id_usuario": actor.ID, "usuario": actor.Usuario, "nombre": nombre, "rol": actor.Rol}})
}

type cambiarClaveReq struct {
	Actual string `json:"actual"`
	Nueva  string `json:"nueva"`
}

func (a *API) cambiarMiClave(w http.ResponseWriter, r *http.Request) {
	var in cambiarClaveReq
	if !leer(w, r, &in) {
		return
	}
	actor, _ := application.ActorDe(r.Context())
	if err := a.svc.CambiarMiClave(r.Context(), actor.ID, in.Actual, in.Nueva); err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 200, map[string]string{"estado": "ok"})
}

// ---------- usuarios del sistema ----------

func (a *API) usuarios(w http.ResponseWriter, r *http.Request) {
	us, err := a.svc.Usuarios(r.Context())
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 200, us)
}

type usuarioReq struct {
	Usuario string     `json:"usuario"`
	Nombre  string     `json:"nombre"`
	Clave   string     `json:"clave"`
	Rol     domain.Rol `json:"rol"`
}

func (a *API) crearUsuario(w http.ResponseWriter, r *http.Request) {
	var in usuarioReq
	if !leer(w, r, &in) {
		return
	}
	u, err := a.svc.CrearUsuario(r.Context(), in.Usuario, in.Nombre, in.Clave, in.Rol)
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 201, u)
}

func (a *API) editarUsuario(w http.ResponseWriter, r *http.Request) {
	id, ok := idDe(w, r)
	if !ok {
		return
	}
	var in application.CambiosUsuario
	if !leer(w, r, &in) {
		return
	}
	u, err := a.svc.ActualizarUsuario(r.Context(), id, in)
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 200, u)
}

type nuevaClaveReq struct {
	Nueva string `json:"nueva"`
}

func (a *API) restablecerClave(w http.ResponseWriter, r *http.Request) {
	id, ok := idDe(w, r)
	if !ok {
		return
	}
	var in nuevaClaveReq
	if !leer(w, r, &in) {
		return
	}
	if err := a.svc.RestablecerClave(r.Context(), id, in.Nueva); err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 200, map[string]string{"estado": "ok"})
}

func (a *API) auditoria(w http.ResponseWriter, r *http.Request) {
	limite, _ := strconv.Atoi(r.URL.Query().Get("limite"))
	lista, err := a.svc.Auditoria(r.Context(), limite)
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 200, lista)
}

// ---------- clientes y vehículos ----------

func (a *API) buscarClientes(w http.ResponseWriter, r *http.Request) {
	lista, err := a.svc.BuscarClientes(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 200, lista)
}

func (a *API) cliente(w http.ResponseWriter, r *http.Request) {
	id, ok := idDe(w, r)
	if !ok {
		return
	}
	c, err := a.svc.Cliente(r.Context(), id)
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 200, c)
}

type clienteEdicionReq struct {
	Nombre   string `json:"nombre"`
	Telefono string `json:"telefono"`
}

func (a *API) editarCliente(w http.ResponseWriter, r *http.Request) {
	id, ok := idDe(w, r)
	if !ok {
		return
	}
	var in clienteEdicionReq
	if !leer(w, r, &in) {
		return
	}
	c, err := a.svc.ActualizarCliente(r.Context(), id, in.Nombre, in.Telefono)
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 200, c)
}

func (a *API) agregarMensualidad(w http.ResponseWriter, r *http.Request) {
	id, ok := idDe(w, r)
	if !ok {
		return
	}
	var in application.MensualidadEntrada
	if !leer(w, r, &in) {
		return
	}
	m, err := a.svc.AgregarMensualidad(r.Context(), id, in)
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 201, map[string]any{"id_mensualidad": m.ID, "id_cliente": m.ClienteID,
		"fecha_inicio": m.Inicio.Format(time.DateOnly), "fecha_fin": m.Fin.Format(time.DateOnly)})
}

type tipoReq struct {
	Tipo domain.TipoVehiculo `json:"tipo"`
}

func (a *API) editarVehiculo(w http.ResponseWriter, r *http.Request) {
	id, ok := idDe(w, r)
	if !ok {
		return
	}
	var in tipoReq
	if !leer(w, r, &in) {
		return
	}
	v, err := a.svc.ActualizarTipoVehiculo(r.Context(), id, in.Tipo)
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 200, v)
}

// ---------- celdas ----------

type celdaReq struct {
	Codigo string              `json:"codigo"`
	Tipo   domain.TipoVehiculo `json:"tipo"`
}

func (a *API) crearCelda(w http.ResponseWriter, r *http.Request) {
	var in celdaReq
	if !leer(w, r, &in) {
		return
	}
	c, err := a.svc.CrearCelda(r.Context(), in.Codigo, in.Tipo)
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 201, c)
}

type celdaEdicionReq struct {
	Codigo *string              `json:"codigo,omitempty"`
	Tipo   *domain.TipoVehiculo `json:"tipo,omitempty"`
}

func (a *API) editarCelda(w http.ResponseWriter, r *http.Request) {
	id, ok := idDe(w, r)
	if !ok {
		return
	}
	var in celdaEdicionReq
	if !leer(w, r, &in) {
		return
	}
	c, err := a.svc.ActualizarCelda(r.Context(), id, in.Codigo, in.Tipo)
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 200, c)
}

type estadoReq struct {
	Estado domain.EstadoCelda `json:"estado"`
}

func (a *API) estadoCelda(w http.ResponseWriter, r *http.Request) {
	id, ok := idDe(w, r)
	if !ok {
		return
	}
	var in estadoReq
	if !leer(w, r, &in) {
		return
	}
	c, err := a.svc.CambiarEstadoCelda(r.Context(), id, in.Estado)
	if err != nil {
		a.responderError(w, err)
		return
	}
	escribir(w, 200, c)
}
