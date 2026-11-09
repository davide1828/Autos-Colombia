# -*- coding: utf-8 -*-
"""
Generador del Documento Word (.docx) — Evidencia de Aprendizaje 2 (EA2)
Sistema de Información Parqueadero "Autos Colombia"
Estudiante: Cristian David España Figueroa
"""
import os, json
from docx import Document
from docx.shared import Inches, Pt, RGBColor
from docx.enum.text import WD_ALIGN_PARAGRAPH
from docx.enum.table import WD_TABLE_ALIGNMENT
from docx.oxml import OxmlElement, parse_xml
from docx.oxml.ns import nsdecls, qn

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
SALIDA = os.path.join(AQUI, "España_Cristian_EA2.docx")

doc = Document()

# Configuración de márgenes
for section in doc.sections:
    section.top_margin = Inches(0.85)
    section.bottom_margin = Inches(0.85)
    section.left_margin = Inches(0.85)
    section.right_margin = Inches(0.85)

# Colores institucionales
COLOR_AZUL = RGBColor(15, 76, 129)     # #0F4C81
COLOR_TEXTO = RGBColor(28, 37, 48)     # #1C2530
COLOR_MUTED = RGBColor(100, 116, 139)  # #64748B

def set_run_font(run, name="Arial", size_pt=10, bold=False, italic=False, color=COLOR_TEXTO):
    run.font.name = name
    run.font.size = Pt(size_pt)
    run.font.bold = bold
    run.font.italic = italic
    run.font.color.rgb = color

def add_p(doc, text, align=WD_ALIGN_PARAGRAPH.LEFT, space_before=0, space_after=6, font_size=10, bold=False, italic=False, color=COLOR_TEXTO):
    p = doc.add_paragraph()
    p.alignment = align
    p.paragraph_format.space_before = Pt(space_before)
    p.paragraph_format.space_after = Pt(space_after)
    p.paragraph_format.line_spacing = 1.15
    run = p.add_run(text)
    set_run_font(run, size_pt=font_size, bold=bold, italic=italic, color=color)
    return p

def add_h1(doc, text):
    return add_p(doc, text, space_before=14, space_after=6, font_size=15, bold=True, color=COLOR_AZUL)

def add_h2(doc, text):
    return add_p(doc, text, space_before=10, space_after=4, font_size=12, bold=True, color=COLOR_AZUL)

def add_table(doc, rows_data, col_widths, is_header=True):
    t = doc.add_table(rows=len(rows_data), cols=len(rows_data[0]))
    t.alignment = WD_TABLE_ALIGNMENT.CENTER
    for i, row in enumerate(rows_data):
        for j, val in enumerate(row):
            cell = t.cell(i, j)
            cell.width = Inches(col_widths[j])
            p = cell.paragraphs[0]
            p.paragraph_format.space_before = Pt(2)
            p.paragraph_format.space_after = Pt(2)
            p.paragraph_format.line_spacing = 1.1
            run = p.add_run(str(val))
            if is_header and i == 0:
                set_run_font(run, size_pt=9, bold=True, color=RGBColor(255, 255, 255))
                # Fondo azul encabezado
                shading = parse_xml(r'<w:shd {} w:fill="0F4C81"/>'.format(nsdecls('w')))
                cell._tc.get_or_add_tcPr().append(shading)
            else:
                set_run_font(run, size_pt=8.5, color=COLOR_TEXTO)
                if i % 2 == 0:
                    shading = parse_xml(r'<w:shd {} w:fill="F8FAFC"/>'.format(nsdecls('w')))
                    cell._tc.get_or_add_tcPr().append(shading)
    # Espacio tras tabla
    p_after = doc.add_paragraph()
    p_after.paragraph_format.space_before = Pt(0)
    p_after.paragraph_format.space_after = Pt(6)
    return t

def add_img(doc, path, width_in=6.2, caption=""):
    if os.path.exists(path):
        p = doc.add_paragraph()
        p.alignment = WD_ALIGN_PARAGRAPH.CENTER
        p.paragraph_format.space_before = Pt(6)
        p.paragraph_format.space_after = Pt(2)
        p.add_run().add_picture(path, width=Inches(width_in))
        if caption:
            p_cap = doc.add_paragraph()
            p_cap.alignment = WD_ALIGN_PARAGRAPH.CENTER
            p_cap.paragraph_format.space_before = Pt(2)
            p_cap.paragraph_format.space_after = Pt(8)
            r = p_cap.add_run(caption)
            set_run_font(r, size_pt=8.5, italic=True, color=COLOR_MUTED)

