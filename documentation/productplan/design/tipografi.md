Berikut adalah **Spesifikasi Tipografi Lengkap (Typography System)** untuk **NexusOS**.

Spesifikasi ini dirancang untuk mendukung arsitektur **Hybrid**, menyeimbangkan estetika modern (seperti _Horizon UI_ atau _Shadcn_) dengan kebutuhan kepadatan data tinggi (seperti _Excel_ atau _Modernize_).

---

### 1. Keluarga Font (Typeface Family)

Kita menggunakan **Geist** (font dari Vercel) untuk memaksimalkan nuansa "Next.js Native" dan keterbacaan pada UI modern.

- **Primary Font (UI):** `Geist Sans`
  - _Karakter:_ Modern, geometris, sangat terbaca di ukuran kecil.
  - _Penggunaan:_ Headings, Body text, Button labels, Navigation.
- **Data Font (Technical):** `Geist Mono`
  - _Karakter:_ Fixed-width, membedakan `0` dan `O` dengan jelas.
  - _Penggunaan:_ Tabel data finansial, ID Transaksi, Snippet kode, API Keys.

---

### 2. Skala & Hirarki (Type Scale)

Sistem ini menggunakan **Tailwind v4 t-shirt sizing** sebagai basis, namun dengan _line-height_ (leading) yang dikustomisasi untuk mode _Comfort_ vs _Compact_.

#### A. Headings (Judul)

Dirancang untuk hierarki visual yang tajam namun tidak memakan tempat berlebih (terutama di dashboard).

| Token       | Ukuran (Size)   | Berat (Weight) | Tracking (Letter Spacing) | Penggunaan Utama                 |
| :---------- | :-------------- | :------------- | :------------------------ | :------------------------------- |
| **Display** | 36px (2.25rem)  | Bold (700)     | -0.02em                   | Angka KPI Besar, Halaman Landing |
| **H1**      | 24px (1.5rem)   | SemiBold (600) | -0.01em                   | Judul Halaman (Page Title)       |
| **H2**      | 20px (1.25rem)  | SemiBold (600) | -0.01em                   | Judul Section / Card Title       |
| **H3**      | 18px (1.125rem) | Medium (500)   | Normal                    | Sub-section, Modal Title         |
| **H4**      | 16px (1rem)     | Medium (500)   | Normal                    | Group Header di Sidebar          |

#### B. Body & Content (Teks Utama)

Disinilah fitur **Dual-Density** bekerja paling aktif.

| Token         | Size: Comfort (SaaS) | Size: Compact (Enterprise) | Berat   | Penggunaan                    |
| :------------ | :------------------- | :------------------------- | :------ | :---------------------------- |
| **Body-Lg**   | 16px                 | 14px                       | Regular | Paragraf artikel, Chat bubble |
| **Body-Base** | **14px**             | **13px**                   | Regular | **Default UI**, Input text    |
| **Body-Sm**   | 13px                 | 12px                       | Regular | Label form, Teks sekunder     |
| **Caption**   | 12px                 | 11px                       | Medium  | Tooltip, Status Badge, Footer |

---

### 3. Aturan Jarak Baris (Line Height / Leading)

Kunci kenyamanan vs kepadatan data terletak di sini.

- **Comfort Mode (SaaS/Marketing):** Menggunakan `150% - 160%` (Tailwind `leading-relaxed`). Memberikan ruang napas agar mata tidak lelah saat membaca laporan analitik panjang.
- **Compact Mode (Data/Excel):** Menggunakan `120% - 130%` (Tailwind `leading-tight`). Memungkinkan menampilkan 20+ baris data dalam satu layar tanpa _scroll_ berlebih.

---

### 4. Spesifikasi Fungsional (Functional Specs)

Detail khusus untuk komponen interaktif.

#### Button Text

- **Style:** Uppercase (Opsional) atau Sentence Case (Modern).
- **Weight:** `Medium (500)` atau `SemiBold (600)`.
- **Size:** Sama dengan `Body-Base`.
- **Tracking:** Sedikit diperluas (`0.01em`) untuk keterbacaan di tombol berwarna kontras.

#### Tabular Data (Tabel & Angka)

Fitur wajib untuk aplikasi Enterprise/Fintech agar angka sejajar vertikal.

- **Font:** `Geist Mono` atau `Geist Sans` dengan fitur OpenType `tnum` (Tabular Figures).
- **CSS:** `font-variant-numeric: tabular-nums;`
- **Align:** Selalu `Right Align` untuk data mata uang dan angka desimal.

#### Code Snippets & ID

- **Font:** `Geist Mono`.
- **Color:** Biasanya `Slate-600` (Light) atau `Slate-400` (Dark).
- **Background:** Diberikan background tipis (`bg-slate-100`) dan `rounded-md`.

---

### 5. Implementasi Tailwind v4 (`theme.css`)

Berikut adalah konfigurasi CSS Variables yang bisa Anda copy-paste untuk mengaktifkan sistem tipografi dinamis ini.

```css
@theme {
  /* Definisikan Font Family */
  --font-sans: 'Geist Sans', ui-sans-serif, system-ui, sans-serif;
  --font-mono: 'Geist Mono', ui-monospace, SFMono-Regular, monospace;

  /* Definisikan Ukuran Font (Base values) */
  --text-display: 2.25rem;
  --text-h1: 1.5rem;
  --text-h2: 1.25rem;
  --text-body: 0.875rem; /* 14px default */
  --text-small: 0.8125rem; /* 13px */
  --text-caption: 0.75rem; /* 12px */
}

/* Base Typo Rules (Comfort Mode / Default) */
:root {
  --leading-body: 1.6; /* Lega */
  --tracking-base: 0em;
}

/* Compact Mode (Enterprise / Data Dense) */
[data-density='compact'] {
  --text-body: 0.8125rem; /* Turun ke 13px */
  --text-small: 0.75rem; /* Turun ke 12px */
  --leading-body: 1.3; /* Rapat */
}

/* Utility Classes Custom */
.type-h1 {
  font-family: var(--font-sans);
  font-size: var(--text-h1);
  font-weight: 600;
  letter-spacing: -0.02em;
  line-height: 1.2;
}

.type-body {
  font-family: var(--font-sans);
  font-size: var(--text-body);
  line-height: var(--leading-body); /* Berubah dinamis */
}

.type-tabular {
  font-family: var(--font-mono);
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.01em;
}
```

### Rekomendasi Penerapan di Figma

Instruksikan desainer Anda untuk:

1.  Membuat **Text Styles** bernama: `Display`, `H1`, `H2`, `Body/Base`, `Body/Small`.
2.  Jangan kunci ukuran font (px). Gunakan **Figma Variables** untuk `FontSize` dan `LineHeight`.
3.  Buat **Modes** di Figma Variables: Satu kolom "Comfort", satu kolom "Compact". Ini memungkinkan desainer melihat perubahan kepadatan teks secara instan hanya dengan menukar mode di layer parent.
