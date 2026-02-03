Berikut adalah **Business Requirement Document (BRD)** yang komprehensif dan mendetail untuk proyek **"NexusOS"**. Dokumen ini dirancang sebagai acuan utama bagi tim manajemen produk, _engineering_, dan desain untuk mengeksekusi visi produk yang telah kita diskusikan.

Dokumen ini menggabungkan data pasar terkini (2025-2026) untuk memastikan produk ini secara teknis unggul dari **DashTail** dan secara fungsional lebih fleksibel daripada **Modernize**.

---

# Business Requirement Document (BRD)

**Nama Proyek:** NexusOS (The Adaptive Enterprise Dashboard)
**Versi Dokumen:** 1.0 (Final Draft)
**Tanggal:** 23 Mei 2025
**Status:** Approved for Development

---

## 1. Ringkasan Eksekutif (Executive Summary)

**Latar Belakang:**
Pasar _admin dashboard_ saat ini terpolarisasi. Developer dipaksa memilih antara:

1.  **Kecepatan & Estetika (SaaS-focus):** Seperti _DashTail_ atau _Horizon UI_, yang ringan dan cantik namun lemah dalam fitur manajemen data kompleks.
2.  **Stabilitas & Kelengkapan (Enterprise-focus):** Seperti _Modernize_ (AdminMart), yang memiliki fitur lengkap namun menggunakan arsitektur berat (MUI/Redux) yang sulit dikustomisasi.

**Tujuan Bisnis:**
Mengembangkan **NexusOS**, dashboard _hybrid_ pertama yang menyatukan arsitektur modern (Next.js 16 + Tailwind v4) dengan fitur data kelas berat (TanStack Table + RBAC). Tujuannya adalah merebut 15-20% pangsa pasar dari pengguna Modernize yang menginginkan _stack_ lebih ringan, dan pengguna DashTail yang membutuhkan fitur bisnis lebih dalam.

---

## 2. Ruang Lingkup Proyek (Project Scope)

### 2.1 In-Scope (Fokus Rilis V1)

- **Core Framework:** Pengembangan berbasis **Next.js 16** (App Router) dan **React 19**.
- **Design System:** Implementasi **Tailwind CSS v4** dengan **Shadcn UI** (Headless components).
- **Modul Enterprise:** Advanced Data Grid, RBAC (Role-Based Access Control), dan Audit Logs.
- **Modul SaaS:** Integrasi AI (Chat/GenAI) dan Subscription Billing UI.
- **Legacy Support:** Versi **React Vite** (SPA) tanpa ketergantungan Next.js server.

### 2.2 Out-of-Scope (Fase Selanjutnya)

- Versi Angular atau Vue (dijadwalkan untuk V2/V3).
- Aplikasi Mobile Native (React Native).
- Backend API penuh (Node.js/Python) – V1 fokus pada _Frontend Starter Kit_ dengan integrasi _BaaS_ (Firebase/Supabase).

---

## 3. Profil Pengguna (Target Audience)

| Persona                                   | Deskripsi & Kebutuhan (Pain Points)                                                                              | Solusi NexusOS                                                            |
| :---------------------------------------- | :--------------------------------------------------------------------------------------------------------------- | :------------------------------------------------------------------------ |
| **The Indie Founder** (Target SaaS)       | Ingin meluncurkan produk AI dalam <48 jam. Kecewa dengan DashTail karena harus coding ulang Auth & Billing.      | **SaaS Launchpad:** Pre-built Auth, Stripe UI, & AI Hooks.                |
| **The Corporate Dev** (Target Enterprise) | Membangun internal tool logistik. Modernize terlalu berat (MUI), Horizon UI terlalu boros tempat (_whitespace_). | **Compact Mode & Data Grid:** UI padat informasi & Tabel performa tinggi. |
| **The Agency** (Target Legacy)            | Klien menggunakan backend Laravel/.NET lama. Tidak bisa deploy Node.js server.                                   | **Vite/Static Version:** Versi HTML/React murni yang mudah di-_embed_.    |

---

## 4. Spesifikasi Fungsional (Functional Requirements - FR)

### FR-01: The "Chameleon" Density Engine (Unique Value Proposition)

- **FR-01.1:** Sistem harus memiliki toggle global "Density Mode".
  - _Comfort Mode:_ Padding 16px, Font 14px/16px, Border Radius 8px (Default untuk SaaS).
  - _Compact Mode:_ Padding 4px-8px, Font 12px/13px, Border Radius 4px, Border tajam (Untuk Enterprise).
- **FR-01.2:** Perubahan mode tidak boleh me-reload halaman (_instant state change_ menggunakan CSS variables).

### FR-02: Enterprise Data Grid (Killer Feature)

Menggantikan kelemahan tabel standar pada kompetitor _Horizon UI_.

- **FR-02.1:** Integrasi **TanStack Table v8** (Headless).
- **FR-02.2 Fitur Wajib:**
  - _Server-side Pagination:_ Mampu menangani 100.000+ baris data.
  - _Column Pinning:_ Membekukan kolom "Action" atau "ID" saat scroll horizontal.
  - _Multi-column Sorting:_ Sortir berdasarkan "Status" lalu "Tanggal".
  - _Bulk Actions:_ Checkbox row selection dengan action bar melayang (Delete/Export).

### FR-03: AI-Native Integration Layer

Mengungguli _MatDash_ yang hanya menyediakan UI statis.

