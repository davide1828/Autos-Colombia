# -*- coding: utf-8 -*-
"""
Generador de Mockups de Alta Fidelidad para la Iteración 2
(Gestión de Usuarios, Gestión de Celdas y Auditoría)
Paleta de diseño: Azul #0F4C81, Asfalto #1C2530, Fondo #F8FAFC, Bordes #E2E8F0
"""
from PIL import Image, ImageDraw, ImageFont
import os

FONT_TITLE = ImageFont.truetype(r"C:\Windows\Fonts\arialbd.ttf", 20)
FONT_SUB = ImageFont.truetype(r"C:\Windows\Fonts\arial.ttf", 13)
FONT_H2 = ImageFont.truetype(r"C:\Windows\Fonts\arialbd.ttf", 15)
FONT_BOLD = ImageFont.truetype(r"C:\Windows\Fonts\arialbd.ttf", 12)
FONT_TEXT = ImageFont.truetype(r"C:\Windows\Fonts\arial.ttf", 12)
FONT_SMALL = ImageFont.truetype(r"C:\Windows\Fonts\arial.ttf", 10)
FONT_BADGE = ImageFont.truetype(r"C:\Windows\Fonts\arialbd.ttf", 10)

C_PRIMARY = "#0F4C81"
C_PRIMARY_LIGHT = "#E8F1F8"
C_TEXT = "#1C2530"
C_MUTED = "#64748B"
C_BG = "#F8FAFC"
C_WHITE = "#FFFFFF"
C_BORDER = "#CBD5E1"
C_BORDER_LIGHT = "#E2E8F0"

C_GREEN_BG = "#DCFCE7"
C_GREEN_TXT = "#15803D"
C_RED_BG = "#FEE2E2"
C_RED_TXT = "#B91C1C"
C_BLUE_BG = "#E0F2FE"
C_BLUE_TXT = "#0369A1"
C_AMBER_BG = "#FEF3C7"
C_AMBER_TXT = "#B45309"

def draw_header(draw, W, user_name="Cristian España", user_role="ADMINISTRADOR", active_tab="usuarios"):
    # Barra superior
    draw.rectangle([(0, 0), (W, 64)], fill=C_PRIMARY)
    draw.text((25, 20), "AUTOS COLOMBIA", fill=C_WHITE, font=FONT_TITLE)
    draw.text((220, 26), "· Sistema de Gestión", fill=C_PRIMARY_LIGHT, font=FONT_SUB)
    
    # Usuario info derecha
    u_info = f"{user_name}  [{user_role}]"
    b_u = FONT_BOLD.getbbox(u_info)
    uw = b_u[2] - b_u[0]
    draw.text((W - uw - 40, 24), u_info, fill=C_WHITE, font=FONT_BOLD)
    
    # Barra de navegación (Pestañas)
    draw.rectangle([(0, 64), (W, 108)], fill=C_WHITE, outline=C_BORDER_LIGHT, width=1)
    tabs = [
        ("usuarios", "Gestión de Usuarios"),
        ("celdas", "Gestión de Celdas"),
        ("clientes", "Clientes y Mensualidades"),
        ("patio", "Patio Operativo (EA1)"),
        ("auditoria", "Bitácora de Auditoría")
    ]
    tx = 25
    for tab_id, tab_label in tabs:
        bbox = FONT_BOLD.getbbox(tab_label) if tab_id == active_tab else FONT_SUB.getbbox(tab_label)
        tw = bbox[2] - bbox[0]
        if tab_id == active_tab:
            draw.rectangle([(tx - 10, 68), (tx + tw + 10, 104)], fill=C_PRIMARY_LIGHT)
            draw.line([(tx - 10, 106), (tx + tw + 10, 106)], fill=C_PRIMARY, width=3)
            draw.text((tx, 78), tab_label, fill=C_PRIMARY, font=FONT_BOLD)
        else:
            draw.text((tx, 78), tab_label, fill=C_MUTED, font=FONT_SUB)
        tx += tw + 35

