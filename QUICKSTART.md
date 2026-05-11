# Quick Start Scripts & Commands

**New here?** Read [SETUP.md](./SETUP.md) first!

---

## Fastest Way - Docker One-Liner (Windows)

```powershell
powershell -ExecutionPolicy Bypass -File .\dalamNamaTuhan.ps1
```

✅ Done in ~15 seconds. API ready at http://localhost:3001

Catatan keamanan: pastikan `.env` berisi `JWT_SECRET`. Aplikasi akan berhenti kalau secret ini belum diisi.

### Run-only mode

Kalau image sudah ada dan kamu cuma mau hidupkan container lagi tanpa build ulang:

```powershell
powershell -ExecutionPolicy Bypass -File .\dalamNamaTuhan.ps1 -NoBuild -NoMigrate
```

Use case:
- `-NoBuild` = skip build image
- `-NoMigrate` = skip migration
- cocok kalau kamu cuma restart service yang sudah ada

### Restart fast (new)

Gunakan opsi `-Restart` untuk restart container lebih cepat tanpa `down/up` (script akan skip build/migrate):

```powershell
# Restart all services (fast)
powershell -ExecutionPolicy Bypass -File .\dalamNamaTuhan.ps1 -Restart

# Restart only API service by name
powershell -ExecutionPolicy Bypass -File .\dalamNamaTuhan.ps1 -Restart -Service golang-christapi
```

### Migration only

Kalau kamu cuma mau apply migration baru ke database existing:

```powershell
powershell -ExecutionPolicy Bypass -File .\dalamNamaTuhan.ps1 -MigrateOnly
```

---

## Setup Automation Scripts

### Windows (PowerShell) — Recommended

```powershell
# Auto-setup everything
.\dalamNamaTuhan.ps1

# Troubleshoot if error "running scripts is disabled"
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
.\dalamNamaTuhan.ps1
```

**What it does:**
- ✅ Checks Docker running
- ✅ Builds image
- ✅ Starts containers
- ✅ Waits for PostgreSQL healthy
- ✅ Runs migrations
- ✅ Shows connection info

**Run-only mode:**
- pakai `-NoBuild -NoMigrate`
- tidak rebuild image
- tidak jalankan migration lagi

---

### Linux / macOS / Windows WSL / Git Bash

```bash
# Make executable
chmod +x dalamNamaTuhan.sh

# Run
./dalamNamaTuhan.sh
```

Same automation as PowerShell version.

---

## Common Docker Commands

```powershell
# View logs (all services)
docker compose logs -f

# View API logs only
docker compose logs -f golang-christapi

# View Database logs only
docker compose logs -f postgre-chrisapi

# Access API container shell
docker compose exec golang-christapi sh

# Access Database psql shell
docker compose exec postgre-chrisapi psql -U christ_user -d christ_db

# Stop all services
docker compose down

# Start services again
docker compose up -d

# Rebuild image (if you changed code)
docker compose build --no-cache
docker compose up -d

# Show running containers
docker compose ps

# Show container details
docker ps
```

---

## Database Access

### Via DBeaver (Recommended)
```
Connection Type: PostgreSQL
Host: localhost
Port: 5433
Database: christ_db
User: christ_user
Password: christ_password
```

### Via psql Command
```powershell
# Windows
psql -h localhost -U christ_user -d christ_db -W
# (masukkan password saat diminta)

# macOS/Linux
psql -h localhost -U christ_user -d christ_db
```

---

## API Testing

### Test Login Endpoint
```powershell
# Windows PowerShell
Invoke-WebRequest -Uri "http://localhost:3001/api/v1/auth/login" `
  -Method POST `
  -ContentType "application/json" `
  -Body '{"email":"test@test.com","password":"test123"}'

# Or use curl (if installed)
curl -X POST http://localhost:3001/api/v1/auth/login `
  -H "Content-Type: application/json" `
  -d '{"email":"test@test.com","password":"test123"}'
```

### Expected Response (400 Bad Request)
```json
{
  "error": "missing token"
}
```

---

## Local Development (No Docker)

### Setup
```powershell
# 1. Create database
createdb christ_db

# 2. Edit .env.local
# File .env.local sudah disiapkan untuk local development

# 3. Run migrations
migrate -path migrations -database "postgres://user:password@localhost:5432/christ_db?sslmode=disable" up

# 4. Start server
go run cmd/server/main.go
```

Kalau kamu pakai PostgreSQL dari Docker, pakai `DB_PORT=5433` dan tetap jalankan API local di `API_PORT=3000`.

### Stop Server
```
Press Ctrl+C
```

### View Code & Logs
- API handlers: `internal/*/handler.go`
- Database models: `internal/*/model.go`
- Routes: `routes/routes.go`

---

## After Setup - What's Next?

1. **Review structure**: Open [docs/schema.sql](./docs/schema.sql)
2. **Check checklist**: Read [CHECKLIST.md](./CHECKLIST.md)
3. **Explore handlers**: Look at `internal/auth`, `internal/bible`, etc.
4. **Run tests**: `go test ./...`
5. **Read docs**: [DOCKER.md](./DOCKER.md), [SETUP.md](./SETUP.md)

---

## Need Help?

- **Setup issues?** → [SETUP.md](./SETUP.md) troubleshooting
- **Docker help?** → [DOCKER.md](./DOCKER.md)
- **Schema?** → [docs/schema.sql](./docs/schema.sql)
- **Development guide?** → [CHECKLIST.md](./CHECKLIST.md)

---

## What the scripts do

1. **Check Docker** — pastikan Docker Desktop running
2. **Setup .env** — copy dari `.env.example` jika belum ada
3. **Build image** — compile Go app ke Docker image
4. **Start services** — jalankan `docker compose up -d` (postgres + api)
5. **Wait for DB** — tunggu PostgreSQL healthy
6. **Run migrations** — apply SQL migrations ke database
7. **Show status** — display containers status
8. **Print info** — tampilkan API URL, DB connection info, useful commands

---

## Output example

```
🙏 Bismillah... Starting ChristAPI setup...

🐳 Checking Docker...
✅ Docker is running

⚙️  Checking environment variables...
✅ .env already exists

🔨 Building Docker image...
[+] Building 45.2s (19/19) FINISHED
✅ Build complete

🚀 Starting services (postgres, api)...
✅ Services started

⏳ Waiting for PostgreSQL to be healthy...
✅ PostgreSQL is healthy

🔄 Running database migrations...
1/u init (326.187593ms)
✅ Migrations complete

📊 Service status:
NAME               IMAGE          COMMAND                  STATUS
golang-christapi   christapi-api  "./server"               Up 2 seconds
postgre-chrisapi   postgres:16    "docker-entrypoint..."  Up 12 seconds (healthy)

========================================
🎉 ChristAPI is ready!
========================================

📍 API Server:
   http://localhost:3001

🗄️  Database:
   Host: localhost
   Port: 5433
   Database: christ_db
   User: christ_user
   Password: christ_password

📚 Useful commands:
   docker compose logs -f                 # View logs
   docker compose exec golang-christapi sh # Access API container
   docker compose down                    # Stop services

DBeaver connection string:
   postgres://christ_user:christ_password@localhost:5433/christ_db
```

---

## Stop services

```powershell
# PowerShell
docker compose down

# Or bash/sh
docker compose down
```

---

## More info

- See [DOCKER.md](DOCKER.md) untuk dokumentasi Docker lengkap
- See [README.md](README.md) untuk dokumentasi API

