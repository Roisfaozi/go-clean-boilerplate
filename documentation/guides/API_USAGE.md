# Panduan Penggunaan API & Manajemen Akses

Dokumen ini menjelaskan alur kerja utama (workflow) dalam menggunakan API Casbin Project, mulai dari pendaftaran pengguna hingga manajemen hak akses berbasis peran (RBAC) dengan dukungan multi-tenancy.

## Daftar Isi

1.  [Manajemen Pengguna (User Management)](#1-manajemen-pengguna-user-management)
    *   [Registrasi Pengguna Baru](#11-registrasi-pengguna-baru)
    *   [Login (Autentikasi)](#12-login-autentikasi)
2.  [Multi-Tenancy Context (Headers)](#2-multi-tenancy-context-headers)
3.  [Manajemen Peran (Role Management)](#3-manajemen-peran-role-management)
    *   [Menetapkan Peran ke Pengguna (Assign Role)](#31-menetapkan-peran-ke-pengguna-assign-role)
4.  [Manajemen Izin & Akses (Permission Management)](#4-manajemen-izin--akses-permission-management)
    *   [Memberikan Izin (Grant Permission)](#41-memberikan-izin-grant-permission)
    *   [Pengecekan Batch (Batch Permission Check)](#42-pengecekan-batch-batch-permission-check)

---

## 1. Manajemen Pengguna (User Management)

### 1.1. Registrasi Pengguna Baru

Setiap pengguna baru yang mendaftar akan secara otomatis diberikan peran **`role:user`** di domain `global`.

*   **Endpoint:** `POST /api/v1/users/register`
*   **Payload:**
    ```json
    {
      "username": "johndoe",
      "password": "password123",
      "name": "John Doe",
      "email": "johndoe@example.com"
    }
    ```

### 1.2. Login (Autentikasi)

Gunakan username dan password untuk mendapatkan **Access Token** (JWT).

*   **Endpoint:** `POST /api/v1/auth/login`
*   **Response Sukses (200 OK):**
    ```json
    {
      "data": {
        "access_token": "eyJhbGciOiJIUzI1NiIs...",
        "token_type": "Bearer"
      }
    }
    ```

---

## 2. Multi-Tenancy Context (Headers)

Banyak endpoint dalam aplikasi ini bersifat **Organization-Aware**. Untuk mengakses data dalam konteks organisasi tertentu, Anda wajib mengirimkan salah satu header berikut:

-   `X-Organization-ID`: UUID organisasi.
-   `X-Organization-Slug`: Slug unik organisasi (misal: `acme-corp`).

Jika header ini tidak disertakan, sistem akan menggunakan konteks `global`.

---

## 3. Manajemen Peran (Role Management)

### 3.1. Menetapkan Peran ke Pengguna (Assign Role)

Anda dapat menetapkan peran kepada pengguna dalam organisasi tertentu menggunakan field `domain`.

*   **Endpoint:** `POST /api/v1/permissions/assign-role`
*   **Payload:**
    ```json
    {
      "user_id": "uuid-user",
      "role": "role:admin",
      "domain": "acme-corp" 
    }
    ```
    *Catatan: Jika `domain` kosong, akan otomatis menggunakan `"global"`.*

---

## 4. Manajemen Izin & Akses (Permission Management)

### 4.1. Memberikan Izin (Grant Permission)

Menghubungkan **Role** dengan **Resource** dan **Action** di domain tertentu.

*   **Endpoint:** `POST /api/v1/permissions/grant`
*   **Payload:**
    ```json
    {
      "role": "role:editor",
      "path": "/api/v1/projects",
      "method": "POST",
      "domain": "acme-corp"
    }
    ```

### 4.2. Pengecekan Batch (Batch Permission Check)

Sangat berguna untuk Frontend (misal: menentukan tombol mana yang harus muncul). Mendukung pengecekan lintas domain dalam satu request.

*   **Endpoint:** `POST /api/v1/permissions/check-batch`
*   **Payload:**
    ```json
    {
      "items": [
        { "resource": "/api/v1/users", "action": "GET", "domain": "global" },
        { "resource": "/api/v1/projects", "action": "POST", "domain": "acme-corp" },
        { "resource": "/api/v1/billing", "action": "READ", "domain": "finance-dept" }
      ]
    }
    ```
*   **Response:**
    ```json
    {
      "data": {
        "results": {
          "/api/v1/users:GET": true,
          "/api/v1/projects:POST": true,
          "/api/v1/billing:READ": false
        }
      }
    }
    ```

---

## 5. Melihat Jejak Audit (Audit Logs)

Sistem secara otomatis mencatat aktivitas penting. Pencatatan audit sekarang mencakup `organization_id` untuk kepatuhan multi-tenancy.

*   **Endpoint:** `POST /api/v1/audit-logs/search`
*   **Contoh Filter (Mencari aksi LOGIN):**
    ```json
    {
      "filter": {
        "action": { "type": "equals", "from": "LOGIN" }
      }
    }
    ```
