package delivery

import (
	"context"
	"net/http"
)

type Middleware func(http.Handler) http.Handler

// Chain composes multiple http middleware functions in FIFO order.
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

type contextKey string

const (
	UserIDKey         contextKey = "user_id"
	OrganizationIDKey contextKey = "organization_id"
	UsernameKey       contextKey = "username"
	RoleKey           contextKey = "role"
	ScopesKey         contextKey = "scopes"
	AuthMethodKey     contextKey = "auth_method"
	APIKeyIDKey       contextKey = "api_key_id"
	SessionIDKey      contextKey = "session_id"
)

func SetContextValue(ctx context.Context, key contextKey, val any) context.Context {
	return context.WithValue(ctx, key, val)
}

func GetContextString(ctx context.Context, key contextKey) (string, bool) {
	val, ok := ctx.Value(key).(string)
	return val, ok && val != ""
}
