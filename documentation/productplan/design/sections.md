Berikut adalah definisi spesifikasi lengkap untuk **Level 2: MOLECULES (Komponen Molekul)** dalam Design System NexusOS.

Spesifikasi ini menggabungkan logika **Atomic Design**, fitur **AI-First** (terinspirasi dari MatDash), dan arsitektur **Dual-Density** (SaaS vs Enterprise) yang telah kita bangun sebelumnya.

---

# NexusOS Molecules: Functional Components

Molekul adalah kelompok atom yang bekerja bersama sebagai satu unit fungsional. Pada level ini, "Chameleon Engine" (logika kepadatan) mulai mengatur **tata letak (layout)** dan **posisi**, bukan hanya ukuran.

---

### **A. Smart Form Field**

_Komponen input cerdas yang sadar konteks, validasi, dan memiliki kemampuan AI._

**Komposisi Atom:** `Label` + `Input Atom` + `Icon (Help/AI)` + `Error Message`

#### 1. Spesifikasi Visual & Layout

Variabel `--spacing-input-y` dan `--gap-form` mengontrol kepadatan.

- **Mode A: Comfort (SaaS/Onboarding)**
  - **Layout:** **Vertical Stack** (Label di atas Input).
  - **Spacing:** `gap-2` (8px) antara label dan input.
  - **Visual:** Label menggunakan font `text-sm` weight `medium`. Input terlihat "chunky" dan mudah diketuk.
  - _User Psychology:_ Fokus pada kejelasan instruksi, meminimalkan kesalahan user baru.

- **Mode B: Compact (Enterprise/Data Entry)**
  - **Layout:** **Horizontal / Grid Layout** (Label di Kiri, Input di Kanan).
  - **Grid Ratio:** Label 30% : Input 70%.
  - **Spacing:** `gap-4` horizontal, `items-center` secara vertikal.
  - **Visual:** Label menggunakan font `text-xs` warna `muted-fg`.
  - _User Psychology:_ Efisiensi vertikal. User bisa memindai (scan) form panjang dengan gerakan mata vertikal lurus ke bawah tanpa zig-zag.

#### 2. Fitur AI "Magic Fill" (Inspired by MatDash)

Fitur untuk membedakan NexusOS dari template standar.

- **Posisi:** Absolut di dalam input, sebelah kanan (`right-3`).
- **Ikon:** `Sparkles` (Lucide) dengan ukuran 14px.
- **Style:**
  - _Idle:_ Warna `Slate-400` (subtle).
  - _Hover:_ Gradient Text `Violet-500` ke `Fuchsia-500` + Tooltip "Auto-generate with AI".
  - _Loading:_ Animasi `spin-slow` dengan warna gradient.
- **Interaksi:** Klik ikon → Trigger API request → Streaming text masuk ke input field.

#### 3. States (Tailwind Classes)

- **Default:** Border `slate-200` (Light) / `slate-700` (Dark).
- **Error:** Border `red-500` + Teks error `text-xs` di bawah input.
- **Focus:** Ring `indigo-500/20` (Nexus Blue).

---

### **B. Global Search Bar (Command Menu)**

_Pusat navigasi keyboard-first untuk power user._

**Komposisi Atom:** `Search Icon` + `Input Transparent` + `Badge (Shortcut)`

#### 1. Implementasi Teknis

Menggunakan library **`cmdk`** (basis dari Shadcn Command) untuk aksesibilitas dan performa tinggi.

#### 2. Spesifikasi Visual

- **Container:**
  - _Desktop:_ Lebar tetap (misal: 250px) atau `w-full` di dalam sidebar.
  - _Mobile:_ Hanya ikon kaca pembesar yang men-trigger modal fullscreen.
- **Style:**
  - Background: `bg-slate-100` (Light) / `bg-slate-900` (Dark).
  - Border: `border-transparent` (agar menyatu dengan header/sidebar).
  - Radius: Sesuaikan variabel global `--radius-md`.

#### 3. Interaksi & Behavior

- **Shortcut Badge:** Tampilkan badge `⌘K` (Mac) atau `Ctrl+K` (Windows) di ujung kanan input.
  - _Visual Badge:_ `bg-white` (Light) / `bg-slate-800` (Dark), `text-xs`, shadow-sm.
- **Focus State:** Saat diklik, tidak hanya border yang menyala, tapi langsung membuka **Modal Overlay** (Dialog) di tengah layar dengan background backdrop blur.
- **Enterprise Utility:**
  - Di mode Enterprise, Command Menu ini harus bisa melakukan "Jump to ID" (misal: ketik `#INV-2024` langsung buka detail invoice).

---

### **C. Toast Notification (Smart Alerts)**

_Sistem notifikasi non-intrusif (Toaster)._

**Komposisi Atom:** `Icon (Status)` + `Title` + `Description` + `Close Button`

#### 1. Varian Style (Semantic)

Menggunakan sistem border berwarna (color-coded borders) untuk identifikasi cepat tanpa membaca teks.

- **Success:**
  - Border Kiri: `border-l-4 border-emerald-500`.
  - Icon: `CheckCircle2` warna Emerald.
- **Error/Critical:**
  - Border Kiri: `border-l-4 border-red-500`.
  - Icon: `AlertCircle` warna Red.
  - _Behavior:_ Persisten (tidak hilang otomatis) sampai user klik close.
- **AI Processing (Thinking State):**
  - Background: `bg-slate-900` (Dark) atau `bg-white` (Light).
  - Border: `border border-indigo-500/30`.
  - **Efek Khusus:** Tambahkan animasi CSS `shimmer` (kilau bergerak) berwarna ungu/ungu muda di background untuk menandakan AI sedang "berpikir" atau memproses data di latar belakang.

#### 2. Positioning Logic (Hybrid Mode)

Posisi toast berubah berdasarkan kepadatan informasi di layar.

- **Mode SaaS (Comfort):**
  - **Posisi:** `top-right`.
  - _Alasan:_ Area mata user SaaS biasanya "F-Pattern", dimulai dari kiri atas ke kanan atas. Notifikasi di sini terlihat jelas dan ramah.
- **Mode Enterprise (Compact):**
  - **Posisi:** `bottom-right`.
  - _Alasan:_ Bagian atas aplikasi Enterprise biasanya penuh dengan _Utility Bar_, _Filter Bar_, dan _Breadcrumbs_ yang padat. Notifikasi di atas akan menutupi navigasi penting. Memindahkannya ke bawah menjaga area kerja tetap bersih.

---

### **Prompt Implementasi untuk Developer**

_(Serahkan ini ke tim Frontend/Cursor AI)_

```markdown
**Task:** Create Level 2 Molecule Components for NexusOS.

**1. SmartInput Component:**

- Wrap native `<input>` with a `div` container.
- Accept prop `density="comfort" | "compact"`.
- If `density="compact"`, apply `flex-row items-center justify-between` to label and input container.
- Add an optional `isAI={true}` prop. If true, render a `SparklesIcon` button absolutely positioned right-3 inside the input. On hover, show tooltip "Auto-fill with AI".

**2. CommandMenu Trigger:**

- Use `shadcn/ui` dialog trigger.
- Style looks like an input but acts as a button.
- Include `CMD+K` badge using `<Kbd>` component logic.

**3. Toaster Logic:**

- Customize `sonner` or `shadcn/toast`.
- Create a specific variant `variant="ai-processing"` that applies a purple shimmer animation class (`animate-shimmer`).
- Configure `Toaster` provider to accept a `position` prop that changes dynamically based on the global app state (Zustand store `useDensityStore`).
```
