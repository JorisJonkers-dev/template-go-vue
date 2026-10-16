// Package assembly wires the service together: each context's use cases and adapters on the store,
// and the routes the server serves. main builds the service here, and so do the integration tests.
package assembly

import (
	"log/slog"

	"github.com/JorisJonkers-dev/template-go-vue/internal/notes/adapters/persistence"
	"github.com/JorisJonkers-dev/template-go-vue/internal/notes/app"
	"github.com/JorisJonkers-dev/template-go-vue/internal/platform/httpapi"
	"github.com/JorisJonkers-dev/template-go-vue/internal/platform/httpx"
	"github.com/JorisJonkers-dev/template-go-vue/internal/platform/pg"
	"github.com/JorisJonkers-dev/template-go-vue/internal/platform/webui"
	"github.com/JorisJonkers-dev/template-go-vue/internal/server"
	"github.com/JorisJonkers-dev/template-go-vue/web"
)

// Config is what the service is built from.
type Config struct {
	Logger *slog.Logger
	Store  *pg.Store
	// DevUser stands in for the platform's forward-auth on a local run: every request without an
	// identity runs as it. It must never be set in a deployment.
	DevUser string
}

// Routes builds the API from each context's web adapter, and serves the embedded web app beside it.
func Routes(cfg Config) (server.Routes, error) {
	api, err := httpapi.New(cfg.Logger, app.New(persistence.New(cfg.Store.Queries())))
	if err != nil {
		return server.Routes{}, err
	}
	if cfg.DevUser != "" {
		cfg.Logger.Warn("DEV_USER is set: every request without an identity runs as it", "subject", cfg.DevUser)
		api = httpx.DevIdentity(cfg.DevUser, api)
	}
	return server.Routes{API: api, Web: webui.Handler(web.Dist()), Ready: cfg.Store.Ping}, nil
}
