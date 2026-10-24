# -*- coding: utf-8 -*-
import sys, json
from PIL import Image
from reportlab.lib.pagesizes import letter
from reportlab.lib.units import cm
from reportlab.lib import colors
from reportlab.lib.styles import ParagraphStyle
from reportlab.lib.enums import TA_CENTER, TA_JUSTIFY
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont
from reportlab.platypus import (BaseDocTemplate, PageTemplate, Frame, Paragraph, Spacer, PageBreak,
                                Table, TableStyle, Image as RLImage, KeepTogether, NextPageTemplate, Preformatted)

D = "/usr/share/fonts/truetype/dejavu/"
pdfmetrics.registerFont(TTFont("DV", D + "DejaVuSans.ttf"))
pdfmetrics.registerFont(TTFont("DV-B", D + "DejaVuSans-Bold.ttf"))
pdfmetrics.registerFont(TTFont("DV-I", D + "DejaVuSans-Oblique.ttf"))
pdfmetrics.registerFont(TTFont("DV-BI", D + "DejaVuSans-BoldOblique.ttf"))
pdfmetrics.registerFont(TTFont("DVM", D + "DejaVuSansMono.ttf"))
pdfmetrics.registerFontFamily("DV", normal="DV", bold="DV-B", italic="DV-I", boldItalic="DV-BI")

AZUL = colors.HexColor("#0f4c81"); GRIS = colors.HexColor("#f2f5f8"); BORDE = colors.HexColor("#c5cfd9")
cfg = json.load(open(sys.argv[1], encoding="utf-8")) if len(sys.argv) > 1 else {}
DOCENTE = cfg.get("docente", "")
MATERIA = cfg.get("materia", "")
GITHUB = cfg.get("github", "")
FIGMA = cfg.get("figma", "")
NOMBRE = cfg.get("nombre", "Cristian España")
SALIDA = cfg.get("salida", "/mnt/user-data/outputs/España_Cristian_EA1.pdf")

def S(name, **kw):
    base = dict(fontName="DV", fontSize=9.5, leading=14, textColor=colors.HexColor("#1c2530"))
    base.update(kw); return ParagraphStyle(name, **base)

body = S("body", alignment=TA_JUSTIFY, spaceAfter=6)
small = S("small", fontSize=8, leading=11, textColor=colors.HexColor("#4a5663"))
cap = S("cap", fontSize=8, leading=11, alignment=TA_CENTER, textColor=colors.HexColor("#4a5663"), spaceAfter=10, fontName="DV-I")
h1 = S("h1", fontName="DV-B", fontSize=16, leading=20, textColor=AZUL, spaceBefore=4, spaceAfter=10)
h2 = S("h2", fontName="DV-B", fontSize=12, leading=16, textColor=AZUL, spaceBefore=10, spaceAfter=6)
h3 = S("h3", fontName="DV-B", fontSize=10, leading=14, spaceBefore=6, spaceAfter=3)
cell = S("cell", fontSize=8.3, leading=11)
cellb = S("cellb", fontSize=8.3, leading=11, fontName="DV-B", textColor=colors.white)
code = ParagraphStyle("code", fontName="DVM", fontSize=7.2, leading=9.2, backColor=GRIS, borderColor=BORDE, borderWidth=0.5, borderPadding=5)

def P(t, st=body): return Paragraph(t, st)

def tabla(filas, anchos, cab=True):
    data = [[P(c, cellb if (cab and i == 0) else cell) if isinstance(c, str) else c for c in f] for i, f in enumerate(filas)]
    t = Table(data, colWidths=anchos, repeatRows=1 if cab else 0)
    st = [("GRID", (0, 0), (-1, -1), 0.4, BORDE), ("VALIGN", (0, 0), (-1, -1), "TOP"),
          ("LEFTPADDING", (0, 0), (-1, -1), 5), ("RIGHTPADDING", (0, 0), (-1, -1), 5),
          ("TOPPADDING", (0, 0), (-1, -1), 4), ("BOTTOMPADDING", (0, 0), (-1, -1), 4)]
    if cab: st += [("BACKGROUND", (0, 0), (-1, 0), AZUL)]
    for i in range(1, len(data)):
        if i % 2 == 0: st.append(("BACKGROUND", (0, i), (-1, i), GRIS))
    t.setStyle(TableStyle(st)); return t

def img(path, ancho_max, alto_max):
    w, h = Image.open(path).size
    esc = min(ancho_max / w, alto_max / h)
    return RLImage(path, width=w * esc, height=h * esc)

def figura(path, titulo, ancho=16.4 * cm, alto=15 * cm):
    return KeepTogether([img(path, ancho, alto), Spacer(1, 3), P(titulo, cap)])

DG = "/home/claude/parqueadero/docs/diagramas/"
PT = "/home/claude/parqueadero/docs/pantallas/"
W = letter[0] - 4.4 * cm