# ==============================================================================
# 1. PORTADA
# ==============================================================================
add_p(doc, "S20 · EVIDENCIA DE APRENDIZAJE 2", space_before=30, space_after=4, font_size=11, bold=True, color=COLOR_AZUL)
add_p(doc, "Sistema de información del parqueadero “Autos Colombia”", space_before=4, space_after=8, font_size=20, bold=True, color=COLOR_AZUL)
add_p(doc, "Diseño de Sistema de Información – Iteración 2: Gestión de usuarios y gestión de celdas", space_before=4, space_after=35, font_size=12, italic=True, color=COLOR_MUTED)

tabla_portada = [
    ["Estudiante", NOMBRE],
    ["Programa", "Tecnología en Desarrollo de Software"],
    ["Institución", "Institución Universitaria Digital de Antioquia (IU Digital)"],
    ["Asignatura", MATERIA],
    ["Docente", DOCENTE],
    ["Código del curso", CODIGO],
    ["Periodo", PERIODO],
    ["Repositorio GitHub", GITHUB],
    ["Diseño en Figma", FIGMA],
    ["Ciudad y fecha", "Medellín, Colombia · Noviembre de 2026"]
]
add_table(doc, tabla_portada, [2.2, 4.3], is_header=False)
doc.add_page_break()

# ==============================================================================
# 2. CONTENIDO E INTRODUCCIÓN
# ==============================================================================
add_h1(doc, "Contenido")
add_p(doc, "1. Introducción y Enfoque de la Iteración 2\n"
           "2. Requerimientos Funcionales y No Funcionales (Gestión de Usuarios y Celdas)\n"
           "3. Diagramas de Casos de Uso en STARUML\n"
           "4. Historias de Usuario Asociadas (Formato Ágil y Gherkin)\n"
           "5. Mockups de Interfaz de Usuario de las Historias de Gestión\n"
           "6. Diagramas de Secuencia de las Iteraciones 1 y 2 en STARUML\n"
           "7. Diagrama de Componentes de las Iteraciones 1 y 2 en STARUML\n"
           "8. Formato de Plan de Pruebas Resumido (Gestión de Usuarios y Celdas)\n"
           "9. Repositorio en GitHub y Evidencias de Código Fuente\n"
           "10. Matriz de Trazabilidad entre Representaciones\n"
           "11. Conclusiones")

add_h1(doc, "1. Introducción y Enfoque de la Iteración 2")
add_p(doc, "El sistema de información del parqueadero “Autos Colombia” es una solución integral diseñada para modernizar y controlar de forma eficiente las operaciones de estacionamiento por mensualidad. "
           "Tras completar con éxito en la Iteración 1 el flujo operativo del patio (control de accesos, cálculo de permanencia, bahías de estacionamiento y registro de novedades), "
           "esta Iteración 2 se enfoca en dotar a la plataforma de gobernabilidad, seguridad criptográfica y gestión de inventario de celdas mediante un modelo multiusuario con control de acceso basado en roles (RBAC).")
add_p(doc, "El sistema incorpora la administración de identidades para diferenciar las atribuciones del Administrador y del Encargado, políticas estrictas de contraseñas mediante hashing PBKDF2-SHA256 con sal aleatoria, "
           "gestión dinámica del estado operativo de las celdas (habilitación e inhabilitación preventiva por mantenimiento) y una bitácora de auditoría inmutable forzada por disparadores transaccionales a nivel de PostgreSQL.")
doc.add_page_break()

