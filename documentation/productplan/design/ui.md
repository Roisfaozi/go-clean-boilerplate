Berdasarkan analisis mendalam terhadap **Atomic Design** yang digunakan oleh **Untitled UI**, serta standar teknis dari kompetitor papan atas seperti **MatDash** (Shadcn/Tailwind) dan **DashTail** (Next.js/Tailwind v4), berikut adalah definisi lengkap komponen **Atom** yang harus disiapkan untuk **NexusOS**.

Ingat, dalam fase ini kita tidak mendesain "Halaman", melainkan **Token** dan **Elemen Primitif** yang akan menjadi fondasi sistem _Hybrid_ (SaaS + Enterprise).

---

### 1. Design Tokens (Fondasi Variabel)

Sebelum membuat komponen visual, Anda wajib mendefinisikan _Invisible Atoms_ ini di Figma Variables dan `theme.css` Tailwind v4. Ini adalah kunci fitur "Chameleon" (Comfort vs Compact).

- **Color Semantics (Nebula Palette):** Jangan gunakan Hex code langsung. Gunakan token semantik.
  - `bg-surface-primary`: Latar belakang kartu utama.
  - `bg-surface-secondary`: Latar belakang sidebar/header (sedikit lebih gelap/terang).
  - `text-primary`: Judul (Slate-900).
  - `text-muted`: Label/Secondary text (Slate-500).
  - `border-default`: Garis pemisah halus (Slate-200 di Light, Slate-800 di Dark).
- **Spacing & Sizing Variables (The Density Engine):**
  - `--spacing-layout`: Padding halaman (32px vs 16px).
  - `--component-height`: Tinggi input/tombol (44px vs 32px).
- **Radius Variables:**
  - `--radius-lg`: Sudut kartu (12px vs 4px).
  - `--radius-md`: Sudut tombol/input (8px vs 2px).

### 2. Interactive Atoms (Elemen Interaksi Dasar)

Komponen ini adalah yang paling sering digunakan. Fokus pada _State_ (Hover, Active, Disabled, Focus).

#### A. Buttons (Tombol)

Mengacu pada standar **Shadcn UI** yang dimodifikasi untuk _Enterprise grade_.

- **Variants:**
  - `Primary`: Solid background (Brand Indigo).
  - `Secondary/Outline`: Border only (Netral).
  - `Ghost`: Transparent (Untuk toolbar tabel).
  - `Destructive`: Merah (Hapus data).
- **Sizes (Responsive):**
  - _Comfort:_ Tinggi 44px, Padding X 20px, Font 14px.
  - _Compact:_ Tinggi 32px, Padding X 12px, Font 12px.
- **States:** Default, Hover (`brightness-110`), Pressed (`scale-95`), Disabled (`opacity-50`), Loading (dengan Spinner).

#### B. Inputs (Formulir Data)

Kritikal untuk aplikasi Enterprise. Harus mendukung input data cepat.

- **Types:** Text, Number, Password, Email, DatePicker trigger.
- **Visual Specs:**
  - _Border:_ 1px solid `border-default`. Saat fokus, gunakan `ring-2` warna `primary-100`.
  - _Helper Text:_ Teks kecil di bawah input untuk instruksi/error.
- **Hybrid Behavior:** Di mode _Compact_, hilangkan padding vertikal berlebih agar form terlihat padat seperti Excel.

#### C. Toggle / Switch / Checkbox / Radio

- **Checkbox:** Gunakan `accent-color` brand. Di tabel enterprise, checkbox digunakan untuk _multi-row selection_.
- **Switch:** Untuk pengaturan on/off cepat (misal: Dark Mode toggle).

### 3. Typography Atoms (Sistem Teks)

Menggunakan font keluarga **Geist** (Sans & Mono) untuk nuansa teknis modern.

- **Headings:** H1 (24px), H2 (20px), H3 (18px) - Berat: SemiBold.
- **Body:**
  - `Body-Base`: 14px (Reading).
  - `Body-Small`: 13px (Data tables & Labels).
  - `Body-Mono`: 12px (Code snippets, API Keys, ID Transaksi).
- **Links:** Warna `primary`, underline saat hover.

### 4. Feedback & Status Atoms (Indikator)

Penting untuk memberikan konteks visual instan kepada user.

- **Badges (Chips):**
  - _Style:_ `Subtle` (Background transparan + Teks berwarna) lebih disukai daripada `Solid` agar tidak terlalu mencolok di tabel yang padat data.
  - _Variants:_ Success (Hijau), Warning (Kuning), Error (Merah), Info (Biru), Neutral (Abu-abu).
  - _Shape:_ Pill (Bulat penuh) untuk SaaS, Rectangle (Sudut 4px) untuk Enterprise.
- **Avatars:**
  - Lingkaran untuk foto user.
  - Kotak (`rounded-md`) untuk logo perusahaan/proyek.
- **Skeleton Loader:**
  - Animasi _pulse_ abu-abu untuk menggantikan konten saat data sedang dimuat (penting untuk UX _perceived performance_).

### 5. Iconography (Visual Pendukung)

Menggunakan **Lucide React** sebagai standar industri modern.

- **Stroke Width:**
  - _Comfort:_ 1.5px atau 2px (Lebih tebal, ramah).
  - _Compact:_ 1px (Lebih tipis, presisi).
- **Size:** Base size 20px (Comfort) dan 16px (Compact).

### 6. Special Atoms (NexusOS USP)

Atom khusus yang membedakan NexusOS dari kompetitor standar.

- **AI Sparkle Icon:** Ikon khusus dengan gradien warna (Ungu/Pink) untuk menandakan fitur yang ditenagai AI (misal: tombol "Generate Summary").
- **Divider (Separator):**
  - Garis horizontal/vertikal sangat tipis (`1px`, warna `slate-100`).
- **Scrollbar:** Kustomisasi CSS scrollbar agar tipis dan tidak mengganggu visual tabel yang lebar.

---

### **Daftar Checklist untuk Desainer (Deliverables):**

Instruksikan desainer Anda untuk menyerahkan satu file Figma "Library" yang berisi:

1.  **Variable Collection:** `Color`, `Spacing`, `Radius` (dengan mode _Comfort_ & _Compact_).
2.  **Typography Styles:** Heading & Body (Responsive).
3.  **Component Set (Variants):**
    - Button (4 Varian, 5 State).
    - Input (Default, Error, Disabled).
    - Checkbox & Radio.
    - Badge (5 Warna semantic).
    - Avatar.
    - Icon set (Lucide wrapper).

Dengan mendefinisikan atom-atom ini secara ketat di awal, tim developer (Next.js/Tailwind) dapat membangun UI dengan kecepatan tinggi karena mereka hanya perlu "menyusun lego" tanpa memikirkan nilai pixel manual.
