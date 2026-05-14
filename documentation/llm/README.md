# LLM Analysis Index

Folder ini berisi paket analisis lintas-sistem untuk membantu memahami arsitektur, request flow, proses bisnis, dan area risiko di codebase ini. Dokumen-dokumen di sini melengkapi dokumentasi utama di `documentation/`, bukan menggantikannya sebagai source of truth.

Dokumen ini paling berguna untuk:

- engineer baru yang perlu onboarding cepat ke struktur backend
- reviewer yang ingin menilai alur auth, tenant, RBAC, API key, dan worker
- maintainer yang ingin memahami dampak hardening atau perubahan arsitektur
- agent atau tool analisis yang butuh entry point ke dokumen-dokumen turunan

## Scope

Fokus folder ini adalah analisis sistem. Untuk detail implementasi dan referensi utama, tetap rujuk ke:

- [`../../README.md`](../../README.md)
- [`../ARCHITECTURE.md`](../ARCHITECTURE.md)
- [`../guides/TESTING.md`](../guides/TESTING.md)
- source code di `internal/`, `pkg/`, `cmd/`, dan `tests/`

## Status

- Status keseluruhan: `active analysis`
- Dokumen di folder ini adalah dokumen analitis, bukan spesifikasi final
- Beberapa file dapat bersifat parsial dan perlu divalidasi ulang terhadap codebase saat ini sebelum dipakai sebagai dasar perubahan besar

## Urutan Baca

Urutan baca yang direkomendasikan:

1. [00-analysis-priority.md](./00-analysis-priority.md)
2. [04-architecture-security-audit.md](./04-architecture-security-audit.md)
3. [01-module-map.md](./01-module-map.md)
4. [02-request-flows.md](./02-request-flows.md)
5. [03-business-processes.md](./03-business-processes.md)
6. [05-api-key-analysis.md](./05-api-key-analysis.md)
7. [06-hardening-impact-analysis.md](./06-hardening-impact-analysis.md)

Jika tujuan Anda spesifik, gunakan jalur baca berikut:

- Memahami arsitektur dan boundary modul: `01` lalu `02`
- Menilai risiko security dan authorization: `04` lalu `06`
- Memahami proses bisnis utama: `03`
- Meninjau area API key: `05` lalu `06`
- Menentukan prioritas pembacaan: `00`

## Ringkasan Dokumen

- `00-analysis-priority.md`: menjawab analisis mana yang perlu dilakukan lebih dulu dan alasannya
- `01-module-map.md`: memetakan modul utama, tanggung jawab, dan relasi antar area sistem
- `02-request-flows.md`: menjelaskan flow end-to-end per request dari middleware sampai persistence
- `03-business-processes.md`: menguraikan proses bisnis utama per fitur dan interaksi antarmodul
- `04-architecture-security-audit.md`: audit kelemahan arsitektur, security risk, dan titik enforcement
- `05-api-key-analysis.md`: analisis mendalam alur, enforcement, dan risiko pada fitur API key
- `06-hardening-impact-analysis.md`: menilai dampak teknis dan operasional jika hardening diterapkan

## Cara Memakai Folder Ini

- Gunakan file di folder ini untuk membangun peta mental sistem sebelum mengubah code path yang sensitif
- Validasi temuan penting dengan membaca file implementasi terkait sebelum mengambil keputusan desain
- Perlakukan dokumen ini sebagai panduan analisis kerja, bukan pengganti review source code
