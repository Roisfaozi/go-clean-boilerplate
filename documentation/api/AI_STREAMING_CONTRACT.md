# API Contract: NexusAI Streaming Service

This document defines the interface for the real-time AI assistant (NexusAI). It uses Server-Sent Events (SSE) to provide a smooth, word-by-word streaming experience.

## 1. Endpoint Metadata

- **URL**: `POST /api/v1/ai/chat`
- **Auth**: Required (Bearer Token or Session Cookie)
- **Request Format**: `application/json`
- **Response Format**: `text/event-stream` (Standard SSE)

---

## 2. Request Schema (JSON)

The frontend sends the full conversation history and the current application context.

```json
{
  "messages": [
    {
      "role": "user",
      "content": "Why are there so many failed logins today?"
    }
  ],
  "context": {
    "current_page": "/dashboard/audit",
    "organization_id": "uuid-org-123",
    "data": [
      { "id": 1, "action": "LOGIN", "status": "failed", "ip": "10.0.0.1" },
      { "id": 2, "action": "LOGIN", "status": "failed", "ip": "10.0.0.1" }
    ]
  }
}
```

### Field Definitions

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `messages` | `Array` | Yes | List of previous messages. Role must be `user` or `assistant`. |
| `context` | `Object` | No | Data to help the AI understand what the user is looking at. |
| `context.current_page` | `String` | No | The current route/URL path. |
| `context.data` | `Any` | No | Raw JSON data from the current view (e.g., table rows). |

---

## 3. Response Schema (Server-Sent Events)

The response must be streamed using the `text/event-stream` MIME type. Each data chunk must be prefixed with `data: `.

### Data Chunk Format (JSON)
```json
data: {"text": "Based"}
data: {"text": " on"}
data: {"text": " the"}
data: {"text": " logs,"}
```

### Finish Event (Optional)
Used to signify the end of transmission and send metadata like token usage.
```text
event: finish
data: {"message_id": "msg_999", "usage": {"prompt_tokens": 50, "completion_tokens": 100}}
```

---

## 4. Backend Implementation Guide (Go)

### Example Implementation (Gin)

```go
func HandleAIChat(c *gin.Context) {
    // 1. Set SSE Headers
    c.Writer.Header().Set("Content-Type", "text/event-stream")
    c.Writer.Header().Set("Cache-Control", "no-cache")
    c.Writer.Header().Set("Connection", "keep-alive")
    c.Writer.Header().Set("Transfer-Encoding", "chunked")

    // 2. Bind JSON Request
    var req AIChatRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    // 3. Setup LLM Stream (e.g., OpenAI or Gemini)
    // Ensure you inject the system prompt based on req.Context
    stream, err := llmProvider.GetStream(c.Request.Context(), req)
    if err != nil {
        c.JSON(500, gin.H{"error": "ai service unavailable"})
        return
    }

    // 4. Stream Loop
    c.Stream(func(w io.Writer) bool {
        if chunk, ok := <-stream; ok {
            payload, _ := json.Marshal(map[string]string{"text": chunk})
            fmt.Fprintf(w, "data: %s

", string(payload))
            return true
        }
        fmt.Fprintf(w, "event: finish
data: {}

")
        return false 
    })
}
```

---

## 5. Security & Performance Policies

1. **Rate Limiting**: AI inference is resource-intensive. Limit users to 5-10 requests per minute.
2. **Context Scrubbing**: The backend **must** remove sensitive fields (passwords, secrets, personal PII) from `context.data` before sending it to an external LLM provider.
3. **Timeout**: Connections should be automatically closed if the LLM doesn't respond within 30 seconds.
4. **Token Limits**: Truncate conversation history if it exceeds the LLM's context window.

---
*Created: February 12, 2026*  
*Project: NexusOS Enterprise*
