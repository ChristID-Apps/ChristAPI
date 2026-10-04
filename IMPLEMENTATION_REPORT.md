# Implementation Report

## Ringkasan

Perbaikan backlog dari [CODEBASE_AUDIT.md](CODEBASE_AUDIT.md) diterapkan tanpa rewrite arsitektur. Setelah policy dikonfirmasi, draft visibility, target points, attendance-backed streak, dan role revocation juga diperbaiki. Migration 21 berhasil diterapkan ke database disposable; satu orphan `users.site_id` dipertahankan di bawah FK `NOT VALID` tanpa mengubah row.

## Finding yang Berhasil Diperbaiki

### A-02 - Draft visibility
**Status:** Resolved
**Test:** Unit tests untuk status filter dan visibility; hasil full suite akhir dicatat pada bagian verification.
**Masalah:** Member dapat membaca berita/aktivitas draft dan mengirim submission ke activity draft.

**Perubahan:** Member list/detail activity/news hanya melihat `published`; admin memiliki route list/detail untuk semua status; submission query menolak activity non-published.

**File:** [routes/routes.go](routes/routes.go), [internal/activities/handler.go](internal/activities/handler.go), [internal/news/handler.go](internal/news/handler.go), [internal/news/repository.go](internal/news/repository.go), [internal/activitysubmissions/repository.go](internal/activitysubmissions/repository.go)

**Test:** Unit tests untuk status filter dan visibility; `go test ./...` perlu dijalankan setelah implementasi lanjutan.

**Catatan:** Policy yang dipilih: draft admin-only. Member mendapat 404 untuk activity draft detail dan submission tidak memenuhi query.

**Test:** Validasi missing/zero/positive `user_id`; hasil full suite akhir dicatat pada bagian verification.
**Status:** Resolved

**Masalah:** Admin earn mengkredit admin pemanggil sendiri.

**Perubahan:** Request kini memerlukan `user_id` positif sebagai penerima; handler mengkredit target tersebut tanpa site restriction. Postman example diperbarui.

**File:** [internal/points/handler.go](internal/points/handler.go), [internal/points/handler_test.go](internal/points/handler_test.go), [docs/ChristAPI.postman_collection.json](docs/ChristAPI.postman_collection.json)

**Test:** Validasi missing/zero/positive `user_id`; `go test ./...` perlu dijalankan setelah implementasi lanjutan.

**Test:** Unit test eligibility dengan/tanpa attendance; hasil full suite akhir dicatat pada bagian verification.

### A-05 - Attendance eligibility untuk streak
**Status:** Resolved

**Masalah:** Streak points dapat diklaim tanpa attendance.

**Perubahan:** Check-in streak memerlukan attendance record user pada hari UTC yang sama sebelum menulis streak log atau points. Attendance tidak harus terkait activity yang sama.

**File:** [internal/streaks/repository.go](internal/streaks/repository.go), [internal/streaks/handler.go](internal/streaks/handler.go), [internal/streaks/repository_test.go](internal/streaks/repository_test.go)

**Test:** Middleware/routes tests dan full suite akan diverifikasi ulang setelah seluruh follow-up edits.

**Catatan:** Tanggal attendance memakai UTC, sedangkan period streak tetap memakai Asia/Jakarta sesuai source existing.

### A-10 - Live role revocation
**Status:** Resolved

**Masalah:** Admin role berasal dari JWT claim dan dapat stale hingga token expired.

**Perubahan:** `AuthMiddleware` membaca `role_id` live dari user row; `AdminOnly` memeriksa role context tersebut.

**File:** [internal/middleware/auth.go](internal/middleware/auth.go), [internal/middleware/admin.go](internal/middleware/admin.go)

**Test:** `go test ./...` akan diverifikasi ulang setelah seluruh follow-up edits.

**Catatan:** Perubahan role berlaku pada request berikutnya; tidak ada blacklist atau cache baru.

### A-13 - One-shot migration flow
**Status:** Resolved

**Masalah:** Migration profile/helper dapat keluar tanpa menjalankan `migrate` karena service profile tidak aktif dan command shell dipecah.

**Perubahan:** Semua helper mengaktifkan profile `manual-migration`; command Compose diteruskan sebagai satu argumen shell dan menunggu PostgreSQL healthy.

**File:** [docker-compose.yml](docker-compose.yml), [Makefile](Makefile), [dalamNamaTuhan.sh](dalamNamaTuhan.sh), [dalamNamaTuhan.ps1](dalamNamaTuhan.ps1), [resetDatabase.sh](resetDatabase.sh), [DOCKER.md](DOCKER.md)