# ==============================================================================
# 3. REQUERIMIENTOS FUNCIONALES Y NO FUNCIONALES
# ==============================================================================
add_h1(doc, "2. Requerimientos Funcionales y No Funcionales")
add_p(doc, "Los requerimientos de la Iteración 2 formalizan las capacidades del sistema para la gestión de usuarios, celdas y auditoría, articuladas con las siguientes reglas de negocio:")
add_p(doc, "• RN-07 (Autenticación y Sesión Segura): Acceso mediante credenciales individuales, hash PBKDF2-SHA256 (100.000 iteraciones) y bloqueo preventivo por IP tras 5 fallos consecutivos.\n"
           "• RN-08 (Gobernabilidad y Último Administrador): El sistema debe mantener siempre al menos un usuario activo con rol ADMINISTRADOR; se prohíbe inactivar o degradar al último administrador.\n"
           "• RN-09 (Integridad de Celdas Ocupadas): Una celda ocupada no puede pasar a INACTIVA ni ser modificada en su código o tipo mientras aloje un vehículo.\n"
           "• RN-10 (No Solapamiento de Mensualidades): Ningún cliente puede tener periodos de mensualidad superpuestos.\n"
           "• RN-11 (Auditoría Inmutable): Toda modificación administrativa genera un registro de auditoría que no puede ser alterado ni eliminado.")

add_h2(doc, "Tabla de Requerimientos Funcionales (RF-08 a RF-21)")
rf_data = [
    ["Id", "Requerimiento Funcional", "Regla", "Prioridad"],
    ["RF-08", "Autenticación de usuarios con emisión de token de sesión seguro (Bearer/JWT) con expiración y perfil.", "RN-07", "Alta"],
    ["RF-09", "Control de acceso basado en roles (RBAC): ADMIN gestiona usuarios, celdas y auditoría; ENCARGADO opera patio y cambia clave.", "RN-07", "Alta"],
    ["RF-10", "Registro de usuarios del sistema con login único (3-30 car), nombre, rol y contraseña inicial segura.", "RN-07", "Alta"],
    ["RF-11", "Modificación y estado de usuarios (activo/inactivo), impidiendo la inactivación o degradación del último admin.", "RN-08", "Alta"],
    ["RF-12", "Cambio de contraseña por el propio usuario validando su clave actual y la política de robustez.", "RN-07", "Media"],
    ["RF-13", "Restablecimiento administrativo de credenciales para cualquier operador sin conocer la anterior.", "RN-07", "Media"],
    ["RF-14", "Revocación inmediata de sesión al inactivar a un usuario o cambiarle el rol.", "RN-07, RN-08", "Alta"],
    ["RF-15", "Registro de celdas con código alfanumérico único en mayúsculas (ej. A-01) y tipo CARRO/MOTO.", "RN-03", "Alta"],
    ["RF-16", "Inhabilitación y reactivación de celdas (DISPONIBLE / INACTIVA), impidiendo inhabilitar si está ocupada.", "RN-09", "Alta"],
    ["RF-17", "Modificación de características de celda (código/tipo) validando que no tenga vehículo dentro.", "RN-03, RN-09", "Media"],
    ["RF-18", "Gestión de clientes y mensualidades validando que no existan periodos solapados para el mismo cliente.", "RN-01, RN-10", "Alta"],
    ["RF-19", "Actualización de tipo de vehículo validando que no tenga ingreso activo en el patio.", "RN-03", "Media"],
    ["RF-20", "Registro automático en bitácora de auditoría de cada modificación sobre usuarios, celdas y vehículos.", "RN-11", "Alta"],
    ["RF-21", "Consulta de la bitácora de auditoría en orden cronológico descendente.", "RN-11", "Media"],
]
add_table(doc, rf_data, [0.8, 3.8, 1.1, 0.8])

