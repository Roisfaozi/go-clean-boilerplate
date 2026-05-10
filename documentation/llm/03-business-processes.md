# Business Processes by Feature

Dokumen ini menguraikan business process utama satu per satu. Fokusnya bukan pada detail endpoint semata, tetapi pada trigger, aturan bisnis, perubahan data, dan side effect.

## 1. User Registration

### Tujuan bisnis

Mengonversi identitas baru menjadi user aktif yang siap memakai platform dengan workspace awal.

### Trigger

`POST /api/v1/auth/register`

### Alur

1. sistem cek username belum dipakai
2. sistem cek email belum dipakai
3. password di-hash
4. user dibuat dalam status aktif
5. default role diberikan
6. default workspace or organization dibuat
7. membership owner dibuat untuk organization tersebut
8. audit REGISTER dijadwalkan
9. user langsung login dan mendapat session

### Hasil bisnis

- user identity tercipta
- tenant awal tercipta
- user siap masuk aplikasi tanpa setup manual

## 2. Login

### Tujuan bisnis

Memberi akses ke user aktif sambil menegakkan kebijakan keamanan login.

### Aturan bisnis

- account lockout diterapkan
- invalid username tetap menjalani dummy password check
- hanya user status aktif yang boleh login
- session baru disimpan di Redis

### Side effect

- audit LOGIN
- publish event login
- refresh token issued bersama access token

## 3. Refresh Token

### Tujuan bisnis

Memperpanjang akses tanpa memaksa login ulang, tetapi tetap menjaga kontrol session.

### Aturan bisnis

- refresh token harus valid
- user masih harus aktif
- session lama direvoke
- session baru dibuat

### Nilai bisnis

Sistem mendapatkan UX token refresh sambil tetap mempertahankan kemampuan revoke session.

## 4. Logout and Session Revocation

### Trigger

- logout satu session
- revoke all sessions

### Aturan bisnis

- session dihapus dari Redis
- audit dibuat

### Nilai bisnis

Tim operasi atau user bisa mematikan akses aktif secara langsung.

## 5. Organization Creation

### Tujuan bisnis

Membuat tenant baru yang siap dipakai.

### Alur

1. slug dicek unik
2. organization dibuat
3. owner membership dibuat
4. Casbin grouping policy untuk owner dibuat dalam domain org

### Hasil bisnis

- tenant baru aktif
- owner langsung punya akses admin tenant

## 6. Invite Organization Member

### Tujuan bisnis

Memungkinkan kolaborasi tim, termasuk untuk email yang belum menjadi akun penuh.

### Aturan bisnis

- organization harus ada
- jika email belum terdaftar, shadow user dibuat
- membership invited dibuat
- invitation token disimpan
- invitation lama dibersihkan
- email invitation dikirim async

### Nilai bisnis

Platform mendukung team onboarding sebelum user menyelesaikan registrasi penuh.

## 7. Update Member Role or Status

### Tujuan bisnis

Mengubah posisi dan kewenangan member di dalam tenant.

### Aturan bisnis

- membership harus sudah ada
- perubahan role di DB harus disinkronkan ke Casbin grouping policy
- role lama dibersihkan sebelum role baru ditambahkan

### Nilai bisnis

Role bisnis di organization benar-benar menjadi role teknis di authorization engine.

## 8. Create and Manage Roles

### Tujuan bisnis

Menyediakan entitas role yang bisa dipakai untuk governance.

### Aturan bisnis

- nama role harus unik
- `role:superadmin` tidak boleh dihapus
- saat role dihapus, policy Casbin terkait ikut dibersihkan

### Nilai bisnis

Admin dapat mengelola katalog role tanpa meninggalkan policy yatim.

## 9. Assign Role to User

### Tujuan bisnis

Mengaitkan user dengan role tertentu di domain tertentu.

### Aturan bisnis

- domain default adalah `global`
- user dan role harus ada
- role lama di domain itu dibersihkan dahulu
- role baru ditambahkan ke grouping policy

### Nilai bisnis

