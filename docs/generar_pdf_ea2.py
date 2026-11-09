# -*- coding: utf-8 -*-
"""
Generador del Documento PDF Oficial — Evidencia de Aprendizaje 2 (EA2)
Sistema de Información Parqueadero "Autos Colombia"
Estudiante: Cristian David España Figueroa
"""
import sys, os, json
from PIL import Image
from reportlab.lib.pagesizes import letter
from reportlab.lib.units import cm
from reportlab.lib import colors
from reportlab.lib.styles import ParagraphStyle
from reportlab.lib.enums import TA_CENTER, TA_JUSTIFY, TA_LEFT, TA_RIGHT
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont
from reportlab.platypus import (BaseDocTemplate, PageTemplate, Frame, Paragraph, Spacer, PageBreak,
                                Table, TableStyle, Image as RLImage, KeepTogether)

# Registro de fuentes TrueType de Windows
pdfmetrics.registerFont(TTFont("Arial", r"C:\Windows\Fonts\arial.ttf"))
pdfmetrics.registerFont(TTFont("Arial-B", r"C:\Windows\Fonts\arialbd.ttf"))
pdfmetrics.registerFont(TTFont("Arial-I", r"C:\Windows\Fonts\ariali.ttf"))
pdfmetrics.registerFont(TTFont("Arial-BI", r"C:\Windows\Fonts\arialbi.ttf"))
pdfmetrics.registerFontFamily("Arial", normal="Arial", bold="Arial-B", italic="Arial-I", boldItalic="Arial-BI")

AZUL = colors.HexColor("#0f4c81")
AZUL_CLARO = colors.HexColor("#e8f1f8")
GRIS = colors.HexColor("#f8fafc")
BORDE = colors.HexColor("#cbd5e1")
TEXTO = colors.HexColor("#1c2530")

AQUI = os.path.dirname(os.path.abspath(__file__))
cfg_path = os.path.join(AQUI, "pdf_config_ea2.json")
cfg = json.load(open(cfg_path, encoding="utf-8")) if os.path.exists(cfg_path) else {}

NOMBRE = cfg.get("nombre", "Cristian David España Figueroa")
MATERIA = cfg.get("materia", "Ingeniería de Software")
DOCENTE = cfg.get("docente", "Esteban Montoya Molina")
CODIGO = cfg.get("codigo", "PREICA2602B020081")
PERIODO = cfg.get("periodo", "2026_2_B2")
GITHUB = cfg.get("github", "https://github.com/davide1828/Autos-Colombia.git")
FIGMA = cfg.get("figma", "https://www.figma.com/design/DhCeukFFGQef4egiIl3UbY/")
SALIDA = os.path.join(AQUI, "España_Cristian_EA2.pdf")

W = letter[0] - 4.4 * cm

def S(name, **kw):
    base = dict(fontName="Arial", fontSize=9.5, leading=14, textColor=TEXTO)
    base.update(kw)
    return ParagraphStyle(name, **base)

body = S("body", alignment=TA_JUSTIFY, spaceAfter=6)
body_bold = S("body_bold", fontName="Arial-B", spaceAfter=6)
small = S("small", fontSize=8, leading=11, textColor=colors.HexColor("#4a5663"))
cap = S("cap", fontSize=8, leading=11, alignment=TA_CENTER, textColor=colors.HexColor("#4a5663"), spaceAfter=10, fontName="Arial-I")
h1 = S("h1", fontName="Arial-B", fontSize=15, leading=19, textColor=AZUL, spaceBefore=12, spaceAfter=8)
h2 = S("h2", fontName="Arial-B", fontSize=11.5, leading=15, textColor=AZUL, spaceBefore=9, spaceAfter=5)
h3 = S("h3", fontName="Arial-B", fontSize=9.5, leading=13, spaceBefore=6, spaceAfter=3)

cell = S("cell", fontSize=8, leading=10.5)
cell_bold = S("cell_bold", fontSize=8, leading=10.5, fontName="Arial-B")
cellb = S("cellb", fontSize=8, leading=10.5, fontName="Arial-B", textColor=colors.white)

def P(t, st=body): return Paragraph(t, st)

def tabla(filas, anchos, cab=True):
    data = []
    for i, f in enumerate(filas):
        fila = []
        for c in f:
            if isinstance(c, str):
                st = cellb if (cab and i == 0) else cell
                fila.append(Paragraph(c, st))
            else:
                fila.append(c)
        data.append(fila)
    t = Table(data, colWidths=anchos, repeatRows=1 if cab else 0)
    st = [
        ("GRID", (0, 0), (-1, -1), 0.4, BORDE),
        ("VALIGN", (0, 0), (-1, -1), "TOP"),
        ("LEFTPADDING", (0, 0), (-1, -1), 4),
        ("RIGHTPADDING", (0, 0), (-1, -1), 4),
        ("TOPPADDING", (0, 0), (-1, -1), 3),
        ("BOTTOMPADDING", (0, 0), (-1, -1), 3)
    ]
    if cab:
        st += [("BACKGROUND", (0, 0), (-1, 0), AZUL)]
    for i in range(1, len(data)):
        if i % 2 == 0:
            st.append(("BACKGROUND", (0, i), (-1, i), GRIS))
    t.setStyle(TableStyle(st))
    return t

def img(path, ancho_max, alto_max):
    if not os.path.exists(path):
        return P(f"[Imagen no encontrada: {path}]", cap)
    w, h = Image.open(path).size
    esc = min(ancho_max / w, alto_max / h)
    return RLImage(path, width=w * esc, height=h * esc)

def figura(path, titulo, ancho=16.4 * cm, alto=13 * cm):
    return KeepTogether([img(path, ancho, alto), Spacer(1, 3), P(titulo, cap)])

