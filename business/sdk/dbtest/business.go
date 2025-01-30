package dbtest

import (
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/rmsj/service/business/domain/homebus"
	"github.com/rmsj/service/business/domain/homebus/stores/homedb"
	"github.com/rmsj/service/business/domain/productbus"
	"github.com/rmsj/service/business/domain/productbus/stores/productdb"
	"github.com/rmsj/service/business/domain/userbus"
	"github.com/rmsj/service/business/domain/userbus/stores/userdb"
	"github.com/rmsj/service/business/domain/vproductbus"
	"github.com/rmsj/service/business/domain/vproductbus/stores/vproductdb"
	"github.com/rmsj/service/business/sdk/delegate"
	"github.com/rmsj/service/foundation/logger"
)

// BusDomain represents all the business domain apis needed for testing.
type BusDomain struct {
	Delegate *delegate.Delegate
	Home     *homebus.Business
	Product  *productbus.Business
	User     *userbus.Business
	VProduct *vproductbus.Business
}

func newBusDomains(log *logger.Logger, db *sqlx.DB) BusDomain {
	dlg := delegate.New(log)
	userBus := userbus.NewBusiness(log, dlg, userdb.NewStore(log, db, time.Hour))
	productBus := productbus.NewBusiness(log, userBus, dlg, productdb.NewStore(log, db))
	homeBus := homebus.NewBusiness(log, userBus, dlg, homedb.NewStore(log, db))
	vproductBus := vproductbus.NewBusiness(vproductdb.NewStore(log, db))

	return BusDomain{
		Delegate: dlg,
		Home:     homeBus,
		Product:  productBus,
		User:     userBus,
		VProduct: vproductBus,
	}
}
