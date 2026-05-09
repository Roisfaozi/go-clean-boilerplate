# Frontend Package Extraction Plan

Dokumen ini memetakan apa saja dari `web/` dan `web-lovable/` yang layak dipindah ke package monorepo, area mana yang sebaiknya tetap tinggal di level app, dan jika ada dua implementasi yang mirip maka mana yang lebih baik dijadikan baseline.

## Tujuan

- Menentukan kandidat package yang realistis
- Menghindari shared package yang terlalu cepat
- Menetapkan source of truth per area
- Mengurangi duplikasi antara `web/` dan `web-lovable/`

## Prinsip Umum

### 1. Shared package hanya untuk hal yang netral

Boleh masuk package:

- UI primitives
- utility murni
- hooks ringan
- DTO / schema bersama
- reusable visual patterns

Belum boleh masuk package:

- auth flow
- tenant flow
- realtime transport
- route guard
- server actions
- API client yang terikat backend strategy saat ini

### 2. `web` tetap source of truth aplikasi utama

`web` saat ini lebih matang untuk:

- auth cookie/session
- tenancy
- realtime ticket-based
- integrasi backend utama

### 3. `web-lovable` lebih kuat sebagai sumber UI pattern

`web-lovable` saat ini lebih kaya untuk:

- reusable presentational component
- CRUD shell
- form widgets
- upload widgets
- design pattern dan showcase

## Ringkasan Keputusan

| Area | Baseline | Status |
|---|---|---|
| `packages/ui` | `web-lovable` | layak dipindah lebih dulu |
| `packages/hooks` | `web` | hanya hook netral |
| `packages/utils` | campuran, mulai dari `web-lovable` | parsial |
| `packages/api-types` | `web-lovable` | layak setelah audit DTO |
| `packages/patterns` | `web-lovable` | layak setelah netralisasi dependency |
| shared API client | `web` secara correctness, tapi jangan dipackagekan dulu | tunda |
| auth / tenant / realtime logic | `web` | tetap di app |

## 1. Kandidat `packages/ui`

Ini adalah area paling aman untuk diekstrak lebih dulu karena sudah overlap jelas di dua app.

### Kandidat kuat

- `accordion`
- `alert-dialog`
- `alert`
- `aspect-ratio`
- `avatar`
- `badge`
- `breadcrumb`
- `button`
- `calendar`
- `card`
- `carousel`
- `chart`
- `checkbox`
- `collapsible`
- `command`
- `context-menu`
- `dialog`
- `drawer`
- `dropdown-menu`
- `form`
- `hover-card`
- `input-otp`
- `input`
- `label`
- `menubar`
- `navigation-menu`
- `pagination`
- `popover`
- `progress`
- `radio-group`
- `resizable`
- `scroll-area`
- `select`
- `separator`
- `sheet`
- `sidebar`
- `skeleton`
- `slider`
- `sonner`
- `switch`
- `table`
- `tabs`
- `textarea`
- `toast`
- `toaster`
- `toggle-group`
- `toggle`
- `tooltip`

### Referensi overlap

- [web/src/components/ui](/home/user/Documents/Riset/Casbin/web/src/components/ui/button.tsx:1)
- [web-lovable/src/components/ui](/home/user/Documents/Riset/Casbin/web-lovable/src/components/ui/button.tsx:1)

### Baseline terbaik

Pilih **`web-lovable`** sebagai baseline awal untuk `packages/ui`.

### Alasan

- lebih netral
- lebih sedikit opini app-specific
- lebih cocok sebagai reusable primitive

Contoh pada [button.tsx](/home/user/Documents/Riset/Casbin/web-lovable/src/components/ui/button.tsx:1) dibanding [button.tsx](/home/user/Documents/Riset/Casbin/web/src/components/ui/button.tsx:1):

- versi `web-lovable` lebih generik
- versi `web` sudah mulai tercampur density token dan variant yang lebih spesifik ke app

### Catatan migrasi

Saat dipindahkan ke package:

- ganti alias `@/` dan `~/`
- hilangkan dependency ke provider app-specific
- pastikan token CSS yang dipakai tersedia di kedua app

## 2. Kandidat `packages/utils`

### Kandidat kuat

- `cn`
- formatter netral
- string helper
- date helper murni

### Referensi

- [web/src/lib/utils.ts](/home/user/Documents/Riset/Casbin/web/src/lib/utils.ts:1)
- [web-lovable/src/lib/utils.ts](/home/user/Documents/Riset/Casbin/web-lovable/src/lib/utils.ts:1)

### Baseline terbaik

Gunakan **versi minimal dari `web-lovable`** sebagai baseline awal, lalu tambahkan utilitas dari `web` satu per satu bila benar-benar reusable.

