# Research

## Purpose

Folder ini untuk investigasi durable berbasis evidence.

Gunakan ketika perlu memahami behavior lintas file atau lintas module, membandingkan beberapa pendekatan, atau menyimpan audit teknis yang masih berguna setelah task sekarang selesai.

## Simpan di Sini Bila

- perlu membandingkan beberapa pendekatan teknis atau produk
- perlu catatan evidence lintas file/module yang tidak cocok dimasukkan ke `llm/cache/`
- perlu memisahkan fakta, inference, dan recommendation dalam audit besar

## Format Minimum

Setiap research note sebaiknya berisi:

- pertanyaan riset
- file path, command, atau evidence source
- temuan fakta
- inference yang jelas dipisah dari fakta
- batasan atau `needs confirmation`
- recommendation terpisah dari fakta

## Perbedaan Dengan Folder Lain

- `llm/cache/`
  - hanya untuk fakta yang sudah cukup stabil dan tervalidasi
- `llm/tasks/`
  - state aktif sekarang
- `llm/recommendations/`
  - follow-up non-urgent
- `llm/plans/`
  - sequence implementasi

## Jangan Simpan di Sini

- checklist kerja aktif
- fakta stabil yang sudah layak dipromosikan ke `llm/cache/`
- rekomendasi tanpa evidence