Sistem memodelkan satu role efektif per domain untuk satu user, sehingga governance lebih sederhana.

## 10. Assign Access Right

### Tujuan bisnis

Memberikan sekumpulan izin ke role dalam bentuk abstraksi yang lebih mudah dipahami daripada endpoint mentah.

### Alur

1. access right diambil dari repository
2. endpoint di bawah access right dibaca
3. setiap endpoint diterjemahkan menjadi policy Casbin
4. audit ASSIGN_ACCESS_RIGHT dibuat

### Nilai bisnis

Admin bisa mengelola izin secara logical grouping, bukan route per route.

## 11. User Creation by Admin

### Tujuan bisnis

Membuat akun user secara administratif.

### Alur

1. cek username dan email unik
2. hash password
3. create user
4. assign default role global
5. create audit CREATE
6. trigger webhook `user.created`

### Nilai bisnis

Aktivitas administratif dapat diaudit dan diintegrasikan ke sistem eksternal.

## 12. User Update and Status Update

### Tujuan bisnis

Memelihara identity record user.

### Aturan bisnis

- username baru harus unik
- password baru harus di-hash
- perubahan penting tercatat ke audit

### Catatan

Usecase user adalah domain pendukung penting, tetapi bukan pusat produk seperti auth or permission.

## 13. Avatar Upload

### Tujuan bisnis

Mengelola profile media user dengan pipeline file yang fleksibel.

### Dua jalur

- upload biasa via storage provider
- TUS resumable upload dengan avatar hook

### Nilai bisnis

Platform siap untuk file upload yang lebih tahan gangguan koneksi.

## 14. API Key Issuance

### Tujuan bisnis

Menyediakan kredensial machine-to-machine per organization.

### Aturan bisnis

- raw key dibuat aman
- yang disimpan hanya hash
- identity hasil auth berisi org, user, username, scopes, dan expiry
- hasil auth di-cache di Redis

### Nilai bisnis

Integrasi backend dan automation script bisa menggunakan credential selain JWT user session.

## 15. Webhook Trigger

### Tujuan bisnis

Mengekspos event internal ke sistem eksternal.

### Alur

1. event business terjadi
2. webhook yang subscribe event itu diambil
3. tiap webhook dibuatkan task async
4. worker menembak URL target

### Nilai bisnis

Sistem mendukung integrasi event-driven tanpa menghambat request utama.

## 16. Project Lifecycle

### Tujuan bisnis

Menyediakan contoh resource tenant-aware.

### Alur inti

- create project dalam organization
- list project per organization
- update project
- delete project

### Catatan

Project belum memiliki orchestration yang dalam. Saat ini ia lebih merepresentasikan resource bisnis dasar daripada core domain utama.

## 17. Audit Logging

### Tujuan bisnis

Menyediakan jejak aktivitas yang dapat dipakai untuk compliance, observability, dan forensic review.

### Dua mode

- write direct jika tidak berada dalam transaction context
- write ke outbox jika berada dalam transaction context

### Nilai bisnis

Audit tidak tertinggal dari perubahan data utama dan tetap dapat diekspor atau dibroadcast.

## 18. Dashboard Stats

### Tujuan bisnis

Menyediakan ringkasan operasional untuk dashboard.

### Sumber data

- users
- roles
- audit logs
- organization members

### Catatan

Sebagian insight masih placeholder, jadi fitur ini belum setara kematangan dengan auth atau permission.

## Fitur yang Paling Core

Jika diprioritaskan berdasarkan business value aktual:

1. registration and login
2. organization and membership
3. roles and permissions
4. audit logging
5. session revocation

## Fitur yang Lebih Supportive

- webhook
- API key
- project
- stats
- avatar upload
- realtime notifications

## Kesimpulan

Business process paling inti di codebase ini membentuk satu rantai:

- user dibuat
- tenant dibentuk
- membership dan role dipasang
- permission diberlakukan
- aktivitas dicatat
- integrasi dan realtime memperluas efeknya

Itulah inti nilai platform ini.
