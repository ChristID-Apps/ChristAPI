# Christ API

REST API backend dibangun dengan **Go 1.25** + **Fiber v2**. Fokus: cepat dikembangkan, mudah dibaca, dan siap dikembangkan lebih lanjut.

**Version:** 1.2.0+ | **Status:** Active Development

**Features:**
- 🔐 Authentication (Email/Password + Google Sign-In)
- ✉️ OTP Verification & Email notifications
- 👥 Role Management (Admin, User, etc.)
- 📇 Contact Management
- 📰 News & Articles
- ⭐ Points System
- 🛡️ JWT + Admin Middleware

---

## � Kamu Baru Clone Project? Mulai Di Sini!

```
git clone <repo-url>
cd christ-api
          ↓
    👇 Pilih yang cocok 👇

┌─────────────────────────────────────┐
│ WINDOWS + Punya Docker Desktop?     │
├─────────────────────────────────────┤
│ ✅ Recommend: Run 1 command setup   │
│                                     │
│ powershell -ExecutionPolicy Bypass  │
│   -File .\dalamNamaTuhan.ps1        │
│                                     │
│ ⏱️  Selesai dalam ~15 detik        │
│                                     │
│ 📖 Detail: [SETUP.md](./SETUP.md)  │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│ LINUX/macOS/Prefer Local Dev?       │
├─────────────────────────────────────┤
│ 📖 Read: [SETUP.md](./SETUP.md)     │
│                                     │
│ Pilih 2 cara:                       │
│ • Docker (recommended untuk team)   │
│ • Local (direct Go + PostgreSQL)    │
└─────────────────────────────────────┘
```

---

## 🚀 Quick Start

### Windows (1 Command):
```powershell
powershell -ExecutionPolicy Bypass -File .\dalamNamaTuhan.ps1
```
**API ready at:** http://localhost:3001

Kalau container sudah pernah dibuat dan kamu cuma mau menjalankan lagi tanpa build ulang:
```powershell
powershell -ExecutionPolicy Bypass -File .\dalamNamaTuhan.ps1 -NoBuild -NoMigrate
```

Kalau kamu cuma mau apply migration baru ke database existing:
```powershell
powershell -ExecutionPolicy Bypass -File .\dalamNamaTuhan.ps1 -MigrateOnly
```

### Atau baca [SETUP.md](./SETUP.md) untuk:
- ✅ Penjelasan lengkap setiap step
- ✅ Mode full setup dan mode run-only
- ✅ Troubleshooting tips
- ✅ Local setup (non-Docker)
- ✅ Database access commands

Catatan singkat: `.env.local` dipakai untuk `go run`, sedangkan `.env.docker` dipakai Docker Compose.

## Tutorial Local Development: Database Docker, API Go di Host

> Jalur ini menjalankan **PostgreSQL saja di Docker**, sedangkan API dijalankan langsung dengan `go run`. Jangan menjalankan service `api` dari Compose pada jalur ini. Docker Compose development memetakan PostgreSQL ke `localhost:5433`; port internal container tetap `5432`.

### Windows (PowerShell)

1. Pasang dan jalankan Docker Desktop, Go 1.25+, dan Git. Clone repository lalu buka terminal di folder project.
2. Buat file konfigurasi lokal:
   ```powershell
   Copy-Item .env.docker.example .env.docker
   Copy-Item .env.example .env.local
   notepad .env.docker
   notepad .env.local
   ```
3. Isi `.env.docker` untuk container PostgreSQL: pertahankan `POSTGRES_USER=christ_user` (dipakai healthcheck), isi `POSTGRES_PASSWORD` dan `POSTGRES_DB`, lalu samakan `DB_USER`, `DB_PASSWORD`, dan `DB_NAME`. Biarkan `DB_HOST=postgres` dan `DB_PORT=5432` di file Docker. Isi `JWT_SECRET` dengan nilai lokal acak.
4. Atur titik gereja **sekali saja** di `.env.docker`: isi `ATTENDANCE_LATITUDE` dan `ATTENDANCE_LONGITUDE` dengan koordinat pin gereja. Ambil latitude/longitude dari pin peta (misalnya klik kanan titik di Google Maps lalu salin koordinat). Radius tetap 500 meter. Tidak perlu mengaktifkan precise location; akurasi yang dilaporkan perangkat dicatat, bukan dijadikan syarat check-in.
5. Di `.env.local`, samakan `ATTENDANCE_LATITUDE` dan `ATTENDANCE_LONGITUDE` serta kredensial DB, tetapi gunakan `DB_HOST=localhost`, `DB_PORT=5433`, `DB_SSLMODE=disable`, `API_PORT=3000`, dan isi `JWT_SECRET`. File ini dipakai Go yang berjalan di Windows; jangan gunakan `DB_HOST=postgres` di sini.
6. Jalankan **database saja** dan tunggu sampai sehat:
   ```powershell
   docker compose up -d postgres
   docker compose ps
   docker compose logs postgres
   ```
