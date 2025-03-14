package commands

import (
	"context"
	"fmt"
	"net/mail"
	"time"

	"github.com/timmaaaz/ichor/business/domain/users/status/approvalbus"
	"github.com/timmaaaz/ichor/business/domain/users/status/approvalbus/stores/approvaldb"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
	"github.com/timmaaaz/ichor/business/domain/users/userbus/stores/userdb"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// UserAdd adds new users into the database.
func UserAdd(log *logger.Logger, cfg sqldb.Config, username, firstName, lastName, email, password string) error {
	if username == "" || firstName == "" || lastName == "" || email == "" || password == "" {
		fmt.Println("help: useradd <username> <firstname> <lastname> <email> <password>")
		return ErrHelp
	}

	db, err := sqldb.Open(cfg)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userApprovalStatusBus := approvalbus.NewBusiness(log, nil, approvaldb.NewStore(log, db))
	userBus := userbus.NewBusiness(log, nil, userApprovalStatusBus, userdb.NewStore(log, db))

	addr, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("parsing email: %w", err)
	}

	nu := userbus.NewUser{
		Username:  userbus.MustParseName(username),
		FirstName: userbus.MustParseName(firstName),
		LastName:  userbus.MustParseName(lastName),
		Email:     *addr,
		Password:  password,
		Roles:     []userbus.Role{userbus.Roles.Admin, userbus.Roles.User},
		Enabled:   true,
	}

	usr, err := userBus.Create(ctx, nu)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	fmt.Println("user id:", usr.ID)
	return nil
}
