package delivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdaptHTTPHandler converts a standard http.HandlerFunc to gin.HandlerFunc for phased cutover.
func AdaptHTTPHandler(handler http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, param := range c.Params {
			c.Request.SetPathValue(param.Key, param.Value)
		}

		ctx := c.Request.Context()
		if val, exists := c.Get("user_id"); exists {
			if s, ok := val.(string); ok && s != "" {
				ctx = SetContextValue(ctx, UserIDKey, s)
			}
		}
		if val, exists := c.Get("organization_id"); exists {
			if s, ok := val.(string); ok && s != "" {
				ctx = SetContextValue(ctx, OrganizationIDKey, s)
			}
		}
		if val, exists := c.Get("username"); exists {
			if s, ok := val.(string); ok && s != "" {
				ctx = SetContextValue(ctx, UsernameKey, s)
			}
		}
		if val, exists := c.Get("user_role"); exists {
			if s, ok := val.(string); ok && s != "" {
				ctx = SetContextValue(ctx, RoleKey, s)
			}
		}
		if val, exists := c.Get("session_id"); exists {
			if s, ok := val.(string); ok && s != "" {
				ctx = SetContextValue(ctx, SessionIDKey, s)
			}
		}
		if val, exists := c.Get("auth_method"); exists {
			if s, ok := val.(string); ok && s != "" {
				ctx = SetContextValue(ctx, AuthMethodKey, s)
			}
		}
		if val, exists := c.Get("api_key_id"); exists {
			if s, ok := val.(string); ok && s != "" {
				ctx = SetContextValue(ctx, APIKeyIDKey, s)
			}
		}
		if val, exists := c.Get("api_key_scopes"); exists {
			if scopes, ok := val.([]string); ok {
				ctx = SetContextValue(ctx, ScopesKey, scopes)
			}
		}

		handler(c.Writer, c.Request.WithContext(ctx))
	}
}

// AdaptGinEngineHandler wraps gin.Engine.ServeHTTP as standard http.Handler.
func AdaptGinEngineHandler(engine *gin.Engine) http.Handler {
	return engine
}
