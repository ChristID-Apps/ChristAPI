# Codebase Audit

**Cakupan:** audit statis terhadap source Go, route, query SQL, migration, konfigurasi, Docker, CI, test, dan dokumentasi yang ada di repository pada 1 Oktober 2026. Database tidak diubah dan source code tidak disentuh.

**Klasifikasi evidence:** **Terverifikasi** berarti perilaku atau kondisi terlihat langsung di repository. **Terindikasi** berarti risiko didukung source tetapi dampak akhirnya bergantung pada input atau lingkungan. **Perlu Konfirmasi** berarti keputusan produk/deployment diperlukan sebelum menganggapnya sebagai bug.

**Status verifikasi awal:** audit awal menjalankan `go test`, `go vet`, dan `gofmt`; implementasi lanjutan serta hasil test/build terbaru dicatat pada [IMPLEMENTATION_REPORT.md](IMPLEMENTATION_REPORT.md).

## Ringkasan

Repository ini adalah backend Go/Fiber dengan SQL langsung melalui `database/sql` dan PostgreSQL. Struktur package per domain masih sesuai ukuran project; audit tidak menemukan alasan untuk mengganti arsitektur atau menambah framework.

Perbaikan implementasi menutup draft visibility untuk member, Google subject matching, admin points target, attendance-backed streak, live role revocation, dan schema FK baru. Risiko tersisa yang penting: secret lama masih ada di Git history, upload filesystem belum persisten, dan satu orphan `users.site_id` membuat FK `NOT VALID` belum dapat divalidasi penuh.

P0 tidak ditemukan dari pemeriksaan repository ini. Prioritas bukan penilaian kualitas project; prioritas berikut hanya mengurutkan risiko yang memiliki bukti.

## Temuan P0 - Critical

Tidak ada temuan P0 yang dapat dibuktikan dari source dan konfigurasi yang diperiksa.

## Temuan P1 - High

### A-01 - Secret runtime dilacak di Git

- **Priority:** P1 - High
- **Status:** Terverifikasi; penggunaan di luar development perlu konfirmasi
- **Category:** Security, Configuration
- **Lokasi:** [.env.docker](.env.docker), [`.gitignore`](.gitignore), [dalamNamaTuhan.sh](dalamNamaTuhan.sh), [Makefile](Makefile)
- **Masalah:** `.env.docker` dilacak Git dan berisi variabel `JWT_SECRET`, `DB_PASSWORD`, serta kredensial PostgreSQL. `.gitignore` secara eksplisit mengecualikan `.env.docker` dari ignore. Helper migrasi juga menyimpan connection string dengan credential literal.
- **Evidence:** `git ls-files` memasukkan `.env.docker`; nilai tidak disalin ke laporan. `docker-compose.yml` memakai file itu untuk API, database, dan migration container.
- **Dampak:** siapa pun yang dapat membaca repository dapat memperoleh credential tersebut. Jika `JWT_SECRET` yang sama dipakai pada deployment yang dapat diakses, token aplikasi berpotensi dipalsukan; credential database dapat membuka akses ke data.
- **Alasan penting:** status repository privat tidak menghilangkan risiko bila akses meluas, file masuk ke mirror/backup, atau credential development dipakai ulang.
- **Rekomendasi arah perbaikan:** pindahkan nilai sensitif ke environment lokal yang tidak dilacak atau secret manager; sediakan template tanpa nilai rahasia. Rotasi nilai yang pernah dipakai di lingkungan bersama/produksi. Hindari mengulang password di script/Makefile.
- **Effort:** Small
- **Confidence:** High untuk keterlacakan; Medium untuk dampak runtime karena deployment tidak diketahui.

### A-02 - Pembaca member dapat mengambil konten draft

- **Priority:** P1 - High
- **Status:** Resolved setelah policy dikonfirmasi
- **Category:** Authorization, Data exposure
- **Lokasi:** [routes/routes.go](routes/routes.go), [internal/activities/repository.go](internal/activities/repository.go), [internal/news/repository.go](internal/news/repository.go), [internal/activitysubmissions/repository.go](internal/activitysubmissions/repository.go)
- **Masalah:** sebelum fix, list member tidak membatasi status; detail UUID juga dapat membaca draft dan submission menerima activity draft.
- **Evidence:** member handler memaksa `published`; admin memiliki list/detail routes tersendiri; submission query mensyaratkan `a.status = 'published'`.
- **Dampak:** member tidak dapat membaca atau mengirim submission untuk draft; admin tetap dapat mengakses status lain.
- **Alasan penting:** product owner mengonfirmasi draft hanya untuk admin.
- **Rekomendasi arah perbaikan:** filter diterapkan di query/member boundary agar parameter caller tidak dapat melonggarkan policy.
- **Effort:** Small
- **Confidence:** High bahwa query saat ini mengembalikan semua status; Medium untuk sensitivitas data menurut produk.

### A-03 - Google login tidak mengikat ulang token ke subject tersimpan

- **Priority:** P1 - High
- **Status:** Terindikasi
- **Category:** Authentication, Account takeover risk
- **Lokasi:** [internal/auth/handler.go](internal/auth/handler.go), [internal/auth/service.go](internal/auth/service.go), [internal/auth/repository.go](internal/auth/repository.go)
- **Masalah:** handler memvalidasi tanda tangan/audience ID token Google dan mengambil `email` serta `sub`. Pada login, service mencari user dengan email dan hanya memastikan `AuthProvider == "google"`; ia tidak membandingkan `googleID` dari token dengan `user.GoogleID`. Claim `email_verified` juga tidak diperiksa. Pembandingan subject dilakukan di flow submit username, tetapi tidak di flow login yang sudah terdaftar.
- **Evidence:** `GoogleLoginOrRegister(email, googleID, ...)` tidak memakai argumen `googleID` setelah lookup; request email dari body tidak dipakai sebagai identitas, tetapi claim email dari token menjadi kunci lookup.
- **Dampak:** bila token Google valid dapat membawa email yang tidak terverifikasi atau berubah, code path berpotensi mengautentikasi user Google lain yang memiliki email sama tanpa memastikan subject pemilik akun.
- **Alasan penting:** ini menyentuh batas identitas login. Validitas token saja tidak membuktikan kecocokan dengan subject yang sebelumnya dipautkan ke akun lokal.
- **Rekomendasi arah perbaikan:** bandingkan subject token dengan `google_id` yang tersimpan untuk akun existing, dan terapkan aturan `email_verified` yang sesuai dokumentasi Google sebelum memakai email untuk account linking.
- **Effort:** Small
- **Confidence:** Medium; eksploitabilitas bergantung pada claim dan perilaku identity provider.

