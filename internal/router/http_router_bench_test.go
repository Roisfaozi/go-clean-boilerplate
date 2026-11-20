package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
)

func BenchmarkStdMiddlewareChain_Health(b *testing.B) {
	mux := http.NewServeMux()
	health := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.Handle("GET /api/v1/health", delivery.Chain(
		health,
		middleware.HTTPRequestIDMiddleware(),
		middleware.HTTPPrometheusMiddleware(),
		middleware.HTTPSecurityMiddleware(),
		middleware.HTTPCORSMiddleware([]string{"http://localhost:3000"}),
	))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			b.Fatalf("unexpected status %d", rec.Code)
		}
	}
}

func BenchmarkStdServeMux_ParameterizedRoute(b *testing.B) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/projects/{id}", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(r.PathValue("id")))
	})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/proj-123", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Body.String() != "proj-123" {
			b.Fatalf("unexpected body %q", rec.Body.String())
		}
	}
}
