# Christ API 🚀

Sebuah REST API backend yang dibangun dengan **Go** dan **Fiber Framework** untuk mengelola data dengan sistem autentikasi JWT.

---

## Untuk Pemula 👶

Jika ini pertama kalimu, jangan khawatir! Repository ini dirancang agar mudah dipahami.

---

## Requirements

Sebelum memulai, pastikan sudah install:

- **Go** versi 1.25.0 atau lebih tinggi ([Download](https://golang.org/))
- **PostgreSQL** ([Download](https://www.postgresql.org/))
- **Git** ([Download](https://git-scm.com/))

---

## Setup Cepat

### 1. Clone Project
```bash
git clone <repo-url>
cd christ-api
```

### 2. Setup Database
```bash
# Buat database PostgreSQL
createdb christ_api

# Jalankan migration jika ada
# psql christ_api < migrations/init.sql
```

### 3. Setup Environment
Buat file `.env` di root folder:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_username
DB_PASSWORD=your_password
DB_NAME=christ_api
JWT_SECRET=your_secret_key_here
PORT=3000
```

### 4. Download Dependencies
```bash
go mod download
```

### 5. Jalankan Project
```bash
go run cmd/server/main.go
```

Server berjalan di: `http://localhost:3000`

---

## Struktur Folder yang Mudah Dipahami

```
christ-api/
├── cmd/
│   └── server/
│       └── main.go          ← Entry point (dimulai dari sini)
│
├── internal/                 ← Core business logic
│   ├── auth/                 ← Fitur login & autentikasi
│   │   ├── handler.go        ← Menerima request HTTP
│   │   ├── service.go        ← Logika bisnis
│   │   ├── repository.go     ← Query database
│   │   └── model.go          ← Struktur data
│   │
│   └── middleware/           ← Proses request sebelum sampai handler
│       ├── auth.go           ← Cek JWT token
│       └── logger.go         ← Catat setiap request
│
├── pkg/                      ← Utility & helper umum
│   ├── database/             ← Koneksi database
│   └── jwt/                  ← Token management
│
├── routes/
│   └── routes.go            ← Daftar semua endpoint API
│
├── go.mod                   ← Daftar dependency
└── .env                     ← Konfigurasi (jangan commit ke git!)
```

### 📚 Penjelasan Singkat

- **cmd/**: Tempat program dimulai (entry point)
- **internal/**: Folder yang paling penting - tempat logika aplikasi
- **pkg/**: Helper & utility yang bisa dipakai di mana saja
- **routes/**: Daftar semua endpoint (path API)

---

## API Endpoint

### 1. Login (Tidak perlu token)
```bash
POST /api/login
Content-Type: application/json

{
  "username": "john",
  "password": "password123"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### 2. Profile (Perlu token)
```bash
GET /api/profile
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

**Response:**
```json
{
  "message": "you are logged in"
}
```

---

## Cara Kerja Sederhana

### 1. User Kirim Request
```
Request HTTP → /api/login
```

### 2. Handler Terima
```go
// handler.go menerima request
func Login(c *fiber.Ctx) error {
    // Ambil data dari request
    // Panggil service
}
```

### 3. Service Proses Logika
```go
// service.go proses data
func (s *Service) Login(username, password string) {
    // Validasi
    // Generate token
}
```

### 4. Repository Query Database
```go
// repository.go ambil data dari DB
func (r *Repository) GetUser(username string) {
    // SELECT * FROM users
}
```

### 5. Kembali Response
```
← JSON response ← Service ← Repository
```

---

## JWT Authentication

### Apa itu JWT?
- Token yang berisi informasi user
- Dikirim di setiap request yang butuh login
- Server memverifikasi token

### Cara Pakai:
1. Login dulu, dapat token
2. Simpan token di client
3. Kirim token di header: `Authorization: Bearer <token>`
4. Server verifikasi token, baru akses resource

---

## Development Tips untuk Junior 🎯

### 1. Baca File Secara Berurutan
- `cmd/server/main.go` (mulai dari sini)
- `routes/routes.go` (lihat endpoint)
- `internal/auth/handler.go` (lihat request/response)
- `internal/auth/service.go` (lihat logika)
- `internal/auth/repository.go` (lihat query DB)

### 2. Testing Endpoint
Gunakan **Postman** atau **curl**:
```bash
# Login
curl -X POST http://localhost:3000/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"john","password":"pass123"}'

# Akses protected route (pakai token dari login)
curl -X GET http://localhost:3000/api/profile \
  -H "Authorization: Bearer TOKEN_DARI_LOGIN"
```

### 3. Baca Error dengan Teliti
- Error message biasanya bilang apa yang salah
- Cek log di terminal saat run `go run cmd/server/main.go`

### 4. Jangan Takut untuk Debug
- Print nilai di console: `fmt.Println(data)`
- Atau gunakan debugger Go

---

## Common Issues & Solusi

| Masalah | Solusi |
|---------|--------|
| Error: `cannot find package` | Jalankan `go mod download` |
| Database connection error | Cek `.env` sudah benar dan PostgreSQL running |
| JWT token invalid | Pastikan `JWT_SECRET` di `.env` sama di semua tempat |
| Port 3000 sudah dipakai | Ubah `PORT` di `.env` atau stop app lain |

---

## Menambah Fitur Baru

Misalnya ingin tambah fitur **User Management**:

### 1. Buat folder baru
```bash
mkdir -p internal/user
```

### 2. Buat 4 file (handler, service, repository, model)
```bash
touch internal/user/{handler,service,repository,model}.go
```

### 3. Ikuti pattern yang ada di `internal/auth/`

### 4. Daftarkan route di `routes/routes.go`

---

## Resources untuk Belajar

- [Fiber Documentation](https://docs.gofiber.io/) - Framework yang dipakai
- [Go Tutorial](https://golang.org/doc/) - Bahasa yang dipakai
- [JWT Explained](https://jwt.io/introduction) - Autentikasi yang dipakai
- [PostgreSQL Basics](https://www.postgresql.org/docs/) - Database yang dipakai

---

## Sebelum Push ke Git

✅ Checklist:

- [ ] `.env` sudah di `.gitignore` (jangan push!)
- [ ] Semua code sudah tested
- [ ] Tidak ada hardcoded password/secret
- [ ] Error handling sudah ada
- [ ] Database migration clear

---

## Bantuan & Pertanyaan?

Kalau ada yang tidak paham:
1. Baca komentar di code
2. Baca file `go.mod` untuk tau dependency apa saja
3. Tanya senior atau lihat documentation

---

**Happy Coding! 🎉**

---

*Last Updated: 30 Maret 2026*
