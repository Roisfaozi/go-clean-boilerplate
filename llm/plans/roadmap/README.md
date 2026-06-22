# Roadmap Plans

## Purpose

Folder ini untuk plan multi-tahap yang lintas modul, lintas app, atau lintas milestone.

Isi roadmap harus tetap berguna walau implementasi tersebar di banyak sesi dan tidak selesai dalam satu branch kecil.

## Kapan Pakai Folder Ini

Gunakan `llm/plans/roadmap/` bila:

- target state perlu rollout bertahap
- ada dependency antar fase atau antar tim/agent
- compatibility concern lebih besar daripada satu patch rutin
- success criteria perlu didefinisikan per milestone

Contoh repo ini:

- penyelarasan kontrak API lintas backend, `apps/web`, `apps/client`, `packages/api-types`
- stabilisasi worker + webhook + audit + retry semantics
- observability rollout untuk tracing, metrics, pprof, and alertability
- evolusi design system frontend atau shared package boundary

## Struktur Minimal Dokumen

Roadmap sebaiknya memuat:

- current state summary
- target state
- rollout phases / milestones
- dependency antar fase
- migration or compatibility risk
- verification and rollback thinking
- explicit out-of-scope items

## Bedakan Dari Folder Lain

- `llm/plans/improve/`
  - improvement bertahap area existing, biasanya lebih sempit
- `llm/tasks/`
  - state aktif sekarang
- `llm/research/`
  - evidence dan comparison notes
- `llm/recommendations/`
  - ide perbaikan yang belum jadi roadmap

## Jangan Simpan di Sini

- patch satu module kecil
- audit sementara tanpa target state
- backlog ide tanpa sequence nyata

## Kualitas Roadmap Yang Baik

Roadmap bagus di repo ini harus:

- menyebut live owners seperti `internal/config/app.go`, `internal/router/router.go`, `apps/web`, `apps/client`, `packages/*`
- membedakan fakta sekarang vs target masa depan
- punya milestone yang bisa diverifikasi, bukan slogan umum
