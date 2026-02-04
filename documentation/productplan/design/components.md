Berdasarkan analisis arsitektur **Atomic Design** dan tren dari kompetitor seperti **Untitled UI** dan **MatDash**, berikut adalah spesifikasi komponen lengkap untuk **NexusOS**.

Sistem ini dirancang untuk **Tailwind CSS v4** dan **Shadcn UI**, dengan fokus pada fitur unik kita: **"Fluid Density"** (kemampuan berubah dari mode SaaS yang lega ke Enterprise yang padat).

---

### **Struktur Direktori Atomic (Next.js 16)**

Agar proyek tetap terorganisir, kita akan membagi folder `components` berdasarkan metodologi Atomic Design:

```text
/components
  /atoms        (Base primitives: Button, Input, Icon, Badge)
  /sections    (Simple groups: SearchBar, FormField, DatePicker)
  /shared    (Complex blocks: HyperGrid, AI-ChatWindow, Sidebar)
  /templates    (Page structures: DashboardLayout, AuthLayout)
```

---

### **1. Level 1: ATOMS (Primitif & Token)**

_Komponen terkecil yang tidak bisa dipecah lagi. Fokus pada variasi visual dan "Density Variables"._

#### **A. Button (Tombol)**

Menggunakan `shadcn/ui` button sebagai basis, dimodifikasi untuk _Hybrid Mode_.

- **Variants:**
  - `primary`: Background Nebula-600, Text White.
  - `outline`: Border Slate-300, Text Slate-700.
  - `ghost`: Hover only (untuk toolbar tabel).
  - `magic`: Gradient border (Violet-Indigo) untuk fitur AI.
- **Density Logic (Tailwind v4):**
  - _SaaS Mode:_ `h-11 px-5 text-sm rounded-xl` (Nyaman disentuh).
  - _Enterprise Mode:_ `h-8 px-3 text-xs rounded-md` (Efisien ruang).

#### **B. Input & Select**

- **Visual:** Border minimalis (`border-slate-200`).
- **States:** Default, Hover, Focus (Ring Indigo-500/20), Error (Border Red-500).
- **Density Logic:**
  - _SaaS:_ Tinggi `44px`, Padding dalam luas.
  - _Enterprise:_ Tinggi `32px`, Padding rapat. Font `Geist Mono` untuk input angka.

#### **C. Badge / Status Chip**

Digunakan masif di dalam tabel Enterprise.

- **Style:** `subtle` (Background transparan, Teks berwarna) vs `solid`.
- **Shape:** `rounded-full` (SaaS) vs `rounded-sm` (Enterprise).

#### **D. Iconography**

- **Library:** **Lucide React** (Standar industri untuk performa ringan).
- **Size:** Variabel `--icon-size` yang berubah dari `20px` (SaaS) ke `16px` (Enterprise).

---

### **2. Level 2: MOLECULES (Gabungan Sederhana)**

_Menggabungkan Atom menjadi fungsi dasar._

#### **A. Smart Form Field**

Menggabungkan: `Label` + `Input` + `Error Message` + `Hint`.

- **Fitur AI:** Tambahkan tombol "Sparkles" kecil di dalam input (kanan) untuk _Autofill with AI_ (Inspirasi: **MatDash**).
- **Layout:**
  - _SaaS:_ Spacing vertikal `gap-2`.
  - _Enterprise:_ Spacing vertikal `gap-1` atau layout horizontal (Label kiri, Input kanan) untuk menghemat ruang vertikal.

#### **B. Search Bar Global (Command Menu)**

- **Komponen:** Input + Search Icon + Shortcut Badge (`⌘K`).
- **Behavior:** Membuka modal `cmdk` (Shadcn Command) untuk navigasi cepat tanpa mouse. Ini krusial untuk power user di mode Enterprise.

#### **C. Toast Notification**

- **Style:**
  - _Success:_ Border kiri hijau tebal.
  - _AI Processing:_ Animasi _shimmer_ ungu saat AI sedang bekerja di background.