**Test:** Aksi read-only `version` mengembalikan 20; aksi `up` menerapkan migration 21. `docker compose --profile manual-migration config --quiet` lulus.

**Catatan:** Migration tetap one-shot manual, bukan startup otomatis.

### A-03 - Google identity binding
**Status:** Resolved

**Masalah:** Login Google untuk akun existing mencocokkan email/provider tanpa memastikan subject Google sama.

**Perubahan:** Login dan Google username submission kini memerlukan `email_verified`; service membandingkan `sub` dengan `google_id` tersimpan.

**Test:** PostgreSQL migration berhasil; version `21`, `dirty=false`. FK metadata terverifikasi; orphan counts: enam relasi nol; `users.site_id` tetap satu orphan.

**Test:** Unit tests verified/unverified email dan subject sama/berbeda; `go test ./...` lulus.

**Catatan:** Akun Google lama yang tidak memiliki `google_id` tersimpan tidak dapat login sampai datanya ditinjau. Tidak ada account-linking otomatis.

### A-06 - Attendance report query
**Status:** Resolved

**Masalah:** Query memilih `users.full_name`, kolom yang tidak ada.

**Perubahan:** Query mengambil nama dengan `LEFT JOIN contacts` dan `COALESCE`.

**File:** [internal/attendance/repository.go](internal/attendance/repository.go)

**Test:** `go test ./...` lulus untuk compile; belum ada PostgreSQL integration test.

**Catatan:** Jalankan query pada schema migrated yang terisolasi sebelum release.

### A-07 - Activity UUID lookup
**Status:** Resolved

**Masalah:** Lookup hanya mencari dalam 100 activity terbaru.

**Perubahan:** UUID menjadi filter query dan lookup meminta satu hasil; UUID kosong tetap menghasilkan not-found.

**File:** [internal/activities/model.go](internal/activities/model.go), [internal/activities/repository.go](internal/activities/repository.go)

**Test:** `go test ./...` lulus; belum ada PostgreSQL repository test.

**Catatan:** Scanner dan response detail yang ada tetap dipakai.

### A-08 - News affected rows
**Status:** Resolved

**Masalah:** Update/delete mengembalikan sukses ketika UUID tidak cocok.

**Perubahan:** Memeriksa `RowsAffected`, mengembalikan `sql.ErrNoRows` untuk record hilang/soft-deleted, dan handler memetakan ke 404.

**File:** [internal/news/repository.go](internal/news/repository.go), [internal/news/handler.go](internal/news/handler.go)

**Test:** Package lulus `go test`; belum ada PostgreSQL test untuk row count.

**Catatan:** Tidak mengubah success response.

### A-12 - Error detail exposure
**Status:** Resolved

**Masalah:** Response server error memuat tipe dan detail error internal.

**Perubahan:** Error tetap dicatat ke log dengan method/path/status; response 5xx sekarang memakai tipe generik dan detail aman.

**File:** [pkg/response/response.go](pkg/response/response.go), [pkg/response/response_test.go](pkg/response/response_test.go)

**Test:** Unit test memastikan pesan internal tidak muncul dalam body; lulus.

**Catatan:** Envelope JSON dipertahankan. Isi `detail` untuk 5xx berubah menjadi pesan aman.

### A-16 - Attendance down migration
**Status:** Resolved

**Masalah:** Tidak ada rollback untuk migration 14.

**Perubahan:** Menambahkan down migration yang menjatuhkan `attendance_records`.

**File:** [migrations/14_add_attendance_records.down.sql](migrations/14_add_attendance_records.down.sql)

**Test:** File SQL diverifikasi; migration tidak dijalankan pada database.

**Catatan:** Rollback menghapus seluruh data attendance. Pastikan backup dan approval sebelum menjalankan `down` pada database berisi data.

### A-17 - Invalid attendance dates
**Status:** Resolved

**Masalah:** Input tanggal invalid diam-diam dianggap tidak dikirim atau diganti tanggal hari ini.

**Perubahan:** Semua endpoint attendance yang menerima tanggal kini mengembalikan 422 untuk format/tanggal invalid.

**File:** [internal/attendance/handler.go](internal/attendance/handler.go), [internal/attendance/handler_test.go](internal/attendance/handler_test.go)

**Test:** Unit test kosong/valid/invalid date; lulus.