# ---------- encabezado / pie ----------
def pie(c, d):
    c.saveState(); c.setFont("DV", 7.5); c.setFillColor(colors.HexColor("#4a5663"))
    c.drawString(2.2 * cm, 1.2 * cm, "Parqueadero Autos Colombia · EA1 · Iteración 1")
    c.drawRightString(letter[0] - 2.2 * cm, 1.2 * cm, "Página %d" % d.page)
    c.setStrokeColor(BORDE); c.line(2.2 * cm, 1.6 * cm, letter[0] - 2.2 * cm, 1.6 * cm); c.restoreState()

def portada(c, d):
    c.saveState(); c.setFillColor(AZUL); c.rect(0, letter[1] - 9 * cm, letter[0], 9 * cm, fill=1, stroke=0); c.restoreState()

doc = BaseDocTemplate(SALIDA, pagesize=letter, leftMargin=2.2 * cm, rightMargin=2.2 * cm, topMargin=2.2 * cm, bottomMargin=2.2 * cm,
                      title="Sistema de información del parqueadero Autos Colombia - EA1", author=NOMBRE)
fr = Frame(doc.leftMargin, doc.bottomMargin, W, letter[1] - 4.4 * cm, id="f")
doc.addPageTemplates([PageTemplate("portada", [fr], onPage=portada), PageTemplate("normal", [fr], onPage=pie)])

E = []
# ---------- PORTADA ----------
E += [Spacer(1, 1.2 * cm),
      P("S20 · EVIDENCIA DE APRENDIZAJE 1", S("p0", fontName="DV-B", fontSize=11, textColor=colors.white, leading=14)),
      Spacer(1, 6),
      P("Sistema de información del parqueadero “Autos Colombia”", S("p1", fontName="DV-B", fontSize=22, leading=28, textColor=colors.white)),
      Spacer(1, 8),
      P("Diseño de Sistema de Información – Iteración 1: Gestión de la entrada y salida de vehículos", S("p2", fontSize=13, leading=18, textColor=colors.white)),
      Spacer(1, 4.4 * cm)]
datos = [["Estudiante", NOMBRE], ["Programa", "Tecnología en Desarrollo de Software"], ["Institución", "IU Digital de Antioquia"],
         ["Asignatura", MATERIA or "______________________________"], ["Docente", DOCENTE or "______________________________"],
         ["Ciudad y fecha", "Medellín, Colombia · 21 de octubre de 2026"]]
t = Table([[P("<b>%s</b>" % a, S("pc", fontSize=10.5)), P(b, S("pc2", fontSize=10.5))] for a, b in datos], colWidths=[4 * cm, W - 4 * cm])
t.setStyle(TableStyle([("LINEBELOW", (0, 0), (-1, -1), 0.4, BORDE), ("TOPPADDING", (0, 0), (-1, -1), 8), ("BOTTOMPADDING", (0, 0), (-1, -1), 8)]))
E += [t, NextPageTemplate("normal"), PageBreak()]

# ---------- CONTENIDO ----------
E += [P("Contenido", h1)]
for i, x in enumerate(["Introducción", "Entregables de la Iteración 1", "   1. Requerimientos funcionales y no funcionales", "   2. Historias de usuario",
                       "   3. Arquitectura del sistema", "   4. Mockups de las historias de usuario", "   5. Modelo entidad-relación, relacional y físico",
                       "   6. Diagramas de casos de uso", "   7. Diagrama de clases", "   8. Diagrama de paquetes", "   9. Código fuente y repositorio",
                       "   10. Coherencia entre representaciones (trazabilidad)", "Conclusiones"]):
    E.append(P(x, S("toc", spaceAfter=3)))
E += [PageBreak()]

