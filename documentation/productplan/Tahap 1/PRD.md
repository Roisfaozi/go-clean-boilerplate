Berikut adalah **Product Requirements Document (PRD)** yang komprehensif untuk **NexusOS**. Dokumen ini dirancang sebagai "source of truth" bagi tim produk, *engineering*, dan desain.

Dokumen ini menggabungkan wawasan pasar terbaru (Next.js 16, React 19, Tailwind v4) untuk memastikan produk ini unggul secara teknis dibandingkan **DashTail** dan secara fungsional lebih superior dari **Modernize**.

---

# Product Requirements Document (PRD): NexusOS

| Metadata Proyek | Detail |
| :--- | :--- |
| **Nama Produk** | NexusOS (The Adaptive Enterprise Dashboard) |
| **Versi Dokumen** | 2.0 (Final Draft) |
| **Status** | **Approved for Development** |
| **Target Rilis** | Q4 2026 (Early Access) |
| **Visi Produk** | Jembatan arsitektur yang menyatukan kecepatan peluncuran SaaS (AI-Ready) dengan kekokohan sistem Enterprise (Data-Heavy). |

---

## 1. Pendahuluan & Masalah Bisnis

### 1.1 Latar Belakang
Pasar *template admin* saat ini terfragmentasi menjadi tiga kutub ekstrem berdasarkan data pasar 2025-2026:
1.  **Performance/Cheap:** Didominasi oleh **DashTail** (Tailwind v4, Next.js 16) dengan harga ~$14, namun fitur bisnisnya dangkal.
2.  **Enterprise/Heavy:** Didominasi oleh **Modernize** (MUI, Full Apps), namun arsitekturnya terasa berat dan sulit dikustomisasi karena "vendor lock-in" pada library UI.
3.  **Design-Led:** Didominasi oleh **Horizon UI**, sangat estetik dan terintegrasi Figma, namun lemah dalam menangani tabel data kompleks.

### 1.2 Solusi NexusOS
NexusOS hadir sebagai solusi **"Hybrid"**. Menggunakan stack performa tinggi (Next.js 16 + Tailwind v4) seperti DashTail, namun menyertakan kedalaman logika bisnis (RBAC, Advanced Tables) seperti Modernize, dan fleksibilitas desain (Shadcn UI) untuk menghindari kaku-nya Material Design.

---

## 2. Profil Pengguna (User Personas)

### Persona A: "The Velocity Founder" (Target Modul SaaS)
*   **Profil:** Indie Hacker atau CTO Startup AI.
*   **Pain Point:** "Saya tidak butuh 100 komponen UI. Saya butuh Auth, Stripe, dan AI Chat yang sudah jalan dalam 2 jam."
*   **Kebutuhan NexusOS:** Integrasi **NextAuth**, **Billing Portal**, dan **Vercel AI SDK Hooks**.

### Persona B: "The Corporate Architect" (Target Modul Enterprise)
*   **Profil:** Senior Dev di perusahaan logistik/fintech.
*   **Pain Point:** "Horizon UI terlalu banyak *whitespace* untuk data gudang. Modernize terlalu berat untuk dimodifikasi."
*   **Kebutuhan NexusOS:** **High-Density Data Grid** (TanStack Table), **RBAC**, dan **Audit Logs**.

---

## 3. Spesifikasi Fungsional (Functional Requirements - FR)

### FR-01: Core Architecture & Tech Stack
Harus menggunakan teknologi standar industri 2026 untuk menjamin *longevity*.

*   **FR-01.1 Framework:** **Next.js 16 (App Router)**. Wajib mendukung *React Server Components* (RSC) untuk efisiensi data fetching.
*   **FR-01.2 Rendering Engine:** **React 19**. Memanfaatkan *React Compiler* untuk menghilangkan kebutuhan `useMemo`/`useCallback` manual.
*   **FR-01.3 Styling:** **Tailwind CSS v4**. Menggunakan compiler berbasis Rust untuk build time instan (<100ms).
*   **FR-01.4 UI Primitives:** **Shadcn UI**. Komponen harus bersifat *copy-paste* (dimiliki penuh oleh developer), bukan *npm package* tertutup.

### FR-02: The "Chameleon" Density Engine (Unique Value Proposition)
Fitur pembeda utama untuk melayani dua persona sekaligus.

*   **FR-02.1 Mode Toggle:** Switch global di navbar (Comfort vs Compact).
*   **FR-02.2 Comfort Mode:** Padding 16px, Radius 8px-12px (Cocok untuk SaaS/Horizon UI style).
*   **FR-02.3 Compact Mode:** Padding 4px-8px, Font 12px, Radius 4px (Cocok untuk Enterprise/Excel style).
*   **FR-02.4 Implementasi:** Menggunakan *CSS Variables* global yang dire-inject secara *runtime* tanpa *page reload*.

### FR-03: AI-Native Module (SaaS Differentiator)
Mengungguli **MatDash** yang hanya menyediakan UI statis.

