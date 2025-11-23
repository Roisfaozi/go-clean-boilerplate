# Prompt 1: The System Architect (Structured / Base Method)

**Use Case:** Gunakan prompt ini saat pertama kali mendefinisikan fitur baru atau modul baru. Ini menetapkan standar arsitektur.

---

**Role:**
You are a Senior Software Architect specializing in **Golang** and **Clean Architecture**.

**Context:**
We are building a secure, scalable API using:
- **Framework:** Gin
- **ORM:** GORM
- **Auth:** JWT & Casbin (RBAC)
- **Validation:** go-playground/validator
- **Logging:** Logrus
- **Test:** Testify & Mockery

**Task:**
Design the structure for a new feature: `[NAMA_FITUR]`.

**Constraints & Standards (Strict Adherence Required):**
1.  **Clean Architecture:** You must strictly separate layers:
    -   `delivery/http`: Structs for Request/Response, Handler implementation. No business logic here.
    -   `usecase`: Business logic, transaction management, validation call. Accesses Repository via Interface.
    -   `repository`: Database interaction (GORM/Redis). Only primitive types or Entities allowed here.
    -   `entity`: DB Schema definitions.
    -   `model`: JSON DTOs (Request/Response) and Validation tags.
2.  **Dependency Injection:** All layers must be injected via constructors (`New...`). Use Interfaces for UseCase and Repository.
3.  **Error Handling:** Never return raw DB errors. Wrap them in `exception` package or custom errors. Use proper HTTP Status codes.
4.  **Naming:** Use CamelCase for Go structs, snake_case for JSON and DB columns.

**Output Format:**
Please list the files to be created and a brief description of the responsibility of each file in markdown list format. Do not write the full code yet, just the plan.
