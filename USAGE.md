# Panduan Penggunaan API & Manajemen Akses

Dokumen ini menjelaskan alur kerja utama (workflow) dalam menggunakan API Casbin Project, mulai dari pendaftaran pengguna hingga manajemen hak akses berbasis peran (RBAC).

## Daftar Isi

1.  [Manajemen Pengguna (User Management)](#1-manajemen-pengguna-user-management)
    *   [Registrasi Pengguna Baru](#11-registrasi-pengguna-baru)
    *   [Login (Autentikasi)](#12-login-autentikasi)
    *   [Melihat Profil Pengguna](#13-melihat-profil-pengguna)
2.  [Manajemen Peran (Role Management)](#2-manajemen-peran-role-management)
    *   [Membuat Peran Baru](#21-membuat-peran-baru)
    *   [Menetapkan Peran ke Pengguna (Assign Role)](#22-menetapkan-peran-ke-pengguna-assign-role)
3.  [Manajemen Izin & Akses (Permission & Access Management)](#3-manajemen-izin--akses-permission--access-management)
    *   [Konsep Dasar](#konsep-dasar)
    *   [Langkah 1: Daftarkan Endpoint](#langkah-1-daftarkan-endpoint)
    *   [Langkah 2: Buat Access Right](#langkah-2-buat-access-right)
    *   [Langkah 3: Hubungkan Endpoint ke Access Right](#langkah-3-hubungkan-endpoint-ke-access-right)
    *   [Langkah 4: Berikan Izin ke Peran (Grant Permission)](#langkah-4-berikan-izin-ke-peran-grant-permission)

---

## 1. Manajemen Pengguna (User Management)

### 1.1. Registrasi Pengguna Baru

Setiap pengguna baru yang mendaftar akan secara otomatis diberikan peran **`role:user`**.

*   **Endpoint:** `POST /api/v1/users/register`
*   **Akses:** Publik (Tanpa Token)
*   **Payload:**
    ```json
    {
      "username": "johndoe",
      "password": "password123",
      "name": "John Doe",
      "email": "johndoe@example.com"
    }
    ```
*   **Response Sukses (201 Created):**
    Mengembalikan data pengguna yang baru dibuat beserta ID-nya. Simpan `id` ini untuk keperluan administrasi selanjutnya.

### 1.2. Login (Autentikasi)

Gunakan username dan password untuk mendapatkan **Access Token** (JWT) dan Refresh Token (via Cookie).

*   **Endpoint:** `POST /api/v1/auth/login`
*   **Akses:** Publik
*   **Payload:**
    ```json
    {
      "username": "johndoe",
      "password": "password123"
    }
    ```
*   **Response Sukses (200 OK):**
    ```json
    {
      "data": {
        "access_token": "eyJhbGciOiJIUzI1NiIs...",
        "token_type": "Bearer",
        "expires_in": 900
      }
    }
    ```
    > **Penting:** Gunakan `access_token` ini di header `Authorization: Bearer <token>` untuk setiap request ke endpoint terproteksi.

### 1.3. Melihat Profil Pengguna

*   **Endpoint:** `GET /api/v1/users/me`
*   **Akses:** Terproteksi (Perlu Token)
*   **Header:** `Authorization: Bearer <access_token>`

---

## 2. Manajemen Peran (Role Management)

Hanya pengguna dengan peran Admin (atau yang memiliki izin khusus) yang dapat mengelola peran.

### 2.1. Membuat Peran Baru

Jika peran standar (`role:user`, `role:admin`) belum cukup, Anda bisa membuat peran baru.

*   **Endpoint:** `POST /api/v1/roles`
*   **Akses:** Admin
*   **Payload:**
    ```json
    {
      "name": "role:editor",
      "description": "Editor konten dengan akses tulis terbatas"
    }
    ```

### 2.2. Menetapkan Peran ke Pengguna (Assign Role)

Untuk menjadikan seorang pengguna sebagai Admin atau peran lainnya.

*   **Endpoint:** `POST /api/v1/permissions/assign-role`
*   **Akses:** Admin
*   **Payload:**
    ```json
    {
      "user_id": "uuid-user-yang-disimpan-tadi",
      "role": "role:admin"
    }
    ```

---

## 3. Manajemen Izin & Akses (Permission & Access Management)

Sistem ini memisahkan definisi endpoint fisik dari hak akses logis untuk fleksibilitas yang lebih baik.

### Konsep Dasar

1.  **Endpoint**: URL API fisik dan metode HTTP-nya (misal: `GET /api/v1/reports`).
2.  **Access Right**: Nama logis untuk sekumpulan endpoint (misal: `view_reports`).
3.  **Permission (Policy)**: Aturan yang menghubungkan **Role** dengan **Resource (URL)** dan **Action (Method)**.

### Langkah 1: Daftarkan Endpoint

Misalnya Anda membuat fitur baru untuk melihat laporan penjualan.

*   **Endpoint:** `POST /api/v1/endpoints`
*   **Payload:**
    ```json
    {
      "path": "/api/v1/sales/reports",
      "method": "GET"
    }
    ```
    *Simpan `id` endpoint yang dihasilkan (misal: 10).*

### Langkah 2: Buat Access Right

Buat representasi logis dari hak akses tersebut.

*   **Endpoint:** `POST /api/v1/access-rights`
*   **Payload:**
    ```json
    {
      "name": "sales:view_reports",
      "description": "Izinkan melihat laporan penjualan"
    }
    ```
    *Simpan `id` access right yang dihasilkan (misal: 5).*

### Langkah 3: Hubungkan Endpoint ke Access Right

Hubungkan endpoint fisik ke hak akses logis. Satu Access Right bisa memiliki banyak Endpoint.

*   **Endpoint:** `POST /api/v1/access-rights/link`
*   **Payload:**
    ```json
    {
      "access_right_id": 5,
      "endpoint_id": 10
    }
    ```

### Langkah 4: Berikan Izin ke Peran (Grant Permission)

Ini adalah langkah yang mengaktifkan akses di level **Casbin**. Tanpa langkah ini, Role tidak bisa mengakses Endpoint meskipun Access Right sudah dibuat.

*   **Endpoint:** `POST /api/v1/permissions/grant`
*   **Payload:**
    ```json
    {
      "role": "role:editor",
      "path": "/api/v1/sales/reports",
      "method": "GET"
    }
    ```

Sekarang, semua pengguna yang memiliki peran `role:editor` dapat mengakses endpoint `GET /api/v1/sales/reports`.
