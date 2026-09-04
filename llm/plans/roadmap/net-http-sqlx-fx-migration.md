# Migration Plan: stdlib `net/http`, `sqlx`, Raw SQL, and Fx

## 1. Tujuan

Migrasikan backend Go dari Gin + GORM ke:

- `net/http` dan `http.ServeMux` Go 1.25 untuk routing dan middleware;
- `github.com/jmoiron/sqlx` di atas `database/sql` untuk koneksi, transaksi, dan scanning;
- raw SQL terparameterisasi untuk seluruh operasi repository dan query dynamic;
- `go.uber.org/fx` untuk dependency injection dan lifecycle.

Perilaku API, auth/session, tenant isolation, API-key scope, Casbin, audit, webhook, worker, upload TUS, SSE, WebSocket, dan frontend proxy harus tetap kompatibel kecuali perubahan kontrak disetujui eksplisit.

## 2. Scope dan non-goals

### Scope

- `cmd/api`, `internal/config`, `internal/router`, `internal/middleware`;
- seluruh module di `internal/modules/*`;
- `pkg/tx`, `pkg/database`, `pkg/querybuilder`;
- Casbin persistence adapter dan transactional enforcer;
- seed tooling, repository tests, integration tests, E2E, dependency manifest, dan docs.

### Non-goals

- tidak mengubah schema atau data production tanpa migration terpisah dan approval;
- tidak mengubah endpoint, response, cookie, status code, atau authorization policy karena alasan refactor;
- tidak menghapus Gin/GORM sebelum seluruh consumer dan test berpindah;
- frontend hanya disentuh bila proxy atau kontrak benar-benar terpengaruh.

## 3. Prinsip eksekusi

1. Satu fase harus lulus checkpoint sebelum fase berikutnya dimulai.
2. Setiap repository SQL harus memiliki test success, not-found, DB error, tenant filter, pagination, dan transaction path yang relevan.
3. Semua nilai user tetap menjadi bind parameter; nama kolom, operator, sort direction, dan table hanya boleh berasal dari allowlist.
4. Context tetap membawa request context, organization context, dan transaction handle; jangan membawa global config ke usecase.
5. Migration dilakukan per boundary, bukan search-and-replace seluruh repo.
6. Setiap fase wajib menyimpan hasil command dan daftar regression yang ditemukan.

## 4. Baseline dan aturan validasi

Jalankan sebelum edit:

```bash
make doctor
pnpm lint
make test
make test-integration
make test-e2e
```

Integration dan E2E membutuhkan Docker serta environment database/Redis. Catat kegagalan baseline sebagai baseline issue, bukan sebagai hasil migrasi.

Validasi wajib pada setiap checkpoint:

```bash
pnpm lint
make test
```

Tambahkan sesuai boundary:

```bash
make test-integration
make test-e2e
```

Validasi final juga mencakup:

```bash
go vet ./...
go test -race ./...
pnpm typecheck
pnpm build
scripts/guard-time-conventions.sh
```

`make test-all` boleh dipakai sebagai ringkasan, tetapi hasil tiga command individual tetap menjadi sumber kebenaran karena target tersebut selalu mengakhiri proses dengan exit code 0.

## 5. Urutan fase dan dependency

```text
F0 baseline
 -> F1 dependency/config contract
 -> F2 DBTX + sqlx connection/transaction
 -> F3 raw SQL foundation/querybuilder
 -> F4 repository migration by module
 -> F5 Casbin SQL adapter
 -> F6 HTTP compatibility layer
 -> F7 controller/middleware/router migration
 -> F8 Fx composition/lifecycle
 -> F9 seed/docs/observability cleanup
 -> F10 remove Gin/GORM and final validation
```

F4 dan F5 boleh berjalan paralel setelah F2/F3, tetapi F5 harus selesai sebelum repository permission/auth yang memakai transactional Casbin diklaim selesai. F6 dapat dimulai dengan adapter HTTP sementara, tetapi F7 membutuhkan F2 dan kontrak response yang sudah dipetakan.

## F0 - Baseline, inventory, dan contract freeze

### Owner
`cmd/api`, `internal/config`, `internal/router`, semua module delivery/repository, `tests`, `apps/web`, `apps/client`, `packages/api-types`.

### Langkah

1. Buat snapshot dependency dan status branch; jangan menghapus perubahan user.
2. Inventaris seluruh import Gin/GORM, `*gorm.DB`, `gin.Context`, GORM scopes, preload/join, soft-delete, `gorm.Expr`, `gorm-adapter`, dan `otelgorm`.
3. Daftar semua endpoint dari `internal/router/router.go`, route module, Swagger, frontend proxy, dan Postman.
4. Bekukan kontrak: method/path, status code, body JSON, error format, header/cookie, auth strata, API-key scope, tenant requirement.
5. Tandai setiap write flow yang mengubah DB + Casbin, DB + audit/outbox, cache, worker, atau upload hook.
6. Ukur benchmark dasar untuk hot query/list endpoint bila performance menjadi acceptance criterion.

