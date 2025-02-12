package tranapp

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/productbus"
	"github.com/timmaaaz/ichor/business/domain/userbus"
)

// Product represents an individual product.
type Product struct {
	ID          string  `json:"id"`
	UserID      string  `json:"userID"`
	Name        string  `json:"name"`
	Cost        float64 `json:"cost"`
	Quantity    int     `json:"quantity"`
	DateCreated string  `json:"dateCreated"`
	DateUpdated string  `json:"dateUpdated"`
}

// Encode implements the encoder interface.
func (app Product) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func toAppProduct(prd productbus.Product) Product {
	return Product{
		ID:          prd.ID.String(),
		UserID:      prd.UserID.String(),
		Name:        prd.Name.String(),
		Cost:        prd.Cost,
		Quantity:    prd.Quantity,
		DateCreated: prd.DateCreated.Format(time.RFC3339),
		DateUpdated: prd.DateUpdated.Format(time.RFC3339),
	}
}

// =============================================================================

// NewTran represents an example of cross domain transaction at the
// application layer.
type NewTran struct {
	Product NewProduct `json:"product"`
	User    NewUser    `json:"user"`
}

// Validate checks the data in the model is considered clean.
func (app NewTran) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

// Decode implements the decoder interface.
func (app *NewTran) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

// =============================================================================

// NewUser contains information needed to create a new user.
// NewUser defines the data needed to add a new user.
type NewUser struct {
	RequestedBy     string   `json:"requestedBy" validate:"omitempty"` // we might be able to infer this instead of taking it as an argument
	TitleID         string   `json:"titleID" validate:"omitempty"`
	OfficeID        string   `json:"officeID" validate:"omitempty"`
	WorkPhoneID     string   `json:"workPhoneID" validate:"omitempty"`
	CellPhoneID     string   `json:"cellPhoneID" validate:"omitempty"`
	Username        string   `json:"username" validate:"required"`
	FirstName       string   `json:"firstName" validate:"required"`
	LastName        string   `json:"lastName" validate:"required"`
	Email           string   `json:"email" validate:"required,email"`
	Birthday        string   `json:"birthday" validate:"required"`
	Roles           []string `json:"roles" validate:"required"`
	SystemRoles     []string `json:"systemRoles" validate:"required"`
	Password        string   `json:"password" validate:"required"`
	PasswordConfirm string   `json:"passwordConfirm" validate:"eqfield=Password"`
	Enabled         bool     `json:"enabled"`
}

// Validate checks the data in the model is considered clean.
func (app NewUser) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusNewUser(app NewUser) (userbus.NewUser, error) {
	requestedBy, err := uuid.Parse(app.RequestedBy)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	titleID, err := uuid.Parse(app.TitleID)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	officeID, err := uuid.Parse(app.OfficeID)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	workPhoneID, err := uuid.Parse(app.WorkPhoneID)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	cellPhoneID, err := uuid.Parse(app.CellPhoneID)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	username, err := userbus.ParseName(app.Username)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	firstName, err := userbus.ParseName(app.FirstName)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	lastName, err := userbus.ParseName(app.LastName)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	addr, err := mail.ParseAddress(app.Email)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	birthday, err := time.Parse(time.RFC3339, app.Birthday)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	roles, err := userbus.ParseRoles(app.Roles)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	systemRoles, err := userbus.ParseRoles(app.Roles)
	if err != nil {
		return userbus.NewUser{}, fmt.Errorf("parse: %w", err)
	}

	bus := userbus.NewUser{
		RequestedBy: requestedBy,
		TitleID:     titleID,
		OfficeID:    officeID,
		WorkPhoneID: workPhoneID,
		CellPhoneID: cellPhoneID,
		Username:    username,
		FirstName:   firstName,
		LastName:    lastName,
		Email:       *addr,
		Birthday:    birthday,
		Roles:       roles,
		SystemRoles: systemRoles,
		Password:    app.Password,
		Enabled:     app.Enabled,
	}

	return bus, nil
}

// =============================================================================

// NewProduct is what we require from clients when adding a Product.
type NewProduct struct {
	Name     string  `json:"name" validate:"required"`
	Cost     float64 `json:"cost" validate:"required,gte=0"`
	Quantity int     `json:"quantity" validate:"required,gte=1"`
}

// Validate checks the data in the model is considered clean.
func (app NewProduct) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}

	return nil
}

func toBusNewProduct(app NewProduct) (productbus.NewProduct, error) {
	name, err := productbus.ParseName(app.Name)
	if err != nil {
		return productbus.NewProduct{}, fmt.Errorf("parse: %w", err)
	}

	bus := productbus.NewProduct{
		Name:     name,
		Cost:     app.Cost,
		Quantity: app.Quantity,
	}

	return bus, nil
}
