# Comparison: Starter-Pack Template vs. Casbin Config

## Tujuan

Membandingkan starter-pack asli dari `ai-native-template` dengan kondisi riil di repositori `Casbin`.

## Persamaan Utama

- Menggunakan `AGENTS.md` sebagai entrypoint utama.
- Menyimpan cache memori dalam `llm/cache/`.
- Menyimpan playbook standar dalam `llm/workflows/`.
- Memisahkan backend (`typescript-go`) dan frontend di setup.

## Perbedaan Kritis (Gap Config)

| Fitur                 | ai-native-template (Carbon) | Casbin (Current Repo)                                    |
| --------------------- | --------------------------- | -------------------------------------------------------- |
| **Framework Backend** | Go (Asumsi dasar)           | Go dengan **Gin + GORM + Casbin**                        |
| **Monorepo Tooling**  | Bebas/Generik               | `pnpm` + `turbo`                                         |
| **Frontend Stack**    | Generik (Next/React)        | `apps/web` (Next.js) & `apps/client` (React Router)      |
| **Autentikasi**       | Session/JWT standar         | JWT + **Redis Session** + SSO                            |
| **Otorisasi**         | Biasa                       | Otorisasi **berbasis Casbin dengan Tenant (Organisasi)** |
| **Testing**           | Unit/Integrasi              | Unit + **Integration + E2E Docker (Testcontainers)**     |
| **Task Queue**        | Standard Go Channels        | **Asynq (Redis)** + Scheduler                            |
| **File Upload**       | S3 biasa                    | **TUS Protocol (`tusd`)**                                |
| **Realtime**          | WebSocket standar           | **SSE + WebSocket dengan Redis Ticket**                  |

## Kesimpulan

Repo Casbin _sudah_ mengimplementasikan filosofi starter-pack dengan sangat dalam, bahkan jauh lebih spesifik dan restriktif (contoh: _tenant org, casbin enforcer, api-key middleware, worker system_) dibanding template Carbon yang generik. Struktur cache dan workflow saat ini sudah 100% tervalidasi terhadap live code (`phase-compliance.md`).
