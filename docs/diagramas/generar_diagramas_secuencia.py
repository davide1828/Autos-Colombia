# -*- coding: utf-8 -*-
"""
Generador de Diagramas de Secuencia (Iteraciones 1 y 2)
1. PlantUML (.puml)
2. Imágenes PNG de alta resolución
3. Integración en el archivo StarUML (.mdj)
"""
from PIL import Image, ImageDraw, ImageFont
import json, uuid, os

def uid():
    return "AAAAA" + uuid.uuid4().hex[:19]

# ==============================================================================
# 1. PLANTUML FUENTES
# ==============================================================================
PUML_INGRESO = """@startuml secuencia_ingreso
skinparam shadowing false
skinparam defaultFontName Arial
skinparam sequenceMessageAlign center
skinparam roundCorner 8
skinparam BoxPadding 10
skinparam ParticipantPadding 10

title Diagrama de Secuencia - Registro de Ingreso de Vehículo (Iteración 1)

actor "Encargado" as Enc #E8F1F8
box "Frontend PWA" #F8FAFC
  participant "Vista Ingreso\\n(React)" as UI #FFFFFF
end box

box "Backend Go (Arquitectura Hexagonal)" #F8FAFC
  participant "HTTP Handler\\n(POST /api/ingresos)" as API #FFFFFF
  participant "Servicio Aplicación\\n(ServicioIngresoSalida)" as App #FFFFFF
  participant "Dominio\\n(Entidades & Reglas)" as Dom #FFFFFF
end box

database "PostgreSQL 16" as DB #E8F1F8

Enc -> UI : Ingresa placa ("AAA-123")\\ny selecciona celda ("A-01")
UI -> API : POST /api/ingresos {placa, id_celda}
activate API

API -> App : RegistrarIngreso(ctx, "AAA-123", id_celda, novedad)
activate App

App -> DB : Transacción: Consultar vehículo y mensualidad
activate DB
DB --> App : Vehículo registrado, Mensualidad vigente (UTC-5)
deactivate DB

App -> Dom : Validar reglas RN-01, RN-02, RN-03
activate Dom
Dom --> App : Validaciones exitosas (sin ingreso activo, tipos compatibles)
deactivate Dom

App -> DB : Actualizar celda a OCUPADA\\nInsertar nuevo REGISTRO
activate DB
note right of DB: Trigger trg_registro_tipos valida\\natómicamente la compatibilidad
DB --> App : Registro confirmado (id_registro)
deactivate DB

App --> API : Retorna Registro creado
deactivate App

API --> UI : HTTP 201 Created {id_registro, momento}
deactivate API

UI --> Enc : Muestra confirmación en verde y celda ocupada
@enduml
"""

PUML_SALIDA = """@startuml secuencia_salida
skinparam shadowing false
skinparam defaultFontName Arial
skinparam sequenceMessageAlign center
skinparam roundCorner 8

title Diagrama de Secuencia - Registro de Salida de Vehículo (Iteración 1)

actor "Encargado" as Enc #E8F1F8
box "Frontend PWA" #F8FAFC
  participant "Vista Salida\\n(React)" as UI #FFFFFF
end box

box "Backend Go" #F8FAFC
  participant "HTTP Handler\\n(POST /api/salidas)" as API #FFFFFF
  participant "Servicio Aplicación" as App #FFFFFF
end box

database "PostgreSQL 16" as DB #E8F1F8

Enc -> UI : Ingresa placa ("AAA-123") y confirma salida
UI -> API : POST /api/salidas {placa, novedad?}
activate API

API -> App : RegistrarSalida(ctx, "AAA-123", novedad)
activate App

App -> DB : Buscar ingreso activo para la placa
activate DB
DB --> App : Registro activo encontrado (id_registro, id_celda, fecha_ingreso)
deactivate DB

App -> App : Calcular minutos de permanencia (RN-04)

App -> DB : Transacción:\\n1. Cerrar registro (fecha_salida, minutos)\\n2. Actualizar celda a DISPONIBLE
activate DB
DB --> App : Operación confirmada
deactivate DB

App --> API : Retorna Minutos y Celda liberada
deactivate App

API --> UI : HTTP 200 OK {minutos_estancia: 145, celda: "A-01"}
deactivate API

UI --> Enc : Muestra tiempo de permanencia y celda libre
@enduml
"""