add_h2(doc, "Tabla de Requerimientos No Funcionales (RNF-08 a RNF-13)")
rnf_data = [
    ["Id", "Categoría", "Requerimiento No Funcional Verificable", "Criterio de Aceptación"],
    ["RNF-08", "Seguridad Criptográfica", "Contraseñas protegidas mediante PBKDF2 (HMAC-SHA256, 16 bytes de sal, 100k iteraciones) y verificación en tiempo constante.", "100 % pruebas unitarias superadas en seguridad; uso de subtle.ConstantTimeCompare."],
    ["RNF-09", "Anti Fuerza Bruta", "Limitación por IP (máximo 5 fallos/minuto); retorno de HTTP 429 con cabecera Retry-After: 60.", "Validado en pruebas de HTTP API con RateLimiter."],
    ["RNF-10", "Integridad Concurrente", "Transacciones ACID con bloqueos pesimistas (FOR SHARE / FOR UPDATE) en operaciones concurrentes sobre celdas y conteo de administradores.", "Sin condiciones de carrera bajo pruebas concurrentes en PostgreSQL."],
    ["RNF-11", "Inviolabilidad", "Tabla de auditoría inmutable a nivel de motor de base de datos.", "Trigger trg_auditoria_inmutable arroja excepción ante UPDATE o DELETE."],
    ["RNF-12", "Rendimiento Auth", "Validación de credenciales y extracción de sesión en menos de 5 ms.", "Medido en peticiones HTTP Go."],
    ["RNF-13", "Mantenibilidad", "Arquitectura hexagonal desacoplada en Go con doble implementación de persistencia (PostgreSQL y Memoria).", "Casos de uso ejecutables sin base de datos externa para pruebas ágiles."]
]
add_table(doc, rnf_data, [0.8, 1.5, 2.5, 1.7])
doc.add_page_break()

# ==============================================================================
# 4. CASOS DE USO EN STARUML
# ==============================================================================
add_h1(doc, "3. Diagramas de Casos de Uso en STARUML")
add_p(doc, "El modelo fue construido en StarUML (archivo docs/diagramas/parqueadero_iteracion_2.mdj). Representa la interacción de los actores Administrador y Encargado con los casos de uso de gestión.")
add_img(doc, os.path.join(AQUI, "diagramas", "casos_uso_iteracion_2.png"), width_in=6.4,
        caption="Figura 1. Diagrama de Casos de Uso — Iteración 2 (StarUML)")

cu_data = [
    ["Caso de Uso", "Actores", "Descripción", "Precondición", "Inclusiones <<include>>"],
    ["CU-06 Iniciar sesión", "Admin, Encargado", "Autentica al usuario en el sistema y emite token Bearer.", "Usuario registrado y activo.", "Validar credenciales y tasa."],
    ["CU-07 Cambiar contraseña", "Admin, Encargado", "Actualiza la clave personal confirmando la actual.", "Sesión activa.", "Validar robustez contraseña."],
    ["CU-08 Administrar usuarios", "Administrador", "Alta, edición, inactivación y reseteo de claves.", "Admin autenticado.", "Validar clave; Garantizar último admin; Registrar auditoría."],
    ["CU-09 Administrar celdas", "Administrador", "Alta de celdas, edición e inhabilitación preventiva.", "Admin autenticado.", "Validar celda no ocupada; Registrar auditoría."],
    ["CU-10 Administrar clientes", "Admin, Encargado", "Clientes, vehículos y mensualidades sin cruces.", "Usuario autenticado / Admin.", "Validar no solapamiento; Validar tipo vehículo; Registrar auditoría."],
    ["CU-11 Consultar auditoría", "Administrador", "Consulta cronológica de eventos de auditoría.", "Admin autenticado.", "Ninguna (consulta pura)."]
]
add_table(doc, cu_data, [1.4, 1.1, 1.6, 1.1, 1.3])
doc.add_page_break()

# ==============================================================================
# 5. HISTORIAS DE USUARIO
# ==============================================================================
add_h1(doc, "4. Historias de Usuario Asociadas (Formato Ágil y Gherkin)")
add_p(doc, "A continuación se presentan las Historias de Usuario (HU-06 a HU-10) con sus respectivos criterios de aceptación estructurados en formato Dado / Cuando / Entonces.")

add_h2(doc, "HU-06 · Gestión de usuarios del sistema (Rol: Administrador)")
add_p(doc, "Como Administrador del parqueadero, quiero registrar, modificar y administrar las cuentas de los usuarios asignando roles (ADMINISTRADOR / ENCARGADO), para controlar el acceso a la plataforma.")
add_p(doc, "• CA-06.1 Registro válido: Dado que el Admin diligencia login único, nombre válido, clave robusta y rol, Cuando solicita la creación, Entonces el sistema crea el usuario activo, almacena el hash PBKDF2, registra auditoría y retorna HTTP 201 Created.\n"
           "• CA-06.2 Login duplicado: Dado que ya existe un usuario con el mismo login, Cuando se intenta registrar, Entonces rechaza con HTTP 409 Conflict.\n"
           "• CA-06.3 Contraseña débil: Dado una clave < 10 caracteres o sin letras/números, Cuando se envía, Entonces responde HTTP 400 Bad Request.\n"
           "• CA-06.4 Inactivación de usuario: Dado un Encargado activo, Cuando el Admin cambia su estado a inactivo, Entonces el sistema persiste el cambio y cualquier petición posterior con su token retorna HTTP 401.\n"
           "• CA-06.5 Protección del último administrador: Dado un único admin activo, Cuando se intenta inactivar o degradar su rol, Entonces bloquea con HTTP 409 Conflict ('debe quedar al menos un administrador activo').")