# ---------- INTRODUCCIÓN ----------
E += [P("Introducción", h1),
 P("El parqueadero “Autos Colombia” presta el servicio de parqueo por mensualidad y hoy necesita un sistema de información que le permita llevar sus operaciones diarias: "
   "registrar la entrada y salida de los vehículos, administrar usuarios y celdas, registrar las novedades ocurridas sobre un vehículo y gestionar los pagos."),
 P("Este documento presenta la <b>Iteración 1</b> del desarrollo, enfocada en la <b>gestión de la entrada y salida de vehículos</b>. Se siguió la ruta vista en la unidad: "
   "primero se entiende la necesidad y se transforma en requerimientos, reglas e historias de usuario con criterios de aceptación; luego se representa la solución con "
   "diagramas de casos de uso, clases, paquetes y arquitectura, y con el modelo de datos (entidad-relación, relacional y físico); finalmente se construye un prototipo funcional."),
 P("La solución es una <b>aplicación web responsive (PWA)</b> que puede usarse desde celulares, tablets y computadores sin instalar nada. Se implementó con React en el "
   "frontend, Go en el backend con arquitectura hexagonal y PostgreSQL como gestor de datos. Las iteraciones siguientes cubrirán la gestión de usuarios y celdas "
   "(administración completa) y la gestión de pagos."),
 P("<b>Alcance de la iteración.</b> Se asume, como en el ejercicio guía de la unidad, que los clientes, vehículos, mensualidades y celdas ya están registrados "
   "(datos iniciales) y que el encargado es quien realiza los registros. No se contemplan reservas futuras: el ingreso es inmediato. La salida incluye la posibilidad "
   "de dejar constancia del estado del vehículo mediante una novedad."),
 P("<b>Reglas de negocio adoptadas.</b>"),
 tabla([["Id", "Regla", "Necesidad que atiende"],
        ["RN-01", "Solo puede ingresar un vehículo registrado cuyo cliente tenga mensualidad vigente.", "Controlar que solo usen el servicio quienes lo tienen contratado."],
        ["RN-02", "Un vehículo tiene como máximo un registro activo (ingreso sin salida).", "Evitar ingresos duplicados."],
        ["RN-03", "Una celda tiene como máximo un vehículo activo, y solo se ocupa si está disponible y es del mismo tipo del vehículo (carro/moto).", "Conocer qué celda ocupa cada vehículo."],
        ["RN-04", "La salida exige un ingreso activo y libera la celda.", "Mantener actualizada la disponibilidad."],
        ["RN-05", "Las novedades se registran sobre un vehículo que está dentro (al ingresar, al salir o durante la estancia) y requieren tipo y descripción.", "Dejar constancia de lo ocurrido con cada vehículo."],
        ["RN-06", "Los registros y novedades no se eliminan; quedan como historial.", "Trazabilidad ante reclamaciones."]],
       [1.4 * cm, 9.2 * cm, W - 10.6 * cm]),
 PageBreak()]

# ---------- ENTREGABLES ----------
E += [P("Entregables de la Iteración 1", h1), P("1. Requerimientos funcionales y no funcionales", h2),
      P("Un requerimiento funcional describe un comportamiento o servicio que el sistema debe ofrecer; la función permite aplicar una regla y atender una necesidad. "
        "Los requerimientos no funcionales son condiciones de calidad y restricciones, redactadas de forma verificable.")]
E += [P("Requerimientos funcionales", h3),
 tabla([["Id", "Requerimiento", "Regla"],
  ["RF-01", "El sistema debe permitir al encargado registrar el ingreso de un vehículo indicando la placa, la celda y registrando automáticamente la fecha y hora.", "RN-02, RN-03"],
  ["RF-02", "El sistema debe validar, antes de confirmar un ingreso, que el vehículo esté registrado, que su cliente tenga mensualidad vigente, que no tenga un ingreso activo y que la celda esté disponible y sea del tipo del vehículo; en caso contrario debe rechazar el ingreso con un mensaje que explique el motivo.", "RN-01, RN-02, RN-03"],
  ["RF-03", "El sistema debe permitir al encargado registrar la salida de un vehículo con ingreso activo, guardando la fecha y hora de salida, calculando el tiempo de permanencia y liberando la celda.", "RN-04"],
  ["RF-04", "El sistema debe permitir registrar una novedad (daño, faltante u otra) con su descripción sobre un vehículo que esté dentro, ya sea al ingresar, al salir o durante su estancia.", "RN-05"],
  ["RF-05", "El sistema debe permitir consultar las celdas y filtrarlas por estado (disponible, ocupada, inactiva).", "RN-03"],
  ["RF-06", "El sistema debe mostrar los vehículos que están dentro del parqueadero con su placa, cliente, celda y hora de ingreso.", "RN-02"],
  ["RF-07", "El sistema debe conservar el historial de registros de ingreso/salida y de novedades sin permitir su eliminación.", "RN-06"]],
  [1.3 * cm, W - 3.7 * cm, 2.4 * cm])]
E += [P("Requerimientos no funcionales", h3),
 tabla([["Id", "Calidad", "Requerimiento verificable"],
  ["RNF-01", "Rendimiento", "El 95 % de las consultas y registros debe responder en máximo 2 segundos, con 200 celdas y 5 usuarios concurrentes."],
  ["RNF-02", "Usabilidad", "4 de cada 5 encargados deben completar un registro de ingreso sin ayuda, tras una inducción de 10 minutos. Las acciones principales deben quedar a máximo 2 toques desde el inicio."],
  ["RNF-03", "Compatibilidad y portabilidad", "La aplicación debe funcionar en celulares, tablets y computadores, con las dos últimas versiones de Chrome, Edge, Firefox y Safari, y mostrarse sin desplazamiento horizontal desde 360 px de ancho."],
  ["RNF-04", "Seguridad", "Todas las operaciones deben exigir autenticación; las consultas a la base de datos deben ser parametrizadas; en producción la comunicación debe usar HTTPS."],
  ["RNF-05", "Integridad", "La base de datos debe impedir que un vehículo o una celda tengan más de un registro activo, incluso con operaciones simultáneas."],
  ["RNF-06", "Disponibilidad", "El sistema debe estar disponible al menos el 99 % del horario de operación del parqueadero en cada mes."],
  ["RNF-07", "Mantenibilidad", "Las reglas de negocio deben estar separadas de la interfaz y de la base de datos (arquitectura en capas) y cubiertas por pruebas automáticas."]],
  [1.5 * cm, 3.2 * cm, W - 4.7 * cm])]