def draw_badge(draw, x, y, text, bg, fg):
    bbox = FONT_BADGE.getbbox(text)
    w = bbox[2] - bbox[0] + 16
    draw.rounded_rectangle([(x, y), (x + w, y + 20)], radius=10, fill=bg)
    draw.text((x + 8, y + 4), text, fill=fg, font=FONT_BADGE)
    return w

# ==============================================================================
# MOCKUP 1: GESTIÓN DE USUARIOS
# ==============================================================================
def make_mockup_usuarios():
    W, H = 1100, 720
    im = Image.new("RGB", (W, H), C_BG)
    d = ImageDraw.Draw(im)
    draw_header(d, W, active_tab="usuarios")
    
    # Contenedor principal
    d.rectangle([(25, 125), (W - 25, H - 25)], fill=C_WHITE, outline=C_BORDER, width=1)
    
    # Título y botón Nuevo
    d.text((45, 145), "Usuarios del Sistema (RBAC)", fill=C_TEXT, font=FONT_TITLE)
    d.text((45, 175), "Administre cuentas de acceso, roles asignados y estados de los operadores.", fill=C_MUTED, font=FONT_SUB)
    
    # Botón '+ Nuevo Usuario'
    d.rounded_rectangle([(W - 215, 145), (W - 45, 185)], radius=6, fill=C_PRIMARY)
    d.text((W - 195, 155), "+ Nuevo Usuario", fill=C_WHITE, font=FONT_BOLD)
    
    # Filtros / Buscador
    d.rounded_rectangle([(45, 215), (350, 250)], radius=4, fill=C_WHITE, outline=C_BORDER)
    d.text((58, 225), "Buscar por usuario o nombre...", fill=C_MUTED, font=FONT_TEXT)
    
    d.rounded_rectangle([(365, 215), (490, 250)], radius=4, fill=C_WHITE, outline=C_BORDER)
    d.text((380, 225), "Rol: Todos ▼", fill=C_TEXT, font=FONT_TEXT)
    
    d.rounded_rectangle([(505, 215), (630, 250)], radius=4, fill=C_WHITE, outline=C_BORDER)
    d.text((520, 225), "Estado: Activos ▼", fill=C_TEXT, font=FONT_TEXT)
    
    # Encabezado Tabla
    d.rectangle([(45, 270), (W - 45, 305)], fill=C_PRIMARY_LIGHT)
    cols = [(65, "USUARIO"), (200, "NOMBRE COMPLETO"), (460, "ROL ASIGNADO"), (620, "ESTADO"), (750, "CREACIÓN"), (900, "ACCIONES")]
    for x, txt in cols:
        d.text((x, 280), txt, fill=C_PRIMARY, font=FONT_BOLD)
        
    # Filas de la tabla
    filas = [
        ("admin", "Cristian David España", "ADMINISTRADOR", "ACTIVO", "2026-10-20", ["Editar", "Clave", "Inactivar (Bloqueado: Último Admin)"]),
        ("emolina", "Esteban Montoya Molina", "ADMINISTRADOR", "ACTIVO", "2026-10-22", ["Editar", "Clave", "Inactivar"]),
        ("carlos.perez", "Carlos Andrés Pérez", "ENCARGADO", "ACTIVO", "2026-10-24", ["Editar", "Clave", "Inactivar"]),
        ("luisa.gomez", "Luisa Fernanda Gómez", "ENCARGADO", "ACTIVO", "2026-10-25", ["Editar", "Clave", "Inactivar"]),
        ("jorge.ruiz", "Jorge Hernán Ruiz", "ENCARGADO", "INACTIVO", "2026-10-28", ["Editar", "Clave", "Activar"]),
    ]
    
    y = 315
    for u, nom, rol, est, fec, acts in filas:
        d.line([(45, y + 42), (W - 45, y + 42)], fill=C_BORDER_LIGHT, width=1)
        d.text((65, y + 12), u, fill=C_TEXT, font=FONT_BOLD)
        d.text((200, y + 12), nom, fill=C_TEXT, font=FONT_TEXT)
        
        # Badge Rol
        if rol == "ADMINISTRADOR":
            draw_badge(d, 460, y + 10, rol, C_BLUE_BG, C_BLUE_TXT)
        else:
            draw_badge(d, 460, y + 10, rol, C_PRIMARY_LIGHT, C_PRIMARY)
            
        # Badge Estado
        if est == "ACTIVO":
            draw_badge(d, 620, y + 10, est, C_GREEN_BG, C_GREEN_TXT)
        else:
            draw_badge(d, 620, y + 10, est, C_RED_BG, C_RED_TXT)
            
        d.text((750, y + 12), fec, fill=C_MUTED, font=FONT_TEXT)
        
        # Acciones
        d.text((900, y + 12), "Editar  ·  Clave", fill=C_PRIMARY, font=FONT_BOLD)
        if "Inactivar (Bloqueado" in acts[2]:
            d.text((980, y + 12), "· [Protegido]", fill=C_MUTED, font=FONT_SMALL)
        elif "Inactivar" in acts[2]:
            d.text((980, y + 12), "· Inactivar", fill=C_RED_TXT, font=FONT_TEXT)
        else:
            d.text((980, y + 12), "· Activar", fill=C_GREEN_TXT, font=FONT_TEXT)
            
        y += 45

    # Mensaje de ayuda inferior
    d.rounded_rectangle([(45, y + 30), (W - 45, y + 75)], radius=6, fill=C_PRIMARY_LIGHT, outline=C_PRIMARY)
    d.text((60, y + 42), "Regla RN-08: El sistema impide desactivar o degradar al último Administrador activo para asegurar la gobernabilidad.", fill=C_PRIMARY, font=FONT_TEXT)
    
    out = r"c:\universidad\ingenieria de software\parqueadero_EA1_codigo\parqueadero\docs\pantallas\11_admin_usuarios.png"
    im.save(out, dpi=(300, 300))
    print("Generado:", out)

