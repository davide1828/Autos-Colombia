# -*- coding: utf-8 -*-
"""
Generador del Diagrama de Componentes (Iteraciones 1 y 2)
1. PlantUML (.puml)
2. Renderizado PNG de alta resolución
3. Integración en el archivo StarUML (.mdj)
"""
from PIL import Image, ImageDraw, ImageFont
import json, uuid, os

def uid():
    return "AAAAA" + uuid.uuid4().hex[:19]

d_diag = r"c:\universidad\ingenieria de software\parqueadero_EA1_codigo\parqueadero\docs\diagramas"

# ==============================================================================
# 1. PLANTUML
# ==============================================================================
PUML_COMPONENTES = """@startuml diagrama_componentes
skinparam shadowing false
skinparam defaultFontName Arial
skinparam componentStyle uml2
skinparam packageBackgroundColor #F8FAFC
skinparam packageBorderColor #64748B
skinparam componentBackgroundColor #FFFFFF
skinparam componentBorderColor #0F4C81
skinparam interfaceBorderColor #0F4C81

title Diagrama de Componentes — Sistema de Información Autos Colombia (Iteraciones 1 y 2)

package "Frontend (Cliente PWA - React 18 + Vite)" {
  [Interfaz de Usuario (Vistas & Componentes)] as UI
  [Cliente API HTTP (Fetch & Auth)] as APIClient
  UI ..> APIClient : consume
}

cloud "Protocolo HTTP / HTTPS (JSON)" as HTTP

APIClient -- HTTP

package "Backend Go (Arquitectura Hexagonal)" {
  
  package "Capa de Adaptadores Primarios (Transporte HTTP)" {
    [HTTP Router & Middleware] as Router
    [Controladores Operativos (EA1)] as CtrlEA1
    [Controladores Gestión & Auth (EA2)] as CtrlEA2
    
    HTTP -- Router
    Router --> CtrlEA1
    Router --> CtrlEA2
  }

  package "Capa de Aplicación (Casos de Uso)" {
    interface "ServicioOperaciones" as ISvcOp
    interface "ServicioGestion" as ISvcGes
    
    [Servicio Ingreso y Salida] as SvcEA1
    [Servicio Gestión y Usuarios] as SvcEA2
    
    ISvcOp - SvcEA1
    ISvcGes - SvcEA2
    
    CtrlEA1 ..> ISvcOp : invoca
    CtrlEA2 ..> ISvcGes : invoca
  }

  package "Capa de Dominio (Reglas de Negocio)" {
    [Entidades y Validaciones de Negocio] as Dominio
    SvcEA1 ..> Dominio : valida RN-01 a RN-06
    SvcEA2 ..> Dominio : valida RN-07 a RN-11
  }

  package "Capa de Seguridad & Criptografía" {
    interface "Hasheador" as IHash
    [Módulo PBKDF2 (HMAC-SHA256)] as SecPBKDF2
    [Firmador de Tokens de Sesión] as SecToken
    
    IHash - SecPBKDF2
    SvcEA2 ..> IHash : hash / verificación
    CtrlEA2 ..> SecToken : emite / valida
  }

  package "Capa de Adaptadores Secundarios (Persistencia)" {
    interface "RepositorioOperaciones" as IRepoOp
    interface "RepositorioGestion" as IRepoGes
    
    [Adaptador PostgreSQL] as RepoPG
    [Adaptador Memoria (Testing)] as RepoMem
    
    IRepoOp - RepoPG
    IRepoGes - RepoPG
    
    IRepoOp -- RepoMem
    IRepoGes -- RepoMem
    
    SvcEA1 ..> IRepoOp : persiste
    SvcEA2 ..> IRepoGes : persiste
  }
}

database "PostgreSQL 16 (Motor de Base de Datos)" {
  [Tablas Operativas (vehiculo, celda, registro, novedad)] as TablasEA1
  [Tablas Gestión (usuario, auditoria, cliente, mensualidad)] as TablasEA2
  [Triggers de Integridad (trg_registro_tipos, trg_auditoria_inmutable)] as Triggers
  
  RepoPG --> TablasEA1
  RepoPG --> TablasEA2
  TablasEA1 ..> Triggers : ejecuta
  TablasEA2 ..> Triggers : ejecuta
}
@enduml
"""

with open(os.path.join(d_diag, "13_diagrama_componentes.puml"), "w", encoding="utf-8") as f:
    f.write(PUML_COMPONENTES)

# ==============================================================================
# 2. GENERADOR DE IMAGEN PNG DEL DIAGRAMA DE COMPONENTES
# ==============================================================================
W, H = 1350, 920
im = Image.new("RGB", (W, H), "#FFFFFF")
d = ImageDraw.Draw(im)