# Encabezado y pie de página
def pie(c, d):
    c.saveState()
    c.setFont("Arial", 7.5)
    c.setFillColor(colors.HexColor("#4a5663"))
    c.drawString(2.2 * cm, 1.2 * cm, "Parqueadero Autos Colombia · EA2 · Iteración 2")
    c.drawRightString(letter[0] - 2.2 * cm, 1.2 * cm, f"Página {d.page}")
    c.setStrokeColor(BORDE)
    c.line(2.2 * cm, 1.5 * cm, letter[0] - 2.2 * cm, 1.5 * cm)
    c.restoreState()

def portada(c, d):
    c.saveState()
    c.setFillColor(AZUL)
    c.rect(0, letter[1] - 8.5 * cm, letter[0], 8.5 * cm, fill=1, stroke=0)
    c.restoreState()

doc = BaseDocTemplate(SALIDA, pagesize=letter, leftMargin=2.2 * cm, rightMargin=2.2 * cm, topMargin=2.2 * cm, bottomMargin=2.2 * cm,
                      title="Sistema de Información Autos Colombia - EA2", author=NOMBRE)
fr = Frame(doc.leftMargin, doc.bottomMargin, W, letter[1] - 4.4 * cm, id="f")
doc.addPageTemplates([PageTemplate("portada", [fr], onPage=portada), PageTemplate("normal", [fr], onPage=pie)])

E = []

# ==============================================================================
# PORTADA
# ==============================================================================
E += [
    Spacer(1, 1.0 * cm),
    P("S20 · EVIDENCIA DE APRENDIZAJE 2", S("p0", fontName="Arial-B", fontSize=9.5, leading=13, textColor=colors.HexColor("#b9d5ee"))),
    Spacer(1, 4),
    P("Sistema de información del<br/>parqueadero “Autos Colombia”", S("p1", fontName="Arial-B", fontSize=21, leading=25, textColor=colors.white)),
    Spacer(1, 6),
    P("Diseño de Sistema de Información – Iteración 2: Gestión de usuarios y gestión de celdas",
      S("p2", fontName="Arial", fontSize=11, leading=15, textColor=colors.HexColor("#e8f1f8"))),
    Spacer(1, 3.8 * cm),
    tabla([
        ["Estudiante", NOMBRE],
        ["Programa", "Tecnología en Desarrollo de Software"],
        ["Institución", "Institución Universitaria Digital de Antioquia (IU Digital)"],
        ["Asignatura", MATERIA],
        ["Docente", DOCENTE],
        ["Código curso", CODIGO],
        ["Periodo", PERIODO],
        ["Repositorio GitHub", GITHUB],
        ["Diseño en Figma", FIGMA],
        ["Ciudad y fecha", "Medellín, Colombia · Noviembre de 2026"]
    ], [4.5 * cm, W - 4.5 * cm], cab=False),
    PageBreak()
]

# ==============================================================================
# TABLA DE CONTENIDO E INTRODUCCIÓN
# ==============================================================================
E += [
    P("Contenido del Documento", h1),
    P("1. Introducción y Enfoque de la Iteración 2<br/>"
      "2. Requerimientos Funcionales y No Funcionales (Gestión de Usuarios y Celdas)<br/>"
      "3. Diagramas de Casos de Uso en STARUML<br/>"
      "4. Historias de Usuario Asociadas (Formato Ágil y Gherkin)<br/>"
      "5. Mockups de Interfaz de Usuario de las Historias de Gestión<br/>"
      "6. Diagramas de Secuencia de las Iteraciones 1 y 2 en STARUML<br/>"
      "7. Diagrama de Componentes de las Iteraciones 1 y 2 en STARUML<br/>"
      "8. Formato de Plan de Pruebas Resumido (Gestión de Usuarios y Celdas)<br/>"
      "9. Repositorio en GitHub y Evidencias de Código Fuente<br/>"
      "10. Matriz de Trazabilidad entre Representaciones<br/>"
      "11. Conclusiones"),
    Spacer(1, 10),
    P("1. Introducción y Enfoque de la Iteración 2", h1),
    P("El sistema de información del parqueadero <b>“Autos Colombia”</b> es una solución integral diseñada bajo principios de ingeniería de software para gestionar las operaciones cotidianas de un parqueadero que presta servicios de mensualidad. "
      "Tras haber completado y validado en la <b>Iteración 1</b> el núcleo operativo de patio (registro de ingreso, salida, permanencia, control de bahías y novedades bajo una arquitectura limpia y desacoplada), la presente <b>Iteración 2</b> aborda la gobernabilidad, administración y seguridad del sistema."),
    P("En esta iteración se incorporan dos ejes fundamentales: la <b>Gestión de Usuarios con control de acceso basado en roles (RBAC)</b> y la <b>Gestión Integral de Celdas</b>, respaldadas por un subsistema de <b>Auditoría Inmutable</b> a nivel de persistencia relacional. "
      "El sistema evoluciona de un esquema monousuario a una plataforma multiusuario donde se diferencian estrictamente las atribuciones del <i>Administrador</i> (gestión de identidades, configuración del patio, asignación de celdas y auditoría) y del <i>Encargado</i> (operación diaria y actualización de su propia clave)."),
    P("Todo el diseño formal presentado en este documento se encuentra respaldado por su implementación ejecutable en el backend (Go 1.22 con arquitectura hexagonal), pruebas automatizadas con 0 fallos, base de datos relacional PostgreSQL 16 con disparadores de integridad transaccional, y artefactos UML modelados en <b>StarUML</b>."),
    PageBreak()
]

