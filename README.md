# Christ API

REST API backend dibangun dengan Go + Fiber. Fokus: cepat dikembangkan, mudah dibaca, dan siap dikembangkan lebih lanjut.

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
│   ├── auth/            → Authentication
│   ├── bible/           → Bible module
│   ├── contacts/        → Contacts
│   ├── news/            → News
│   └── ...
├── pkg/                 → Reusable packages
│   ├── database/        → DB connection
│   └── jwt/             → JWT utilities
├── routes/              → API endpoints registration
├── migrations/          → SQL migrations
├── SETUP.md             → 👈 Start here!
└── docker-compose.yml   → Container orchestration
```

### Prinsip Struktur:
Setiap fitur di `internal/<feature>/` punya 4 file:
- **handler.go** — Parse request, return response (thin layer)
- **service.go** — Business logic
- **repository.go** — Database queries (parameterized)
- **model.go** — Data structures

---

## ⚙️ Development

1. Buat fitur baru di `internal/<feature>` dengan empat file:
   - `handler.go` — hanya parsing request dan return response
   - `service.go` — logika utama, tanpa DB query langsung
   - `repository.go` — query ke DB (gunakan parameterized query)
   - `model.go` — struct data

2. Daftarkan route di `routes/routes.go`.
3. Unit test untuk `service` (mock repository) dan integration test untuk handler (httptest atau Postman).

Prinsip singkat: keep handlers thin, push logic ke service, query hanya di repository.

Contoh commit message singkat dan jelas:
```
feat(news): add list and create endpoints
fix(auth): handle nil DB connection
```

---

## Step-by-step: Tambah endpoint baru

Berikut langkah praktis untuk menambah fitur/endpoint baru. Contoh menggunakan feature `news`.

1) Buat folder feature

```bash
mkdir -p internal/news
```

2) `model.go` — definisikan struct data

```go
package news

import "time"

type News struct {
  ID int64 `json:"id"`
  UUID string `json:"uuid"`
  Title string `json:"title"`
  Content string `json:"content"`
  SiteID *int64 `json:"site_id,omitempty"`
  CreatedAt time.Time `json:"created_at"`
  UpdatedAt time.Time `json:"updated_at"`
}
```

3) `repository.go` — semua query DB, satu interface + satu impl

```go
package news

import (
  "context"
  "database/sql"
)

type Repository interface {
  FindByID(ctx context.Context, id int64) (*News, error)
  List(ctx context.Context, siteID *int64, limit, offset int) ([]News, error)
  Create(ctx context.Context, n *News) (*News, error)
}

type repo struct { db *sql.DB }
func NewRepository(db *sql.DB) Repository { return &repo{db: db} }

// implement FindByID, List, Create — gunakan QueryRowContext/QueryContext
```

4) `service.go` — business logic, bergantung pada interface repo

```go
package news

import "context"

type Service interface {
  Get(ctx context.Context, id int64) (*News, error)
  Create(ctx context.Context, n *News) (*News, error)
}

type service struct { repo Repository }
func NewService(r Repository) Service { return &service{repo: r} }

func (s *service) Get(ctx context.Context, id int64) (*News, error) {
  if id <= 0 { return nil, ErrInvalidID }
  return s.repo.FindByID(ctx, id)
}

// Create: validasi ringkas lalu repo.Create
```

5) `handler.go` — HTTP layer, tipis, register route

```go
package news

import "github.com/gofiber/fiber/v2"

type Handler struct{ svc Service }
func NewHandler(s Service) *Handler { return &Handler{svc: s} }

func (h *Handler) RegisterRoutes(app *fiber.App) {
  g := app.Group("/api")
  g.Get("/news", h.list)
  g.Post("/news", h.create)
}

// h.list/h.create: parse request, call service, return JSON/status
```

6) Wiring di `main.go` atau `routes.Setup`

```go
db := database.DB // atau NewPostgresConnection()
repo := news.NewRepository(db)
svc := news.NewService(repo)
handler := news.NewHandler(svc)
handler.RegisterRoutes(app)
```

7) Testing cepat

- Unit test service: mock `Repository` (interface) dan tes logika.
- Repo tests: gunakan `github.com/DATA-DOG/go-sqlmock` untuk assert query.
- Handler tests: gunakan Fiber's app + `httptest` untuk request/response.

Tips singkat:
- Jangan letakkan business logic di handler.
- Repository harus mengembalikan errors, bukan panic.
- Selalu gunakan parameterized query dan context.

---

## Testing

- Run unit tests:
```
go test ./...
```

- Rekomendasi:
  - Untuk repository tests: gunakan `github.com/DATA-DOG/go-sqlmock` agar tidak perlu DB nyata.
  - Untuk handler tests: gunakan `net/http/httptest` atau jalankan server lokal dan pakai Postman.

Contoh cepat curl:
```
# Login
curl -X POST http://localhost:3000/api/login \
  -H "Content-Type: application/json" \
  -d '{"email":"a@b.com","password":"pass"}'

# List news
curl http://localhost:3000/api/news
```

---

## Migrations

- Folder `migrations/` berisi SQL siap pakai. Untuk development cukup jalankan:
```
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f migrations/0001_create_news.sql
```
- Untuk production gunakan tool migration (contoh: `golang-migrate`).

---

## Lint & format

- Format: `gofmt -w .`
- Vet: `go vet ./...`
- (Opsional) Static analysis: `staticcheck ./...`

---

## Best practices singkat

- Hindari panic di repository; kembalikan error.
- Gunakan parameterized queries untuk mencegah SQL injection.
- Prefer dependency injection untuk repos/services (lebih mudah testing).
- Simpan secrets di `.env` dan jangan commit.

---

Butuh bantuan lebih lanjut? sebutkan fitur yang mau ditambah atau testing yang ingin dibuat — saya bantu contoh kode dan test case singkat.

Checklists: lihat [CHECKLIST.md](CHECKLIST.md) untuk panduan pre-commit, CI, migration, dan release.

---

**Enable Git pre-commit hooks (recommended)**

We added a simple pre-commit hook to ensure Go files are formatted before committing. To enable it for your local clone:

```bash
# run once per clone
git config core.hooksPath .githooks
chmod +x .githooks/pre-commit
```

What it does:
- Runs `gofmt -l .` and blocks the commit if any files are unformatted.
- If the hook blocks your commit, run `gofmt -w .`, `git add` the changes, then commit again.

You can remove or disable the hook by resetting `core.hooksPath`.