# ==============================================================================
# MOCKUP 2: MODAL CREAR / EDITAR USUARIO
# ==============================================================================
def make_mockup_modal_usuario():
    W, H = 800, 650
    im = Image.new("RGB", (W, H), C_BG)
    d = ImageDraw.Draw(im)
    
    # Fondo modal
    d.rectangle([(0, 0), (W, H)], fill="#00000033") # Sombra simulada
    d.rounded_rectangle([(100, 40), (700, 600)], radius=8, fill=C_WHITE, outline=C_BORDER, width=2)
    
    # Cabecera modal
    d.rectangle([(100, 40), (700, 100)], fill=C_PRIMARY)
    d.text((130, 58), "Nuevo Usuario del Sistema", fill=C_WHITE, font=FONT_TITLE)
    d.text((660, 55), "X", fill=C_WHITE, font=FONT_TITLE)
    
    # Campo Nombre
    d.text((130, 120), "Nombre Completo *", fill=C_TEXT, font=FONT_BOLD)
    d.rounded_rectangle([(130, 142), (670, 180)], radius=4, fill=C_WHITE, outline=C_BORDER)
    d.text((145, 153), "Santiago Morales Restrepo", fill=C_TEXT, font=FONT_TEXT)
    
    # Campo Login (Usuario)
    d.text((130, 195), "Nombre de Usuario (Login) *", fill=C_TEXT, font=FONT_BOLD)
    d.text((310, 197), "(3-30 caracteres, minúsculas, números, punto o guion)", fill=C_MUTED, font=FONT_SMALL)
    d.rounded_rectangle([(130, 217), (670, 255)], radius=4, fill=C_WHITE, outline=C_BORDER)
    d.text((145, 228), "smorales", fill=C_TEXT, font=FONT_TEXT)
    
    # Campo Rol
    d.text((130, 270), "Rol en el Sistema *", fill=C_TEXT, font=FONT_BOLD)
    # Radiobuttons
    d.ellipse([(135, 298), (150, 313)], outline=C_PRIMARY, width=2)
    d.ellipse([(138, 301), (147, 310)], fill=C_PRIMARY)
    d.text((160, 298), "ENCARGADO  (Operación de patio, entradas, salidas y novedades)", fill=C_TEXT, font=FONT_TEXT)
    
    d.ellipse([(135, 330), (150, 345)], outline=C_BORDER, width=2)
    d.text((160, 330), "ADMINISTRADOR  (Acceso total: gestión de usuarios, celdas y auditoría)", fill=C_TEXT, font=FONT_TEXT)
    
    # Campo Contraseña
    d.text((130, 375), "Contraseña Inicial *", fill=C_TEXT, font=FONT_BOLD)
    d.text((260, 377), "(Mínimo 10 caracteres, al menos una letra y un número)", fill=C_MUTED, font=FONT_SMALL)
    d.rounded_rectangle([(130, 397), (670, 435)], radius=4, fill=C_WHITE, outline=C_BORDER)
    d.text((145, 408), "••••••••••••••  [ContraseñaSegura2026]", fill=C_TEXT, font=FONT_TEXT)
    
    # Indicador de robustez
    d.rounded_rectangle([(130, 445), (350, 452)], radius=3, fill=C_GREEN_TXT)
    d.text((360, 442), "Fortaleza: Fuerte (Cumple política RN-07)", fill=C_GREEN_TXT, font=FONT_SMALL)
    
    # Checkbox Activo
    d.rectangle([(132, 480), (148, 496)], outline=C_PRIMARY, width=2, fill=C_PRIMARY)
    d.text((136, 481), "✓", fill=C_WHITE, font=FONT_BOLD)
    d.text((160, 481), "Usuario activo para iniciar sesión inmediatamente", fill=C_TEXT, font=FONT_TEXT)
    
    # Botones acción
    d.rounded_rectangle([(440, 535), (540, 575)], radius=6, fill=C_WHITE, outline=C_BORDER)
    d.text((460, 547), "Cancelar", fill=C_MUTED, font=FONT_BOLD)
    
    d.rounded_rectangle([(555, 535), (670, 575)], radius=6, fill=C_PRIMARY)
    d.text((575, 547), "Guardar Usuario", fill=C_WHITE, font=FONT_BOLD)
    
    out = r"c:\universidad\ingenieria de software\parqueadero_EA1_codigo\parqueadero\docs\pantallas\12_modal_usuario.png"
    im.save(out, dpi=(300, 300))
    print("Generado:", out)