# ==============================================================================
# ENTREGABLE 1: REQUERIMIENTOS FUNCIONALES Y NO FUNCIONALES
# ==============================================================================
E += [
    P("2. Requerimientos Funcionales y No Funcionales", h1),
    P("Un requerimiento funcional define los servicios y comportamientos del software en respuesta a estímulos específicos. "
      "Para la Iteración 2, los requerimientos se formulan alineados con las reglas de negocio implementadas y garantizadas tanto en la capa de dominio como mediante restricciones y triggers en la base de datos."),
    P("Reglas de Negocio Incorporadas en la Iteración 2", h2),
    P("• <b>RN-07 (Autenticación y Sesión Segura):</b> Todo acceso al sistema exige credenciales individuales. Las contraseñas se almacenan mediante hash con sal criptográfica individual (PBKDF2-SHA256, 100.000 iteraciones). Se bloquean solicitudes tras 5 intentos fallidos consecutivos por IP.<br/>"
      "• <b>RN-08 (Gobernabilidad y Último Administrador):</b> El sistema debe asegurar en todo momento la presencia de al menos un usuario activo con rol ADMINISTRADOR. Ninguna operación puede degradar o inactivar al último administrador existente.<br/>"
      "• <b>RN-09 (Integridad de Celdas Ocupadas):</b> Una celda en estado OCUPADA no puede pasar a INACTIVA ni ser modificada en su código o tipo mientras mantenga un vehículo dentro.<br/>"
      "• <b>RN-10 (No Solapamiento de Mensualidades):</b> No se permiten periodos de mensualidad superpuestos para el mismo cliente.<br/>"
      "• <b>RN-11 (Auditoría Inmutable):</b> Toda modificación administrativa sobre usuarios, celdas y clientes queda registrada en la tabla de auditoría, la cual rechaza cualquier UPDATE o DELETE mediante trigger transaccional."),
    Spacer(1, 4),
    P("Requerimientos Funcionales (RF-08 a RF-21)", h2),
    tabla([
        ["Id", "Requerimiento Funcional", "Regla", "Prioridad"],
        ["RF-08", "El sistema debe permitir a los usuarios (Administrador y Encargado) autenticarse mediante login y contraseña, retornando un token de sesión seguro (Bearer) con expiración y el perfil del usuario.", "RN-07", "Alta"],
        ["RF-09", "El sistema debe controlar los accesos según el rol (RBAC): restringir la gestión de usuarios, celdas, vehículos y auditoría al ADMINISTRADOR, y permitir al ENCARGADO operar patio y cambiar su clave.", "RN-07", "Alta"],
        ["RF-10", "El sistema debe permitir al Administrador registrar nuevos usuarios indicando login único (3-30 caracteres alfanuméricos), nombre, rol (ADMINISTRADOR/ENCARGADO) y contraseña inicial segura.", "RN-07", "Alta"],
        ["RF-11", "El sistema debe permitir al Administrador actualizar datos de usuarios y alternar su estado (activo/inactivo), bloqueando la inactivación o degradación del último administrador activo.", "RN-08", "Alta"],
        ["RF-12", "El sistema debe permitir al usuario autenticado actualizar su contraseña personal validando su contraseña actual y verificando que la nueva clave cumpla con los estándares de robustez.", "RN-07", "Media"],
        ["RF-13", "El sistema debe permitir al Administrador restablecer la contraseña de cualquier usuario sin necesidad de conocer la clave anterior, aplicando las directivas de seguridad.", "RN-07", "Media"],
        ["RF-14", "El sistema debe revocar de inmediato el acceso de cualquier usuario que sea inactivado o degradado de rol, consultando la base de datos en cada petición protegida.", "RN-07, RN-08", "Alta"],
        ["RF-15", "El sistema debe permitir al Administrador crear celdas con código alfanumérico único en mayúsculas de hasta 10 caracteres (ej. A-01) y tipo de vehículo (CARRO/MOTO) en estado inicial DISPONIBLE.", "RN-03", "Alta"],
        ["RF-16", "El sistema debe permitir al Administrador alternar el estado de una celda entre DISPONIBLE e INACTIVA (por mantenimiento o daño), impidiendo inhabilitarla si se encuentra ocupada.", "RN-09", "Alta"],
        ["RF-17", "El sistema debe permitir al Administrador editar código o tipo de una celda, validando que no tenga ningún vehículo dentro al momento de la modificación.", "RN-03, RN-09", "Media"],
        ["RF-18", "El sistema debe permitir registrar clientes con su documento, datos de contacto, vehículos y registrar mensualidades validando que no existan periodos solapados para el mismo cliente.", "RN-01, RN-10", "Alta"],
        ["RF-19", "El sistema debe permitir al Administrador corregir el tipo de un vehículo (CARRO/MOTO), validando que no se encuentre dentro del patio para no violar la compatibilidad con su celda.", "RN-03", "Media"],
        ["RF-20", "El sistema debe registrar automáticamente en la bitácora de auditoría cada creación, modificación o inactivación de usuarios, celdas y vehículos, impidiendo su edición o borrado.", "RN-11", "Alta"],
        ["RF-21", "El sistema debe permitir al Administrador consultar el histórico cronológico de auditoría con datos de momento UTC, usuario responsable, entidad afectada, ID y detalle.", "RN-11", "Media"]
    ], [1.8 * cm, W - 5.4 * cm, 2.0 * cm, 1.6 * cm]),
    Spacer(1, 6),
    P("Requerimientos No Funcionales (RNF-08 a RNF-13)", h2),
    tabla([
        ["Id", "Categoría", "Requerimiento No Funcional Verificable", "Métrica / Criterio"],
        ["RNF-08", "Seguridad Criptográfica", "Las contraseñas deben almacenarse hasheadas con PBKDF2 (HMAC-SHA256, sal de 16 bytes, 100.000 iteraciones). Las validaciones deben ejecutarse en tiempo constante para mitigar ataques de temporización.", "100 % pruebas unitarias superadas en seguridad; uso de subtle.ConstantTimeCompare."],
        ["RNF-09", "Anti Fuerza Bruta", "El endpoint de login debe limitar los intentos fallidos por IP (máximo 5 fallos por minuto). Si se excede, debe responder HTTP 429 con cabecera Retry-After: 60.", "Validado en pruebas de HTTP API con RateLimiter por IP."],
        ["RNF-10", "Integridad Concurrente", "Las operaciones críticas sobre celdas, vehículos y validación del último administrador deben ejecutarse en transacciones ACID con bloqueos pesimistas (FOR SHARE / FOR UPDATE).", "Sin condiciones de carrera bajo pruebas concurrentes en PostgreSQL."],
        ["RNF-11", "Inviolabilidad", "La tabla de auditoría debe ser inmutable a nivel de base de datos. Ningún usuario o proceso puede modificar ni borrar registros de bitácora.", "Trigger trg_auditoria_inmutable que arroja excepción ante UPDATE o DELETE."],
        ["RNF-12", "Rendimiento Auth", "La verificación de autenticación y extracción del actor en cada petición no debe incrementar la latencia en más de 5 ms.", "Medido en peticiones HTTP con servidor en Go."],
        ["RNF-13", "Mantenibilidad", "Arquitectura hexagonal en Go: desacoplamiento absoluto de dominio, casos de uso, HTTP y persistencia, con doble adaptador (PostgreSQL y Memoria).", "Casos de uso ejecutables sin base de datos externa para pruebas ágiles."]
    ], [1.8 * cm, 2.8 * cm, W - 9.0 * cm, 4.4 * cm]),
    PageBreak()
]

