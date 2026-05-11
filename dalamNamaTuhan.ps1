# Setup & run ChristAPI with Docker (PowerShell)
# Usage: .\dalamNamaTuhan.ps1

$ErrorActionPreference = "Stop"

Write-Host "[*] Bismillah... Starting ChristAPI setup..." -ForegroundColor Cyan
Write-Host ""

# 1. Check if Docker is running
Write-Host "[*] Checking Docker..." -ForegroundColor Yellow
try {
    $null = & docker ps 2>&1
    Write-Host "[OK] Docker is running" -ForegroundColor Green
} catch {
    Write-Host "[ERROR] Docker is not running. Please start Docker Desktop and try again." -ForegroundColor Red
    exit 1
}
Write-Host ""

# 2. Check if .env exists, if not copy from .env.example
Write-Host "[*] Checking environment variables..." -ForegroundColor Yellow
if (-not (Test-Path ".env")) {
    if (-not (Test-Path ".env.example")) {
        Write-Host "[ERROR] .env.example not found" -ForegroundColor Red
        exit 1
    }
    Write-Host "[*] .env not found, copying from .env.example..."
    Copy-Item ".env.example" ".env"
    Write-Host "[OK] .env created" -ForegroundColor Green
} else {
    Write-Host "[OK] .env already exists" -ForegroundColor Green
}
Write-Host ""

# 3. Build Docker image
Write-Host "[*] Building Docker image..." -ForegroundColor Yellow
& docker compose build --no-cache | Out-Null
Write-Host "[OK] Build complete" -ForegroundColor Green
Write-Host ""

# 4. Start services
Write-Host "[*] Starting services (postgres, api)..." -ForegroundColor Yellow
& docker compose down 2>&1 | Out-Null
& docker compose up -d | Out-Null
Write-Host "[OK] Services started" -ForegroundColor Green
Write-Host ""

# 5. Wait for postgres to be healthy (simplified approach)
Write-Host "[*] Waiting for PostgreSQL to be healthy..." -ForegroundColor Yellow
Start-Sleep -Seconds 3  # Give PostgreSQL time to start
for ($attempt = 1; $attempt -le 15; $attempt++) {
    $result = & docker exec postgre-chrisapi pg_isready -U christ_user 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] PostgreSQL is healthy" -ForegroundColor Green
        break
    }
    Write-Host "    Waiting... (attempt $attempt/15)"
    Start-Sleep -Seconds 2
}
Write-Host ""

# 6. Run migrations
Write-Host "[*] Running database migrations..." -ForegroundColor Yellow
& docker compose run --rm migrate -path=/migrations -database "postgres://christ_user:christ_password@postgre-chrisapi:5432/christ_db?sslmode=disable" up | Out-Null
Write-Host "[OK] Migrations complete" -ForegroundColor Green
Write-Host ""

# 7. Show status
Write-Host "[*] Service status:" -ForegroundColor Yellow
& docker compose ps
Write-Host ""

# 8. Show access info
Write-Host "========================================" -ForegroundColor Green
Write-Host "[SUCCESS] ChristAPI is ready!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "[*] API Server:" -ForegroundColor Cyan
Write-Host "    http://localhost:3001" -ForegroundColor White
Write-Host ""
Write-Host "[*] Database:" -ForegroundColor Cyan
Write-Host "    Host: localhost" -ForegroundColor White
Write-Host "    Port: 5433" -ForegroundColor White
Write-Host "    Database: christ_db" -ForegroundColor White
Write-Host "    User: christ_user" -ForegroundColor White
Write-Host "    Password: christ_password" -ForegroundColor White
Write-Host ""
Write-Host "[*] Useful commands:" -ForegroundColor Cyan
Write-Host "    docker compose logs -f                 # View logs" -ForegroundColor White
Write-Host "    docker compose exec golang-christapi sh # Access API container" -ForegroundColor White
Write-Host "    docker compose down                    # Stop services" -ForegroundColor White
Write-Host ""
Write-Host "[*] DBeaver connection string:" -ForegroundColor Cyan
Write-Host "    postgres://christ_user:christ_password@localhost:5433/christ_db" -ForegroundColor White
Write-Host ""
