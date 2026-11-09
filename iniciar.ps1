# Iniciar Autos Colombia - Parqueadero
Write-Host "Iniciando Autos Colombia..." -ForegroundColor Cyan

# 0. Cargar variables del .env
if (Test-Path ".env") {
    Get-Content ".env" | Where-Object { $_ -notmatch "^\s*#" -and $_ -match "=" } | ForEach-Object {
        $parts = $_ -split "=", 2
        [System.Environment]::SetEnvironmentVariable($parts[0].Trim(), $parts[1].Trim(), "Process")
    }
    Write-Host "[0/3] Variables de entorno cargadas desde .env" -ForegroundColor Gray
}

# 1. Asegurar Docker PostgreSQL
Write-Host "[1/3] Levantando base de datos PostgreSQL en Docker..." -ForegroundColor Yellow
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

# 3. Iniciar Backend (las variables ya estan en el entorno desde el .env)
Write-Host "[3/3] Iniciando servidor Go en http://localhost:8080..." -ForegroundColor Green
Write-Host "Usuario admin: $($env:ADMIN_USUARIO)  |  Clave: $($env:ADMIN_CLAVE)" -ForegroundColor Cyan
Push-Location backend
go run ./cmd/api
Pop-Location
