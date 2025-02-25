package checkapp

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/rmsj/service/foundation/logger"
	"github.com/rmsj/service/foundation/web"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Build string
	Log   *logger.Logger
	DB    *sqlx.DB
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {
	const version = "v1"

	api := newApp(cfg.Build, cfg.Log, cfg.DB)

	app.HandlerFuncNoMid(http.MethodGet, version, "/readiness", api.readiness)
	app.HandlerFuncNoMid(http.MethodGet, version, "/liveness", api.liveness)
}
