Berikut adalah **Spesifikasi Desain UI/UX & Creative Brief** yang dirancang khusus untuk **Desainer Produk (UI/UX)**.

Dokumen ini tidak berisi kode CSS/React, melainkan instruksi visual, logika interaksi, dan struktur sistem desain di Figma agar sesuai dengan arsitektur "Hybrid" NexusOS.

---

# Creative Brief & UI Specifications: NexusOS

**Target Audiens Dokumen:** UI/UX Designer, Product Designer
**Tools Wajib:** Figma (dengan fitur _Variables_ & _Auto Layout_ aktif)
**Visi Visual:** "Fluid Density" — Sebuah antarmuka yang bisa bernafas lega untuk startup (SaaS) namun bisa memadat presisi untuk korporat (Enterprise).

---

## 1. Filosofi Desain & Moodboard

Kita menggabungkan dua dunia yang biasanya terpisah. Desainer harus membuat **satu sistem** yang memiliki dua "Wajah" (Modes):

| Atribut Visual         | Mode A: "Comfort" (SaaS Focus)                   | Mode B: "Compact" (Enterprise Focus)         |
| :--------------------- | :----------------------------------------------- | :------------------------------------------- |
| **Inspirasi**          | _Horizon UI, Linear, Vercel_                     | _Modernize, Linear (Density view), Excel_    |
| **Vibe**               | Modern, Airy, Friendly, Marketing-ready          | Technical, Dense, utilitarian, High-contrast |
| **Ruang (Whitespace)** | Luas (Relaxed spacing)                           | Ketat (Tight spacing)                        |
| **Border Radius**      | Membulat (12px - 16px)                           | Tajam/Kecil (2px - 4px)                      |
| **Penggunaan**         | Dashboard founder, Analytics high-level, AI Chat | Data entry, Logistik, Tabel 100+ baris       |

---

## 2. Sistem Warna: "Nebula Palette" (Custom)

Jangan gunakan warna default library. Gunakan palet khusus ini untuk menciptakan identitas **"Trustworthy & Future-Tech"**.

### A. Primary & Secondary Brands

Warna biru-indigo yang dalam, bukan biru terang standar. Terlihat profesional di mata korporat, tapi modern untuk startup.

- **Primary 500 (Main Action):** `#6366F1` (Indigo cerah namun mata tidak sakit)
- **Primary 700 (Hover/Focus):** `#4338CA` (Deep Indigo)
- **Primary 50 (Surface Tint):** `#EEF2FF` (Untuk background baris tabel yang dipilih)
- **Secondary 500 (Alt Action):** `#14B8A6` (Teal - untuk tombol sekunder)
- **Secondary 400 (Dark Mode):** `#2DD4BF` (Teal terang)

### B. Neutrals: "Slate Carbon"

Gunakan _Slate_ (abu-abu kebiruan) bukan _Gray_ murni, untuk memberikan nuansa "Tech".

- **Surface 900 (Text Main):** `#0F172A` (Hampir hitam, teks utama)
- **Surface 500 (Text Muted):** `#64748B` (Teks sekunder/label)
- **Border 200:** `#E2E8F0` (Garis pemisah halus)
- **Background Page:** `#F8FAFC` (Bukan putih murni, supaya mata tidak lelah)

### C. Semantic States (Signal Colors)

- **Success:** `#10B981` (Emerald - bukan Green biasa)
- **Warning:** `#F59E0B` (Amber)
- **Error:** `#DC2626` (Red-600 untuk kontras lebih baik)
- **Info:** `#3B82F6` (Blue-500 untuk badge info, processing)
- **AI/Magic:** Gunakan **Gradient** dari _Violet_ ke _Fuchsia_ (`#8B5CF6` → `#D946EF`) untuk semua elemen yang berhubungan dengan AI.

---

## 3. Spesifikasi Token Desain (Figma Variables)

**Instruksi Kritis untuk Desainer:** Jangan gunakan nilai _hardcoded_ (misal: mengetik "16px" manual). Anda wajib membuat **Collection** di Figma Variables dengan dua mode kolom: **Comfort** dan **Compact**.

### A. Spacing & Sizing Tokens

| Nama Variable (Figma)   | Nilai Mode Comfort | Nilai Mode Compact | Penggunaan                            |
| :---------------------- | :----------------- | :----------------- | :------------------------------------ |
| `spacing-layout-pad`    | 32px               | 16px               | Padding halaman utama                 |
| `spacing-component-gap` | 16px               | 8px                | Jarak antar elemen dalam kartu        |
| `spacing-table-cell-y`  | 16px               | 6px                | Tinggi padding baris tabel (Krusial!) |
| `size-input-height`     | 44px               | 32px               | Tinggi tombol dan input field         |
| `size-icon-base`        | 20px               | 16px               | Ukuran icon                           |

