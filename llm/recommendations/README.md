# Recommendations

## Purpose

Folder ini untuk peluang perbaikan non-urgent yang sudah punya evidence, tetapi belum menjadi plan eksekusi aktif.

Isi folder ini harus cukup konkret untuk dipromosikan ke `llm/plans/` saat prioritas naik.

## Kapan Simpan di Sini

Simpan recommendation bila:

- ada debt teknis di luar scope patch saat ini
- ada gap hardening, performance, docs, tests, atau architecture yang layak ditindaklanjuti nanti
- ada refactor yang jelas manfaatnya tetapi belum ada approval/slot

Contoh repo ini:

- hardening tambahan untuk querybuilder deny-list
- penyelarasan kontrak frontend-backend yang belum urgent
- coverage gap pada webhook/audit integration paths
- cleanup shared package ownership antar frontend apps

## Struktur Minimal Dokumen

Setiap recommendation sebaiknya punya:

- masalah
- dampak
- evidence path / file / command
- usulan perbaikan
- prioritas atau urgency
- risiko jika dibiarkan

## Bedakan Dari Folder Lain

- `llm/cache/`
  - fakta stabil yang sudah tervalidasi
- `llm/research/`
  - investigasi evidence-heavy yang mungkin belum menghasilkan recommendation final
- `llm/plans/`
  - jalur eksekusi yang sudah lebih matang
- `llm/tasks/`
  - state kerja aktif sekarang

## Jangan Simpan di Sini

- fakta runtime stabil
- plan implementasi aktif
- audit sementara tanpa rekomendasi jelas
- opini tanpa evidence path
