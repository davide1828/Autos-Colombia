# -*- coding: utf-8 -*-
"""
Generador de diagramas de Casos de Uso (Iteración 2)
1. Archivo StarUML (.mdj)
2. Imagen PNG de alta resolución para el documento
"""
import json, uuid, math
from PIL import Image, ImageDraw, ImageFont

def uid():
    return "AAAAA" + uuid.uuid4().hex[:19]

# ==============================================================================
# 1. GENERACIÓN DEL ARCHIVO STARUML (.mdj)
# ==============================================================================
proj_id = uid()
model_id = uid()
diag_id = uid()

# Actores
act_admin_id = uid()
act_enc_id = uid()

# Casos de uso principales
uc_06_id = uid() # Iniciar sesión
uc_07_id = uid() # Cambiar propia contraseña
uc_08_id = uid() # Administrar usuarios
uc_09_id = uid() # Administrar celdas
uc_10_id = uid() # Administrar clientes y mensualidades
uc_11_id = uid() # Consultar auditoría

# Casos de uso incluidos
uc_vcred_id = uid()  # Validar credenciales y tasa
uc_vclave_id = uid() # Validar política contraseña
uc_vadmin_id = uid() # Garantizar último admin
uc_vocup_id = uid()  # Validar celda no ocupada
uc_vmens_id = uid()  # Validar no solapamiento
uc_vveh_id = uid()   # Validar tipo vehículo
uc_raud_id = uid()   # Registrar auditoría

def mk_assoc(source_id, target_id, name=""):
    aid = uid()
    end1_id = uid()
    end2_id = uid()
    return {
        "_type": "UMLAssociation",
        "_id": aid,
        "_parent": {"$ref": source_id},
        "name": name,
        "end1": {"_type": "UMLAssociationEnd", "_id": end1_id, "_parent": {"$ref": aid}, "reference": {"$ref": source_id}},
        "end2": {"_type": "UMLAssociationEnd", "_id": end2_id, "_parent": {"$ref": aid}, "reference": {"$ref": target_id}}
    }

def mk_include(source_id, target_id):
    iid = uid()
    return {
        "_type": "UMLInclude",
        "_id": iid,
        "_parent": {"$ref": source_id},
        "source": {"$ref": source_id},
        "target": {"$ref": target_id}
    }

# Estructura del modelo StarUML
model_elements = [
    {"_type": "UMLActor", "_id": act_admin_id, "_parent": {"$ref": model_id}, "name": "Administrador"},
    {"_type": "UMLActor", "_id": act_enc_id, "_parent": {"$ref": model_id}, "name": "Encargado"},

    {"_type": "UMLUseCase", "_id": uc_06_id, "_parent": {"$ref": model_id}, "name": "CU-06 Iniciar sesión"},
    {"_type": "UMLUseCase", "_id": uc_07_id, "_parent": {"$ref": model_id}, "name": "CU-07 Cambiar propia contraseña"},
    {"_type": "UMLUseCase", "_id": uc_08_id, "_parent": {"$ref": model_id}, "name": "CU-08 Administrar usuarios del sistema"},
    {"_type": "UMLUseCase", "_id": uc_09_id, "_parent": {"$ref": model_id}, "name": "CU-09 Administrar celdas"},
    {"_type": "UMLUseCase", "_id": uc_10_id, "_parent": {"$ref": model_id}, "name": "CU-10 Administrar clientes y mensualidades"},
    {"_type": "UMLUseCase", "_id": uc_11_id, "_parent": {"$ref": model_id}, "name": "CU-11 Consultar bitácora de auditoría"},

    {"_type": "UMLUseCase", "_id": uc_vcred_id, "_parent": {"$ref": model_id}, "name": "Validar credenciales y tasa de intentos"},
    {"_type": "UMLUseCase", "_id": uc_vclave_id, "_parent": {"$ref": model_id}, "name": "Validar política de robustez de contraseña"},
    {"_type": "UMLUseCase", "_id": uc_vadmin_id, "_parent": {"$ref": model_id}, "name": "Garantizar al menos un administrador activo"},
    {"_type": "UMLUseCase", "_id": uc_vocup_id, "_parent": {"$ref": model_id}, "name": "Validar que la celda no esté ocupada"},
    {"_type": "UMLUseCase", "_id": uc_vmens_id, "_parent": {"$ref": model_id}, "name": "Validar no solapamiento de fechas de mensualidad"},
    {"_type": "UMLUseCase", "_id": uc_vveh_id, "_parent": {"$ref": model_id}, "name": "Validar tipo vehículo (fuera del parqueadero)"},
    {"_type": "UMLUseCase", "_id": uc_raud_id, "_parent": {"$ref": model_id}, "name": "Registrar evento de auditoría inmutable"},

    # Asociaciones
    mk_assoc(act_enc_id, uc_06_id),
    mk_assoc(act_enc_id, uc_07_id),
    mk_assoc(act_enc_id, uc_10_id),

    mk_assoc(act_admin_id, uc_06_id),
    mk_assoc(act_admin_id, uc_07_id),
    mk_assoc(act_admin_id, uc_08_id),
    mk_assoc(act_admin_id, uc_09_id),
    mk_assoc(act_admin_id, uc_10_id),
    mk_assoc(act_admin_id, uc_11_id),

    # Includes
    mk_include(uc_06_id, uc_vcred_id),
    mk_include(uc_07_id, uc_vclave_id),
    mk_include(uc_08_id, uc_vclave_id),
    mk_include(uc_08_id, uc_vadmin_id),
    mk_include(uc_08_id, uc_raud_id),
    mk_include(uc_09_id, uc_vocup_id),
    mk_include(uc_09_id, uc_raud_id),
    mk_include(uc_10_id, uc_vmens_id),
    mk_include(uc_10_id, uc_vveh_id),
    mk_include(uc_10_id, uc_raud_id),
]