## Temuan P2 - Medium

### A-04 - Pemberian poin melalui route admin hanya menambah saldo admin pemanggil

- **Priority:** P2 - Medium
- **Status:** Resolved setelah kontrak API dikonfirmasi
- **Category:** Business flow, Data integrity
- **Lokasi:** [routes/routes.go](routes/routes.go), [internal/points/handler.go](internal/points/handler.go), [internal/points/service.go](internal/points/service.go)
- **Masalah:** sebelum fix, target user diambil dari JWT admin.
- **Evidence:** request kini memerlukan `user_id` positif dan service mengkredit user tersebut; contoh Postman telah diperbarui.
- **Dampak:** poin masuk ke anggota yang dipilih; validasi admin tetap berjalan.
- **Alasan penting:** alur finansial internal (saldo dan ledger) dapat berbeda dari maksud operasional meskipun transaction menjaga konsistensi database.
- **Rekomendasi arah perbaikan:** consumer mengirim `user_id`; site scope tidak diterapkan sesuai keputusan bahwa `site_id` tidak dipakai sebagai policy.
- **Effort:** Small
- **Confidence:** High untuk perilaku aktual; Medium untuk mismatch terhadap business requirement.

### A-05 - Check-in streak memberi poin tanpa bukti kehadiran

- **Priority:** P2 - Medium
- **Status:** Resolved setelah eligibility dikonfirmasi
- **Category:** Business rule, Points integrity
- **Lokasi:** [routes/routes.go](routes/routes.go), [internal/streaks/handler.go](internal/streaks/handler.go), [internal/streaks/repository.go](internal/streaks/repository.go)
- **Masalah:** sebelum fix, streak check-in tidak membutuhkan attendance.
- **Dampak:** streak log dan points kini hanya diproses jika user sudah memiliki attendance record pada UTC date hari itu.
- **Alasan penting:** product owner memilih attendance harian apa pun sebagai eligibility; activity ID attendance tidak harus sama.
- **Rekomendasi arah perbaikan:** client memanggil attendance check-in lebih dulu; streak yang belum eligible menerima 409.
- **Effort:** Medium
- **Confidence:** High untuk flow yang ada; Low untuk pelanggaran policy karena policy tidak ditemukan.

### A-06 - Admin attendance report memilih kolom yang tidak ada

- **Priority:** P2 - Medium
- **Status:** Terverifikasi
- **Category:** Correctness, SQL
- **Lokasi:** [internal/attendance/repository.go](internal/attendance/repository.go), [migrations/1_init.up.sql](migrations/1_init.up.sql), [docs/schema.sql](docs/schema.sql)
- **Masalah:** query `GetAdminReport` memilih `u.full_name` dari `users`. Migration dan schema snapshot menyimpan nama pada `contacts.full_name`; table `users` tidak mendefinisikan `full_name`.
- **Evidence:** migration awal mendefinisikan kolom `users` tanpa `full_name`; snapshot hanya mendefinisikan `contacts.full_name`.
- **Dampak:** endpoint `GET /api/admin/attendance` gagal saat query dijalankan terhadap schema repository.
- **Alasan penting:** laporan admin attendance tidak berfungsi pada skema yang dimaksud.
- **Rekomendasi arah perbaikan:** join `contacts` melalui `users.contact_id` dan tetapkan perilaku bila contact tidak ada/soft-deleted; tambahkan query integration test bila test database tersedia.
- **Effort:** Small
- **Confidence:** High

### A-07 - Lookup aktivitas berhenti pada 100 hasil terbaru

- **Priority:** P2 - Medium
- **Status:** Terverifikasi
- **Category:** Correctness, Query
- **Lokasi:** [internal/activities/repository.go](internal/activities/repository.go)
- **Masalah:** `GetByUUID` memanggil list dengan `Limit: 100`, lalu mencari UUID dalam hasil tersebut. List diurutkan berdasarkan `created_at DESC`, sehingga aktivitas lebih lama dari 100 item pertama tidak dapat ditemukan melalui endpoint detail meskipun masih ada.
- **Dampak:** `GET /api/activities/:uuid` dapat melaporkan aktivitas valid tidak ditemukan ketika jumlah aktivitas melewati batas tersebut.
- **Alasan penting:** batas pagination list bocor ke lookup record tunggal.
- **Rekomendasi arah perbaikan:** gunakan query langsung berdasarkan UUID dengan kondisi deleted yang sama.
- **Effort:** Small
- **Confidence:** High

### A-08 - Update/delete news dapat memberi sukses untuk UUID yang tidak ada

- **Priority:** P2 - Medium
- **Status:** Terverifikasi
- **Category:** Correctness, API consistency
- **Lokasi:** [internal/news/repository.go](internal/news/repository.go), [internal/news/handler.go](internal/news/handler.go)
- **Masalah:** repository `Update` dan `SoftDelete` hanya memeriksa error `Exec`, tidak memeriksa `RowsAffected`. Handler karena itu mengembalikan sukses meski tidak ada baris dengan UUID tersebut; delete berulang juga tetap sukses.
- **Dampak:** client/admin tidak dapat membedakan perubahan berhasil dari target yang tidak ada atau sudah tidak relevan.
- **Alasan penting:** hasil HTTP tidak merepresentasikan keadaan database.
- **Rekomendasi arah perbaikan:** periksa `RowsAffected` dan kembalikan `sql.ErrNoRows` bila tidak ada record yang berubah, konsisten dengan pola `UpdateImage`.
- **Effort:** Small
- **Confidence:** High

### A-09 - Beberapa list tidak membatasi jumlah hasil

