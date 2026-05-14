Berikut adalah **Spesifikasi Teknis Visual (Visual Specs)** lengkap untuk **Cards, Panels, Shadows, dan Borders** dalam sistem desain NexusOS.

Spesifikasi ini dirancang untuk mengakomodasi transisi mulus antara _Light Mode_ (berbasis kedalaman/shadow) dan _Dark Mode_ (berbasis kontras/border), serta mendukung arsitektur **Tailwind CSS v4** yang digunakan oleh pesaing utama seperti DashTail.

---

# 1. Cards & Panels Strategy

**Filosofi:** Objek di Light Mode "mengapung" di atas kanvas putih, sedangkan objek di Dark Mode "tertanam" atau dibatasi garis tegas untuk memisahkan diri dari _void_ gelap.

### A. Light Mode (Elevation Based)

- **Surface:** `bg-white`
- **Border:** `border-transparent` (Utama) atau `border-slate-100` (Sangat tipis untuk definisi tambahan).
- **Shadow:** Dominan. Menggunakan _layered shadows_ (bayangan berlapis) agar terlihat lembut seperti **Horizon UI**.
- **Hover Effect:** Kartu naik sedikit (`-translate-y-1`) dan bayangan membesar (`shadow-xl`).

### B. Dark Mode (Structure Based)

- **Surface:** `bg-slate-900` (Sedikit lebih terang dari background halaman `slate-950`).
- **Shadow:** **DISABLED** (`shadow-none`). Bayangan tidak terlihat di background gelap.
- **Border (Wajib):** `border-1 border-slate-800`.
- **Inner Highlight (Glass Effect):** Gunakan teknik _Inner Ring_ untuk memberikan efek dimensi "kaca" di bagian atas kartu tanpa menggunakan shadow eksternal.
  - _Code:_ `ring-1 ring-inset ring-white/5` (Opacity 5-10%).

---

# 2. Shadow Specifications (Elevation Scale)

Kita menggunakan sistem "Soft-Diffused Shadows" yang mencampurkan warna abu-abu netral dengan sedikit _tint_ warna brand (Indigo) agar tidak terlihat kotor.

| Token Name           | Tailwind Utility | Spesifikasi Visual (CSS Value)                                                         | Penggunaan Ideal                                              |
| :------------------- | :--------------- | :------------------------------------------------------------------------------------- | :------------------------------------------------------------ |
| **None**             | `shadow-none`    | `0 0 #0000`                                                                            | Dark mode default, elemen flat.                               |
| **Micro (XS)**       | `shadow-xs`      | `0px 1px 2px rgba(15, 23, 42, 0.05)`                                                   | Tombol sekunder, Input fields (inset).                        |
| **Small (SM)**       | `shadow-sm`      | `0px 1px 3px rgba(15, 23, 42, 0.08), 0px 1px 2px -1px rgba(15, 23, 42, 0.1)`           | Card kecil, item dalam list, dropdown items.                  |
| **Medium (MD)**      | `shadow-md`      | `0px 4px 6px -1px rgba(15, 23, 42, 0.08), 0px 2px 4px -1px rgba(15, 23, 42, 0.04)`     | **Standard Card Dashboard**, Panel utama.                     |
| **Large (LG)**       | `shadow-lg`      | `0px 10px 15px -3px rgba(15, 23, 42, 0.08), 0px 4px 6px -2px rgba(15, 23, 42, 0.04)`   | Dropdown Menu, Popover, Sticky Header.                        |
| **Extra Large (XL)** | `shadow-xl`      | `0px 20px 25px -5px rgba(15, 23, 42, 0.08), 0px 10px 10px -5px rgba(15, 23, 42, 0.03)` | Modal (Dialog), Floating Action Button (FAB).                 |
| **Inner**            | `shadow-inner`   | `inset 0 2px 4px 0 rgba(0, 0, 0, 0.05)`                                                | Input field aktif (pressed state), Panel "well" (tempat log). |

**Implementasi Tailwind v4 (`theme.css`):**

