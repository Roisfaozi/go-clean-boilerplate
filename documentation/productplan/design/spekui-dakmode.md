Berikut adalah **Spesifikasi Dark Mode (NexusOS: The Eclipse Protocol)**.

Dokumen ini melengkapi spesifikasi desain sebelumnya. Mengingat **Horizon UI** memiliki dark mode yang sangat _polished_, dan **Modernize** menggunakan standar Material Design, NexusOS harus mengambil pendekatan **"High-Contrast Engineering"**.

Tujuannya bukan sekadar "membalik warna" (invert), melainkan menciptakan lingkungan visual yang nyaman untuk penggunaan durasi lama (data entry) dan terlihat futuristik untuk presentasi SaaS.

---

# Nexus Design System: Dark Mode Specification

**Kode Nama:** "Eclipse"
**Filosofi:** Menghindari _Pure Black_ (`#000000`). Menggunakan _Deep Slate_ untuk mengurangi ketegangan mata (eye strain) dan _Neon Accents_ untuk hierarki visual.

---

## 1. Palet Warna Gelap ("Eclipse Palette")

Kita menggunakan pendekatan _semantic mapping_. Variabel CSS yang sama akan berubah nilai heksadesimalnya saat class `.dark` aktif di `<html>`.

### A. Base Foundations (Backgrounds)

Menggunakan **Slate** (Blue-Grey) untuk memberikan nuansa "Tech/Code Editor" yang premium, mirip dashboard Vercel atau Linear.

| Token           | Light Value (Ref) | **Dark Value (Eclipse)**  | Penjelasan Visual                                                                        |
| :-------------- | :---------------- | :------------------------ | :--------------------------------------------------------------------------------------- |
| `background`    | `#FFFFFF`         | **`#020617` (Slate-950)** | Background utama halaman. Sangat dalam, hampir hitam, tapi ada _tint_ biru.              |
| `surface`       | `#F8FAFC`         | **`#0F172A` (Slate-900)** | Untuk Kartu/Panel. Sedikit lebih terang dari background agar tercipta kedalaman (depth). |
| `surface-hover` | `#F1F5F9`         | **`#1E293B` (Slate-800)** | State saat mouse hover.                                                                  |
| `overlay`       | `#FFFFFF`         | **`#1E293B` (Slate-800)** | Untuk Dropdown menu, Modal, dan Popover.                                                 |

### B. Typography & Borders (Contrast Control)

Masalah utama dark mode adalah teks putih murni (`#FFF`) pada background gelap seringkali terlalu silau ("halalation effect").

| Token           | Light Value (Ref) | **Dark Value (Eclipse)**  | Penjelasan Visual                                                       |
| :-------------- | :---------------- | :------------------------ | :---------------------------------------------------------------------- |
| `foreground`    | `#0F172A`         | **`#F8FAFC` (Slate-50)**  | Teks utama. Putih lembut (Off-white), bukan `#FFFFFF` murni.            |
| `muted-fg`      | `#64748B`         | **`#94A3B8` (Slate-400)** | Teks sekunder. Abu-abu kebiruan terang.                                 |
| `border`        | `#E2E8F0`         | **`#1E293B` (Slate-800)** | Garis pemisah. Harus sangat halus agar UI tidak terlihat "kotak-kotak". |
| `border-active` | `#CBD5E1`         | **`#334155` (Slate-700)** | Border input saat aktif/focus.                                          |

### C. Brand Colors (Adaptive Vibrancy)

Warna `Primary` (Indigo) di mode terang seringkali terlalu gelap/sulit terbaca di mode gelap. Kita harus menggesernya menjadi lebih pastel/neon.

| Token     | Light Value  | **Dark Value**   | Alasan Teknis                                                                                |
| :-------- | :----------- | :--------------- | :------------------------------------------------------------------------------------------- |
| `primary` | `indigo-600` | **`indigo-500`** | Geser 1 tingkat lebih terang agar kontras teks putih di atas tombol tetap terjaga (WCAG AA). |
| `accent`  | `violet-500` | **`violet-400`** | Untuk elemen AI, gunakan warna yang lebih "bercahaya" (neon).                                |
| `danger`  | `red-600`    | **`red-500`**    | Merah gelap sulit dilihat di dark mode; gunakan merah tomat cerah.                           |

---

## 2. Strategi Komponen (Behavioral Changes)

Di Dark Mode, kita tidak bisa mengandalkan _Shadow_ (bayangan) untuk kedalaman karena bayangan tidak terlihat di background gelap. Kita harus menggantinya dengan **Border** atau **Highlight**.

### A. Cards & Panels

- **Light Mode:** Mengandalkan `shadow-lg`, border `transparent` atau sangat tipis.
- **Dark Mode:**
  - **Shadow:** Hilangkan (`shadow-none`).
  - **Border:** Wajib ada. Gunakan `border-1 border-slate-800`.
  - **Highlight:** (Opsional) Tambahkan _inner glow_ putih tipis di bagian atas kartu untuk efek kaca (glassmorphism halus).

### B. The Hyper-Grid (Tabel Data)

