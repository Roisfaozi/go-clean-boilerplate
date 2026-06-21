# Test Playbooks

## Purpose

Folder ini untuk flow verifikasi reusable: manual UI, API, browser, integration, atau E2E.

Gunakan ketika langkah test cukup spesifik ke repo ini dan layak dipakai ulang lintas task.

## Kapan Simpan di Sini

Simpan playbook bila:

- login/setup actor perlu diulang berkali-kali
- tenant/org context penting untuk hasil test
- ada route/auth/proxy/worker behavior yang butuh langkah verifikasi stabil
- verifikasi manual lebih bernilai daripada sekadar command dump

Contoh kuat di repo ini:

- login + pilih org + cek route tenant-protected
- upload avatar via TUS lalu verifikasi profile update
- buat webhook lalu trigger event dan cek delivery log
- batch permission check vs admin permission CRUD
- frontend proxy auth/cookie behavior di `apps/web` dan `apps/client`

## Available Playbooks

- `llm/test-playbooks/security-boundary-regression-playbook.md`: quick checklist for security-sensitive regression planning.
- `llm/test-playbooks/security-boundary-change-types.md`: command-level playbooks by change type, including auth/session, tenant/Casbin/API key, upload/storage, worker/audit/webhook, realtime, and frontend contract.
- `llm/test-playbooks/benchmark-before-after.md`: before/after benchmark discipline for performance-sensitive work.

## Struktur Minimal Dokumen

Setiap playbook sebaiknya punya:

- scope flow
- prerequisite
- actor/role yang dipakai
- seed/data awal bila perlu
- langkah test
- expected result
- cleanup bila perlu
- command terkait bila ada

## Repo-Specific Notes

- backend integration dan E2E sering butuh Docker
- `apps/client` lint runs Biome; older logs before Phase 7 may still show placeholder-only lint
- auth/tenant/Casbin behavior harus dicek ke `internal/router/router.go` dan middleware terkait
- proxy behavior bisa perlu dicek di `apps/web/src/app/api/v1/[...path]/route.ts` atau `apps/client/app/routes/api-proxy.ts`
- worker/webhook/audit behavior kadang butuh eventual-consistency wait step

## Bedakan Dari Folder Lain

- `llm/tasks/`
  - result sementara atau scratch notes
- `llm/research/`
  - analisis dan evidence, bukan langkah test operasional
- `llm/cache/`
  - fakta stabil, bukan prosedur test

## Jangan Simpan di Sini

- hasil test ad hoc tanpa flow reusable
- command list tanpa actor, context, dan expected result
- checklist terlalu umum seperti “run app and click around”