```css
@theme {
  --shadow-md:
    0 4px 6px -1px rgb(15 23 42 / 0.08), 0 2px 4px -2px rgb(15 23 42 / 0.08);
  --shadow-lg:
    0 10px 15px -3px rgb(15 23 42 / 0.08), 0 4px 6px -4px rgb(15 23 42 / 0.08);
}
```

---

# 3. Border Specifications (Structure Scale)

Border di NexusOS berfungsi sebagai pemisah data yang tegas, terutama untuk mode **Enterprise (Compact)**.

### A. Ketebalan (Border Width)

| Token        | Width   | Penggunaan                                                                             |
| :----------- | :------ | :------------------------------------------------------------------------------------- |
| `border-0`   | 0px     | Card di Light Mode (bersih).                                                           |
| `border`     | **1px** | **Default Global**. Input, Card (Dark Mode), Table Row divider.                        |
| `border-2`   | 2px     | State Aktif (misal: Tab yang dipilih), Avatar Ring, Focus ring.                        |
| `border-l-4` | 4px     | Indikator status di sebelah kiri alert/notifikasi (misal: Garis merah di Error Alert). |

### B. Warna & Gaya (Border Colors)

Warna border harus adaptif (berubah otomatis di Dark Mode).

| Semantic Token   | Light Value  | Dark Value   | Penggunaan                                   |
| :--------------- | :----------- | :----------- | :------------------------------------------- |
| `border-subtle`  | `slate-100`  | `slate-800`  | Garis tabel internal, divider halus.         |
| `border-DEFAULT` | `slate-200`  | `slate-700`  | Input field normal, Card border (Dark Mode). |
| `border-strong`  | `slate-300`  | `slate-600`  | Input field (Hover), Button secondary.       |
| `border-active`  | `indigo-500` | `indigo-400` | Input (Focus), Selected Item.                |
| `border-error`   | `red-500`    | `red-500`    | Input validasi error.                        |

### C. Glassmorphism Highlight (Dark Mode Only)

Untuk mencapai efek "Premium Dark" seperti yang diminta:

```css
/* Utility Class Custom untuk Dark Mode Card */
.card-glass-dark {
  @apply border border-slate-800 bg-slate-900 shadow-none;
  /* Efek Inner Glow / Highlight Atas */
  @apply ring-1 ring-inset ring-white/10;
}
```

_Visual:_ Ini akan menciptakan garis putih super tipis (transparan) di _dalam_ border utama, memberikan efek seolah-olah kartu tersebut terkena cahaya dari atas, memisahkan visual kartu dari background yang pekat.

---

# 4. Rekomendasi Integrasi Desain (Figma & Code)

1.  **Variable Collection di Figma:**
    - Buat collection bernama **"Elevation"**.
    - Isi dengan mode: _Light_ (berisi nilai shadow) dan _Dark_ (berisi nilai shadow `none` tapi nilai stroke/border aktif).
    - Ini memungkinkan desainer mengganti mode dan melihat kartu berubah dari "Shadowy" menjadi "Bordery" secara otomatis.

2.  **Tailwind v4 Configuration:**
    - Gunakan fitur `@theme` baru di Tailwind v4 untuk menimpa default shadow yang terlalu hitam/kasar. Sesuaikan alpha channel (opacity) menjadi lebih rendah (`0.08` alih-alih default `0.1`) untuk tampilan SaaS modern ala **DashTail**.

3.  **Komponen Shadcn UI:**
    - Update komponen `Card` di `components/ui/card.tsx` dengan logic kondisional:
      ```tsx
      // Contoh implementasi class
      className={cn(
        "rounded-xl bg-card text-card-foreground",
        // Light Mode: Shadow lembut, border tipis
        "shadow-md border-slate-100",
        // Dark Mode: No shadow, border tegas, inner ring highlight
        "dark:shadow-none dark:border-slate-800 dark:ring-1 dark:ring-white/5",
        className
      )}
      ```

Dengan spesifikasi ini, **NexusOS** akan memiliki kedalaman visual yang "mahal" di Light Mode dan struktur yang tajam serta futuristik di Dark Mode, mengungguli tampilan standar **Modernize** atau **MatDash**.