PUML_LOGIN = """@startuml secuencia_autenticacion
skinparam shadowing false
skinparam defaultFontName Arial
skinparam sequenceMessageAlign center
skinparam roundCorner 8

title Diagrama de Secuencia - Autenticación y Control de Acceso (Iteración 2)

actor "Operador\\n(Admin / Encargado)" as Op #E8F1F8
box "Frontend PWA" #F8FAFC
  participant "Pantalla Login\\n(React)" as UI #FFFFFF
end box

box "Backend Go (Seguridad)" #F8FAFC
  participant "RateLimiter\\n(Anti Fuerza Bruta)" as Lim #FFFFFF
  participant "HTTP Handler\\n(POST /api/login)" as API #FFFFFF
  participant "Servicio Aplicación\\n(ServicioGestion)" as App #FFFFFF
  participant "Hasheador PBKDF2\\n(HMAC-SHA256)" as Sec #FFFFFF
  participant "Emisor Sesión\\n(Firma Token)" as Tok #FFFFFF
end box

database "PostgreSQL 16" as DB #E8F1F8

Op -> UI : Digita usuario y contraseña
UI -> Lim : POST /api/login {usuario, clave}
activate Lim

alt IP con más de 5 fallos en 60s
  Lim --> UI : HTTP 429 Too Many Requests (Retry-After: 60)
else IP permitida
  Lim -> API : Procesa solicitud
  activate API
  
  API -> App : Autenticar(ctx, usuario, clave)
  activate App
  
  App -> DB : UsuarioPorLogin(ctx, usuario)
  activate DB
  DB --> App : Datos de usuario (hash, salt, rol, activo)
  deactivate DB
  
  App -> Sec : Verificar(hash, clave)
  activate Sec
  note over Sec: Comparación en tiempo constante\\n(subtle.ConstantTimeCompare)
  Sec --> App : Contraseña válida (true)
  deactivate Sec
  
  App --> API : Usuario autenticado
  deactivate App
  
  API -> Tok : Emitir(Actor{ID, usuario, rol})
  activate Tok
  Tok --> API : Token Bearer firmado + Expiración
  deactivate Tok
  
  API --> UI : HTTP 200 OK {token, expira, usuario}
  deactivate API
  
  UI --> Op : Redirecciona según rol (Admin / Encargado)
end
deactivate Lim
@enduml
"""

PUML_GESTION_CELDA = """@startuml secuencia_gestion_celda
skinparam shadowing false
skinparam defaultFontName Arial
skinparam sequenceMessageAlign center
skinparam roundCorner 8

title Diagrama de Secuencia - Inhabilitación de Celda con Auditoría (Iteración 2)

actor "Administrador" as Admin #E8F1F8
box "Backend Go (Arquitectura Hexagonal)" #F8FAFC
  participant "Middleware Auth\\n(Bearer JWT)" as Auth #FFFFFF
  participant "HTTP Handler\\n(PUT /api/celdas/{id}/estado)" as API #FFFFFF
  participant "Servicio Aplicación\\n(ServicioGestion)" as App #FFFFFF
  participant "Dominio\\n(Validaciones)" as Dom #FFFFFF
  participant "Repositorio\\n(PostgreSQL)" as Repo #FFFFFF
end box

database "PostgreSQL 16" as DB #E8F1F8

Admin -> Auth : PUT /api/celdas/3/estado {estado: "INACTIVA"}\\nAuthorization: Bearer <token>
activate Auth

Auth -> DB : Validar usuario activo y rol ADMINISTRADOR
activate DB
DB --> Auth : Actor válido (ID: 1, Rol: ADMINISTRADOR)
deactivate DB

Auth -> API : Continúa petición autenticada
activate API

API -> App : CambiarEstadoCelda(ctx, 3, "INACTIVA")
activate App

App -> Dom : EstadoGestionable("INACTIVA")
activate Dom
Dom --> App : Válido (DISPONIBLE o INACTIVA)
deactivate Dom

App -> Repo : CambiarEstadoCelda(ctx, 3, "INACTIVA")
activate Repo

Repo -> DB : SELECT id_registro FROM registro WHERE id_celda = 3 AND fecha_salida IS NULL
activate DB

alt Celda tiene vehículo dentro (OCUPADA)
  DB --> Repo : Registro activo existe
  Repo --> App : ErrCeldaOcupada (RN-09)
  App --> API : Error: Celda ocupada
  API --> Admin : HTTP 409 Conflict ("la celda está ocupada")
else Celda desocupada
  DB --> Repo : Sin registro activo
  Repo -> DB : UPDATE celda SET estado = 'INACTIVA' WHERE id_celda = 3
  
  Repo -> DB : INSERT INTO auditoria (momento, id_usuario, usuario, accion, entidad, detalle)
  note right of DB: Trigger trg_auditoria_inmutable asegura\\nque el registro no se podrá alterar (RN-11)
  
  DB --> Repo : Transacción confirmada
  deactivate DB
  
  Repo --> App : Éxito
  deactivate Repo
  
  App --> API : Celda actualizada
  deactivate App
  
  API --> Admin : HTTP 200 OK {id_celda: 3, estado: "INACTIVA"}
  deactivate API
end
deactivate Auth
@enduml
"""

