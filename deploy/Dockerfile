# Stage 1: Builder
FROM golang:1.25-alpine AS builder

# Install dependencies needed for build (e.g. gcc for cgo if needed, though we try to avoid it)
RUN apk add --no-cache git make build-base

WORKDIR /app

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# -ldflags="-w -s" strips debug information to reduce size
# CGO_ENABLED=0 ensures a static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/api/main.go

# Stage 2: Runner
FROM alpine:latest

# Install necessary runtime dependencies (e.g., certificates, timezone)
RUN apk --no-cache add ca-certificates tzdata

# Create a non-root user
RUN adduser -D -u 1000 appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/main .

# Copy migration files (if migration is run by the app on start, or by a separate tool)
COPY --from=builder /app/db/migrations ./db/migrations

# Set ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Set timezone (Optional, example for Jakarta)
ENV TZ=Asia/Jakarta

# Expose port
EXPOSE 8080

# Healthcheck
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
CMD ["./main"]