- **Priority:** P2 - Medium
- **Status:** Terverifikasi; dampak skala perlu konfirmasi
- **Category:** Performance, Reliability
- **Lokasi:** [internal/news/repository.go](internal/news/repository.go), [internal/activitysubmissions/repository.go](internal/activitysubmissions/repository.go), [internal/rewards/repository.go](internal/rewards/repository.go), [internal/contacts/repository.go](internal/contacts/repository.go)
- **Masalah:** news menerima limit caller tanpa batas atas; daftar submission dan redemption tidak menggunakan pagination. `contacts.List` menerima limit berapa pun dan menghitung offset dari page/limit. Sejumlah endpoint dapat mengambil dan memetakan seluruh hasil ke memory.
- **Dampak:** query/payload dapat membesar seiring data; `limit` atau offset sangat besar menambah beban DB dan memory. Volume produksi tidak tersedia untuk mengukur bottleneck saat ini.
- **Alasan penting:** berbeda dari optimasi spekulatif, source secara langsung mengizinkan beban hasil yang tidak dibatasi di beberapa endpoint.
- **Rekomendasi arah perbaikan:** tambahkan pagination dan batas maksimum pada endpoint yang relevan setelah menyepakati kontrak API; lakukan bertahap berdasarkan volume data.
- **Effort:** Medium
- **Confidence:** High untuk ketiadaan batas; Medium untuk dampak operasional.

### A-10 - Google role dan status authorization menggunakan claim token lama

- **Priority:** P2 - Medium
- **Status:** Resolved
- **Category:** Authorization
- **Lokasi:** [internal/middleware/auth.go](internal/middleware/auth.go), [internal/middleware/admin.go](internal/middleware/admin.go), [pkg/jwt/jwt.go](pkg/jwt/jwt.go)
- **Masalah:** sebelum fix, `AdminOnly` menggunakan role ID claim token.
- **Dampak:** role change sekarang dibaca dari row user pada setiap request authenticated, sehingga token lama tidak mempertahankan akses admin setelah request berikutnya.
- **Alasan penting:** product owner meminta privilege dicabut segera.
- **Rekomendasi arah perbaikan:** pertahankan live role lookup; tidak diperlukan token blacklist.
- **Effort:** Small
- **Confidence:** Medium; belum ada endpoint khusus yang mengubah role user yang terlihat dalam route saat ini.

### A-11 - Upload hanya memvalidasi nama file dan tidak persisten di Docker

- **Priority:** P2 - Medium
- **Status:** Terverifikasi
- **Category:** Security, Reliability, Deployment
- **Lokasi:** [internal/activities/handler.go](internal/activities/handler.go), [internal/news/handler.go](internal/news/handler.go), [internal/rewards/handler.go](internal/rewards/handler.go), [internal/contacts/handler.go](internal/contacts/handler.go), [Dockerfile](Dockerfile), [docker-compose.yml](docker-compose.yml)
- **Masalah:** handler membatasi ukuran dan ekstensi, tetapi tidak mendeteksi MIME/content atau decode gambar. URL hasil upload disajikan publik dari `/uploads`. Image Docker runtime hanya menyalin binary dan Compose tidak memasang volume untuk `uploads`.
- **Dampak:** file non-image dapat disimpan dengan ekstensi gambar; file upload hilang saat container API dibuat ulang dan file yang sudah ada di host tidak tersedia di runtime container. Risiko content-sniffing bergantung pada cara file dikonsumsi browser/client.
- **Alasan penting:** upload merupakan fitur yang terlihat API, sementara umur file mengikuti umur container.
- **Rekomendasi arah perbaikan:** validasi content sesuai format yang diizinkan. Tentukan penyimpanan persisten (misalnya volume lokal terkelola) hanya jika deployment memang mengandalkan upload lokal; dokumentasikan backup dan akses file.
- **Effort:** Small untuk validasi; Medium untuk persistensi
- **Confidence:** High

### A-12 - Error detail internal dikirim kembali ke client

- **Priority:** P2 - Medium
- **Status:** Terverifikasi
- **Category:** Security, Error handling
- **Lokasi:** [pkg/response/response.go](pkg/response/response.go), pemanggil `ErrorDetail` di handler
- **Masalah:** `ErrorDetail` mengirim tipe error dan `err.Error()` ke response, sekaligus menulis error mentah ke log. Beberapa query/handler meneruskan error database atau provider secara langsung.
- **Dampak:** pesan database, nama objek/kolom, atau detail provider dapat terekspos ke pemanggil ketika operasi gagal. Seberapa sensitif pesan tersebut bergantung pada error yang terjadi.
- **Alasan penting:** error internal tidak selalu aman sebagai kontrak API publik.
- **Rekomendasi arah perbaikan:** gunakan pesan client yang stabil dan simpan detail lengkap di log terkontrol; tetap pertahankan error domain yang memang aman untuk user.
- **Effort:** Medium
- **Confidence:** High untuk mekanisme; Medium untuk isi data yang terekspos pada runtime.

### A-13 - Docker Compose memulai service migration tanpa argumen migration

- **Priority:** P2 - Medium
- **Status:** Resolved dan diuji pada DB disposable
- **Category:** Configuration, Reliability
- **Lokasi:** [docker-compose.yml](docker-compose.yml), [Makefile](Makefile), [dalamNamaTuhan.sh](dalamNamaTuhan.sh)
- **Masalah:** profile migration tidak aktif secara default dan command shell sebelumnya dipecah menjadi beberapa argumen sehingga helper dapat exit tanpa menjalankan migration.
- **Evidence:** helper sekarang mengaktifkan `manual-migration`; aksi `version` membaca version 20 dan `up` berhasil menerapkan migration 21.
- **Dampak:** migration one-shot berjalan hanya saat diminta dan menunggu PostgreSQL healthy.
- **Rekomendasi arah perbaikan:** gunakan helper atau `docker compose --profile manual-migration run --rm migrate`.
- **Effort:** Small
- **Confidence:** Medium-High; command container tidak dieksekusi dalam audit.

### A-14 - Test business flow belum menjadi gate CI

- **Priority:** P2 - Medium
- **Status:** Terverifikasi
- **Category:** Testing, Reliability
- **Lokasi:** [.github/workflows/ci.yml](.github/workflows/ci.yml), [internal/middleware/rate_limit_test.go](internal/middleware/rate_limit_test.go)
- **Masalah:** CI menjalankan `go vet`, pemeriksaan `gofmt`, dan `go build`, tetapi tidak menjalankan `go test`. Repository memiliki satu file test Go untuk rate limiter; package business lainnya dilaporkan `no test files` oleh `go test ./...`.
- **Evidence:** test dijalankan selama audit dan lulus, tetapi tidak menguji business flow utama. Tidak ada bukti integration test DB atau test authorization pada workflow.
- **Dampak:** perubahan pada auth, transaksi points/rewards, status submission, dan query SQL dapat lolos CI tanpa regresi perilaku terdeteksi.
- **Alasan penting:** lulus build/vet tidak memvalidasi hasil transaksi dan aturan business.
- **Rekomendasi arah perbaikan:** tambahkan test terarah bersama perubahan behavior; pertimbangkan menambahkan `go test ./...` sebagai gate CI setelah test stabil. Integration test SQL memerlukan strategi database yang eksplisit.
- **Effort:** Medium
- **Confidence:** High