E += [PageBreak()]

# ---------- HISTORIAS ----------
E += [P("2. Historias de usuario", h2),
 P("Cada historia sigue la plantilla <i>Como [rol], quiero [capacidad], para [beneficio]</i> y se acompaña de criterios de aceptación en formato "
   "<b>Dado / Cuando / Entonces</b>, incluyendo rechazos y excepciones.")]
def historia(id_, titulo, texto, rf, criterios):
    filas = [[P("<b>%s · %s</b>" % (id_, titulo), cellb), P("Relacionada con: %s" % rf, S("hr", fontSize=8, textColor=colors.white))],
             [P("<i>%s</i>" % texto, cell), ""]]
    for i, (n, d, c, e) in enumerate(criterios):
        filas.append([P("<b>CA-%s.%d</b> %s" % (id_[3:], i + 1, n), cell),
                      P("<b>Dado</b> %s<br/><b>Cuando</b> %s<br/><b>Entonces</b> %s" % (d, c, e), cell)])
    t = Table(filas, colWidths=[4.6 * cm, W - 4.6 * cm])
    st = [("GRID", (0, 0), (-1, -1), 0.4, BORDE), ("BACKGROUND", (0, 0), (-1, 0), AZUL), ("SPAN", (0, 1), (-1, 1)),
          ("BACKGROUND", (0, 1), (-1, 1), GRIS), ("VALIGN", (0, 0), (-1, -1), "TOP"),
          ("TOPPADDING", (0, 0), (-1, -1), 4), ("BOTTOMPADDING", (0, 0), (-1, -1), 4)]
    t.setStyle(TableStyle(st)); return [t, Spacer(1, 8)]

E += historia("HU-01", "Registrar ingreso", "Como encargado, quiero registrar el ingreso de un vehículo en una celda, para saber qué vehículo ocupa cada celda y desde cuándo.", "RF-01, RF-02", [
  ("Ingreso válido", "un vehículo registrado, con mensualidad vigente y sin ingreso activo, y una celda disponible de su mismo tipo", "el encargado confirma el ingreso", "se crea el registro con la fecha y hora actuales y la celda queda ocupada"),
  ("Mensualidad vencida", "un vehículo cuyo cliente no tiene mensualidad vigente", "el encargado intenta confirmar el ingreso", "el sistema rechaza el ingreso, informa que no hay mensualidad vigente y no crea ningún registro"),
  ("Vehículo ya dentro", "un vehículo que ya tiene un ingreso activo", "el encargado intenta registrar otro ingreso", "el sistema rechaza el ingreso e informa que el vehículo ya está dentro"),
  ("Celda no utilizable", "una celda ocupada, inactiva o de un tipo distinto al del vehículo", "el encargado intenta asignarla", "el sistema rechaza el ingreso, informa el motivo y la celda no cambia de estado"),
  ("Placa no registrada", "una placa que no existe en el sistema", "el encargado intenta registrar el ingreso", "el sistema informa que el vehículo no está registrado"),
  ("Ingreso con novedad", "un ingreso válido y una novedad con tipo y descripción", "el encargado confirma el ingreso", "se registran el ingreso y la novedad de tipo ENTRADA en la misma operación")])
E += historia("HU-02", "Registrar salida", "Como encargado, quiero registrar la salida de un vehículo, para liberar su celda y conocer cuánto tiempo permaneció.", "RF-03", [
  ("Salida válida", "un vehículo con ingreso activo", "el encargado confirma la salida", "se guarda la fecha y hora de salida, se muestra el tiempo de permanencia y la celda queda disponible"),
  ("Sin ingreso activo", "un vehículo sin ingreso activo", "el encargado intenta registrar la salida", "el sistema informa que el vehículo no tiene un ingreso activo y no modifica ninguna celda"),
  ("Salida con novedad", "un vehículo con ingreso activo y una novedad con tipo y descripción", "el encargado confirma la salida", "se registran la salida y la novedad de tipo SALIDA (por ejemplo, un daño detectado)")])
E += historia("HU-03", "Registrar novedad", "Como encargado, quiero registrar una novedad sobre un vehículo que está dentro, para dejar constancia de daños, faltantes u otros hechos.", "RF-04", [
  ("Novedad válida", "un vehículo con ingreso activo", "el encargado registra una novedad con tipo y descripción", "la novedad queda asociada al registro activo con la fecha y hora actuales"),
  ("Descripción vacía", "un vehículo con ingreso activo", "el encargado intenta guardar una novedad sin descripción", "el sistema rechaza la novedad y pide completar la descripción"),
  ("Vehículo fuera", "un vehículo sin ingreso activo", "el encargado intenta registrar una novedad", "el sistema informa que el vehículo no está dentro del parqueadero")])
