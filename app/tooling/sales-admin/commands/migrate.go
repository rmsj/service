package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rmsj/service/business/data/dbschema"
	"github.com/rmsj/service/business/sys/database"
)

// ErrHelp provides context that help was given.
var ErrHelp = errors.New("provided help")

// Migrate creates the schema in the database.
func Migrate(cfg database.Config) error {
	db, err := database.Open(cfg)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer database.Close(db)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := dbschema.Migrate(ctx, db); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	fmt.Println("migrations complete")
	return nil
}