- **Position:** Pojok kanan atas (SaaS) vs Pojok kanan bawah (Enterprise - agar tidak menutupi navigasi utama).

---

### **3. Level 3: ORGANISMS (Komponen Bisnis Kompleks)**

_Bagian ini adalah USP (Unique Selling Point) utama NexusOS._

#### **A. The Hyper-Grid (Data Table)**

Target: Mengalahkan performa tabel **Modernize** dan **DashTail**.

- **Komponen Penyusun:** Table Header, Pagination, Filter Popover.
- **Fitur Wajib:**
  - **Sticky Header & Columns:** Kolom pertama (ID/Nama) terkunci saat scroll horizontal.
  - **Density Toggle:** Tombol di toolbar tabel untuk switch instan antara _Comfort_ (padding 16px) dan _Compact_ (padding 6px).
  - **Row Actions:** Menu titik tiga (`...`) yang muncul hanya saat _hover_ row untuk mengurangi _visual noise_.
  - **Zebra Striping:** Hanya aktif di _Dark Mode_ atau _Enterprise Mode_ untuk keterbacaan data padat.

#### **B. AI Command Center (Widget Chat)**

Bukan sekadar pop-up, tapi panel _dockable_.

- **Structure:**
  - _Header:_ Status AI (Online/Thinking/Streaming).
  - _Body:_ Chat bubble dengan dukungan Markdown (untuk kode/tabel).
  - _Input Area:_ Textarea yang membesar otomatis + tombol Attachment.
- **Mode:**
  - _Float:_ Seperti widget bantuan biasa.
  - _Split View:_ Membagi layar menjadi dua (Kiri: Dashboard, Kanan: AI Chat) untuk _context-aware assistance_.

#### **C. Metric Cards (KPI)**

- **SaaS Variant:** Angka besar, Icon besar, Background putih bersih dengan shadow lembut.
- **Enterprise Variant:** Ukuran kartu lebih kecil, tanpa shadow (hanya border), disertai _Sparkline Chart_ (grafik garis mini) untuk melihat tren tanpa membuka detail.

---

### **4. Level 4: TEMPLATES (Struktur Layout)**

_Kerangka halaman tanpa data nyata._

#### **A. The "Chameleon" Dashboard Layout**

Satu layout yang beradaptasi dengan mode user.

- **Sidebar:**
  - _Expanded:_ Menu lengkap dengan akordion (SaaS).
  - _Collapsed (Rail):_ Hanya ikon navigasi (Enterprise).
- **Navbar:** Berisi Global Search, Density Switcher, Theme Toggle (Dark/Light), dan User Profile.

#### **B. Authentication Split Layout**

- **Kiri:** Form Login/Register (Atom & Molekul).
- **Kanan:** Visual artistik atau _Testimonial Carousel_ (Penting untuk branding SaaS).

---

### **Prompt untuk Developer (Implementasi)**

Gunakan prompt ini untuk membantu Tim Dev men-generate kode awal menggunakan AI (Cursor/V0):

```markdown
Generate a "Smart Data Table" organism using React, Tailwind v4, and Shadcn UI.
Requirements:

1.  **Atomic Structure:** Use atomic components (Button, Input, Badge) defined previously.
2.  **Dual Density:** Accept a prop `density="comfort" | "compact"`.
    - If "compact": Use `py-1 text-xs`, show vertical borders.
    - If "comfort": Use `py-4 text-sm`, remove vertical borders.
3.  **Features:** Include a sticky header, row selection checkbox, and a pagination footer.
4.  **Styling:** Use `border-slate-200` for light mode and `border-slate-800` for dark mode.
```

### **Langkah Selanjutnya:**

1.  **Desainer:** Mulai buat _Library Figma_ berdasarkan daftar **Atoms** di atas. Pastikan variabel _Density_ sudah diset.
2.  **Developer:** Buat folder `components/ui` dan mulai _coding_ komponen `Button` dan `Input` dengan dukungan varian _Compact_.

Apakah Anda ingin kita mendalami detail **Logika CSS untuk Density Switcher** agar developer bisa langsung copy-paste?
