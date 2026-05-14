# Hardening Impact Analysis

Dokumen ini menganalisis dampak yang akan terjadi jika seluruh audit dan saran hardening dari paket analisis sebelumnya benar-benar diimplementasikan. Fokusnya adalah impact ke core logic aplikasi, test suite, data contracts, operasional, dan cara mengatasinya.

## Ringkasan

Jika seluruh hardening dijalankan, sistem akan berubah dari:

- backend enterprise-capable yang masih memiliki beberapa fallback longgar

menjadi:

- platform yang lebih strict terhadap tenant boundary, authorization, API key restriction, dan operational failure mode

Perubahan ini akan meningkatkan security posture, tetapi hampir pasti menimbulkan:

- perubahan kontrak request
- pecahnya sebagian test lama
- kebutuhan refactor di middleware, repository, dan route classification
- tambahan policy migration
- tambahan beban operasional pada observability dan worker path

## Area Impact Utama

Impact utama akan muncul di enam area:

1. request lifecycle dan route contract
2. authorization core logic
3. tenant isolation dan repository access pattern
4. API key model dan scope enforcement
5. Redis dependency posture
6. test architecture dan compatibility

## 1. Impact ke Request Lifecycle dan Route Contract

Jika hardening diterapkan, middleware chain akan menjadi jauh lebih ketat.

### Perubahan yang akan terjadi

- route tenant yang dulu masih bisa lolos tanpa context yang benar akan mulai gagal
- route authorized yang dulu tetap jalan saat Casbin tidak aktif akan fail startup atau fail request
- request dengan API key valid tetapi scope tidak cocok akan ditolak
- route yang dulu fallback diam ke domain `global` akan perlu diklasifikasikan ulang

### Dampak ke aplikasi

- lebih banyak request akan berhenti di middleware sebelum business usecase dijalankan
- error `401`, `403`, atau `400` akan lebih sering muncul secara sengaja
- frontend atau client internal yang mengandalkan fallback lama akan pecah

### Dampak ke core logic

- sebagian controller tidak lagi menerima request yang dulu lolos
- usecase akan lebih jarang memproses request invalid karena ditahan lebih awal
- validasi organisasi dan authorization tidak lagi implicit

### Cara mengatasinya

- buat klasifikasi route yang eksplisit:
  - `global-only`
  - `tenant-required`
  - `tenant-optional`
- buat matriks route contract per endpoint:
  - auth requirement
  - tenant requirement
  - authorization requirement
  - accepted auth methods
- lakukan rollout bertahap bila client existing masih memakai contract lama

## 2. Impact ke Authorization Core Logic

Jika Casbin dibuat mandatory di production dan domain global versus tenant dipertegas, maka authorization bukan lagi best-effort control, tetapi menjadi hard contract.

### Perubahan yang akan terjadi

- role dan permission yang ada mungkin tidak cukup untuk semua route setelah klasifikasi diperketat
- policy yang sebelumnya “kebetulan bekerja” karena fallback domain bisa gagal
- route tenant-only akan berhenti memakai fallback `global`

### Dampak ke modul inti

Modul yang paling terdampak:

- [casbin_middleware.go](/home/user/Documents/Riset/Casbin/internal/middleware/casbin_middleware.go)
- [tenant_middleware.go](/home/user/Documents/Riset/Casbin/internal/middleware/tenant_middleware.go)
- [permission_usecase.go](/home/user/Documents/Riset/Casbin/internal/modules/permission/usecase/permission_usecase.go)
- [app.go](/home/user/Documents/Riset/Casbin/internal/config/app.go)

### Risiko yang akan muncul

- policy DB perlu dibersihkan atau di-seed ulang
- role tertentu tiba-tiba kehilangan akses yang sebelumnya dianggap normal
- bug domain mismatch akan menjadi lebih terlihat

### Cara mengatasinya

- buat inventory semua route protected dan domain yang seharusnya dipakai
- audit semua policy existing di Casbin
- tambahkan seed atau migration untuk policy yang wajib ada
- kelompokkan route authorized menjadi:
  - global-authorized
  - tenant-authorized

## 3. Impact ke Tenant Isolation dan Repository Pattern

Jika tenant hardening diterapkan penuh, repository pattern akan berubah cukup dalam.

### Perubahan yang akan terjadi

- repository tenant-aware yang belum memakai scope akan perlu diubah
- query global admin yang tadinya bergantung pada “tidak ada context” harus dibuat eksplisit
- bug data leakage yang sebelumnya diam akan muncul saat route mulai menolak atau data tampak “hilang”

### Dampak ke core architecture

- base helper query kemungkinan perlu ditambahkan
- coding standard repository akan menjadi lebih ketat
- developer harus sadar tabel mana yang tenant-bound dan mana yang global

### Risiko teknis

- beberapa usecase existing mungkin diam-diam membaca data lintas tenant
- patch hardening yang terburu-buru bisa malah membuat query legitimate global menjadi gagal

