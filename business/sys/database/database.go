// Package database provides support for access the database.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq" // Calls init function.
	"go.uber.org/zap"
	glogger "gorm.io/gorm/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/rmsj/service/foundation/web"
)

// lib/pq errorCodeNames
// https://github.com/lib/pq/blob/master/error.go#L178
const UniqueViolation = "23505"

const timezone = "utc"

// Set of error variables for CRUD operations.
var (
	ErrDBNotFound        = errors.New("not found")
	ErrDBDuplicatedEntry = errors.New("duplicated entry")
)

// Config is the required properties to use the database.
type Config struct {
	User            string
	Password        string
	Host            string
	Name            string
	Port            int
	MaxIdleConns    int
	MaxOpenConns    int
	MaxConnLifeTime int
	DisableTLS      bool
	Debug           bool
	Logger          *zap.SugaredLogger
}

type DBModeler interface {
	TableName() string
}

type Store struct {
	db  *gorm.DB
	log *zap.SugaredLogger
}

// Open knows how to open a database connection based on the configuration.
func Open(cfg Config) (*gorm.DB, error) {

	sslMode := "require"
	if cfg.DisableTLS {
		sslMode = "disable"
	}

	// we need our own logger to be able to centralize everything with zap
	myLogger := newLogger(cfg.Logger)
	myLogger.SlowThreshold = 200 * time.Millisecond
	myLogger.LogLevel = glogger.Warn
	if cfg.Debug {
		myLogger.LogLevel = glogger.Info
	}
	gormCfg := gorm.Config{
		Logger: myLogger,
	}

	//dsnStr := "host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s"
	//dns := fmt.Sprintf(dsnStr, cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, sslMode, timezone)
	dsnStr := "host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s"
	dns := fmt.Sprintf(dsnStr, cfg.Host, cfg.User, cfg.Password, cfg.Name, cfg.Port, sslMode, timezone)
	db, err := gorm.Open(postgres.Open(dns), &gormCfg)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()

	if err != nil {
		return nil, err
	}

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// Close closes connection with database
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	sqlDB.Close()
	return nil
}

// StatusCheck returns nil if it can successfully talk to the database. It
// returns a non-nil error otherwise.
func StatusCheck(ctx context.Context, db *gorm.DB) error {

	// Get generic database object sql.DB to use its functions
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// First check we can ping the database.
	var pingError error
	for attempts := 1; ; attempts++ {

		// Ping
		pingError = sqlDB.Ping()
		if pingError == nil {
			break
		}
		time.Sleep(time.Duration(attempts) * 100 * time.Millisecond)
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	// Make sure we didn't timeout or be cancelled.
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Run a simple query to determine connectivity. Running this query forces a
	// round trip through the database.
	const q = `SELECT true`
	var tmp bool
	return db.Raw(q).Scan(&tmp).Error
}

// Transaction runs passed function and do commit/rollback at the end.
func Transaction(ctx context.Context, db *gorm.DB, log *zap.SugaredLogger, fn func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	traceID := web.GetTraceID(ctx)

	// Begin the transaction.
	log.Infow("begin tran", "traceid", traceID)

	// any error will automatically rollback the transaction
	err := db.Transaction(fn, opts...)

	// Execute the code inside the transaction. If the function
	// fails, return the error and the defer function will roll back.
	if err != nil {

		// Checks if the error is of code 23505 (unique_violation).
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == UniqueViolation {
			return ErrDBDuplicatedEntry
		}
		return fmt.Errorf("exec tran: %w", err)
	}

	return nil
}

// ExecDDL is a helper function to execute a CUD operation with
// logging and tracing.
func ExecDDL(ctx context.Context, db *gorm.DB, log *zap.SugaredLogger, sql string, data any) error {

	log.Infow("database.ExecSQL", "traceid", web.GetTraceID(ctx), "query", sql)

	var dbErr error
	if data != nil {
		dbErr = db.Exec(sql, data).Error
	} else {
		dbErr = db.Exec(sql).Error
	}

	if dbErr != nil {
		return dbErr
	}

	return nil
}

// ExecSQL is a helper function to execute a CUD operation with
// logging and tracing.
func ExecSQL(ctx context.Context, db *gorm.DB, log *zap.SugaredLogger, sql string, data any) error {

	sqlStr := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		if data != nil {
			return db.Exec(sql, data)
		} else {
			return db.Exec(sql)
		}
	})
	log.Infow("database.ExecSQL", "traceid", web.GetTraceID(ctx), "query", sqlStr)

	var dbErr error
	if data != nil {
		dbErr = db.Exec(sql, data).Error
	} else {
		dbErr = db.Exec(sql).Error
	}

	if dbErr != nil {
		// Checks if the error is of code 23505 (unique_violation).
		if pqerr, ok := dbErr.(*pq.Error); ok && pqerr.Code == UniqueViolation {
			return ErrDBDuplicatedEntry
		}
		return dbErr
	}

	return nil
}

