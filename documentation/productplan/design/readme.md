Berikut adalah **Spesifikasi UI Design System (Nexus Design System v1.0)** yang lengkap. Spesifikasi ini dirancang khusus untuk mengakomodasi arsitektur **Hybrid (SaaS + Enterprise)** menggunakan stack **Next.js 16, Tailwind v4, dan Shadcn UI**.

Dokumen ini adalah "Bible" bagi UI Designer dan Frontend Developer Anda.

---

# Nexus Design System: Spesifikasi Visual & Teknis

**Filosofi Desain:** _"Fluid Density"_
Sistem ini tidak memiliki satu tampilan statis. Ia memiliki dua _state_ global yang dikontrol oleh variabel CSS:

1.  **Comfort Mode (SaaS):** Fokus pada estetika, _white space_, tipografi besar, dan _onboarding_ yang ramah (Inspirasi: Horizon UI, Vercel).
2.  **Compact Mode (Enterprise):** Fokus pada kepadatan data, garis tegas, font kecil, dan efisiensi ruang (Inspirasi: Linear, Excel, Modernize).

---

## 1. Palet Warna: "Nebula" (Tailwind v4 Variables)

Kita tidak menggunakan warna Tailwind default. Kita menggunakan _Semantic Tokens_ yang mendukung _Dark Mode_ secara native.

### A. Base Colors (Neutral)

Menggunakan **Slate** (sedikit kebiruan) untuk nuansa teknis modern, bukan Gray (terlalu kusam) atau Zinc (terlalu tajam).

| Token           | Light Mode (Hex)      | Dark Mode (Hex)       | Penggunaan    |
| :-------------- | :-------------------- | :-------------------- | :------------ |
| `background`    | `#FFFFFF`             | `#020617` (Slate-950) | Halaman utama |
| `surface`       | `#F8FAFC` (Slate-50)  | `#0F172A` (Slate-900) | Kartu/Panel   |
| `surface-hover` | `#F1F5F9` (Slate-100) | `#1E293B` (Slate-800) | Hover state   |
| `border`        | `#E2E8F0` (Slate-200) | `#1E293B` (Slate-800) | Garis pemisah |
| `foreground`    | `#0F172A` (Slate-900) | `#F8FAFC` (Slate-50)  | Teks utama    |
| `muted-fg`      | `#64748B` (Slate-500) | `#94A3B8` (Slate-400) | Teks sekunder |

### B. Brand Colors (Themable)

NexusOS default menggunakan **"Deep Indigo"** agar terlihat profesional namun modern (AI-ready).

| Token        | Light Mode         | Dark Mode          | Penggunaan                       |
| :----------- | :----------------- | :----------------- | :------------------------------- |
| `primary`    | `indigo-600`       | `indigo-400`       | Tombol utama, Active State       |
| `primary-fg` | `white`            | `slate-900`        | Teks di atas tombol utama        |
| `secondary`  | `teal-500` #14B8A6 | `teal-400` #2DD4BF | Tombol sekunder, Link alternatif |
| `accent`     | `violet-500`       | `violet-400`       | Gradien AI, Highlights           |
| `info`       | `blue-500` #3B82F6 | `blue-400` #60A5FA | Badge info, Processing state     |

---

## 2. The Chameleon Engine (Dual-Density Variables)

Ini adalah inti dari spesifikasi. Variabel ini **berubah nilai** tergantung mode yang dipilih user.

### A. Radius & Spacing

| Variable            | Value: Comfort (SaaS) | Value: Compact (Enterprise) |
| :------------------ | :-------------------- | :-------------------------- |
| `--radius-lg`       | `12px` (Membulat)     | `6px` (Hampir tajam)        |
| `--radius-sm`       | `6px`                 | `2px`                       |
| `--spacing-page`    | `32px`                | `16px`                      |
| `--spacing-card`    | `24px`                | `12px`                      |
| `--spacing-input-y` | `10px`                | `4px`                       |

### B. Input & Component Sizing

| Component     | Height: Comfort (SaaS) | Height: Compact (Enterprise) |
| :------------ | :--------------------- | :--------------------------- |
| `Button`      | `44px`                 | `32px`                       |
| `Input Field` | `44px`                 | `32px`                       |
| `Table Row`   | `64px` (Dengan Avatar) | `36px` (Teks padat)          |
| `Icon Size`   | `20px`                 | `16px`                       |

---

## 3. Tipografi: "Engineering Precision"

Menggunakan font keluarga **Geist** (Vercel) untuk memaksimalkan kesan "Next.js Native".

- **Primary Font:** `Geist Sans` (UI Elements)
- **Data Font:** `Geist Mono` (Angka di tabel, Kode, ID Transaksi)

### Skala Tipografi (Responsive)

