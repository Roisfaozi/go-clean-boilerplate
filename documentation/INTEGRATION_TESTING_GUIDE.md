# Integration Testing Guide

This project uses `testcontainers-go` to perform integration testing with real MySQL and Redis instances.

## 🚀 Optimization: Singleton Container Pattern

To avoid high resource usage and slow execution times (common when spinning up containers for every test), we implement the **Singleton Container Pattern**.

### How it works:
1.  **Shared Instance**: Only one MySQL and one Redis container are started for the entire test suite execution within a single process.
2.  **Once Initialization**: We use `sync.Once` in `tests/integration/setup/test_container.go` to ensure containers start only once.
3.  **Sequential Execution**: Tests are run sequentially (`-p 1`) to prevent data race conditions while sharing the same database.
4.  **Lightweight Cleanup**: Instead of restarting containers, we use `TRUNCATE` on all tables between test cases to ensure a clean slate.

## 🛠 Running Tests

Run the integration suite using the following command:
```bash
make test-integration
```

This runs:
```bash
go test -v ./tests/integration/... -tags=integration -p 1 -timeout=10m
```

## 🏗 Setup Anatomy

- `test_container.go`: Manages the lifecycle of Docker containers.
- `test_database.go`: Handles migrations, seeding (default roles/policies), and truncation.
- `fixtures/`: Contains factories (e.g., `UserFactory`) to generate test data programmatically.

## 💡 Best Practices
- **Do not use `t.Parallel()`**: Since we share one database instance, parallel tests will cause conflicts.
- **Use Factories**: Always use `fixtures` to create dependent data (like roles) before creating the primary entity.
- **Seed Policies**: Default policies for `role:user` are seeded automatically to support profiles and auth flows.