package unitest

import (
	"context"

	"github.com/rmsj/service/business/domain/authbus"
	"github.com/rmsj/service/business/domain/productbus"
	"github.com/rmsj/service/business/domain/userbus"
)

// User represents an app user specified for the test.
type User struct {
	userbus.User
	Products []productbus.Product
}

// SeedData represents data that was seeded for the test.
type SeedData struct {
	Users           []User
	Admins          []User
	PassResetTokens []authbus.PasswordResetToken
}

// Table represent fields needed for running an unit test.
type Table struct {
	Name    string
	ExpResp any
	ExcFunc func(ctx context.Context) any
	CmpFunc func(got any, exp any) string
}