E += historia("HU-04", "Consultar celdas", "Como encargado, quiero ver las celdas y su estado, para asignar rápidamente una celda disponible al vehículo que llega.", "RF-05", [
  ("Filtrar disponibles", "celdas en distintos estados", "el encargado consulta las celdas disponibles", "el sistema muestra solo las celdas disponibles con su código y tipo"),
  ("Sin disponibles", "que todas las celdas estén ocupadas o inactivas", "el encargado consulta las disponibles", "el sistema muestra la lista vacía y el ingreso no puede confirmarse")])
E += historia("HU-05", "Consultar vehículos dentro", "Como encargado, quiero ver los vehículos que están dentro, para controlar la ocupación del parqueadero.", "RF-06", [
  ("Listado de activos", "vehículos con ingreso activo", "el encargado abre la vista de estado", "el sistema lista placa, cliente, celda y hora de ingreso, del más antiguo al más reciente")])
E += [PageBreak()]

# ---------- ARQUITECTURA ----------
E += [P("3. Arquitectura del sistema", h2),
 P("Se adoptó una <b>arquitectura en capas con enfoque hexagonal</b>, de modo que cada parte tenga responsabilidades explícitas: la presentación recibe datos, acciones y mensajes; "
   "la aplicación y el dominio coordinan y aplican las reglas del negocio; la persistencia consulta y almacena. El dominio no depende ni de HTTP ni de la base de datos: "
   "la aplicación define un <i>puerto</i> (interfaz <font name='DVM'>Repositorio</font>) y el adaptador de PostgreSQL lo implementa."),
 figura(DG + "arquitectura.png", "Figura 1. Arquitectura en capas del sistema (fuente PlantUML: docs/diagramas/01_arquitectura.puml)", alto=10 * cm),
 tabla([["Elemento", "Tecnología", "Responsabilidad"],
        ["Frontend", "React + Vite, PWA responsive", "Formularios, mensajes al usuario y navegación; se adapta a celular, tablet y PC."],
        ["Presentación (API)", "Go, net/http, JSON", "Valida la entrada, controla el acceso con token y traduce errores de negocio a códigos HTTP (400, 401, 404, 409, 422)."],
        ["Aplicación", "Go", "Casos de uso: registrar ingreso, salida y novedad; consultar celdas y vehículos dentro."],
        ["Dominio", "Go", "Entidades y reglas (placa válida, celda utilizable, novedad válida, duración de la estancia)."],
        ["Persistencia", "PostgreSQL 16, database/sql", "Transacciones: ocupar celda + crear registro + guardar novedad se hacen juntos o no se hacen."]],
       [3.2 * cm, 4.4 * cm, W - 7.6 * cm]),
 Spacer(1, 6),
 P("<b>Nota sobre el framework web.</b> La capa HTTP se construyó con la biblioteca estándar de Go para que el proyecto compile sin dependencias externas en el entorno de pruebas. "
   "Al estar aislada en un adaptador, puede reemplazarse por Fiber sin tocar la aplicación ni el dominio.", small),
 PageBreak()]

# ---------- MOCKUPS ----------
E += [P("4. Mockups de las historias de usuario", h2),
 P("Los mockups muestran la apariencia visual (tipografía, color y jerarquía) de las pantallas asociadas a las historias de usuario. La interfaz se diseñó pensando primero en el "
   "celular: formularios de una sola columna, botones grandes y navegación inferior con las cuatro acciones del encargado (Ingreso, Salida, Novedad y Estado). Cada pantalla "
   "muestra los datos, las acciones y los mensajes que el usuario necesita para saber qué ocurrió y cómo continuar. Las imágenes son capturas del prototipo funcional; los mockups editables se encuentran en Figma."),
 P("Archivo en Figma: %s" % (FIGMA or "[pegar aquí el enlace del archivo de Figma]"), S("fg", fontName="DV-B", fontSize=9)), Spacer(1, 4)]
def fila_pantallas(items):
    celdas = [[img(PT + f, 3.6 * cm, 7.2 * cm), P(c, cap)] for f, c in items]
    t = Table([[x[0] for x in celdas], [x[1] for x in celdas]], colWidths=[W / len(celdas)] * len(celdas))
    t.setStyle(TableStyle([("ALIGN", (0, 0), (-1, -1), "CENTER"), ("VALIGN", (0, 0), (-1, 0), "TOP")])); return t
