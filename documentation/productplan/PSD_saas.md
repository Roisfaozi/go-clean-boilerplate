Berikut adalah **Dokumen Spesifikasi Produk (Product Specification Document)** untuk **NexusOS Edisi SaaS**.

Berbeda dengan edisi Enterprise yang fokus pada "Kepadatan Data", edisi SaaS ini dirancang untuk **"Kecepatan Peluncuran (Time-to-Market)"**. Target utamanya adalah _Indie Hackers_, _Startups_, dan _Agencies_ yang ingin meluncurkan produk AI/SaaS dalam hitungan hari, bukan bulan.

Strategi ini menggabungkan arsitektur modern **MatDash** (React 19 + Tailwind) dengan fungsionalitas bisnis **Saasable** (Auth + Billing).

---

# Product Specification: NexusOS SaaS Edition (Launchpad)

**Versi Dokumen:** 1.0
**Target Market:** SaaS Startups, Micro-SaaS Builders, AI Wrapper Apps.
**Value Proposition:** "The 'Vercel-style' Starter Kit for AI SaaS. Built for speed, optimized for revenue."

---

## 1. Arsitektur Teknis (The Growth Stack)

Dipilih untuk performa maksimal dan _developer experience_ (DX) yang superior, mengalahkan arsitektur lama yang digunakan kompetitor.

| Komponen          | Spesifikasi Terpilih         | Alasan Strategis (Market Trend 2026)                                                                                           |
| :---------------- | :--------------------------- | :----------------------------------------------------------------------------------------------------------------------------- |
| **Framework**     | **Next.js 16 (App Router)**  | Standar industri untuk SEO dan performa. Mendukung _React Server Components_ untuk loading data yang cepat.                    |
| **Styling**       | **Tailwind CSS v4**          | Menggunakan _Rust-based compiler_ untuk build time instan. Kompetitor seperti **DashTail** sudah mengadopsi ini.               |
| **UI Library**    | **Shadcn UI**                | Memberikan kepemilikan kode penuh (_copy-paste_). Lebih fleksibel daripada MUI (Modernize) untuk custom branding unik.         |
| **State**         | **Zustand / Context API**    | Ringan dan sederhana. **MatDash** baru saja meninggalkan Redux demi Context API, kita ikuti tren ini untuk mengurangi _bloat_. |
| **Data Fetching** | **SWR / TanStack Query**     | Untuk _caching_ dan _real-time updates_ yang krusial di dashboard analitik SaaS.                                               |
| **Auth**          | **NextAuth.js (Auth.js) v5** | Integrasi pre-built untuk Google, GitHub, dan Email Magic Links.                                                               |

---

## 2. Modul "SaaS Business Core" (Fitur Wajib)

Fitur ini membedakan produk Anda dari sekadar "template admin" biasa. Ini adalah infrastruktur bisnis yang siap pakai.

### A. Authentication & Onboarding

- **Multi-Auth Support:** Template login yang sudah terintegrasi dengan logika NextAuth/Supabase. Mendukung Social Login dan Email/Password.
- **Onboarding Wizard:** Form _multi-step_ setelah pendaftaran (Contoh: "Set up your workspace" -> "Invite Team" -> "Connect Data").
- **Role Management:** Logika simpel untuk _Admin_ vs _Member_ dalam satu tim (Multi-tenancy lite).

### B. Subscription & Billing UI

- **Pricing Page Template:** Toggle bulanan/tahunan yang interaktif.
- **Billing Portal:** Halaman akun pengguna yang menampilkan status langganan aktif, tanggal pembaruan, dan riwayat invoice (Status: Paid/Pending).
- **Usage Meter:** Komponen UI visual (Progress Bar) untuk menampilkan penggunaan kuota (misal: "750/1000 AI Credits used").

### C. Landing Page Terintegrasi

- Jangan biarkan user mencari template lain untuk halaman depan. Sertakan **Landing Page** dengan desain yang selaras dengan dashboard (Hero, Features, Pricing, FAQ, Footer).
- **SEO Optimized:** Menggunakan metadata Next.js yang benar.

---

