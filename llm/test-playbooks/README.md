# Test Playbooks

## Purpose

Folder ini untuk flow test manual, browser, API, integration, atau E2E yang reusable.

Gunakan ketika langkah verifikasi sering diulang dan cukup spesifik ke repo ini sehingga layak dijadikan playbook.

## Simpan di Sini Bila

- flow perlu diulang lintas task atau lintas agent
- login, seed, fixture, atau tenant setup penting untuk verifikasi
- ada langkah verifikasi UI/API end-to-end yang spesifik ke route, proxy, auth, atau worker behavior repo ini

## Format Minimum

Setiap playbook sebaiknya punya:

- scope flow
- prerequisite
- actor atau role yang dipakai
- langkah test
- expected result
- cleanup bila perlu
- command terkait bila ada

## Catatan Repo Ini

- backend integration dan E2E sering butuh Docker
- `apps/client` lint bukan verifikasi kuat
- route, auth, tenant, dan Casbin behavior harus tetap dicek terhadap `internal/router/router.go` dan middleware terkait
- frontend proxy behavior bisa perlu pengecekan di `apps/web/src/app/api/v1/[...path]/route.ts` atau `apps/client/app/routes/api-proxy.ts`

## Jangan Simpan di Sini

- test result sementara tanpa flow reusable
- ad hoc command dump tanpa konteks actor dan expected result