**Catatan:** Parameter tanggal kosong tetap menggunakan default yang lama.

### A-19 - News rows iteration error
**Status:** Resolved

**Masalah:** News list tidak memeriksa error setelah iterasi rows.

**Perubahan:** Repository sekarang mengembalikan `rows.Err()`.

**File:** [internal/news/repository.go](internal/news/repository.go)

**Test:** `go test ./...` lulus; error iterasi belum diuji dengan database/fault injection.

**Catatan:** Tidak mengubah response normal.

## Finding yang Sebagian Diperbaiki

### A-01 - Secret dalam Git
**Status:** Partially Resolved

**Masalah:** `.env.local` dan `.env.docker` dilacak Git dan memuat konfigurasi sensitif; helper juga menanam/mencetak credential.

**Perubahan:** Kedua file dikeluarkan dari Git index tetapi salinan lokal dipertahankan dan di-ignore. Ditambahkan `.env.docker.example`, secret placeholder di `.env.example` dikosongkan, dan helper/guide tidak lagi memuat atau mencetak password literal.

**File:** [.gitignore](.gitignore), [.env.example](.env.example), [.env.docker.example](.env.docker.example), [Makefile](Makefile), [dalamNamaTuhan.sh](dalamNamaTuhan.sh), [dalamNamaTuhan.ps1](dalamNamaTuhan.ps1), [resetDatabase.sh](resetDatabase.sh), [SETUP.md](SETUP.md), [DOCKER.md](DOCKER.md), [QUICKSTART.md](QUICKSTART.md)

**Test:** `git ls-files .env.local .env.docker` tidak menghasilkan nama; `git check-ignore` mengonfirmasi keduanya di-ignore; kedua file lokal tetap ada.

**Catatan:** Nilai lama tetap dapat diakses melalui Git history. Rotasi credential dan pembersihan history tidak dilakukan; rotasi diperlukan bila nilai pernah dipakai di shared/staging/production. Migration scripts tidak menampilkan nilai secret.

### A-09 - List pagination limits
**Status:** Partially Resolved

**Masalah:** News/contact menerima limit yang tidak dibatasi; submission/redemption list tidak berpaginasi.

**Perubahan:** News dan contacts mempertahankan default lama, membatasi maksimum 100, dan menormalkan negative offset pada news.

**File:** [internal/news/repository.go](internal/news/repository.go), [internal/contacts/repository.go](internal/contacts/repository.go)

**Test:** Package tests/build lulus; tidak ada test SQL pagination.

**Catatan:** Submission dan redemption tidak diubah karena batas hasil baru dapat memotong response dan consumer eksternal tidak diketahui.

### A-11 - Upload validation/persistence
**Status:** Partially Resolved

**Masalah:** Upload hanya memvalidasi nama/ukuran dan filesystem Docker tidak persisten.

**Perubahan:** Validator bersama memeriksa batas 5 MB, extension, serta `Content-Type` terdeteksi untuk JPEG/PNG/WebP; semua empat upload handler menggunakannya.

**File:** [internal/uploadimage/validate.go](internal/uploadimage/validate.go), [internal/uploadimage/validate_test.go](internal/uploadimage/validate_test.go), [internal/activities/handler.go](internal/activities/handler.go), [internal/news/handler.go](internal/news/handler.go), [internal/rewards/handler.go](internal/rewards/handler.go), [internal/contacts/handler.go](internal/contacts/handler.go)

**Test:** Unit tests untuk JPEG/PNG/WebP, mismatch, extension invalid, dan ukuran berlebih; lulus.

**Catatan:** Content tidak didecode penuh. Volume/persistence tidak ditambahkan karena deployment model belum dikonfirmasi.

### A-14 - CI test gate
**Status:** Partially Resolved

**Masalah:** CI tidak menjalankan test dan coverage business flow terbatas.

**Perubahan:** Workflow menambahkan `go test ./...`; test unit untuk identity, tanggal, upload, dan response ditambahkan.

**File:** [.github/workflows/ci.yml](.github/workflows/ci.yml), [internal/auth/google_test.go](internal/auth/google_test.go), [internal/attendance/handler_test.go](internal/attendance/handler_test.go), [internal/uploadimage/validate_test.go](internal/uploadimage/validate_test.go), [pkg/response/response_test.go](pkg/response/response_test.go)

**Test:** `go test ./...` lulus.

**Catatan:** Belum ada integration tests PostgreSQL untuk query, constraints, transaction rollback, atau concurrency.