# ==============================================================================
# ENTREGABLE 2: CASOS DE USO EN STARUML
# ==============================================================================
E += [
    P("3. Diagramas de Casos de Uso en STARUML", h1),
    P("El diagrama de casos de uso modela el comportamiento del sistema desde el punto de vista de los actores externos, capturando las metas funcionales del negocio. "
      "El modelo fue diseñado y exportado en <b>StarUML</b> (archivo <font name='Arial-B'>docs/diagramas/parqueadero_iteracion_2.mdj</font>), estructurando dos actores y seis casos de uso principales."),
    figura(os.path.join(AQUI, "diagramas", "casos_uso_iteracion_2.png"),
           "Figura 1. Diagrama de Casos de Uso de la Iteración 2 (StarUML / docs/diagramas/08_casos_uso_iteracion_2.puml)", ancho=16.0 * cm, alto=11.0 * cm),
    P("Especificación de los Casos de Uso", h2),
    tabla([
        ["Caso de Uso", "Actores", "Descripción", "Precondición", "Postcondición", "Inclusiones <<include>>"],
        ["CU-06 Iniciar sesión", "Admin, Encargado", "Autentica al usuario en el sistema mediante usuario y clave, generando token de sesión.", "Usuario registrado y activo.", "Token Bearer emitido; reinicio de intentos fallidos.", "Validar credenciales y tasa."],
        ["CU-07 Cambiar contraseña", "Admin, Encargado", "Permite actualizar la contraseña personal confirmando la clave actual.", "Sesión activa.", "Contraseña actualizada; auditoría registrada.", "Validar robustez contraseña."],
        ["CU-08 Administrar usuarios", "Administrador", "Permite crear usuarios, editar nombres, alternar roles, inactivar y resetear claves.", "Admin autenticado.", "Usuario guardado; bitácora registrada.", "Validar robustez clave.<br/>Garantizar último admin.<br/>Registrar auditoría inmutable."],
        ["CU-09 Administrar celdas", "Administrador", "Permite registrar celdas, editar código/tipo e inhabilitar por mantenimiento.", "Admin autenticado.", "Celda creada/actualizada; auditoría registrada.", "Validar celda no ocupada.<br/>Registrar auditoría inmutable."],
        ["CU-10 Administrar clientes", "Admin, Encargado", "Registra clientes, asocia vehículos y renueva mensualidades sin solape.", "Usuario autenticado / Admin.", "Contrato registrado; vigencia actualizada.", "Validar no solapamiento fechas.<br/>Validar vehículo fuera de patio.<br/>Registrar auditoría inmutable."],
        ["CU-11 Consultar auditoría", "Administrador", "Consulta cronológica descendente de la bitácora inalterable de auditoría.", "Admin autenticado.", "Registros devueltos sin modificación.", "Ninguna (consulta pura)."]
    ], [2.4 * cm, 2.0 * cm, 3.4 * cm, 2.3 * cm, 2.5 * cm, W - 12.6 * cm]),
    PageBreak()
]