- **FR-03.1:** Integrasi **Vercel AI SDK** (`useChat`, `useCompletion`).
- **FR-03.2:** Komponen UI Chat mendukung _Streaming Text_ (efek mengetik) dan rendering Markdown (untuk kode).
- **FR-03.3:** **Prompt Manager UI:** Halaman CRUD untuk menyimpan template _system prompts_.

### FR-04: Security & Compliance Module

Fitur pembeda utama dari template murah ($14) seperti _DashTail_.

- **FR-04.1 RBAC System:**
  - Menyediakan Higher-Order Component (HOC) atau Wrapper `<PermissionGate>` untuk menyembunyikan elemen UI berdasarkan role user.
- **FR-04.2 Audit Logs:**
  - Halaman template visual untuk menampilkan log aktivitas (Who, When, What, IP Address).
- **FR-04.3 API Key Management:**
  - UI untuk generate/revoke API tokens bagi user.

### FR-05: Authentication & SaaS Logic

- **FR-05.1:** Dukungan multi-provider (NextAuth v5): Google, GitHub, Email Magic Link.
- **FR-05.2:** Halaman Login/Register yang terpisah dari layout dashboard utama.
- **FR-05.3:** Halaman _Pricing_ dengan toggle Bulanan/Tahunan yang fungsional (UI state logic).

---

## 5. Spesifikasi Non-Fungsional (Non-Functional Requirements - NFR)

### NFR-01: Performance & Tech Stack

- **Framework:** **Next.js 16** dengan pemanfaatan _React Server Components_ (RSC) untuk mengurangi _client-side bundle size_.
- **Styling:** **Tailwind CSS v4** (Rust compiler) untuk _build time_ instan (<100ms HMR).
- **Lighthouse Score:** Target skor Performance > 95 pada mode produksi.

### NFR-02: Code Quality & Maintainability

- **Type Safety:** **TypeScript 5.6+** dengan konfigurasi `strict: true`. Dilarang menggunakan `any` tanpa alasan krusial.
- **State Management:** Gunakan **Zustand** untuk _global client state_ dan **TanStack Query/SWR** untuk _server state_. Hindari Redux (mengikuti jejak MatDash yang meninggalkannya).
- **Component Ownership:** Kode komponen UI (Button, Input, Modal) harus berada di dalam folder proyek (`/components/ui`) menggunakan arsitektur **Shadcn UI**, bukan di-_import_ dari `node_modules` tertutup (seperti MUI).

---

## 6. Strategi Integrasi Desain (Design Handoff)

- **Figma Parity:** Variabel di Figma (Colors, Spacing, Radius) harus 1:1 dengan `tailwind.config.js`. Ini adalah nilai jual utama _Horizon UI_ yang harus diadopsi.
- **Dark Mode Strategy:** Dukungan _native_ dark mode menggunakan CSS variables, bukan class `dark:` manual di setiap elemen, agar tema mudah diganti.

---

## 7. Roadmap & Milestone Peluncuran

| Fase                   | Durasi      | Deliverables Utama                                                              | Target Pasar               |
| :--------------------- | :---------- | :------------------------------------------------------------------------------ | :------------------------- |
| **Fase 1: MVP**        | Minggu 1-4  | Next.js 16 Setup, Auth Pages, Basic Dashboard, AI Chat UI (Dummy).              | Early Adopters / Indie Dev |
| **Fase 2: Core V1**    | Minggu 5-8  | **Advanced Data Grid**, RBAC System, Billing UI, Integrasi AI SDK Fungsional.   | SaaS Founders              |
| **Fase 3: Enterprise** | Minggu 9-12 | **Compact Mode**, Audit Logs, Versi Vite (Legacy Support), Dokumentasi Lengkap. | Corporate Devs             |

---

## 8. Strategi Monetisasi (Pricing Model)

Menggunakan model _Tiered Licensing_ untuk memaksimalkan _Average Order Value (AOV)_:

1.  **Starter License ($49):**
    - Untuk Solo Developer/Freelancer.
    - Single Project.
    - Mencakup: Next.js Source Code, Figma File.
2.  **Company License ($129):**
    - Untuk Tim Kecil/Agensi.
    - Unlimited Projects.
    - Mencakup: Akses Repo GitHub Private, Priority Support.
3.  **Enterprise / Extended License ($399):**
    - Untuk SaaS yang memungut bayaran dari user (_Charge End Users_).
    - Mencakup: Modul Audit Logs, **Versi Legacy (Vite/HTML)**, Hak Redistribusi Komersial.

---

## 9. Analisis Risiko & Mitigasi

- **Risiko:** Perubahan API Next.js yang cepat (misal: perubahan pada App Router/Server Actions).
  - _Mitigasi:_ Tim dev harus memantau rilis _Canary_ Next.js dan menggunakan _Dependencies_ versi _fixed_ (bukan `^version`) di `package.json` untuk stabilitas rilis.
- **Risiko:** Kompleksitas _Advanced Data Grid_ membuat bundle size membengkak.
  - _Mitigasi:_ Gunakan _Code Splitting_ agresif dan _Lazy Loading_ untuk komponen tabel yang berat. Pastikan TanStack Table di-_tree-shake_ dengan benar.

---

**Persetujuan:**
Dokumen ini disetujui sebagai landasan pengembangan NexusOS. Setiap perubahan fitur utama (misal: penggantian framework CSS) harus melalui revisi dokumen ini.