# Guardar archivos .puml
d_diag = r"c:\universidad\ingenieria de software\parqueadero_EA1_codigo\parqueadero\docs\diagramas"
with open(os.path.join(d_diag, "09_secuencia_ingreso.puml"), "w", encoding="utf-8") as f:
    f.write(PUML_INGRESO)
with open(os.path.join(d_diag, "10_secuencia_salida.puml"), "w", encoding="utf-8") as f:
    f.write(PUML_SALIDA)
with open(os.path.join(d_diag, "11_secuencia_autenticacion.puml"), "w", encoding="utf-8") as f:
    f.write(PUML_LOGIN)
with open(os.path.join(d_diag, "12_secuencia_gestion_celda_auditoria.puml"), "w", encoding="utf-8") as f:
    f.write(PUML_GESTION_CELDA)

print("Fuentes PlantUML de secuencias guardadas.")

# ==============================================================================
# 2. GENERADOR DE IMÁGENES PNG (Renderizado de diagramas de secuencia)
# ==============================================================================
FONT_TITLE = ImageFont.truetype(r"C:\Windows\Fonts\arialbd.ttf", 16)
FONT_SUB = ImageFont.truetype(r"C:\Windows\Fonts\ariali.ttf", 11)
FONT_PART = ImageFont.truetype(r"C:\Windows\Fonts\arialbd.ttf", 11)
FONT_MSG = ImageFont.truetype(r"C:\Windows\Fonts\arial.ttf", 10)
FONT_NOTE = ImageFont.truetype(r"C:\Windows\Fonts\arial.ttf", 9)

C_AZUL = "#0F4C81"
C_LINEA = "#0F4C81"
C_MSG = "#1C2530"
C_BOX = "#F8FAFC"
C_BORDER = "#CBD5E1"