staruml_project = {
    "_type": "Project",
    "_id": proj_id,
    "name": "Parqueadero Autos Colombia",
    "ownedElements": [
        {
            "_type": "UMLModel",
            "_id": model_id,
            "_parent": {"$ref": proj_id},
            "name": "Iteración 2 - Gestión de Usuarios y Celdas",
            "ownedElements": model_elements + [
                {
                    "_type": "UMLUseCaseDiagram",
                    "_id": diag_id,
                    "_parent": {"$ref": model_id},
                    "name": "Diagrama de Casos de Uso - Gestión de Usuarios y Celdas"
                }
            ]
        }
    ]
}

with open(r"c:\universidad\ingenieria de software\parqueadero_EA1_codigo\parqueadero\docs\diagramas\parqueadero_iteracion_2.mdj", "w", encoding="utf-8") as f:
    json.dump(staruml_project, f, indent=2, ensure_ascii=False)
print("Archivo StarUML generado exitosamente: docs/diagramas/parqueadero_iteracion_2.mdj")


# ==============================================================================
# 2. GENERACIÓN DE IMAGEN PNG DE ALTA RESOLUCIÓN
# ==============================================================================
W, H = 1450, 1020
img = Image.new("RGB", (W, H), "#FFFFFF")
draw = ImageDraw.Draw(img)

# Paleta
C_AZUL = "#0F4C81"
C_CELESTE = "#E8F1F8"
C_BORDE_CU = "#0F4C81"
C_TEXTO = "#1C2530"
C_FONDO_SYS = "#F8FAFC"
C_BORDE_SYS = "#94A3B8"
C_LINEA = "#0F4C81"
C_LINEA_INC = "#64748B"

# Tipografías
font_title = ImageFont.truetype(r"C:\Windows\Fonts\arialbd.ttf", 22)
font_sys = ImageFont.truetype(r"C:\Windows\Fonts\arialbd.ttf", 16)
font_actor = ImageFont.truetype(r"C:\Windows\Fonts\arialbd.ttf", 14)
font_cu_title = ImageFont.truetype(r"C:\Windows\Fonts\arialbd.ttf", 12)
font_cu_sub = ImageFont.truetype(r"C:\Windows\Fonts\arial.ttf", 11)
font_inc = ImageFont.truetype(r"C:\Windows\Fonts\ariali.ttf", 11)

# Encabezado
draw.text((40, 22), "Diagrama de Casos de Uso — Iteración 2: Gestión de Usuarios y Celdas", fill=C_AZUL, font=font_title)
draw.text((40, 52), "Sistema de Información Parqueadero Autos Colombia · StarUML Export", fill="#64748B", font=font_inc)
draw.line([(40, 72), (W - 40, 72)], fill=C_AZUL, width=2)

