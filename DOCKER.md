# Docker Setup Guide

## Prerequisites
- Docker Desktop installed ([Download](https://www.docker.com/products/docker-desktop))
- Docker Compose (included with Docker Desktop)

## Quick Start

### 1. Create `.env` file
Copy dari `.env.example`:
```bash
cp .env.example .env
```

Edit `.env` dan ubah `JWT_SECRET` ke nilai yang aman untuk production:
```
DB_HOST=postgres
DB_PORT=5432
DB_USER=christ_user
DB_PASSWORD=christ_password
DB_NAME=christ_db
DB_SSLMODE=disable
JWT_SECRET=your-secure-secret-key-here
```

### 2. Build dan Run dengan Docker Compose
```bash
docker-compose up -d
```

Ini akan:
- Build image Go aplikasi
- Start PostgreSQL database
- Start API server pada port 3000

### 3. Verify Services Running
```bash
docker-compose ps
```

### Useful Commands

#### View Logs
```bash
docker-compose logs -f
# atau untuk service spesifik
docker-compose logs -f api
docker-compose logs -f postgres
```

#### Stop Services
```bash
docker-compose down
```

#### Stop dan Remove Data
```bash
docker-compose down -v
```

#### Rebuild Image
```bash
docker-compose build --no-cache
```

#### Access Database
```bash
docker-compose exec postgres psql -U christ_user -d christ_db
```

#### Access API Container Shell
```bash
docker-compose exec api sh
```

## Using Makefile (Optional)

Kalau mau pakai `Makefile` di Windows, ada 2 hal yang harus ada:

1. `make` harus ter-install
2. Docker Desktop harus sedang berjalan

### Cara install `make` di Windows

Pilih salah satu:

- **Scoop**:
   ```powershell
   scoop install make
   ```
- **Chocolatey**:
   ```powershell
   choco install make
   ```
- **WSL / Git Bash / MSYS2**:
   install GNU Make lewat environment tersebut, lalu jalankan command `make` dari shell itu.

Setelah itu, buka PowerShell / Git Bash / terminal yang sama dan jalankan:

```bash
make docker-build       # Build image
make docker-up          # Start services
make docker-down        # Stop services
make docker-logs        # View logs
make docker-restart     # Restart services
make docker-shell       # Access API shell
make docker-db-shell    # Access database shell
make docker-clean       # Stop dan hapus volume
```

### Kalau `make` tidak ada

Kamu tetap bisa jalanin semua command tanpa Makefile, misalnya:

```powershell
docker compose build
docker compose up -d
docker compose logs -f
docker compose down -v
```

### Catatan penting untuk Windows

- Kalau pakai **PowerShell**, pastikan command `make` memang ditemukan di `PATH`.
- Kalau `docker-compose` tidak ada, pakai `docker compose` langsung.
- Untuk workflow paling gampang di Windows, tetap pakai `dalamNamaTuhan.ps1` karena itu sudah otomatis build, start, dan migration.

## Production Setup

Untuk production, pastikan untuk:

1. **Change Database Credentials**
   - Update `DB_USER`, `DB_PASSWORD` di `.env`

2. **Change JWT Secret**
   - Generate secret yang kuat untuk `JWT_SECRET`
   ```bash
   openssl rand -hex 32
   ```

3. **Enable SSL Mode**
   - Ubah `DB_SSLMODE` ke `require` atau `verify-full`

4. **Use Docker Secrets** (jika di Docker Swarm)
   - Jangan store credentials di `.env` file
   - Gunakan Docker Secrets feature

5. **Environment Variables**
   ```bash
   docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
   ```

## Architecture

```
┌─────────────────────────────────┐
│       Docker Network            │
│   (christ-network)              │
├─────────────────────────────────┤
│ API Service (Port 3000)         │
│ - Go Fiber Application          │
│ - Routes: /api/*                │
│ - Services: auth, bible, etc    │
└──────────────┬──────────────────┘
               │
               │ (DB_HOST=postgres)
               │
┌──────────────▼──────────────────┐
│ PostgreSQL (Port 5432)          │
│ - Database: christ_db           │
│ - User: christ_user             │
└─────────────────────────────────┘
```

## Troubleshooting

### "Connection refused" error
- Pastikan PostgreSQL container healthy: `docker-compose ps`
- Check logs: `docker-compose logs postgres`
- Wait beberapa detik untuk database startup

### "Database does not exist" error
- Pastikan `schema.sql` ada di `docs/` folder
- Volume belum ter-mount dengan benar
- Try `docker-compose down -v && docker-compose up -d` untuk reset database

### Port 3000 already in use
Ubah port di `docker-compose.yml`:
```yaml
ports:
  - "3001:3000"  # Change to 3001
```

### Rebuild needed setelah code changes
```bash
docker-compose build && docker-compose up -d
```

## Next Steps
- Baca [main README](README.md) untuk dokumentasi API
- Check [database schema](docs/schema.sql)
- Lihat [testing guide](k6/) untuk load testing
