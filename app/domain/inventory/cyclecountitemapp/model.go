package cyclecountitemapp

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/inventory/cyclecountitembus"
	"github.com/timmaaaz/ichor/foundation/timeutil"
)

// QueryParams holds the raw query parameters from the HTTP request.
type QueryParams struct {
	Page       string
	Rows       string
	OrderBy    string
	ID         string
	SessionID  string
	ProductID  string
	LocationID string
	Status     string
	CountedBy  string
}

// =============================================================================
// Response model
// =============================================================================

// CycleCountItem is the app-layer response model. All fields are strings for JSON serialization.
type CycleCountItem struct {
	ID              string `json:"id"`
	SessionID       string `json:"sessionId"`
	ProductID       string `json:"productId"`
	LocationID      string `json:"locationId"`
	SystemQuantity  string `json:"systemQuantity"`
	CountedQuantity string `json:"countedQuantity"`
	Variance        string `json:"variance"`
	Status          string `json:"status"`
	CountedBy       string `json:"countedBy"`
	CountedDate     string `json:"countedDate"`
	CreatedDate     string `json:"createdDate"`
	UpdatedDate     string `json:"updatedDate"`
}

func (app CycleCountItem) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToAppCycleCountItem converts a bus model to an app-layer response model.
func ToAppCycleCountItem(bus cyclecountitembus.CycleCountItem) CycleCountItem {
	countedQuantity := ""
	if bus.CountedQuantity != nil {
		countedQuantity = strconv.Itoa(*bus.CountedQuantity)
	}

	variance := ""
	if bus.Variance != nil {
		variance = strconv.Itoa(*bus.Variance)
	}

	countedBy := ""
	if bus.CountedBy != uuid.Nil {
		countedBy = bus.CountedBy.String()
	}

	countedDate := ""
	if !bus.CountedDate.IsZero() {
		countedDate = bus.CountedDate.Format(timeutil.FORMAT)
	}

	return CycleCountItem{
		ID:              bus.ID.String(),
		SessionID:       bus.SessionID.String(),
		ProductID:       bus.ProductID.String(),
		LocationID:      bus.LocationID.String(),
		SystemQuantity:  strconv.Itoa(bus.SystemQuantity),
		CountedQuantity: countedQuantity,
		Variance:        variance,
		Status:          bus.Status.String(),
		CountedBy:       countedBy,
		CountedDate:     countedDate,
		CreatedDate:     bus.CreatedDate.Format(timeutil.FORMAT),
		UpdatedDate:     bus.UpdatedDate.Format(timeutil.FORMAT),
	}
}

// ToAppCycleCountItems converts a slice of bus models to app-layer response models.
func ToAppCycleCountItems(bus []cyclecountitembus.CycleCountItem) []CycleCountItem {
	app := make([]CycleCountItem, len(bus))
	for i, v := range bus {
		app[i] = ToAppCycleCountItem(v)
	}
	return app
}

// =============================================================================
// Create model
// =============================================================================

// NewCycleCountItem is the app-layer create request model.
type NewCycleCountItem struct {
	SessionID      string `json:"sessionId" validate:"required,min=36,max=36"`
	ProductID      string `json:"productId" validate:"required,min=36,max=36"`
	LocationID     string `json:"locationId" validate:"required,min=36,max=36"`
	SystemQuantity string `json:"systemQuantity" validate:"required"`
}

func (app *NewCycleCountItem) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

