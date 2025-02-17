// Package all binds all the routes into the specified app.
package all

import (
	"time"

	"github.com/rmsj/service/app/domain/checkapp"
	"github.com/rmsj/service/app/domain/productapp"
	"github.com/rmsj/service/app/domain/rawapp"
	"github.com/rmsj/service/app/domain/tranapp"
	"github.com/rmsj/service/app/domain/userapp"
	"github.com/rmsj/service/app/domain/vproductapp"
	"github.com/rmsj/service/app/sdk/mux"
	"github.com/rmsj/service/business/domain/productbus"
	"github.com/rmsj/service/business/domain/productbus/stores/productdb"
	"github.com/rmsj/service/business/domain/userbus"
	"github.com/rmsj/service/business/domain/userbus/stores/userdb"
	"github.com/rmsj/service/business/domain/vproductbus"
	"github.com/rmsj/service/business/domain/vproductbus/stores/vproductdb"
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
	vproductBus := vproductbus.NewBusiness(vproductdb.NewStore(cfg.Log, cfg.DB))

	checkapp.Routes(app, checkapp.Config{
		Build: cfg.Build,
		Log:   cfg.Log,
		DB:    cfg.DB,
	})

	productapp.Routes(app, productapp.Config{
		Log:        cfg.Log,
		UserBus:    userBus,
		ProductBus: productBus,
		AuthClient: cfg.AuthClient,
	})

	rawapp.Routes(app)

	tranapp.Routes(app, tranapp.Config{
		Log:        cfg.Log,
		DB:         cfg.DB,
		UserBus:    userBus,
		ProductBus: productBus,
		AuthClient: cfg.AuthClient,
	})

	userapp.Routes(app, userapp.Config{
		Log:        cfg.Log,
		UserBus:    userBus,
		AuthClient: cfg.AuthClient,
	})

	vproductapp.Routes(app, vproductapp.Config{
		Log:         cfg.Log,
		UserBus:     userBus,
		VProductBus: vproductBus,
		AuthClient:  cfg.AuthClient,
	})
}
