# Improve Plans

## Purpose

Folder ini untuk improvement plan bertahap pada area yang sudah ada di repo, bukan fitur baru dari nol.

Dokumen di sini harus cukup detail untuk dipakai ulang lintas sesi, lintas agent, atau lintas reviewer.

## Kapan Pakai Folder Ini

Simpan plan di `llm/plans/improve/` bila:

- ada area existing yang perlu diperbaiki bertahap
- pekerjaan terlalu besar untuk sekali patch
- urutan perbaikan penting agar blast radius terkendali
- ada debt teknis yang sudah jelas jalur eksekusinya

Contoh repo ini:

- hardening auth/session/tenant/Casbin
- cleanup frontend proxy dan shared type drift
- peningkatan querybuilder safety coverage
- stabilisasi upload, webhook, atau worker side effects
- perapihan `llm/` starter-pack agar parity makin dekat ke Carbon tetapi tetap repo-specific

## Struktur Minimal Dokumen

Setiap improve plan sebaiknya punya:

- tujuan dan pain saat ini
- scope dan non-goals
- urutan fase kecil
- ownership file/module per fase
- risk atau blast radius
- verification per fase
- blocker, dependency, atau approval gate

## Bedakan Dari Folder Lain

- `llm/tasks/`
  - state aktif cepat berubah
- `llm/research/`
  - investigasi dan evidence, belum tentu jadi eksekusi
- `llm/recommendations/`
  - usulan perbaikan non-urgent tanpa sequence implementasi matang
- `llm/plans/roadmap/`
  - plan lintas milestone atau lintas area yang lebih besar

## Format Yang Disarankan

Pola yang bagus:

1. current state
2. target state
3. phase 1 / phase 2 / phase 3
4. verification target
5. deferred risk or follow-up

## Jangan Simpan di Sini

- satu patch kecil yang bisa selesai sekarang
- checklist harian yang cepat basi
- recommendation tanpa evidence atau tanpa urutan kerja