*   **FR-03.1 Streaming Chat:** Komponen UI Chat yang terhubung langsung ke **Vercel AI SDK**. Mendukung *streaming response*, *Markdown rendering*, dan *code syntax highlighting*.
*   **FR-03.2 Prompt Management:** Halaman CRUD untuk menyimpan, mengedit, dan menguji *System Prompts* (fitur yang jarang ada di kompetitor).
*   **FR-03.3 AI Form Fill:** Tombol "Magic Fill" pada form input yang menggunakan LLM untuk memparsing teks tidak terstruktur menjadi JSON form data.

### FR-04: Enterprise Data Grid (The "Heavy Lifting")
Mengatasi kelemahan **Horizon UI** dalam menangani data padat.

*   **FR-04.1 Engine:** Integrasi **TanStack Table v8** (Headless).
*   **FR-04.2 Server-Side Operations:** Pagination, Filtering, dan Sorting yang terhubung ke URL parameters (bukan client-side filtering).
*   **FR-04.3 Advanced Features:**
    *   *Column Pinning* (Bekukan kolom kiri/kanan).
    *   *Row Selection* (Shift+Click untuk bulk select).
    *   *Density Toggle* (Sinkron dengan Chameleon Engine).

### FR-05: Security & Governance
*   **FR-05.1 RBAC:** Sistem *Permission Guard* (`<Can I="delete" a="Post">`) untuk menyembunyikan elemen UI berdasarkan role.
*   **FR-05.2 Authentication:** Template login/register terintegrasi **NextAuth v5** (mendukung Google, GitHub, Magic Link).
*   **FR-05.3 Audit Logs:** Halaman template untuk mencatat log aktivitas user (Who, When, What action).

---

## 4. Spesifikasi Non-Fungsional (NFR)

*   **NFR-01 Performa:** Skor Lighthouse > 95. *First Contentful Paint* (FCP) < 0.8s (mengalahkan rata-rata React apps).
*   **NFR-02 Type Safety:** **TypeScript 5.6+** dengan mode `strict: true`. Tidak boleh ada penggunaan `any` implisit.
*   **NFR-03 State Management:** Gunakan **Zustand** (Client State) dan **TanStack Query** (Server State). Hindari Redux untuk mengurangi *boilerplate*, mengikuti jejak MatDash yang sudah meninggalkannya.
*   **NFR-04 Legacy Support:** Sediakan versi **Vite React (SPA)** terpisah untuk developer yang ingin integrasi ke backend Laravel/Django tanpa Node.js server.

---

## 5. Strategi Desain & Integrasi (Design System)

*   **Figma Parity:** File Figma harus menggunakan **Variables** (Color, Spacing, Radius) yang memiliki nama token sama persis dengan `tailwind.config.js`. Ini mengadopsi keunggulan utama **Horizon UI**.
*   **Theming:** Dukungan *Multi-theme* (Violet, Blue, Emerald) dan *Dark Mode* native yang bekerja otomatis dengan komponen Shadcn.

---

## 6. Roadmap Peluncuran (Execution Plan)

| Fase | Durasi | Deliverables Utama | Fokus Pasar |
| :--- | :--- | :--- | :--- |
| **Fase 1: MVP (Launchpad)** | Minggu 1-4 | Next.js 16 + Tailwind v4 Setup, Auth Pages, Basic Dashboard, AI Chat UI (Streaming ready). | SaaS Founders |
| **Fase 2: Enterprise Core** | Minggu 5-8 | **Advanced Data Grid** (TanStack), RBAC System, Billing UI, Compact Mode. | Corporate Devs |
| **Fase 3: Expansion** | Minggu 9-12 | Versi **Vite (SPA)** untuk legacy backend, Dokumentasi Integrasi (.NET/Laravel), Audit Logs. | Agencies / Legacy Systems |

---

## 7. Strategi Harga & Lisensi (Monetization)

Menggunakan model hibrida untuk bersaing dengan harga murah DashTail ($14) dan kelengkapan Modernize ($59+).

1.  **Solo License ($49):**
    *   1 Developer, Unlimited Projects.
    *   Termasuk: Next.js Source Code, Figma File.
    *   *Strategy:* Harga masuk akal untuk kualitas premium, mengalahkan lisensi "Single Project" kompetitor.
2.  **Team License ($129):**
    *   Up to 5 Developers.
    *   Akses Git Repository Private.
3.  **Enterprise / Extended ($399):**
    *   Hak Redistribusi (SaaS di mana user akhir membayar).
    *   Termasuk Modul Audit Logs & Versi Legacy (Vite).
    *   *Strategy:* Bersaing agresif dengan lisensi Extended WowDash ($700).

---

## 8. Kriteria Kesuksesan (Success Metrics)

*   **Teknis:** Build time di bawah 30 detik di Vercel.
*   **Adopsi:** 500 penjualan dalam 3 bulan pertama.
*   **Kepuasan:** Rating rata-rata 4.8/5 di marketplace (fokus pada kualitas dokumentasi dan kemudahan kustomisasi).

**Persetujuan Dokumen:**
Dokumen ini berlaku sebagai acuan final untuk memulai tahap desain (Figma) dan pengembangan (Coding). Segala perubahan fitur major harus melalui revisi dokumen ini.