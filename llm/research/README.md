# Research

## Purpose

Folder ini untuk investigasi durable berbasis evidence.

Gunakan saat perlu memahami behavior lintas file atau lintas module, membandingkan pendekatan, atau menyimpan audit teknis yang masih berguna setelah task sekarang selesai.

## Kapan Simpan di Sini

Simpan research note bila:

- perlu membandingkan beberapa pendekatan teknis atau product tradeoff
- perlu mengumpulkan evidence lintas file/module sebelum menyimpulkan
- temuan belum cukup stabil untuk masuk `llm/cache/`
- temuan belum cukup final untuk langsung jadi `llm/plans/` atau `llm/recommendations/`

Contoh repo ini:

- audit wiring auth/tenant/Casbin di banyak middleware dan usecase
- pembandingan owner route antara `apps/web`, `apps/client`, dan backend proxies
- evaluasi upload/storage flow end-to-end
- investigasi async path worker/audit/webhook yang belum cukup stabil dijadikan cache fact

## Struktur Minimal Dokumen

Research note sebaiknya memisahkan jelas:

- pertanyaan riset
- source/evidence path
- fakta terverifikasi
- inference
- `needs confirmation`
- recommendation terpisah jika ada

## Repo-Specific Rules

- live code tetap sumber utama
- jika route/auth/tenant behavior dibahas, verifikasi ke `internal/router/router.go`, middleware, dan target usecase
- jika frontend boundary dibahas, verifikasi owner app plus proxy files
- jangan promosikan ke `llm/cache/` sampai cukup stabil dan tervalidasi

## Bedakan Dari Folder Lain

- `llm/cache/`
  - hanya fakta stabil
- `llm/recommendations/`
  - usulan tindak lanjut non-urgent
- `llm/plans/`
  - sequence implementasi
- `llm/tasks/`
  - scratch notes dan active state

## Jangan Simpan di Sini

- checklist kerja aktif
- fakta stabil yang sudah siap jadi cache
- rekomendasi tanpa evidence
- dump command tanpa analisis