add_h2(doc, "HU-07 · Autenticación y cambio de credenciales (Rol: Administrador / Encargado)")
add_p(doc, "Como usuario del sistema, quiero iniciar sesión de forma segura y actualizar mi contraseña personal, para acceder a mis funciones y proteger mi cuenta.")
add_p(doc, "• CA-07.1 Login exitoso: Dado un usuario activo y clave correcta, Cuando envía POST a /api/login, Entonces recibe HTTP 200 OK con token Bearer, expiración y perfil.\n"
           "• CA-07.2 Credenciales erróneas: Dado credenciales incorrectas, Cuando solicita login, Entonces recibe HTTP 401 Unauthorized y suma 1 fallo a la IP.\n"
           "• CA-07.3 Bloqueo anti fuerza bruta: Dado 5 fallos consecutivos en < 60s, Cuando reintenta, Entonces recibe HTTP 429 Too Many Requests (Retry-After: 60).\n"
           "• CA-07.4 Cambio de contraseña: Dado un usuario autenticado que ingresa su clave actual correcta y nueva clave robusta, Cuando confirma, Entonces actualiza el hash PBKDF2 y retorna HTTP 200 OK.")

add_h2(doc, "HU-08 · Gestión y mantenimiento de celdas (Rol: Administrador)")
add_p(doc, "Como Administrador, quiero registrar celdas y habilitar o inhabilitar bahías por obras o mantenimiento, para gestionar la capacidad real del patio.")
add_p(doc, "• CA-08.1 Registro de celda: Dado código único y tipo CARRO/MOTO, Cuando se registra, Entonces se guarda en estado DISPONIBLE y retorna HTTP 201 Created.\n"
           "• CA-08.2 Inhabilitación por mantenimiento: Dado una celda DISPONIBLE, Cuando el Admin la pasa a INACTIVA, Entonces se excluye de los selectores de ingreso.\n"
           "• CA-08.3 Bloqueo en celda ocupada: Dado una celda OCUPADA con vehículo dentro, Cuando se intenta inhabilitar o editar tipo, Entonces rechaza con HTTP 409 Conflict ('la celda está ocupada; registre primero la salida del vehículo').")

add_h2(doc, "HU-09 · Gestión de clientes y mensualidades (Rol: Administrador / Encargado)")
add_p(doc, "Como operador del parqueadero, quiero registrar clientes, vehículos y mensualidades, para autorizar el ingreso a vehículos con contrato al día.")
add_p(doc, "• CA-09.1 Mensualidad válida: Dado cliente y rango de fechas no solapado, Cuando se agrega, Entonces se crea la mensualidad y actualiza vigencia.\n"
           "• CA-09.2 Rechazo por solapamiento: Dado cliente con mensualidad vigente, Cuando se intenta registrar periodo que se cruza, Entonces rechaza con HTTP 409 Conflict.\n"
           "• CA-09.3 Cambio de tipo con vehículo dentro: Dado vehículo alojado en patio, Cuando se intenta cambiar su tipo, Entonces bloquea con HTTP 409 Conflict.")

add_h2(doc, "HU-10 · Auditoría inmutable de operaciones (Rol: Administrador)")
add_p(doc, "Como Administrador, quiero consultar la bitácora inalterable de auditoría, para mantener la trazabilidad de las operaciones administrativas.")
add_p(doc, "• CA-10.1 Registro automático: Toda modificación administrativa inserta una fila en auditoria con fecha UTC, usuario y detalle.\n"
           "• CA-10.2 Inmutabilidad por trigger: Cualquier sentencia UPDATE o DELETE sobre auditoria es abortada por la base de datos con excepción irrevocable.")
doc.add_page_break()

