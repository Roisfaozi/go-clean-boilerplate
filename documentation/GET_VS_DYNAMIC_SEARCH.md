# Perbedaan Antara FindAll (GET) dan Dynamic Search (POST /search)

Dokumen ini menjelaskan perbedaan fundamental antara dua pendekatan utama untuk mengambil daftar resource di API ini: `FindAll` yang menggunakan HTTP GET, dan `FindAllDynamic` yang menggunakan HTTP POST ke endpoint `/search`.

---

## 1. `FindAll` (HTTP GET) - Pencarian Statis/Sederhana

### Karakteristik:
-   **Metode HTTP:** `GET`
-   **Filter:** Didefinisikan secara statis dan hardcoded di backend. Filter dikirim melalui **query parameters di URL**.
-   **Kompleksitas:** Cocok untuk pencarian sederhana dengan beberapa kriteria filter dasar.
-   **Keterbatasan:**
    -   Terbatas pada filter yang sudah ditentukan di kode.
    -   Panjang URL dapat menjadi batasan untuk jumlah/panjang filter.
    -   Merepresentasikan filter kompleks di URL sangat tidak praktis.

### Contoh Penggunaan `FindAll` (HTTP GET)

Saat ini, hanya endpoint User yang mendukung filter melalui GET query parameters. Endpoint Role dan Access Right (termasuk Endpoints) tidak memiliki filter bawaan untuk `GET` mereka, mereka hanya mengembalikan semua data.

#### Modul: User (`GET /api/v1/users`)

Endpoint ini memungkinkan Anda mencari dan memaginasi pengguna berdasarkan `username` (yang memfilter kolom `name` di database) dan `email`.

**Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/users?page=1&limit=10&username=John&email=@example.com" \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

**Penjelasan Parameter:**
-   `page` (opsional): Nomor halaman hasil. Default: `1`.
-   `limit` (opsional): Jumlah item per halaman. Default: `10`.
-   `username` (opsional): Memfilter user di mana kolom `name` (bukan `username` entitas) mengandung string ini (menggunakan `LIKE %username%`).
-   `email` (opsional): Memfilter user di mana kolom `email` mengandung string ini (menggunakan `LIKE %email%`).

#### Modul: Role (`GET /api/v1/roles`)

Endpoint ini akan mengembalikan **semua** role yang ada tanpa opsi filter tambahan.

**Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/roles" \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

#### Modul: Access Rights (`GET /api/v1/access-rights`)

Endpoint ini akan mengembalikan **semua** access right yang ada tanpa opsi filter tambahan.

**Request:**
```bash
curl -X GET "http://localhost:8080/api/v1/access-rights" \
  -H "Authorization: Bearer <YOUR_TOKEN>"
```

---

## 2. `FindAllDynamic` (HTTP POST /search) - Pencarian Dinamis/Lanjutan

### Karakteristik:
-   **Metode HTTP:** `POST`
-   **Endpoint:** Menggunakan path `/search` (misal: `/api/v1/users/search`).
-   **Filter:** Dikirim melalui **request body dalam format JSON**.
-   **Kompleksitas:** Sangat cocok untuk pencarian lanjutan yang fleksibel, dengan banyak kriteria, operator yang beragam, dan sorting dinamis.
-   **Keunggulan:**
    -   Mendukung objek filter JSON yang kompleks dan bersarang.
    -   Tidak ada batasan panjang URL.
    -   Lebih aman karena data filter tidak terekspos di URL.
    -   Konsisten dalam format `snake_case` untuk nama field dan operator.

### Contoh Penggunaan `FindAllDynamic` (HTTP POST /search)

Untuk contoh lengkap dan mendetail mengenai penggunaan endpoint `POST /search` dengan berbagai filter dan sorting, silakan lihat dokumen:
[`documentation/DYNAMIC_SEARCH_EXAMPLES.md`](DYNAMIC_SEARCH_EXAMPLES.md)

---

## 3. Kapan Menggunakan yang Mana?

-   Gunakan **`FindAll` (GET)** jika Anda hanya membutuhkan daftar dasar resource tanpa filter, atau dengan filter yang sangat minimal dan sudah hardcoded (seperti pagination dan satu atau dua filter string sederhana).
-   Gunakan **`FindAllDynamic` (POST /search)** jika Anda membutuhkan kemampuan pencarian yang fleksibel, seperti:
    -   Filter berdasarkan banyak kolom yang berbeda.
    -   Menggunakan operator perbandingan (lebih besar dari, kurang dari, dll).
    -   Menggunakan operator daftar (`in`, `not_in`).
    -   Melakukan sorting multi-kolom.
    -   Filter kompleks dengan logic AND antar kriteria.

Pendekatan `POST /search` memberikan pengalaman yang jauh lebih kuat dan fleksibel untuk kebutuhan pencarian real-time atau "advanced search" di aplikasi web.
