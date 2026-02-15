# Task: Implement Account Lockout (Brute Force Protection)

## 🎯 Objective
Prevent brute-force attacks by locking a user account after specific consecutive failed login attempts.

## 🛠 Specifications

### 1. Configuration (`internal/config/config.go`)
Add `Security` struct to `AppConfig`:
- `MaxLoginAttempts`: Default `5`.
- `LockoutDuration`: Default `30m` (30 minutes).

### 2. Redis Keys
Use Redis to track attempts.
- Key: `auth:attempts:{username}` (Counter)
- Key: `auth:locked:{username}` (Flag, TTL = LockoutDuration)

### 3. Implementation Logic (`internal/modules/auth/usecase/auth_usecase.go`)
Modify the `Login` method:

1.  **Check Lockout**: Before verifying password, check if `auth:locked:{username}` exists.
    - If exists, return `ErrAccountLocked` with remaining time.
2.  **Verify Password**:
    - **If Valid**: Delete `auth:attempts:{username}` (Reset counter). Proceed to login.
    - **If Invalid**:
        - Increment `auth:attempts:{username}`.
        - Check if attempts >= `MaxLoginAttempts`.
        - If limit reached:
            - Set `auth:locked:{username}` with TTL.
            - Log audit event `ACCOUNT_LOCKED`.
            - Return `ErrAccountLocked`.
        - If limit not reached: Return `ErrInvalidCredentials`.

### 4. Testing
- Unit Test: Verify counter increment and lockout trigger.
- Integration Test: Simulate 5 failed logins and ensure 6th attempt is blocked even with correct password (until TTL expires).