### Output
Inventory, contract matrix, transaction matrix, route security matrix, baseline test report.

### Gate
Tidak ada desain raw SQL sebelum seluruh dynamic list dan transaction-sensitive flow memiliki owner dan test target.

## F1 - Dependency manifest dan compatibility contracts

### Owner
`go.mod`, package config, `internal/config`, `pkg/tx`, test helpers.

### Langkah

1. Tambahkan `sqlx` dan `fx` pada branch migrasi; pertahankan Gin/GORM sementara.
2. Pilih driver tetap MySQL yang sekarang digunakan dan tetapkan placeholder `?`.
3. Definisikan interface `DBTX` yang kompatibel dengan `*sqlx.DB` dan `*sqlx.Tx`.
4. Definisikan sentinel error repository, terutama `ErrNotFound`, tanpa mengekspos `gorm.ErrRecordNotFound` ke usecase.
5. Tentukan naming convention SQL, `db` tags, scan strategy, time representation, dan nullable column strategy.
6. Tambahkan lint/static checks yang mencegah string interpolation untuk value SQL.

### Verifikasi
`go test ./pkg/... ./internal/...`, `go vet ./...`, `pnpm lint`.

### Gate
Interface tidak boleh memaksa usecase mengetahui apakah implementasinya GORM atau SQL.

## F2 - Koneksi sqlx dan transaction manager

### Owner
`internal/config/gorm.go` menjadi `sqlx.go`, `pkg/tx/*`, `internal/config/app.go`, health check.

### Langkah

1. Implementasikan DSN MySQL dengan setting charset, parse time, UTC, pool idle/open/lifetime yang sama.
2. Buat `*sqlx.DB`, `PingContext`, dan close function.
3. Ganti `TransactionManager` agar memanggil `BeginTxx(ctx, nil)` dan menyimpan `*sqlx.Tx` pada context.
4. Sediakan `DBFromContext`/`DBTXFromContext` yang mengembalikan abstraction, bukan `*gorm.DB`.
5. Pastikan rollback pada error, rollback pada panic, commit error, dan context cancellation memiliki test.
6. Update health endpoint agar memakai `PingContext` SQL dan Redis tanpa mengubah response.
7. Instrumentasi SQL menggunakan OpenTelemetry SQL instrumentation yang kompatibel, bukan `otelgorm`.

### Verifikasi
Test transaction commit/rollback/panic dengan `sqlmock`; test health; `make test`; integration database.

### Gate
Tidak boleh ada repository baru yang mengambil `*gorm.DB` dari context.

## F3 - Raw SQL foundation dan query builder aman

### Owner
`pkg/querybuilder/*`, `pkg/database/scopes.go`, model/entity tags.

### Langkah

1. Pisahkan metadata query dari GORM tags; gunakan explicit allowlist per resource/entity.
2. Implementasikan builder yang menghasilkan `(sql string, args []any, error)`.
3. Dukung equals, contains, in, between, gt, gte, lt, lte, ne dengan placeholder.
4. Validasi slice kosong untuk `IN`; tentukan hasil aman, misalnya false predicate, bukan `IN ()`.
5. Validasi field, table alias, order column, dan direction sebelum SQL dibuat.
6. Pertahankan denylist password, token, secret, key, salt, serta field sensitif lain secara fail-closed.
7. Implementasikan pagination/count terpisah dan batasi page/page-size maksimum.
8. Ganti organization scope dan organization visibility scope menjadi fragment SQL + args yang dapat dikomposisi.
9. Tambahkan test table-driven untuk injection payload, invalid field, invalid operator, NULL/global organization, soft-delete, sort, dan count.

### Verifikasi
`pkg/querybuilder`, `pkg/database` tests; `go test -race ./pkg/querybuilder/... ./pkg/database/...`; `go vet ./...`.

### Gate
Review keamanan wajib sebelum builder dipakai oleh repository user/audit/admin.

## F4 - Migrasi repository raw SQL per tingkat kesulitan

Semua module mengikuti urutan: entity scan tags -> repository constructor -> `getDB(ctx)` -> read -> write -> transaction -> dynamic query -> tests -> module constructor.

### F4.1 Project

`internal/modules/project/repository/project_repository.go`, entity/model/tests.

- SELECT tenant-scoped dan soft-delete eksplisit.
- INSERT/UPDATE/DELETE dengan affected-row check.
- Pastikan ID dari path tidak dapat melewati organization scope.

Verifikasi: project unit + integration tenant isolation.

### F4.2 Access

`internal/modules/access/repository/*`.