### A-15 - Snapshot schema dan instruksi setup tidak selaras dengan migration terkini

- **Priority:** P2 - Medium
- **Status:** Terverifikasi
- **Category:** Documentation, Database, Deployment
- **Lokasi:** [docs/schema.sql](docs/schema.sql), [SETUP.md](SETUP.md), [docker-compose.yml](docker-compose.yml), [migrations/](migrations/)
- **Masalah:** snapshot memuat tabel Bible legacy dan tidak memuat seluruh schema fitur saat ini.
- **Dampak:** snapshot saja tidak mendukung seluruh query API. Compose kini tidak lagi memuatnya dan setup menjadikan migrations satu-satunya jalur bootstrap.
- **Alasan penting:** setup lokal/CI dapat gagal atau berbeda dari database yang dibuat lewat seluruh migration.
- **Rekomendasi arah perbaikan:** gunakan migrations untuk bootstrap baru; simpan snapshot hanya sebagai artifact legacy.
- **Effort:** Medium
- **Confidence:** High

### A-16 - Migration attendance tidak memiliki rollback

- **Priority:** P2 - Medium
- **Status:** Terverifikasi
- **Category:** Database, Operations
- **Lokasi:** [migrations/14_add_attendance_records.up.sql](migrations/14_add_attendance_records.up.sql), [migrations/](migrations/)
- **Masalah:** migration 14 awalnya tidak memiliki pasangan down.
- **Dampak:** down migration kini menjatuhkan `attendance_records`, sehingga semua attendance data terhapus saat rollback.
- **Alasan penting:** rollback adalah bagian dari operasi release dan pemulihan schema.
- **Rekomendasi arah perbaikan:** jalankan rollback hanya setelah backup dan persetujuan; jangan gunakan sebagai rollback data-preserving.
- **Effort:** Small
- **Confidence:** High

## Temuan P3 - Low

### A-17 - Parser tanggal attendance mengubah input invalid menjadi default

- **Priority:** P3 - Low
- **Status:** Terverifikasi
- **Category:** Validation, API consistency
- **Lokasi:** [internal/attendance/handler.go](internal/attendance/handler.go)
- **Masalah:** `parseDateOrEmpty` mengembalikan string kosong untuk format tanggal yang invalid. Handler summary/history lalu menganggapnya tidak ada filter tanggal; admin report mengganti tanggal invalid dengan tanggal hari ini.
- **Dampak:** typo pada query tidak menghasilkan error validasi dan dapat mengembalikan periode/data berbeda dari yang diminta.
- **Alasan penting:** dapat membingungkan client dan menyamarkan input salah, tetapi tidak mengubah data.
- **Rekomendasi arah perbaikan:** bila kontrak API mengharapkan tanggal eksplisit, balas 422 untuk format invalid; konsistenkan perilaku default tanggal.
- **Effort:** Small
- **Confidence:** High

### A-18 - Beberapa response dan error mapping berbeda antar-module

- **Priority:** P3 - Low
- **Status:** Terverifikasi; dampak client perlu konfirmasi
- **Category:** Consistency, Maintainability
- **Lokasi:** [internal/points/handler.go](internal/points/handler.go), [pkg/response/response.go](pkg/response/response.go), [internal/middleware/rate_limit.go](internal/middleware/rate_limit.go)
- **Masalah:** response points memakai key `UserId`, `Balance`, `History` yang berbeda dari konvensi snake_case; rate limit mengembalikan bentuk JSON berbeda dari `ApiResponse`; beberapa handler juga memetakan error database menjadi 400/422/500 secara berbeda.
- **Dampak:** client perlu menangani variasi kontrak; perubahan konsistensi dapat menjadi breaking change bila client sudah bergantung pada key/status saat ini.
- **Alasan penting:** hanya layak diseragamkan saat kontrak client dan versi perubahan jelas; bukan alasan refactor massal.
- **Rekomendasi arah perbaikan:** pertahankan response existing untuk perubahan internal; tetapkan kontrak pada endpoint baru dan ubah yang lama hanya dengan kompatibilitas client yang direncanakan.
- **Effort:** Medium
- **Confidence:** High untuk perbedaan; Medium untuk biaya client.

### A-19 - Query news tidak memeriksa error iterasi rows

- **Priority:** P3 - Low
- **Status:** Terverifikasi
- **Category:** Reliability, SQL
- **Lokasi:** [internal/news/repository.go](internal/news/repository.go)
- **Masalah:** setelah loop `rows.Next()`, `List` mengembalikan hasil dengan `nil` tanpa memeriksa `rows.Err()`. Repository lain umumnya memeriksa error iterasi.
- **Dampak:** kegagalan saat membaca hasil dapat disajikan sebagai daftar sukses yang terpotong.
- **Alasan penting:** error iterasi berbeda dari error query awal dan perlu diteruskan agar client tidak menerima hasil parsial seolah lengkap.
- **Rekomendasi arah perbaikan:** periksa dan kembalikan `rows.Err()` setelah loop.
- **Effort:** Small
- **Confidence:** High

### A-20 - Constraint referensial tidak konsisten pada migration

- **Priority:** P3 - Low
- **Status:** Partially Resolved
- **Category:** Database, Data integrity
- **Lokasi:** [migrations/1_init.up.sql](migrations/1_init.up.sql), [migrations/](migrations/)
- **Masalah:** beberapa relasi aplikasi belum ditegakkan oleh migration chain.
- **Evidence:** migration 21 menambah tujuh FK `NOT VALID`, melewati FK legacy yang cocok, semuanya dengan `ON DELETE SET NULL`.
- **Dampak:** relasi baru divalidasi; orphan existing tidak diubah. Read-only check menemukan satu orphan `users.site_id` dan nol orphan pada enam relasi lain.
- **Alasan penting:** migration mempertahankan data, tetapi constraints belum membuktikan semua row existing valid.
- **Rekomendasi arah perbaikan:** petakan dan perbaiki orphan `users.site_id` sesuai data owner, lalu `VALIDATE CONSTRAINT`; down migration hanya melepas FK baru.
- **Effort:** Medium
- **Confidence:** High untuk migration; Medium untuk dampak.