Tabel Enterprise sangat krusial.

- **Zebra Striping:** Wajib aktif di Dark Mode untuk memandu mata.
  - _Row Genap:_ `bg-transparent`
  - _Row Ganjil:_ `bg-slate-900/50` (50% opacity).
- **Hover Row:** Gunakan warna `bg-indigo-500/10` (Indigo dengan opacity 10%) agar highlight terlihat jelas tapi tidak menutup teks.

### C. AI Chat Interface

- **User Bubble:** Background `Slate-800` (Neutral).
- **AI Bubble:** Background `Indigo-950` dengan border `Indigo-500/30`. Memberikan kesan bahwa AI "hidup" atau berpendar.
- **Code Block (Markdown):** Gunakan tema syntax highlighting **"Tokyo Night"** atau **"Dracula"** yang kontrasnya tinggi.

---

## 3. Implementasi Teknis (Tailwind v4 CSS)

Tambahkan konfigurasi ini ke file CSS global Anda. Ini memanfaatkan fitur **CSS Variables fallback** untuk transisi otomatis.

```css
@layer base {
  :root {
    /* Light Mode Tokens */
    --background: 0 0% 100%; /* #FFFFFF */
    --foreground: 222.2 84% 4.9%; /* Slate-950 */
    --card: 0 0% 100%;
    --card-foreground: 222.2 84% 4.9%;
    --popover: 0 0% 100%;
    --popover-foreground: 222.2 84% 4.9%;
    --primary: 243 75% 59%; /* Indigo-600 */
    --primary-foreground: 210 40% 98%;
    --border: 214.3 31.8% 91.4%; /* Slate-200 */
    --input: 214.3 31.8% 91.4%;
  }

  .dark {
    /* Eclipse Mode Tokens (Override) */
    --background: 222.2 84% 4.9%; /* Slate-950 (#020617) */
    --foreground: 210 40% 98%; /* Slate-50 */

    --card: 222.2 84% 4.9%; /* Sama dengan BG untuk flat look */
    --card-foreground: 210 40% 98%;

    --popover: 222.2 84% 4.9%;
    --popover-foreground: 210 40% 98%;

    --primary: 243 100% 70%; /* Indigo-400 (Lebih terang/neon) */
    --primary-foreground: 222.2 47.4% 11.2%; /* Teks hitam di tombol neon */

    --border: 217.2 32.6% 17.5%; /* Slate-800 (Garis tegas) */
    --input: 217.2 32.6% 17.5%;
  }
}

/* Utilitas Global untuk Transisi Halus */
* {
  @apply transition-colors duration-200 ease-in-out;
}
```

---

## 4. Visualisasi Data (Charts)

Library chart (Recharts/ApexCharts) sering rusak di dark mode jika warnanya _hardcoded_.

**Spesifikasi Warna Chart (Dark Mode Only):**

- **Series A (Utama):** `#818CF8` (Indigo-400) - Terang & Jelas.
- **Series B (Sekunder):** `#2DD4BF` (Teal-400) - Kontras tinggi terhadap Indigo.
- **Grid Lines:** `#334155` (Slate-700) - Putus-putus dan tipis.
- **Tooltip Background:** `#1E293B` (Slate-800) + Border.

---

## 5. Deliverables untuk Tim Desain (Prompt Update)

Berikut adalah prompt untuk Anda berikan kepada AI atau desainer Figma untuk membuat varian Dark Mode secara otomatis:

```markdown
**Task:** Create the "Eclipse" Dark Mode variant for the NexusOS Design System.

**Constraints:**

1.  **Do not simply invert colors.** Use a dedicated "Slate" palette (#020617 for bg, #1E293B for borders).
2.  **Card Logic:** Removing drop shadows in dark mode and replacing them with a `1px` border using color `Slate-800`.
3.  **Primary Action Color:** Shift the Brand Indigo from `#4F46E5` (600) to `#818CF8` (400) to ensure accessibility against the dark background.
4.  **Inputs:** Dark mode inputs should have a background of `#0F172A` (Slate-900) to distinguish them from the page background `#020617`.

**Output Required:**

- A side-by-side view of the "Enterprise Dashboard" in Light vs. Dark mode.
- A specific "Data Grid" state in Dark Mode showing active row highlighting (use subtle indigo tint).
- Design the "AI Chat" component in dark mode with a glowing effect on the AI avatar.
```

### Action Plan Selanjutnya:

1.  **Desainer:** Duplikasi halaman "Components" di Figma, bungkus dalam _Section_, dan set mode variabel ke "Dark". Perbaiki kontras manual jika ada warna yang "mati".
2.  **Developer:** Copy-paste konfigurasi CSS variabel di atas ke `globals.css` di Next.js.
3.  **Review:** Cek halaman Login. Pastikan form input terlihat jelas batasnya di mode gelap.

Apakah Anda ingin saya buatkan **Daftar Komponen Shadcn UI** yang perlu di-_override_ manual styling-nya untuk dark mode ini? (Biasanya Calendar dan Popover butuh perhatian khusus).
