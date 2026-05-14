# Test Suite Documentation

This directory contains the comprehensive test suite for the Go Clean Boilerplate project. Specifically designed to ensure robustness, security, and scalability.

## 📂 Structure

```
tests/
├── e2e/                # End-to-End tests (HTTP Client -> Server -> DB)
│   ├── api/            # Test cases per module (Auth, User, etc.)
│   └── setup/          # Test server & client helpers
│
├── integration/        # Integration tests (UseCase -> Repository -> DB)
│   ├── modules/        # Module-specific integration tests
│   ├── scenarios/      # Complex business scenarios (User Lifecycle)
│   └── setup/          # Singleton Docker containers (MySQL, Redis)
│
├── fixtures/           # Data factories for generating test data
└── helpers/            # Shared test utilities (Assertions, Waiters)
```

## 🚀 Key Concepts

### Singleton Containers
To optimize performance, we use a **Singleton Container Pattern**. Instead of spinning up new Docker containers for every test package, we initialize one global MySQL and Redis instance shared across all integration tests.
- **Speed**: Reduces suite execution time by ~80%.
- **Resource**: Low memory footprint on developer machines/CI.

### Comprehensive vs Standard Tests
- **Standard (`*_test.go`)**: Validates "Happy Paths" and core functionality.
- **Comprehensive (`*_comprehensive_test.go`)**: Validates "Unhappy Paths", Edge Cases, Security vulnerabilities (SQLi, XSS), and complex logic.

## 🏃 Running Tests

Run the full suite using Make:
```bash
make test-all
```

Or run specific layers:
```bash
make test-unit          # Fast logic checks
make test-integration   # DB/Cache interactions
make test-e2e           # Full system verification
```