# Límite del sistema
box_sys = [(310, 90), (W - 40, H - 35)]
draw.rectangle(box_sys, fill=C_FONDO_SYS, outline=C_BORDE_SYS, width=2)
draw.rectangle([(310, 90), (310 + 460, 125)], fill=C_AZUL)
draw.text((325, 98), "Sistema Autos Colombia (Módulo de Gestión)", fill="#FFFFFF", font=font_sys)

# Función para dibujar actor estilo UML
def draw_actor(x, y, label):
    # Cabeza
    draw.ellipse([(x - 14, y - 40), (x + 14, y - 12)], fill="#FFFFFF", outline=C_AZUL, width=2)
    # Tronco
    draw.line([(x, y - 12), (x, y + 16)], fill=C_AZUL, width=2)
    # Brazos
    draw.line([(x - 22, y - 2), (x + 22, y - 2)], fill=C_AZUL, width=2)
    # Piernas
    draw.line([(x, y + 16), (x - 18, y + 42)], fill=C_AZUL, width=2)
    # Pierna der
    draw.line([(x, y + 16), (x + 18, y + 42)], fill=C_AZUL, width=2)
    # Nombre
    bbox = font_actor.getbbox(label)
    lw = bbox[2] - bbox[0]
    draw.text((x - lw // 2, y + 48), label, fill=C_AZUL, font=font_actor)

# Función para dibujar caso de uso ovalado
def draw_usecase(cx, cy, rw, rh, title, subtitle=None, bg=C_CELESTE, border=C_BORDE_CU):
    draw.ellipse([(cx - rw, cy - rh), (cx + rw, cy + rh)], fill=bg, outline=border, width=2)
    if subtitle:
        b1 = font_cu_title.getbbox(title)
        w1 = b1[2] - b1[0]
        b2 = font_cu_sub.getbbox(subtitle)
        w2 = b2[2] - b2[0]
        draw.text((cx - w1 // 2, cy - 12), title, fill=C_TEXTO, font=font_cu_title)
        draw.text((cx - w2 // 2, cy + 4), subtitle, fill=C_TEXTO, font=font_cu_sub)
    else:
        b1 = font_cu_title.getbbox(title)
        w1 = b1[2] - b1[0]
        draw.text((cx - w1 // 2, cy - 6), title, fill=C_TEXTO, font=font_cu_title)

# Función para dibujar flecha discontinua <<include>>
def draw_include(p1, p2, label="<<include>>"):
    draw.line([p1, p2], fill=C_LINEA_INC, width=1)
    # Punto medio para etiqueta
    mx = (p1[0] + p2[0]) // 2
    my = (p1[1] + p2[1]) // 2
    draw.rectangle([(mx - 32, my - 8), (mx + 32, my + 8)], fill=C_FONDO_SYS)
    draw.text((mx - 28, my - 7), label, fill=C_LINEA_INC, font=font_inc)
    # Flecha punta
    angle = math.atan2(p2[1] - p1[1], p2[0] - p1[0])
    l_arrow = 10
    a1 = angle + math.pi - 0.4
    a2 = angle + math.pi + 0.4
    x3 = p2[0] + l_arrow * math.cos(a1)
    y3 = p2[1] + l_arrow * math.sin(a1)
    x4 = p2[0] + l_arrow * math.cos(a2)
    y4 = p2[1] + l_arrow * math.sin(a2)
    draw.polygon([p2, (x3, y3), (x4, y4)], fill=C_LINEA_INC)

# Posiciones de los actores
admin_pos = (130, 420)
enc_pos = (130, 780)

draw_actor(admin_pos[0], admin_pos[1], "Administrador")
draw_actor(enc_pos[0], enc_pos[1], "Encargado")

# Definición de Casos de Uso Principales (Columna Izquierda dentro del sistema)
col1_x = 510
ucs_col1 = [
    ("CU-06 Iniciar sesión", "(Auth y JWT)", 180, (col1_x, 180)),
    ("CU-07 Cambiar contraseña", "(Personal)", 180, (col1_x, 310)),
    ("CU-08 Administrar usuarios", "(Alta, roles, estado, reset)", 195, (col1_x, 460)),
    ("CU-09 Administrar celdas", "(Crear, inhabilitar, editar)", 190, (col1_x, 610)),
    ("CU-10 Administrar clientes", "(Vehículos y mensualidades)", 195, (col1_x, 760)),
    ("CU-11 Consultar auditoría", "(Bitácora inmutable)", 180, (col1_x, 900)),
]

# Dibujar CUs Columna 1
for title, sub, rw, (cx, cy) in ucs_col1:
    draw_usecase(cx, cy, rw, 32, title, sub)

# Definición de Casos de Uso Secundarios / Incluidos (Columna Derecha)
col2_x = 1140
ucs_col2 = [
    ("Validar credenciales y tasa", "(Máx 5 fallas/min)", 175, (col2_x, 180)),
    ("Validar robustez contraseña", "(10+ car, letra, num)", 175, (col2_x, 310)),
    ("Garantizar último admin", "(Mínimo 1 admin activo)", 175, (col2_x, 430)),
    ("Registrar auditoría inmutable", "(Trigger no modif / no borrado)", 180, (col2_x, 560)),
    ("Validar celda no ocupada", "(Para inhabilitar o editar)", 175, (col2_x, 680)),
    ("Validar no solapamiento fechas", "(Vigencia mensualidad)", 175, (col2_x, 800)),
    ("Validar vehículo fuera de patio", "(Para cambio de tipo)", 175, (col2_x, 915)),
]

# Dibujar CUs Columna 2
for title, sub, rw, (cx, cy) in ucs_col2:
    draw_usecase(cx, cy, rw, 30, title, sub, bg="#F1F5F9", border="#475569")

# --- CONEXIONES DE ACTORES ---
# Encargado
draw.line([(enc_pos[0] + 35, enc_pos[1] - 30), (col1_x - 180, 180)], fill=C_LINEA, width=2)
draw.line([(enc_pos[0] + 35, enc_pos[1] - 15), (col1_x - 180, 310)], fill=C_LINEA, width=2)
draw.line([(enc_pos[0] + 35, enc_pos[1] + 5), (col1_x - 195, 760)], fill=C_LINEA, width=2)

# Administrador
draw.line([(admin_pos[0] + 35, admin_pos[1] - 40), (col1_x - 180, 180)], fill=C_LINEA, width=2)
draw.line([(admin_pos[0] + 35, admin_pos[1] - 25), (col1_x - 180, 310)], fill=C_LINEA, width=2)
draw.line([(admin_pos[0] + 35, admin_pos[1] - 5), (col1_x - 195, 460)], fill=C_LINEA, width=2)
draw.line([(admin_pos[0] + 35, admin_pos[1] + 15), (col1_x - 190, 610)], fill=C_LINEA, width=2)
draw.line([(admin_pos[0] + 35, admin_pos[1] + 30), (col1_x - 195, 760)], fill=C_LINEA, width=2)
draw.line([(admin_pos[0] + 35, admin_pos[1] + 45), (col1_x - 180, 900)], fill=C_LINEA, width=2)

# --- RELACIONES <<include>> ---
# CU-06 -> Validar credenciales
draw_include((col1_x + 180, 180), (col2_x - 175, 180))
# CU-07 -> Validar robustez
draw_include((col1_x + 180, 310), (col2_x - 175, 310))
# CU-08 -> Validar robustez
draw_include((col1_x + 160, 440), (col2_x - 170, 330))
# CU-08 -> Garantizar último admin
draw_include((col1_x + 195, 460), (col2_x - 175, 435))
# CU-08 -> Registrar auditoría
draw_include((col1_x + 180, 480), (col2_x - 175, 545))
# CU-09 -> Validar celda no ocupada
draw_include((col1_x + 190, 620), (col2_x - 175, 675))
# CU-09 -> Registrar auditoría
draw_include((col1_x + 180, 595), (col2_x - 175, 570))
# CU-10 -> Validar no solapamiento
draw_include((col1_x + 195, 765), (col2_x - 175, 795))
# CU-10 -> Validar vehículo fuera
draw_include((col1_x + 180, 785), (col2_x - 175, 905))
# CU-10 -> Registrar auditoría
draw_include((col1_x + 170, 740), (col2_x - 160, 590))

# Guardar imagen PNG
output_png = r"c:\universidad\ingenieria de software\parqueadero_EA1_codigo\parqueadero\docs\diagramas\casos_uso_iteracion_2.png"
img.save(output_png, format="PNG", dpi=(300, 300))
print(f"Imagen PNG generada exitosamente: {output_png}")

