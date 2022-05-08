// Package db contains user related CRUD functionality.
package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/rmsj/service/business/sys/database"
)

// Create inserts a new user into the database.
func Create(ctx context.Context, db *gorm.DB, usr User) (User, error) {

	if err := db.WithContext(ctx).Create(&usr).Error; err != nil {
		// Checks if the error is of code 23505 (unique_violation).
		if pqerr, ok := err.(*pq.Error); ok && pqerr.Code == database.UniqueViolation {
			return usr, database.ErrDBDuplicatedEntry
		}
		return usr, fmt.Errorf("inserting user: %w", err)
	}

	return usr, nil
}

// Update replaces a user document in the database.
func Update(ctx context.Context, db *gorm.DB, usr User) (User, error) {

	if err := db.WithContext(ctx).Save(&usr).Error; err != nil {
		// Checks if the error is of code 23505 (unique_violation).
		if pqerr, ok := err.(*pq.Error); ok && pqerr.Code == database.UniqueViolation {
			return usr, database.ErrDBDuplicatedEntry
		}
		return usr, fmt.Errorf("updating userID[%s]: %w", usr.UserID, err)
	}

	return usr, nil
}

// Delete removes a user from the database.
func Delete(ctx context.Context, db *gorm.DB, userID string) error {

	var usr User
	if err := db.WithContext(ctx).Where("user_id = ?", userID).Delete(&usr).Error; err != nil {
		return fmt.Errorf("deleting userID[%s]: %w", userID, err)
	}

	return nil
}

// Query retrieves a list of existing users from the database.
func Query(ctx context.Context, db *gorm.DB, pageNumber int, rowsPerPage int) ([]User, error) {

	offset := (pageNumber - 1) * rowsPerPage
	var users []User
	if err := db.WithContext(ctx).Limit(rowsPerPage).Offset(offset).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("selecting users: %w", err)
	}

	return users, nil
}

// QueryByID gets the specified user from the database.
func QueryByID(ctx context.Context, db *gorm.DB, userID string) (User, error) {

	var usr User
	sql := `user_id = ?`
	if err := db.WithContext(ctx).Where(sql, userID).First(&usr).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return User{}, database.ErrDBNotFound
		}
		return User{}, fmt.Errorf("selecting by id[%q]: %w", userID, err)
	}

	return usr, nil
}

// QueryByEmail gets the specified user from the database by email.
func QueryByEmail(ctx context.Context, db *gorm.DB, email string) (User, error) {

	var usr User
	sql := `email = ?`
	if err := db.WithContext(ctx).Where(sql, email).First(&usr).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return User{}, database.ErrDBNotFound
		}
		return User{}, fmt.Errorf("selecting email[%q]: %w", email, err)
	}

	return usr, nil
}