### B. Corner Radius Tokens

| Nama Variable  | Nilai Mode Comfort | Nilai Mode Compact |
| :------------- | :----------------- | :----------------- | ------------------ |
| `radius-card`  | 16px               | 4px                | Sudut kartu/panel  |
| `radius-input` | 8px                | 2px                | Sudut tombol/input |

### C. Typography (Font: Geist Sans / Inter)

Fokus pada keterbacaan angka (tabular nums).

- **Body Base:** 14px (Comfort) vs 13px (Compact).
- **Line Height:** 150% (Comfort - mudah dibaca) vs 120% (Compact - muat banyak data).

---

## 4. Spesifikasi Komponen (Atomic Specs)

### 1. The Hyper-Grid (Tabel Data)

Ini adalah komponen paling penting.

- **Visual Requirements:**
  - **Header:** Font tebal, uppercase, warna `Slate-500`, background `Slate-50`.
  - **Pinning Column:** Berikan visual _drop shadow_ halus di sebelah kanan kolom pertama untuk menandakan kolom itu "beku" saat di-scroll horizontal.
  - **Row Hover:** Saat mouse di atas baris, background berubah warna (jangan pakai outline).
  - **Compact Mode Visual:** Saat mode ini aktif, _grid lines_ (garis batas sel) harus terlihat jelas (seperti Excel). Di mode Comfort, garis vertikal dihilangkan.

### 2. AI Command Center (Chat Interface)

- **Layout:** Jangan buat seperti widget kecil. Buat tampilan _immersive_ (layar penuh) seperti ChatGPT.
- **AI Bubble:** Background gradien tipis atau border tipis berwarna _Magic_ (Violet).
- **User Bubble:** Background `Slate-100`, teks hitam.
- **Interaction:** Desain _state_ "Thinking..." dengan animasi _pulse_ atau _shimmer_, bukan _spinner_ loading bulat biasa.

### 3. Smart Inputs & Forms

- **Input Field:**
  - _Default:_ Border `Slate-200`.
  - _Focus:_ Border `Primary-500` + _Ring/Glow_ halus (4px spread).
  - _AI Action:_ Tambahkan tombol kecil icon "Stars/Magic" di sebelah kanan dalam input field untuk fitur "Auto-fill with AI".

### 4. Navigation (Sidebar)

- **State Collapsed:** Sidebar harus bisa mengecil menjadi hanya ikon (lebar ~64px).
- **Active State:** Menu yang aktif memiliki background `Primary-50` dan garis vertikal tebal (`3px`) di sebelah kiri warna `Primary-500`.

---

## 5. Deliverables yang Diharapkan (Checklist)

Desainer diminta menyerahkan file Figma dengan struktur berikut:

1.  **Halaman "Design Tokens":**
    - Menampilkan palet warna lengkap.
    - Menampilkan tipografi scale.
    - **Wajib:** Tampilan tabel perbandingan _Variables_ (Comfort vs Compact).

2.  **Halaman "Component Library":**
    - Button (Primary, Secondary, Ghost, Destructive).
    - Input Fields (Text, Select, Datepicker).
    - **Master Table Component:** Tabel yang bisa di-resize dengan Auto Layout sempurna.

3.  **Halaman "Key Screens" (Mockup):**
    - **Screen A: Dashboard SaaS (Comfort Mode).** Tampilkan grafik besar, kartu statistik dengan banyak whitespace.
    - **Screen B: Logistics Data Grid (Compact Mode).** Tampilkan tabel dengan 20 kolom dan 50 baris data, terlihat padat tapi rapi.
    - **Screen C: AI Chat Page.** Tampilan percakapan dengan AI.

4.  **Prototyping:**
    - Buat interaksi sederhana di Figma di mana tombol **"Switch Density"** diklik, dan seluruh tampilan layout berubah dari Comfort ke Compact secara instan (menggunakan fitur _Figma Modes_).

---

**Catatan Khusus untuk Desainer:**

> "Kunci keberhasilan proyek ini bukan pada seberapa cantik warnanya, tapi pada seberapa rapi **Auto Layout** dan **Variables** yang Anda buat. Developer akan meng-export token Anda langsung ke Tailwind. Jika Auto Layout Anda berantakan, kode kami akan berantakan."