- Migrasikan endpoint/access-right registry reads dan writes.
- Pastikan unique conflict dan endpoint-method lookup tetap sama.

Verifikasi: access repository/usecase tests dan seed compatibility.

### F4.3 Role

`internal/modules/role/repository/*`.

- Migrasikan role CRUD, org/global visibility, soft-delete, dan list/search.
- Pertahankan cleanup policy sebagai usecase concern.

Verifikasi: role repository/usecase tests, integration role cleanup.

### F4.4 User

`internal/modules/user/repository/*`.

- Migrasikan user CRUD, SSO identity, status, hard-delete soft-deleted records.
- Tulis explicit join/subquery membership.
- Pertahankan sensitive-field filter, count, pagination, dan `deleted_at = 0`.

Verifikasi: user repository tests, query security tests, auth integration.

### F4.5 Organization

`organization/repository/organization_repository.go`, `organization_member_repository.go`, `invitation_repository.go`.

- Migrasikan organization + owner member atomic create.
- Migrasikan membership status, invitations, restore/hard-delete.
- Pertahankan Redis cache invalidation setelah commit semantics ditentukan.

Verifikasi: organization unit, tenant isolation, membership/cache integration.

### F4.6 Audit dan webhook

`audit/repository/*`, `webhook/repository/*`.

- Migrasikan outbox, pending/retry query, batch processing, log visibility, dan atomic status update.
- Ganti `gorm.Expr("retry_count + 1")` dengan ekspresi SQL statis yang tidak menerima input user.
- Pastikan claim/update outbox tidak memproses item yang sama secara tidak sengaja.

Verifikasi: audit/webhook repository tests, worker integration, rollback dan retry scenarios.

### F4.7 API key

`internal/modules/api_key/repository/*`.

- Migrasikan issuance, hash lookup, revoke, expiry, organization binding, dan scope lookup.
- Jangan pernah mencari berdasarkan raw API key; pertahankan hash-only lookup.
- Pastikan cache Redis invalidation tetap dilakukan pada revoke/rotate.

Verifikasi: API-key unit tests dan protected-route integration.

### F4.8 Auth token dan auth repository

`internal/modules/auth/repository/*`, terutama token repository dan auth-related SQL.

- Pertahankan Redis sebagai session/token store bila memang menjadi source of truth.
- Migrasikan tabel token/identity yang masih memakai GORM.
- Pertahankan not-found mapping, expiry, revocation, cleanup, dan SSO identity behavior.

Verifikasi: auth repository/usecase tests, session integration, refresh/logout E2E.

### F4.9 Stats

`internal/modules/stats/*`.

- Migrasikan aggregation query ke SQL eksplisit.
- Pastikan timezone, millisecond timestamps, NULL aggregation, dan tenant visibility sama.

Verifikasi: stats tests, integration aggregation, realtime broadcaster smoke test.

### F4.10 Module constructor cutover

- Ubah constructor setiap module dari `*gorm.DB` menjadi `*sqlx.DB` atau `DBTX`.
- Hilangkan GORM dari interface publik repository.
- Jalankan dependency/import audit untuk memastikan module tidak mengimpor GORM.

Verifikasi: package test module, `go vet ./...`, unit suite penuh.

## F5 - Casbin SQL adapter dan transactional policy

### Owner
`internal/config/casbin.go`, `internal/modules/permission/usecase/transactional_enforcer.go`, adapter baru, permission tests.

### Langkah

1. Pilih adapter SQL internal berbasis `sqlx` agar transaction handle bisa dikontrol penuh.
2. Implementasikan load policy dan grouping policy dari tabel Casbin dengan column mapping yang sama.
3. Implementasikan add/remove single dan batch policy, filtered removal, clear/load behavior, dan auto-save semantics.
4. Buat constructor adapter yang menerima `DBTX`, sehingga adapter normal memakai `*sqlx.DB` dan transactional adapter memakai `*sqlx.Tx`.
5. Pertahankan Redis watcher dan production guard: nil enforcer atau zero policy tetap abort di environment strict.
6. Ubah transactional enforcer agar mengambil `DBTXFromContext`, bukan `*gorm.DB`.
7. Uji atomicity: DB mutation gagal -> policy rollback; policy mutation gagal -> DB rollback.
8. Bandingkan hasil enforcement subject/domain/object/action sebelum dan sesudah cutover.

### Verifikasi

`internal/modules/permission/test/*`, Casbin adapter tests, middleware tests, integration tenant/Casbin, dan E2E role/permission flows.

### Approval gate

Tidak boleh menghapus `gorm-adapter` sebelum test policy load, watcher, transaction rollback, role cleanup, dan production zero-policy guard lulus.

## F6 - HTTP compatibility layer

### Owner
`internal/delivery`, `pkg/response`, middleware helper baru, controller adapter.

