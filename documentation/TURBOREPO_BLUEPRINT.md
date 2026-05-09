# Turborepo Blueprint

Dokumen ini adalah blueprint implementasi Turborepo untuk repo ini. Fokusnya bukan migrasi besar sekaligus, tetapi transisi bertahap dari struktur sekarang ke monorepo yang stabil untuk Go backend + beberapa frontend React.

## Tujuan

- Menjadikan repo ini monorepo polyglot yang rapi
- Memisahkan aplikasi frontend dari package React bersama
- Tetap mempertahankan workflow Go yang sudah ada
- Menghindari refactor besar yang mencampur auth, tenant, dan realtime logic terlalu dini

## Kondisi Saat Ini

- Backend Go berada di root repo dan dikelola dengan `go.mod` + `Makefile`
- Frontend utama ada di `web/`
- Frontend eksploratif / UI-lab ada di `web-lovable/`
- Belum ada workspace monorepo di root
- Ada `pnpm-workspace.yaml` di dalam `web/`, tetapi itu bukan workspace root monorepo
- Belum ada `turbo.json`

## Prinsip Arsitektur

### 1. Root repo tetap menjadi rumah backend Go

Backend tidak perlu dipindah ke `apps/api` pada fase awal. Repo ini sudah punya banyak asumsi terhadap root:

- `go.mod`
- `Makefile`
- `db/`
- `internal/`
- `pkg/`
- `tests/`

Untuk fase awal, lebih aman mempertahankan backend di root.

### 2. Frontend diperlakukan sebagai app terpisah

Target state:

```txt
apps/
  web/
  web-lovable/

packages/
  ui/
  utils/
  api-types/
  config-eslint/
  config-ts/

internal/
pkg/
cmd/
db/
tests/
Makefile
go.mod
package.json
pnpm-workspace.yaml
turbo.json
```

### 3. Yang dibagi hanya yang benar-benar reusable

Boleh di-shared ke `packages/*`:

- presentational UI components
- generic form primitives
- layout primitives
- table shells
- dialog, badges, cards, tabs
- DTO types dan schema yang netral
- utility murni

Jangan di-shared dulu:

- auth store
- organization store
- API client penuh
- realtime transport
- route guard
- server actions Next.js
- kode yang tahu cookie/token flow
- kode yang tahu tenant header strategy

## Tooling Yang Disarankan

- `pnpm` sebagai workspace package manager
- `turbo` sebagai task orchestrator
- `Makefile` tetap untuk backend Go
- `go` tooling tetap native

Tidak perlu dulu:

- Nx
- Nix
- migrasi backend ke package/app khusus

## Target Struktur Monorepo

### Root

```txt
.
├── apps/
│   ├── web/
│   └── web-lovable/
├── packages/
│   ├── ui/
│   ├── utils/
│   ├── api-types/
│   ├── config-eslint/
│   └── config-ts/
├── cmd/
├── db/
├── deploy/
├── docs/
├── documentation/
├── internal/
├── pkg/
├── tests/
├── Makefile
├── go.mod
├── package.json
├── pnpm-workspace.yaml
└── turbo.json
```

### `apps/web`

Frontend utama yang terintegrasi dengan backend saat ini:

- auth cookie/session
- tenant-aware
- realtime ticket-based
- SSR/App Router

### `apps/web-lovable`

Frontend eksploratif:

- design system lab
- kandidat redesign
- tempat validasi pola UI

Secara prinsip, app ini tidak boleh menjadi sumber auth/API contract utama sebelum diselaraskan.

### `packages/ui`

Isi awal yang cocok dipindahkan:

- button
- input
- textarea
- badge
- card
- dialog
- table shell
- page header
- empty state
- loading state
- generic form section

### `packages/utils`

Untuk:

- helper string/date/formatting
- className helpers
- fungsi murni tanpa browser/runtime coupling

### `packages/api-types`

Untuk:

- response types
- entity DTO
- zod schema yang benar-benar shared

Catatan: package ini tidak boleh berisi API client implementation pada fase awal.

## Fase Implementasi

### Fase 0: Persiapan

Tujuan:

- buat fondasi monorepo tanpa mengubah perilaku aplikasi

Langkah:

1. Tambahkan `package.json` di root
2. Tambahkan `pnpm-workspace.yaml` di root
3. Tambahkan `turbo.json` di root
4. Biarkan `web/` dan `web-lovable/` tetap di lokasi sekarang untuk sementara
5. Standarkan script dasar di kedua app

Outcome:

- root bisa menjalankan task lint/build/test frontend
- backend Go belum tersentuh

### Fase 1: Inisiasi Workspace Root

#### Root `package.json`

Contoh awal:

```json
{
  "name": "casbin-monorepo",
  "private": true,
  "packageManager": "pnpm@10.8.0",
  "scripts": {
    "dev:web": "pnpm --dir web dev",
    "dev:web-lovable": "pnpm --dir web-lovable dev",
    "build:web": "pnpm --dir web build",
    "build:web-lovable": "pnpm --dir web-lovable build",
    "lint:web": "pnpm --dir web lint",
    "lint:web-lovable": "pnpm --dir web-lovable lint",
    "test:web-lovable": "pnpm --dir web-lovable test",
    "go:test": "go test ./...",
    "go:test-integration": "make test-integration",
    "go:test-e2e": "make test-e2e",
    "docs": "make docs",
    "lint": "turbo run lint",
    "build": "turbo run build",
    "test": "turbo run test"
  },
  "devDependencies": {
    "turbo": "^2.1.0"
  }
}
```

#### Root `pnpm-workspace.yaml`

Contoh awal:

```yaml
packages:
  - "web"
  - "web-lovable"
  - "apps/*"
  - "packages/*"
```

