# Christ API

REST API backend sederhana dibangun dengan Go + Fiber. Fokus: cepat dikembangkan, mudah dibaca, dan siap dikembangkan lebih lanjut.

Ringkas dan langsung ke poin: berikut cara setup, cara ngoding, dan testing.

---

## Quick start

1. Copy repository

```bash
git clone <repo-url>
cd christ-api
```

2. Buat database PostgreSQL

```bash
createdb christ_api
```

3. Buat file `.env` di root (conto):

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=christ_api
DB_SSLMODE=disable
JWT_SECRET=changeme
PORT=3000
```

4. Install deps dan run

```bash
go mod download
go run cmd/server/main.go
```

Server default: http://localhost:3000

Untuk menjalankan migrations SQL yang ada: gunakan `psql` atau `golang-migrate` (rekomendasi untuk production).

---

## Struktur proyek (cepat)

- `cmd/server/main.go` — entry point
- `routes/` — daftar endpoint
- `internal/` — fitur (feature-based): setiap fitur biasanya punya `handler.go`, `service.go`, `repository.go`, `model.go`
- `pkg/database` — koneksi DB
- `migrations/` — SQL migration

Konvensi penting:
- Handler = thin (HTTP parsing + response)
- Service = business logic
- Repository = semua query database (parameterized queries)

Catatan: beberapa modul saat ini menggunakan `database.DB` global; idealnya gunakan dependency injection (constructor) untuk testability.

---

## Environment

Pastikan environment variables di `.env` sudah diatur. Kunci yang biasa dipakai:
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`
- `JWT_SECRET` — rahasiakan
- `PORT` — server port

---

## Cara ngoding (singkat & praktis)

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

