# Project Guide: Go Clean Boilerplate API

Selamat datang di panduan proyek **Go Clean Boilerplate API**! Dokumen ini dirancang khusus untuk membantu Anda, sebagai *junior developer*, memahami secara mendalam struktur, teknologi, dan implementasi dari proyek ini. Kami akan membahas **mengapa** setiap keputusan dibuat, bukan hanya **bagaimana** implementasinya.

---

## 1. Pendahuluan (Introduction)

### 1.1 Apa itu Go Clean Boilerplate?

**Go Clean Boilerplate** adalah sebuah proyek RESTful API berbasis Go (Golang) yang dibangun dengan menerapkan prinsip-prinsip **Clean Architecture** dan **modularity**. Tujuan utamanya adalah menyediakan fondasi yang kuat, aman, skalabel, dan mudah dipelihara untuk pengembangan backend. Proyek ini dilengkapi dengan fitur-fitur penting seperti otentikasi JWT, otorisasi Role-Based Access Control (RBAC) menggunakan Casbin, sistem pencarian dinamis, serta kemampuan komunikasi real-time melalui WebSocket dan Server-Sent Events (SSE).

### 1.2 Tujuan Proyek

*   **Pembelajaran:** Sebagai contoh implementasi Clean Architecture dan praktik terbaik Go untuk junior developer.
*   **Fondasi Cepat:** Menyediakan boilerplate yang siap pakai untuk proyek-proyek backend baru yang membutuhkan fitur-fitur dasar yang kuat.
*   **Skalabilitas & Pemeliharaan:** Memastikan kode mudah dipahami, diuji, dan diperluas di masa depan.
*   **Keamanan:** Mengintegrasikan solusi otentikasi (JWT) dan otorisasi (Casbin RBAC) yang robust.

### 1.3 Fitur Utama

*   **Clean Architecture:** Struktur kode modular dan terorganisir berdasarkan domain dan layer.
*   **Otorisasi RBAC Lanjutan (Casbin):** Kontrol akses yang sangat granular dan dinamis.
*   **Otentikasi Aman (JWT & Redis):** Token akses stateless, token refresh stateful dengan revokasi instan.
*   **Dynamic Search & Filtering:** Mekanisme pencarian dan pengurutan data yang fleksibel dan aman.
*   **Komunikasi Real-time (SSE & WebSocket):** Dukungan untuk event streaming satu arah dan komunikasi bidireksional.
*   **Validasi Robust:** Validasi input request terpusat dengan pesan error yang *user-friendly*.
*   **Standardisasi Respons:** Struktur respons JSON yang konsisten untuk sukses dan error.
*   **Migrasi Database:** Manajemen skema database berbasis versi.
*   **Dokumentasi API Otomatis:** Integrasi Swagger/OpenAPI.

---

## 2. Teknologi yang Digunakan (Technology Stack)

Pemilihan teknologi di proyek ini didasarkan pada performa, komunitas, kemudahan penggunaan, dan keselarasan dengan prinsip skalabilitas dan pemeliharaan.

### 2.1 Go (Golang)

*   **Mengapa Go?**
    *   **Performa:** Go adalah bahasa yang dikompilasi (compiled language) yang menawarkan performa tinggi, mirip dengan C++ atau Java, tetapi dengan sintaksis yang lebih sederhana.
    *   **Concurrency:** Desain bawaan Go dengan Goroutines dan Channels mempermudah penulisan aplikasi concurrent yang efisien dan skalabel (misalnya untuk WebSocket atau SSE).
    *   **Sintaksis Sederhana:** Mudah dipelajari dan dibaca, meminimalkan *cognitive load* dan mendorong konsistensi antar developer.
    *   **Deployment Mudah:** Hasil kompilasi berupa *single binary* yang memudahkan proses deployment tanpa banyak dependensi.
    *   **Struktur Proyek Jelas:** Konvensi Go mendorong struktur proyek yang rapi.

### 2.2 Gin Web Framework

*   **Mengapa Gin?**
    *   **Performa Tinggi:** Salah satu framework web Go tercepat, ideal untuk membangun RESTful API.
    *   **Middleware:** Mendukung middleware, yang sangat membantu untuk fungsionalitas seperti autentikasi, otorisasi Casbin, logging, dan CORS.
    *   **Routing:** Sistem routing yang kuat dan mudah digunakan.
    *   **API-centric:** Dirancang khusus untuk membangun API, bukan aplikasi web *full-stack* dengan template rendering.

### 2.3 GORM (Go Object Relational Mapping)

*   **Mengapa GORM?**
    *   **ORM:** Menyediakan abstraksi untuk berinteraksi dengan database relasional (misalnya MySQL). Mengurangi kebutuhan untuk menulis query SQL mentah yang berulang.
    *   **Fitur Lengkap:** Mendukung migrasi, relasi (satu-ke-satu, satu-ke-banyak, banyak-ke-banyak), *soft delete*, transaksi, dll.
    *   **Mudah Digunakan:** API yang intuitif untuk operasi CRUD.
    *   **Fleksibilitas:** Memungkinkan eksekusi SQL mentah ketika abstraksi ORM tidak memadai.

### 2.4 Casbin

