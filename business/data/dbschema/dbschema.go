// Package dbschema contains the database schema, migrations and seeding data.
package dbschema

import (
	"context"
	_ "embed" // Calls init function.
	"fmt"

	"github.com/ardanlabs/darwin"
	"gorm.io/gorm"

	"github.com/rmsj/service/business/sys/database"
)

var (
	//go:embed sql/schema.sql
	schemaDoc string

	//go:embed sql/seed.sql
	seedDoc string

	//go:embed sql/delete.sql
	deleteDoc string
)

// Migrate attempts to bring the schema for db up to date with the migrations
// defined in this package.
func Migrate(ctx context.Context, db *gorm.DB) error {
	if err := database.StatusCheck(ctx, db); err != nil {
		return fmt.Errorf("status check database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	driver, err := darwin.NewGenericDriver(sqlDB, darwin.PostgresDialect{})
	if err != nil {
		return fmt.Errorf("construct darwin driver: %w", err)
	}

	d := darwin.New(driver, darwin.ParseMigrations(schemaDoc))
	return d.Migrate()
}

// Seed runs the set of seed-data queries against db. The queries are ran in a
// transaction and rolled back if any fail.
func Seed(ctx context.Context, db *gorm.DB) error {
	if err := database.StatusCheck(ctx, db); err != nil {
		return fmt.Errorf("status check database: %w", err)
	}

	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Exec(seedDoc).Error; err != nil {
		if err := tx.Rollback().Error; err != nil {
			return err
		}
		return err
	}

	return tx.Commit().Error
}

// DeleteAll runs the set of Drop-table queries against db. The queries are ran in a
// transaction and rolled back if any fail.
func DeleteAll(db *gorm.DB) error {
	tx := db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Exec(deleteDoc).Error; err != nil {
		if err := tx.Rollback().Error; err != nil {
			return err
		}
		return err
	}

	return tx.Commit().Error
}