# ==============================================================================
# 6. MOCKUPS DE LAS HISTORIAS DE GESTIÓN
# ==============================================================================
add_h1(doc, "5. Mockups de Interfaz de Usuario de las Historias de Gestión")
add_p(doc, "Los mockups reflejan las pantallas correspondientes a la gestión de usuarios, celdas y auditoría, diseñadas con la identidad visual del parqueadero Autos Colombia.")
add_img(doc, os.path.join(AQUI, "pantallas", "11_admin_usuarios.png"), width_in=6.4,
        caption="Figura 2. Mockup de Gestión de Usuarios del Sistema (Listado RBAC, estados, filtros y protección del último administrador)")
add_img(doc, os.path.join(AQUI, "pantallas", "12_modal_usuario.png"), width_in=5.4,
        caption="Figura 3. Mockup de Creación y Edición de Usuario (Formulario modal con indicador de fortaleza de contraseña RN-07)")
doc.add_page_break()
add_img(doc, os.path.join(AQUI, "pantallas", "14_admin_celdas.png"), width_in=6.4,
        caption="Figura 4. Mockup de Gestión de Celdas y Bahías (Resumen de capacidad, tipos CARRO/MOTO y conmutador de inhabilitación por mantenimiento RN-09)")
add_img(doc, os.path.join(AQUI, "pantallas", "15_admin_auditoria.png"), width_in=6.4,
        caption="Figura 5. Mockup de Bitácora de Auditoría Inmutable (Consulta cronológica de eventos administrativos protegida por trigger RN-11)")
doc.add_page_break()

# ==============================================================================
# 7. DIAGRAMAS DE SECUENCIA EN STARUML
# ==============================================================================
add_h1(doc, "6. Diagramas de Secuencia de las Iteraciones 1 y 2 en STARUML")
add_p(doc, "Se modelaron en StarUML las interacciones dinámicas temporales para los flujos principales de ambas iteraciones:")
add_img(doc, os.path.join(AQUI, "diagramas", "secuencia_01_ingreso.png"), width_in=6.4,
        caption="Figura 6. Diagrama de Secuencia 1 — Registro de Ingreso de Vehículo (Iteración 1: validaciones transaccionales y trigger de compatibilidad)")
add_img(doc, os.path.join(AQUI, "diagramas", "secuencia_02_salida.png"), width_in=6.4,
        caption="Figura 7. Diagrama de Secuencia 2 — Registro de Salida de Vehículo (Iteración 1: cálculo de permanencia y liberación de celda RN-04)")
doc.add_page_break()
add_img(doc, os.path.join(AQUI, "diagramas", "secuencia_03_login.png"), width_in=6.4,
        caption="Figura 8. Diagrama de Secuencia 3 — Autenticación y Emisión de Sesión (Iteración 2: rate limiting, hash PBKDF2 y token Bearer RN-07)")
add_img(doc, os.path.join(AQUI, "diagramas", "secuencia_04_gestion_celdas.png"), width_in=6.4,
        caption="Figura 9. Diagrama de Secuencia 4 — Inhabilitación de Celda con Auditoría (Iteración 2: validación de rol, guardia de celda ocupada RN-09 y persistencia inmutable RN-11)")
doc.add_page_break()

# ==============================================================================
# 8. DIAGRAMA DE COMPONENTES EN STARUML
# ==============================================================================
add_h1(doc, "7. Diagrama de Componentes de las Iteraciones 1 y 2 en STARUML")
add_p(doc, "El sistema implementa una Arquitectura Hexagonal estricta que desacopla la lógica de negocio pura (Dominio y Aplicación) de los adaptadores de infraestructura (HTTP REST, Criptografía PBKDF2 y Base de Datos PostgreSQL 16).")
add_img(doc, os.path.join(AQUI, "diagramas", "diagrama_componentes.png"), width_in=6.4,
        caption="Figura 10. Diagrama de Componentes del Sistema Autos Colombia (Iteraciones 1 y 2: Frontend PWA, Backend Go Hexagonal y PostgreSQL 16)")
