package http_test

import (
	"testing"

	"github.com/gin-gonic/gin"
	roleHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/test/mocks"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestRegisterAuthorizedRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockUseCase := new(mocks.MockRoleUseCase)
	v := validator.New()
	log := logrus.New()

	roleController := roleHttp.NewRoleController(mockUseCase, log, v)

	roleHttp.RegisterAuthorizedRoutes(router.Group("/api/v1"), roleController)

	routes := router.Routes()

	expectedRoutes := map[string]string{
		"POST":   "/api/v1/roles",
		"GET":    "/api/v1/roles",
		"PUT":    "/api/v1/roles/:id",
		"DELETE": "/api/v1/roles/:id",
	}
	// Add /search as POST /api/v1/roles/search
	expectedRoutesDynamic := map[string]string{
		"POST": "/api/v1/roles/search",
	}

	foundRoutes := make(map[string]string)
	for _, route := range routes {
		foundRoutes[route.Method + " " + route.Path] = route.Path
	}

	for method, path := range expectedRoutes {
		key := method + " " + path
		_, exists := foundRoutes[key]
		assert.True(t, exists, "Expected route %s %s to be registered", method, path)
	}

	for method, path := range expectedRoutesDynamic {
		key := method + " " + path
		_, exists := foundRoutes[key]
		assert.True(t, exists, "Expected route %s %s to be registered", method, path)
	}
}