# ==============================================================================
# ENTREGABLE 3: HISTORIAS DE USUARIO
# ==============================================================================
E += [
    P("4. Historias de Usuario Asociadas a la Gestión", h1),
    P("Cada historia de usuario sigue el formato ágil estándar <i>Como [rol], quiero [capacidad], para [beneficio]</i>, acompañada de criterios de aceptación estructurados en lenguaje Gherkin (<i>Dado / Cuando / Entonces</i>)."),
    
    P("HU-06 · Gestión de usuarios del sistema (Relacionada con: RF-09, RF-10, RF-11, RF-14)", h2),
    P("<b>Como</b> Administrador del parqueadero, <b>quiero</b> registrar, modificar y administrar las cuentas de los usuarios del sistema asignando roles (ADMINISTRADOR / ENCARGADO), <b>para</b> controlar el acceso a la plataforma y aplicar el principio de menor privilegio."),
    P("• <b>CA-06.1 Registro válido:</b> Dado que el Administrador diligencia un login único (3-30 car), nombre válido, clave robusta y rol, Cuando solicita la creación, Entonces el sistema crea el usuario activo, almacena el hash PBKDF2, registra la auditoría y retorna HTTP 201 Created.<br/>"
      "• <b>CA-06.2 Login duplicado:</b> Dado que ya existe un usuario con el mismo login, Cuando se intenta registrar, Entonces el sistema rechaza con HTTP 409 Conflict ('ya existe un usuario con ese nombre de usuario').<br/>"
      "• <b>CA-06.3 Contraseña débil:</b> Dado que se ingresa una clave menor a 10 caracteres o sin números/letras, Cuando se envía la solicitud, Entonces responde HTTP 400 Bad Request explicando la política.<br/>"
      "• <b>CA-06.4 Inactivación de usuario:</b> Dado un Encargado activo, Cuando el Admin cambia su estado a inactivo, Entonces el sistema persiste el cambio y cualquier uso posterior de su token es denegado con HTTP 401.<br/>"
      "• <b>CA-06.5 Protección del último administrador activo:</b> Dado un único administrador activo, Cuando se intenta inactivar o cambiar su rol, Entonces el sistema rechaza con HTTP 409 Conflict ('debe quedar al menos un administrador activo')."),
    
    P("HU-07 · Autenticación y cambio de credenciales (Relacionada con: RF-08, RF-12, RF-13)", h2),
    P("<b>Como</b> usuario del sistema (Administrador o Encargado), <b>quiero</b> iniciar sesión de forma segura y actualizar mi contraseña personal, <b>para</b> acceder a las funciones operativas y mantener mi cuenta protegida."),
    P("• <b>CA-07.1 Login exitoso:</b> Dado un usuario activo y su contraseña correcta, Cuando envía credenciales a /api/login, Entonces recibe HTTP 200 OK con token Bearer, expiración y perfil.<br/>"
      "• <b>CA-07.2 Credenciales erróneas:</b> Dado un usuario o contraseña incorrectos, Cuando solicita login, Entonces recibe HTTP 401 Unauthorized ('usuario o contraseña incorrectos') y suma 1 fallo a la IP.<br/>"
      "• <b>CA-07.3 Bloqueo por fuerza bruta:</b> Dado que una IP acumula 5 intentos fallidos en < 60s, Cuando reintenta, Entonces recibe HTTP 429 Too Many Requests con cabecera Retry-After: 60.<br/>"
      "• <b>CA-07.4 Cambio de contraseña personal:</b> Dado un usuario autenticado que digita su clave actual correcta y una nueva clave válida, Cuando confirma el cambio, Entonces actualiza el hash PBKDF2 y retorna HTTP 200 OK.<br/>"
      "• <b>CA-07.5 Confirmación de clave actual errónea:</b> Dado que el usuario falla al digitar su clave actual, Cuando intenta cambiarla, Entonces el sistema responde HTTP 400 Bad Request sin alterar la clave."),
    
    P("HU-08 · Gestión y mantenimiento de celdas (Relacionada con: RF-15, RF-16, RF-17)", h2),
    P("<b>Como</b> Administrador, <b>quiero</b> registrar celdas y habilitar o inhabilitar bahías por obras o mantenimiento, <b>para</b> reflejar la disponibilidad real del patio y evitar asignaciones erróneas."),
    P("• <b>CA-08.1 Registro de celda:</b> Dado un código único (ej. A-05) y tipo CARRO/MOTO, Cuando se registra, Entonces se guarda en estado DISPONIBLE y retorna HTTP 201 Created.<br/>"
      "• <b>CA-08.2 Inhabilitación por mantenimiento:</b> Dado una celda DISPONIBLE, Cuando el Admin cambia su estado a INACTIVA, Entonces el sistema la excluye de los selectores operativos de ingreso.<br/>"
      "• <b>CA-08.3 Bloqueo en celda ocupada:</b> Dado una celda en estado OCUPADA con un vehículo dentro, Cuando el Admin intenta inhabilitarla o cambiar su tipo, Entonces el sistema rechaza con HTTP 409 Conflict ('la celda está ocupada; registre primero la salida del vehículo')."),
    
    P("HU-09 · Gestión de clientes y mensualidades (Relacionada con: RF-18, RF-19)", h2),
    P("<b>Como</b> operador del parqueadero, <b>quiero</b> registrar clientes, asociar sus vehículos y renovar contratos de mensualidad, <b>para</b> autorizar el acceso exclusivo a vehículos con pago al día."),
    P("• <b>CA-09.1 Mensualidad válida:</b> Dado un cliente y un rango de fechas sin conflicto, Cuando se agrega, Entonces se genera la mensualidad y se actualiza la vigencia del cliente.<br/>"
      "• <b>CA-09.2 Rechazo por solapamiento:</b> Dado un cliente con mensualidad vigente, Cuando se intenta registrar un periodo que se cruza con el existente, Entonces el sistema rechaza con HTTP 409 Conflict.<br/>"
      "• <b>CA-09.3 Modificación de tipo con vehículo dentro:</b> Dado un vehículo alojado en patio, Cuando se intenta modificar su tipo, Entonces se bloquea con HTTP 409 Conflict para mantener la compatibilidad con la celda."),
    
    P("HU-10 · Auditoría inmutable de operaciones (Relacionada con: RF-20, RF-21)", h2),
    P("<b>Como</b> Administrador, <b>quiero</b> consultar la bitácora inalterable de auditoría, <b>para</b> mantener la trazabilidad de las operaciones y responder ante reclamaciones."),
    P("• <b>CA-10.1 Registro automático:</b> Toda modificación administrativa inserta una fila en auditoria con marca de tiempo UTC y actor.<br/>"
      "• <b>CA-10.2 Inmutabilidad por trigger:</b> Cualquier sentencia UPDATE o DELETE sobre auditoria es abortada por la base de datos con excepción irrevocable."),
    PageBreak()
]

# ==============================================================================
# ENTREGABLE 4: MOCKUPS DE LAS HISTORIAS DE GESTIÓN
# ==============================================================================
E += [
    P("5. Mockups de Interfaz de Usuario de las Historias de Gestión", h1),
    P("Los mockups fueron diseñados manteniendo la línea visual corporativa del parqueadero Autos Colombia: paleta sobria (Azul #0F4C81, Asfalto #1C2530, fondo #F8FAFC), "
      "tipografía clara y jerarquizada, tarjetas de contenido, badges de estado con alto contraste y formularios modales accesibles."),
    figura(os.path.join(AQUI, "pantallas", "11_admin_usuarios.png"),
           "Figura 2. Mockup de Gestión de Usuarios del Sistema (Listado RBAC, estados, filtros y protección del último administrador)", ancho=16.0 * cm, alto=9.5 * cm),
    Spacer(1, 4),
    figura(os.path.join(AQUI, "pantallas", "12_modal_usuario.png"),
           "Figura 3. Mockup de Creación y Edición de Usuario (Formulario modal con indicador de fortaleza de contraseña RN-07 y asignación de rol)", ancho=13.0 * cm, alto=9.5 * cm),
    PageBreak(),
    figura(os.path.join(AQUI, "pantallas", "14_admin_celdas.png"),
           "Figura 4. Mockup de Gestión de Celdas y Bahías (Resumen de capacidad, tipos CARRO/MOTO y conmutador de inhabilitación por mantenimiento RN-09)", ancho=16.0 * cm, alto=9.5 * cm),
    Spacer(1, 4),
    figura(os.path.join(AQUI, "pantallas", "15_admin_auditoria.png"),
           "Figura 5. Mockup de Bitácora de Auditoría Inmutable (Consulta cronológica de eventos administrativos protegida por trigger RN-11)", ancho=16.0 * cm, alto=9.5 * cm),
    PageBreak()
]

