// Package crud binds the crud domain set of routes into the specified app.
package crud

import (
	"time"

	"github.com/rmsj/service/app/domain/checkapp"
	"github.com/rmsj/service/app/domain/homeapp"
	"github.com/rmsj/service/app/domain/productapp"
	"github.com/rmsj/service/app/domain/tranapp"
	"github.com/rmsj/service/app/domain/userapp"
	"github.com/rmsj/service/app/sdk/mux"
	"github.com/rmsj/service/business/domain/homebus"
	"github.com/rmsj/service/business/domain/homebus/stores/homedb"
	"github.com/rmsj/service/business/domain/productbus"
	"github.com/rmsj/service/business/domain/productbus/stores/productdb"
	"github.com/rmsj/service/business/domain/userbus"
	"github.com/rmsj/service/business/domain/userbus/stores/userdb"
	"github.com/rmsj/service/business/sdk/delegate"
	"github.com/rmsj/service/foundation/web"
)

// Routes constructs the add value which provides the implementation of
// of RouteAdder for specifying what routes to bind to this instance.
func Routes() add {
	return add{}
}

type add struct{}

// Add implements the RouterAdder interface.
func (add) Add(app *web.App, cfg mux.Config) {

	// Construct the business domain packages we need here so we are using the
	// sames instances for the different set of domain apis.
	dlg := delegate.New(cfg.Log)
	userBus := userbus.NewBusiness(cfg.Log, dlg, userdb.NewStore(cfg.Log, cfg.DB, time.Minute))
	productBus := productbus.NewBusiness(cfg.Log, userBus, dlg, productdb.NewStore(cfg.Log, cfg.DB))
	homeBus := homebus.NewBusiness(cfg.Log, userBus, dlg, homedb.NewStore(cfg.Log, cfg.DB))

	checkapp.Routes(app, checkapp.Config{
		Build: cfg.Build,
		Log:   cfg.Log,
		DB:    cfg.DB,
	})

	homeapp.Routes(app, homeapp.Config{
		UserBus:    userBus,
		HomeBus:    homeBus,
		AuthClient: cfg.AuthClient,
	})

	productapp.Routes(app, productapp.Config{
		UserBus:    userBus,
		ProductBus: productBus,
		AuthClient: cfg.AuthClient,
	})

	tranapp.Routes(app, tranapp.Config{
		UserBus:    userBus,
		ProductBus: productBus,
		Log:        cfg.Log,
		AuthClient: cfg.AuthClient,
		DB:         cfg.DB,
	})

	userapp.Routes(app, userapp.Config{
		UserBus:    userBus,
		AuthClient: cfg.AuthClient,
	})
}