### Alasan

`web-lovable` punya utilitas dasar yang sangat bersih:

- `cn`

Sedangkan `web` mencampur utilitas netral dan utilitas app-specific:

- formatter angka
- helper file/upload
- helper redirect Next.js
- generator id

### Rekomendasi

Buat package awal yang kecil:

- `packages/utils/src/cn.ts`
- `packages/utils/src/format.ts`
- `packages/utils/src/index.ts`

Sisanya tetap di app sampai benar-benar layak diekstrak.

## 3. Kandidat `packages/hooks`

### Kandidat kuat

- `use-mobile`
- `use-toast`

### Referensi

- [web/src/hooks/use-mobile.tsx](/home/user/Documents/Riset/Casbin/web/src/hooks/use-mobile.tsx:1)
- [web-lovable/src/hooks/use-mobile.tsx](/home/user/Documents/Riset/Casbin/web-lovable/src/hooks/use-mobile.tsx:1)
- [web/src/hooks/use-toast.ts](/home/user/Documents/Riset/Casbin/web/src/hooks/use-toast.ts:1)
- [web-lovable/src/hooks/use-toast.ts](/home/user/Documents/Riset/Casbin/web-lovable/src/hooks/use-toast.ts:1)

### Baseline terbaik

Pilih **`web`** sebagai baseline.

### Alasan

- implementasinya setara
- versi `web` sudah eksplisit `"use client"`
- lebih aman untuk dipakai lintas Next.js dan Vite

### Jangan dipindah dulu

- `use-presence`
- `use-permission`
- `use-ai-chat`
- `use-audit-stream`
- `use-realtime`

Karena semua itu sudah tahu contract backend dan store aplikasi.

## 4. Kandidat `packages/api-types`

Ini salah satu area yang paling bernilai untuk di-share.

### Kandidat kuat

- `User`
- `Role`
- `Organization`
- `OrgMember`
- `Project`
- `AccessRight`
- `Permission`
- `PaginatedResponse`
- schema `zod` untuk request/response

### Referensi

- [web-lovable/src/lib/api/schemas.ts](/home/user/Documents/Riset/Casbin/web-lovable/src/lib/api/schemas.ts:1)
- [web/src/lib/api/auth.ts](/home/user/Documents/Riset/Casbin/web/src/lib/api/auth.ts:1)
- [web/src/lib/api/organizations.ts](/home/user/Documents/Riset/Casbin/web/src/lib/api/organizations.ts:1)

### Baseline terbaik

Pilih **`web-lovable`** sebagai baseline.

### Alasan

- type dan schema lebih terkonsolidasi
- validasi runtime lebih jelas
- lebih mudah dijadikan package seperti `packages/api-types`

### Caveat

Sebelum dijadikan shared package, schema harus diaudit terhadap backend nyata karena beberapa service di `web-lovable` masih memakai asumsi endpoint yang tidak persis sama dengan backend utama.

## 5. Kandidat `packages/patterns`

Ini adalah reusable component yang lebih tinggi levelnya dari primitive UI, tetapi masih cukup netral jika dirapikan.

### Kandidat kuat dari `web-lovable`

- [components/layout/page-header.tsx](/home/user/Documents/Riset/Casbin/web-lovable/src/components/layout/page-header.tsx:1)
- [components/forms/multi-select.tsx](/home/user/Documents/Riset/Casbin/web-lovable/src/components/forms/multi-select.tsx:1)
- [features/shared/crud-table.tsx](/home/user/Documents/Riset/Casbin/web-lovable/src/features/shared/crud-table.tsx:1)
- `crud-form-dialog`
- `delete-dialog`
- `patterns/stat-card`
- `patterns/data-table`
- `upload/*` presentational components
- `realtime/*` presentational widgets

### Baseline terbaik

Pilih **`web-lovable`**.

### Alasan

`web-lovable` lebih kuat di area:

- reusable admin pattern
- generic CRUD shell
- interaction pattern yang lebih kaya
- visual widget yang lebih matang

### Catatan penting

Jangan langsung masukkan semua ini ke `packages/ui`.

Lebih sehat dibuat sebagai:

- `packages/patterns`
- atau `packages/admin-ui`

Karena dependency dan kompleksitasnya lebih tinggi dari primitive UI.

## 6. Shared API Client: Belum Layak

### Referensi

- [web/src/lib/api/client.ts](/home/user/Documents/Riset/Casbin/web/src/lib/api/client.ts:1)
- [web-lovable/src/lib/api/client.ts](/home/user/Documents/Riset/Casbin/web-lovable/src/lib/api/client.ts:1)

### Baseline paling benar