### Langkah

1. Definisikan handler type internal berbasis `http.Handler` dan `http.HandlerFunc`.
2. Buat helper `WriteJSON`, `WriteError`, `DecodeJSON`, content-type, request ID, dan status mapping.
3. Buat helper path/query parsing menggunakan `r.PathValue`, `r.URL.Query`, dan typed parser.
4. Pastikan error response, validation response, empty body, dan status code identik dengan baseline.
5. Buat adapter sementara agar handler lama dapat diuji di bawah `net/http` tanpa mengubah semua controller sekaligus.
6. Buat middleware adapter untuk auth, API key, CSRF, tenant, Casbin, user status, rate limit, logging, recovery, CORS, security, metrics, dan idempotency.
7. Pertahankan urutan middleware: API key -> token/session -> scope -> user session/status -> tenant -> Casbin.

### Verifikasi

Unit test setiap helper/middleware, router contract tests, auth/tenant integration. Bandingkan response golden sebelum dan sesudah.

## F7 - Migrasi controller dan router ke stdlib

### Urutan module

1. access dan stats;
2. project;
3. role dan permission;
4. user;
5. organization;
6. webhook, audit, dan API key;
7. auth;
8. special surfaces: health, metrics, SSE, WebSocket, TUS, Swagger.

### Langkah per module

1. Ubah controller signature ke `http.HandlerFunc`.
2. Ganti Gin binding dengan `DecodeJSON` dan validator yang sama.
3. Ganti `c.Param`, `c.Query`, `c.Get`, `c.Set`, `c.Request.Context`, dan response helpers.
4. Pertahankan context organization, authenticated user, API-key identity, dan ticket state.
5. Pindahkan route registration dari Gin group ke `http.ServeMux` method/path patterns.
6. Tambahkan route test untuk path, method, status, middleware order, trailing slash, dan OPTIONS/CORS.
7. Migrasikan module route files dan hapus hanya helper Gin yang sudah obsolete.

### Special surfaces

- SSE: flush dengan `http.Flusher`, disconnect via request context, auth tetap wajib.
- WebSocket: lakukan auth/ticket/origin check sebelum upgrade.
- TUS: gunakan `http.Handler`/`http.StripPrefix` tanpa mengubah metadata trust boundary.
- Metrics: gunakan `promhttp.Handler()` dan BasicAuth middleware sendiri.
- Swagger: ganti `gin-swagger` dengan handler kompatibel atau dokumentasikan migration decision.
- OTEL: ganti `otelgin` dengan `otelhttp` dan ukur span/attribute parity.

### Verifikasi

Controller/route unit tests, `make test`, `make test-integration`, dan `make test-e2e` untuk seluruh route strata.

## F8 - Fx composition root dan lifecycle

### Owner
`cmd/api/main.go`, `internal/config/app.go`, `internal/config/app_helpers.go`, module constructors, worker/realtime constructors.

### Langkah

1. Pisahkan provider config, logger, tracer, SQL DB, Redis, JWT, validator, storage, Casbin, transaction manager, dan managers.
2. Pisahkan provider repository, usecase, controller, middleware, router, worker, scheduler, dan server.
3. Hilangkan module constructor yang membuat dependency lintas module secara tersembunyi; semua dependency harus di-inject Fx.
4. Buat `fx.Module` per bounded context bila membantu ownership, tanpa memindahkan business logic ke config.
5. Gunakan `fx.Lifecycle` untuk start/stop HTTP server, worker processor, scheduler, WebSocket manager, broadcaster, dan tracer.
6. Pastikan shutdown order: stop accepting HTTP, stop scheduler/worker intake, flush/close managers, close Redis/SQL, shutdown tracer.
7. Ubah `main` menjadi load config, build `fx.App`, register signal handling, dan `app.Run`.
8. Tambahkan graph validation test untuk duplicate providers, missing dependencies, dan dependency cycle.

### Verifikasi

Application construction test, startup/shutdown integration, `make test`, `make test-integration`, E2E smoke.

## F9 - Seed, dokumentasi, observability, dan tooling cleanup

### Owner
`db/seeds/main.go`, `cmd/gen`, `docs/*`, `documentation/*`, `README.md`, `Makefile`, `llm/cache/*`, `llm/conventions/*`, `llm/workflows/*`, telemetry, dan tests.

### Langkah

1. Migrasikan seed transaction dan seluruh seed CRUD ke `sqlx` + raw SQL.
2. Pastikan seed idempotent dan policy seed memakai adapter/SQL yang sama dengan runtime.
3. Hapus GORM test helpers dan buat fixtures SQL atau `sqlmock` sesuai kebutuhan.
4. Regenerate Swagger melalui tooling yang masih didukung; bandingkan route/method/response dengan contract freeze.
5. Ganti Makefile command yang menyebut GORM/Gin dan tambahkan target migration smoke bila perlu.
6. Update documentation architecture, database, testing, dan local setup.
7. Hapus instrumentasi `otelgorm`; pastikan SQL spans dan HTTP spans tetap tersedia.
8. Audit dependency graph dan lockfile.