Catatan:

- Pada tahap awal, `web` dan `web-lovable` masih boleh tetap di root agar migrasi tidak terlalu besar
- Nanti saat fase restrukturisasi, entry `web` dan `web-lovable` bisa dihapus dan diganti ke `apps/*` saja

#### Root `turbo.json`

Contoh awal:

```json
{
  "$schema": "https://turbo.build/schema.json",
  "tasks": {
    "dev": {
      "cache": false,
      "persistent": true
    },
    "build": {
      "dependsOn": ["^build"],
      "outputs": [
        ".next/**",
        "dist/**",
        "build/**"
      ]
    },
    "lint": {
      "dependsOn": ["^lint"],
      "outputs": []
    },
    "test": {
      "dependsOn": ["^test"],
      "outputs": []
    },
    "typecheck": {
      "dependsOn": ["^typecheck"],
      "outputs": []
    }
  }
}
```

### Fase 2: Standarkan Script App

Setiap frontend sebaiknya memiliki script minimal:

- `dev`
- `build`
- `lint`
- `test` jika ada
- `typecheck`

Contoh tambahan yang perlu ada jika belum ada:

#### `web/package.json`

```json
{
  "scripts": {
    "typecheck": "tsc --noEmit"
  }
}
```

#### `web-lovable/package.json`

```json
{
  "scripts": {
    "typecheck": "tsc --noEmit -p tsconfig.app.json"
  }
}
```

### Fase 3: Pindahkan Frontend ke `apps/`

Setelah workspace root stabil:

```txt
web           -> apps/web
web-lovable   -> apps/web-lovable
```

Lalu ubah root `pnpm-workspace.yaml` menjadi:

```yaml
packages:
  - "apps/*"
  - "packages/*"
```

Hal yang perlu diperiksa setelah pemindahan:

- env path
- CI path
- Docker/dev script yang refer ke path lama
- asset/public path
- import alias
- README frontend

### Fase 4: Ekstrak `packages/ui`

Mulai dari komponen yang paling aman:

1. button
2. input
3. badge
4. dialog
5. card
6. table shell
7. page header
8. empty/loading state

Struktur awal:

```txt
packages/ui/
├── package.json
├── tsconfig.json
└── src/
    ├── button.tsx
    ├── input.tsx
    ├── badge.tsx
    ├── dialog.tsx
    ├── card.tsx
    └── index.ts
```

Contoh `packages/ui/package.json`:

```json
{
  "name": "@casbin/ui",
  "version": "0.0.0",
  "private": true,
  "main": "./src/index.ts",
  "types": "./src/index.ts",
  "peerDependencies": {
    "react": "^18 || ^19",
    "react-dom": "^18 || ^19"
  }
}
```

### Fase 5: Ekstrak `packages/api-types`

Isi awal:

- role DTO
- organization DTO
- project DTO
- user DTO
- paginated response schema

Jangan dulu pindahkan:

- fetch client
- axios client
- auth refresh logic
- tenant injection logic

### Fase 6: Rapikan CI

Contoh pembagian tanggung jawab:

- `turbo run lint build test` untuk frontend/package JS
- `make test` / `go test ./...` untuk Go
- `make test-integration` untuk integration backend
- `make test-e2e` untuk backend e2e

Contoh command root yang sehat:

```bash
pnpm lint
pnpm build
pnpm test
pnpm go:test
pnpm go:test-integration
```

## Blueprint Inisiasi

Urutan inisiasi yang saya sarankan:

1. Tambahkan `package.json` root
2. Tambahkan `pnpm-workspace.yaml` root
3. Tambahkan `turbo.json` root
4. Install `pnpm` dependency di root
5. Verifikasi `turbo run build` dan `turbo run lint`
6. Tambahkan `typecheck` di masing-masing frontend
7. Pindahkan frontend ke `apps/` hanya setelah langkah 1-6 stabil
8. Buat `packages/ui`
9. Port komponen presentational sedikit demi sedikit

## Apa Yang Tidak Perlu Dilakukan Dulu

- Memindahkan backend Go ke `apps/api`
- Membuat shared auth package
- Membuat shared realtime package
- Menyatukan seluruh API client dua frontend
- Menyatukan semua store
- Memaksa `web-lovable` menjadi frontend utama

## Risiko Utama

### 1. Shared package terlalu cepat

Risiko:

- komponen umum jadi tahu auth/tenant
- boundary rusak

Mitigasi:

- hanya ekstrak komponen presentational

### 2. Frontend path dipindah terlalu cepat

Risiko:

- script CI rusak
- env rusak

Mitigasi:

- workspace root dulu, pemindahan folder belakangan

### 3. Mencampur logic `web` dan `web-lovable`

Risiko:

- auth dan realtime model tercampur

Mitigasi:

- `web` tetap source of truth aplikasi utama
- `web-lovable` hanya sumber pola UI sampai integrasi diselaraskan

## Keputusan Yang Saya Rekomendasikan

- Gunakan `pnpm workspace + turbo`
- Pertahankan backend Go di root
- Jadikan `web` app utama
- Jadikan `web-lovable` app eksplorasi
- Ekstrak `packages/ui` dulu
- Ekstrak `packages/api-types` setelah itu
- Tunda shared auth/API/realtime sampai contract dua app benar-benar seragam

## Deliverable Implementasi Minimum

Minimal agar repo ini resmi masuk fase monorepo:

- root `package.json`
- root `pnpm-workspace.yaml`
- root `turbo.json`
- script `typecheck` di frontend
- command root untuk lint/build/test frontend

Itu sudah cukup untuk memulai tanpa membuat repo ini tidak stabil.
