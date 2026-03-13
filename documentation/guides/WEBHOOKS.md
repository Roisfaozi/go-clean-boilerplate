# Outbound Webhooks Guide

NexusOS Admin provides an outbound webhook system that allows you to notify external systems (like Slack, Zapier, or your own microservices) when specific events occur within the platform.

## 1. Architecture

The webhook system is designed for reliability and high performance:

1.  **Event Trigger**: A business action (e.g., user registration) occurs.
2.  **Task Enqueue**: The system identifies all active webhooks subscribed to that event and enqueues tasks in **Redis** via `asynq`.
3.  **Background Processing**: Dedicated workers pick up the tasks and execute the HTTP POST requests.
4.  **Logging**: Every attempt (success or failure) is logged in the `webhook_logs` table for auditing and debugging.

## 2. Security

To ensure that webhook payloads are authentic and haven't been tampered with, NexusOS includes several security measures:

### HMAC-SHA256 Signature
Every request includes an `X-Webhook-Signature` header. This is a hex-encoded HMAC-SHA256 hash of the request body, using the webhook's **Secret Key**.

**Verification Example (Go):**
```go
func VerifySignature(secret string, body []byte, signature string) bool {
    h := hmac.New(sha256.New, []byte(secret))
    h.Write(body)
    expected := hex.EncodeToString(h.Sum(nil))
    return hmac.Equal([]byte(expected), []byte(signature))
}
```

### Replay Attack Prevention
The `X-Webhook-Timestamp` header contains the Unix millisecond timestamp when the request was generated. Recipients should verify that this timestamp is within a reasonable window (e.g., 5 minutes) to prevent replay attacks.

## 3. Webhook Payloads

All webhooks are sent as `POST` requests with a `application/json` content type.

### Standard Headers
- `Content-Type: application/json`
- `X-Webhook-ID`: The unique ID of the webhook configuration.
- `X-Webhook-Event`: The type of event (e.g., `user.created`).
- `X-Webhook-Signature`: HMAC-SHA256 signature.
- `X-Webhook-Timestamp`: Unix millisecond timestamp.

## 4. Supported Events

| Event Key      | Description                                | Payload Content                          |
| :------------- | :----------------------------------------- | :--------------------------------------- |
| `user.created` | Triggered when a new user registers.       | `id`, `username`, `email`, `fullname`    |
| `org.created`  | Triggered when a new organization is made. | `id`, `name`, `slug`, `owner_id`         |

## 5. Management API

| Action            | Endpoint                                | Access Right     |
| :---------------- | :-------------------------------------- | :--------------- |
| Create Webhook    | `POST /api/v1/webhooks`                 | `webhook:manage` |
| List Webhooks     | `GET /api/v1/webhooks`                  | `webhook:manage` |
| Get Webhook Logs  | `GET /api/v1/webhooks/:id/logs`         | `webhook:manage` |
| Update Webhook    | `PUT /api/v1/webhooks/:id`              | `webhook:manage` |
| Delete Webhook    | `DELETE /api/v1/webhooks/:id`           | `webhook:manage` |

## 6. Best Practices

1.  **Idempotency**: Your endpoint should be able to handle the same event multiple times in case of retries.
2.  **Timeout**: Respond quickly (within 2-5 seconds). The NexusOS worker has a 10-second timeout.
3.  **Security**: Always validate the `X-Webhook-Signature`.
4.  **Availability**: If your server returns a `5xx` error, the worker will attempt to retry (configurable in `internal/worker/processor.go`).
