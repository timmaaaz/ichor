// Package userbus provides business access to user domain.
package userbus

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/userapprovalstatusbus"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/business/sdk/order"
	"github.com/timmaaaz/ichor/business/sdk/page"
	"github.com/timmaaaz/ichor/business/sdk/sqldb"
	"github.com/timmaaaz/ichor/foundation/convert"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
	"golang.org/x/crypto/bcrypt"
)

// Set of error variables for CRUD operations.
var (
	ErrNotFound              = errors.New("user not found")
	ErrUniqueEmail           = errors.New("email is not unique")
	ErrAuthenticationFailure = errors.New("authentication failed")
)

// Storer interface declares the behavior this package needs to persist and
// retrieve data.
type Storer interface {
	NewWithTx(tx sqldb.CommitRollbacker) (Storer, error)
	Create(ctx context.Context, usr User) error
	Update(ctx context.Context, usr User) error
	Delete(ctx context.Context, usr User) error
	Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]User, error)
	Count(ctx context.Context, filter QueryFilter) (int, error)
	QueryByID(ctx context.Context, userID uuid.UUID) (User, error)
	QueryByEmail(ctx context.Context, email mail.Address) (User, error)
}

// Business manages the set of APIs for user access.
type Business struct {
	log      *logger.Logger
	storer   Storer
	delegate *delegate.Delegate
	uas      *userapprovalstatusbus.Business
}

// NewBusiness constructs a user business API for use.
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, uas *userapprovalstatusbus.Business, storer Storer) *Business {
	return &Business{
		log:      log,
		delegate: delegate,
		storer:   storer,
		uas:      uas,
	}
}

// NewWithTx constructs a new business value that will use the
// specified transaction in any store related calls.
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	bus := Business{
		log:      b.log,
		delegate: b.delegate,
		storer:   storer,
	}

	return &bus, nil
}