## 3. Modul "AI-Native" (Pembeda Utama)

Untuk bersaing langsung dengan **MatDash** dan **WowDash** yang sudah memiliki fitur AI.

### A. Streaming Chat Interface (ChatGPT Clone)

- **Real-time Streaming:** UI Chat yang mendukung efek _typing_ per karakter (bukan loading spinner biasa).
- **Markdown Support:** Render blok kode, tabel, dan list secara otomatis dalam chat bubble.
- **Empty State Templates:** Kartu saran prompt ("Draft an email", "Summarize this text") saat chat kosong.

### B. AI Generator Templates

- **Text-to-Image UI:** Layout grid untuk menampilkan hasil generasi gambar (seperti Midjourney web).
- **AI Writer/Editor:** Integrasi _rich-text editor_ (Tiptap) dengan toolbar "Ask AI to rewrite".

---

## 4. Dashboard & Analytics (SaaS Metrics)

Fokus pada metrik pertumbuhan, bukan sekadar grafik penjualan e-commerce.

- **KPI Cards:** MRR (Monthly Recurring Revenue), Churn Rate, ARPU (Average Revenue Per User), dan Active Users.
- **Interactive Charts:** Gunakan **Recharts** atau **ApexCharts** untuk grafik pertumbuhan user yang minimalis.
- **Activity Feed:** Widget daftar aktivitas terbaru ("User X upgraded to Pro", "User Y cancelled").

---

## 5. Developer Experience & Utilities

Fitur yang membuat developer "jatuh cinta" pada produk Anda.

- **Command Palette (`Cmd+K`):** Navigasi global instan untuk berpindah halaman atau mengganti tema (Dark/Light) tanpa menyentuh mouse.
- **API Key Management:** UI untuk user men-generate dan me-revoke API token (penting untuk SaaS developer-tools).
- **Theme Customizer:** Panel konfigurasi sederhana untuk mengubah warna primer (Primary Color) dan radius border secara global.

---

## 6. Roadmap Pengembangan (SaaS Version)

### Fase 1: MVP (Launch - Minggu 1-4)

- **Fokus:** Infrastruktur Dasar & Auth.
- **Deliverables:**
  - Setup Next.js 16 + Tailwind v4 + Shadcn.
  - Halaman Auth (Login/Register) fungsional dengan NextAuth.
  - Dashboard Analytics sederhana.
  - Landing Page satu halaman.

### Fase 2: AI & Billing (Minggu 5-8)

- **Fokus:** Nilai Jual Utama.
- **Deliverables:**
  - Integrasi UI Chat AI (Streaming ready).
  - Halaman Pricing & Billing UI.
  - Command Palette (`Cmd+K`).

### Fase 3: Expansion & Polish (Minggu 9+)

- **Fokus:** Kelengkapan Aplikasi.
- **Deliverables:**
  - Aplikasi tambahan: Kanban Board, Email/Inbox UI.
  - Dokumentasi lengkap "How to connect Stripe" & "How to connect OpenAI".

---

## 7. Strategi Harga & Lisensi (SaaS Market)

Menggunakan model yang agresif untuk menarik _early adopters_.

| Tipe Lisensi          | Harga          | Target         | Fitur                                                                      |
| :-------------------- | :------------- | :------------- | :------------------------------------------------------------------------- |
| **Hobby / Solo**      | **$29 - $49**  | Indie Hacker   | Single Project, Full Source Code, Lifetime Updates.                        |
| **Startup / Team**    | **$99 - $129** | Startup Kecil  | Unlimited Projects, Figma Files, Priority Support.                         |
| **Agency / Extended** | **$299**       | Software House | Hak untuk menggunakan di produk klien yang berbayar (SaaS Redistribution). |

**Kesimpulan:**
NexusOS SaaS Edition harus terasa **"Ringan, Cepat, dan Pintar"**. Hindari fitur enterprise yang berat (seperti tabel data kompleks dengan 100 kolom). Fokuslah pada **estetika**, **integrasi AI**, dan **alur pembayaran (billing flow)** yang mulus. Ini adalah apa yang dicari pembeli template di tahun 2026.