# ==============================================================================
# ENTREGABLE 5: DIAGRAMAS DE SECUENCIA (ITERACIONES 1 Y 2)
# ==============================================================================
E += [
    P("6. Diagramas de Secuencia de las Iteraciones 1 y 2 en STARUML", h1),
    P("Los diagramas de secuencia modelan la interacción temporal y el intercambio de mensajes entre los componentes del sistema (Frontend, Capa HTTP, Servicios de Aplicación, Reglas de Dominio y Base de Datos). "
      "Se incluyen los diagramas representativos tanto de la <b>Iteración 1</b> (ingreso y salida) como de la <b>Iteración 2</b> (login con PBKDF2 y gestión de celdas con auditoría)."),
    figura(os.path.join(AQUI, "diagramas", "secuencia_01_ingreso.png"),
           "Figura 6. Diagrama de Secuencia 1 — Registro de Ingreso de Vehículo (Iteración 1: validaciones de mensualidad, celda disponible y trigger de tipos)", ancho=16.0 * cm, alto=9.0 * cm),
    Spacer(1, 4),
    figura(os.path.join(AQUI, "diagramas", "secuencia_02_salida.png"),
           "Figura 7. Diagrama de Secuencia 2 — Registro de Salida de Vehículo (Iteración 1: cálculo de permanencia y liberación de celda RN-04)", ancho=16.0 * cm, alto=9.0 * cm),
    PageBreak(),
    figura(os.path.join(AQUI, "diagramas", "secuencia_03_login.png"),
           "Figura 8. Diagrama de Secuencia 3 — Autenticación y Emisión de Sesión (Iteración 2: rate limiting, hash PBKDF2 con sal y emisión de token Bearer RN-07)", ancho=16.0 * cm, alto=9.0 * cm),
    Spacer(1, 4),
    figura(os.path.join(AQUI, "diagramas", "secuencia_04_gestion_celdas.png"),
           "Figura 9. Diagrama de Secuencia 4 — Inhabilitación de Celda con Auditoría (Iteración 2: validación de rol, guardia de celda ocupada RN-09 y persistencia inmutable RN-11)", ancho=16.0 * cm, alto=9.0 * cm),
    PageBreak()
]

# ==============================================================================
# ENTREGABLE 6: DIAGRAMA DE COMPONENTES (ITERACIONES 1 Y 2)
# ==============================================================================
E += [
    P("7. Diagrama de Componentes de las Iteraciones 1 y 2 en STARUML", h1),
    P("El diagrama de componentes describe la organización modular del software y las dependencias entre subsistemas. "
      "La solución implementa una <b>Arquitectura Hexagonal (Puertos y Adaptadores)</b> estricta, lo que garantiza que las reglas de negocio permanezcan completamente desacopladas de los protocolos de transporte y de la base de datos."),
    figura(os.path.join(AQUI, "diagramas", "diagrama_componentes.png"),
           "Figura 10. Diagrama de Componentes del Sistema Autos Colombia (Iteraciones 1 y 2: Frontend PWA, Backend Go Hexagonal y PostgreSQL 16)", ancho=16.0 * cm, alto=10.5 * cm),
    P("Descripción de Módulos y Capas", h2),
    P("• <b>Frontend (PWA React + Vite):</b> Capa de presentación para dispositivos móviles y de escritorio. Contiene vistas modulares y el cliente HTTP (<font name='Arial-B'>api.js</font>) que administra las cabeceras de autorización Bearer.<br/>"
      "• <b>Adaptadores Primarios (HTTP API):</b> Router nativo Go 1.22 con validación de métodos (405) y middleware de autenticación, control de tasa (anti fuerza bruta) y cabeceras de seguridad.<br/>"
      "• <b>Capa de Aplicación:</b> Casos de uso (<font name='Arial-B'>ServicioIngresoSalida</font> y <font name='Arial-B'>ServicioGestion</font>) que orquestan el negocio sin conocer la base de datos concreta.<br/>"
      "• <b>Dominio:</b> Entidades puras y validaciones invariantes (placas colombianas, cálculo de vigencia UTC-5, política de contraseñas, validación de códigos de celda).<br/>"
      "• <b>Seguridad Criptográfica:</b> Hasheador PBKDF2 (HMAC-SHA256, 16 bytes de sal aleatoria, 100.000 iteraciones) y módulo de firma de sesiones.<br/>"
      "• <b>Adaptadores Secundarios (Persistencia):</b> Repositorio PostgreSQL transaccional y Repositorio en Memoria para pruebas unitarias de alta velocidad.<br/>"
      "• <b>Motor de Base de Datos:</b> PostgreSQL 16 con disparadores <font name='Arial-B'>trg_registro_tipos</font> (impide asociar carro en celda moto a nivel físico) y <font name='Arial-B'>trg_auditoria_inmutable</font> (impide alterar la bitácora)."),
    PageBreak()
]