## Temuan P4 - Optional

Tidak ada improvement P4 yang direkomendasikan secara spesifik. Tidak ada bukti kebutuhan untuk menambah cache, queue, ORM, repository interface, atau arsitektur baru.

## Architecture Audit

- **Terverifikasi:** entrypoint server menginisialisasi konfigurasi dan DB, lalu `routes.Setup` membuat repository/handler. Domain package kebanyakan memakai Handler → Service → Repository, tetapi submission mengakses repository langsung dan beberapa Service hanya meneruskan pemanggilan.
- **Terverifikasi:** repository memakai `database/sql` dan SQL eksplisit. Route wiring menggunakan `database.DB` global; dependency berupa concrete struct, tidak ada container DI.
- **Penilaian:** pola ini sederhana dan masih masuk akal untuk project sekarang. Coupling global mengurangi kemudahan test unit terisolasi, tetapi audit tidak menemukan bukti bahwa perubahan arsitektur besar diperlukan.
- **Tindakan:** jangan menambah interface/layer generik hanya untuk menyamakan semua package. Perbaiki bug pada boundary yang memiliki dampak nyata.

## Code Flow Audit

| Flow | Hasil audit |
|---|---|
| Credential registration → OTP | Contact dan user dibuat dalam transaction. OTP disimpan dan email dikirim setelah commit. Jika email gagal, akun tertinggal di `pending_otp`, tetapi endpoint resend tersedia. Pengiriman email dan penulisan OTP tidak atomic; perlu retry/operational handling. |
| OTP verify | Repository mengecek `EXISTS`, lalu menghapus OTP dalam query terpisah dan service mengubah status user setelahnya. Dua request concurrent berpotensi sama-sama membaca OTP valid sebelum penghapusan; efek dan policy konkurensi belum dites. |
| Login/logout | Password bcrypt dan status account diperiksa. Middleware membaca ulang state user, `last_logout_at`, dan role ID; perubahan role berlaku pada request berikutnya. Token HS256 berlaku 24 jam. |
| Google login | ID token Google diverifikasi, tetapi pencocokan akun existing memakai email/provider tanpa membandingkan subject (A-03). Username submission memiliki pemeriksaan subject. |
| Approval | Route berada di admin group. Service menerima user ID path dan mengubah state; approval tidak membatasi scope site. Apakah admin boleh approve akun dari semua site perlu konfirmasi. |
| Activity/Bible submission | Member list/detail hanya melihat `published`; admin mendapat list/detail semua status. Submission menolak activity non-published, lalu memeriksa config, reflection, dan uniqueness. Registration/occurrence enforcement tidak ditambahkan. |
| Attendance check-in | Attendance row, saldo, dan ledger diperbarui dalam transaction; unique `(user_id, attendance_date)` mencegah double row. Pemeriksaan awal `SELECT ... FOR UPDATE` tidak mengunci row yang belum ada; concurrent check-in kedua dapat berbenturan pada unique constraint dan diterjemahkan menjadi 500 alih-alih 409. Data dan poin tetap di-rollback oleh transaction. |
| Points earn/spend | Update saldo dan ledger berada dalam transaction dan row user dikunci. `/admin/points/earn` menerima `user_id` target eksplisit; spend tetap memakai ID authenticated user. |
| Reward redemption | Reward dan user dikunci, saldo/stock/ledger/redemption ditulis dalam satu transaction. Partial unique index mencegah lebih dari satu redemption pending per user/reward. Rejection mengembalikan saldo dan stock dalam transaction. |
| Streak check-in | Sebelum log/points, user wajib memiliki attendance record pada UTC attendance date hari itu. Period streak tetap dihitung memakai Asia/Jakarta. Log dan ledger berada di transaction. |
| News | Member hanya melihat berita `published`, admin dapat memilih status; update/delete memeriksa `RowsAffected`. |

## Database Audit

- Database yang dikonfigurasi adalah PostgreSQL; akses melalui `database/sql` + pgx. Migration adalah schema evolusi utama, sedangkan `docs/schema.sql` merupakan snapshot berbeda.
- Unique dan FK dipakai pada banyak tabel inti. Migration 21 menambahkan FK untuk `users.contact_id`, `users.role_id`, `users.site_id`, `contacts.site_id`, `roles.site_id`, `news.author_id`, dan `news.site_id`; FK baru `NOT VALID` untuk mempertahankan orphan lama. Database lokal kini version 21.
- Constraint unik attendance `(user_id, attendance_date)` menjaga satu check-in harian. Constraint unik submission `(activity_id, user_id)` menjaga satu submission per aktivitas/user. Migration 10 mengubah log streak agar aktivitas berbeda bisa tercatat pada tanggal sama.
- Balance points disimpan di `users.points_balance` dan direplikasi ke ledger `balance_after`. Flow penting umumnya melakukan lock dan transaction, tetapi tidak ada mekanisme rekonsiliasi saldo ledger dalam repository yang diperiksa.
- Migration 14 memiliki down migration yang menghapus seluruh data attendance. Snapshot legacy tidak dimuat saat bootstrap Compose.
- Migration 21 telah dijalankan pada database disposable yang disetujui user. Pemeriksaan read-only menemukan satu orphan `users.site_id`; constraint terkait belum tervalidasi.

## SQL & Query Audit

- Query input umumnya memakai placeholder PostgreSQL `$n`; dynamic SQL yang ditemukan menyusun nama kolom dari daftar literal internal, bukan input bebas.
- Temuan A-06/A-07/A-08/A-19 sudah diperbaiki pada source: attendance report join contacts, activity lookup filter UUID, news checks affected rows, news list returns `rows.Err()`.
- Query member news/activity dan activity submission kini membatasi `published` (A-02). News/contact limit dibatasi maksimum 100; submission/redemption lists tetap unpaginated (A-09 partial).
- Audit statis tidak menemukan interpolasi langsung parameter request ke nilai SQL pada query penting yang diperiksa; ini bukan bukti formal bahwa seluruh query bebas injection.
- Index ada untuk sebagian kolom tanggal/status/foreign key, tetapi tidak ada query plan atau ukuran data untuk membuktikan kekurangan index lain. Jangan menambahkan index tanpa workload evidence.