Kalau bicara correctness terhadap backend repo ini, pilih **`web`**.

### Alasan

`web` sudah:

- cookie/session aware
- SSR aware
- auto refresh
- hard logout aware
- tenant header injection aware

Sedangkan `web-lovable` masih:

- bearer token di `localStorage`
- belum tenant-aware dengan benar
- belum align ke flow backend utama

### Keputusan

Jangan buat `packages/api-client` dulu.

Kalau ingin sharing bertahap, pecah menjadi hal kecil:

- `ApiError`
- validator helper
- response shape helper

Bukan full transport layer.

## 7. Area Yang Harus Tetap di `web/`

Karena sangat terkait dengan backend utama dan source of truth aplikasi.

### Tetap di app utama

- auth actions dan auth flow
- websocket provider
- auth provider
- density provider
- tenant-aware organization logic
- dashboard context/provider
- SSR-dependent data flow
- server actions Next.js

### Referensi penting

- [web/src/components/shared/providers/websocket-provider.tsx](/home/user/Documents/Riset/Casbin/web/src/components/shared/providers/websocket-provider.tsx:1)
- [web/src/lib/api/client.ts](/home/user/Documents/Riset/Casbin/web/src/lib/api/client.ts:1)

## 8. Area Yang Harus Tetap di `web-lovable/`

Sampai posisinya benar-benar diputuskan.

### Tetap di app eksploratif

- auth variation pages
- error variation pages
- design system showcase
- component showcase
- mock-heavy demo screens

Area ini lebih cocok sebagai:

- playground
- gallery
- sumber inspirasi visual

bukan package produksi langsung.

## 9. Perbandingan Implementasi Yang Mirip

### `ui/*`

- **Menang:** `web-lovable`
- **Alasan:** lebih netral dan reusable

### `lib/utils.ts`

- **Menang untuk baseline package:** `web-lovable`
- **Alasan:** lebih kecil, lebih bersih
- **Catatan:** utilitas tambahan dari `web` diambil selektif

### `hooks/use-mobile.tsx`

- **Menang:** `web`
- **Alasan:** setara, tetapi versi `web` lebih aman untuk package client hook

### `hooks/use-toast.ts`

- **Menang:** `web`
- **Alasan:** implementasi setara, versi `web` sedikit lebih aman sebagai baseline hook shared

### `lib/api/client.ts`

- **Menang secara correctness:** `web`
- **Keputusan:** jangan dipackagekan dulu

### `lib/api/organizations.ts`

- **Menang secara integrasi backend:** `web`
- **Menang secara type/schema discipline:** `web-lovable`
- **Keputusan:** share schema, bukan service client

## 10. Prioritas Ekstraksi

Urutan yang paling aman:

1. `packages/ui`
2. `packages/hooks`
3. `packages/utils`
4. `packages/api-types`
5. `packages/patterns`
6. evaluasi ulang shared API helper kecil

## 11. Peta File Awal

### Masuk `packages/ui`

Ambil dari `web-lovable` terlebih dulu:

- `src/components/ui/*`

Lalu konsumsi ulang dari `web` dan `web-lovable`.

### Masuk `packages/hooks`

Ambil dari `web`:

- `src/hooks/use-mobile.tsx`
- `src/hooks/use-toast.ts`

### Masuk `packages/utils`

Mulai dari:

- `src/lib/utils.ts` yang dinetralkan menjadi `cn` dan helper murni

### Masuk `packages/api-types`

Ambil dan rapikan dari `web-lovable`:

- `src/lib/api/schemas.ts`

### Masuk `packages/patterns`

Ambil selektif dari `web-lovable`:

- `components/layout/page-header.tsx`
- `components/forms/multi-select.tsx`
- `features/shared/crud-table.tsx`
- `features/shared/crud-form-dialog.tsx`
- `features/shared/delete-dialog.tsx`

### Tetap di app

Di `web`:

- `src/lib/api/client.ts`
- `src/components/shared/providers/*`
- `src/app/actions/*`
- seluruh auth/tenant/realtime transport

Di `web-lovable`:

- auth variation pages
- showcase pages
- mock-heavy feature pages

## 12. Rekomendasi Akhir

Strategi yang paling sehat:

- jadikan `web` sebagai source of truth aplikasi
- jadikan `web-lovable` sebagai source of truth visual pattern
- ekstrak primitive UI dari `web-lovable`
- ekstrak app-aware hooks ringan dari `web`
- ekstrak DTO/schema dari `web-lovable`
- tunda sharing auth/API/realtime logic

Dengan pendekatan itu, package yang dihasilkan akan benar-benar reusable dan tidak membawa kekacauan boundary dari dua app yang saat ini masih berbeda model integrasinya.
