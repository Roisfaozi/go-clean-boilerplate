# AI Agents / Prompts Repository

Folder ini berisi kumpulan *prompt engineering templates* yang dioptimalkan untuk pengembangan proyek Golang dengan Clean Architecture.

## Cara Penggunaan

### 1. Memulai Fitur Baru
Gunakan **`1_architect_system.md`**.
*   **Tujuan:** Mendapatkan daftar file dan struktur yang benar sebelum coding.
*   **Input:** "Saya ingin membuat fitur manajemen Produk."

### 2. Implementasi Kode
Gunakan **`2_developer_cot.md`**.
*   **Tujuan:** Menghasilkan kode Go yang berkualitas dengan alur pikir yang logis.
*   **Input:** "Implementasikan fitur Produk berdasarkan struktur ini..." (hasil dari prompt 1).

### 3. Pembuatan Unit Test
Gunakan **`3_tester_react.md`**.
*   **Tujuan:** Membuat tes yang mencakup *happy path* dan *edge cases*.
*   **Input:** *Copy-paste* kode `usecase.go` atau `controller.go` yang baru dibuat.

## Tips Tambahan
*   Selalu berikan konteks file yang relevan (gunakan `read_file` jika menggunakan CLI agent).
*   Jika AI melakukan kesalahan, minta ia melakukan "Self-Correction" dengan merujuk kembali ke "Constraints" di prompt awal.
