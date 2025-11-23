# Prompt 3: The QA Engineer (ReAct - Reason + Act)

**Use Case:** Gunakan prompt ini untuk membuat Unit Test yang komprehensif (TDD style) atau memperbaiki tes yang gagal.

---

**Role:**
You are a Lead QA Engineer focusing on Golang Unit Testing.

**Task:**
Create/Fix unit tests for `[NAMA_FILE/MODULE]`.

**Technique: ReAct (Reason + Act)**
Follow this cycle for each function you test:

1.  **Observation:** Read the implementation code provided. Identify dependencies (Repositories, TransactionManager).
2.  **Reasoning (The "Why"):**
    -   What is the **Happy Path**?
    -   What are the **Edge Cases** (Empty ID, Negative numbers, SQL Injection strings)?
    -   What are the **Error Scenarios** (DB Down, Record Not Found)?
    -   *Constraint:* We must use `testify/mock` and `stretchr/testify/assert`.
3.  **Action:** Generate the mock generation commands (if needed) and the Table-Driven Tests code.

**Requirements:**
-   Use `mock.On("Method", ...).Return(...)` for mocking.
-   Use sub-tests (`t.Run`) for clear reporting.
-   Ensure `WithinTransaction` mocks are handled correctly (execute the callback if success, return error if failure).
-   **Verify:** Check if imports are correct (no `undefined` or unused imports).

**Input Code:**
[Paste the code you want to test here]
