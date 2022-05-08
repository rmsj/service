// Package product provides an example of a core business API. Right now these
// calls are just wrapping the data/store layer. But at some point you will
// want auditing or something that isn't specific to the data/store layer.
package product

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/rmsj/service/business/core/product/db"
	"github.com/rmsj/service/business/sys/database"
	"github.com/rmsj/service/business/sys/validate"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound  = errors.New("product not found")
	ErrInvalidID = errors.New("ID is not in its proper form")
)

// Core manages the set of APIs for product access.
type Core struct {
	log *zap.SugaredLogger
	db  *gorm.DB
}

// NewCore constructs a core for product api access.
func NewCore(l *zap.SugaredLogger, d *gorm.DB) Core {
	return Core{
		log: l,
		db:  d,
	}
}

// Create adds a Product to the database. It returns the created Product with
// fields like ID and DateCreated populated.
func (c Core) Create(ctx context.Context, np NewProduct, now time.Time) (Product, error) {
	if err := validate.Check(np); err != nil {
		return Product{}, fmt.Errorf("validating data: %w", err)
	}

	dbPrd := db.Product{
		ProductID:   validate.GenerateID(),
		Name:        np.Name,
		Cost:        np.Cost,
		Quantity:    np.Quantity,
		UserID:      np.UserID,
		DateCreated: now,
		DateUpdated: now,
	}

	var errDB error
	dbPrd, errDB = db.Create(ctx, c.db, dbPrd)

	if errDB != nil {
		return Product{}, fmt.Errorf("create: %w", errDB)
	}

	return toProduct(dbPrd), nil
}

// Update modifies data about a Product. It will error if the specified ID is
// invalid or does not reference an existing Product.
func (c Core) Update(ctx context.Context, productID string, up UpdateProduct, now time.Time) error {
	if err := validate.CheckID(productID); err != nil {
		return ErrInvalidID
	}

	if err := validate.Check(up); err != nil {
		return fmt.Errorf("validating data: %w", err)
	}

	dbPrd, err := db.QueryByID(ctx, c.db, productID)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("updating product productID[%s]: %w", productID, err)
	}

	if up.Name != nil {
		dbPrd.Name = *up.Name
	}
	if up.Cost != nil {
		dbPrd.Cost = *up.Cost
	}
	if up.Quantity != nil {
		dbPrd.Quantity = *up.Quantity
	}
	dbPrd.DateUpdated = now

	_, err = db.Update(ctx, c.db, dbPrd)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}

	return nil
}

// Delete removes the product identified by a given ID.
func (c Core) Delete(ctx context.Context, productID string) error {
	if err := validate.CheckID(productID); err != nil {
		return ErrInvalidID
	}

	if err := db.Delete(ctx, c.db, productID); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query gets all Products from the database.
func (c Core) Query(ctx context.Context, pageNumber int, rowsPerPage int) ([]Product, error) {
	dbPrds, err := db.Query(ctx, c.db, pageNumber, rowsPerPage)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return toProductSlice(dbPrds), nil
}

// QueryByID finds the product identified by a given ID.
func (c Core) QueryByID(ctx context.Context, productID string) (Product, error) {
	if err := validate.CheckID(productID); err != nil {
		return Product{}, ErrInvalidID
	}

	dbPrd, err := db.QueryByID(ctx, c.db, productID)
	if err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return Product{}, ErrNotFound
		}
		return Product{}, fmt.Errorf("query: %w", err)
	}

	return toProduct(dbPrd), nil
}

// QueryByUserID finds the products identified by a given User ID.
func (c Core) QueryByUserID(ctx context.Context, userID string) ([]Product, error) {
	if err := validate.CheckID(userID); err != nil {
		return nil, ErrInvalidID
	}

	dbPrds, err := db.QueryByUserID(ctx, c.db, userID)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return toProductSlice(dbPrds), nil
}
