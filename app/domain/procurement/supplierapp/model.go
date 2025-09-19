package supplierapp

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus"
	"github.com/timmaaaz/ichor/business/domain/procurement/supplierbus/types"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

type QueryParams struct {
	Page    string
	Rows    string
	OrderBy string

	SupplierID     string
	ContactInfosID string
	Name           string
	PaymentTerms   string
	LeadTimeDays   string
	Rating         string
	IsActive       string
	CreatedDate    string
	UpdatedDate    string
}

type Supplier struct {
	SupplierID     string `json:"supplier_id"`
	ContactInfosID string `json:"contact_infos_id"`
	Name           string `json:"name"`
	PaymentTerms   string `json:"payment_terms"`
	LeadTimeDays   string `json:"lead_time_days"`
	Rating         string `json:"rating"`
	IsActive       string `json:"is_active"`
	CreatedDate    string `json:"created_date"`
	UpdatedDate    string `json:"updated_date"`
}

func (app Supplier) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

func ToAppSupplier(bus supplierbus.Supplier) Supplier {
	return Supplier{
		SupplierID:     bus.SupplierID.String(),
		ContactInfosID: bus.ContactInfosID.String(),
		Name:           bus.Name,
		PaymentTerms:   bus.PaymentTerms,
		LeadTimeDays:   fmt.Sprintf("%d", bus.LeadTimeDays),
		Rating:         bus.Rating.String(),
		IsActive:       fmt.Sprintf("%t", bus.IsActive),
		CreatedDate:    bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:    bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

func ToAppSuppliers(bus []supplierbus.Supplier) []Supplier {
	app := make([]Supplier, len(bus))
	for i, v := range bus {
		app[i] = ToAppSupplier(v)
	}
	return app
}

type NewSupplier struct {
	ContactInfosID string `json:"contact_infos_id" validate:"required,min=36,max=36"`
	Name           string `json:"name" validate:"required"`
	PaymentTerms   string `json:"payment_terms" validate:"required"`
	LeadTimeDays   string `json:"lead_time_days" validate:"required"`
	Rating         string `json:"rating" validate:"required"`
	IsActive       string `json:"is_active" validate:"required"`
}

func (app *NewSupplier) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app NewSupplier) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewSupplier(app NewSupplier) (supplierbus.NewSupplier, error) {
	leadTimeDays, err := strconv.Atoi(app.LeadTimeDays)
	if err != nil {
		return supplierbus.NewSupplier{}, fmt.Errorf("invalid lead time days: %w", err)
	}

	ContactInfosID, err := uuid.Parse(app.ContactInfosID)
	if err != nil {
		return supplierbus.NewSupplier{}, fmt.Errorf("parse: %w", err)
	}

	rating, err := types.ParseRoundedFloat(app.Rating)
	if err != nil {
		return supplierbus.NewSupplier{}, fmt.Errorf("invalid rating: %w", err)
	}

	isActive, err := strconv.ParseBool(app.IsActive)
	if err != nil {
		return supplierbus.NewSupplier{}, fmt.Errorf("invalid is active: %w", err)
	}

	return supplierbus.NewSupplier{
		ContactInfosID: ContactInfosID,
		Name:           app.Name,
		PaymentTerms:   app.PaymentTerms,
		LeadTimeDays:   leadTimeDays,
		Rating:         rating,
		IsActive:       isActive,
	}, nil
}

type UpdateSupplier struct {
	ContactInfosID *string `json:"contact_infos_id" validate:"omitempty,min=36,max=36"`
	Name           *string `json:"name" validate:"omitempty"`
	PaymentTerms   *string `json:"payment_terms" validate:"omitempty"`
	LeadTimeDays   *string `json:"lead_time_days" validate:"omitempty"`
	Rating         *string `json:"rating" validate:"omitempty"`
	IsActive       *string `json:"is_active" validate:"omitempty"`
}

func (app *UpdateSupplier) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app UpdateSupplier) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateSupplier(app UpdateSupplier) (supplierbus.UpdateSupplier, error) {

	dest := supplierbus.UpdateSupplier{}

	if app.ContactInfosID != nil {
		ContactInfosID, err := uuid.Parse(*app.ContactInfosID)
		if err != nil {
			return supplierbus.UpdateSupplier{}, errs.NewFieldsError("contact_infos_id", err)
		}
		dest.ContactInfosID = &ContactInfosID
	}

	if app.Name != nil {
		dest.Name = app.Name
	}

	if app.PaymentTerms != nil {
		dest.PaymentTerms = app.PaymentTerms
	}

	if app.LeadTimeDays != nil {
		leadTimeDays, err := strconv.Atoi(*app.LeadTimeDays)
		if err != nil {
			return supplierbus.UpdateSupplier{}, errs.NewFieldsError("lead_time_days", err)
		}
		dest.LeadTimeDays = &leadTimeDays
	}

	if app.Rating != nil {
		rating, err := types.ParseRoundedFloat(*app.Rating)
		if err != nil {
			return supplierbus.UpdateSupplier{}, errs.NewFieldsError("rating", err)
		}
		dest.Rating = &rating
	}

	if app.IsActive != nil {
		isActive, err := strconv.ParseBool(*app.IsActive)
		if err != nil {
			return supplierbus.UpdateSupplier{}, errs.NewFieldsError("is_active", err)
		}
		dest.IsActive = &isActive
	}

	return dest, nil
}
