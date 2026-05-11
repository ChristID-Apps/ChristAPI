# 🚀 Setup ChristAPI - Panduan Lengkap

## Untuk Pengguna Baru (Setelah Clone)

Setelah clone project ini, ada **2 cara** untuk menjalankan:

---

## ✅ Cara 1: DOCKER (Recommended - Windows)

**Paling mudah, semua otomatis!**

### Requirement:
- ✅ [Docker Desktop](https://www.docker.com/products/docker-desktop) (installed & running)

### Setup (One Command):

```powershell
# Buka PowerShell di folder project
powershell -ExecutionPolicy Bypass -File .\dalamNamaTuhan.ps1
```

Kalau kamu cuma mau menjalankan container yang sudah ada, pakai mode run-only:

```powershell
powershell -ExecutionPolicy Bypass -File .\dalamNamaTuhan.ps1 -NoBuild -NoMigrate
```

Kalau kamu cuma mau apply migration baru ke database existing:

```powershell
powershell -ExecutionPolicy Bypass -File .\dalamNamaTuhan.ps1 -MigrateOnly
```

**Apa yang terjadi otomatis:**
- ✅ Build Docker image
- ✅ Start PostgreSQL container
- ✅ Start API container  
- ✅ Run database migrations
- ✅ Tampilkan status & connection info

**Kalau pakai mode run-only:**
- ✅ Lewatkan build image
- ✅ Lewatkan migrasi
- ✅ Hanya start ulang container dan cek status

**Output (dalam ~15 detik):**
```
[SUCCESS] ChristAPI is ready!

API Server:
    http://localhost:3001

Database:
    Host: localhost
    Port: 5433
    User: christ_user
    Password: christ_password
    Database: christ_db
```

### Akses Project:
- **API**: http://localhost:3001
- **Database**: `localhost:5433` (gunakan DBeaver atau psql)

### Useful Commands:
```powershell
# Lihat logs API
docker compose logs -f golang-christapi

# Lihat logs Database
docker compose logs -f postgre-chrisapi

# Masuk ke container API
docker compose exec golang-christapi sh

# Stop semua services
docker compose down

# Mulai lagi
docker compose up -d

# Apply migration saja
powershell -ExecutionPolicy Bypass -File .\dalamNamaTuhan.ps1 -MigrateOnly
```

---

## 🛠️ Cara 2: LOCAL SETUP (Development)

**Jika ingin development tanpa Docker**

### Requirement:
- ✅ Go 1.25+ ([Download](https://golang.org/doc/install))
- ✅ PostgreSQL ([Download](https://www.postgresql.org/download/))
- ✅ golang-migrate CLI (opsional, untuk migrations)

### Step 1: Setup Database PostgreSQL

**Windows (Command Prompt):**
```powershell
# Buat database
createdb -U postgres christ_db

# Atau gunakan psql
psql -U postgres -c "CREATE DATABASE christ_db;"
```

**macOS/Linux:**
```bash
createdb christ_db
```

### Step 2: Edit `.env.local`

File `.env.local` sudah disiapkan di root folder. Sesuaikan kalau perlu:

```env
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=christ_db
DB_SSLMODE=disable
API_PORT=3000
JWT_SECRET=your-secret-key-here-change-in-production
```

**Sesuaikan:**
- `DB_USER` & `DB_PASSWORD` dengan PostgreSQL setup kamu
- Kalau PostgreSQL kamu jalan di Docker, pakai `DB_PORT=5433`
- Kalau PostgreSQL kamu jalan lokal langsung, pakai `DB_PORT=5432`
- `JWT_SECRET` dengan string random (production: gunakan secret manager)

Untuk Docker full-stack, pakai `.env.docker` yang sudah disiapkan di root project.

### Step 3: Run Migrations

**Option A: Gunakan golang-migrate CLI**

```powershell
# Install (Windows)
choco install migrate
# atau download manual: https://github.com/golang-migrate/migrate/releases

# Run migrations
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/christ_db?sslmode=disable" up
```

**Option B: Manual dengan psql**

```powershell
psql -U postgres -d christ_db -f docs/schema.sql
```

### Step 4: Download Dependencies & Run

```powershell
# Download dependencies
go mod download

# Run server
go run cmd/server/main.go
```

**Expected output:**
```
✅ PostgreSQL Connected
🚀 Server running on :3000

 ┌───────────────────────────────┐ 
 │      Fiber v2.52.12           │ 
 │   http://127.0.0.1:3000       │ 
 │                               │ 
 │ Handlers ............ 33      │ 
 └───────────────────────────────┘ 
```

### Akses Project:
- **API**: http://localhost:3000
- **Database**: `localhost:5433` jika DB pakai Docker, atau `localhost:5432` jika DB lokal

---

## 📋 Troubleshooting

### ❌ "Docker is not running"
→ Buka **Docker Desktop** dulu, tunggu sampai siap

### ❌ "Port 5433 already in use"
→ Ada container lain pakai port itu:
```powershell
# Cek apa yang pakai port 5433
netstat -ano | findstr 5433

# Stop container lain atau ganti port di docker-compose.yml
```

### ❌ "PostgreSQL failed to become healthy"
→ Restart dari awal:
```powershell
docker compose down
docker volume rm christapi_christ_postgres_data
powershell -ExecutionPolicy Bypass -File .\dalamNamaTuhan.ps1
```

### ❌ ".env tidak terbaca" (API crash)
→ Pastikan `.env` ada di root folder project (sama level dengan `docker-compose.yml`)

### ❌ Database user/password error
→ Cek `.env` dan `docker-compose.yml` - pastikan credential sama:
- `.env`: `DB_USER=christ_user`, `DB_PASSWORD=christ_password`
- `docker-compose.yml`: `POSTGRES_USER`, `POSTGRES_PASSWORD` harus sama

---

## 📚 Dokumentasi Lengkap

- **[DOCKER.md](./DOCKER.md)** - Setup Docker detail & best practices
- **[QUICKSTART.md](./QUICKSTART.md)** - Quick reference
- **[docs/schema.sql](./docs/schema.sql)** - Database schema

---

## 🎯 Pilih Setup yang Cocok untuk Kamu:

| Aspek | Docker | Local |
|-------|--------|-------|
| Setup Time | ~15 detik | ~5 menit |
| Dependency | Docker Desktop | Go, PostgreSQL |
| Isolation | Ya (clean container) | Tidak (system-wide) |
| Production | ✅ Recommended | ❌ Development only |
| Windows | ✅ Best | ⚠️ Requires extra setup |

**Rekomendasi:**
- **Windows user?** → Gunakan **Docker** (jauh lebih mudah)
- **Linux/macOS & prefer local dev?** → Gunakan **Local**
- **Team project?** → **Docker** (semua environment sama)

---

## ✨ Setelah Setup

### Test API:
```powershell
# Test login endpoint
curl -X POST http://localhost:3001/api/v1/auth/login `
  -H "Content-Type: application/json" `
  -d '{"email":"test@test.com","password":"test123"}'
```

### Next Steps:
1. Baca [CHECKLIST.md](./CHECKLIST.md) untuk development checklist
2. Buka [docs/schema.sql](./docs/schema.sql) untuk struktur database
3. Explore handlers di `internal/` folder untuk mengerti architecture

---

**Questions?** Baca [DOCKER.md](./DOCKER.md) atau cek logs:
```powershell
docker compose logs -f
```