func (app NewCycleCountItem) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusNewCycleCountItem(app NewCycleCountItem) (cyclecountitembus.NewCycleCountItem, error) {
	sessionID, err := uuid.Parse(app.SessionID)
	if err != nil {
		return cyclecountitembus.NewCycleCountItem{}, errs.Newf(errs.InvalidArgument, "parse sessionID: %s", err)
	}

	productID, err := uuid.Parse(app.ProductID)
	if err != nil {
		return cyclecountitembus.NewCycleCountItem{}, errs.Newf(errs.InvalidArgument, "parse productID: %s", err)
	}

	locationID, err := uuid.Parse(app.LocationID)
	if err != nil {
		return cyclecountitembus.NewCycleCountItem{}, errs.Newf(errs.InvalidArgument, "parse locationID: %s", err)
	}

	systemQuantity, err := strconv.Atoi(app.SystemQuantity)
	if err != nil {
		return cyclecountitembus.NewCycleCountItem{}, errs.Newf(errs.InvalidArgument, "parse systemQuantity: %s", err)
	}

	return cyclecountitembus.NewCycleCountItem{
		SessionID:      sessionID,
		ProductID:      productID,
		LocationID:     locationID,
		SystemQuantity: systemQuantity,
	}, nil
}

// =============================================================================
// Update model
// =============================================================================

// UpdateCycleCountItem is the app-layer update request model.
type UpdateCycleCountItem struct {
	CountedQuantity *string `json:"countedQuantity" validate:"omitempty"`
	Status          *string `json:"status" validate:"omitempty"`
}

func (app *UpdateCycleCountItem) Decode(data []byte) error {
	return json.Unmarshal(data, app)
}

func (app UpdateCycleCountItem) Validate() error {
	if err := errs.Check(app); err != nil {
		return errs.Newf(errs.InvalidArgument, "validate: %s", err)
	}
	return nil
}

func toBusUpdateCycleCountItem(app UpdateCycleCountItem, userID uuid.UUID) (cyclecountitembus.UpdateCycleCountItem, error) {
	bus := cyclecountitembus.UpdateCycleCountItem{}

	if app.CountedQuantity != nil {
		q, err := strconv.Atoi(*app.CountedQuantity)
		if err != nil {
			return cyclecountitembus.UpdateCycleCountItem{}, errs.Newf(errs.InvalidArgument, "parse countedQuantity: %s", err)
		}
		bus.CountedQuantity = &q

		// Auto-inject CountedBy and CountedDate when CountedQuantity is set.
		now := time.Now()
		bus.CountedBy = &userID
		bus.CountedDate = &now
	}

	if app.Status != nil {
		st, err := cyclecountitembus.ParseStatus(*app.Status)
		if err != nil {
			return cyclecountitembus.UpdateCycleCountItem{}, errs.Newf(errs.InvalidArgument, "parse status: %s", err)
		}
		bus.Status = &st
	}

	return bus, nil
}

// =============================================================================
// Filter
// =============================================================================

func parseFilter(qp QueryParams) (cyclecountitembus.QueryFilter, error) {
	var filter cyclecountitembus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return cyclecountitembus.QueryFilter{}, fmt.Errorf("parse id: %w", err)
		}
		filter.ID = &id
	}

	if qp.SessionID != "" {
		id, err := uuid.Parse(qp.SessionID)
		if err != nil {
			return cyclecountitembus.QueryFilter{}, fmt.Errorf("parse sessionID: %w", err)
		}
		filter.SessionID = &id
	}

	if qp.ProductID != "" {
		id, err := uuid.Parse(qp.ProductID)
		if err != nil {
			return cyclecountitembus.QueryFilter{}, fmt.Errorf("parse productID: %w", err)
		}
		filter.ProductID = &id
	}

	if qp.LocationID != "" {
		id, err := uuid.Parse(qp.LocationID)
		if err != nil {
			return cyclecountitembus.QueryFilter{}, fmt.Errorf("parse locationID: %w", err)
		}
		filter.LocationID = &id
	}

	if qp.Status != "" {
		st, err := cyclecountitembus.ParseStatus(qp.Status)
		if err != nil {
			return cyclecountitembus.QueryFilter{}, fmt.Errorf("parse status: %w", err)
		}
		filter.Status = &st
	}

	if qp.CountedBy != "" {
		id, err := uuid.Parse(qp.CountedBy)
		if err != nil {
			return cyclecountitembus.QueryFilter{}, fmt.Errorf("parse countedBy: %w", err)
		}
		filter.CountedBy = &id
	}

	return filter, nil
}