*   **Mengapa Casbin?**
    *   **Otorisasi Robust:** Library otorisasi yang sangat kuat, mendukung berbagai model kontrol akses (RBAC, ABAC, ACL, dll.). Proyek ini menggunakan RBAC.
    *   **Kebijakan Dinamis:** Kebijakan otorisasi disimpan di database, memungkinkan pembaruan izin secara *runtime* tanpa perlu mengubah kode atau me-restart aplikasi.
    *   **Casbin Model:** Didefinisikan dalam file `.conf` (`internal/config/casbin_model.conf`), sangat fleksibel.
    *   **Adaptor:** Menggunakan `gorm-adapter` untuk menyimpan kebijakan di database.

### 2.5 Redis

*   **Mengapa Redis?**
    *   **In-Memory Data Store:** Sangat cepat karena menyimpan data di memori.
    *   **Sesi & Token Refresh:** Digunakan sebagai penyimpanan untuk token refresh (stateful) dan sesi pengguna, memungkinkan revokasi token secara instan (misalnya saat logout atau ban).
    *   **Pub/Sub (Potensi):** Dapat digunakan untuk sistem *message queuing* sederhana atau *event-driven architecture* di masa depan.

### 2.6 JWT (JSON Web Tokens)

*   **Mengapa JWT?**
    *   **Autentikasi Stateless:** Server tidak perlu menyimpan status sesi pengguna. Ini sangat skalabel untuk aplikasi terdistribusi.
    *   **Ringkas & Self-Contained:** Berisi semua informasi yang diperlukan untuk mengidentifikasi pengguna.
    *   **Keamanan:** Ditandatangani secara kriptografis untuk memastikan integritasnya.

### 2.7 Lainnya

*   **`go-playground/validator`:** Library validasi request yang kuat dan kaya fitur.
*   **`logrus`:** Library logging terstruktur yang fleksibel.
*   **`golang-migrate`:** Untuk mengelola migrasi skema database dengan versi.
*   **`air`:** Tool untuk *live-reloading* kode selama pengembangan, meningkatkan produktivitas.
*   **`swag`:** Tool untuk menggenerasi dokumentasi Swagger/OpenAPI dari anotasi kode Go.
*   **`stretchr/testify` & `vektra/mockery`:** Toolset lengkap untuk unit testing dan mocking.

---

## 3. Struktur dan Prinsip Arsitektur (Architecture and Principles)

Proyek ini sangat menganut prinsip **Clean Architecture**, yang berfokus pada pemisahan *concerns* (pemisahan tanggung jawab) dan independensi dari framework, UI, database, atau agen eksternal lainnya.

### 3.1 Clean Architecture Overview

Clean Architecture mengorganisir kode ke dalam lapisan-lapisan (layers) berbentuk konsentris, di mana dependensi hanya boleh mengalir ke arah dalam. Lapisan terluar bergantung pada lapisan di dalamnya, tetapi lapisan di dalam tidak boleh mengetahui detail lapisan di luarnya.

*   **Entities (Inti):** Berisi aturan bisnis paling umum dan tinggi-level. Ini adalah objek data inti aplikasi.
*   **Use Cases (Aturan Bisnis Aplikasi):** Berisi aturan bisnis spesifik untuk aplikasi Anda. Mengkoordinasikan aliran data ke dan dari Entities.
*   **Interface Adapters:** Mengadaptasi data dari format yang paling nyaman untuk Use Cases dan Entities ke format yang paling nyaman untuk agen eksternal (Database, Web, UI). Ini mencakup Controllers, Presenters, Gateways (Repository).
*   **Frameworks & Drivers (Terluar):** Database, Web Frameworks, UI, dll. Lapisan ini harus yang paling mudah diganti.

### 3.2 Struktur Folder Proyek

Struktur folder mencerminkan prinsip-prinsip Clean Architecture dan konvensi proyek Go standar.

```
.
├── cmd/                # Aplikasi utama/entry point.
│   └── api/            # Server API (main.go).
│
├── db/                 # Skrip terkait database.
│   ├── migrations/     # Migrasi skema database (.sql).
│   └── seeds/          # Skrip untuk mengisi data awal (seeding).
│
├── docs/               # Dokumentasi API yang digenerasi otomatis (Swagger).
│
├── documentation/      # Panduan proyek dan dokumentasi tambahan.
│   ├── API_ACCESS_WORKFLOW.md      # Detail Alur Akses API dan Hak Akses Role.
│   ├── DYNAMIC_SEARCH_EXAMPLES.md  # Contoh penggunaan Dynamic Search.
│   ├── GET_VS_DYNAMIC_SEARCH.md    # Perbedaan GET dan Dynamic Search POST.
│   └── SSE_USAGE.md                # Panduan penggunaan Server-Sent Events.
│
├── pkg/                # Package yang dapat digunakan kembali secara luas di dalam proyek.
│   ├── exception/      # Definisi error kustom.
│   ├── jwt/            # Utilitas untuk JSON Web Tokens.
│   ├── password/       # Utilitas hashing password.
│   ├── querybuilder/   # Implementasi Dynamic Search Query Builder.
│   ├── response/       # Struktur respons API standar.
│   ├── sse/            # Manajer Server-Sent Events.
│   ├── tx/             # Manajer transaksi database.
│   ├── validation/     # Validasi kustom dan format error.
│   └── ws/             # Manajer WebSocket dan penanganan client.
│
└── internal/           # Kode internal proyek yang tidak boleh diimpor oleh proyek Go eksternal.
    ├── config/         # Konfigurasi aplikasi, inisialisasi dependensi (DI Container).
    ├── middleware/     # Middleware HTTP (Autentikasi, Otorisasi, CORS).
    ├── mocking/        # Mock object untuk pengujian.
    ├── router/         # Konfigurasi routing utama.
    │
    └── modules/        # Modul-modul spesifik domain (Core Business Logic).
        ├── <nama_module_1>/ # Contoh: auth, user, role, permission, access.
        │   ├── delivery/   # (Interface Adapter) Handler HTTP (Controller).
        │   ├── usecase/    # (Use Case) Logika bisnis spesifik modul.
        │   ├── repository/ # (Interface Adapter) Abstraksi akses data.
        │   ├── model/      # (Entities/DTOs) Struktur data untuk request/response, entitas database.
        │   └── entity/     # (Entities) Representasi entitas domain/database.
        ├── <nama_module_2>/
        └── ...
```
**Penjelasan `pkg/` vs `internal/`:**
*   **`pkg/`**: Direktori ini berisi kode-kode yang dapat **digunakan kembali secara luas** di dalam proyek ini, dan **dapat diimpor oleh proyek Go eksternal lainnya** jika modul ini dijadikan library (meskipun saat ini belum). Ini adalah utilitas generik yang tidak terikat pada satu domain bisnis tertentu.
*   **`internal/`**: Sesuai dengan konvensi Go, kode di dalam direktori `internal/` **tidak dapat diimpor oleh proyek Go di luar modul ini**. Ini digunakan untuk fungsionalitas inti proyek yang bersifat pribadi dan tidak dimaksudkan untuk diekspos. Modul bisnis utama ditempatkan di sini untuk menjaga agar implementasi internal domain bisnis tetap tersembunyi dari dunia luar.