// Create is a helper function to execute a CUD operation with
// logging and tracing.
func Create(ctx context.Context, db *gorm.DB, log *zap.SugaredLogger, model DBModeler) error {

	sqlStr := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.WithContext(ctx).Create(&model)
	})

	log.Infow("database.Exec", "traceid", web.GetTraceID(ctx), "insert", sqlStr)

	if err := db.WithContext(ctx).Create(&model).Error; err != nil {
		// Checks if the error is of code 23505 (unique_violation).
		if pqerr, ok := err.(*pq.Error); ok && pqerr.Code == UniqueViolation {
			return ErrDBDuplicatedEntry
		}
		return err
	}

	return nil
}

// Update is a helper function to execute a CUD operation with
// logging and tracing.
func Update(ctx context.Context, db *gorm.DB, log *zap.SugaredLogger, model DBModeler) error {

	sqlStr := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.WithContext(ctx).Save(&model)
	})

	log.Infow("database.Exec", "traceid", web.GetTraceID(ctx), "update", sqlStr)

	if err := db.WithContext(ctx).Save(&model).Error; err != nil {
		// Checks if the error is of code 23505 (unique_violation).
		if pqerr, ok := err.(*pq.Error); ok && pqerr.Code == UniqueViolation {
			return ErrDBDuplicatedEntry
		}
		return err
	}

	return nil
}

// Delete is a helper function to execute a CUD operation with
// logging and tracing.
func Delete(ctx context.Context, db *gorm.DB, log *zap.SugaredLogger, model DBModeler) error {

	sqlStr := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.WithContext(ctx).Delete(&model)
	})

	log.Infow("database.Exec", "traceid", web.GetTraceID(ctx), "query", sqlStr)

	return db.WithContext(ctx).Save(&model).Error
}

// GetOne is a helper function for executing queries that return a
// single value to be unmarshalled into a struct type.
func GetOne(ctx context.Context, db *gorm.DB, log *zap.SugaredLogger, dest DBModeler, query string, data ...any) error {
	sqlStr := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Where(query, data...).First(dest)
	})

	log.Infow("database.GetOne", "traceid", web.GetTraceID(ctx), "query", sqlStr)

	err := db.Where(query, data...).First(dest).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDBNotFound
		}
		return err
	}

	return nil
}

// GetMany is a helper function for executing queries that return a
// collection of data to be unmarshalled into a slice.
func GetMany(ctx context.Context, db *gorm.DB, log *zap.SugaredLogger, dest any, query string, data ...any) error {

	sqlStr := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Where(query, data...).Find(dest)
	})
	log.Infow("database.GetMany", "traceid", web.GetTraceID(ctx), "query", sqlStr)

	err := db.Where(query, data...).Find(&dest).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDBNotFound
		}
		return err
	}

	return nil
}
