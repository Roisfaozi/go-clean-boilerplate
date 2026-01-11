# Maintenance Guide (Cleanup Jobs & Scheduler)

This project includes an automated maintenance system built on top of the **Asynq Scheduler**. It ensures the database remains clean and performant by pruning stale data automatically.

## Scheduled Tasks

| Task Name | Schedule | Description |
| :--- | :--- | :--- |
| `cleanup:expired_tokens` | Every 6 hours | Deletes expired password reset tokens from the database. |
| `cleanup:soft_deleted_entities` | Daily (03:00 AM) | Permanently deletes users that were soft-deleted more than 30 days ago. |
| `cleanup:prune_audit_logs` | Weekly (Sun 04:00 AM) | Prunes audit logs older than 180 days (6 months). |

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