### Cara mengatasinya

- buat daftar resmi:
  - tabel tenant-bound
  - tabel global
  - tabel campuran atau butuh special handling
- buat helper eksplisit seperti:
  - `tenantDB(ctx)`
  - `globalDB(ctx)`
- hindari bypass dengan cara “jangan isi context”; bypass harus bernama jelas dan terdokumentasi
- tambahkan negative test lintas tenant di setiap modul tenant-aware

## 4. Impact ke Fitur API Key

Jika semua saran API key diterapkan, impact-nya akan sangat besar terhadap perilaku client dan authorization flow.

### Perubahan yang akan terjadi

- API key tidak lagi hanya memvalidasi identity, tetapi juga membatasi capability
- client yang selama ini memakai API key tanpa scope enforcement akan mulai gagal
- middleware baru atau jalur authorization baru akan diperlukan

### Dampak ke desain

Sistem harus memilih model scope:

1. scope ke endpoint dan method
2. scope ke resource dan action
3. scope ke access-right
4. scope diintegrasikan ke Casbin

### Dampak ke modul

Modul yang terdampak:

- [api_key_usecase.go](/home/user/Documents/Riset/Casbin/internal/modules/api_key/usecase/api_key_usecase.go)
- [api_key_middleware.go](/home/user/Documents/Riset/Casbin/internal/middleware/api_key_middleware.go)
- kemungkinan [permission_usecase.go](/home/user/Documents/Riset/Casbin/internal/modules/permission/usecase/permission_usecase.go) jika disatukan dengan model permission existing

### Risiko

- existing key dengan `scopes=[]` akan ambigu
- precedence JWT versus API key bisa membingungkan
- route tertentu bisa over-protected atau under-protected jika model scope tidak konsisten

### Cara mengatasinya

- pilih satu model scope saja
- sarankan model yang paling konsisten dengan codebase:
  - `scope -> access-right`
  - atau `scope -> resource/action`
- lakukan rollout dalam dua fase:
  - observability mode
  - enforcement mode
- tentukan behavior eksplisit untuk key lama tanpa scope

## 5. Impact ke Audit, Webhook, dan Async Side Effects

Jika hardening menambah audit trail, webhook governance, dan kontrol lifecycle lebih ketat, jumlah side effect akan bertambah.

### Perubahan yang akan terjadi

- event audit bertambah
- task worker bertambah
- kemungkinan volume audit log meningkat tajam
- beberapa aksi admin yang dulu sunyi kini menjadi audited

### Dampak operasional

- beban worker naik
- volume `audit_logs` naik
- queue Redis bisa bertambah sibuk
- export audit dan retention strategy mungkin perlu disesuaikan

### Risiko

- duplicate event atau side effect jika idempotency belum kuat
- latensi async processing meningkat
- observability path sendiri menjadi beban baru

### Cara mengatasinya

- bedakan event yang wajib audit dari event yang cukup log biasa
- jaga event kritis tetap transactional jika memengaruhi compliance
- tambahkan idempotency atau deduplication untuk webhook dan audit tertentu
- ukur kapasitas worker dan review retention audit log

## 6. Impact ke Redis Dependency Posture

Jika saran tentang Redis dijalankan, tim harus memilih failure mode per fungsi, bukan satu kebijakan generik.

### Perubahan yang akan terjadi

- auth session verification bisa menjadi lebih strict
- worker, presence, ticket, dan cache akan diperlakukan dengan posture yang berbeda
- local development dan test mungkin berubah karena sebagian flow tidak lagi best-effort

### Risiko

- salah memilih posture bisa menukar security dengan availability secara tidak terkendali
- sistem bisa terlalu keras untuk cache non-kritis atau terlalu longgar untuk auth kritis

### Cara mengatasinya

- pecahkan concern Redis menjadi:
  - auth critical
  - worker critical
  - cache optional
  - presence optional
- definisikan per subsistem:
  - fail closed
  - fail open
  - degrade mode
- buat runbook pemulihan untuk auth, worker, dan presence

## 7. Impact ke Developer Experience dan Base App Convention

Hardening yang serius akan mengubah cara tim menulis kode.

### Perubahan yang akan terjadi

- menambah modul baru menjadi lebih berat karena harus memikirkan tenant, authz, audit, dan side effect sejak awal
- review PR harus memeriksa contract security, bukan hanya business logic
- developer tidak bisa lagi mengandalkan fallback diam

### Dampak ke base app

- template module generator mungkin perlu diubah
- checklist engineering perlu diperketat
- helper coding untuk tenant-aware repository dan route classification menjadi penting

### Cara mengatasinya

- buat checklist engineering untuk modul baru
- update generator atau template modul agar memasukkan concern tenant/authz dari awal
- buat panduan “cara menambah tenant-aware module dengan aman”
- sediakan helper test dan helper repository agar guardrails tidak terasa seperti beban berlebihan

## 8. Impact ke Test Suite