### A-15 - Schema bootstrap drift
**Status:** Partially Resolved

**Masalah:** Snapshot legacy menjadi jalur schema alternatif dari migrations.

**Perubahan:** Compose tidak lagi mount `docs/schema.sql`; SETUP/DOCKER menjelaskan migrations sebagai sumber schema dan snapshot sebagai legacy.

**File:** [docker-compose.yml](docker-compose.yml), [SETUP.md](SETUP.md), [DOCKER.md](DOCKER.md), [QUICKSTART.md](QUICKSTART.md)

**Test:** `docker compose config --quiet` lulus.

**Catatan:** Snapshot dipertahankan. Database volume existing tidak diperiksa atau diubah; evaluasi terpisah diperlukan sebelum migrasi volume lama.

### A-20 - Foreign key coverage
**Status:** Partially Resolved

**Masalah:** Migration chain tidak menegakkan beberapa relasi nullable di users, contacts, roles, dan news.

**Perubahan:** Migration 21 menambahkan tujuh FK `NOT VALID` dengan `ON DELETE SET NULL`, dan melewati constraint legacy yang cocok. Migration diterapkan pada DB disposable.

**File:** [migrations/21_add_missing_foreign_keys.up.sql](migrations/21_add_missing_foreign_keys.up.sql), [migrations/21_add_missing_foreign_keys.down.sql](migrations/21_add_missing_foreign_keys.down.sql)

**Test:** PostgreSQL migration berhasil; version `21`, `dirty=false`. Orphan counts: enam relasi nol; `users.site_id` tetap satu orphan.

**Catatan:** Row orphan tidak diubah. FK `users.site_id` dan seluruh constraint baru masih `NOT VALID`; petakan/bersihkan row sesuai data owner sebelum `VALIDATE CONSTRAINT`.

## Finding yang Masih Blocked

Tidak ada finding yang sepenuhnya blocked setelah policy dikonfirmasi. A-20 masih partial sampai orphan site reference ditangani dan constraints divalidasi.

## Finding yang Tidak Diperbaiki

### A-18 - Response/error consistency
**Status:** Won't Fix

**Masalah:** Format/key response berbeda antar-module.

**Perubahan:** Tidak menyeragamkan response lama.

**File:** Tidak berubah.

**Test:** Tidak berlaku.

**Catatan:** Perubahan tidak memiliki bukti kontrak client dan dapat menjadi breaking; response existing dipertahankan.

## Perubahan File

- Configuration/secrets: [.env.example](.env.example), [.env.docker.example](.env.docker.example), [.gitignore](.gitignore), staged removal dari index [.env.local](.env.local) dan [.env.docker](.env.docker). Salinan kedua file env tetap lokal.
- Docker/setup: [docker-compose.yml](docker-compose.yml), [Makefile](Makefile), [dalamNamaTuhan.sh](dalamNamaTuhan.sh), [dalamNamaTuhan.ps1](dalamNamaTuhan.ps1), [resetDatabase.sh](resetDatabase.sh), [SETUP.md](SETUP.md), [DOCKER.md](DOCKER.md), [QUICKSTART.md](QUICKSTART.md).
- Source: auth, activities, attendance, contacts, news, rewards, uploadimage, response.
- Migration/CI: [migrations/14_add_attendance_records.down.sql](migrations/14_add_attendance_records.down.sql), [.github/workflows/ci.yml](.github/workflows/ci.yml).
- Documentation: [PROJECT_ANALYSIS.md](PROJECT_ANALYSIS.md), [CODEBASE_AUDIT.md](CODEBASE_AUDIT.md), report ini.
- User data directory `uploads/` sudah untracked sebelum implementasi dan tidak diubah.

## Test yang Ditambahkan

- Google verified email dan subject match/mismatch.
- Member/admin draft status filtering.
- Admin points target ID validation.
- Live admin role source is database-backed.
- Attendance prerequisite for streak check-in.
- Attendance date parsing untuk kosong/valid/invalid.
- MIME/extension/size validation untuk upload image.
- Error response tidak mengembalikan pesan internal.
- Route inventory test memastikan semua method/path API terdaftar; seluruh protected route diuji menolak request tanpa token, dan enam public auth route diuji mencapai validasi request.
- Migration 21 dijalankan pada PostgreSQL disposable; belum ada test repository PostgreSQL untuk query lain.