### Dokumentasi yang wajib disinkronkan

- `README.md`: stack utama, cara menjalankan API, dependency, command test, dan troubleshooting.
- `documentation/README.md`: indeks dan status dokumentasi.
- `documentation/api/README.md` dan kontrak API lain: handler, auth, response, dan route behavior.
- `documentation/architecture/ARCHITECTURE.md`, `SYSTEM_ARCHITECTURE.md`, dan `ARCHITECTURE_VISUAL_AND_SEQUENCE.md`: `net/http`, sqlx, raw SQL, Fx graph, lifecycle, dan request flow.
- `documentation/architecture/MULTI_TENANCY.md`: tenant predicate raw SQL, transaction context, dan Casbin ordering.
- `documentation/guides/GETTING_STARTED.md`, `DEVELOPER_FLOW.md`, `MAINTENANCE.md`, `TESTING.md`, dan `WORKTREE_FLOW.md`: setup, command, test prerequisites, Docker, dan shutdown/debug flow.
- `documentation/guides/SEARCH.md`: filter/sort allowlist, placeholder SQL, pagination, dan sensitive-field restrictions.
- `documentation/guides/OBSERVABILITY.md`: `otelhttp`, SQL instrumentation, metrics, health, worker, SSE, dan WebSocket.
- `documentation/guides/STORAGE.md`, `RESUMABLE_UPLOAD.md`, `CLIENT_UPLOAD_GUIDE.md`, `WEBHOOKS.md`: hanya bila perubahan HTTP/lifecycle memengaruhi boundary tersebut.
- `docs/swagger.yaml`, `docs/swagger.json`, dan `docs/docs.go`: regenerate setelah route/controller migration; jangan edit generated output sebagai source utama.
- `llm/cache/project-overview.md`, `architecture.md`, `backend-map.md`, `module-map.md`, `domain-rules.md`: update fakta runtime setelah fase stabil.
- `llm/cache/querybuilder-security.md`, `casbin-permission-system.md`, `authentication-system.md`, `tenant-organization-system.md`, `worker-audit-webhook-system.md`, `realtime-system.md`, dan `tus-upload-system.md`: update boundary yang berubah.
- `llm/conventions/golang.md`, `database.md`, dan `testing.md`: update pola constructor, DBTX, raw SQL, mock, dan command validasi.
- `llm/workflows/go-service.md`, `api-endpoint.md`, `database-migration.md`, `cross-stack-change.md`, `verification-before-completion` bila workflow atau checklist migrasi berubah.

### Urutan update dokumentasi

1. Setelah setiap fase, update hanya dokumen yang owner boundary-nya berubah.
2. Setelah test fase lulus, bandingkan klaim dokumen dengan live code dan command aktual.
3. Setelah F7, regenerate Swagger dan audit kedua frontend proxy sebelum mengubah kontrak API docs.
4. Setelah F8, update diagram dependency/lifecycle dan contoh startup/shutdown Fx.
5. Setelah F10, lakukan stale-reference sweep untuk Gin, GORM, `gorm.DB`, `gin.Context`, `otelgorm`, dan constructor lama.
6. Tandai dokumen historis atau kontrak deprecated secara eksplisit; jangan meninggalkan instruksi lama tanpa label.

### Aturan source of truth

- Runtime code dan test menentukan behavior.
- Migration files menentukan schema.
- Generated Swagger mengikuti annotations/route code dan harus diregenerate, bukan ditulis manual sebagai solusi permanen.
- `llm/cache` hanya diperbarui setelah fakta diverifikasi dari live code yang sudah lulus test.
- Plan/task menyimpan keputusan aktif dan blocker; cache tidak boleh menyimpan asumsi yang belum terbukti.

### Verifikasi dokumentasi

- `git diff --check` untuk seluruh perubahan docs.
- Link/path audit untuk file lokal dan command audit terhadap `package.json`, `Makefile`, serta `go.mod`.
- Search stale references setelah tiap cutover dan wajib zero stale runtime instruction sebelum F10.
- `pnpm go:docs` setelah perubahan route/controller.
- `make test-integration` untuk validasi seed/docs yang bergantung pada database.
- Review manual README, getting started, testing, architecture, API, dan troubleshooting sebagai user baru.

### Verifikasi

Seed against clean Docker MySQL, `pnpm go:docs`, `make test-integration`, E2E from seeded state, stale-reference sweep, dan docs diff review.

## F10 - Removal, hardening, dan final acceptance

