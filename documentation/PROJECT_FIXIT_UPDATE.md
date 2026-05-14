# 📄 Technical Documentation: System Architecture & Optimization Update (NexusOS)

## 📌 Ringkasan Eksekutif
Pembaruan ini merupakan bagian dari **"Project Fixit"** yang bertujuan untuk memperkuat fondasi arsitektur NexusOS sebelum implementasi fitur Multi-tenancy yang kompleks. Fokus utama adalah pada **Inversion of Control (IoC)**, **pemrosesan asinkron**, dan **efisiensi database**.

---

## 🏗️ 1. Dekopling Arsitektur & Inversion of Control (IoC)
### Masalah
Sebelumnya, layer `UseCase` memiliki ketergantungan langsung (tight coupling) pada infrastruktur (`pkg/ws`, `pkg/sse`) dan library eksternal (`Casbin`). Hal ini melanggar prinsip *Clean Architecture* dan membuat unit testing sulit dilakukan karena memerlukan mock yang sangat kompleks dari pihak ketiga.

### Solusi
Kami memperkenalkan layer abstraksi (Interface) di antara logika bisnis dan implementasi teknis.

*   **`NotificationPublisher` Interface:** Menghilangkan ketergantungan `UseCase` pada protokol pengiriman tertentu.
*   **`AuthzManager` Interface:** Menghilangkan ketergantungan `UseCase` pada library otorisasi tertentu (Casbin).

#### Pola Penggunaan Baru:
UseCase sekarang memanggil method pada interface, dan implementasi nyata (adapter) disuntikkan saat inisialisasi aplikasi di `internal/config/app.go`.

---

## ⚡ 2. Audit Logging Asinkron (Performance Optimization)
### Masalah
Setiap aksi penting (Login, Register, Update) menulis Audit Log ke database secara **sinkron**. Hal ini menambah latensi sebesar 20-50ms pada setiap request di jalur kritis, karena aplikasi harus menunggu database selesai menulis log sebelum merespon user.

### Solusi (Background Processing)
Audit Log kini dipindahkan ke **Background Worker** menggunakan Redis dan library `Asynq`.

*   **Alur:** `UseCase` mengirim payload ke Redis (Queue) → `AuthUseCase` langsung merespon user → Worker memproses payload dan menulis ke DB MySQL secara terpisah.
*   **Keuntungan:** Latensi API berkurang, throughput meningkat, dan kegagalan penulisan log tidak akan menggagalkan transaksi utama user.

---

## 🛡️ 3. Type Safety & Anti-Primitive Obsession
### Masalah
Fungsi-fungsi utama seperti `GetTicket` atau `GenerateTokenPair` menerima banyak argumen string berurutan (misal: `userID, orgID, sessionID, role, username`). Hal ini sangat rawan terhadap bug "salah urutan parameter" karena compiler tidak bisa membedakan antar string tersebut.

### Solusi (Context Structs)
Parameter-parameter tersebut kini dibungkus dalam struct formal yang *type-safe*.

*   **`jwt.UserContext`**: Digunakan secara khusus untuk pembuatan token JWT.
*   **`model.UserSessionContext`**: Digunakan untuk pertukaran data session identitas user antar layer internal.

---

## 💾 4. Optimasi Database & GORM
### A. Composite Indexes (Soft Delete Optimization)
GORM menggunakan fitur *Soft Delete*, yang berarti setiap query secara implisit akan ditambahkan klausa `WHERE deleted_at = 0`. Tanpa index yang tepat, query pada tabel yang memiliki jutaan baris akan melambat secara eksponensial.

**Index Baru (Migration 000018):**
Kami menambahkan index gabungan (Composite Indexes) pada tabel-tabel utama:
*   `idx_user_org_deleted`: `(organization_id, deleted_at)`
*   `idx_user_status_deleted`: `(status, deleted_at)`
*   `idx_audit_user_deleted`: `(user_id, deleted_at)`
*   *Indeks serupa diterapkan pada tabel Role, AccessRight, dan Project.*

### B. Pagination SkipCount
Menambahkan opsi `SkipCount` pada `DynamicFilter`.
*   **Kegunaan:** Jika UI hanya membutuhkan data (misal: *infinite scroll* atau data yang sudah diketahui jumlahnya), kita bisa melewati query `COUNT(*)` yang berat. Hal ini sangat krusial untuk performa pada tabel `audit_logs` yang pertumbuhannya sangat cepat.

---

## ⚙️ 5. Konfigurasi & Environment
Nilai default untuk otorisasi sekarang diatur secara terpusat di `AppConfig` dan dapat di-override melalui file `.env`.

| Variable | Deskripsi | Default |
| :--- | :--- | :--- |
| `CASBIN_DEFAULT_ROLE` | Role otomatis untuk user baru | `role:user` |
| `CASBIN_DEFAULT_DOMAIN` | Domain default otorisasi global | `global` |

---

## 🧪 6. Panduan Pengembang (Developer Guide)
Untuk menjaga konsistensi arsitektur ini, harap ikuti aturan berikut:

1.  **Jangan Impor Infrastruktur ke UseCase:** Jangan melakukan impor `pkg/ws` atau `pkg/sse` di dalam file UseCase. Gunakan interface `NotificationPublisher`.
2.  **Audit Logging:** Gunakan `taskDistributor.DistributeTaskAuditLog` untuk aksi-aksi umum. Gunakan logging sinkron hanya jika integritas log tersebut krusial untuk transaksi finansial.
3.  **Update Mocks:** Jika Anda mengubah interface, segera jalankan `make mocks` untuk memperbarui file mock testing.
4.  **Verifikasi dengan Test:** Sebelum melakukan commit, pastikan menjalankan `make test`. Seluruh pengujian (Unit, Security, Integration) harus lulus dengan **Exit Code: 0**.

---

## 🛡️ 7. Testing Berbasis Waktu (Clock Provider)
### Masalah
Penggunaan `time.Now()` secara langsung di dalam kode menyulitkan pengujian token yang memiliki masa berlaku (expiry). Hal ini sering menyebabkan tes di-skip karena ketergantungan pada kecepatan eksekusi CPU.

### Solusi
Gunakan interface `util.Clock` yang disuntikkan (dependency injection) ke dalam repository atau usecase.

*   **`util.RealClock`**: Digunakan saat aplikasi berjalan normal (produksi).
*   **`util.MockClock`**: Digunakan saat unit testing untuk "membekukan" waktu di titik tertentu.

#### Cara Penggunaan:
```go
// Di dalam Repository/UseCase
expiration := session.ExpiresAt.Sub(r.clock.Now())

// Di dalam Unit Test
mockClock.SetTime(fixedTime)
// Assertion akan selalu akurat karena waktu tidak berubah
```

---

## 🏢 8. Refinement Multi-tenancy
### A. Global User murni
Tabel `users` tidak lagi dijadikan filter utama untuk `organization_id`. Pencarian member organisasi sekarang selalu menggunakan subquery ke tabel `organization_members` untuk mendukung satu user di banyak organisasi (model Global User).

### B. Isolasi Audit Log
`AuditRepository` sekarang secara ketat mengikuti `OrganizationScope`. Pengguna hanya dapat melihat audit log yang sesuai dengan konteks organisasi yang aktif di session mereka.

---
**Status Implementasi:** ✅ **All Tests Passing (Unit, Integration, E2E).**
**Update Terakhir:** 23 Februari 2026