Ini kemungkinan area yang paling terasa.

Jika hardening diimplementasikan, banyak test lama akan gagal bukan karena bug baru, tetapi karena kontrak keamanan berubah.

### Perubahan yang akan terjadi

- response yang dulu `200` bisa menjadi `401`, `403`, atau `400`
- test tenant route perlu selalu membawa org context yang benar
- test authorized route perlu seed policy yang benar
- test API key perlu mencakup scope allow dan deny
- test Redis failure perlu ditambahkan

### Kategori test yang wajib ditambah

- cross-tenant denial tests
- Casbin mandatory-in-prod tests
- domain mismatch tests
- API key scope allow and deny tests
- Redis failure behavior tests
- webhook governance tests
- invitation edge-case tests

### Cara mengatasinya

- buat matriks perubahan kontrak route sebelum mengubah test
- update integration dan e2e test lebih dulu
- setelah itu baru sesuaikan unit test usecase dan mocks
- buat test helpers untuk:
  - user bertenant
  - tenant header
  - Casbin policy seed
  - API key dengan scope tertentu
  - Redis failure simulation atau cache miss path

## 9. Impact ke Data dan Migration

Walau banyak hardening ada di level logic, beberapa perubahan hampir pasti menyentuh data layer.

### Perubahan yang mungkin terjadi

- policy Casbin perlu reseed atau migration tambahan
- data API key lama perlu aturan transisi
- event audit baru menambah volume dan variasi data
- field seperti `is_active` pada API key mungkin mulai benar-benar dipakai

### Risiko

- data existing tidak kompatibel dengan contract baru
- key lama atau policy lama berperilaku ambigu
- rollout code lebih cepat daripada readiness data

### Cara mengatasinya

- pisahkan migration data dari rollout code
- tentukan backward compatibility plan untuk API key lama
- buat verifier script untuk policy Casbin minimum
- review kebutuhan retention dan indexing saat audit event bertambah

## 10. Impact ke Backward Compatibility

Hardening yang benar hampir selalu menghasilkan breaking behavior.

### Contoh perubahan perilaku

- client lama tidak kirim `X-Organization-ID`, kini gagal
- API key valid tapi tanpa scope yang cukup kini gagal
- environment yang dulu jalan tanpa Casbin kini gagal startup
- route tenant yang dulu “kebetulan global” kini jadi deny

### Cara mengatasinya

- buat compatibility matrix sebelum rollout
- umumkan breaking change ke consumer internal dan frontend
- gunakan rollout bertahap:
  - observe
  - warn
  - enforce
- jika perlu, sediakan temporary compatibility mode dengan batas waktu yang jelas

## Strategi Mitigasi yang Disarankan

### 1. Kunci kontrak sistem dulu

Definisikan per route:

- auth method
- tenant requirement
- authorization requirement
- expected side effects

### 2. Ubah test menjadi penjaga kontrak

Jangan hanya test business logic. Test juga:

- tenant boundary
- authorization boundary
- API key restriction
- failure mode Redis

### 3. Rollout bertahap

Jangan lakukan enforcement penuh dalam satu langkah.

Urutan yang disarankan:

1. observability and assertion
2. warning and compatibility review
3. hard enforcement

### 4. Pisahkan policy change dari behavior change

Casbin seed, API key scope model, route classification, dan middleware enforcement sebaiknya tidak dirilis dalam satu commit besar.

### 5. Invest di helper dan guardrails

Kalau tidak, maintenance cost akan naik dan developer akan mencari jalan pintas.

## Checklist Mitigasi

### Sebelum implementasi

- buat route contract matrix
- buat inventory tabel tenant-bound
- buat inventory route protected dan policy existing
- putuskan failure mode Redis per fungsi
- putuskan model scope API key

### Saat implementasi

- tambah test helper dan policy seed helper
- lakukan perubahan route classification per grup
- terapkan observability mode untuk perubahan yang berpotensi breaking
- siapkan migration data dan policy

### Setelah implementasi

- audit test failure yang muncul karena contract berubah
- ukur kenaikan audit volume dan worker load
- review false positive atau deny yang tidak diinginkan
- hapus compatibility mode sementara bila sudah aman

## Kesimpulan

Jika seluruh audit dan saran hardening dijalankan, impact utamanya adalah aplikasi berubah dari backend yang fleksibel menjadi platform yang lebih tegas terhadap boundary keamanan dan tenancy.

Itu adalah arah yang benar, tetapi efek sampingnya besar:

- lebih banyak request akan ditolak secara sengaja
- lebih banyak test lama akan gagal
- lebih banyak modul akan perlu refactor
- lebih banyak keputusan desain harus dibuat eksplisit

Cara mengatasinya bukan dengan mengurangi hardening, tetapi dengan:

- memperjelas kontrak sistem
- menjadikan test suite sebagai contract suite
- melakukan rollout bertahap
- memberi developer guardrails agar implementasi aman menjadi jalur paling mudah