## Security Audit

- **Ada:** bcrypt untuk password; JWT memvalidasi algoritme HS256; status user/approval/logout divalidasi middleware; rate limit per IP; parameter SQL pada query yang diperiksa; CORS origin dapat dikonfigurasi; upload memiliki batas size dan extension.
- **Risiko tersisa:** secret terdahulu tetap ada di Git history (A-01 partial), upload belum persisten di Docker (A-11 partial), dan FK site belum tervalidasi karena orphan (A-20 partial).
- `ErrorDetail` mengirim detail error server kepada client (A-12). CORS bukan pengganti authorization; endpoint member tetap harus membatasi status dan ownership.
- Rate limiter menyimpan counter in-memory, jadi limit terpisah antar replica/restart. `c.IP()` dan trusted proxy behavior bergantung pada deployment; Compose tidak menunjukkan reverse proxy production.
- Tidak ada rate limit khusus per email/account untuk OTP; bucket OTP berlaku per IP. Konsekuensi brute force bergantung pada batas IP aktual dan proxy topology; angka konfigurasi tidak direproduksi di sini.
- Site ID diterima pada registration/contact/login dan beberapa query. Tidak ditemukan policy site-scope generik. Jika site adalah batas tenant/privasi, cross-site access dan penetapan site perlu pemeriksaan eksplisit sebelum dianggap aman.
- Tidak ada bukti pada repository tentang TLS termination, secret manager, backup, security headers, atau request-size policy production. Hal ini bukan bukti deployment production tidak memilikinya.

## Authentication & Authorization Audit

- Public routes: login/register/OTP/Google dan pembacaan Bible. Route bisnis utama memakai `AuthMiddleware`; `/admin` memakai tambahan `AdminOnly`.
- Middleware memastikan token valid, akun aktif/approved, dan role admin terkini dari database (A-10 resolved).
- `AdminOnly` mengizinkan role code `admin` atau `super_admin`, bukan policy matrix umum. Ini masih dapat dipertahankan selama kebutuhan RBAC sederhana.
- Handler member points membatasi user ke ID sendiri; reward redemptions member juga memakai user ID dari context. Jalur ownership yang diperiksa tidak menerima user ID member bebas.
- Tidak ada site-scoped authorization terpusat. Status sebagai admin tidak secara otomatis memeriksa site tertentu.
- Google login mencocokkan subject dan email verification (A-03 resolved); admin points memakai target `user_id` (A-04 resolved); streak memerlukan attendance harian (A-05 resolved).

## Code Consistency Audit

- Variasi Handler/Service/Repository dan penggunaan response berbeda antar modul ada, tetapi tidak seluruhnya menyebabkan bug. Ikuti pola lokal bila kontrak berjalan.
- Pagination tidak seragam: activity membatasi limit hingga 100; points membatasi pilihan limit; news, contact, submission, dan redemption memiliki batas berbeda atau tidak ada (A-09).
- Response JSON points memakai kapitalisasi berbeda, rate-limit punya envelope berbeda; perubahan dapat breaking dan tidak perlu diseragamkan tanpa kebutuhan client (A-18).
- Error mapping tidak seragam. Temuan berdampak yang terverifikasi dicatat terpisah (A-08, A-12, A-19), bukan menganggap seluruh variasi sebagai defect.

## Code Quality Audit

- Pemakaian SQL terparameterisasi, mapping nullable secara eksplisit, validasi domain pada beberapa service, dan transaction untuk mutasi saldo/reward adalah pola baik.
- Tidak ada indikasi yang mengharuskan refactor menyeluruh. Tidak adanya repository interface bukan masalah tersendiri pada struktur saat ini.
- Risiko maintainability yang berdampak: behavior/query coverage terbatas (A-14), pagination dan error mapping tidak seragam, serta beberapa flow langsung menggunakan global DB.
- Fungsi yang panjang atau variasi style tidak diangkat sebagai temuan kecuali menghasilkan risiko nyata yang dirinci di bagian lain.

## Transaction & Data Integrity Audit

- **Baik:** register contact+user, attendance+points+ledger, points delta+ledger, reward redeem/reject, dan streak+ledger berada dalam transaction.
- Reward redemption mengunci reward dan user sebelum perubahan saldo/stock; unique partial index melindungi redemption pending bersamaan.
- Attendance memeriksa row lama dengan `FOR UPDATE`, tetapi unique key merupakan perlindungan utama saat row belum ada. Race concurrent berakhir sebagai konflik DB yang tidak dipetakan ke `ErrAlreadyCheckedIn`; kemungkinan response 500, bukan double-credit.
- OTP delete/verify/state transition berjalan sebagai operasi terpisah, sehingga race tidak diuji. Risiko yang terlihat lebih pada konsistensi response/status daripada bukti account takeover.
- Saldo dan ledger tidak memiliki constraint database yang memastikan saldo `users` sama dengan agregasi ledger. Repository menjaga konsistensi pada flow yang diperiksa; rekonsiliasi/operasi manual tidak diketahui.

## Performance Audit

- Middleware melakukan query DB pada tiap request authenticated untuk validasi status/logout; trade-off ini memberi revocation cepat dengan satu query tambahan.
- A-09 mengidentifikasi daftar yang tidak dibatasi. Activity UUID lookup mengambil list 100 alih-alih query langsung (A-07).
- News search menggunakan `ILIKE` pada title/content; snapshot mencantumkan index full-text tetapi query repository juga memakai `ILIKE`. Efektivitas index perlu query plan/data volume dan tidak dapat dipastikan statis.
- Tidak ada evidence beban yang membenarkan cache, Redis, queue, atau service terpisah. Belum ada dasar untuk rekomendasi tersebut.

## Testing Audit

- `go test ./...`: lulus. Hanya `internal/middleware/rate_limit_test.go` menjalankan test; package lain melaporkan tidak ada test.
- `go vet ./...`: lulus tanpa output.
- `gofmt -l .`: tidak menampilkan file.
- `go build ./...`: tidak dijalankan karena tool menyatakan pemanggilan dilewati.
- CI menjalankan `go vet`, format check, dan build; tidak menjalankan test. Tidak ditemukan workflow integration test PostgreSQL atau test migration.
- Prioritas test berikutnya adalah query attendance report, status/authorization draft, Google account matching, ledger transaction failure, serta attendance concurrency. Tambahkan hanya test yang sesuai setup DB yang dipilih.

