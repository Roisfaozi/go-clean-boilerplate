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
		handler(c.Writer, c.Request)
	}
}

// AdaptGinEngineHandler wraps gin.Engine.ServeHTTP as standard http.Handler.
func AdaptGinEngineHandler(engine *gin.Engine) http.Handler {
	return engine
}
