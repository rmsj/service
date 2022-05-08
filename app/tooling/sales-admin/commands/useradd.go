package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/rmsj/service/business/core/user"
	"github.com/rmsj/service/business/sys/database"
	"github.com/rmsj/service/business/web/auth"
)

// UserAdd adds new users into the database.
func UserAdd(log *zap.SugaredLogger, cfg database.Config, name, email, password string) error {
	if name == "" || email == "" || password == "" {
		fmt.Println("help: useradd <name> <email> <password>")
		return ErrHelp
	}

	db, err := database.Open(cfg)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer database.Close(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	core := user.NewCore(log, db)

	nu := user.NewUser{
		Name:            name,
		Email:           email,
		Password:        password,
		PasswordConfirm: password,
		Roles:           pq.StringArray{auth.RoleAdmin, auth.RoleUser},
	}

	usr, err := core.Create(ctx, nu, time.Now())
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	fmt.Println("user id:", usr.ID)
	return nil
}