# ==============================================================================
# ENTREGABLE 7: PLAN DE PRUEBAS RESUMIDO
# ==============================================================================
E += [
    P("8. Formato de Plan de Pruebas Resumido", h1),
    P("El aseguramiento de la calidad del software se fundamenta en pruebas automatizadas y verificaciones de integración. "
      "A continuación se presenta la matriz formal del <b>Plan de Pruebas Resumido</b> de la Iteración 2, cubriendo los escenarios de éxito, condiciones de borde y rechazos de negocio."),
    tabla([
        ["Id", "Módulo", "Caso de Prueba", "Precondiciones", "Datos de Entrada", "Resultado Esperado", "Resultado Obtenido", "Estado"],
        ["CP-01", "Seguridad", "Autenticación válida de usuario y emisión de sesión", "Usuario registrado y activo en BD.", "usuario: 'admin', clave: 'ClaveSegura123'", "HTTP 200 OK con token Bearer firmado y expiración.", "Token emitido; reinicio de intentos.", "Aprobado"],
        ["CP-02", "Seguridad", "Rechazo por credenciales erróneas", "Usuario registrado o inexistente.", "usuario: 'admin', clave: 'Invalida'", "HTTP 401 Unauthorized ('usuario o contraseña incorrectos').", "HTTP 401; sin filtrar motivo.", "Aprobado"],
        ["CP-03", "Seguridad", "Bloqueo por fuerza bruta tras 5 fallos", "IP sin bloqueo previo.", "6 intentos fallidos consecutivos en < 60s", "HTTP 429 Too Many Requests con cabecera Retry-After: 60.", "HTTP 429 en el 6to intento; IP bloqueada.", "Aprobado"],
        ["CP-04", "Usuarios", "Registro exitoso de nuevo usuario por Admin", "Admin autenticado con token Bearer.", "usuario: 'oper1', nombre: 'Carlos Ruiz', rol: 'ENCARGADO'", "HTTP 201 Created. Usuario activo; auditoría generada.", "Usuario persistido; hash PBKDF2 guardado.", "Aprobado"],
        ["CP-05", "Usuarios", "Rechazo por nombre de usuario duplicado", "Existe usuario con login 'admin'.", "usuario: 'admin', rol: 'ADMINISTRADOR'", "HTTP 409 Conflict ('ya existe un usuario con ese nombre').", "HTTP 409; unicidad garantizada.", "Aprobado"],
        ["CP-06", "Usuarios", "Rechazo por contraseña débil o inválida", "Admin autenticado.", "clave: 'corta1' (< 10 caracteres)", "HTTP 400 Bad Request explicando directivas de clave.", "HTTP 400 con detalle de la política.", "Aprobado"],
        ["CP-07", "Usuarios", "Protección del último administrador activo", "Solo 1 administrador activo.", "Admin intenta inactivar su cuenta o degradar rol.", "HTTP 409 Conflict ('debe quedar al menos un admin activo').", "Error atómico ErrUltimoAdministrador.", "Aprobado"],
        ["CP-08", "Usuarios", "Cambio personal de contraseña con clave actual", "Usuario autenticado.", "clave_actual: 'Vieja123', clave_nueva: 'Nueva2026'", "HTTP 200 OK. Contraseña actualizada y auditada.", "Hash actualizado; sesión renovada.", "Aprobado"],
        ["CP-09", "Celdas", "Registro exitoso de nueva celda", "Admin autenticado.", "codigo: 'A-05', tipo: 'CARRO'", "HTTP 201 Created. Celda en estado DISPONIBLE.", "Celda creada en mayúsculas.", "Aprobado"],
        ["CP-10", "Celdas", "Rechazo por código de celda duplicado", "Existe celda con código 'A-01'.", "codigo: 'a-01', tipo: 'CARRO'", "HTTP 409 Conflict ('ya existe una celda con ese código').", "Código normalizado a mayúsculas y rechazado.", "Aprobado"],
        ["CP-11", "Celdas", "Inhabilitación de celda por mantenimiento", "Celda DISPONIBLE sin vehículos.", "estado: 'INACTIVA'", "HTTP 200 OK. Celda pasa a INACTIVA y se excluye de ingresos.", "Celda inhabilitada y auditada.", "Aprobado"],
        ["CP-12", "Celdas", "Bloqueo de inhabilitación en celda ocupada", "Celda con vehículo en estancia.", "id_celda: 2 (ocupada), estado: 'INACTIVA'", "HTTP 409 Conflict ('la celda está ocupada; registre salida').", "Rechazo inmediato; celda protegida.", "Aprobado"],
        ["CP-13", "Clientes", "Rechazo de mensualidad solapada", "Cliente con mensualidad del 01/11 al 30/11.", "periodo: 2026-11-15 al 2026-12-15", "HTTP 409 Conflict ('la mensualidad se solapa').", "Validación de fechas ejecutada.", "Aprobado"],
        ["CP-14", "Auditoría", "Inmutabilidad de la tabla de auditoría", "Filas existentes en tabla auditoria.", "Sentencia SQL UPDATE o DELETE directa en motor.", "Trigger trg_auditoria_inmutable aborta la transacción.", "Excepción lanzada; registros intactos.", "Aprobado"]
    ], [1.3 * cm, 1.8 * cm, 2.5 * cm, 2.1 * cm, 2.1 * cm, 2.6 * cm, 2.3 * cm, 1.7 * cm]),
    Spacer(1, 8),
    P("Evidencia de Ejecución Automatizada en el Backend", h2),
    P("La totalidad de los casos de prueba de dominio, aplicación, seguridad y adaptadores HTTP y PostgreSQL se ejecutan mediante el comando "
      "<font name='Arial-B'>go test -race ./...</font>, obteniendo un resultado de <b>0 fallos y 100 % de aprobación</b> en todas las suites."),
    PageBreak()
]