### 3.3 Prinsip SOLID dan Dependency Inversion (DIP)

Proyek ini sangat mengedepankan prinsip-prinsip pengembangan perangkat lunak, khususnya:

*   **SOLID Principles:**
    *   **Single Responsibility Principle (SRP):** Setiap modul/kelas/fungsi memiliki satu alasan untuk berubah.
    *   **Open/Closed Principle (OCP):** Entitas perangkat lunak harus terbuka untuk ekstensi, tetapi tertutup untuk modifikasi.
    *   **Liskov Substitution Principle (LSP):** Objek dalam program harus dapat diganti dengan *instance* dari *subtipe*-nya tanpa mengubah kebenaran program.
    *   **Interface Segregation Principle (ISP):** Banyak *interface* kecil lebih baik daripada satu *interface* besar.
    *   **Dependency Inversion Principle (DIP):** Modul tingkat tinggi tidak boleh bergantung pada modul tingkat rendah. Keduanya harus bergantung pada abstraksi. Abstraksi tidak boleh bergantung pada detail. Detail harus bergantung pada abstraksi.
*   **Dependency Inversion (DIP):** Ini adalah kunci dalam Clean Architecture. Layer Use Case (tingkat tinggi) tidak langsung memanggil implementasi Repository (tingkat rendah). Sebaliknya, Use Case bergantung pada **interface** Repository yang didefinisikan di layer Use Case. Implementasi Repository kemudian memenuhi interface ini. Ini memungkinkan penggantian implementasi database dengan mudah tanpa memengaruhi logika bisnis.
    *   **Contoh:** `user/usecase/interface.go` akan mendefinisikan `UserRepository` interface, dan `user/repository/user_repository.go` akan menjadi implementasinya.

---

## 4. Implementasi Fitur Utama (Core Feature Implementation)

Bagian ini akan menjelaskan bagaimana fitur-fitur utama proyek ini diimplementasikan dan diintegrasikan ke dalam arsitektur yang sudah ada.

### 4.1 Alur Autentikasi (Login, Refresh Token, Logout)

Autentikasi di proyek ini menggunakan kombinasi JWT (JSON Web Tokens) untuk akses stateless dan Refresh Token yang stateful (disimpan di Redis) untuk keamanan dan mekanisme revokasi.

*   **Login (`POST /auth/login`):**
    1.  Pengguna mengirimkan `username` dan `password`.
    2.  `AuthUseCase` memverifikasi kredensial pengguna dari `UserRepository`.
    3.  Jika valid, `JWTManager` (dari `pkg/jwt`) membuat sepasang Access Token (berumur pendek) dan Refresh Token (berumur panjang).
    4.  Refresh Token disimpan di Redis melalui `TokenRepository` dengan TTL (Time To Live) yang sesuai.
    5.  Access Token dikembalikan di body respons, Refresh Token dikirimkan sebagai `HttpOnly Cookie`.
    6.  **Broadcast Event:** `WebSocketManager` digunakan untuk melakukan broadcast event "UserLoggedIn" ke channel `global_notifications` (jika ada implementasinya). Ini menunjukkan kemampuan aplikasi untuk memberitahu sistem lain secara real-time.
*   **Refresh Token (`POST /auth/refresh`):**
    1.  Client mengirimkan request tanpa Access Token, tetapi Refresh Token harus ada di `HttpOnly Cookie`.
    2.  `AuthUseCase` mengambil Refresh Token dari cookie.
    3.  `JWTManager` memvalidasi Refresh Token.
    4.  `TokenRepository` memeriksa keberadaan Refresh Token di Redis (memastikan belum di-revoke).
    5.  Jika valid, `TokenRepository` menghapus Refresh Token lama dari Redis.
    6.  `JWTManager` membuat pasangan Access dan Refresh Token baru.
    7.  Refresh Token baru disimpan di Redis, dan Access Token baru dikembalikan.
