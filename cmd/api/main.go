package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strings"

	_ "github.com/Roisfaozi/go-clean-boilerplate/docs"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/config"
	"go.uber.org/fx"
)

// @title           Go Clean Boilerplate API
// @version         1.0
// @description     This is a clean and modular boilerplate for Go REST APIs with RBAC, Audit Logs, and WebSockets.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description "Type 'Bearer ' followed by a space and the access token."
func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if cfg.JWT.AccessTokenSecret == "" || cfg.JWT.RefreshTokenSecret == "" {
		log.Fatal("JWT secrets are not set. Please check your .env file or environment variables.")
	}

	if cfg.JWT.AccessTokenSecret == cfg.JWT.RefreshTokenSecret {
		log.Fatal("JWT_ACCESS_SECRET and JWT_REFRESH_SECRET must be different.")
	}

	if strings.EqualFold(cfg.Cookie.SameSite, "none") {
		if cfg.Cookie.Secure == nil || !*cfg.Cookie.Secure {
			log.Fatal("COOKIE_SAME_SITE=none requires COOKIE_SECURE=true (browser rejects SameSite=None without Secure)")
		}
	}

	if cfg.Pprof.Enabled {
		go func() {
			pprofAddr := fmt.Sprintf("localhost:%d", cfg.Pprof.Port)
			log.Printf("Starting pprof server on %s", pprofAddr)
			if err := http.ListenAndServe(pprofAddr, nil); err != nil {
				log.Printf("Failed to start pprof server: %v", err)
			}
		}()
	}

	fxApp := fx.New(
		config.CoreFxModule,
		fx.Supply(cfg),
	)

	fxApp.Run()
}
