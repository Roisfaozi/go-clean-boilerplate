Berikut adalah spesifikasi teknis dan visual untuk **Level 3: ORGANISMS (Komponen Bisnis Kompleks)**.

Komponen ini adalah **USP (Unique Selling Point)** dari NexusOS, dirancang untuk melampaui standar _Modernize_ (yang kuat di fitur) dan _DashTail_ (yang kuat di performa) dengan menggabungkan estetika SaaS dan kepadatan Enterprise.

---

# 1. Organism A: The Hyper-Grid (Advanced Data Table)

**Core Tech:** TanStack Table v8 (Headless) + Tailwind v4 + Shadcn UI.
**Target:** Menangani 10.000+ baris data tanpa lag, dengan UX yang adaptif.

### A. Komponen Penyusun (Molecules)

1.  **Toolbar:** Global Search, Density Toggle (Comfort/Compact), Column Visibility Dropdown, Filter Badges.
2.  **Pagination Footer:** Rows per page selector (10/20/50/100), Page navigation, "Showing x-y of z results".
3.  **Bulk Actions:** Muncul melayang (floating bar) di bawah header hanya saat 1+ baris dipilih (Delete, Export, Edit Status).

### B. Fitur Wajib & Spesifikasi Visual

#### 1. Sticky Header & Columns (Excel-like Experience)

- **Behavior:** Header tabel (`<thead>`) dan Kolom Pertama (biasanya ID atau Nama) menggunakan `position: sticky`.
- **Visual Logic:**
  - Saat _scroll_ vertikal: Header mendapat border-bottom yang lebih tegas (`border-b-2`) dan sedikit _elevation_ (`shadow-sm`).
  - Saat _scroll_ horizontal: Kolom yang terkunci (pinned) mendapat bayangan vertikal di sisi kanan (`shadow-[4px_0_24px_rgba(0,0,0,0.02)]`) untuk menciptakan efek "melayang" di atas kolom lain.

#### 2. The Chameleon Density Engine (Toggle)

Tombol switch di toolbar mengubah variabel CSS global untuk tabel ini secara real-time.

| Atribut CSS        | **Comfort Mode (SaaS)** | **Compact Mode (Enterprise)**     |
| :----------------- | :---------------------- | :-------------------------------- |
| `Padding-Y` (Cell) | `py-4` (16px)           | `py-1.5` (6px)                    |
| `Font Size`        | `text-sm` (14px)        | `text-xs` (13px)                  |
| `Header Style`     | Clean, Text Grey        | Uppercase, Bold, Bg-Muted         |
| `Grid Lines`       | Horizontal Only         | Horizontal & Vertical (Full Grid) |

#### 3. Row Actions (On-Demand UI)

Untuk mengurangi kekacauan visual (visual noise) pada tabel padat:

- **Default State:** Kolom aksi di ujung kanan kosong/transparan.
- **Hover State:** Saat kursor di atas baris (`tr:hover`), tombol aksi (Menu titik tiga `...` atau tombol Edit/Delete) muncul dengan animasi _fade-in_ cepat (`duration-100`).
- **Implementation:** Gunakan class Tailwind `opacity-0 group-hover:opacity-100`.

#### 4. Smart Zebra Striping (Context Aware)

Tidak seperti _Modernize_ yang statis, striping di NexusOS bersifat kondisional:

- **Light Mode:** Polos (White background). Fokus pada _whitespace_.
- **Dark Mode / Enterprise Mode:** Otomatis mengaktifkan Zebra Striping.
  - _Rumus:_ `odd:bg-white/5` (sangat halus) untuk membantu mata melacak data secara horizontal di layar gelap.

---

# 2. Organism B: AI Command Center (Dockable Widget)

**Inspirasi:** _MatDash_ memiliki fitur AI, tapi kita akan membuatnya lebih terintegrasi ke workflow, bukan sekadar modal.

### A. Struktur Visual

1.  **Header (Status Bar):**
    - Indikator Status: Dot hijau (Online) atau animasi _pulse_ ungu (Processing/Thinking).
    - Action: Tombol "Dock/Undock" dan "Close".
