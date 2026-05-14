# Architecture and Security Audit

Dokumen ini memuat temuan utama tentang kelemahan arsitektur, security risk, dan area yang perlu diperhatikan lebih tajam. Fokusnya adalah behavior aktual codebase, bukan checklist generik.

## Ringkasan Prioritas

Temuan paling penting adalah:

1. tenant isolation sangat bergantung pada disiplin middleware dan repository scope
2. authorization bisa dibypass jika Casbin dimatikan atau context domain tidak terbentuk
3. API key authentication belum menunjukkan enforcement scope yang kuat
4. ada perbedaan tingkat kematangan antar modul; core security matang, tetapi beberapa domain masih tipis

## Saran Prioritas Eksekusi

Urutan tindakan yang disarankan untuk tim:

1. kunci boundary tenant dan authorization
2. tutup gap API key scope enforcement
3. perjelas failure mode Redis dan degrade mode
4. rapikan peripheral domains agar tidak memberi kesan production-ready yang berlebihan

Target konkret untuk iterasi pertama:

- audit semua route tenant dan repository tenant-aware
- fail fast untuk config production yang menjalankan route protected tanpa Casbin
- definisikan model scope API key dan middleware enforcement-nya
- dokumentasikan blast radius ketika Redis down

Target konkret untuk iterasi kedua:

- tambah test lintas tenant dan test auth downgrade
- tambah guardrails coding pattern untuk repository tenant-aware
- tambah review terhadap webhook governance dan shadow-user invitation flow
- tandai fitur stats yang masih placeholder

## Temuan Arsitektur

### 1. Tenant scope bersifat opt-in secara repository

Fakta teknis:

- tenant context diinject oleh `TenantMiddleware`
- repository tertentu memanggil `database.OrganizationScope(ctx)`
- scope akan diam jika `organization_id` tidak ada di context

Implikasi:

- jika route tenant lupa memasang middleware, query bisa berjalan tanpa filter tenant
- jika repository baru lupa memakai scope, data leakage lintas tenant dapat terjadi

Severity:

- tinggi

Kenapa penting:

- model keamanan tenant di codebase ini bukan enforced globally oleh ORM layer
- ia bergantung pada kedisiplinan implementasi

Saran:

- buat daftar tabel yang wajib tenant-scoped dan validasikan satu per satu di repository
- tambahkan test yang membuktikan user Org A tidak bisa membaca data Org B untuk setiap modul tenant
- pertimbangkan helper repository atau wrapper query yang membuat scope tenant menjadi default, bukan opsional
- dokumentasikan route mana yang memang boleh bypass tenant scope agar bypass yang terjadi bersifat eksplisit

### 2. Casbin authorization adalah toggle, bukan invariant

Fakta teknis:

- middleware Casbin akan bypass saat enforcer `nil`
- config memungkinkan Casbin dimatikan

Implikasi:

- salah konfigurasi environment dapat menurunkan posture authorization secara drastis
- route yang diasumsikan protected bisa menjadi hanya authenticated

Severity:

- tinggi

Catatan:

- ini bisa masuk akal untuk local development
- tetapi berbahaya jika tidak diberi guard yang lebih keras untuk production

Saran:

- untuk environment production, ubah perilaku dari bypass menjadi startup failure jika enforcer tidak tersedia
- tambahkan health or startup assertion yang memeriksa Casbin benar-benar aktif saat route authorized didaftarkan
- bedakan dengan jelas konfigurasi dev, test, dan prod di dokumentasi dan `.env.example`

### 3. Domain authorization tergantung context organization

Fakta teknis:

- domain default adalah `global`
- Casbin middleware memakai `organization_id` dari context jika ada

Implikasi:

- jika tenant context gagal dibangun, policy evaluation bisa jatuh ke `global`
- hasilnya bisa menjadi deny atau, pada desain policy tertentu, menjadi allow yang tidak diinginkan

Severity:

- tinggi

Saran:

- pada route yang seharusnya tenant-only, tambahkan validasi eksplisit bahwa `organization_id` harus ada sebelum Casbin dievaluasi
- hindari fallback diam ke domain `global` untuk endpoint yang semestinya tenant-bound
- kelompokkan route menjadi global-only dan tenant-only, lalu dokumentasikan perilaku domain masing-masing

### 4. Mixed maturity antar domain

Modul yang matang:

- auth
- organization
- permission
- audit

Modul yang lebih sederhana:

- project
- stats

Implikasi:

- sistem terlihat kaya fitur, tetapi tidak semua domain punya kedalaman business logic yang sama
- arsitektur inti siap production lebih dulu daripada domain produk sekundernya

Severity:

- sedang

### 5. Shared infrastructure multiplexing on Redis

Redis dipakai untuk:

- session
- lockout
- worker
- presence
- ticket
- membership cache
- API key cache
- optional limiter

Implikasi:

- Redis menjadi single operational hot spot
- gangguan Redis berdampak ke auth, realtime, queue, dan tenant performance sekaligus

Severity:

- sedang sampai tinggi, tergantung deployment

Saran:

- pecah monitoring Redis berdasarkan fungsi: auth, queue, cache, presence
- dokumentasikan layanan mana yang gagal total dan mana yang hanya degraded saat Redis unavailable
- jika deployment menuntut availability tinggi, siapkan strategi HA atau operational runbook khusus Redis

## Temuan Security

### 1. API key scopes belum terlihat enforced di request path

Fakta teknis:

- API key identity memuat `scopes`
- middleware API key hanya menyuntik identity ke context
- tidak terlihat middleware atau authorization path yang mengevaluasi scopes tersebut

Implikasi:

- scope saat ini lebih terlihat sebagai metadata daripada hard authorization boundary
- risiko privilege lebih luas dari yang dibayangkan oleh pembuat API key

Severity:

- tinggi

Saran:

- tentukan apakah scope API key akan memetakan endpoint, action, resource, atau access-right
- enforce scope di middleware sebelum request masuk ke controller
- bila ingin satu model otorisasi, integrasikan scope API key ke Casbin; bila tidak, buat aturan evaluasi terpisah yang tegas
- tambahkan test bahwa API key tanpa scope relevan tidak bisa mengakses route tertentu

### 2. Auth posture kuat, tetapi sangat tergantung Redis availability

Kekuatan:

- JWT divalidasi
- session diverifikasi di Redis
- revoke session benar-benar efektif

Risiko:

- jika Redis terganggu, validasi session ikut terganggu
- pada outage tertentu, availability auth akan turun tajam

Severity:

- sedang

Saran:

- definisikan posture saat Redis gagal: fail closed, fail open, atau maintenance mode
- buat runbook operasional untuk pemulihan auth, worker, dan presence
- pastikan monitoring membedakan error Redis pada auth dengan error Redis pada cache non-kritis

### 3. Lockout logic baik, tetapi memusatkan kontrol di username key

Fakta teknis:

- login attempts dan lockout terlihat berbasis username

Implikasi:

- username tertentu bisa menjadi target denial-style lockout jika proteksi tambahan tidak ada
- perlu pengawasan pada kombinasi IP limit dan username lockout

Severity:

- sedang

Saran:

- kombinasikan lockout per username dengan rate limit berbasis IP atau fingerprint lain
- buat alert untuk pola lockout berulang terhadap username tertentu
- evaluasi apakah threshold lockout sudah seimbang antara security dan abuse resistance

### 4. Shadow user invite flow perlu pengawasan ekstra

Kekuatan:

- memudahkan onboarding kolaboratif

Risiko:

- identity placeholder diciptakan sebelum user menyelesaikan registrasi penuh
- validasi merge account dan claim invitation harus sangat jelas agar tidak menimbulkan edge case

Severity:

- sedang

Saran:

- dokumentasikan lifecycle shadow user dari invite sampai aktivasi penuh
- pastikan ada aturan yang jelas untuk merge akun jika email yang diundang kemudian register sendiri
- tambahkan test untuk duplicate invite, expired invite, dan claim invite oleh akun yang berbeda

### 5. Webhook outbound adalah jalur exfiltration yang sah

Fakta teknis:

- webhook dapat mengirim payload event ke URL eksternal

Implikasi:

- bila governance lemah, webhook bisa menjadi jalur keluarnya data organisasi
- perlu pembatasan event, audit webhook change, dan ideally signature verification ketat

Severity:

- sedang

Saran:

- audit semua perubahan webhook configuration
- batasi event yang boleh diekspos ke webhook per organization bila perlu
- pertimbangkan allowlist domain atau validasi tambahan untuk deployment yang sensitif
- pastikan setiap delivery memakai signature yang konsisten dan mudah diverifikasi client

### 6. Realtime channel perlu governance channel naming dan authorization yang disiplin

Risiko utama:

- jika naming channel atau broadcast rule longgar, data realtime bisa menyebar ke audience yang salah
- route WS sudah memakai ticket, tetapi authorization pasca-connect harus tetap disiplin

Severity:

- sedang

Saran:

- dokumentasikan channel yang bersifat tenant-scoped dan global-scoped
- audit subscription and broadcast rules agar tidak ada channel yang menerima audience terlalu luas
- pastikan reconnect flow tidak melonggarkan authorization yang diterapkan saat connect awal

## Kekuatan Arsitektur

Hal-hal yang justru kuat di codebase ini:

- clean architecture modular cukup konsisten
- transaction manager jelas
- transactional enforcer adalah keputusan desain yang matang
- audit outbox pattern mengurangi inconsistency
- session-backed JWT memberi revoke capability nyata
- tenant membership cache mengurangi pressure ke DB
- async side effects dipisah dari jalur request utama

## Area yang Perlu Hardening

### Guardrails tenant isolation

Rekomendasi:

- buat checklist wajib untuk repository tenant-aware
- tambahkan test lintas tenant pada semua modul tenant
- pertimbangkan pola base repository atau helper untuk mendorong scope by default

### Guardrails Casbin in production

Rekomendasi:

- fail fast di production bila route protected berjalan tanpa enforcer
- bedakan mode dev dan prod secara eksplisit

### API key scope enforcement

Rekomendasi:

- definisikan scope model yang benar-benar digunakan
- tambahkan middleware enforcement sebelum controller
- selaraskan dengan Casbin atau kebijakan terpisah

### Redis dependency posture

Rekomendasi:

- dokumentasikan blast radius Redis failure
- tetapkan fallback atau degrade mode yang jelas
- perjelas SLA untuk auth and worker availability

### Stats and peripheral domains

Rekomendasi:

- tandai fitur yang masih placeholder
- jangan posisikan stats sebagai source of truth observability sebelum didukung metrics nyata

## Checklist Implementasi

Checklist ini disusun agar tim bisa mengubah audit menjadi pekerjaan yang dapat dieksekusi.

### Minggu 1

- inventaris semua route tenant-only
- inventaris semua repository yang menyentuh tabel bertenant
- tetapkan keputusan production untuk Casbin: wajib aktif atau tidak
- definisikan posture Redis outage untuk auth dan worker

### Minggu 2

- tambahkan test lintas tenant untuk modul `organization`, `project`, `audit`, dan modul tenant-aware lain
- tambahkan test untuk memastikan route protected gagal jika Casbin tidak aktif di mode production
- tetapkan spesifikasi scope API key dan route enforcement target

### Minggu 3

- implementasikan enforcement scope API key
- tambah audit trail untuk perubahan webhook dan API key
- rapikan dokumentasi stats agar jelas mana yang placeholder dan mana yang berbasis data nyata

## Pertanyaan Terbuka untuk Tim

Beberapa keputusan perlu disepakati, bukan hanya diimplementasikan:

- apakah seluruh route authorized harus gagal startup jika Casbin mati
- apakah semua resource tenant wajib memiliki repository abstraction yang scope-aware by default
- apakah API key akan menggunakan model scope sendiri atau mengikuti model permission Casbin
- apakah shadow user boleh tetap dipertahankan, atau harus diganti dengan invitation pre-registration tanpa membuat user record dulu
- seberapa besar toleransi terhadap outage Redis untuk jalur auth

## Residual Risk by Area

### Sangat kritikal

- tenant leakage akibat middleware atau scope yang hilang
- authorization downgrade akibat Casbin disabled
- API key over-privilege karena scope tidak enforced

### Kritis menengah

- Redis outage memukul banyak subsistem sekaligus
- shadow user edge cases pada invitation flow
- webhook misuse sebagai outbound data channel

### Menengah

- ketidakseimbangan kematangan antar modul
- stats yang masih semi-placeholder

## Kesimpulan

Secara umum, fondasi security dan architecture aplikasi ini cukup kuat untuk ukuran boilerplate enterprise. Namun model keamanan utamanya masih bergantung pada tiga disiplin implementasi:

- tenant middleware harus selalu benar dipasang
- repository tenant-aware harus selalu benar memakai scope
- Casbin dan context domain harus selalu tersedia dan sinkron

Jika tiga hal itu dijaga, core platform ini kuat. Jika tidak, risiko paling besar bukan bug kecil, tetapi bypass boundary sistem.