add_p(doc, "• Frontend (PWA React + Vite): Vistas interactivas y cliente HTTP centralizado (api.js).\n"
           "• Adaptadores Primarios: Router Go 1.22, Middleware de autenticación Bearer y Rate Limiter por IP.\n"
           "• Capa de Aplicación: Servicios de orquestación (ServicioIngresoSalida y ServicioGestion).\n"
           "• Dominio: Entidades y reglas de negocio puras (RN-01 a RN-11) independientes de bases de datos.\n"
           "• Seguridad Criptográfica: Módulo PBKDF2 (HMAC-SHA256, 16 bytes de sal, 100.000 iteraciones).\n"
           "• Adaptadores Secundarios: Repositorio PostgreSQL transaccional y Repositorio en Memoria para testing.\n"
           "• Motor PostgreSQL 16: Tablas relacionales y disparadores trg_registro_tipos y trg_auditoria_inmutable.")
doc.add_page_break()

# ==============================================================================
# 9. PLAN DE PRUEBAS RESUMIDO
# ==============================================================================
add_h1(doc, "8. Formato de Plan de Pruebas Resumido")
add_p(doc, "A continuación se detalla la matriz formal del Plan de Pruebas Resumido de la Iteración 2, validada mediante la ejecución automatizada de pruebas unitarias y de integración.")

test_plan_data = [
    ["Id", "Módulo", "Caso de Prueba", "Precondiciones", "Datos de Entrada", "Resultado Esperado", "Estado"],
    ["CP-01", "Seguridad", "Autenticación válida y emisión de sesión", "Usuario activo en BD.", "usuario: 'admin', clave: 'ClaveSegura123'", "HTTP 200 OK con token Bearer firmado y expiración.", "Aprobado"],
    ["CP-02", "Seguridad", "Rechazo por credenciales erróneas", "Usuario registrado o no.", "usuario: 'admin', clave: 'Invalida'", "HTTP 401 Unauthorized ('usuario o clave incorrectos').", "Aprobado"],
    ["CP-03", "Seguridad", "Bloqueo por fuerza bruta tras 5 fallos", "IP sin bloqueo previo.", "6 fallos consecutivos en < 60s", "HTTP 429 Too Many Requests con Retry-After: 60.", "Aprobado"],
    ["CP-04", "Usuarios", "Registro exitoso de usuario por Admin", "Admin autenticado.", "usuario: 'oper1', rol: 'ENCARGADO'", "HTTP 201 Created. Usuario activo y auditoría generada.", "Aprobado"],
    ["CP-05", "Usuarios", "Rechazo por login duplicado", "Existe usuario 'admin'.", "usuario: 'admin', rol: 'ADMIN'", "HTTP 409 Conflict ('ya existe un usuario con ese nombre').", "Aprobado"],
    ["CP-06", "Usuarios", "Rechazo por contraseña débil", "Admin autenticado.", "clave: 'corta1' (< 10 caracteres)", "HTTP 400 Bad Request explicando directivas de clave.", "Aprobado"],
    ["CP-07", "Usuarios", "Protección del último admin activo", "Solo 1 admin activo.", "Admin intenta inactivar su cuenta.", "HTTP 409 Conflict ('debe quedar al menos un admin activo').", "Aprobado"],
    ["CP-08", "Usuarios", "Cambio personal de contraseña", "Usuario autenticado.", "clave_actual: 'OK', clave_nueva: 'Nueva123'", "HTTP 200 OK. Contraseña actualizada y auditada.", "Aprobado"],
    ["CP-09", "Celdas", "Registro exitoso de nueva celda", "Admin autenticado.", "codigo: 'A-05', tipo: 'CARRO'", "HTTP 201 Created. Celda en estado DISPONIBLE.", "Aprobado"],
    ["CP-10", "Celdas", "Rechazo por código duplicado", "Existe celda 'A-01'.", "codigo: 'a-01', tipo: 'CARRO'", "HTTP 409 Conflict ('ya existe una celda con ese código').", "Aprobado"],
    ["CP-11", "Celdas", "Inhabilitación de celda por mantenimiento", "Celda DISPONIBLE.", "estado: 'INACTIVA'", "HTTP 200 OK. Celda pasa a INACTIVA y se excluye de ingresos.", "Aprobado"],
    ["CP-12", "Celdas", "Bloqueo de inhabilitación en celda ocupada", "Celda con vehículo.", "id_celda: 2 (ocupada), estado: 'INACTIVA'", "HTTP 409 Conflict ('la celda está ocupada; registre salida').", "Aprobado"],
    ["CP-13", "Clientes", "Rechazo de mensualidad solapada", "Cliente con mensualidad activa.", "periodo que cruza fechas activas", "HTTP 409 Conflict ('la mensualidad se solapa').", "Aprobado"],
    ["CP-14", "Auditoría", "Inmutabilidad de la tabla de auditoría", "Filas en auditoria.", "Sentencia SQL UPDATE o DELETE directa", "Trigger trg_auditoria_inmutable aborta con excepción.", "Aprobado"]
]
add_table(doc, test_plan_data, [0.7, 1.1, 1.8, 1.1, 1.2, 1.8, 0.9])
add_p(doc, "Resultado de la suite de pruebas automatizadas: 100 % de aprobación y 0 fallos ejecutando 'go test -race ./...'.")
doc.add_page_break()

