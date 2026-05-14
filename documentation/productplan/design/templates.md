Berikut adalah **Spesifikasi Level 4: TEMPLATES (Struktur Layout)** untuk NexusOS.

Template adalah level di mana atom, molekul, dan organisme disatukan menjadi struktur halaman yang nyata. Fokus utama di sini adalah **Responsive Logic** dan **Density Adaptation** (kemampuan layout berubah drastis antara mode SaaS dan Enterprise).

---

# 1. Template A: The "Chameleon" Dashboard Layout

**Core Concept:** Satu layout induk (`DashboardLayout.tsx`) yang membungkus konten. Layout ini tidak statis; ia "bernapas" dan berubah bentuk berdasarkan preferensi user (Comfort vs Compact).

### A. The Responsive Grid Structure

- **Teknologi:** CSS Grid atau Flexbox dengan Tailwind v4.
- **Zoning:**
  1.  **Sidebar (Left):** Fixed width atau Collapsed width.
  2.  **Navbar (Top):** Sticky, `z-index` tinggi.
  3.  **Main Content (Right-Bottom):** Area dinamis yang memiliki `padding` variabel.

---

### B. Spesifikasi Sidebar (Navigation)

Sidebar adalah elemen yang paling terpengaruh oleh perubahan mode.

| Atribut           | **Mode: Comfort (SaaS Focus)**                | **Mode: Compact (Enterprise Focus)**         |
| :---------------- | :-------------------------------------------- | :------------------------------------------- |
| **Lebar (Width)** | `280px` (Lebar, lega)                         | `72px` (Rail / Icon-only)                    |
| **Menu Item**     | Icon + Label Teks + Chevron                   | Hanya Icon (Teks muncul via Tooltip)         |
| **Grouping**      | Ada judul grup (misal: "ANALYTICS")           | Judul grup diganti garis pemisah (Divider)   |
| **Sub-menu**      | **Accordion:** Expand ke bawah (push content) | **Popover:** Muncul melayang di samping icon |
| **Visual Style**  | Background putih/slate, border tipis.         | Background lebih gelap/kontras untuk fokus.  |

**Interaksi Penting:**

- **Toggle Trigger:** Tombol kecil (`<ChevronsLeft />`) di bagian bawah sidebar atau di samping logo header untuk mengubah mode ini secara manual.
- **Mobile Behavior:** Pada mobile, sidebar selalu menjadi **Drawer (Off-canvas)** yang meluncur dari kiri, terlepas dari mode apa yang aktif.

---

### C. Spesifikasi Navbar (Header)

Area ini harus sangat bersih untuk menyeimbangkan kepadatan data di bawahnya.

- **Tinggi (Height):**
  - _Comfort:_ `80px` (Border-bottom transparan atau halus).
  - _Compact:_ `56px` (Border-bottom tegas).
- **Komponen Wajib:**
  1.  **Global Search (Kiri):** Menggunakan molekul _Command Menu_ yang sudah didefinisikan (Input transparan dengan icon search).
  2.  **Density Switcher (Tengah/Kanan):**
      - _Component:_ Segmented Control (Toggle Switch 3-way: Comfort / Compact).
      - _Fungsi:_ Ini adalah "Remote Control" utama NexusOS. Saat diklik, variabel CSS global `--spacing` dan `--font-size` berubah instan.
  3.  **Theme Toggle:** Icon Sun/Moon.
  4.  **User Profile:** Avatar + Nama (SaaS) atau Avatar saja (Enterprise).

---

# 2. Template B: Authentication Split Layout

**Inspirasi:** _DashTail_ dan _Horizon UI_ menggunakan pola ini karena konversi tinggi.
**Struktur:** Layout 2 Kolom (50:50 pada Desktop, 100% pada Mobile).

### A. Panel Kiri (Functional Zone)

- **Konten:** Logo NexusOS (Top-left), Heading ("Welcome Back"), Subheading, dan Form Login/Register (Molekul Smart Input).
- **Alignment:** Vertikal & Horizontal Centered (`flex items-center justify-center`).
- **Padding:** Luas (`p-12`) agar user tidak merasa sesak saat mengisi form.
- **Background:** Putih bersih (Light) atau Slate-950 (Dark).

### B. Panel Kanan (Visual / Branding Zone)

Area ini vital untuk SaaS branding. Jangan biarkan kosong.

- **Konten Visual (Pilih salah satu varian):**
  1.  **Varian "Abstract 3D":** Objek 3D abstrak yang melayang dengan _Glassmorphism_ (sesuai tren _Horizon UI_).
  2.  **Varian "Social Proof":** Carousel testimonial user dengan foto besar dan kutipan.
  3.  **Varian "Data Visualization":** Screenshot miring (skewed) dari dashboard NexusOS itu sendiri untuk memamerkan fitur "Compact Mode".
- **Styling:**
  - Background: Gunakan warna `Primary-600` atau Gradient Brand (Indigo ke Violet).
  - Overlay: Tambahkan _noise texture_ tipis untuk kesan premium.
  - **Behavior:** `Hidden on mobile` (Display: none pada viewport < 1024px).

---

# 3. Implementasi Teknis (Tailwind v4 Specs)

Berikan instruksi ini kepada developer untuk mengatur layout global:

```css
/* theme.css - Global Layout Variables */
@theme {
  /* Comfort Defaults */
  --sidebar-width: 280px;
  --navbar-height: 80px;
  --main-padding: 2rem;
}

/* Override saat mode Compact aktif */
[data-density='compact'] {
  --sidebar-width: 72px;
  --navbar-height: 56px;
  --main-padding: 1rem;
}

/* Utility Class untuk Layout Utama */
.layout-grid {
  display: grid;
  /* Sidebar fixed, konten sisanya fluid */
  grid-template-columns: var(--sidebar-width) 1fr;
  grid-template-rows: var(--navbar-height) 1fr;
  height: 100vh;
  transition: grid-template-columns 0.3s cubic-bezier(0.4, 0, 0.2, 1); /* Smooth Animation */
}
```

---

# 4. Deliverables untuk Tim Desain (Figma Checklist)

Minta desainer Anda membuat **2 Frame Master** di Figma:

1.  **"Master Dashboard - Comfort Mode":**
    - Sidebar lebar (Menu Accordion terbuka).
    - Kartu metrik dengan shadow besar.
    - Tabel data dengan padding sel 16px.
2.  **"Master Dashboard - Compact Mode":**
    - Sidebar sempit (Rail).
    - Tabel data "Hyper-Grid" dengan border sel terlihat dan padding 6px.
    - Navbar lebih tipis.

**Catatan Integrasi:** Pastikan desainer menggunakan fitur **"Auto Layout Wrapping"** pada _Auth Layout_ agar saat layar mengecil (mobile), panel kanan turun ke bawah atau hilang secara otomatis tanpa perlu membuat desain mobile terpisah dari nol.

Referensi struktur ini sejalan dengan praktik terbaik pada template modern seperti **DashTail** (yang menawarkan variasi layout menu) dan **Modernize** (yang fokus pada skalabilitas struktur).