## Documentation & Schema Drift Audit

- `SETUP.md` menawarkan `docs/schema.sql` sebagai jalur setup manual, tetapi file tersebut tidak menggambarkan schema fitur terbaru pada migration (A-15).
- README/QUICKSTART/Postman memerlukan verifikasi kontrak terhadap `routes/routes.go`; dokumentasi lama menyebut pola `/api/v1` pada contoh tertentu sedangkan route saat ini terdaftar pada `/api`. Contoh login di SETUP menggunakan `/api/v1/auth/login`, sementara implementasi route adalah `/api/login`.
- `PROJECT_ANALYSIS.md` berguna sebagai indeks, tetapi audit ini tidak mengambil kesimpulannya sebagai fakta sebelum memeriksa source. Temuan dalam laporan ini didukung path implementasi/migration terkait.
- Migration 14 kini memiliki down migration (A-16). Tidak ada OpenAPI source yang ditemukan pada struktur yang diperiksa; cakupan dokumentasi kontrak API karenanya perlu dijaga manual.

## Configuration & Deployment Audit

- Compose menyajikan API pada port host yang dipetakan ke 3000 container, PostgreSQL dengan volume data, dan migration service terpisah. API menunggu healthcheck PostgreSQL.
- Compose menggunakan migrations sebagai bootstrap; snapshot SQL legacy tidak lagi dimuat (A-15).
- Service migration berada pada profile one-shot; helper mengaktifkan profile dan menunggu PostgreSQL healthy. Command shell yang sebelumnya exit tanpa menjalankan migration telah diperbaiki dan diuji terhadap migration 21 (A-13).
- Dockerfile runtime hanya menyalin binary, bukan `docs/` atau `uploads/`; `uploads` juga tidak dipasang sebagai volume (A-11). API dapat membuat direktori upload saat request, tetapi file tidak bertahan saat container diganti.
- `.env.docker` dilacak Git (A-01). Deployment sebenarnya, reverse proxy, TLS, storage persisten, dan secret rotation tidak dapat disimpulkan dari Compose.

## Technical Debt

| ID | Masalah dan penyebab | Dampak saat ini | Risiko bila dibiarkan | Kapan ditangani | Perbaikan minimal |
|---|---|---|---|---|---|
| A-01 | Secret file dikeluarkan dari index, tetapi history lama ada | Local files tetap ada dan di-ignore | Credential dapat tetap diakses dari history | Segera tentukan rotasi/history policy | Rotasi bila dipakai bersama; bersihkan history hanya sesuai prosedur |
| A-02/A-03 | Draft visibility dan Google identity checks | Member hanya published; Google subject/email verified diperiksa | Perubahan behavior perlu dipantau client | Selesai; tambah integration test bila tersedia | Pertahankan query/auth boundary |
| A-04/A-05/A-10 | Target points, streak attendance gate, live role validation | Implementasi sesuai policy yang dipilih user | Request lama points perlu target ID; attendance/streak timezone berbeda | Selesai; monitor rollout | Update client contract dan smoke test flows |
| A-06/A-07/A-08/A-19 | Query/affected-row defects | Fixed pada source | Query belum diuji lewat integration DB | Saat PostgreSQL test env tersedia | Tambah repository integration tests |
| A-09 | List/pagination bounds | News/contact maksimum 100 | Submission/redemption tetap unbounded | Berdasarkan volume/consumer | Tambah pagination dengan kontrak kompatibel |
| A-13/A-15/A-16/A-21 | Compose migration/profile, bootstrap, rollback, FK | Migration 21 applied; bootstrap mengandalkan chain | A-20 satu orphan membuat FK NOT VALID | Petakan orphan sebelum validate | Jalankan cleanup terotorisasi, validasi FK, dokumentasikan rollback |
| A-14 | CI/test coverage | CI menjalankan test; unit coverage bertambah | Business SQL/transaction belum integration-tested | Bertahap | Siapkan PostgreSQL integration test disposable |

## Impact vs Effort

| Kelompok | Finding | Alasan |
|---|---|---|
| High Impact / Low Effort | A-01, A-02, A-03, A-04, A-10 | Source mitigasi selesai; rotasi secret dan rollout client perlu tindak lanjut operasional. |
| High Impact / Medium Effort | A-05, A-11, A-20 | Eligibility dan auth fixed; storage menunggu keputusan; FK memerlukan orphan cleanup/validation. |
| High Impact / High Effort | Tidak ada yang terbukti memerlukan pekerjaan besar | Tidak ditemukan dasar untuk rewrite atau migrasi arsitektur besar. |
| Low Impact / Low Effort | A-07, A-08, A-16, A-17, A-19 | Perbaikan lokal pada query, validasi tanggal, dan migration. |
| Low Impact / Medium Effort | A-09, A-12, A-13, A-14, A-15, A-18, A-19 | Sebagian telah selesai; sisa mencakup pagination endpoint lain dan integration test. |
| Low Impact / High Effort | Tidak ada yang direkomendasikan | Skala data/arsitektur belum membuktikan investasi besar diperlukan. |

## Do Not Touch

- Pertahankan SQL langsung `database/sql` dan struktur domain saat ini; audit tidak menemukan kebutuhan ORM atau generic repository.
- Pertahankan transaksi dan row lock pada saldo/ledger/reward; pola tersebut melindungi konsistensi.
- Pertahankan constraint unik attendance, submission, dan pending redemption; jangan mengubahnya tanpa keputusan business rule.
- Pertahankan JWT state validation, bcrypt, parameterized query, dan rate limiter yang sudah ada sambil memperbaiki isu spesifik.
- Jangan menyamakan seluruh Service/Repository atau response lama melalui refactor massal; response adalah kontrak client.
- Jangan mengganti mekanisme penyimpanan atau menambah cache/queue sebelum deployment dan beban aktual diketahui.

## Hal yang Perlu Dikonfirmasi