2.  **Body (Stream Area):**
    - Mendukung rendering **Markdown** penuh (untuk menampilkan tabel hasil generate AI atau kode).
    - _User Bubble:_ `bg-slate-100` (Light) / `bg-slate-800` (Dark).
    - _AI Bubble:_ Border tipis gradient `border-indigo-200` + Background `bg-indigo-50/50`.
3.  **Input Area (Prompter):**
    - Textarea _auto-grow_.
    - Tombol "Magic Attach": Ikon klip kertas yang bisa membaca konteks halaman saat ini (misal: "Analisis tabel di halaman ini").

### B. Mode Tampilan (Layout Modes)

#### 1. Float Mode (Default)

- Seperti widget Intercom/Chatbot biasa.
- Posisi: `fixed bottom-6 right-6`.
- Dimensi: Lebar 380px, Tinggi 600px.
- _Use case:_ Pertanyaan cepat ("Cara buat invoice baru?").

#### 2. Split View Mode (Co-Pilot)

Ini fitur "Killer" untuk Enterprise.

- **Behavior:** Saat tombol "Dock" diklik, area konten utama dashboard (`<main>`) mengecil lebarnya (misal dari 100% ke 70%), dan AI Chat mengisi 30% sisa layar di sebelah kanan.
- **Transisi:** Gunakan CSS `grid-template-columns` dengan transisi halus agar layout tidak "melompat" kasar.
- _Use case:_ User bekerja di Data Table sambil meminta AI menganalisis data tersebut secara real-time di panel samping.

---

# 3. Organism C: Metric Cards (KPI)

Satu komponen, dua identitas berbeda. Dikontrol oleh prop `variant="saas" | "enterprise"`.

### A. Variant 1: SaaS (Marketing & Overview)

- **Visual:** "Loud & Proud".
- **Structure:**
  - Icon: Ukuran besar (48px), dalam container lingkaran dengan background warna pastel (`bg-indigo-100 text-indigo-600`).
  - Value: Font `text-3xl` bold.
  - Background: Putih bersih + `shadow-lg` (soft diffused).
  - Border: Tidak ada (mengandalkan shadow).
- **Tujuan:** Memberikan kesan "Progress" dan "Success" yang menyenangkan.

### B. Variant 2: Enterprise (Monitoring & Density)

- **Visual:** "Quiet & Precision".
- **Structure:**
  - Icon: Kecil (16px) atau tidak ada.
  - Value: Font `text-xl` medium.
  - **Sparkline Chart:** Grafik garis mini di sebelah kanan/bawah angka untuk menunjukkan tren 7 hari terakhir tanpa perlu klik detail. Gunakan library `Recharts` (TinyLineChart).
  - Background: `bg-slate-50` (Light) / `bg-slate-900` (Dark).
  - Border: `border border-slate-200`. **Tanpa Shadow**.
  - Ukuran: Lebih pendek (`h-24` vs `h-32` pada SaaS) untuk memuat lebih banyak kartu dalam satu baris.

---

### **Prompt Implementasi untuk Developer (Cursor/V0):**

Gunakan prompt ini untuk men-generate kode awal Organism:

```markdown
**Task:** Build the "Hyper-Grid" Data Table Organism using TanStack Table v8, Tailwind v4, and Shadcn UI.

**Requirements:**

1.  **Component Props:** Accept a `data` array, `columns` definition, and a `density` prop ("comfort" | "compact").
2.  **Styling Logic:**
    - Use Tailwind v4 `@theme` variables.
    - If `density="compact"`, apply `py-1 text-xs border-r border-border` (vertical borders active).
    - If `density="comfort"`, apply `py-4 text-sm` (no vertical borders).
3.  **Sticky Logic:** Implement sticky header (`top-0 z-20`) and sticky first column (`left-0 z-20`). Add a conditional shadow class to the sticky element when scrolling starts.
4.  **Zebra Striping:** Use `dark:odd:bg-muted/50` to enable striping only in dark mode or when a specific `enterprise` prop is true.
5.  **Row Actions:** Include a dummy "Actions" column at the end. The cell content should utilize `opacity-0 group-hover:opacity-100 transition-opacity` to only show on hover.
```
