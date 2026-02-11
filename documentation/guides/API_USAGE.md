# Panduan Penggunaan API & Manajemen Akses

Dokumen ini menjelaskan alur kerja utama (workflow) dalam menggunakan API Casbin Project, mulai dari pendaftaran pengguna hingga manajemen hak akses berbasis peran (RBAC).

## Daftar Isi

1.  [Manajemen Pengguna (User Management)](#1-manajemen-pengguna-user-management)
    - [Registrasi Pengguna Baru](#11-registrasi-pengguna-baru)
    - [Login (Autentikasi)](#12-login-autentikasi)
    - [Melihat Profil Pengguna](#13-melihat-profil-pengguna)
2.  [Manajemen Peran (Role Management)](#2-manajemen-peran-role-management)
    - [Membuat Peran Baru](#21-membuat-peran-baru)
    - [Menetapkan Peran ke Pengguna (Assign Role)](#22-menetapkan-peran-ke-pengguna-assign-role)
3.  [Manajemen Organisasi (Organization Management)](#3-manajemen-organisasi-organization-management)
4.  [Manajemen Proyek (Project Management)](#4-manajemen-proyek-project-management)
5.  [Dashboard & Statistik (Stats)](#5-dashboard--statistik-stats)
6.  [Manajemen Izin Global (Global Permissions)](#6-manajemen-izin-global-global-permissions)
7.  [Resumable Upload (Tus)](#7-resumable-upload-tus)
8.  [Melihat Jejak Audit (Audit Logs)](#8-melihat-jejak-audit-audit-logs)

---

## 1. Manajemen Pengguna (User Management)

### 1.1. Registrasi Pengguna Baru

Setiap pengguna baru yang mendaftar akan secara otomatis diberikan peran **`role:user`**.

- **Endpoint:** `POST /api/v1/users/register`
- **Akses:** Publik (Tanpa Token)
- **Payload:**
  ```json
  {
    "username": "johndoe",
    "password": "password123",
    "name": "John Doe",
    "email": "johndoe@example.com"
  }
  ```
- **Response Sukses (201 Created):**
  Mengembalikan data pengguna yang baru dibuat beserta ID-nya. Simpan `id` ini untuk keperluan administrasi selanjutnya.

### 1.2. Login (Autentikasi)

Gunakan username dan password untuk mendapatkan sesi.

- **Endpoint:** `POST /api/v1/auth/login`
- **Akses:** Publik
- **Payload:**
  ```json
  {
    "username": "johndoe",
    "password": "password123"
  }
  ```
- **Response Sukses (200 OK):**
  ```json
  {
    "data": {
      "user": {
        "id": "uuid",
        "username": "johndoe",
        "role": "role:user"
      },
      "access_token": "eyJ...",
      "refresh_token": "eyJ..."
    }
  }
  ```
  > **Keamanan Baru:** Pada Frontend (Next.js), token ini **otomatis disimpan di HttpOnly Cookies** oleh Server Action. Anda tidak perlu menyertakan header `Authorization` secara manual jika memanggil API lewat proxy `/api/v1/...`.

### 1.3. Melihat Profil Pengguna Aktif

Gunakan endpoint ini untuk memverifikasi sesi dan mendapatkan data user terbaru.

- **Endpoint:** `GET /api/v1/auth/me`
- **Akses:** Terproteksi (Sesi Aktif)
- **Response:** Mengembalikan objek user yang sedang terotentikasi.

---

## 2. Manajemen Peran (Role Management)

Hanya pengguna dengan peran Admin (atau yang memiliki izin khusus) yang dapat mengelola peran.

### 2.1. Membuat Peran Baru

Jika peran standar (`role:user`, `role:admin`) belum cukup, Anda bisa membuat peran baru.

- **Endpoint:** `POST /api/v1/roles`
- **Akses:** Admin
- **Payload:**
  ```json
  {
    "name": "role:editor",
    "description": "Editor konten dengan akses tulis terbatas"
  }
  ```

### 2.2. Menetapkan Peran ke Pengguna (Assign Role)

Untuk menjadikan seorang pengguna sebagai Admin atau peran lainnya.

- **Endpoint:** `POST /api/v1/permissions/assign-role`
- **Akses:** Admin
- **Payload:**
  ```json
  {
    "user_id": "uuid-user-yang-disimpan-tadi",
    "role": "role:admin"
  }
  ```

### 2.2.1. Memberikan Peran Organisasi (Organization Role)

Peran dalam organisasi diatur terpisah dari peran global.

- **Owner**: Pemilik organisasi, akses penuh.
- **Admin**: Mengelola anggota dan pengaturan.
- **Member**: Akses standar ke resource.
- **Viewer**: Akses baca saja.

---

## 3. Manajemen Organisasi (Organization Management)

Organisasi adalah unit isolasi data utama (Multi-tenancy).

### 3.1. Membuat Organisasi

- **Endpoint:** `POST /api/v1/organizations`
- **Payload:**
  ```json
  { "name": "Acme Corp", "slug": "acme-corp" }
  ```
- **Catatan:** Pengguna yang membuat otomatis menjadi **Owner**.

### 3.2. Mengundang Anggota

Hanya Owner/Admin organisasi yang bisa mengundang.

- **Endpoint:** `POST /api/v1/organizations/:org_id/members/invite`
- **Header:** `X-Org-ID: <org_uuid>` (otomatis dihandle frontend via URL context)
- **Payload:**
  ```json
  { "email": "colleague@acme.com", "role": "member" }
  ```

### 3.3. Menerima Undangan

- **Endpoint:** `POST /api/v1/organizations/invitations/accept`
- **Payload:**
  ```json
  { "token": "jwt_token_from_email_link" }
  ```

---

## 4. Manajemen Proyek (Project Management)

Proyek adalah resource yang terikat pada organisasi.

### 4.1. Membuat Proyek

- **Endpoint:** `POST /api/v1/projects`
- **Payload:**
  ```json
  { "name": "Project Alpha", "description": "Top Secret" }
  ```
- **Konteks:** Organisasi harus ditentukan (biasanya via context session di backend atau header jika API langsung).

### 4.2. Melihat Daftar Proyek

- **Endpoint:** `GET /api/v1/projects`
- **Response:** Mengembalikan daftar proyek milik organisasi aktif pengguna.

---

## 5. Dashboard & Statistik (Stats)

Endpoint untuk data visualisasi dashboard.

### 5.1. Ringkasan (Summary)

- **Endpoint:** `GET /api/v1/stats/summary`
- **Output:** Total User, Total Org, Active Sessions.

### 5.2. Aktivitas (Activity)

- **Endpoint:** `GET /api/v1/stats/activity`
- **Output:** Grafik login/registrasi 7 hari terakhir.

---

## 6. Manajemen Izin Global (Global Permissions)

Sistem ini memisahkan definisi endpoint fisik dari hak akses logis untuk fleksibilitas yang lebih baik.

### Konsep Dasar

1.  **Endpoint**: URL API fisik dan metode HTTP-nya (misal: `GET /api/v1/reports`).
2.  **Access Right**: Nama logis untuk sekumpulan endpoint (misal: `view_reports`).
3.  **Permission (Policy)**: Aturan yang menghubungkan **Role** dengan **Resource (URL)** dan **Action (Method)**.

### Langkah 1: Daftarkan Endpoint

Misalnya Anda membuat fitur baru untuk melihat laporan penjualan.

- **Endpoint:** `POST /api/v1/endpoints`
- **Payload:**
  ```json
  {
    "path": "/api/v1/sales/reports",
    "method": "GET"
  }
  ```
  _Simpan `id` endpoint yang dihasilkan (misal: 10)._

### Langkah 2: Buat Access Right

Buat representasi logis dari hak akses tersebut.

- **Endpoint:** `POST /api/v1/access-rights`
- **Payload:**
  ```json
  {
    "name": "sales:view_reports",
    "description": "Izinkan melihat laporan penjualan"
  }
  ```
  _Simpan `id` access right yang dihasilkan (misal: 5)._

### Langkah 3: Hubungkan Endpoint ke Access Right

Hubungkan endpoint fisik ke hak akses logis. Satu Access Right bisa memiliki banyak Endpoint.

- **Endpoint:** `POST /api/v1/access-rights/link`
- **Payload:**
  ```json
  {
    "access_right_id": "uuid-access-right",
    "endpoint_id": "uuid-endpoint"
  }
  ```

### Langkah 3.1: Hapus Hubungan Endpoint (Unlink)

Jika Anda ingin menghapus endpoint dari grup hak akses tertentu.

- **Endpoint:** `POST /api/v1/access-rights/unlink`
- **Payload:** (Sama dengan payload link)
  ```json
  {
    "access_right_id": "uuid-access-right",
    "endpoint_id": "uuid-endpoint"
  }
  ```

### Langkah 4: Berikan Izin ke Peran (Grant Permission)

Ini adalah langkah yang mengaktifkan akses di level **Casbin**. Tanpa langkah ini, Role tidak bisa mengakses Endpoint meskipun Access Right sudah dibuat.

- **Endpoint:** `POST /api/v1/permissions/grant`
- **Payload:**
  ```json
  {
    "role": "role:editor",
    "path": "/api/v1/sales/reports",
    "method": "GET"
  }
  ```

Now, semua pengguna yang memiliki peran `role:editor` dapat mengakses endpoint `GET /api/v1/sales/reports`.

---

## 7. Resumable Upload (Tus)

Layanan upload terpusat yang mendukung fitur _pause_ dan _resume_. Digunakan untuk file besar atau koneksi tidak stabil.

- **Endpoint Utama:** `POST /api/v1/upload/files/`
- **Akses:** Terproteksi (Sesi Aktif)
- **Alur Kerja:**
  1.  **Inisialisasi**: Kirim `POST` dengan header `Upload-Length` dan `Upload-Metadata` (berisi target/tipe upload).
  2.  **Upload**: Patch data binary ke URL yang diberikan di header `Location`.
  3.  **Hook**: Setelah selesai, sistem akan memproses file sesuai tipenya (misal: update avatar).

> **Detail Teknis:** Lihat panduan lengkap di [Panduan Resumable Upload (Tus)](RESUMABLE_UPLOAD.md).

---

## 8. Melihat Jejak Audit (Audit Logs)

Sistem secara otomatis mencatat aktivitas penting seperti Login, Register, dan perubahan User.

- **Endpoint:** `POST /api/v1/audit-logs/search`
- **Akses:** Superadmin
- **Contoh Filter (Mencari aksi LOGIN):**
  ```json
  {
    "filter": {
      "action": { "type": "equals", "from": "LOGIN" }
    },
    "sort": [{ "colId": "created_at", "sort": "desc" }]
  }
  ```
