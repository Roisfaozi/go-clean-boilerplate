# Dokumentasi Koleksi Postman: Casbin Project API

Dokumen ini menjelaskan struktur, endpoint, dan skenario pengujian yang tercakup dalam koleksi Postman untuk proyek **Casbin Project API**. Koleksi ini dirancang untuk menguji fungsionalitas *end-to-end* (E2E), termasuk manajemen pengguna, autentikasi, manajemen peran (RBAC), hak akses, dan endpoint API.

---

## 📂 Struktur Folder

Koleksi ini dibagi menjadi beberapa folder utama berdasarkan modul fungsionalitas:

1.  **Users**: Manajemen pengguna (pendaftaran, profil, dan admin actions).
2.  **Authentication**: Login, refresh token, dan logout.
3.  **Roles**: Manajemen peran pengguna (Role-Based Access Control).
4.  **Permissions (Casbin Policies)**: Manajemen kebijakan akses granular menggunakan Casbin.
5.  **Access Scenarios (Protected Routes)**: Skenario pengujian hak akses untuk berbagai peran (Admin, Editor, Viewer).
6.  **Endpoints**: Manajemen metadata endpoint API.
7.  **Access Rights**: Manajemen hak akses abstrak dan penghubungannya ke endpoint.
8.  **Happy Path Workflow**: Alur kerja lengkap pengguna dari pendaftaran hingga akses fitur.

---

## 📝 Detail Endpoint dan Skenario Tes

### 1. Users (Manajemen Pengguna)

Folder ini berisi operasi CRUD untuk pengguna, termasuk endpoint publik dan endpoint khusus admin.

| Request | Method | Endpoint | Auth | Deskripsi & Skenario Tes |
| :--- | :---: | :--- | :---: | :--- |
| **Register New User** | `POST` | `/users/register` | - | Mendaftarkan pengguna baru.<br>✅ **Tes Positif**: Memastikan status `201 Created`, respon JSON valid, dan presence of `userId`. Saves `userId`, `username`, and `password` to environment variables for subsequent tests. |
| **Register User with Existing Username** | `POST` | `/users/register` | - | Mencoba mendaftar dengan username yang sudah ada.<br>❌ **Tes Negatif**: Memastikan status `409 Conflict` atau konflik yang sesuai. |
| **Register User with Bad Payload** | `POST` | `/users/register` | - | Mengirim payload yang tidak lengkap atau salah.<br>❌ **Tes Negatif**: Memastikan status `400 Bad Request`. |
| **[SECURITY] POST User with Malformed JSON** | `POST` | `/users/register` | - | Mengirim JSON yang rusak (syntax error).<br>🔒 **Tes Keamanan**: Memastikan server menangani error dengan `400 Bad Request` tanpa *panic*. |
| **Get Current User** | `GET` | `/users/me` | `authToken` | Mengambil profil pengguna yang sedang login.<br>✅ **Tes Positif**: Memastikan status `200 OK` dan ID pengguna sesuai dengan token. |
| **Update Current User** | `PUT` | `/users/me` | `authToken` | Memperbarui profil pengguna yang sedang login.<br>✅ **Tes Positif**: Memastikan status `200 OK` dan data berhasil diperbarui. |
| **[Admin] Get All Users** | `GET` | `/users` | `adminToken` | Mengambil daftar semua pengguna (Khusus Admin).<br>✅ **Tes Positif**: Memastikan status `200 OK` dan respons berupa array. |
| **[Admin] Get All Users (Filtered & Paginated)** | `GET` | `/users?page=1&limit=5&username=User` | `adminToken` | Mengambil subset pengguna berdasarkan paginasi dan filter.<br>✅ **Tes Positif**: Memastikan status `200 OK`, respons berupa array, dan jumlah item sesuai `limit`. |
| **[Admin] Get User By ID** | `GET` | `/users/:id` | `adminToken` | Mengambil detail pengguna spesifik berdasarkan ID (Khusus Admin).<br>✅ **Tes Positif**: Memastikan status `200 OK` dan ID data sesuai parameter. |
| **[Admin] Delete User** | `DELETE` | `/users/:id` | `adminToken` | Menghapus pengguna spesifik (Khusus Admin).<br>✅ **Tes Positif**: Memastikan status `200 OK`. |

### 2. Authentication (Autentikasi)

| Request | Method | Endpoint | Auth | Deskripsi & Skenario Tes |
| :--- | :---: | :--- | :---: | :--- |
| **Login User** | `POST` | `/auth/login` | - | Login pengguna untuk mendapatkan token.<br>✅ **Tes Positif**: Memastikan status `200 OK` dan menyimpan `accessToken` ke variabel `authToken`. |
| **[Positive] Refresh with Valid Cookie** | `POST` | `/auth/refresh` | Cookie | Menggunakan refresh token dari cookie untuk mendapatkan access token baru.<br>✅ **Tes Positif**: Status `200 OK`, token baru diterima, dan cookie diperbarui. |
| **[Negative] Refresh with No Cookie** | `POST` | `/auth/refresh` | - | Mencoba refresh tanpa cookie.<br>❌ **Tes Negatif**: Memastikan status `401 Unauthorized`. |
| **[Negative] Refresh with Invalid Token** | `POST` | `/auth/refresh` | Cookie | Mencoba refresh dengan token palsu.<br>❌ **Tes Negatif**: Memastikan status `401 Unauthorized`. |
| **[Positive] Logout with Valid Token** | `POST` | `/auth/logout` | `authToken` | Logout pengguna.<br>✅ **Tes Positif**: Status `200 OK` dan cookie refresh token dihapus/dikosongkan. |
| **[Negative] Logout with No Token** | `POST` | `/auth/logout` | - | Logout tanpa token akses.<br>❌ **Tes Negatif**: Memastikan status `401 Unauthorized`. |

