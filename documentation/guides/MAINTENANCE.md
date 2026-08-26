# Maintenance Guide (Cleanup Jobs & Scheduler)

This project includes an automated maintenance system built on top of the **Asynq Scheduler**. It ensures the database remains clean and performant by pruning stale data automatically.

## Scheduled Tasks

| Task Name                       | Schedule              | Description                                                             |
| :------------------------------ | :-------------------- | :---------------------------------------------------------------------- |
| `cleanup:expired_tokens`        | Every 6 hours         | Deletes expired password reset tokens from the database.                |
| `cleanup:soft_deleted_entities` | Daily (03:00 AM)      | Permanently deletes users that were soft-deleted more than 30 days ago. |
| `cleanup:prune_audit_logs`      | Weekly (Sun 04:00 AM) | Prunes audit logs older than 180 days (6 months).                       |

## How It Works

1.  **Repository Layer**: Implements specific cleanup queries (e.g., `DeleteExpiredResetTokens`).
2.  **Worker Handler**: `CleanupTaskHandler` coordinates the repository calls.
3.  **Scheduler**: `internal/worker/scheduler.go` defines the Cron schedules and enqueues tasks into Redis.
4.  **Processor**: The background worker picks up tasks from the queue and executes them.

## Configuration

You can adjust retention periods in `internal/worker/scheduler.go`:

```go
// Example: Change user retention to 60 days
payloadUser, _ := json.Marshal(tasks.CleanupSoftDeletedEntitiesPayload{RetentionDays: 60})
```

## Monitoring

Maintenance logs are visible in the application output with the `worker` context:

```text
INFO Starting cleanup of expired reset tokens
INFO Completed cleanup of expired reset tokens
```

For advanced monitoring, you can use the `asynqmon` tool to view the state of the maintenance queues.

## Configuration Loading Rules (Viper)

`internal/config/config.go` builds `AppConfig` with `v.Unmarshal(&cfg)`. Viper's
`AutomaticEnv()` is a *resolver*, not an environment scanner: it can only resolve
keys Viper already knows, and `Unmarshal` iterates over `v.AllKeys()`.

A key only enters `AllKeys()` when it has a `SetDefault`, exists in a config
file, or is explicitly bound. So a field with **no default is silently dropped**
even when its environment variable is set — no error, no warning.

Because of that, every default-less key that must come from the environment is
registered in `envOnlyKeys` and bound before `Unmarshal`:

```go
var envOnlyKeys = []string{
	"server.frontend_base_url",
	"cookie.secure",
	"redis.dial_timeout",
	"redis.read_timeout",
	"redis.write_timeout",
}

for _, key := range envOnlyKeys {
	_ = v.BindEnv(key)
}
```

Env var names are derived automatically: the key is upper-cased and `.` becomes
`_` via `SetEnvKeyReplacer`, so `cookie.secure` reads `COOKIE_SECURE`.

### Rules when adding config

1. If the value may be guessed safely, add a `SetDefault`.
2. If it must not be guessed (origins, secrets, credentials), add the key to
   `envOnlyKeys` instead — do **not** invent a default.
3. Keep the `BindEnv` loop before `v.Unmarshal`; after it, bindings have no effect.
4. Document the variable in `.env.example`.

Regression guard: `TestNewConfig_EnvOnlyKeysAreBound` in
`internal/config/config_test.go` sets all five variables and asserts they reach
the struct.

**Known limit:** the list is manual. A new default-less field that is not added
to `envOnlyKeys` will reintroduce the same silent-drop bug and the test will not
catch it, since it only checks the keys already listed.

### Operational notes

- `SERVER_FRONTEND_BASE_URL` has no default and is the base of every link the
  backend emails. If unset, those links are generated without a domain.
- `COOKIE_SECURE` is now actually honoured. Set it to `false` for plain-HTTP
  local testing, otherwise the browser refuses to store the session cookie and
  login appears to fail with no clear error.
