package middleware

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type IdempotencyResponse struct {
	Status  int               `json:"status"`
	Headers map[string]string `json:"headers"`
	Body    []byte            `json:"body"`
}

type idempotencyResponseRecorder struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
	body        *bytes.Buffer
}

func (w *idempotencyResponseRecorder) WriteHeader(code int) {
	if !w.wroteHeader {
		w.statusCode = code
		w.wroteHeader = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *idempotencyResponseRecorder) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.statusCode = http.StatusOK
		w.wroteHeader = true
	}
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *idempotencyResponseRecorder) Flush() {
	if fl, ok := w.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

func (w *idempotencyResponseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("hijack not supported")
	}
	return hj.Hijack()
}

func (w *idempotencyResponseRecorder) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func writeIdempotencyReplay(w http.ResponseWriter, cachedResp *IdempotencyResponse) {
	for k, v := range cachedResp.Headers {
		w.Header().Set(k, v)
	}
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}
	w.Header().Set("X-Cache-Lookup", "HIT - Idempotency Replay")
	status := cachedResp.Status
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	_, _ = w.Write(cachedResp.Body)
}

// HTTPIdempotencyMiddleware deduplicates mutating requests (POST, PUT, PATCH) keyed by Idempotency-Key.
func HTTPIdempotencyMiddleware(rdb *redis.Client, log *logrus.Logger, ttl time.Duration) delivery.Middleware {
	return func(next http.Handler) http.Handler {
		if rdb == nil {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost && r.Method != http.MethodPut && r.Method != http.MethodPatch {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get("Idempotency-Key")
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			var bodyBytes []byte
			if r.Body != nil {
				var err error
				bodyBytes, err = io.ReadAll(r.Body)
				if err == nil {
					r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			}

			hasher := sha256.New()
			hasher.Write(bodyBytes)
			payloadHash := hex.EncodeToString(hasher.Sum(nil))

			redisKey := "idempotency:" + key
			lockKey := "idempotency_lock:" + key

			acquired, err := rdb.SetNX(r.Context(), lockKey, payloadHash, 10*time.Second).Result()
			if err != nil {
				if log != nil {
					log.WithError(err).Error("Idempotency middleware: failed to acquire lock")
				}
				next.ServeHTTP(w, r)
				return
			}

			if !acquired {
				existingHash, _ := rdb.Get(r.Context(), lockKey).Result()
				if existingHash != payloadHash {
					response.WriteError(w, http.StatusConflict, errors.New("idempotency key conflict"), "payload mismatch for same idempotency key")
					return
				}

				deadline := time.Now().Add(2 * time.Second)
				ticker := time.NewTicker(50 * time.Millisecond)
				defer ticker.Stop()

				for time.Now().Before(deadline) {
					cachedData, cacheErr := rdb.Get(r.Context(), redisKey).Result()
					if cacheErr == nil && cachedData != "" {
						var cachedResp IdempotencyResponse
						if err := json.Unmarshal([]byte(cachedData), &cachedResp); err == nil {
							writeIdempotencyReplay(w, &cachedResp)
							return
						}
					}

					select {
					case <-r.Context().Done():
						response.WriteError(w, http.StatusRequestTimeout, errors.New("idempotency wait cancelled"), "request cancelled while waiting for idempotent replay")
						return
					case <-ticker.C:
					}
				}

				response.WriteError(w, http.StatusConflict, errors.New("idempotency processing timeout"), "original request still processing")
				return
			}

			defer func() {
				_ = rdb.Del(context.WithoutCancel(r.Context()), lockKey).Err()
			}()

			cachedData, err := rdb.Get(r.Context(), redisKey).Result()
			if err == nil && cachedData != "" {
				var cachedResp IdempotencyResponse
				if err := json.Unmarshal([]byte(cachedData), &cachedResp); err == nil {
					writeIdempotencyReplay(w, &cachedResp)
					return
				}
			}

			rec := &idempotencyResponseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
				body:           bytes.NewBuffer(nil),
			}

			next.ServeHTTP(rec, r)

			if rec.statusCode >= 200 && rec.statusCode < 300 {
				headers := make(map[string]string)
				if contentType := rec.Header().Get("Content-Type"); contentType != "" {
					headers["Content-Type"] = contentType
				}

				respToCache := IdempotencyResponse{
					Status:  rec.statusCode,
					Headers: headers,
					Body:    rec.body.Bytes(),
				}

				cachedBytes, err := json.Marshal(respToCache)
				if err == nil {
					_ = rdb.Set(context.Background(), redisKey, cachedBytes, ttl).Err()
				}
			}
		})
	}
}
