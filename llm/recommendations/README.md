# Recommendations

## Purpose

Folder ini untuk peluang perbaikan non-urgent yang sudah punya evidence, tetapi belum menjadi plan eksekusi aktif.

## Simpan di Sini Bila

- ada debt teknis di luar scope patch sekarang
- ada gap hardening, performance, doc, test, atau architecture yang layak ditindaklanjuti nanti
- ada saran refactor bertahap yang belum mendapat approval atau slot implementasi

## Format Minimum

Setiap recommendation sebaiknya punya:

- masalah
- dampak
- evidence path atau command
- usulan perbaikan
- prioritas, risiko, atau urutan tindak lanjut

## Perbedaan Dengan Folder Lain

- `llm/cache/`
  - fakta stabil yang sudah tervalidasi
- `llm/plans/`
  - plan implementasi yang sudah punya sequence kerja
- `llm/tasks/`
  - state aktif sekarang
- `llm/research/`
  - investigasi dan evidence mapping yang mungkin belum berubah menjadi recommendation

## Jangan Simpan di Sini

- fakta runtime stabil
- plan eksekusi aktif
- audit sementara tanpa rekomendasi jelas