7. Terapkan migration satu kali:
   ```powershell
   docker compose --profile manual-migration run --rm migrate
   ```
8. Jalankan API dari host:
   ```powershell
   go mod download
   go run ./cmd/server
   ```
9. API tersedia di `http://localhost:3000`. Untuk berhenti, tekan `Ctrl+C` pada terminal Go dan jalankan `docker compose stop postgres`. Data database tetap tersimpan di named volume.

### Linux (Bash)

1. Pasang Docker Engine + Docker Compose plugin, Go 1.25+, dan Git. Pastikan user dapat menjalankan `docker` tanpa masalah permission. Clone repository lalu masuk ke folder project.
2. Buat file konfigurasi lokal:
   ```bash
   cp .env.docker.example .env.docker
   cp .env.example .env.local
   nano .env.docker
   nano .env.local
   ```
3. Atur `.env.docker` seperti langkah Windows: pertahankan `POSTGRES_USER=christ_user` (dipakai healthcheck), isi password/database, lalu samakan `DB_USER`, `DB_PASSWORD`, `DB_NAME`; gunakan `DB_HOST=postgres`, `DB_PORT=5432`, dan `JWT_SECRET` lokal.
4. Isi `ATTENDANCE_LATITUDE` dan `ATTENDANCE_LONGITUDE` di `.env.docker` dengan koordinat pin gereja. Salin pasangan yang sama ke `.env.local`. Ambil koordinat dari pin peta; radius server 500 meter. Precise location tidak diwajibkan.
5. Atur `.env.local` dengan nilai database yang sama, tetapi `DB_HOST=localhost`, `DB_PORT=5433`, `DB_SSLMODE=disable`, `API_PORT=3000`, serta `JWT_SECRET`. Go berjalan di host sehingga tidak dapat memakai hostname service `postgres`.
6. Jalankan **database saja**:
   ```bash
   docker compose up -d postgres
   docker compose ps
   docker compose logs postgres
   ```
7. Terapkan migration satu kali:
   ```bash
   docker compose --profile manual-migration run --rm migrate
   ```
8. Jalankan API dari host:
   ```bash
   go mod download
   go run ./cmd/server
   ```
9. API tersedia di `http://localhost:3000`. Untuk berhenti, tekan `Ctrl+C` pada terminal Go lalu jalankan `docker compose stop postgres`. Named volume mempertahankan data.

**Penting:** `.env.local` dan `.env.docker` hanya untuk mesin lokal dan tidak boleh di-commit. `docker compose down -v` menghapus volume PostgreSQL beserta seluruh datanya; jangan gunakan kecuali memang ingin menghapus database lokal.

## Menjalankan Test

Test unit tidak membutuhkan PostgreSQL atau server berjalan:

```bash
go test ./...
```

Perintah yang direkomendasikan sebelum membuat perubahan/PR:

```bash
go test ./...
go vet ./...
go build ./...
```

Test route saat ini memeriksa registrasi route, penolakan request tanpa token pada protected routes, dan validasi dasar untuk auth routes publik. Belum semua success flow handler dan query repository diuji; untuk smoke test endpoint dengan data nyata, jalankan database + migrations + API sesuai tutorial di atas. Integration test PostgreSQL perlu memakai database test terisolasi, bukan database pengguna/production.

## Attendance Geofence dan Bacaan Harian

### Check-in attendance

FE meminta lokasi browser/perangkat setelah jemaat menekan tombol absen, lalu mengirim posisi perangkat—tanpa `site_id`, `activity_id`, titik tujuan, atau radius:

```http
POST /api/attendance/check-in
Authorization: Bearer <token>
Content-Type: application/json
```

```json
{"latitude": -6.2000, "longitude": 106.8167, "accuracy_m": 18.5}
```

Backend mencocokkan posisi perangkat terhadap satu titik gereja dari `ATTENDANCE_LATITUDE` dan `ATTENDANCE_LONGITUDE`. Jarak maksimal 500 meter; tidak ada persyaratan precise location atau ambang akurasi perangkat. Satu check-in berhasil per tanggal Asia/Jakarta memberi 10 poin; pengulangan pada hari yang sama mendapat `409`.

### Submission bacaan Alkitab harian

FE mengirim object JSON fleksibel sebagai `payload` (maksimal 16 KB):

```http
POST /api/bible-reading-submissions
Authorization: Bearer <token>
Content-Type: application/json
```

```json
{"payload":{"reference":"Mazmur 23","reflection":"Tuhan memelihara saya.","source":"daily-reading"}}
```

Riwayat jemaat: `GET /api/bible-reading-submissions/me`. Antrean admin: `GET /api/admin/bible-reading-submissions?status=submitted`. Admin menyetujui dan memberi poin dalam satu request transaksional:

```http
POST /api/admin/bible-reading-submissions/:uuid/approve
Authorization: Bearer <admin-token>
Content-Type: application/json
```

```json
{"points": 10}
```

Untuk menolak, `POST /api/admin/bible-reading-submissions/:uuid/reject` dengan body `{"note":"Alasan penolakan"}`. Jemaat boleh mengirim ulang pada hari yang sama setelah ditolak; submission sebelumnya tetap ada sebagai riwayat. Submission pending/approved membatasi pengajuan baru untuk hari tersebut, dan approval ganda tidak memberikan poin dua kali.

## Catatan Production

Compose dan file env pada repository ini ditujukan untuk development, bukan deployment production langsung. Untuk production:

1. Siapkan PostgreSQL terkelola/terisolasi; jangan buka port database ke internet. Gunakan user database dengan hak minimum dan backup/restore yang sudah diuji.
2. Inject `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`, `JWT_SECRET`, SMTP, dan Google OAuth dari secret manager/platform deployment. Jangan menyalin `.env.local` atau `.env.docker` ke server.
3. Gunakan credential unik dan `JWT_SECRET` acak yang kuat. Set TLS database (`verify-full` bila CA/hostname dapat dikonfigurasi dengan benar) dan terminasi HTTPS pada load balancer/reverse proxy.
4. Jalankan migration sebagai job rilis **satu kali** sebelum/bersamaan dengan deployment, misalnya melalui pipeline migration yang mengakses schema dari artifact. Jangan menjalankan beberapa migration job bersamaan dan jangan mengandalkan startup API untuk mengubah schema.
5. Build dan deploy versi aplikasi yang sama pada seluruh instance. Atur `CORS_ORIGINS` hanya untuk domain frontend yang dipercaya dan gunakan `API_PORT` sesuai platform.
6. Upload saat ini disimpan di filesystem lokal aplikasi. Sebelum menjalankan banyak instance atau mengganti container, siapkan storage persisten yang sesuai (misalnya object storage atau volume terkelola) dan backup file upload.

Lakukan rollout dan rollback pada staging terlebih dahulu. Down migration tertentu bersifat destruktif; backup dan persetujuan data owner diperlukan sebelum rollback.

---

## 📚 Dokumentasi Utama

| File | Untuk Apa |
|------|-----------|
| **[SETUP.md](./SETUP.md)** | 👈 **Baca pertama!** Panduan setup lengkap |
| [QUICKSTART.md](./QUICKSTART.md) | Cheat sheet commands |
| [DOCKER.md](./DOCKER.md) | Docker detail & best practices |
| [CHECKLIST.md](./CHECKLIST.md) | Development workflow |
| [docs/schema.sql](./docs/schema.sql) | Database schema |

---

## 🏗️ Struktur Proyek

