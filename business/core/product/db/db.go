// Package db contains product related CRUD functionality.
package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/rmsj/service/business/sys/database"
)

// Create adds a Product to the database. It returns the created Product with
// fields like ID and DateCreated populated.
func Create(ctx context.Context, db *gorm.DB, prd Product) (Product, error) {
	if err := db.WithContext(ctx).Create(&prd).Error; err != nil {
		// Checks if the error is of code 23505 (unique_violation).
		if pqerr, ok := err.(*pq.Error); ok && pqerr.Code == database.UniqueViolation {
			return prd, database.ErrDBDuplicatedEntry
		}
		return prd, fmt.Errorf("inserting user: %w", err)
	}

	return prd, nil
}

// Update modifies data about a Product. It will error if the specified ID is
// invalid or does not reference an existing Product.
func Update(ctx context.Context, db *gorm.DB, prd Product) (Product, error) {
	if err := db.WithContext(ctx).Save(&prd).Error; err != nil {
		// Checks if the error is of code 23505 (unique_violation).
		if pqerr, ok := err.(*pq.Error); ok && pqerr.Code == database.UniqueViolation {
			return prd, database.ErrDBDuplicatedEntry
		}
		return prd, fmt.Errorf("updating productID[%s]: %w", prd.ProductID, err)
	}

	return prd, nil
}

// Delete removes the product identified by a given ID.
func Delete(ctx context.Context, db *gorm.DB, productID string) error {
	var prd Product
	if err := db.WithContext(ctx).Where("product_id = ?", productID).Delete(&prd).Error; err != nil {
		return fmt.Errorf("deleting productID[%s]: %w", productID, err)
	}

	return nil
}

// Query gets all Products from the database.
func Query(ctx context.Context, db *gorm.DB, pageNumber int, rowsPerPage int) ([]Product, error) {
	offset := (pageNumber - 1) * rowsPerPage
	var prds []Product
	if err := db.WithContext(ctx).Limit(rowsPerPage).Offset(offset).Find(&prds).Error; err != nil {
		return nil, fmt.Errorf("selecting users: %w", err)
	}

	return prds, nil
}

// QueryByID finds the product identified by a given ID.
func QueryByID(ctx context.Context, db *gorm.DB, productID string) (Product, error) {
	var prd Product
	sql := `product_id = ?`
	if err := db.WithContext(ctx).Where(sql, productID).First(&prd).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Product{}, database.ErrDBNotFound
		}
		return Product{}, fmt.Errorf("selecting by id[%q]: %w", productID, err)
	}

	return prd, nil
}

// QueryByUserID finds the product identified by a given User ID.
func QueryByUserID(ctx context.Context, db *gorm.DB, userID string) ([]Product, error) {
	var prds []Product
	if err := db.WithContext(ctx).Find(&prds).Error; err != nil {
		return nil, fmt.Errorf("selecting products userID[%s]: %w", userID, err)
	}

	return prds, nil
}