# ==============================================================================
# MOCKUP 3: GESTIÓN DE CELDAS
# ==============================================================================
def make_mockup_celdas():
    W, H = 1100, 720
    im = Image.new("RGB", (W, H), C_BG)
    d = ImageDraw.Draw(im)
    draw_header(d, W, active_tab="celdas")
    
    # Contenedor principal
    d.rectangle([(25, 125), (W - 25, H - 25)], fill=C_WHITE, outline=C_BORDER, width=1)
    
    # Título y botón
    d.text((45, 145), "Gestión de Celdas y Bahías", fill=C_TEXT, font=FONT_TITLE)
    d.text((45, 175), "Configure la capacidad del patio, asigne tipos (CARRO/MOTO) e inhabilite bahías en obra.", fill=C_MUTED, font=FONT_SUB)
    
    d.rounded_rectangle([(W - 200, 145), (W - 45, 185)], radius=6, fill=C_PRIMARY)
    d.text((W - 180, 155), "+ Registrar Celda", fill=C_WHITE, font=FONT_BOLD)
    
    # Resumen contadores
    tarjetas = [
        ("Total Celdas", "20", C_BLUE_BG, C_BLUE_TXT),
        ("Disponibles", "14", C_GREEN_BG, C_GREEN_TXT),
        ("Ocupadas", "4", C_AMBER_BG, C_AMBER_TXT),
        ("Inactivas (Mantenimiento)", "2", C_RED_BG, C_RED_TXT),
    ]
    cx = 45
    for t_tit, t_num, bg, fg in tarjetas:
        d.rounded_rectangle([(cx, 215), (cx + 225, 275)], radius=6, fill=bg, outline=fg, width=1)
        d.text((cx + 15, 225), t_tit, fill=fg, font=FONT_BOLD)
        d.text((cx + 15, 245), t_num, fill=fg, font=FONT_TITLE)
        cx += 245
        
    # Encabezado Tabla
    d.rectangle([(45, 295), (W - 45, 330)], fill=C_PRIMARY_LIGHT)
    cols = [(65, "CÓDIGO"), (180, "TIPO VEHÍCULO"), (340, "ESTADO ACTUAL"), (520, "VEHÍCULO ALOJADO"), (720, "CONDICIÓN / NOTA"), (920, "GESTIÓN DE ESTADO")]
    for x, txt in cols:
        d.text((x, 305), txt, fill=C_PRIMARY, font=FONT_BOLD)
        
    # Filas celdas
    filas_c = [
        ("A-01", "CARRO", "DISPONIBLE", "— (Libre)", "Apta para parqueo", "Inhabilitar"),
        ("A-02", "CARRO", "OCUPADA", "AAA-123 (Mazda 3)", "En estancia activa", "Bloqueado (Ocupada)"),
        ("A-03", "CARRO", "INACTIVA", "—", "Mantenimiento demarcación", "Habilitar"),
        ("M-01", "MOTO", "DISPONIBLE", "— (Libre)", "Apta para parqueo", "Inhabilitar"),
        ("M-02", "MOTO", "OCUPADA", "XYZ-89E (Yamaha R3)", "En estancia activa", "Bloqueado (Ocupada)"),
        ("M-03", "MOTO", "INACTIVA", "—", "Fuga de aceite en piso", "Habilitar"),
    ]
    
    y = 340
    for cod, tip, est, veh, nota, accion in filas_c:
        d.line([(45, y + 42), (W - 45, y + 42)], fill=C_BORDER_LIGHT, width=1)
        d.text((65, y + 12), cod, fill=C_TEXT, font=FONT_BOLD)
        
        # Tipo badge
        if tip == "CARRO":
            draw_badge(d, 180, y + 10, "🚗 CARRO", C_PRIMARY_LIGHT, C_PRIMARY)
        else:
            draw_badge(d, 180, y + 10, "🏍️ MOTO", C_BLUE_BG, C_BLUE_TXT)
            
        # Estado badge
        if est == "DISPONIBLE":
            draw_badge(d, 340, y + 10, "DISPONIBLE", C_GREEN_BG, C_GREEN_TXT)
        elif est == "OCUPADA":
            draw_badge(d, 340, y + 10, "OCUPADA", C_AMBER_BG, C_AMBER_TXT)
        else:
            draw_badge(d, 340, y + 10, "INACTIVA", C_RED_BG, C_RED_TXT)
            
        d.text((520, y + 12), veh, fill=C_TEXT, font=FONT_TEXT)
        d.text((720, y + 12), nota, fill=C_MUTED, font=FONT_TEXT)
        
        # Botón de acción
        if accion == "Inhabilitar":
            d.rounded_rectangle([(920, y + 8), (1020, y + 32)], radius=4, fill=C_RED_BG, outline=C_RED_TXT)
            d.text((932, y + 13), "Inhabilitar", fill=C_RED_TXT, font=FONT_BOLD)
        elif accion == "Habilitar":
            d.rounded_rectangle([(920, y + 8), (1020, y + 32)], radius=4, fill=C_GREEN_BG, outline=C_GREEN_TXT)
            d.text((935, y + 13), "Habilitar", fill=C_GREEN_TXT, font=FONT_BOLD)
        else:
            d.text((920, y + 13), "🔒 Bloqueado (Ocupada)", fill=C_MUTED, font=FONT_SMALL)
            
        y += 45
        
    # Mensaje regla
    d.rounded_rectangle([(45, y + 25), (W - 45, y + 70)], radius=6, fill=C_PRIMARY_LIGHT, outline=C_PRIMARY)
    d.text((60, y + 37), "Regla RN-09: No es posible inhabilitar ni modificar celdas que tengan un vehículo dentro. Registre primero la salida.", fill=C_PRIMARY, font=FONT_TEXT)
    
    out = r"c:\universidad\ingenieria de software\parqueadero_EA1_codigo\parqueadero\docs\pantallas\14_admin_celdas.png"
    im.save(out, dpi=(300, 300))
    print("Generado:", out)