**Batas cakupan route test:** ini memverifikasi registrasi route dan boundary auth tanpa database. Test ini belum membuktikan business behavior/success path setiap handler, authorization token non-admin terhadap route admin, ataupun response/query repository terhadap PostgreSQL.

## Test Verification

| Perintah | Hasil |
|---|---|
| `go test ./...` | Lulus |
| `go test ./routes` | Lulus; route inventory, protected-route unauthenticated matrix, dan public-auth validation |
| `go vet ./...` | Lulus |
| `gofmt -l .` | Tidak ada file yang perlu diformat |
| `go build ./...` | Lulus |
| `docker compose config --quiet` | Lulus |
| `bash -n dalamNamaTuhan.sh` | Lulus |
| `bash -n resetDatabase.sh` | Lulus |
| `git diff --check` | Lulus |

PowerShell helper tidak diuji karena environment ini Linux. PostgreSQL migrations/query tidak dijalankan terhadap database untuk menghindari perubahan pada database yang mungkin berisi data pengguna.

## Database/Migration Changes

Database disposable lokal berhasil dinaikkan dari version 20 ke 21 (`dirty=false`). Migration 21 menambahkan tujuh FK `NOT VALID`; satu orphan `users.site_id` dipertahankan tanpa perubahan dan membuat constraint belum dapat divalidasi. Migration 14 mendapat down migration `DROP TABLE`, yang akan menghapus seluruh attendance data. Schema bootstrap Docker menggunakan versioned migrations; snapshot tetap legacy.

## API Changes

- Request bodies dan success response tidak berubah.
- Error response 5xx mempertahankan envelope, tetapi `errors.type/detail` kini generik dan tidak membocorkan error internal.
- Google login/username submission kini menolak claim email yang tidak verified atau subject yang tidak cocok.
- Member activity/news hanya membaca `published`; admin memiliki read routes untuk semua status; submission ke draft ditolak.
- `POST /api/admin/points/earn` kini memerlukan `user_id` target.
- Streak check-in mensyaratkan attendance pada UTC date yang sama; tanpa attendance mengembalikan 409.
- Admin role divalidasi live dari user row; perubahan role memengaruhi request berikutnya.
- News update/delete mengembalikan 404 untuk UUID yang tidak ditemukan/soft-deleted.
- Attendance menolak tanggal invalid dengan 422.
- News/contact list limit di atas 100 dibatasi menjadi 100.
- Upload image dengan extension dan content yang tidak cocok ditolak.

## Breaking Changes

Request `/api/admin/points/earn` berubah dan kini wajib menyertakan `user_id`. Member draft visibility, streak prerequisite, role revocation, Google claims, invalid-date handling, 5xx details, dan maximum list limit juga berubah. Client perlu menyesuaikan flow yang memakai endpoint tersebut; success payload points tetap sama dan Postman example diperbarui.

## Risiko yang Masih Tersisa

- Credential lama masih dapat berada di Git history; file hanya dikeluarkan dari index.
- Belum ada PostgreSQL integration test untuk query report, affected rows, transaction, locking, atau migration.
- Upload masih disimpan pada filesystem container tanpa volume persisten.
- Satu orphan `users.site_id` tetap ada; constraint terkait `NOT VALID` sampai owner memutuskan mapping yang benar.
- Submission/redemption list masih tidak berpaginasi.
- Migration down attendance bersifat destructive.

## Hal yang Masih Membutuhkan Konfirmasi

- Rotasi credential dan apakah history repository perlu dibersihkan sesuai prosedur organisasi.
- Apakah filesystem upload harus persisten dan model deployment/storage yang dipakai.
- Apakah `site_id` menjadi security boundary pada routes selain points.
- Pemetaan/pembersihan orphan `users.site_id` sebelum validasi FK.
- Perilaku client terhadap batas pagination baru; migration/rollback deployment policy.

## Rekomendasi Langkah Berikutnya

1. Putar credential yang pernah digunakan bersama/di luar local development; lakukan sesuai prosedur secret management dan jangan mengandalkan penghapusan working tree.
2. Petakan orphan `users.site_id`, putuskan nilai site yang benar, lalu validate constraint setelah data owner menyetujui perubahan.
3. Siapkan database PostgreSQL test terisolasi untuk integration tests query/transaction selain migration 21 yang sudah diuji.
4. Tentukan persistence upload berdasarkan target deployment sebelum menambahkan volume/storage backend.
5. Lakukan smoke test bootstrap Docker pada database disposable baru; jangan reset database pengguna.
