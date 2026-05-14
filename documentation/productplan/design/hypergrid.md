Tentu, ini adalah tambahan spesifik untuk **The Hyper-Grid (Tabel Data)** dalam dokumen **Spesifikasi Dark Mode (Nexus Design System)**.

Spesifikasi ini dirancang untuk mengatasi masalah utama pada tabel _dark mode_ standar (seperti pada _Horizon UI_): hilangnya fokus mata saat membaca baris data yang panjang (tracking error).

---

### Tambahan Spesifikasi Dark Mode: The Hyper-Grid

**Komponen:** `Organism / Data Table`
**Prioritas:** Critical (Wajib untuk Mode Enterprise/Compact)

#### 1. Visual Logic (Zebra Striping & Interaction)

Pada mode gelap, garis batas (`border`) saja seringkali tidak cukup untuk memandu mata pada dataset >50 baris. Kita menggunakan pendekatan **"Alternating Surfaces"** untuk memisahkan data tanpa menambah kekacauan visual.

| State / Bagian           | Spesifikasi Warna (Tailwind Token) | Hex Code & Opacity      | Visual Outcome                                                                                                                  |
| :----------------------- | :--------------------------------- | :---------------------- | :------------------------------------------------------------------------------------------------------------------------------ |
| **Row Default (Genap)**  | `bg-transparent`                   | `N/A`                   | Menyatu dengan background panel (`Slate-950`).                                                                                  |
| **Row Striped (Ganjil)** | `bg-slate-900/50`                  | `#0F172A` (50% Opacity) | Memberikan lapisan "kaca film" sangat tipis. Cukup gelap untuk membedakan baris, tapi tetap transparan.                         |
| **Row Hover**            | `bg-indigo-500/10`                 | `#6366F1` (10% Opacity) | **Interaction Cue.** Highlight Indigo tipis saat kursor lewat. 10% adalah "sweet spot" agar teks putih tetap kontras (WCAG AA). |
| **Selected Row**         | `bg-indigo-500/20`                 | `#6366F1` (20% Opacity) | Lebih terang dari hover untuk menandakan baris yang dicentang.                                                                  |

#### 2. Implementasi Teknis (Tailwind v4 CSS)

Tambahkan aturan ini ke dalam layer `components` atau langsung pada file CSS global Anda untuk memastikan konsistensi di seluruh tabel.

```css
@layer components {
  /* Aturan Global Tabel Enterprise di Dark Mode */
  .dark .hyper-grid-table tr:nth-child(odd) {
    /* Row Ganjil: Slate-900 dengan 50% opacity */
    background-color: rgb(15 23 42 / 0.5);
  }

  .dark .hyper-grid-table tr:nth-child(even) {
    /* Row Genap: Transparan */
    background-color: transparent;
  }

  .dark .hyper-grid-table tr:hover {
    /* Hover Row: Indigo-500 dengan 10% opacity */
    /* Menggunakan !important jika perlu menimpa style bawaan Shadcn */
    background-color: rgb(99 102 241 / 0.1);
    transition: background-color 0.15s ease;
  }

  /* Selected Row (Checkbox Active) */
  .dark .hyper-grid-table tr[data-state='selected'] {
    background-color: rgb(99 102 241 / 0.2);
    border-left: 2px solid rgb(99 102 241); /* Indikator visual tambahan */
  }
}
```

#### 3. Instruksi Modifikasi Komponen Shadcn (`components/ui/table.tsx`)

Komponen standar Shadcn menggunakan `hover:bg-muted/50`. Anda perlu mengubahnya agar sesuai dengan spesifikasi NexusOS.

**Instruksi untuk Developer:**
"Buka file `table.tsx`. Pada elemen `<TableRow>`, ganti class default hover dengan logic kondisional atau utility class baru kita."

```tsx
// components/ui/table.tsx

const TableRow = React.forwardRef<
  HTMLTableRowElement,
  React.HTMLAttributes<HTMLTableRowElement>
>(({ className, ...props }, ref) => (
  <tr
    ref={ref}
    className={cn(
      'border-b transition-colors data-[state=selected]:bg-muted',
      // HAPUS atau GANTI class default Shadcn: "hover:bg-muted/50"

      // GUNAKAN class NexusOS:
      // Light Mode: Hover abu-abu tipis
      'hover:bg-slate-50',

      // Dark Mode: Spesifikasi Hyper-Grid (Indigo Tint)
      'dark:hover:bg-indigo-500/10',

      // Striping Logic (Opsional bisa dipasang di sini atau via CSS global)
      'dark:even:bg-transparent dark:odd:bg-slate-900/50',

      className,
    )}
    {...props}
  />
))
```

#### 4. Referensi Kompetitor (Why we do this?)

- **Modernize:** Menggunakan tabel yang sangat bersih tapi seringkali _flat_ di dark mode, membuat mata lelah saat scanning horizontal. Zebra striping kita mengatasi ini.
- **DashTail:** Menggunakan Tailwind v4, yang memudahkan kita menggunakan opacity modifier (`/10`, `/50`) secara langsung tanpa menulis CSS custom yang berat.
- **MatDash:** Fokus pada "Clean Code", jadi implementasi kita harus langsung pada level komponen (atomic), bukan _override_ CSS global yang berantakan.

Dengan spesifikasi ini, tabel data NexusOS akan terlihat presisi seperti **Excel** di malam hari, namun secantik **SaaS modern** saat berinteraksi.
