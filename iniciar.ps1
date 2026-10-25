# Iniciar Autos Colombia - Parqueadero
Write-Host "Iniciando Autos Colombia..." -ForegroundColor Cyan

# 1. Asegurar Docker PostgreSQL
Write-Host "[1/3] Levantando base de datos PostgreSQL en Docker..." -ForegroundColor Yellow
$env:API_TOKEN = "autos-colombia-token-2026"
docker compose up -d db

# 2. Compilar frontend si es necesario
if (-not (Test-Path "backend\static\index.html")) {
    Write-Host "[2/3] Compilando frontend..." -ForegroundColor Yellow
    Push-Location frontend
    npm install
    npm run build
    Pop-Location
    if (Test-Path "backend\static") { Remove-Item -Recurse -Force "backend\static" }
    Copy-Item -Recurse -Force "frontend\dist" "backend\static"
} else {
    Write-Host "[2/3] Frontend compilado listo en backend\static" -ForegroundColor Green
}

# 3. Iniciar Backend
Write-Host "[3/3] Iniciando servidor Go en http://localhost:8080..." -ForegroundColor Green
Write-Host "Clave de acceso (API_TOKEN): autos-colombia-token-2026" -ForegroundColor Cyan
Push-Location backend
$env:API_TOKEN = "autos-colombia-token-2026"
$env:DATABASE_URL = "postgres://postgres:postgres@localhost:5433/parqueadero?sslmode=disable"
$env:PORT = "8080"
$env:STATIC_DIR = "./static"
go run ./cmd/api
Pop-Location