### 3. Roles (Peran)

| Request | Method | Endpoint | Auth | Deskripsi & Skenario Tes |
| :--- | :---: | :--- | :---: | :--- |
| **Create a New Role** | `POST` | `/roles` | `adminToken` | Membuat peran baru (misal: 'manager').<br>✅ **Tes Positif**: Status `201 Created`, memvalidasi nama role yang dibuat. |
| **List All Roles** | `GET` | `/roles` | `adminToken` | Mendapatkan semua peran yang tersedia.<br>✅ **Tes Positif**: Status `200 OK`, respons berupa array role. |

### 4. Permissions (Kebijakan Casbin)

Mengelola aturan akses granular (Siapa boleh melakukan Apa pada Apa).

| Request | Method | Endpoint | Auth | Deskripsi & Skenario Tes |
| :--- | :---: | :--- | :---: | :--- |
| **Add Policy (Grant Permission)** | `POST` | `/permissions/grant` | `adminToken` | Memberikan izin akses ke sebuah role.<br>✅ **Tes Positif**: Status `201 Created`. |
| **View Policies** | `GET` | `/permissions` | `adminToken` | Melihat semua kebijakan akses.<br>✅ **Tes Positif**: Status `200 OK`, respons array. |
| **Get Permissions for Role** | `GET` | `/permissions/:role` | `adminToken` | Melihat izin khusus untuk role tertentu.<br>✅ **Tes Positif**: Status `200 OK`. |
| **Assign Role to User** | `POST` | `/permissions/assign-role` | `adminToken` | Menetapkan role ke user tertentu.<br>✅ **Tes Positif**: Status `200 OK`. |
| **Update Permission** | `PUT` | `/permissions` | `adminToken` | Mengubah aturan kebijakan yang ada.<br>✅ **Tes Positif**: Status `200 OK`. |
| **Remove Policy (Revoke Permission)** | `DELETE` | `/permissions/revoke` | `adminToken` | Mencabut izin akses.<br>✅ **Tes Positif**: Status `200 OK`. |

### 5. Access Scenarios (Skenario Akses)

Bagian ini adalah **inti pengujian keamanan RBAC**. Folder ini mensimulasikan berbagai pengguna dengan peran berbeda mencoba mengakses sumber daya yang dilindungi.

*   **[SETUP]**: Rangkaian request untuk mendaftarkan user `Admin`, `Editor`, dan `Viewer`, login mereka, dan menetapkan role yang sesuai.
*   **[SCENARIO] Article Access**:
    *   **[VIEWER] GET Articles**: Memastikan Viewer **BISA** membaca (`200 OK`).
    *   **[VIEWER] POST Article**: Memastikan Viewer **TIDAK BISA** menulis (`403 Forbidden`).
    *   **[EDITOR] POST Article**: Memastikan Editor **BISA** menulis (`201 Created`).

### 6. Endpoints & Access Rights

Fitur manajemen metadata API dan hak akses abstrak.

| Request | Method | Endpoint | Auth | Deskripsi & Skenario Tes |
| :--- | :---: | :--- | :---: | :--- |
| **Create an Endpoint** | `POST` | `/endpoints` | `adminToken` | Mendaftarkan endpoint API baru ke sistem.<br>✅ **Tes Positif**: Status `201 Created`. |
| **Create an Access Right** | `POST` | `/access-rights` | `adminToken` | Membuat hak akses logis (misal: 'document:read').<br>✅ **Tes Positif**: Status `201 Created`. |
| **List All Access Rights** | `GET` | `/access-rights` | `adminToken` | Melihat semua hak akses.<br>✅ **Tes Positif**: Status `200 OK`. |
| **Link Endpoint to Access Right** | `POST` | `/access-rights/link` | `adminToken` | Menghubungkan endpoint fisik dengan hak akses logis.<br>✅ **Tes Positif**: Status `200 OK`. |

---

## ⚙️ Variabel Lingkungan (Environment Variables)

Koleksi ini menggunakan variabel berikut yang diatur secara otomatis oleh skrip tes (`pm.environment.set`):

*   `baseURL`: URL dasar API (misal: `http://localhost:8080`).
*   `authToken`: Token JWT pengguna biasa yang sedang dites.
*   `adminToken`: Token JWT khusus admin untuk endpoint terproteksi tinggi.
*   `userId`: ID pengguna yang baru dibuat.
*   `roleName`: Nama role yang baru dibuat.
*   `endpointId`, `accessRightId`: ID untuk entitas terkait.

---

**Catatan:** Koleksi ini dirancang untuk dijalankan secara berurutan (terutama folder *Happy Path* dan *Access Scenarios*) karena adanya ketergantungan data antar request (misal: login butuh user yang sudah diregister).