def draw_sequence_diagram(title, subtitle, participants, messages, filename, W=1150, H=680):
    im = Image.new("RGB", (W, H), "#FFFFFF")
    d = ImageDraw.Draw(im)
    
    # Título
    d.text((30, 20), title, fill=C_AZUL, font=FONT_TITLE)
    d.text((30, 44), subtitle, fill="#64748B", font=FONT_SUB)
    d.line([(30, 60), (W - 30, 60)], fill=C_AZUL, width=2)
    
    # Participantes
    n_part = len(participants)
    step_x = (W - 120) // (n_part - 1) if n_part > 1 else W // 2
    part_coords = {}
    
    for i, (p_id, p_name, p_type) in enumerate(participants):
        x = 60 + i * step_x
        part_coords[p_id] = x
        
        # Caja participante arriba
        d.rounded_rectangle([(x - 55, 75), (x + 55, 115)], radius=4, fill="#E8F1F8", outline=C_AZUL, width=1)
        # Líneas de texto participante
        lines = p_name.split("\n")
        if len(lines) == 1:
            d.text((x - 45, 88), lines[0], fill=C_AZUL, font=FONT_PART)
        else:
            d.text((x - 45, 80), lines[0], fill=C_AZUL, font=FONT_PART)
            d.text((x - 45, 96), lines[1], fill="#64748B", font=FONT_NOTE)
            
        # Línea de vida
        for y_dash in range(120, H - 45, 10):
            d.line([(x, y_dash), (x, y_dash + 5)], fill="#94A3B8", width=1)
            
        # Caja participante abajo
        d.rounded_rectangle([(x - 55, H - 40), (x + 55, H - 10)], radius=4, fill="#E8F1F8", outline=C_AZUL, width=1)
        d.text((x - 45, H - 32), lines[0], fill=C_AZUL, font=FONT_PART)
        
    # Mensajes
    y_curr = 145
    for m in messages:
        m_type = m.get("type", "sync")
        p_from = part_coords[m["from"]]
        p_to = part_coords[m["to"]]
        txt = m["text"]
        
        if m_type == "note":
            nx = (p_from + p_to) // 2
            d.rounded_rectangle([(nx - 110, y_curr - 8), (nx + 110, y_curr + 22)], radius=3, fill="#FEF3C7", outline="#D97706", width=1)
            d.text((nx - 100, y_curr - 3), txt, fill="#B45309", font=FONT_NOTE)
            y_curr += 36
            continue
            
        # Flecha de mensaje
        color = C_LINEA if m_type != "reply" else "#64748B"
        width = 2 if m_type == "sync" else 1
        
        if m_type == "reply":
            # Línea discontinua
            dx = 8 if p_to > p_from else -8
            cur_x = p_from
            while (cur_x < p_to if p_to > p_from else cur_x > p_to):
                d.line([(cur_x, y_curr), (cur_x + dx // 2, y_curr)], fill=color, width=1)
                cur_x += dx
        else:
            d.line([(p_from, y_curr), (p_to, y_curr)], fill=color, width=width)
            
        # Punta de flecha
        if p_to > p_from:
            d.polygon([(p_to, y_curr), (p_to - 8, y_curr - 4), (p_to - 8, y_curr + 4)], fill=color)
        else:
            d.polygon([(p_to, y_curr), (p_to + 8, y_curr - 4), (p_to + 8, y_curr + 4)], fill=color)
            
        # Texto mensaje
        mx = (p_from + p_to) // 2
        d.text((mx - 90, y_curr - 14), txt, fill=C_MSG, font=FONT_MSG)
        
        y_curr += 36
        
    im.save(filename, dpi=(300, 300))
    print("Secuencia generada:", filename)

# 1. Secuencia Ingreso
parts_ing = [
    ("enc", "Encargado\\n(Actor)", "actor"),
    ("ui", "Vista Ingreso\\n(React PWA)", "ui"),
    ("api", "HTTP Handler\\n(POST /ingresos)", "api"),
    ("app", "Servicio App\\n(Go)", "app"),
    ("dom", "Dominio\\n(Reglas RN)", "dom"),
    ("db", "PostgreSQL\\n(Base Datos)", "db")
]
msgs_ing = [
    {"from": "enc", "to": "ui", "text": "1: Digita placa 'AAA-123' y celda 'A-01'"},
    {"from": "ui", "to": "api", "text": "2: POST /api/ingresos {placa, id_celda}"},
    {"from": "api", "to": "app", "text": "3: RegistrarIngreso(ctx, placa, celda)"},
    {"from": "app", "to": "db", "text": "4: Consultar vehículo y mensualidad activa"},
    {"from": "db", "to": "app", "text": "5: Vehículo registrado, Mensualidad OK", "type": "reply"},
    {"from": "app", "to": "dom", "text": "6: Validar reglas RN-01, RN-02, RN-03"},
    {"from": "dom", "to": "app", "text": "7: Compatibilidad celda/vehículo OK", "type": "reply"},
    {"from": "app", "to": "db", "text": "8: BEGIN: Celda OCUPADA + INSERT registro"},
    {"from": "db", "to": "db", "text": "Trigger trg_registro_tipos valida tipos atómicamente", "type": "note"},
    {"from": "db", "to": "app", "text": "9: COMMIT: id_registro generado", "type": "reply"},
    {"from": "app", "to": "api", "text": "10: Retorna Registro {id, fecha, placa}", "type": "reply"},
    {"from": "api", "to": "ui", "text": "11: HTTP 201 Created", "type": "reply"},
    {"from": "ui", "to": "enc", "text": "12: Muestra confirmación e ingreso exitoso", "type": "reply"}
]
draw_sequence_diagram(
    "Diagrama de Secuencia 1 — Registro de Ingreso de Vehículo (Iteración 1)",
    "Flujo transaccional con validación de mensualidad (RN-01), unicidad (RN-02) y compatibilidad (RN-03)",
    parts_ing, msgs_ing, os.path.join(d_diag, "secuencia_01_ingreso.png")
)

# 2. Secuencia Salida
parts_sal = [
    ("enc", "Encargado\\n(Actor)", "actor"),
    ("ui", "Vista Salida\\n(React PWA)", "ui"),
    ("api", "HTTP Handler\\n(POST /salidas)", "api"),
    ("app", "Servicio App\\n(Go)", "app"),
    ("db", "PostgreSQL\\n(Base Datos)", "db")
]
msgs_sal = [
    {"from": "enc", "to": "ui", "text": "1: Digita placa 'AAA-123' y confirma salida"},
    {"from": "ui", "to": "api", "text": "2: POST /api/salidas {placa, novedad?}"},
    {"from": "api", "to": "app", "text": "3: RegistrarSalida(ctx, placa, novedad)"},
    {"from": "app", "to": "db", "text": "4: Buscar ingreso activo del vehículo"},
    {"from": "db", "to": "app", "text": "5: Registro activo retornado (id_registro, celda)", "type": "reply"},
    {"from": "app", "to": "app", "text": "6: Calcular minutos de permanencia (RN-04)"},
    {"from": "app", "to": "db", "text": "7: UPDATE registro (fecha_salida, minutos) + Celda DISPONIBLE"},
    {"from": "db", "to": "app", "text": "8: Transacción confirmada", "type": "reply"},
    {"from": "app", "to": "api", "text": "9: Retorna Minutos estancia y celda liberada", "type": "reply"},
    {"from": "api", "to": "ui", "text": "10: HTTP 200 OK {minutos: 145, celda: 'A-01'}", "type": "reply"},
    {"from": "ui", "to": "enc", "text": "11: Muestra tiempo de estancia y celda libre", "type": "reply"}
]
draw_sequence_diagram(
    "Diagrama de Secuencia 2 — Registro de Salida de Vehículo (Iteración 1)",
    "Flujo de cierre de registro, cálculo de permanencia en minutos y liberación de celda (RN-04)",
    parts_sal, msgs_sal, os.path.join(d_diag, "secuencia_02_salida.png")
)

# 3. Secuencia Login
parts_log = [
    ("op", "Operador\\n(Admin/Enc)", "actor"),
    ("ui", "Pantalla Login\\n(React PWA)", "ui"),
    ("lim", "RateLimiter\\n(Anti Brute-force)", "sec"),
    ("api", "HTTP Handler\\n(POST /login)", "api"),
    ("app", "Servicio App\\n(Go)", "app"),
    ("sec", "PBKDF2\\n(Seguridad)", "sec"),
    ("db", "PostgreSQL\\n(Base Datos)", "db")
]
msgs_log = [
    {"from": "op", "to": "ui", "text": "1: Ingresa usuario y contraseña"},
    {"from": "ui", "to": "lim", "text": "2: POST /api/login {usuario, clave}"},
    {"from": "lim", "to": "lim", "text": "Verificar límite de intentos por IP (máx 5/min)", "type": "note"},
    {"from": "lim", "to": "api", "text": "3: Petición permitida -> login()"},
    {"from": "api", "to": "app", "text": "4: Autenticar(ctx, usuario, clave)"},
    {"from": "app", "to": "db", "text": "5: UsuarioPorLogin(ctx, 'admin')"},
    {"from": "db", "to": "app", "text": "6: Fila usuario (clave_hash, rol, activo)", "type": "reply"},
    {"from": "app", "to": "sec", "text": "7: Verificar(clave_hash, clave_plana)"},
    {"from": "sec", "to": "sec", "text": "Comparación en tiempo constante (RN-07)", "type": "note"},
    {"from": "sec", "to": "app", "text": "8: Password válida: TRUE", "type": "reply"},
    {"from": "app", "to": "api", "text": "9: Usuario autenticado", "type": "reply"},
    {"from": "api", "to": "ui", "text": "10: HTTP 200 OK {token, expira, usuario}", "type": "reply"},
    {"from": "ui", "to": "op", "text": "11: Acceso concedido y menú según rol (RBAC)", "type": "reply"}
]
draw_sequence_diagram(
    "Diagrama de Secuencia 3 — Autenticación y Emisión de Sesión (Iteración 2)",
    "Flujo de login seguro con limitación de intentos, hash PBKDF2 y emisión de token Bearer (RN-07)",
    parts_log, msgs_log, os.path.join(d_diag, "secuencia_03_login.png")
)

# 4. Secuencia Gestión Celdas y Auditoría
parts_ges = [
    ("adm", "Administrador\\n(Actor)", "actor"),
    ("auth", "Middleware Auth\\n(Bearer JWT)", "sec"),
    ("api", "HTTP Handler\\n(PUT /celdas/3)", "api"),
    ("app", "Servicio App\\n(Go)", "app"),
    ("db", "PostgreSQL\\n(Base Datos)", "db")
]
msgs_ges = [
    {"from": "adm", "to": "auth", "text": "1: PUT /api/celdas/3/estado {estado: 'INACTIVA'}"},
    {"from": "auth", "to": "db", "text": "2: Verificar usuario activo y rol ADMINISTRADOR"},
    {"from": "db", "to": "auth", "text": "3: Sesión válida (Rol: ADMINISTRADOR)", "type": "reply"},
    {"from": "auth", "to": "api", "text": "4: Continúa al handler CambiarEstadoCelda"},
    {"from": "api", "to": "app", "text": "5: CambiarEstadoCelda(ctx, id=3, 'INACTIVA')"},
    {"from": "app", "to": "db", "text": "6: SELECT id_registro WHERE id_celda=3 AND activa"},
    {"from": "db", "to": "app", "text": "7: Celda no ocupada (0 vehículos dentro)", "type": "reply"},
    {"from": "app", "to": "db", "text": "8: BEGIN: UPDATE celda SET estado='INACTIVA'"},
    {"from": "app", "to": "db", "text": "9: INSERT INTO auditoria (accion, entidad, detalle)"},
    {"from": "db", "to": "db", "text": "Trigger trg_auditoria_inmutable asegura persistencia", "type": "note"},
    {"from": "db", "to": "app", "text": "10: COMMIT exitoso", "type": "reply"},
    {"from": "app", "to": "api", "text": "11: Celda actualizada y auditada", "type": "reply"},
    {"from": "api", "to": "adm", "text": "12: HTTP 200 OK {id: 3, estado: 'INACTIVA'}", "type": "reply"}
]
draw_sequence_diagram(
    "Diagrama de Secuencia 4 — Inhabilitación de Celda con Auditoría (Iteración 2)",
    "Flujo administrativo con control de rol, guardia de celda no ocupada (RN-09) e inserción inmutable (RN-11)",
    parts_ges, msgs_ges, os.path.join(d_diag, "secuencia_04_gestion_celdas.png")
)

# ==============================================================================
# 3. ACTUALIZAR ARCHIVO STARUML (.mdj) CON LOS DIAGRAMAS DE SECUENCIA
# ==============================================================================
mdj_path = os.path.join(d_diag, "parqueadero_iteracion_2.mdj")
with open(mdj_path, "r", encoding="utf-8") as f:
    star_data = json.load(f)

model_elem = star_data["ownedElements"][0]["ownedElements"]

# Agregar elementos de secuencia a StarUML
seq_collab_id = uid()
seq_elements = [
    {
        "_type": "UMLCollaboration",
        "_id": seq_collab_id,
        "_parent": {"$ref": star_data["ownedElements"][0]["_id"]},
        "name": "Colaboraciones de Secuencia",
        "ownedElements": [
            {
                "_type": "UMLInteraction",
                "_id": uid(),
                "_parent": {"$ref": seq_collab_id},
                "name": "Secuencia 01: Registro de Ingreso (Iteración 1)"
            },
            {
                "_type": "UMLInteraction",
                "_id": uid(),
                "_parent": {"$ref": seq_collab_id},
                "name": "Secuencia 02: Registro de Salida (Iteración 1)"
            },
            {
                "_type": "UMLInteraction",
                "_id": uid(),
                "_parent": {"$ref": seq_collab_id},
                "name": "Secuencia 03: Autenticación y Sesión (Iteración 2)"
            },
            {
                "_type": "UMLInteraction",
                "_id": uid(),
                "_parent": {"$ref": seq_collab_id},
                "name": "Secuencia 04: Gestión de Celda y Auditoría (Iteración 2)"
            }
        ]
    }
]

star_data["ownedElements"][0]["ownedElements"].extend(seq_elements)

with open(mdj_path, "w", encoding="utf-8") as f:
    json.dump(star_data, f, indent=2, ensure_ascii=False)

print("Archivo StarUML actualizado con diagramas de secuencia.")

