# Panduan Penggunaan API & Manajemen Akses

Dokumen ini menjelaskan alur kerja utama (workflow) dalam menggunakan API Casbin Project, mulai dari pendaftaran pengguna hingga manajemen hak akses berbasis peran (RBAC) dengan dukungan multi-tenancy.

## Daftar Isi

1.  [Manajemen Pengguna (User Management)](#1-manajemen-pengguna-user-management)
    - [Registrasi Pengguna Baru](#11-registrasi-pengguna-baru)
    - [Login (Autentikasi)](#12-login-autentikasi)
2.  [Multi-Tenancy Context (Headers)](#2-multi-tenancy-context-headers)
3.  [Manajemen Peran (Role Management)](#3-manajemen-peran-role-management)
    - [Menetapkan Peran ke Pengguna (Assign Role)](#31-menetapkan-peran-ke-pengguna-assign-role)
4.  [Manajemen Izin & Akses (Permission Management)](#4-manajemen-izin--akses-permission-management)
    - [Memberikan Izin (Grant Permission)](#41-memberikan-izin-grant-permission)
    - [Pengecekan Batch (Batch Permission Check)](#42-pengecekan-batch-batch-permission-check)
5.  [Melihat Jejak Audit (Audit Logs)](#5-melihat-jejak-audit-audit-logs)
6.  [Alur Berbasis Tautan Email](#6-alur-berbasis-tautan-email)
    - [Undangan Organisasi](#61-undangan-organisasi)
    - [Lupa & Reset Password](#62-lupa--reset-password)
    - [Verifikasi Email](#63-verifikasi-email)

---

## 1. Manajemen Pengguna (User Management)

### 1.1. Registrasi Pengguna Baru

Setiap pengguna baru yang mendaftar akan secara otomatis diberikan peran **`role:user`** di domain `global`.

- **Endpoint:** `POST /api/v1/users/register`
- **Payload:**
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

- **Endpoint:** `POST /api/v1/auth/login`
- **Response Sukses (200 OK):**
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

- `X-Organization-ID`: UUID organisasi.
- `X-Organization-Slug`: Slug unik organisasi (misal: `acme-corp`).

Jika header ini tidak disertakan, sistem akan menggunakan konteks `global`.

---

## 3. Manajemen Peran (Role Management)

### 3.1. Menetapkan Peran ke Pengguna (Assign Role)

Anda dapat menetapkan peran kepada pengguna dalam organisasi tertentu menggunakan field `domain`.

- **Endpoint:** `POST /api/v1/permissions/assign-role`
- **Payload:**
  ```json
  {
    "user_id": "uuid-user",
    "role": "role:admin",
    "domain": "acme-corp"
  }
  ```
  _Catatan: Jika `domain` kosong, akan otomatis menggunakan `"global"`._

---

## 4. Manajemen Izin & Akses (Permission Management)

### 4.1. Memberikan Izin (Grant Permission)

Menghubungkan **Role** dengan **Resource** dan **Action** di domain tertentu.

- **Endpoint:** `POST /api/v1/permissions/grant`
- **Payload:**
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

- **Endpoint:** `POST /api/v1/permissions/check-batch`
- **Payload:**
  ```json
  {
    "items": [
      { "resource": "/api/v1/users", "action": "GET", "domain": "global" },
      { "resource": "/api/v1/projects", "action": "POST", "domain": "acme-corp" },
      { "resource": "/api/v1/billing", "action": "READ", "domain": "finance-dept" }
    ]
  }
  ```
- **Response:**
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

- **Endpoint:** `POST /api/v1/audit-logs/search`
- **Contoh Filter (Mencari aksi LOGIN):**
  ```json
  {
    "filter": {
      "action": { "type": "equals", "from": "LOGIN" }
    }
  }
  ```

---

## 6. Alur Berbasis Tautan Email

Semua tautan yang dikirim backend dibangun dari `SERVER_FRONTEND_BASE_URL`.
Variabel ini **tidak punya nilai default**: jika kosong, tautan tergenerasi tanpa
domain. Lihat `documentation/guides/MAINTENANCE.md`.

Backend tidak pernah menebak domain frontend dan tidak mengambilnya dari header
`Origin`/`Referer`, karena header tersebut dapat dipalsukan dan akan membuka
celah pengambilalihan akun (tautan sah dengan domain penyerang).

### 6.1. Undangan Organisasi

Mengundang anggota akan mengirim email berisi tautan:

```text
{SERVER_FRONTEND_BASE_URL}/invite/{token}
```

Token berlaku 48 jam. Halaman tujuan meminta nama dan password bila penerima
adalah pengguna baru (shadow user berstatus `invited`), lalu memanggil:

```http
POST /api/v1/organizations/invitations/accept
Content-Type: application/json

{
  "token": "abc123...",
  "name": "Nama Lengkap",
  "password": "PasswordBaru123"
}
```

`name` dan `password` opsional untuk pengguna yang sudah aktif.

### 6.2. Lupa & Reset Password

Permintaan reset:

```http
POST /api/v1/auth/forgot-password
Content-Type: application/json

{ "email": "user@example.com" }
```

Respons **selalu sukses** meski email tidak terdaftar, untuk mencegah
enumerasi akun. Email berisi tautan (token berlaku 15 menit):

```text
{SERVER_FRONTEND_BASE_URL}/reset-password?token={token}
```

Halaman tersebut mengirim:

```http
POST /api/v1/auth/reset-password
Content-Type: application/json

{
  "token": "abc123...",
  "new_password": "PasswordBaru123"
}
```

Password wajib 8–72 karakter. Reset yang berhasil **mencabut seluruh sesi aktif**
pengguna tersebut, sehingga perangkat lain harus login ulang.

### 6.3. Verifikasi Email

Tautan verifikasi (token berlaku 24 jam):

```text
{SERVER_FRONTEND_BASE_URL}/verify-email?token={token}
```

Halaman memverifikasi otomatis saat dibuka:

```http
POST /api/v1/auth/verify-email
Content-Type: application/json

{ "token": "abc123..." }
```

Token bersifat sekali pakai. Untuk meminta tautan baru (perlu sesi aktif):

```http
POST /api/v1/auth/resend-verification
```