| Token        | Size / Weight  | Line Height (Comfort) | Line Height (Compact) |
| :----------- | :------------- | :-------------------- | :-------------------- |
| `text-h1`    | 24px / Bold    | 1.2                   | 1.2                   |
| `text-body`  | 14px / Regular | 1.6 (Mudah baca)      | 1.3 (Padat)           |
| `text-small` | 13px / Medium  | 1.5                   | 1.2                   |
| `text-xs`    | 12px / Medium  | 1.4                   | 1.1                   |

---

## 4. Spesifikasi Komponen UI (Atomic)

### A. Buttons

- **Primary:** Background `primary`, Shadow `shadow-sm` (Comfort) atau `none` (Compact).
- **Secondary:** Border `input`, Background `transparent`, Hover `accent`.
- **Ghost:** Transparent, Hover `surface-hover`.
- **AI Action:** Gradient border (Indigo to Violet), efek _shimmer_ saat loading.

### B. Cards & Containers

- **SaaS Mode:** Border tipis (`slate-200`) + Shadow lembut (`shadow-lg`). Background putih bersih.
- **Enterprise Mode:** Border tegas (`slate-300`). No Shadow. Background sedikit abu (`slate-50`) untuk kontras tinggi terhadap input putih.

### C. Data Grid (The "Money" Component)

Ini komponen paling krusial untuk bersaing dengan **Modernize** dan **DashTail**.

- **Header:** Background `surface`, Font `text-xs` + Uppercase + Bold.
- **Cell:** Border-bottom `1px solid border`.
- **Interaction:**
  - _Comfort:_ Hover row mengubah warna background menjadi biru sangat muda.
  - _Compact:_ Hover row memberikan _outline_ biru pada baris untuk presisi mata.
- **Features Visuals:**
  - _Pinned Column:_ Berikan bayangan vertikal (`box-shadow: 4px 0 8px -4px rgba(...)`) pada kolom yang dibekukan agar terlihat "mengambang" di atas data lain.

---

## 5. Implementasi Teknis (Tailwind v4 CSS)

Serahkan kode ini kepada developer Anda untuk dimasukkan ke `globals.css`. Ini mendefinisikan sistem "Chameleon" secara otomatis.

```css
@theme {
  /* Definisikan Warna Semantic */
  --color-background: var(--background);
  --color-foreground: var(--foreground);
  --color-primary: var(--primary);
  /* ... dll */
}

/* Base Variables (Default / Comfort Mode) */
:root {
  --background: #ffffff;
  --foreground: #0f172a;
  --primary: #4f46e5; /* Indigo 600 */

  /* Density Variables: Comfort (SaaS) */
  --radius: 0.75rem; /* 12px */
  --input-height: 2.75rem; /* 44px */
  --font-size-base: 0.9375rem; /* 15px */
  --table-cell-padding: 1rem;
}

/* Density Variables: Compact (Enterprise) */
[data-density='compact'] {
  --radius: 0.25rem; /* 4px */
  --input-height: 2rem; /* 32px */
  --font-size-base: 0.8125rem; /* 13px */
  --table-cell-padding: 0.375rem; /* 6px */
}

/* Component Classes */
.btn-primary {
  @apply h-[var(--input-height)] rounded-[var(--radius)] bg-primary text-white px-4 text-[length:var(--font-size-base)];
}

.table-cell {
  @apply p-[var(--table-cell-padding)] text-[length:var(--font-size-base)] border-b border-border;
}
```

---

## 6. Aset & Iconography

- **Icon Set:** Gunakan **Lucide React**. Ini standar industri saat ini (digunakan oleh Shadcn).
- **Style:** Stroke 1.5px (Comfort) vs Stroke 1px (Compact).
- **AI Icons:** Gunakan icon khusus "Sparkles" atau "Stars" dengan warna gradient untuk menandakan fitur AI.

---

## 7. Deliverables untuk Tim Desain (Checklist)

Minta desainer Anda menyerahkan file Figma dengan struktur berikut:

1.  **Page: Tokens** (Warna, Typography, Shadows).
2.  **Page: Components** (Button, Input, Table, Card - dalam varian Comfort & Compact menggunakan Figma Variants).
3.  **Page: Layouts** (Sidebar Navigation, Header).
4.  **Page: Screens** (Dashboard SaaS, Dashboard Enterprise, Login, Settings).

**Instruksi Khusus untuk Desainer:**

> "Tolong gunakan fitur **Figma Variables** untuk `Spacing` dan `Radius`. Buat dua _Modes_ di Figma Variable: 'Comfort' dan 'Compact'. Saya ingin bisa mengubah mode satu halaman penuh hanya dengan mengganti variabel mode di layer root."

Apakah Anda ingin saya buatkan **Prompt AI** spesifik untuk men-generate kode komponen (misalnya: _Complex Data Table_) menggunakan spesifikasi di atas?