E += [fila_pantallas([("02_ingreso_lleno.png", "HU-01 · Registrar ingreso: placa, celda libre y novedad opcional"),
                      ("03_ingreso_ok.png", "HU-01 · Ingreso registrado"),
                      ("04_ingreso_rechazo.png", "HU-01 · Rechazo por mensualidad vencida")]),
      Spacer(1, 6),
      fila_pantallas([("07_salida.png", "HU-02 · Registrar salida eligiendo un vehículo dentro"),
                      ("06_novedad.png", "HU-03 · Registrar novedad"),
                      ("05_patio.png", "HU-04 y HU-05 · Patio: celdas y vehículos dentro")]),
      PageBreak(),
      P("Vista en computador (misma aplicación, mismo código)", h3),
      figura(PT + "09_escritorio.png", "Figura 2. Vista de escritorio: formulario y patio a la vez; en tablet se conserva esta disposición", alto=9.5 * cm),
      P("<b>De boceto a mockup y prototipo.</b> El boceto fija la estructura (qué datos y acciones van en cada pantalla); el mockup añade la apariencia visual; el prototipo simula la "
        "interacción y la navegación. Para esta iteración, la interfaz en React funciona como prototipo navegable conectado al backend real.", body),
      Spacer(1, 10)]

# ---------- DATOS ----------
E += [P("5. Modelo entidad-relación, modelo relacional y modelo físico", h2),
 P("<b>Modelo entidad-relación (conceptual).</b> Describe las entidades del problema y sus relaciones con cardinalidad. Un cliente posee varios vehículos y contrata mensualidades; "
   "cada vez que un vehículo ingresa se crea un <i>registro</i> en una celda; sobre ese registro pueden existir cero o más novedades. "
   "La fecha de salida pertenece a un registro específico, no al vehículo."),
 figura(DG + "mer.png", "Figura 3. Modelo entidad-relación", ancho=14 * cm, alto=7 * cm),
 P("<b>Modelo relacional.</b> Traduce las entidades a tablas: PK identifica un registro y FK mantiene una referencia válida. La relación entre vehículo y celda es de muchos a muchos "
   "y se resuelve con la tabla REGISTRO, que además guarda las fechas de la estancia."),
 figura(DG + "relacional.png", "Figura 4. Modelo relacional", ancho=14 * cm, alto=9 * cm),
 PageBreak(),
 P("<b>Modelo físico.</b> Concreta el diseño para PostgreSQL: tipos de datos, restricciones (clave primaria, claves foráneas, datos obligatorios, CHECK) e índices según las consultas del sistema.", body),
 figura(DG + "fisico.png", "Figura 5. Modelo físico (script completo en backend/db/schema.sql)", alto=11.5 * cm),
 tabla([["Decisión", "Implementación"],
        ["Tipos", "Identificadores SERIAL; placa VARCHAR(6); nombre VARCHAR(100); fechas de estancia TIMESTAMPTZ; vigencia de mensualidad DATE."],
        ["Restricciones", "PK y FK en todas las tablas; UNIQUE en documento, placa y código de celda; CHECK en tipo, estado, momento y rangos de fechas; campos obligatorios NOT NULL."],
        ["Reglas en la base de datos", "Dos índices únicos parciales (<font name='DVM'>WHERE fecha_salida IS NULL</font>) garantizan un solo registro activo por vehículo y por celda (RN-02, RN-03), aun con concurrencia."],
        ["Acceso", "Índices por (cliente, fechas) para validar la mensualidad, por estado de celda para consultar disponibles y por registro para listar novedades."]],
       [3.4 * cm, W - 3.4 * cm]),
 Spacer(1, 6),
 P("Fragmento clave del script:", small),
 Preformatted("CREATE TABLE IF NOT EXISTS registro (\n    id_registro   SERIAL PRIMARY KEY,\n    id_vehiculo   INT         NOT NULL REFERENCES vehiculo(id_vehiculo),\n"
              "    id_celda      INT         NOT NULL REFERENCES celda(id_celda),\n    fecha_entrada TIMESTAMPTZ NOT NULL,\n    fecha_salida  TIMESTAMPTZ,\n"
              "    CONSTRAINT ck_registro_fechas CHECK (fecha_salida IS NULL OR fecha_salida >= fecha_entrada)\n);\n"
              "CREATE UNIQUE INDEX uq_registro_vehiculo_activo ON registro(id_vehiculo) WHERE fecha_salida IS NULL;\n"
              "CREATE UNIQUE INDEX uq_registro_celda_activa    ON registro(id_celda)    WHERE fecha_salida IS NULL;", code),
 PageBreak()]

