# TASK: Implement Multi-tenancy Architecture via Strict TDD

You are an expert Golang Backend Engineer. Your mission is to transform the current Single-Tenant backend into a Multi-tenant SaaS platform using **Strict Test-Driven Development (TDD)**.

## 🚨 CRITICAL INSTRUCTION: TDD IS MANDATORY
Do NOT write implementation code before writing the test. For every component, you must follow this cycle:
1.  **RED:** Write the test file (Unit/Integration) based on the requirements. Run it. It MUST fail.
2.  **GREEN:** Write the minimal implementation code to make the test pass.
3.  **REFACTOR:** Clean up the code while keeping the test green.

**If you write implementation code without a preceding failing test, the task is considered FAILED.**

## 1. Context & Documentation (Source of Truth)
Before writing any code, read and understand these files. They contain the schemas, logic, and exact test scenarios:
1.  `documentation/productplan/multi-tenancy-prd.md` (Schema & Business Rules)
2.  `documentation/ops/MULTI_TENANCY_IMPLEMENTATION.md` (Technical Blueprint & Code Snippets)
3.  `documentation/ops/TDD_UNIT_TEST_REQUIREMENTS_ADVANCED.md` (Detailed Unit Test Scenarios)
4.  `documentation/ops/TDD_INTEGRATION_TEST_REQUIREMENTS.md` (Detailed Integration Test Scenarios)
5.  `documentation/ops/MULTI_TENANCY_IMPACT_ANALYSIS.md` (List of affected files and refactor spots)

## 2. Execution Steps (TDD Workflow)

### Phase 1: Foundation (Entities & Migrations)
1.  Create migration files (SQL) in `db/migrations/`.
2.  Run the migration to update the DB schema.
3.  Create Entity structs in `internal/modules/organization/entity/`.

### Phase 2: Repository Layer & Scopes (TDD)
1.  Create `pkg/database/scopes_test.go` -> Fail -> Implement `scopes.go`.
2.  Create `organization_repository_test.go` (Integration) -> Fail -> Implement `OrganizationRepository`.
3.  **Verification:** Ensure `Create` is atomic (3 tables + owner member) and queries are scoped by `organization_id`.

### Phase 3: Logic & Middleware (TDD)
1.  Generate Mocks for the new repositories using `mockery`.
2.  Create `organization_usecase_test.go` -> Fail -> Implement `OrganizationUseCase` (CRUD + Slug logic).
3.  Create `tenant_middleware_test.go` -> Fail -> Implement `TenantMiddleware` using the `CachedOrganizationReader` strategy.
4.  **Verification:** Test "Header Spoofing", "Banned Member", and "Cache Hit/Miss" scenarios.

### Phase 4: API & Integration (TDD)
1.  Create `organization_e2e_test.go` -> Fail -> Implement `OrganizationController`.
2.  Wire everything in `internal/config/app.go` and `internal/router/router.go`.
3.  **Verification:** Use "Centralized Grouping" (Option 2) to register global vs tenant routes.

## 3. Implementation Guardrails
*   **Architecture:** Strictly follow existing Clean Architecture layers.
*   **Identity Model:** Support "Global User, Local Member" (Users table is global, membership table is scoped).
*   **Performance:** Use Redis for membership validation in the middleware.
*   **Security:** Every tenant-scoped query MUST use the GORM scope.

---
**Status:** Ready to start.
**Action:** Begin with Phase 1 (Migrations and Entities).