# ==============================================================================
# ENTREGABLE 8: REPOSITORIO GITHUB Y TRAZABILIDAD
# ==============================================================================
E += [
    P("9. Repositorio en GitHub y Evidencias de Código Fuente", h1),
    P("El código fuente completo del sistema, con la totalidad de los módulos de la Iteración 1 y la Iteración 2, se encuentra versionado bajo Git y alojado en GitHub en el siguiente enlace público:"),
    Spacer(1, 4),
    P(f"<b>URL Oficial del Repositorio:</b> <font color='#0f4c81'><u>{GITHUB}</u></font>", body_bold),
    Spacer(1, 4),
    P("Resumen de Actualizaciones de Código (Iteración 2)", h2),
    P("En el commit <font name='Arial-B'>c928179</font> se incorporaron más de 4.100 líneas de código fuente en Go y sentencias DDL en PostgreSQL, distribuidas en los siguientes componentes:"),
    tabla([
        ["Paquete / Archivo", "Propósito en la Iteración 2", "Líneas Agregadas"],
        ["backend/internal/domain/gestion.go", "Entidades Usuario, Rol, ClienteDetalle, Auditoria y reglas de validación (clave robusta, login normalizado, celdas gestionables).", "+135"],
        ["backend/internal/application/gestion.go", "Casos de uso de autenticación, control RBAC, gestión de usuarios (guardia de último admin), celdas (guardia de ocupación) y auditoría.", "+421"],
        ["backend/internal/adapters/seguridad/pbkdf2.go", "Adaptador criptográfico PBKDF2 con sal aleatoria de 16 bytes y comparación en tiempo constante (subtle.ConstantTimeCompare).", "+91"],
        ["backend/internal/adapters/httpapi/gestion.go", "Endpoints REST para /api/login, /api/sesion, /api/usuarios, /api/celdas, /api/clientes y /api/auditoria.", "+311"],
        ["backend/internal/adapters/httpapi/sesion.go", "Emisión y verificación de tokens de sesión con revocación inmediata al inactivar usuario.", "+66"],
        ["backend/internal/adapters/postgres/gestion.go", "Consultas SQL parametrizadas, transacciones ACID y bloqueos pesimistas para PostgreSQL.", "+376"],
        ["backend/db/schema.sql", "Tablas usuario y auditoria; triggers trg_auditoria_inmutable y trg_registro_tipos.", "+58"],
        ["Suites de Pruebas (*_test.go)", "Pruebas unitarias y de integración para gestión, seguridad, handlers HTTP y PostgreSQL real.", "+2.600+"]
    ], [4.5 * cm, W - 6.5 * cm, 2.0 * cm]),
    Spacer(1, 10),
    P("10. Matriz de Trazabilidad entre Representaciones", h1),
    P("La trazabilidad garantiza que cada requerimiento funcional esté respaldado por una historia de usuario, un caso de uso, componentes de software, tablas en la base de datos, endpoints de API y casos de prueba verificados."),
    tabla([
        ["Requisito", "Historia", "Caso de Uso", "Componentes / Clases", "Tablas en BD", "Endpoint API", "Prueba"],
        ["RF-08, RF-09", "HU-07", "CU-06", "Router, MiddlewareAuth, SvcGestion, PBKDF2", "usuario", "POST /api/login", "CP-01, CP-02, CP-03"],
        ["RF-10, RF-11", "HU-06", "CU-08", "CtrlGestion, SvcGestion, Dominio, RepoPG", "usuario, auditoria", "GET, POST, PATCH /api/usuarios", "CP-04, CP-05, CP-07"],
        ["RF-12, RF-13", "HU-07", "CU-07", "CtrlGestion, SvcGestion, PBKDF2", "usuario, auditoria", "POST /api/sesion/clave", "CP-06, CP-08"],
        ["RF-15, RF-16, RF-17", "HU-08", "CU-09", "CtrlGestion, SvcGestion, RepoPG", "celda, auditoria", "GET, POST, PUT /api/celdas", "CP-09, CP-10, CP-11, CP-12"],
        ["RF-18, RF-19", "HU-09", "CU-10", "CtrlGestion, SvcGestion, Dominio", "cliente, vehiculo, mensualidad", "GET, POST /api/clientes", "CP-13"],
        ["RF-20, RF-21", "HU-10", "CU-11", "CtrlGestion, SvcGestion, Triggers", "auditoria (trigger inmutable)", "GET /api/auditoria", "CP-14"]
    ], [1.7 * cm, 1.4 * cm, 1.4 * cm, 3.8 * cm, 2.8 * cm, 3.0 * cm, W - 14.1 * cm]),
    Spacer(1, 10),
    P("11. Conclusiones", h1),
    P("1. <b>Evolución hacia la madurez arquitectónica:</b> La Iteración 2 permitió transformar un sistema monousuario básico en una plataforma empresarial robusta dotada de gobernabilidad, seguridad y trazabilidad. La implementación del control de acceso basado en roles (RBAC) garantiza que las acciones críticas de configuración del patio y administración de identidades queden reservadas al Administrador, protegiendo la operación."),
    P("2. <b>Seguridad integral y no repudio:</b> La combinación de contraseñas protegidas mediante PBKDF2 con sal individual, comparaciones en tiempo constante, protección perimetral contra fuerza bruta (Rate Limiting) y una bitácora de auditoría inmutable mediante triggers transaccionales en PostgreSQL provee un estándar de seguridad de nivel de producción que cumple plenamente con los postulados de la ingeniería de software moderna."),
    P("3. <b>Integridad del modelo de negocio en persistencia:</b> Reglas de negocio críticas como la imposibilidad de inactivar al último administrador activo, la prohibición de inhabilitar celdas ocupadas y la compatibilidad estricta de tipos celda-vehículo fueron blindadas tanto en la lógica de aplicación como a nivel de restricciones de base de datos, eliminando la posibilidad de estados corruptos ante accesos concurrentes."),
    P("4. <b>Coherencia y trazabilidad absoluta:</b> Se demostró una alineación perfecta entre las necesidades de negocio, los requerimientos formales, los modelos UML en StarUML (casos de uso, secuencias y componentes), las interfaces de usuario (mockups en Figma) y las pruebas automatizadas (100 % aprobadas), garantizando un ciclo de desarrollo ágil y riguroso.")
]

doc.build(E)
print(f"Documento PDF generado exitosamente: {SALIDA}")