1. Apakah `site_id` merupakan batas tenant/security untuk routes selain points, dan apakah admin dibatasi ke site sendiri?
2. Apakah credential lama di `.env.local`/`.env.docker` pernah digunakan di shared, staging, atau production sehingga harus dirotasi/riwayat dibersihkan?
3. Apakah upload harus bertahan lintas deployment, dan apakah filesystem production memerlukan volume persisten?
4. Apakah rollback migration menjadi prosedur deployment resmi? Down migration attendance menghapus semua attendance records.
5. Kapan orphan `users.site_id` yang terdeteksi akan dipetakan/dibersihkan agar FK dapat divalidasi?
6. Berapa volume submission/redemption sebelum pagination ditambahkan tanpa mengubah consumer eksternal?

## Rencana Perbaikan Bertahap

### Tahap 1 - Perbaikan Risiko Tinggi

- A-02/A-03/A-06 sudah diimplementasikan; lakukan client/deployment smoke test sebelum rollout.
- A-01 sudah dikeluarkan dari index; rotasi credential bila pernah dipakai di lingkungan bersama.

### Tahap 2 - Perbaikan Reliability & Maintainability

- A-04/A-05/A-07/A-08/A-10 sudah diimplementasikan sesuai keputusan user.
- Tetapkan batas pagination endpoint berdasarkan kontrak (A-09) dan periksa `rows.Err()` (A-19).
- A-13/A-15/A-16 sudah diimplementasikan; migration 21 applied pada database disposable.
- Tambahkan test terarah pada alur SQL/auth dan jalankan `go test ./...` dalam CI (A-14).
- Pilih kebijakan penyimpanan upload dan error response sesuai deployment (A-11/A-12).

### Tahap 3 - Improvement

- Perjelas invalid date response untuk attendance (A-17).
- Dokumentasikan kontrak response/pagination baru tanpa breaking perubahan lama (A-18).
- A-20 masih memerlukan mapping orphan `users.site_id` dan `VALIDATE CONSTRAINT` setelah data owner menyetujui nilainya.

## Prinsip Audit

Prioritaskan correctness, security, data integrity, reliability, maintainability, lalu performance. Pertahankan implementasi sederhana selama tidak ada dampak yang dibuktikan; labeli keputusan produk/deployment yang belum diketahui sebagai **Perlu Konfirmasi** dan hindari perubahan besar hanya untuk mengikuti pattern populer.

## Status Setelah Implementasi (2026-10-01)

| ID | Status | Perubahan/verifikasi | Catatan tersisa |
|---|---|---|---|
| A-01 | Partially Resolved | `.env.local` dan `.env.docker` dikeluarkan dari Git index; template tanpa secret, ignore rules, helper, dan dokumentasi diperbarui. | Nilai lama masih ada dalam Git history; rotasi/clean-up history butuh keputusan operasional. |
| A-02 | Resolved | Member list/detail memaksa status `published`; admin mendapat list/detail tersendiri; submission draft ditolak. | Policy unit-tested; PostgreSQL query belum integration-tested. |
| A-03 | Resolved | Login memeriksa Google `sub` tersimpan dan `email_verified`; unit tests untuk match/mismatch dan claim ditambahkan. | Belum ada integration test terhadap Google provider. |
| A-04 | Resolved | Admin earn menerima target `user_id`; Postman example dan validasi request diperbarui. | Consumer eksternal harus mengirim `user_id`; site scope tidak diterapkan sesuai keputusan. |
| A-05 | Resolved | Streak mensyaratkan attendance row user pada UTC attendance date hari itu. | Tanggal streak tetap Asia/Jakarta; aturan dipilih sesuai attendance harian apa pun. |
| A-06 | Resolved | Attendance report join ke `contacts` untuk nama. | Belum diuji terhadap PostgreSQL terisolasi. |
| A-07 | Resolved | Lookup UUID memakai predicate langsung dan limit satu. | Belum ada test repository PostgreSQL. |
| A-08 | Resolved | News update/delete memeriksa affected rows dan menolak record soft-deleted sebagai not-found. | Belum ada test repository PostgreSQL. |
| A-09 | Partially Resolved | News/contact limit maksimum 100; default dan response dipertahankan. | Submission/redemption masih tanpa pagination karena kontrak client belum diketahui. |
| A-10 | Resolved | AdminOnly menggunakan role terkini yang dibaca AuthMiddleware dari database. | Live role change berlaku pada request berikutnya. |
| A-11 | Partially Resolved | Keempat upload handler memvalidasi ukuran, extension, dan detected MIME; validator diuji untuk JPEG/PNG/WebP dan mismatch. | Persistensi Docker/storage menunggu model deployment. |
| A-12 | Resolved | Error internal dicatat di log dengan request context; response 5xx tidak memuat detail internal. | Format error client dipertahankan. |
| A-13 | Resolved | Profile diaktifkan oleh semua helper; command shell diperbaiki dan migration 21 berhasil berjalan. | Database lokal sekarang berada di version 21. |
| A-14 | Partially Resolved | CI menambahkan `go test ./...`; unit tests baru dan existing lulus. | Integration tests untuk business SQL/transaction belum ada. |
| A-15 | Partially Resolved | Compose tidak lagi memuat snapshot; setup menjelaskan migrations sebagai jalur bootstrap. | `docs/schema.sql` tetap sebagai snapshot legacy; volume database existing tidak diuji/migrasi otomatis. |
| A-16 | Resolved | Down migration menjatuhkan `attendance_records`. | Rollback menghapus seluruh data attendance pada tabel tersebut. |
| A-17 | Resolved | Attendance mengembalikan 422 untuk tanggal invalid; parser memiliki unit test. | Tidak menambah validasi urutan start/end date yang tidak diminta finding. |
| A-18 | Won't Fix | Tidak menyeragamkan response lama demi menghindari breaking API changes. | Perbedaan existing dipertahankan; response baru sebaiknya mengikuti kontrak yang diputuskan. |
| A-19 | Resolved | News list mengembalikan `rows.Err()` setelah iterasi. | Belum ada fault-injection/integration test rows. |
| A-20 | Partially Resolved | Migration 21 menambah tujuh FK `NOT VALID` dengan `ON DELETE SET NULL`, melewati FK legacy yang cocok. | Satu orphan `users.site_id` dipertahankan; constraints menunggu cleanup sebelum divalidasi. |

Build final dan seluruh check CI dicatat pada [IMPLEMENTATION_REPORT.md](IMPLEMENTATION_REPORT.md).
