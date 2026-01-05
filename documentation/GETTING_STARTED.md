# Getting Started: Running the Casbin DB Project

This guide provides step-by-step instructions to set up and run the Casbin DB project for the first time, including bootstrapping the database with necessary initial data.

---

## 🚀 1. Prerequisites

Before you begin, ensure you have the following installed on your system:

*   **Go**: Version 1.25.5 or higher.
*   **Docker** & **Docker Compose**: Essential for running the database (MySQL) and cache (Redis) services.
*   **Make**: The `Makefile` simplifies common development tasks.
*   **Git**: For cloning the repository.
*   **Air** (Optional but Recommended for Development): For live-reloading the application during development.
    ```bash
    go install github.com/air-verse/air@latest
    ```
*   **Swag CLI** (Optional): For regenerating API documentation if you make changes to Swagger annotations.
    ```bash
    go install github.com/swaggo/swag/cmd/swag@latest
    ```
*   **Golang Migrate** (Optional): If you want to run migrations manually without the Makefile helper.
    ```bash
    go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    ```
*   **C/C++ Compiler (GCC/MinGW-w64)**: Required for running repository tests that use SQLite (due to CGO). Ensure `gcc` is in your system's PATH.

---

## ⚙️ 2. Setup & First Run

Follow these steps to get your project up and running:

### Step 2.1: Clone the Repository

If you haven't already, clone the project repository:

```bash
git clone https://github.com/Roisfaozi/go-clean-boilerplate.git
cd go-clean-boilerplate
```

### Step 2.2: Configure Environment Variables

Create a `.env` file from the example and configure your application settings. This includes database credentials, JWT secrets, and other runtime parameters.

```bash
cp .env.example .env
```
*   **Note**: The provided `docker-compose.yml` uses the default values from `.env.example`, so you can usually proceed without changes for a quick local setup.

### Step 2.3: Start Infrastructure Services

Use Docker Compose to launch the MySQL database and Redis cache containers. These are essential for the application to function.

```bash
docker-compose up -d
```
*   This command will download the necessary Docker images and start the services in the background.

### Step 2.4: Run Database Migrations

Apply the database schema migrations. This will create tables and the very basic default roles (`role:user`, `role:admin`).

```bash
make migrate-up
```

### Step 2.5: Seed Initial Data

After migrations, you can seed more comprehensive initial data, such as a `role:superadmin` and a default super admin user, along with initial Casbin policies.

```bash
make seed-up
```

### Step 2.6: Run the Application

Now that the database and Redis are ready, you can start the Go application.

**For Development (with Live Reload):**
```bash
air
```
*   This command will compile and run your application, automatically restarting it whenever code changes are detected.

**Standard Run (without Live Reload):**
```bash
go run cmd/api/main.go
```

The application will typically start on `http://localhost:8080` (check your `.env` or console output for the exact port).

---

## ✅ 3. Verification

Once the application is running, you can verify the setup:

*   **Access Swagger UI**: Open your web browser and navigate to `http://localhost:8080/api/docs/index.html`. You should see the interactive API documentation.
*   **Test with Postman**:
    1.  Import the Postman collection `postman/Casbin_API_Full_Flow.postman_collection.json` into Postman.
    2.  Set up your Postman environment variables (`baseURL` to `http://localhost:8080`, `apiVersion` to `v1`).
    3.  Run the collection. All requests should execute successfully, demonstrating user registration, login, role assignment, and RBAC checks.

You are now ready to develop and extend the Casbin DB project!

---