FONT_TITLE = ImageFont.truetype(r"C:\Windows\Fonts\arialbd.ttf", 20)
FONT_SUB = ImageFont.truetype(r"C:\Windows\Fonts\ariali.ttf", 12)
FONT_PKG = ImageFont.truetype(r"C:\Windows\Fonts\arialbd.ttf", 13)
FONT_COMP_BOLD = ImageFont.truetype(r"C:\Windows\Fonts\arialbd.ttf", 11)
FONT_COMP_TEXT = ImageFont.truetype(r"C:\Windows\Fonts\arial.ttf", 10)
FONT_INTF = ImageFont.truetype(r"C:\Windows\Fonts\ariali.ttf", 10)

C_AZUL = "#0F4C81"
C_PKG_BG = "#F8FAFC"
C_PKG_BORDER = "#94A3B8"
C_COMP_BG = "#FFFFFF"
C_COMP_BORDER = "#0F4C81"
C_TEXT = "#1C2530"

# Encabezado
d.text((40, 20), "Diagrama de Componentes — Sistema de Información Autos Colombia (Iteraciones 1 y 2)", fill=C_AZUL, font=FONT_TITLE)
d.text((40, 48), "Arquitectura Hexagonal desacoplada: Frontend PWA, API REST Go, Dominio, Seguridad PBKDF2 y PostgreSQL", fill="#64748B", font=FONT_SUB)
d.line([(40, 68), (W - 40, 68)], fill=C_AZUL, width=2)

def draw_component(x, y, w, h, title, subtitle=None):
    # Caja principal del componente
    d.rectangle([(x, y), (x + w, y + h)], fill=C_COMP_BG, outline=C_COMP_BORDER, width=2)
    # Pestañas UML del componente en la esquina superior izquierda
    d.rectangle([(x - 8, y + 8), (x + 2, y + 18)], fill=C_COMP_BG, outline=C_COMP_BORDER, width=1)
    d.rectangle([(x - 8, y + 24), (x + 2, y + 34)], fill=C_COMP_BG, outline=C_COMP_BORDER, width=1)
    
    # Texto
    if subtitle:
        d.text((x + 12, y + 10), title, fill=C_AZUL, font=FONT_COMP_BOLD)
        d.text((x + 12, y + 26), subtitle, fill="#64748B", font=FONT_COMP_TEXT)
    else:
        d.text((x + 12, y + 18), title, fill=C_AZUL, font=FONT_COMP_BOLD)

def draw_package(x, y, w, h, title, bg=C_PKG_BG, border=C_PKG_BORDER):
    # Pestaña carpeta
    d.rectangle([(x, y - 22), (x + 180, y)], fill=bg, outline=border, width=1)
    d.text((x + 8, y - 18), title[:25], fill="#475569", font=FONT_COMP_BOLD)
    # Cuerpo carpeta
    d.rectangle([(x, y), (x + w, y + h)], fill=bg, outline=border, width=1)
    d.text((x + 10, y + 6), title, fill="#334155", font=FONT_PKG)

# 1. FRONTEND PACKAGE
draw_package(40, 110, 320, 240, "Frontend (PWA React + Vite)")
draw_component(65, 145, 270, 50, "Interfaz de Usuario", "Vistas Ingreso, Salida, Usuarios, Celdas")
draw_component(65, 215, 270, 50, "Componentes Reutilizables", "PlacaInput, CeldaSelector, NovedadCampo")
draw_component(65, 285, 270, 50, "Cliente API HTTP", "api.js (Fetch + Bearer Token)")

# FLECHA FRONTEND -> HTTP
d.line([(360, 310), (420, 310)], fill=C_AZUL, width=2)
d.polygon([(420, 310), (410, 305), (410, 315)], fill=C_AZUL)
d.text((368, 290), "HTTP/JSON", fill=C_AZUL, font=FONT_COMP_TEXT)

# 2. BACKEND HEXAGONAL PACKAGE
draw_package(430, 110, 880, 530, "Backend Go (Arquitectura Hexagonal)")

# Sub-paquete HTTP Adapters
draw_package(450, 150, 400, 140, "Adaptadores Primarios (HTTP API)", bg="#FFFFFF", border="#CBD5E1")
draw_component(470, 185, 175, 45, "HTTP Router & Auth", "Mux Go 1.22 + Middleware")
draw_component(660, 185, 175, 45, "Controladores Operativos", "Ingreso, Salida, Novedades")
draw_component(470, 238, 365, 42, "Controladores Gestión & Seguridad", "Sesiones, Usuarios, Celdas, Auditoría")

# Sub-paquete Aplicación
draw_package(450, 320, 400, 140, "Capa de Aplicación (Casos de Uso)", bg="#FFFFFF", border="#CBD5E1")
draw_component(470, 355, 365, 45, "Servicio IngresoSalida (EA1)", "Casos de uso operativos y tiempo estancia")
draw_component(470, 408, 365, 45, "Servicio Gestión (EA2)", "Usuarios, RBAC, celdas, clientes, auditoría")

