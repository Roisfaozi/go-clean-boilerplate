# Prompt 2: The Feature Implementer (Chain of Thought)

**Use Case:** Gunakan prompt ini setelah arsitektur disetujui, untuk menghasilkan kode implementasi.

---

**Role:**
You are a Senior Golang Developer.

**Task:**
Implement the `[NAMA_FITUR]` based on the architecture plan.

**Technique: Chain of Thought (CoT)**
Before writing the code, I want you to **think step-by-step** to ensure data flows correctly through the layers.

**Steps:**
1.  **Think about the Model:** Define the DTOs (`model`) and Entities (`entity`). What validation tags are needed?
2.  **Think about the Interface:** Define the `Repository` interface. What methods are needed for this feature?
3.  **Think about the UseCase:** Define the `UseCase` interface. How does it call the repository? Where does the transaction block (`WithinTransaction`) go?
4.  **Think about the Handler:** How do we parse the request? How do we handle specific errors (`ErrNotFound`, `ErrConflict`) returned by the UseCase?
5.  **Action:** Write the code for each file sequentially (Model -> Entity -> Repo Interface -> Repo Impl -> UseCase Interface -> UseCase Impl -> Handler -> Route).

**Critical Instructions:**
-   Ensure `repo` implementation uses `context`.
-   Ensure `usecase` handles validations (`validate.Struct`).
-   Ensure `handler` returns standardized `WebResponse`.

**Input:**
[Paste the Architect's Plan or feature description here]
