# Test Playbooks

Gunakan folder ini untuk flow test manual, browser, API, integration, atau E2E yang reusable.

## Simpan di sini bila

- flow perlu diulang lintas task
- login / seed / test data setup penting
- ada langkah verifikasi UI/API end-to-end yang spesifik repo

## Isi minimum

- scope flow
- prerequisite
- langkah test
- expected result
- cleanup bila perlu
- command terkait bila ada

## Catatan repo

- backend integration dan E2E biasa butuh Docker
- `apps/client` lint bukan verifikasi kuat
- route/auth/tenant test harus tetap diverifikasi terhadap `internal/router/router.go` dan middleware terkait