// Create adds a new user to the system.
func (b *Business) Create(ctx context.Context, nu NewUser) (User, error) {
	ctx, span := otel.AddSpan(ctx, "business.userbus.create")
	defer span.End()

	hash, err := bcrypt.GenerateFromPassword([]byte(nu.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, fmt.Errorf("generatefrompassword: %w", err)
	}

	var defaultStatus string = "PENDING"

	statuses, err := b.uas.Query(ctx, userapprovalstatusbus.QueryFilter{Name: &defaultStatus}, userapprovalstatusbus.DefaultOrderBy, page.MustParse("1", "1"))
	if err != nil {
		return User{}, fmt.Errorf("query userapprovalstatus pending: %w", err)
	}

	status := statuses[0]

	if status.Name != defaultStatus {
		return User{}, fmt.Errorf("default approval status not found: %w", err)
	}

	now := time.Now()

	usr := User{
		ID:                 uuid.New(),
		RequestedBy:        nu.RequestedBy,
		ApprovedBy:         uuid.Nil,
		UserApprovalStatus: status.ID,
		TitleID:            nu.TitleID,
		OfficeID:           nu.OfficeID,
		WorkPhoneID:        nu.WorkPhoneID,
		CellPhoneID:        nu.CellPhoneID,
		Username:           nu.Username,
		FirstName:          nu.FirstName,
		LastName:           nu.LastName,
		Email:              nu.Email,
		Birthday:           nu.Birthday,
		Roles:              nu.Roles,
		SystemRoles:        nu.SystemRoles,
		PasswordHash:       hash,
		Enabled:            nu.Enabled,
		DateHired:          time.Time{}, // Zero-value
		DateRequested:      now,
		DateApproved:       time.Time{}, // Zero-value
		DateCreated:        now,
		DateUpdated:        now,
	}

	if err := b.storer.Create(ctx, usr); err != nil {
		return User{}, fmt.Errorf("create: %w", err)
	}

	return usr, nil
}

// Update modifies information about a user.
func (b *Business) Update(ctx context.Context, usr User, uu UpdateUser) (User, error) {
	ctx, span := otel.AddSpan(ctx, "business.userbus.update")
	defer span.End()

	err := convert.PopulateSameTypes(uu, &usr)
	if err != nil {
		return User{}, fmt.Errorf("populate user from update user: %w", err)
	}

	if uu.Password != nil {
		pw, err := bcrypt.GenerateFromPassword([]byte(*uu.Password), bcrypt.DefaultCost)
		if err != nil {
			return User{}, fmt.Errorf("generatefrompassword: %w", err)
		}
		usr.PasswordHash = pw
	}

	usr.DateUpdated = time.Now()

	if err := b.storer.Update(ctx, usr); err != nil {
		return User{}, fmt.Errorf("update: %w", err)
	}

	// Other domains may need to know when a user is updated so business
	// logic can be applieb. This represents a delegate call to other domains.
	if err := b.delegate.Call(ctx, ActionUpdatedData(uu, usr.ID)); err != nil {
		return User{}, fmt.Errorf("failed to execute `%s` action: %w", ActionUpdated, err)
	}

	return usr, nil
}

// Delete removes the specified user.
func (b *Business) Delete(ctx context.Context, usr User) error {
	ctx, span := otel.AddSpan(ctx, "business.userbus.delete")
	defer span.End()

	if err := b.storer.Delete(ctx, usr); err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

// Query retrieves a list of existing users.
func (b *Business) Query(ctx context.Context, filter QueryFilter, orderBy order.By, page page.Page) ([]User, error) {
	ctx, span := otel.AddSpan(ctx, "business.userbus.query")
	defer span.End()

	users, err := b.storer.Query(ctx, filter, orderBy, page)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return users, nil
}

// Count returns the total number of users.
func (b *Business) Count(ctx context.Context, filter QueryFilter) (int, error) {
	ctx, span := otel.AddSpan(ctx, "business.userbus.count")
	defer span.End()

	return b.storer.Count(ctx, filter)
}

// QueryByID finds the user by the specified ID.
func (b *Business) QueryByID(ctx context.Context, userID uuid.UUID) (User, error) {
	ctx, span := otel.AddSpan(ctx, "business.userbus.querybyid")
	defer span.End()

	user, err := b.storer.QueryByID(ctx, userID)
	if err != nil {
		return User{}, fmt.Errorf("query: userID[%s]: %w", userID, err)
	}

	return user, nil
}

// QueryByEmail finds the user by a specified user email.
func (b *Business) QueryByEmail(ctx context.Context, email mail.Address) (User, error) {
	ctx, span := otel.AddSpan(ctx, "business.userbus.querybyemail")
	defer span.End()

	user, err := b.storer.QueryByEmail(ctx, email)
	if err != nil {
		return User{}, fmt.Errorf("query: email[%s]: %w", email, err)
	}

	return user, nil
}

// Authenticate finds a user by their email and verifies their passworb. On
// success it returns a Claims User representing this user. The claims can be
// used to generate a token for future authentication.
func (b *Business) Authenticate(ctx context.Context, email mail.Address, password string) (User, error) {
	ctx, span := otel.AddSpan(ctx, "business.userbus.authenticate")
	defer span.End()

	// TODO: Change this to work by username.
	usr, err := b.QueryByEmail(ctx, email)
	if err != nil {
		return User{}, fmt.Errorf("query: email[%s]: %w", email, err)
	}

	if err := bcrypt.CompareHashAndPassword(usr.PasswordHash, []byte(password)); err != nil {
		return User{}, fmt.Errorf("comparehashandpassword: %w", ErrAuthenticationFailure)
	}

	return usr, nil
}

func (b *Business) Approve(ctx context.Context, user User, approvedBy uuid.UUID) error {
	ctx, span := otel.AddSpan(ctx, "business.userbus.approve")
	defer span.End()

	var approvedStatus string = "APPROVED"

	statuses, err := b.uas.Query(ctx, userapprovalstatusbus.QueryFilter{Name: &approvedStatus}, userapprovalstatusbus.DefaultOrderBy, page.MustParse("1", "1"))
	if err != nil {
		return fmt.Errorf("query userapprovalstatus for approved: %w", err)
	}

	status := statuses[0]

	if status.Name != approvedStatus {
		return fmt.Errorf("approved userapprovalstatus not found: %w", err)
	}

	user.DateApproved = time.Now()
	user.ApprovedBy = approvedBy
	user.UserApprovalStatus = status.ID

	if err := b.storer.Update(ctx, user); err != nil {
		return fmt.Errorf("update: %w", err)
	}

	return nil
}
func (b *Business) Deny(ctx context.Context, user User) error {
	ctx, span := otel.AddSpan(ctx, "business.userbus.deny")
	defer span.End()

	var approvedStatus string = "DENIED"

	statuses, err := b.uas.Query(ctx, userapprovalstatusbus.QueryFilter{Name: &approvedStatus}, userapprovalstatusbus.DefaultOrderBy, page.MustParse("1", "1"))
	if err != nil {
		return fmt.Errorf("query userapprovalstatus for denied: %w", err)
	}

	status := statuses[0]

	if status.Name != approvedStatus {
		return fmt.Errorf("denied userapprovalstatus not found: %w", err)
	}

	user.UserApprovalStatus = status.ID

	if err := b.storer.Update(ctx, user); err != nil {
		return fmt.Errorf("update: %w", err)
	}

	return nil
}
func (b *Business) SetUnderReview(ctx context.Context, user User) error {
	ctx, span := otel.AddSpan(ctx, "business.userbus.setunderreview")
	defer span.End()

	var underReviewStatus string = "UNDER REVIEW"

	statuses, err := b.uas.Query(ctx, userapprovalstatusbus.QueryFilter{Name: &underReviewStatus}, userapprovalstatusbus.DefaultOrderBy, page.MustParse("1", "1"))
	if err != nil {
		return fmt.Errorf("query userapprovalstatus for under review: %w", err)
	}

	status := statuses[0]

	if status.Name != underReviewStatus {
		return fmt.Errorf("under review userapprovalstatus not found: %w", err)
	}

	user.UserApprovalStatus = status.ID

	if err := b.storer.Update(ctx, user); err != nil {
		return fmt.Errorf("update: %w", err)
	}

	return nil
}