# ---------- CASOS DE USO ----------
E += [P("6. Diagramas de casos de uso", h2),
 P("Cada caso de uso expresa un objetivo del actor. El único actor de esta iteración es el <b>Encargado</b> del parqueadero. El ingreso incluye la validación de la mensualidad y de la celda, "
   "y tanto el ingreso como la salida pueden extenderse con el registro de una novedad."),
 figura(DG + "casos_uso.png", "Figura 6. Diagrama de casos de uso de la Iteración 1", alto=9.5 * cm),
 P("Especificación de los casos de uso principales", h3),
 tabla([["Campo", "CU-02 Registrar ingreso", "CU-03 Registrar salida"],
  ["Actor", "Encargado", "Encargado"],
  ["Precondiciones", "Encargado autenticado; vehículo, cliente y celdas existentes.", "Encargado autenticado; el vehículo tiene un ingreso activo."],
  ["Flujo principal", "1. Identifica el vehículo por placa.<br/>2. Selecciona una celda disponible.<br/>3. (Opcional) Agrega una novedad.<br/>4. Confirma el ingreso.<br/>5. El sistema comprueba las condiciones.<br/>6. Registra el ingreso, ocupa la celda y muestra el resultado.",
   "1. Identifica el vehículo por placa.<br/>2. (Opcional) Agrega una novedad.<br/>3. Confirma la salida.<br/>4. El sistema comprueba que hay ingreso activo.<br/>5. Registra la salida, libera la celda y muestra el tiempo de permanencia."],
  ["Flujos alternos", "A1. Placa no registrada.<br/>A2. Mensualidad vencida.<br/>A3. Vehículo ya dentro.<br/>A4. Celda ocupada, inactiva o de otro tipo.<br/>En todos: el sistema informa el motivo y no registra nada.",
   "A1. Placa no registrada.<br/>A2. Sin ingreso activo: el sistema informa y no modifica celdas."],
  ["Postcondiciones", "Registro activo creado y celda ocupada.", "Registro cerrado con fecha de salida y celda disponible."]],
  [2.9 * cm, (W - 2.9 * cm) / 2, (W - 2.9 * cm) / 2]),
 PageBreak()]

# ---------- CLASES / PAQUETES ----------
E += [P("7. Diagrama de clases", h2),
 P("Las clases del dominio representan la información y los estados que cambian, y las operaciones que aplican las reglas. Un cliente tiene muchos vehículos; un vehículo y una celda "
   "tienen muchos registros históricos, pero <b>un único registro activo como máximo</b>. La clase <i>Servicio</i> coordina los casos de uso a través del puerto <i>Repositorio</i>."),
 figura(DG + "clases.png", "Figura 7. Diagrama de clases", alto=15.5 * cm),
 PageBreak(),
 P("8. Diagrama de paquetes", h2),
 P("Un paquete agrupa elementos relacionados del modelo o del código. Se organizó en tres grupos: <b>presentación</b> (frontend y adaptador HTTP), <b>aplicación y dominio</b> "
   "(casos de uso, objetos y reglas) y <b>persistencia</b> (adaptador PostgreSQL y scripts SQL). Las dependencias apuntan hacia el dominio, nunca al revés."),
 figura(DG + "paquetes.png", "Figura 8. Diagrama de paquetes", ancho=13 * cm, alto=11 * cm),
 P("Estructura del repositorio", h3),
 Preformatted("parqueadero/\n├── backend/\n│   ├── cmd/api/main.go                    (arranque)\n│   ├── internal/domain/                   (entidades y reglas)\n"
              "│   ├── internal/application/              (casos de uso + pruebas)\n│   ├── internal/adapters/httpapi/         (API REST)\n"
              "│   ├── internal/adapters/postgres/        (persistencia)\n│   └── db/schema.sql, seed.sql            (modelo físico y datos iniciales)\n"
              "├── frontend/                              (React + Vite, PWA)\n├── docs/diagramas/                        (fuentes PlantUML + PNG)\n"
              "├── docs/pantallas/                        (capturas)\n└── docker-compose.yml, README.md", code),
 PageBreak()]

