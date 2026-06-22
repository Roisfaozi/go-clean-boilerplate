# Plans

## Purpose

Folder ini untuk rencana implementasi yang cukup besar, staged, atau durable sehingga tidak layak disimpan hanya di `llm/tasks/todo.md`.

Gunakan folder ini ketika agent atau manusia butuh dokumen eksekusi yang tetap berguna setelah satu task selesai.

## Kapan Simpan di Sini

Simpan plan di `llm/plans/` bila:

- pekerjaan terdiri dari beberapa fase yang saling bergantung
- plan perlu direferensikan ulang lintas sesi
- ada approval gate, dependency map, atau rollout order yang penting dipertahankan
- item terlalu besar untuk sekadar checklist aktif di `llm/tasks/todo.md`

## Struktur Folder

- `llm/plans/improve/`
  - untuk improvement plan bertahap pada area yang sudah ada
  - contoh: hardening auth, cleanup frontend proxy, peningkatan test coverage
- `llm/plans/roadmap/`
  - untuk roadmap multi-tahap lintas modul, app, atau milestone
  - contoh: stabilisasi worker, penyelarasan kontrak API lintas app, observability rollout

## Format Minimum Plan

Setiap plan sebaiknya minimal punya:

- tujuan
- scope dan non-goals
- dependency / urutan fase
- file atau area ownership
- verification plan per fase
- blocker atau approval gate

## Perbedaan Dengan Folder Lain

- `llm/tasks/`
  - untuk state kerja aktif, audit sementara, parity notes, dan checklist harian
- `llm/cache/`
  - untuk fakta stabil repo yang sudah tervalidasi dari live code atau committed docs
- `llm/recommendations/`
  - untuk ide perbaikan non-urgent yang belum menjadi plan eksekusi
- `llm/research/`
  - untuk investigasi berbasis evidence yang memisahkan fakta dan rekomendasi

## Jangan Simpan di Sini

- fakta runtime stabil tanpa eksekusi plan
- catatan task sementara yang akan cepat basi
- rekomendasi tanpa sequence implementasi yang jelas

## Penamaan Yang Disarankan

- gunakan nama yang menjelaskan area dan intent
- contoh:
  - `auth-session-hardening.md`
  - `frontend-proxy-contract-alignment.md`
  - `worker-webhook-stability-roadmap.md`