### Langkah

1. Pastikan tidak ada import Gin, GORM, `gorm-adapter`, `otelgorm`, `soft_delete`, atau `gin-swagger` pada runtime.
2. Hapus dependency lama dari `go.mod` hanya setelah `go mod tidy` dan full build lulus.
3. Hapus adapter compatibility, dead code, old route helpers, dan test fixtures yang tidak lagi dipakai.
4. Jalankan security review raw SQL: injection, allowlist, tenant predicates, soft-delete predicates, affected rows, dan transaction boundaries.
5. Jalankan load/smoke check untuk auth, tenant, list/search, permission, upload, worker, SSE, dan WebSocket.
6. Review diff per fase dan pastikan tidak ada perubahan API yang tidak disetujui.

## 6. Final acceptance matrix

Wajib lulus dan dicatat pada CI/release checklist:

```bash
pnpm lint
make test
make test-integration
make test-e2e
go vet ./...
go test -race ./...
pnpm typecheck
pnpm build
scripts/guard-time-conventions.sh
```

Acceptance tambahan:

- `go.mod` tidak lagi memuat Gin/GORM dependency runtime;
- seluruh route contract dan frontend proxy smoke test lulus;
- production Casbin guard tetap fail-closed;
- tenant isolation dan API-key scope test lulus;
- rollback transaction dan Casbin policy atomicity terbukti;
- startup dan graceful shutdown Fx terbukti;
- tidak ada known regression yang ditutup dengan skip test.

## 7. Risiko dan keputusan yang memerlukan approval

- Apakah breaking internal controller API boleh dilakukan sekaligus, atau wajib compatibility adapter beberapa sprint?
- Apakah Casbin policy tetap berada di database yang sama dan harus atomic dengan entity writes?
- Apakah Swagger harus tetap tersedia pada path `/api/v1/docs/*any` selama migrasi?
- Apakah acceptance performance memerlukan benchmark sebelum/sesudah?
- Apakah branch migration memakai dual-runtime deployment atau maintenance window untuk cutover?

Rekomendasi default: pertahankan endpoint dan schema, lakukan dual-stack internal, migrasikan repository lebih dahulu, lalu HTTP, lalu Fx, dan hapus dependency lama hanya setelah final acceptance matrix lulus.

## 8. Keputusan implementasi preskriptif untuk AI agent

Bagian ini adalah keputusan default. Agent tidak perlu membandingkan alternatif lagi kecuali menemukan fakta live code yang bertentangan; bila itu terjadi, berhenti pada gate fase dan laporkan konflik tersebut.

### F0 - Keputusan baseline

- Gunakan branch `tech/raw-stdlib` sebagai branch migrasi aktif.
- Jangan mengubah schema, endpoint, response, cookie, atau policy pada fase baseline.
- Jadikan live code sebagai source of truth; cache/docs hanya sebagai peta awal.
- Simpan inventory di plan/task, bukan melakukan refactor saat inventory.
- Ambil contract golden dari test HTTP, Swagger, Postman, dan proxy kedua frontend.
- Definition of done: baseline command selesai atau blocker environment terdokumentasi lengkap.

### F1 - Keputusan dependency dan interface

- Gunakan `github.com/jmoiron/sqlx` + driver MySQL yang sudah ada; jangan mengganti database driver.
- Gunakan `go.uber.org/fx`; jangan memperkenalkan wire, dig, atau service locator.
- Definisikan satu interface `DBTX` internal untuk `GetContext`, `SelectContext`, `ExecContext`, `QueryxContext`, dan `QueryRowxContext`.
- Repository menerima `*sqlx.DB` pada constructor dan mengambil `DBTX` transaction handle dari context saat diperlukan.
- Usecase hanya bergantung pada interface repository dan transaction manager, tidak pada `sqlx`.
- Jangan menambahkan generic repository atau active-record abstraction.
- Definition of done: package usecase tidak memiliki import GORM maupun sqlx.

### F2 - Keputusan connection dan transaction

- Rename `internal/config/gorm.go` menjadi implementasi `sqlx` baru; jangan mempertahankan dua pool database runtime.
- Pool utama adalah `*sqlx.DB`; semua operasi memakai `Context` variant.
- Transaction manager memulai `BeginTxx` dan menyimpan hanya `*sqlx.Tx` sebagai `DBTX` di context.
- Gunakan satu transaction owner pada usecase/repository boundary; jangan membuat nested transaction otomatis.
- Untuk commit/rollback failure, kembalikan error asli yang dibungkus dengan `%w`; jangan menggantinya dengan error generik.
- Panic harus rollback lalu dipropagasi ulang.
- Definition of done: test commit, rollback, panic, cancellation, dan commit failure lulus.

### F3 - Keputusan raw SQL dan query builder

