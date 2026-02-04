Berdasarkan analisis sumber daya pesaing seperti **Untitled UI** (yang memiliki library 4.600+ ikon netral) dan standar teknis **DashTail** yang menggunakan **Shadcn UI** (berbasis Lucide React), berikut adalah definisi **Sistem Iconography** lengkap untuk NexusOS.

Sistem ini dirancang untuk mendukung fitur **"Dual-Density"** (Comfort vs Compact) yang menjadi nilai jual utama NexusOS.

---

# Spesifikasi Iconography NexusOS

**Library Basis:** **Lucide React**
**Alasan:** Standar industri untuk ekosistem Shadcn UI/Next.js. Ringan (tree-shakable), berbasis SVG (vektor), dan memiliki gaya goresan (stroke) yang konsisten dan modern, berbeda dengan Material Icons yang terasa "kaku" atau FontAwesome yang terasa "jadul".

---

### 1. Aturan Visual & Dimensi (The Chameleon Logic)

Ikon di NexusOS bukan sekadar hiasan, melainkan alat navigasi dan indikator status. Ikon harus beradaptasi berdasarkan mode yang dipilih user.

#### A. Ukuran & Skala (Size Tokens)

Jangan gunakan pixel manual. Gunakan variabel CSS agar ikon mengecil otomatis saat user pindah ke "Enterprise Mode".

| Token     | Ukuran (SaaS/Comfort) | Ukuran (Enterprise/Compact) | Penggunaan                                         |
| :-------- | :-------------------- | :-------------------------- | :------------------------------------------------- |
| `icon-sm` | 16px                  | 14px                        | Input field icons, metadata (tanggal, user).       |
| `icon-md` | **20px** (Default)    | **16px** (Default)          | **Navigasi Sidebar**, Tombol Action, Header Tabel. |
| `icon-lg` | 24px                  | 20px                        | Modal Header, KPI Cards utama.                     |
| `icon-xl` | 32px                  | 24px                        | Empty States, Dashboard Welcome.                   |

#### B. Ketebalan Garis (Stroke Width)

Ini adalah rahasia agar UI terlihat "mahal" (SaaS) atau "presisi" (Excel-like).

- **Comfort Mode (SaaS):** Gunakan **Stroke 2px**.
  - _Efek:_ Terlihat ramah, tegas, dan mudah dikenali. Mirip gaya **Untitled UI**.
- **Compact Mode (Enterprise):** Gunakan **Stroke 1.5px**.
  - _Efek:_ Terlihat ringan, teknis, dan memberikan lebih banyak _whitespace_ visual di tabel yang padat data.

---

### 2. Gaya & Konsistensi (Style Guide)

#### A. Style: Outlined (Stroked) vs Filled

- **Primary Style:** Selalu gunakan **Outlined (Garis)**. Ini memberikan kesan bersih dan modern.
- **Active State (Sidebar):** Saat menu aktif, **jangan** ubah ikon menjadi _Filled_ (solid). Cukup ubah warnanya menjadi `Primary-600` dan tebalkan stroke-nya sedikit (jika di mode compact).
  - _Alasan:_ Mengubah bentuk ikon (outline ke solid) saat interaksi menciptakan beban kognitif (cognitive load) karena bentuknya berubah.

#### B. Sudut (Corner Radius)

- Pastikan ujung garis ikon (stroke cap) dan sambungan (stroke join) berbentuk **Round**. Ini selaras dengan radius UI "Comfort Mode" yang membulat (12px), menciptakan kesan modern dan tidak tajam.

---

### 3. Ikon Semantik Khusus (Special Categories)

#### A. AI & Magic (Pembeda NexusOS)

Karena kita bersaing dengan **MatDash** yang memiliki fitur AI, kita butuh ikonografi khusus untuk menandakan fitur kecerdasan buatan.

- **Ikon:** `Sparkles`, `Wand2`, `BrainCircuit` (dari Lucide).
- **Perlakuan Khusus:** Jangan gunakan warna solid biasa. Gunakan **Gradient Fill** atau **Gradient Stroke** (Violet to Fuchsia) khusus untuk ikon-ikon ini agar user langsung tahu "Ini fitur AI".

#### B. Indikator Data (Table Controls)

Untuk tabel Enterprise yang padat:

- **Sorting:** Gunakan `ChevronsUpDown` (default), `ChevronUp` (asc), `ChevronDown` (desc). Ukuran harus kecil (`icon-sm`).
- **Menu Row:** Gunakan `MoreHorizontal` (...) bukan vertikal, untuk menghemat ruang vertikal baris tabel.

---

### 4. Implementasi Teknis (Component Wrapper)

Jangan import ikon Lucide secara mentah berulang kali. Buat komponen atomik `Icon` yang menangani logika _Dual Density_ secara otomatis.

**Contoh Logic (React/Tailwind v4):**

```tsx
// components/atoms/Icon.tsx
import { icons } from 'lucide-react'

interface IconProps {
  name: keyof typeof icons
  size?: 'sm' | 'md' | 'lg'
  className?: string
}

const Icon = ({ name, size = 'md', className }: IconProps) => {
  const LucideIcon = icons[name]

  // Tailwind v4 classes mapping
  // Di Compact mode (data-density="compact"), ukuran & stroke mengecil otomatis
  const sizeClasses = {
    sm: 'w-4 h-4 desktop:w-3 desktop:h-3', // Responsive logic
    md: 'w-5 h-5 [data-density=compact]:w-4 [data-density=compact]:h-4',
    lg: 'w-6 h-6 [data-density=compact]:w-5 [data-density=compact]:h-5',
  }

  const strokeWidth = 'stroke-[2px] [data-density=compact]:stroke-[1.5px]'

  return (
    <LucideIcon
      className={`shrink-0 ${sizeClasses[size]} ${strokeWidth} ${className}`}
    />
  )
}
```

---

### 5. Checklist untuk UI Designer (Deliverables)

Minta desainer Anda untuk menyiapkan:

1.  **Icon Set di Figma:** Jangan _detach_ instance. Gunakan library Lucide resmi untuk Figma.
2.  **Base Component:** Buat satu komponen master ikon di Figma yang memiliki properti `Mode` (Comfort/Compact).
    - _Comfort:_ Size 20px, Stroke 2px.
    - _Compact:_ Size 16px, Stroke 1.5px.
3.  **Export Assets:** Tidak perlu export SVG manual jika developer menggunakan `lucide-react`. Cukup berikan nama ikonnya saja (misal: "Gunakan ikon `Users` untuk menu Customer").

Dengan sistem ini, ikonografi NexusOS akan terlihat **konsisten**, **profesional**, dan secara **cerdas beradaptasi** dengan kebutuhan user (apakah mereka butuh tampilan cantik SaaS atau kepadatan data Enterprise).