# ==============================================================================
# MOCKUP 4: AUDITORÍA INMUTABLE
# ==============================================================================
def make_mockup_auditoria():
    W, H = 1100, 720
    im = Image.new("RGB", (W, H), C_BG)
    d = ImageDraw.Draw(im)
    draw_header(d, W, active_tab="auditoria")
    
    # Contenedor principal
    d.rectangle([(25, 125), (W - 25, H - 25)], fill=C_WHITE, outline=C_BORDER, width=1)
    
    d.text((45, 145), "Bitácora de Auditoría del Sistema", fill=C_TEXT, font=FONT_TITLE)
    d.text((45, 175), "Registro inalterable de operaciones administrativas y modificaciones de datos (RN-11).", fill=C_MUTED, font=FONT_SUB)
    
    # Encabezado Tabla
    d.rectangle([(45, 220), (W - 45, 255)], fill=C_PRIMARY_LIGHT)
    cols = [(65, "FECHA / HORA UTC"), (240, "USUARIO"), (370, "ACCIÓN"), (520, "ENTIDAD"), (640, "ID"), (720, "DETALLE DE LA OPERACIÓN")]
    for x, txt in cols:
        d.text((x, 230), txt, fill=C_PRIMARY, font=FONT_BOLD)
        
    logs = [
        ("2026-11-08 17:45:10", "admin", "INHABILITAR_CELDA", "CELDA", "3", "Celda A-03 pasada a INACTIVA por mantenimiento."),
        ("2026-11-08 16:30:22", "admin", "CREAR_USUARIO", "USUARIO", "5", "Creación de usuario 'luisa.gomez' con rol ENCARGADO."),
        ("2026-11-08 15:12:05", "admin", "CAMBIAR_CLAVE", "USUARIO", "4", "Restablecimiento de contraseña para 'carlos.perez'."),
        ("2026-11-08 14:05:40", "admin", "RENOVAR_MENSUALIDAD", "MENSUALIDAD", "18", "Cliente 10203040: vigencia 2026-11-01 a 2026-11-30."),
        ("2026-11-08 12:00:18", "emolina", "CREAR_CELDA", "CELDA", "21", "Nueva celda 'M-05' tipo MOTO registrada en estado DISPONIBLE."),
    ]
    
    y = 265
    for fec, usu, acc, ent, ide, det in logs:
        d.line([(45, y + 42), (W - 45, y + 42)], fill=C_BORDER_LIGHT, width=1)
        d.text((65, y + 12), fec, fill=C_MUTED, font=FONT_TEXT)
        d.text((240, y + 12), usu, fill=C_TEXT, font=FONT_BOLD)
        
        # Badge Acción
        draw_badge(d, 370, y + 10, acc, C_BLUE_BG, C_BLUE_TXT)
        d.text((520, y + 12), ent, fill=C_TEXT, font=FONT_TEXT)
        d.text((640, y + 12), ide, fill=C_MUTED, font=FONT_TEXT)
        d.text((720, y + 12), det, fill=C_TEXT, font=FONT_TEXT)
        y += 45
        
    # Mensaje regla
    d.rounded_rectangle([(45, y + 30), (W - 45, y + 75)], radius=6, fill=C_PRIMARY_LIGHT, outline=C_PRIMARY)
    d.text((60, y + 42), "Protección por Trigger PostgreSQL: Cualquier intento de UPDATE o DELETE sobre esta tabla es rechazado inmediatamente.", fill=C_PRIMARY, font=FONT_TEXT)
    
    out = r"c:\universidad\ingenieria de software\parqueadero_EA1_codigo\parqueadero\docs\pantallas\15_admin_auditoria.png"
    im.save(out, dpi=(300, 300))
    print("Generado:", out)

if __name__ == "__main__":
    make_mockup_usuarios()
    make_mockup_modal_usuario()
    make_mockup_celdas()
    make_mockup_auditoria()
    print("Todos los mockups fueron generados exitosamente.")