# ---------- CÓDIGO ----------
E += [P("9. Código fuente y repositorio", h2),
 P("Repositorio en GitHub: %s" % (GITHUB or "[pegar aquí la URL del repositorio]"), S("gh", fontName="DV-B", fontSize=9)),
 P("El código fuente se ejecutó y verificó de punta a punta: base de datos PostgreSQL con el script del modelo físico y datos de ejemplo, API en Go y frontend en React servido por el mismo backend."),
 P("Cómo ejecutarlo", h3),
 Preformatted("docker compose up -d db                       # PostgreSQL con schema.sql y seed.sql\ncd frontend && npm install && npm run build\ncp -r dist ../backend/static\n"
              "cd ../backend && API_TOKEN=clave go run ./cmd/api   # http://localhost:8080", code),
 P("API REST", h3),
 tabla([["Método y ruta", "Caso de uso", "Respuestas"],
        ["POST /api/ingresos", "Registrar ingreso (RF-01, RF-02)", "201; 400 placa inválida; 404 no registrado; 409 vehículo dentro o celda no disponible; 422 mensualidad vencida o celda incompatible"],
        ["POST /api/salidas", "Registrar salida (RF-03)", "200 con minutos de permanencia; 404; 409 sin ingreso activo"],
        ["POST /api/novedades", "Registrar novedad (RF-04)", "201; 400 novedad inválida; 409 sin ingreso activo"],
        ["GET /api/celdas?estado=", "Consultar celdas (RF-05)", "200"],
        ["GET /api/registros/activos", "Vehículos dentro (RF-06)", "200"]],
       [4.6 * cm, 4.4 * cm, W - 9 * cm]),
 P("Verificación realizada", h3),
 P("<b>Pruebas automáticas del backend</b> (<font name='DVM'>go test -race ./...</font>): dominio, casos de uso, API HTTP, configuración y PostgreSQL real. Cubren el ingreso válido, todos los rechazos "
   "(placa inválida, vehículo no registrado, mensualidad vencida, vehículo ya dentro, celda ocupada, inactiva o de otro tipo), la salida, las novedades, la zona horaria de Colombia en el último día de la mensualidad, "
   "la autenticación y el límite de intentos, y una carrera entre varios encargados por la misma celda. Resultado: <b>55 pruebas, 0 fallos</b> (51 sin base de datos; 4 de integración la requieren). "
   "Cobertura por paquete: dominio y configuración 100 %, API 90 %, aplicación 89 %, persistencia 74 %. En el frontend: 17 pruebas de la lógica de formato y validación, y análisis estático (ESLint) sin errores."),
 P("<b>Prueba de integración</b> contra PostgreSQL real, con peticiones HTTP: acceso sin token (401), ingreso con novedad (201), ingreso repetido (409), celda ocupada (409), "
   "mensualidad vencida (422), celda inactiva (409), celda de otro tipo (422), novedad durante la estancia (201), listado de vehículos dentro, salida con novedad (200), "
   "segunda salida (409) y celda nuevamente disponible. Las tres novedades quedaron guardadas con su momento (ENTRADA, DURANTE_ESTANCIA, SALIDA)."),
 P("<b>Cliente web:</b> el frontend compilado se sirvió desde el backend y se recorrieron con un navegador las pantallas de ingreso, rechazo, salida, novedad y estado en formato móvil y de escritorio (Figuras de la sección 4).", body),
 PageBreak()]

# ---------- TRAZABILIDAD ----------
E += [P("10. Coherencia entre representaciones (trazabilidad)", h2),
 P("Una regla modificada puede afectar varias representaciones. Esta matriz permite revisar que cada necesidad esté cubierta de la historia al dato."),
 tabla([["Requisito", "Historia", "Caso de uso", "Clases", "Tablas", "Endpoint"],
        ["RF-01, RF-02", "HU-01", "CU-02", "Servicio, Vehiculo, Celda, Registro, Mensualidad", "vehiculo, mensualidad, celda, registro", "POST /api/ingresos"],
        ["RF-03", "HU-02", "CU-03", "Servicio, Registro, Celda", "registro, celda", "POST /api/salidas"],
        ["RF-04", "HU-03", "CU-04", "Servicio, Novedad", "novedad", "POST /api/novedades"],
        ["RF-05", "HU-04", "CU-01", "Celda", "celda", "GET /api/celdas"],
        ["RF-06", "HU-05", "CU-05", "Registro, Vehiculo, Cliente", "registro, vehiculo, cliente, celda", "GET /api/registros/activos"],
        ["RF-07", "HU-01 a HU-03", "CU-02 a CU-04", "Registro, Novedad", "registro, novedad (sin borrado)", "—"]],
       [2.1 * cm, 2.1 * cm, 2.2 * cm, 3.7 * cm, 3.4 * cm, W - 13.5 * cm]),
 Spacer(1, 12),
 P("Conclusiones", h1),
 P("La Iteración 1 permitió llevar el problema del parqueadero desde la necesidad hasta un prototipo que funciona. Separar <b>necesidad, regla y función</b> ayudó a que cada requerimiento "
   "tuviera una razón clara y una forma de comprobarse; las historias con criterios Dado/Cuando/Entonces hicieron explícitos los rechazos y excepciones, que son donde más se equivocan los sistemas de este tipo."),
 P("Los diagramas y el modelo de datos se mantuvieron coherentes entre sí: la regla de “un solo registro activo por vehículo y por celda” aparece en las historias, en las clases, en la lógica "
   "de la aplicación y, además, como restricción en la base de datos, lo que protege el sistema incluso si dos encargados operan al mismo tiempo."),
 P("La arquitectura en capas con enfoque hexagonal dejó las reglas de negocio aisladas de la interfaz y de PostgreSQL, y eso hizo posible probarlas de forma automática. "
   "Elegir una aplicación web responsive cumplió el requisito de uso en celulares, tablets y computadores con un único código."),
 P("<b>Limitaciones y trabajo siguiente.</b> El acceso se controla con una clave compartida; las próximas iteraciones deben incorporar usuarios, roles y auditoría de quién registra cada operación "
   "(gestión de usuarios), la administración completa de celdas, clientes y vehículos, y la gestión de pagos que se apoyará en la entidad mensualidad ya modelada. "
   "Los requisitos de rendimiento y usabilidad (RNF-01 y RNF-02) quedan por medirse con pruebas de carga y pruebas con encargados reales.")]
doc.build(E)
print("PDF generado:", SALIDA)