# ==============================================================================
# 10. REPOSITORIO GITHUB Y TRAZABILIDAD
# ==============================================================================
add_h1(doc, "9. Repositorio en GitHub y Evidencias de Código Fuente")
add_p(doc, f"El código fuente completo de las Iteraciones 1 y 2 se encuentra versionado bajo Git y alojado en GitHub en:")
add_p(doc, f"URL del Repositorio: {GITHUB}", bold=True, color=COLOR_AZUL)
add_p(doc, "En el commit c928179 se agregaron más de 4.100 líneas de código implementando la capa de gestión, seguridad PBKDF2, endpoints REST, triggers en PostgreSQL y pruebas automáticas.")

add_h1(doc, "10. Matriz de Trazabilidad entre Representaciones")
traz_data = [
    ["Requisito", "Historia", "Caso de Uso", "Componentes / Clases", "Tablas en BD", "Endpoint API", "Prueba"],
    ["RF-08, RF-09", "HU-07", "CU-06", "Router, MiddlewareAuth, SvcGestion, PBKDF2", "usuario", "POST /api/login", "CP-01, CP-02, CP-03"],
    ["RF-10, RF-11", "HU-06", "CU-08", "CtrlGestion, SvcGestion, Dominio, RepoPG", "usuario, auditoria", "GET, POST, PATCH /api/usuarios", "CP-04, CP-05, CP-07"],
    ["RF-12, RF-13", "HU-07", "CU-07", "CtrlGestion, SvcGestion, PBKDF2", "usuario, auditoria", "POST /api/sesion/clave", "CP-06, CP-08"],
    ["RF-15, RF-16, RF-17", "HU-08", "CU-09", "CtrlGestion, SvcGestion, RepoPG", "celda, auditoria", "GET, POST, PUT /api/celdas", "CP-09, CP-10, CP-11, CP-12"],
    ["RF-18, RF-19", "HU-09", "CU-10", "CtrlGestion, SvcGestion, Dominio", "cliente, vehiculo, mensualidad", "GET, POST /api/clientes", "CP-13"],
    ["RF-20, RF-21", "HU-10", "CU-11", "CtrlGestion, SvcGestion, Triggers", "auditoria (inmutable)", "GET /api/auditoria", "CP-14"]
]
add_table(doc, traz_data, [1.0, 0.8, 0.8, 1.8, 1.2, 1.4, 0.8])

add_h1(doc, "11. Conclusiones")
add_p(doc, "1. La Iteración 2 permitió transformar un prototipo monousuario en una solución multiusuario empresarial dotada de gobernabilidad y control de acceso basado en roles (RBAC).\n"
           "2. Se establecieron estándares rigurosos de seguridad mediante hashing PBKDF2 con sal individual, comparaciones en tiempo constante y auditoría inmutable respaldada por triggers en PostgreSQL.\n"
           "3. Reglas de negocio críticas (garantía del último administrador activo, protección de celdas ocupadas y compatibilidad de tipos celda-vehículo) quedaron blindadas tanto en la lógica de aplicación como a nivel de persistencia relacional.\n"
           "4. Se mantuvo una coherencia y trazabilidad absoluta desde los requerimientos de usuario y modelos en StarUML hasta el código fuente y las pruebas automatizadas (100 % aprobadas).")

doc.save(SALIDA)
print(f"Documento Word generado exitosamente: {SALIDA}")