```
ChristAPI/
├── cmd/server/           → Entry point (main.go)
├── internal/             → Feature modules
│   ├── auth/            → Authentication (Email, OTP, Google Sign-In)
│   ├── role/            → Role Management (Admin, User, etc.)
│   ├── email/           → Email notifications & OTP
│   ├── contacts/        → Contact management
│   ├── news/            → News & articles
│   ├── points/          → Points/reward system
│   ├── sites/           → Site management
│   └── middleware/      → Admin, JWT, etc.
├── pkg/                 → Reusable packages
│   ├── database/        → DB connection & context
│   ├── jwt/             → JWT utilities
│   └── response/        → Response formatting
├── routes/              → API endpoints registration
├── migrations/          → SQL migrations versioned, dijalankan one-shot
├── docs/                → API documentation & schema
├── .githooks/           → Git hooks (pre-commit format check)
├── SETUP.md             → 👈 Start here!
├── QUICKSTART.md        → Common commands
├── DOCKER.md            → Docker guide
├── CHECKLIST.md         → Development workflow
└── docker-compose.yml   → Container orchestration
```

### Prinsip Struktur:
Setiap fitur di `internal/<feature>/` punya 4 file utama:
- **handler.go** — Parse request, return response (thin layer)
- **service.go** — Business logic, no DB queries
- **repository.go** — Database queries (parameterized, context-aware)
- **model.go** — Data structures & constants

---

## 🔐 Authentication & Authorization

**Supported Methods:**
- Email + Password (with OTP verification)
- Google Sign-In (OAuth 2.0)
- Admin user seeding on first setup

**Features:**
- OTP via email with enhanced HTML templates
- JWT-based authentication
- Admin middleware for protected routes
- User approval workflow
- Profile completion flow (post-Google sign-in)
- Automatic contact creation on login

**Key Endpoints:**
```
POST   /api/login                   → Email/password login
POST   /api/register                → Register new user
POST   /api/auth/google             → Google OAuth flow
POST   /api/verify-otp              → Verify OTP
POST   /api/resend-otp              → Resend OTP
```

---

## 👥 Role Management

**Built-in Roles:**
- `admin` — Full system access
- `user` — Standard user permissions

**Role Operations:**
```
GET    /api/admin/roles             → List all roles (admin-only)
POST   /api/admin/roles             → Create role (admin-only)
PATCH  /api/admin/roles/:id         → Update role (admin-only)
```

Each role has:
- `id` — Numeric ID
- `code` — Unique code (e.g., "admin", "user")
- `name` — Display name
- `description` — Role description
- `created_at`, `updated_at` — Timestamps

---

## ✉️ Email & Notifications

**Email Features:**
- OTP delivery with styled HTML templates
- Parameterized SMTP configuration
- GoMail integration

**Template System:**
- Dynamic content injection
- Professional HTML formatting
- Responsive design

---

## 📊 Additional Features

**Contact Management**
- Create & manage contacts
- Link to user profiles

**News & Articles**
- CRUD operations
- Site-based organization
- Public read endpoint: `GET /api/news` (only published, non-deleted news); admin management stays under `/api/admin/news`

**Points System**
- User points tracking
- Reward management

---

## ⚙️ Development Workflow

**Adding a New Endpoint:**

1. Create feature folder
```bash
mkdir -p internal/<feature>
```

2. Define data structure in `model.go`
```go
type Model struct {
  ID        int64     `json:"id"`
  UUID      string    `json:"uuid"`
  Name      string    `json:"name"`
  CreatedAt time.Time `json:"created_at"`
  UpdatedAt time.Time `json:"updated_at"`
}
```

3. Define repository interface in `repository.go`
```go
type Repository interface {
  FindByID(ctx context.Context, id int64) (*Model, error)
  List(ctx context.Context, limit, offset int) ([]Model, error)
  Create(ctx context.Context, m *Model) (*Model, error)
}
```

4. Implement business logic in `service.go`
```go
type Service interface {
  Get(ctx context.Context, id int64) (*Model, error)
  Create(ctx context.Context, m *Model) (*Model, error)
}
```

5. Create HTTP handlers in `handler.go`
```go
func (h *Handler) RegisterRoutes(app *fiber.App) {
  g := app.Group("/api/<feature>")
  g.Get("/:id", h.Get)
  g.Post("/", h.Create)
}
```