- Gunakan placeholder `?` untuk MySQL dan selalu bind value melalui args.
- Nama table, alias, column, operator, dan direction berasal dari allowlist statis; tidak boleh berasal langsung dari request.
- Gunakan explicit query specification per resource, bukan refleksi GORM tag.
- Pertahankan `db` tags untuk scanning dan tambahkan mapping SQL eksplisit bila tag tidak cukup.
- Generate builder result sebagai `SQL`, `Args`, dan `error`; jangan mengembalikan object query mutable.
- Untuk filter `IN` kosong, gunakan predicate yang selalu false dan test-kan perilakunya.
- Untuk `contains`, escape wildcard `%` dan `_` sesuai kontrak pencarian yang dipilih; dokumentasikan keputusan sebelum implementasi.
- Pagination wajib punya maximum page size dari config/default; count dan data query terpisah.
- Definition of done: injection payload, sensitive field, invalid sort/filter, tenant scope, dan NULL behavior memiliki test.

### F4 - Keputusan urutan repository

- Migrasikan satu module sampai test dan import audit lulus sebelum pindah module berikutnya.
- Urutan wajib: project -> access -> role -> user -> organization -> audit -> webhook -> api_key -> auth -> stats.
- Jangan migrasikan permission repository sebagai repository biasa; permission adalah bagian dari F5 karena Casbin transaction coupling.
- Pertahankan nama method interface dan bentuk entity selama tidak dipaksa oleh SQL scan.
- Gunakan `sql.NullString`, `sql.NullInt64`, atau pointer hanya untuk kolom yang memang nullable; jangan mengubah nullability diam-diam.
- Gunakan explicit `INSERT` column list dan explicit `UPDATE` column list; jangan memakai `SELECT *` atau update seluruh struct.
- Untuk not-found, map ke sentinel/domain error yang sama sebelum migrasi.
- Semua delete soft-delete ditulis eksplisit dengan `deleted_at = 0`/timestamp sesuai schema; hard delete memakai query terpisah.
- Definition of done per module: unit repository, usecase regression, tenant/integration test relevan, dan tidak ada import GORM pada module tersebut.

### F4 - Keputusan khusus multi-tenant

- Organization predicate harus ditulis dekat query repository, bukan hanya dipercaya dari middleware.
- Required tenant routes tetap wajib memiliki organization context; raw SQL tidak boleh menganggap context kosong sebagai wildcard secara tidak sengaja.
- Global rows dengan `organization_id IS NULL` hanya boleh muncul pada query yang memang mendukung global visibility.
- Parent organization visibility predicate tetap diterapkan pada child resources.
- Setiap query list/search harus punya test cross-tenant: organization A tidak pernah membaca data B.
- Definition of done: integration test membuktikan required tenant, optional admin organization, global row, dan soft-deleted organization behavior.

### F5 - Keputusan Casbin

- Implementasikan adapter Casbin internal berbasis SQLX/`DBTX`; jangan mencari adapter pihak ketiga baru kecuali adapter tersebut terbukti mendukung transaction handle.
- Pertahankan tabel dan column shape Casbin yang sudah dipakai production.
- Normal enforcer memakai pool utama; transactional enforcer membuat adapter/enforcer yang memakai transaction handle dari context.
- Jangan menggunakan global enforcer untuk policy write di dalam transaction-sensitive usecase.
- Pertahankan Redis watcher dan strict production guard tanpa perubahan fail-open.
- Uji grouping policy, domain, path normalization, action, batch policy, remove filtered policy, dan watcher update.
- Definition of done: policy write rollback dan entity write rollback terbukti dua arah, bukan hanya happy path.

### F6 - Keputusan HTTP compatibility

- Target akhir adalah `http.ServeMux` Go 1.25, bukan router pihak ketiga.
- Gunakan `http.Handler` middleware chain standar.
- Buat satu response/error package internal dan pakai di semua controller baru.
- `http.Request.Context()` menjadi satu-satunya carrier request state; gunakan typed private context keys.
- Decode JSON dengan `json.Decoder`, batasi body size, dan tolak trailing JSON token bila kontrak mengharuskannya.
- Jangan mengubah format error selama refactor; salin behavior baseline lalu tambah test.
- Buat adapter sementara hanya untuk memperkecil batch migrasi, lalu hapus sebelum F10.
- Definition of done: handler stdlib dapat diuji dengan `httptest` tanpa Gin.

### F7 - Keputusan controller dan route