# Sub-paquete Dominio
draw_package(880, 150, 410, 140, "Capa de Dominio (Reglas de Negocio)", bg="#FFFFFF", border="#CBD5E1")
draw_component(905, 185, 360, 45, "Entidades & Validaciones", "Placas, tipos compatibles, clave segura")
draw_component(905, 238, 360, 42, "Reglas de Negocio Puras", "RN-01 a RN-11 (independientes de BD)")

# Sub-paquete Seguridad
draw_package(880, 320, 410, 140, "Capa de Seguridad & Criptografía", bg="#FFFFFF", border="#CBD5E1")
draw_component(905, 355, 360, 45, "Módulo PBKDF2 (HMAC-SHA256)", "16 bytes salt, 100k iteraciones, subtle.Compare")
draw_component(905, 408, 360, 45, "Firmador de Sesión", "Emisión y revocación inmediata de tokens")

# Sub-paquete Persistencia
draw_package(450, 490, 840, 135, "Capa de Persistencia (Adaptadores Secundarios)", bg="#FFFFFF", border="#CBD5E1")
draw_component(470, 525, 390, 80, "Adaptador PostgreSQL (Producción)", "Consultas SQL parametrizadas, transacciones ACID\\nControl de concurrencia FOR SHARE / FOR UPDATE")
draw_component(890, 525, 380, 80, "Adaptador Memoria (Testing)", "Implementación con exclusión mutua (sync.RWMutex)\\nPruebas automáticas unitarias y de integración")

# 3. BASE DE DATOS POSTGRESQL PACKAGE
draw_package(40, 670, 1270, 210, "Base de Datos Relacional (PostgreSQL 16 Engine)")
draw_component(70, 715, 360, 60, "Tablas Operativas (EA1)", "vehiculo, celda, registro, novedad")
draw_component(460, 715, 380, 60, "Tablas Gestión & Seguridad (EA2)", "usuario, auditoria, cliente, mensualidad")
draw_component(870, 715, 410, 60, "Triggers Transaccionales de Motor", "trg_registro_tipos (compatibilidad tipo)\\ntrg_auditoria_inmutable (inviolabilidad bitácora)")

# Resumen de conexiones visuales
d.line([(650, 290), (650, 320)], fill=C_AZUL, width=2) # HTTP -> App
d.line([(850, 390), (880, 390)], fill=C_AZUL, width=2) # App -> Seg
d.line([(850, 220), (880, 220)], fill=C_AZUL, width=2) # App -> Dom
d.line([(650, 460), (650, 490)], fill=C_AZUL, width=2) # App -> Persistencia
d.line([(665, 625), (665, 670)], fill=C_AZUL, width=2) # Persistencia -> DB

png_comp = os.path.join(d_diag, "diagrama_componentes.png")
im.save(png_comp, dpi=(300, 300))
print("Diagrama de componentes generado:", png_comp)

# ==============================================================================
# 3. ACTUALIZAR STARUML (.mdj) CON COMPONENTES
# ==============================================================================
mdj_path = os.path.join(d_diag, "parqueadero_iteracion_2.mdj")
with open(mdj_path, "r", encoding="utf-8") as f:
    star_data = json.load(f)

comp_model_id = uid()
comp_diagram_id = uid()

components_json = {
    "_type": "UMLModel",
    "_id": comp_model_id,
    "_parent": {"$ref": star_data["ownedElements"][0]["_id"]},
    "name": "Modelo de Componentes del Sistema",
    "ownedElements": [
        {"_type": "UMLComponent", "_id": uid(), "name": "Frontend PWA (React + Vite)"},
        {"_type": "UMLComponent", "_id": uid(), "name": "HTTP Router & Middleware"},
        {"_type": "UMLComponent", "_id": uid(), "name": "Controladores API (EA1 & EA2)"},
        {"_type": "UMLComponent", "_id": uid(), "name": "Servicio de Aplicación (Operaciones & Gestión)"},
        {"_type": "UMLComponent", "_id": uid(), "name": "Dominio (Reglas de Negocio)"},
        {"_type": "UMLComponent", "_id": uid(), "name": "Seguridad PBKDF2 & Token"},
        {"_type": "UMLComponent", "_id": uid(), "name": "Adaptador Persistencia (PostgreSQL & Memoria)"},
        {"_type": "UMLComponent", "_id": uid(), "name": "Base de Datos PostgreSQL 16 & Triggers"},
        {
            "_type": "UMLComponentDiagram",
            "_id": comp_diagram_id,
            "_parent": {"$ref": comp_model_id},
            "name": "Diagrama de Componentes - Autos Colombia (Iteraciones 1 y 2)"
        }
    ]
}

star_data["ownedElements"][0]["ownedElements"].append(components_json)

with open(mdj_path, "w", encoding="utf-8") as f:
    json.dump(star_data, f, indent=2, ensure_ascii=False)

print("Archivo StarUML actualizado con diagrama de componentes.")