6. Wire up in `routes/routes.go`
```go
repo := feature.NewRepository(db)
svc := feature.NewService(repo)
handler := feature.NewHandler(svc)
handler.RegisterRoutes(app)
```

7. Add tests

**Core Principles:**
- Handlers: parse request, call service, return response (thin layer)
- Service: business logic only, no DB queries
- Repository: all data access with parameterized queries
- Use `context.Context` in all async operations
- Return errors, never panic

**Commit Messages:**
```
feat(role): add code field to roles and enhance management
fix(auth): handle nil DB connection during Google OAuth
refactor(contacts): improve validation logic
```

---

## 🧪 Testing

**Run all tests:**
```bash
go test ./...
```

Test yang ada saat ini mencakup aturan/unit tertentu dan route guard. Belum semua service, handler, repository, maupun query database memiliki test; jangan menganggap `go test ./...` sebagai pengganti integration test PostgreSQL.

**Example requests:**
```bash
# Register
curl -X POST http://localhost:3001/api/register \
  -H "Content-Type: application/json" \
  -d '{"full_name":"User Example","email":"user@example.com","password":"securepass123"}'

# Login with OTP
curl -X POST http://localhost:3001/api/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","otp_code":"123456"}'

# Resend OTP
curl -X POST http://localhost:3001/api/resend-otp \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com"}'

# Get current user profile (requires JWT token in Authorization header)
curl -H "Authorization: Bearer <token>" \
  http://localhost:3001/api/profile

# Admin: List roles
curl -H "Authorization: Bearer <admin-token>" \
  http://localhost:3001/api/admin/roles
```

---

## 📦 Migrations

Schema dikelola oleh pasangan file SQL bernomor di `migrations/`, bukan oleh snapshot `docs/schema.sql`. Pada workflow database lokal + Go di host, jalankan PostgreSQL lalu terapkan migration one-shot:

```bash
docker compose up -d postgres
docker compose --profile manual-migration run --rm migrate
```

Jalankan perintah migration lagi setelah menambahkan versi baru; migration yang sudah tercatat tidak dijalankan ulang. Untuk melihat versi yang sudah diterapkan:

```bash
docker compose --profile manual-migration run --rm \
  -e MIGRATION_ACTION=version migrate
```

Untuk mencoba rollback satu versi pada **database development disposable saja**:

```bash
docker compose --profile manual-migration run --rm \
  -e MIGRATION_ACTION=down -e MIGRATION_STEPS=1 migrate
```

> **Peringatan:** beberapa down migration menghapus data. Khusus migration attendance, rollback menjatuhkan tabel attendance. Jangan menjalankan `down` pada production tanpa backup, rencana pemulihan, dan persetujuan.

Migration baru harus memiliki pasangan `N_description.up.sql` dan `N_description.down.sql`, diuji pada database disposable, serta direview untuk efek data/lock sebelum rilis. Format penamaan mengikuti migration terakhir yang ada di folder.

---

## 🔍 Lint & Format

**Format code:**
```bash
gofmt -w .
```

**Vet for issues:**
```bash
go vet ./...
```

**Optional static analysis:**
```bash
staticcheck ./...
```

**Pre-commit hook** (recommended — blocks unformatted commits):
```bash
git config core.hooksPath .githooks
chmod +x .githooks/pre-commit
```

---

## ✅ Best Practices

- **No panics in repositories** — return errors always
- **Parameterized queries only** — prevents SQL injection
- **Dependency injection** — wire repos/services at startup, makes testing easier
- **Secrets in .env** — never commit them
- **Use context.Context** — for timeouts and cancellation in all I/O operations
- **Thin handlers** — parse, call service, return response
- **Service contains logic** — no DB queries in services
- **Repository does data access** — all queries parameterized, context-aware

---

## 📚 Additional Resources

- **[SETUP.md](./SETUP.md)** — Complete setup guide (Docker & local)
- **[QUICKSTART.md](./QUICKSTART.md)** — Common commands cheat sheet
- **[DOCKER.md](./DOCKER.md)** — Docker setup and troubleshooting
- **[CHECKLIST.md](./CHECKLIST.md)** — Pre-commit, CI, migration checklist
- **[docs/schema.sql](./docs/schema.sql)** — Database schema reference

---

**Questions?** Open an issue or check the documentation files above.