- Migrasikan controller tanpa memindahkan business logic dari usecase.
- Registrasikan route menggunakan method/path pattern Go 1.25 dan `r.PathValue`.
- Pertahankan urutan middleware persis: API key -> token/session -> scope -> user session/status -> tenant -> Casbin.
- Pertahankan route strata: public, authenticated, tenantAuthorized, authorized, dan special upload/realtime.
- Jangan mengganti route path parameter `:id` menjadi path yang berbeda secara publik; hanya representasi router internal yang berubah.
- Auth, organization, permission, API-key, upload, SSE, dan WebSocket wajib dimigrasikan paling akhir setelah controller sederhana stabil.
- Definition of done: route matrix membandingkan method/path/status/body/header untuk seluruh route utama.

### F8 - Keputusan Fx

- `cmd/api/main.go` hanya memuat config, membuat Fx app, dan menunggu lifecycle.
- Provider dikelompokkan berdasarkan responsibility: config/logging, infrastructure, repositories, usecases, delivery, runtime.
- Gunakan constructor injection; jangan memakai `fx.Populate` untuk production dependency.
- Gunakan `fx.Lifecycle` untuk semua goroutine/server yang perlu start/stop.
- Jika ada dependency cycle, pecahkan ownership constructor; jangan menyelesaikan dengan global variable atau lazy service locator.
- Buat satu `fx.Module` per bounded context hanya bila graph tetap mudah dibaca.
- Definition of done: startup test, shutdown test, missing dependency test, dan cycle-free Fx graph lulus.

### F9 - Keputusan seed, docs, dan telemetry

- Seed memakai raw SQL dan transaction manager yang sama secara konsep dengan runtime; jangan menghidupkan GORM hanya untuk seed.
- Seed harus idempotent dengan `INSERT ... ON DUPLICATE KEY UPDATE` atau select-then-update yang teruji.
- Swagger path publik dipertahankan; pilih handler non-Gin yang paling kompatibel sebelum menghapus `gin-swagger`.
- Gunakan `otelhttp` untuk HTTP dan SQL instrumentation yang kompatibel dengan `database/sql`; jangan mempertahankan `otelgorm` sebagai transitive workaround.
- Update docs setelah behavior parity terbukti, bukan sebelumnya.
- Definition of done: clean database seed, rerun seed, generated docs, dan observability smoke semuanya lulus.

### F9 - Keputusan dokumentasi

- Dokumentasi di-update per fase, bukan dikumpulkan setelah seluruh migrasi selesai.
- Untuk setiap perubahan code, tambahkan perubahan dokumentasi pada commit/PR fase yang sama agar drift langsung terlihat.
- Gunakan daftar wajib di bagian F9 sebagai checklist; file conditional hanya disentuh bila boundary-nya benar-benar berubah.
- Pertahankan nama endpoint, contoh request, response, cookie, dan middleware order dalam docs sampai contract test membuktikan perubahan.
- Semua contoh SQL harus memakai placeholder dan tidak boleh menampilkan secret, password, atau credential nyata.
- Diagram architecture harus menunjukkan Fx composition root, `*sqlx.DB`, transaction context, repository raw SQL, dan `http.ServeMux`.
- Dokumentasikan perbedaan `make test`, `make test-integration`, dan `make test-e2e`; jangan mengarahkan user hanya ke `make test-all`.
- Setelah F10, grep stale references harus zero untuk runtime docs; historical migration notes boleh menyebut Gin/GORM bila diberi konteks deprecated/historical.
- Definition of done: docs build/link review, Swagger regeneration, stale-reference sweep, dan manual setup/test walkthrough lulus.

### F10 - Keputusan removal dan final gate

- Hapus dependency lama hanya dalam commit/fase terpisah setelah seluruh test lulus dengan compatibility code masih aktif.
- Jalankan `go mod tidy`, build, dan import audit setelah removal.
- Jangan menghapus test lama sebelum test ekuivalen raw SQL/stdHTTP tersedia.
- Jangan menutup test dengan `Skip`, memperlonggar assertion, atau mematikan race detector untuk mengatasi regression.
- Bila integration/E2E gagal karena Docker atau environment, tandai sebagai blocked dan jangan klaim migrasi selesai.
- Definition of done hanya tercapai setelah final acceptance matrix dan security review selesai.

## 9. Format kerja agent per task

Untuk setiap task implementasi, agent harus mengikuti urutan berikut:

1. Baca owner path dan keputusan fase terkait.
2. Baca interface, constructor, test terdekat, dan caller utama.
3. Nyatakan satu hypothesis lokal dan satu test yang dapat membantahnya.
4. Implementasikan perubahan terkecil pada satu boundary.
5. Jalankan validasi paling sempit yang tersedia segera setelah edit.
6. Perbaiki hanya slice yang gagal lalu ulangi command yang sama.
7. Jalankan checkpoint fase sebelum membuka fase berikutnya.
8. Catat file berubah, command, hasil, dan blocker pada task/PR notes.

Agent tidak boleh melakukan batch migration lintas module, mengubah schema, atau mengganti kontrak API tanpa approval gate yang sesuai.