*   **Logout (`POST /auth/logout`):**
    1.  Pengguna mengirimkan request dengan Access Token yang valid.
    2.  `AuthUseCase` mendapatkan `userID` dan `sessionID` dari klaim Access Token.
    3.  `TokenRepository` menghapus Refresh Token yang terkait dari Redis, secara efektif me-revoke sesi tersebut.
    4.  Cookie Refresh Token juga dihapus dari respons.

### 4.2 Otorisasi (Casbin RBAC)

Otorisasi adalah inti dari keamanan akses resource dan diimplementasikan menggunakan [Casbin](https://casbin.org/).

*   **Middleware Casbin (`internal/middleware/casbin_middleware.go`):**
    *   Setiap request yang membutuhkan otorisasi akan melewati `CasbinMiddleware`.
    *   Middleware ini mengambil `sub` (pengguna/peran), `obj` (resource yang diakses), dan `act` (aksi/method HTTP) dari request.
    *   **Resource (`obj`):** Untuk setiap request, middleware mencoba menemukan `Access Rights` yang terkait dengan `path` dan `method` request saat ini. Jika ditemukan `Access Right`, `name` dari `Access Right` tersebut akan digunakan sebagai `obj` dalam Casbin. Jika tidak ada `Access Right` yang relevan, `path` dari request akan digunakan secara langsung sebagai `obj`.
    *   **Aksi (`act`):** Metode HTTP (`GET`, `POST`, `PUT`, `DELETE`).
    *   **Subjek (`sub`):** Peran (`role`) pengguna yang diautentikasi (misalnya `role:admin`, `role:user`).
    *   Casbin kemudian melakukan pemeriksaan otorisasi: `enforcer.Enforce(sub, obj, act)`.
    *   Jika `enforce` gagal, request akan ditolak dengan status `403 Forbidden`.
*   **Model Casbin (`internal/config/casbin_model.conf`):**
    ```ini
    [request_definition]
    r = sub, obj, act

    [policy_definition]
    p = sub, obj, act

    [role_definition]
    g = _, _

    [policy_effect]
    e = some(where (p.eft == allow))

    [match_definition]
    m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act || r.obj == "/api/v1/users/me" && r.sub == r.obj.ID
    ```
    *   **`r` (request):** Mendefinisikan format request otorisasi (subjek, objek, aksi).
    *   **`p` (policy):** Mendefinisikan format kebijakan (subjek, objek, aksi).
    *   **`g` (role):** Mendefinisikan hubungan peran (`g(user, role)` berarti user memiliki peran).
    *   **`e` (effect):** Aturan bagaimana kebijakan dievaluasi.
    *   **`m` (match):** Fungsi pencocokan yang kompleks. Ini mencakup:
        *   `g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act`: Pemeriksaan RBAC standar (jika subjek memiliki peran yang sesuai, objek dan aksi harus cocok).
        *   `r.obj == "/api/v1/users/me" && r.sub == r.obj.ID`: Ini adalah aturan khusus untuk **akses profil sendiri**. Artinya, jika objek adalah endpoint `/api/v1/users/me`, maka subjek (UserID) harus sama dengan ID pengguna di path (r.obj.ID), memungkinkan pengguna mengakses resource `me` mereka sendiri tanpa memerlukan peran spesifik di Casbin, selama mereka terautentikasi dan ID mereka cocok. Ini adalah fitur Self-Service yang cerdas.
*   **Konsep Access Rights dan Endpoints:**
    *   `Access Rights` dan `Endpoints` adalah entitas database yang dikelola oleh `superadmin` untuk membentuk `obj` (resource) di kebijakan Casbin.
    *   Sebuah `Endpoint` adalah kombinasi spesifik `path` dan `method` (misalnya `/api/v1/users GET`).
    *   Sebuah `Access Right` adalah pengelompokan logis dari satu atau lebih `Endpoints` (misalnya `user_management` bisa mencakup `GET /api/v1/users`, `POST /api/v1/users`).
    *   Ketika `CasbinMiddleware` mengecek izin, ia akan mencari `Access Right` yang terkait dengan `path` dan `method` request saat ini. Jika ditemukan `Access Right`, `name` dari `Access Right` tersebut akan digunakan sebagai `obj` dalam Casbin. Jika tidak ada `Access Right` yang relevan, `path` dari request akan digunakan secara langsung sebagai `obj`.
    *   Untuk detail lebih lanjut dan contoh pemetaan, lihat [`documentation/API_ACCESS_WORKFLOW.md`](API_ACCESS_WORKFLOW.md).

### 4.3 Dynamic Search (Query Builder)

Fitur ini menyediakan mekanisme pencarian dan pengurutan yang sangat fleksibel untuk resource melalui permintaan `POST /<resource>/search`.

*   **Cara Kerja (`pkg/querybuilder`):**
    *   Library `pkg/querybuilder` adalah inti dari fungsionalitas ini. Ia menerima struktur JSON `DynamicFilter` di body request.
    *   `DynamicFilter` berisi `Filter` (map dari nama field ke objek filter dengan `type` operator, `from` nilai, `to` nilai) dan `Sort` (array objek sort dengan `colId` dan `sort` direction).
    *   `querybuilder` menggunakan **refleksi** untuk menganalisis struct model Go dan memetakan field request yang masuk ke nama kolom database yang sesuai (menggunakan tag `gorm:"column:..."` atau mengkonversi ke `snake_case`).
    *   Ini membangun klausa `WHERE` dan `ORDER BY` SQL yang aman, menggunakan *parameterized queries* untuk mencegah SQL Injection.
    *   Secara otomatis menambahkan klausa *soft delete* (`deleted_by IS NULL` atau `deleted_at = 0`) jika model memiliki field `DeletedBy` atau `DeletedAt`.
*   **Integrasi (Repository, Usecase, Controller):**
    *   **Controller:** Menerima body JSON `DynamicFilter` di endpoint `POST /<resource>/search`. Memanggil `Usecase`.
    *   **Usecase:** Menerima `DynamicFilter`, memvalidasi (jika perlu), dan meneruskannya ke `Repository`.
    *   **Repository:** Menggunakan `querybuilder.GenerateDynamicQuery()` dan `querybuilder.GenerateDynamicSort()` untuk membangun query GORM sebelum mengeksekusi `Find()`.
*   **Contoh Penggunaan:** Lihat [`documentation/DYNAMIC_SEARCH_EXAMPLES.md`](DYNAMIC_SEARCH_EXAMPLES.md) untuk contoh `curl` yang mendetail.

### 4.4 Server-Sent Events (SSE)

SSE memungkinkan server untuk secara *push* mengirimkan event satu arah ke client web melalui koneksi HTTP standar yang tetap terbuka.

*   **Cara Kerja (`pkg/sse/manager.go`):**
    *   `sse.Manager` adalah manajer event yang menangani pendaftaran/pencabutan client, dan *broadcasting* event.
    *   Setiap client memiliki channel pribadi (`chan Event`) untuk menerima event.
    *   `Manager` memiliki Goroutine yang terus berjalan (`run()`) untuk mendengarkan client baru, client yang terputus, dan event yang akan di-broadcast.
    *   Ketika `Broadcast()` dipanggil, event dikirim ke channel semua client yang terhubung.
    *   Event diformat sesuai standar SSE: `event: <nama_event>
data: <json_data>

`.
*   **Integrasi (App Init, Router):**
    *   `sse.Manager` diinisialisasi di `internal/config/app.go` dan dilewatkan sebagai dependensi.
    *   Endpoint `/events` di `internal/router/router.go` menggunakan handler `sseManager.ServeHTTP()` untuk menerima koneksi SSE.
    *   Event dapat di-broadcast dari mana saja di aplikasi (misalnya dari `AuthUseCase` setelah login).
*   **Contoh Penggunaan:** Lihat [`documentation/SSE_USAGE.md`](SSE_USAGE.md) untuk detail dan contoh implementasi frontend.

### 4.5 WebSocket

WebSocket menyediakan saluran komunikasi *bidireksional* dan *full-duplex* di atas satu koneksi TCP.

*   **Cara Kerja (`pkg/ws/`):**
    *   **`ws_manager.go`:** `WebSocketManager` adalah jantung dari fungsionalitas ini. Ia mengelola daftar client yang terhubung, channel-channel tempat client dapat subscribe, dan mekanisme broadcast pesan ke channel tertentu.
    *   **`ws_client.go`:** Setiap client WebSocket direpresentasikan oleh struct `Client` yang memiliki Goroutine `ReadPump()` (untuk membaca pesan dari client) dan `WritePump()` (untuk mengirim pesan ke client).
    *   **`ws_controller.go`:** `WebSocketController` menangani proses *upgrade* koneksi HTTP menjadi koneksi WebSocket menggunakan `gorilla/websocket`.
*   **Integrasi (App Init, Router):**
    *   `WebSocketManager` diinisialisasi di `internal/config/app.go`.
    *   Endpoint `/ws` di `internal/router/router.go` menggunakan handler `WebSocketController.HandleWebSocket()` untuk menerima koneksi WebSocket.
*   **Perbedaan dengan SSE:**
    *   **SSE:** Satu arah (server ke client), berbasis HTTP, lebih sederhana. Ideal untuk notifikasi atau stream data yang tidak memerlukan respons client secara langsung.
    *   **WebSocket:** Dua arah (server ke client dan client ke server), full-duplex, protokol terpisah. Ideal untuk chat, game, atau aplikasi yang membutuhkan interaksi real-time yang intensif dari kedua belah pihak.

---

## 5. Memulai dan Mengembangkan (Getting Started & Development)

Bagian ini akan memandu Anda dalam menyiapkan lingkungan pengembangan, menjalankan aplikasi, dan memahami alur kerja dasar untuk menambah atau memodifikasi fitur.

### 5.1 Prasyarat

Sebelum memulai, pastikan sistem Anda memiliki perangkat lunak berikut terinstal:

1.  **Go:** Versi 1.21 atau lebih tinggi. Ini adalah bahasa pemrograman utama proyek.
2.  **Docker & Docker Compose:** Digunakan untuk menjalankan layanan infrastruktur seperti database (MySQL) dan Redis secara terisolasi dalam kontainer. Ini memastikan lingkungan pengembangan yang konsisten.
3.  **Make:** Utilitas `make` digunakan untuk menjalankan perintah otomatisasi yang ditentukan dalam `Makefile` (misalnya, `make migrate-up`, `make test`).
4.  **Air (Opsional):** Tool untuk *live-reloading* kode saat ada perubahan file, sangat direkomendasikan untuk pengembangan.
    ```bash
    go install github.com/air-verse/air@latest
    ```
5.  **Swag CLI (Opsional):** Digunakan untuk menggenerasi ulang dokumentasi API Swagger dari anotasi kode.
    ```bash
    go install github.com/swaggo/swag/cmd/swag@latest
    ```
6.  **Golang Migrate (Opsional):** Jika Anda ingin menjalankan migrasi secara manual tanpa `Makefile`.
    ```bash
    go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    ```
7.  **C/C++ Compiler (GCC/MinGW-w64):** Diperlukan jika Anda menjalankan test repository yang menggunakan driver SQLite (misalnya `gorm.io/driver/sqlite`). Go menggunakan `CGO_ENABLED=1` untuk mengkompilasi dengan pustaka C. Pastikan `gcc` ada di PATH sistem Anda.

### 5.2 Setup Lingkungan

Ikuti langkah-langkah ini untuk menyiapkan proyek di lingkungan lokal Anda.

1.  **Clone Repositori:**
    ```bash
    git clone https://github.com/yourusername/go-clean-boilerplate.git
    cd go-clean-boilerplate # Pastikan nama folder sudah sesuai
    ```
    *Catatan: Pastikan nama folder proyek Anda di lokal sama persis dengan nama modul di `go.mod` (`go-clean-boilerplate`) untuk menghindari masalah Go Modules.*

2.  **Konfigurasi Environment:**
    Buat file `.env` dari `.env.example`. File ini berisi konfigurasi penting untuk database, Redis, JWT secrets, dan lainnya.
    ```bash
    cp .env.example .env
    ```
    *Tip: Nilai-nilai default dalam `.env.example` biasanya berfungsi langsung dengan `docker-compose.yml` yang disediakan.*

3.  **Mulai Infrastruktur (Database & Redis):**
    Gunakan Docker Compose untuk menjalankan kontainer MySQL dan Redis.
    ```bash
    docker-compose up -d
    ```
    Perintah ini akan membuat dan menjalankan kontainer di latar belakang. Anda dapat memverifikasi statusnya dengan `docker-compose ps`.

### 5.3 Database Migrasi & Seeding

Setelah infrastruktur berjalan, terapkan skema database dan isi data awal.

1.  **Jalankan Migrasi Database:**
    ```bash
    make migrate-up
    ```
    Perintah ini akan menjalankan semua file migrasi SQL di `db/migrations/` dan membuat tabel yang diperlukan.

2.  **Isi Data Awal (Seeding):**
    ```bash
    make seed-up
    ```
    Perintah ini akan menjalankan skrip *seeding* di `db/seeds/` untuk mengisi data awal, seperti peran default (`admin`, `user`), atau pengguna awal.

### 5.4 Menjalankan Aplikasi

Anda dapat menjalankan aplikasi dalam mode pengembangan (dengan hot reload) atau mode standar.

*   **Mode Pengembangan (Direkomendasikan):**
    ```bash
    air
    ```
    Jika Anda telah menginstal `air` (lihat Prasyarat), ini akan memulai server dan secara otomatis me-restart aplikasi setiap kali Anda menyimpan perubahan pada file kode. Server akan berjalan di `http://localhost:8080` (atau port yang ditentukan di `.env`).

*   **Mode Standar:**
    ```bash
    go run cmd/api/main.go
    ```
    Ini akan mengkompilasi dan menjalankan aplikasi. Perubahan kode memerlukan penghentian dan restart manual.

*   **Build Produksi:**
    ```bash
    make build
    ./bin/api # Atau ./bin/api.exe di Windows
    ```
    Ini akan mengkompilasi aplikasi ke dalam sebuah *single binary* yang dapat dieksekusi di direktori `bin/`.

### 5.5 Workflow Pengembangan Umum

Untuk menambah atau memodifikasi fitur baru, ikuti alur kerja umum ini:

1.  **Pahami Persyaratan:** Mulai dengan memahami apa yang perlu dilakukan.
2.  **Definisikan Entitas/Model:** Jika ada data baru, definisikan di `modules/<nama_modul>/entity/` dan `modules/<nama_modul>/model/`.
3.  **Implementasi Repository:** Tambah method baru di `modules/<nama_modul>/repository/interface.go` dan implementasinya di `modules/<nama_modul>/repository/<nama_modul>_repository.go`.
4.  **Implementasi Use Case:** Tambah method baru di `modules/<nama_modul>/usecase/interface.go` dan implementasinya di `modules/<nama_modul>/usecase/<nama_modul>_usecase.go`. Di sinilah logika bisnis utama berada.
5.  **Implementasi Handler (Controller):** Tambah method baru di `modules/<nama_modul>/delivery/http/<nama_modul>_controller.go` untuk menangani request HTTP.
6.  **Daftarkan Route:** Daftarkan endpoint baru di `modules/<nama_modul>/delivery/http/<nama_modul>_routes.go` dan integrasikan ke `internal/router/router.go`.
7.  **Testing:** Selalu tulis test untuk kode Anda.
    *   **Unit Test:** Untuk `Repository` (dengan mock DB), `Usecase` (dengan mock Repository), dan `Controller` (dengan mock Usecase).
    *   **Integration Test:** Jika diperlukan, untuk menguji interaksi antar komponen.
8.  **Migrasi Database:** Jika ada perubahan skema database, buat migrasi baru (`make migrate create <nama_migrasi>`) dan terapkan.
9.  **Dokumentasi:** Perbarui dokumentasi yang relevan (misalnya, Swagger, panduan ini).

*Tip Debugging:*
*   Gunakan `logrus.Debugf()` atau `logrus.Errorf()` untuk melacak aliran eksekusi dan nilai variabel.
*   Gunakan IDE seperti VS Code dengan ekstensi Go untuk *breakpoint debugging*.

---

## 6. Pengujian (Testing Strategy)

Pengujian adalah bagian integral dari proses pengembangan untuk memastikan kode berfungsi dengan benar, mencegah regresi, dan memvalidasi implementasi fitur. Proyek ini mengadopsi beberapa strategi pengujian.

### 6.1 Unit Tests

*   **Tujuan:** Menguji unit kode terkecil (fungsi, method) secara terisolasi.
*   **Struktur:** Setiap modul (`auth`, `user`, `role`, `permission`, `access`) memiliki sub-folder `test/` di dalamnya. Di dalam folder `test/` ini, terdapat file-file test untuk `Controller`, `UseCase`, dan `Repository` (misalnya, `user_controller_test.go`, `user_usecase_test.go`, `user_repository_test.go`).
*   **Filosofi:**
    *   **Repository Tests:** Menguji interaksi dengan database *nyata* (biasanya menggunakan database in-memory seperti SQLite untuk kecepatan, atau kontainer Docker untuk database asli). Ini bisa dianggap sebagai *integration test* level rendah.
    *   **Usecase Tests:** Menguji logika bisnis inti. Dependensi seperti `Repository` atau `JWTManager` akan di-*mock* atau di-*stub*.
    *   **Controller Tests:** Menguji bagaimana handler HTTP memproses request dan menghasilkan respons. Dependensi seperti `UseCase` akan di-*mock*.
*   **Menjalankan Unit Tests:**
    ```bash
    make test
    ```
    Perintah ini akan menjalankan `go test -v ./...` yang akan mengeksekusi semua file test di seluruh proyek.

### 6.2 Mocking: Kapan dan Bagaimana Menggunakan Mockery

*   **Apa itu Mocking?** Mocking adalah teknik dalam unit testing di mana objek nyata diganti dengan objek simulasi (mock) yang meniru perilaku objek asli. Ini memungkinkan pengujian unit kode secara terisolasi dari dependensinya.
*   **Mengapa Menggunakan Mocking?**
    *   **Isolasi:** Memastikan bahwa unit yang sedang diuji hanya bergantung pada perilakunya sendiri, bukan pada kebenaran atau performa dependensi eksternal (misalnya, tidak perlu database sungguhan untuk menguji `UseCase`).
    *   **Kontrol:** Memungkinkan Anda mensimulasikan skenario error atau kondisi *edge case* dari dependensi yang sulit direproduksi di lingkungan nyata.
    *   **Kecepatan:** Test berjalan lebih cepat karena tidak ada interaksi jaringan atau disk yang sebenarnya.
*   **Implementasi di Proyek Ini:**
    *   Proyek ini menggunakan library [Mockery](https://vektra.github.io/mockery/) untuk menggenerasi objek mock secara otomatis dari interface Go.
    *   File konfigurasi `.mockery.yml` mendefinisikan interface mana yang harus di-mock dan di mana file mock yang dihasilkan harus disimpan (misalnya, di `internal/modules/<module>/test/mocks/`).
    *   **Contoh:** `user/usecase/user_usecase_test.go` akan mengimpor `internal/modules/user/test/mocks/mock_user_repository.go` dan menggunakan `mocks.MockUserRepository` untuk mensimulasikan perilaku `UserRepository` interface.
*   **Merekgenerasi Mocks:**
    Jika Anda memodifikasi definisi interface apa pun (misalnya, di `repository/interface.go` atau `usecase/interface.go`), Anda harus meregenerasi mock yang relevan:
    ```bash
    make mocks
    ```

### 6.3 End-to-End Tests (Postman)

*   **Tujuan:** Memverifikasi seluruh alur aplikasi, dari request HTTP hingga respons, termasuk interaksi antar semua komponen (router, controller, usecase, repository, database).
*   **Postman Collections:** Proyek ini menyediakan beberapa Postman Collections di folder `postman/`:
    *   `Casbin Project API.postman_collection.json`: Untuk alur dasar CRUD, Auth, dan RBAC.
    *   `Casbin Project API - Dynamic Search.postman_collection.json`: Berisi berbagai skenario pencarian dinamis (positif, negatif, edge, security).
    *   `Casbin Project API - Realtime.postman_collection.json`: Contoh untuk koneksi WebSocket dan SSE.
*   **Cara Menggunakan:**
    1.  **Impor:** Impor file `.json` ini ke aplikasi Postman Anda.
    2.  **Environment:** Setel variabel lingkungan Postman (misalnya `baseURL`, `apiVersion`, `authToken`) di "Casbin Project Env.postman_environment.json" agar sesuai dengan setup lokal Anda.
    3.  **Jalankan Runner:** Gunakan fitur "Collection Runner" di Postman untuk menjalankan seluruh koleksi. Banyak request dilengkapi dengan skrip test (`pm.test()`) yang secara otomatis memverifikasi respons API (status code, struktur data, dll.).

### 6.4 Troubleshooting Test Repository (CGO/SQLite)

*   **Masalah Umum:** Anda mungkin mengalami error seperti `Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub` saat menjalankan test repository (terutama jika Anda menggunakan `gorm.io/driver/sqlite` untuk testing).
*   **Penyebab:** Go secara default mengkompilasi dengan `CGO_ENABLED=0` di beberapa sistem atau lingkungan, yang mencegah penggunaan pustaka C seperti `sqlite`.
*   **Solusi:**
    1.  **Instal GCC/MinGW-w64:** Pastikan Anda memiliki compiler C/C++ yang terinstal di sistem Anda dan dapat diakses dari PATH (misalnya, `MinGW-w64` di Windows, `build-essential` di Linux, atau `Xcode Command Line Tools` di macOS).
    2.  **Aktifkan CGO:** Jalankan test dengan `CGO_ENABLED=1`:
        ```bash
        CGO_ENABLED=1 go test -v ./...
        ```
        Atau modifikasi `Makefile` target `test` untuk menyertakan `CGO_ENABLED=1`.

---

## 7. Dokumentasi API (API Documentation)

Dokumentasi API sangat penting untuk developer frontend dan tim lain yang berinteraksi dengan API ini.

### 7.1 Swagger UI

*   **Apa itu?** Swagger UI menyediakan antarmuka web interaktif untuk melihat dan menguji endpoint API Anda. Ini digenerasi secara otomatis dari anotasi kode Go Anda.
*   **Akses:** Setelah aplikasi berjalan, buka browser Anda ke:
    `http://localhost:8080/api/docs/index.html`
*   **Cara Menggenerasi Ulang:**
    Jika Anda menambahkan atau memodifikasi endpoint dan anotasi Swagger (`@Summary`, `@Param`, `@Success`, dll.), Anda perlu menggenerasi ulang file-file Swagger:
    ```bash
    swag init -g cmd/api/main.go
    ```
    *Pastikan Anda berada di direktori root proyek saat menjalankan perintah ini.*

### 7.2 Referensi Dokumentasi Tambahan

Selain panduan ini, beberapa dokumen Markdown lain menyediakan detail spesifik:

*   [`documentation/API_ACCESS_WORKFLOW.md`](API_ACCESS_WORKFLOW.md): Detail lengkap tentang semua route API, alur kerja akses, dan privilege untuk setiap peran (superadmin, admin, user).
*   [`documentation/DYNAMIC_SEARCH_EXAMPLES.md`](DYNAMIC_SEARCH_EXAMPLES.md): Contoh penggunaan `curl` yang mendetail untuk endpoint pencarian dinamis dengan berbagai jenis filter dan sorting.
*   [`documentation/GET_VS_DYNAMIC_SEARCH.md`](GET_VS_DYNAMIC_SEARCH.md): Penjelasan perbedaan antara pencarian sederhana dengan `GET` dan pencarian dinamis dengan `POST /search`.
*   [`documentation/SSE_USAGE.md`](SSE_USAGE.md): Panduan tentang bagaimana Server-Sent Events (SSE) diimplementasikan dan cara menggunakannya.

---

## 8. Konvensi Kode (Coding Conventions)

Mengikuti konvensi kode yang konsisten adalah kunci untuk menjaga keterbacaan, pemeliharaan, dan kolaborasi dalam proyek.

### 8.1 Penamaan

*   **Paket (Package):** Gunakan huruf kecil semua (misalnya `user`, `auth`, `pkg`).
*   **Variabel, Fungsi, Metode (Private):** Gunakan `camelCase` dan diawali huruf kecil (misalnya `getUserByID`, `jwtManager`).
*   **Variabel, Fungsi, Metode (Public/Exported):** Gunakan `PascalCase` dan diawali huruf kapital (misalnya `NewJWTManager`, `GenerateTokenPair`).
*   **Konstanta:** Gunakan `PascalCase` atau `ALL_CAPS` jika konstanta global (misalnya `TestAccessSecret`, `MaxMessageSize`).
*   **Interface:** Diawali dengan huruf `I` atau diakhiri dengan `er` (misalnya `UserRepository`, `Writer`).

### 8.2 Formatting

*   Gunakan `go fmt` secara teratur. IDE Anda harus dikonfigurasi untuk menjalankannya secara otomatis saat menyimpan file.
*   `goimports` juga sangat direkomendasikan untuk mengelola impor secara otomatis.

### 8.3 Error Handling

*   Go menggunakan nilai balik `error` secara eksplisit.
*   Selalu periksa error setelah memanggil fungsi yang mengembalikan error.
*   Gunakan `errors.Is()` untuk memeriksa jenis error tertentu.
*   Gunakan `fmt.Errorf("...: %w", err)` untuk *wrapping* error, memungkinkan pelacakan *stack trace* dan inspeksi error asli.
*   Definisikan error kustom di `pkg/exception/error.go` untuk error yang sering terjadi atau memiliki makna khusus.

---

## 9. Kontribusi dan Lisensi (Contribution & License)

### 9.1 Kontribusi

Kami sangat menyambut kontribusi! Jika Anda ingin berkontribusi, ikuti langkah-langkah umum berikut:

1.  Fork repositori proyek.
2.  Buat cabang fitur baru (`git checkout -b feature/nama-fitur-baru`).
3.  Implementasikan perubahan Anda. Pastikan untuk menulis test yang sesuai dan semua test lulus.
4.  Commit perubahan Anda (`git commit -m 'feat: Tambah fitur X'`).
5.  Push ke cabang Anda (`git push origin feature/nama-fitur-baru`).
6.  Buka Pull Request ke repositori utama.

### 9.2 Lisensi

Proyek ini dilisensikan di bawah Lisensi Apache 2.0. Lihat file `LICENSE` untuk detail lebih lanjut.

---
