package user_test

import (
	"time"

	"github.com/timmaaaz/ichor/app/domain/users/userapp"
	"github.com/timmaaaz/ichor/business/domain/users/userbus"
)

func toAppUser(bus userbus.User) userapp.User {
	return userapp.User{
		ID:            bus.ID.String(),
		RequestedBy:   bus.RequestedBy.String(),
		ApprovedBy:    bus.ApprovedBy.String(),
		TitleID:       bus.TitleID.String(),
		OfficeID:      bus.OfficeID.String(),
		WorkPhoneID:   bus.WorkPhoneID.String(),
		CellPhoneID:   bus.CellPhoneID.String(),
		Username:      bus.Username.String(),
		FirstName:     bus.FirstName.String(),
		LastName:      bus.LastName.String(),
		Email:         bus.Email.Address,
		Birthday:      bus.Birthday.Format(time.RFC3339),
		Roles:         userbus.ParseRolesToString(bus.Roles),
		SystemRoles:   userbus.ParseRolesToString(bus.SystemRoles),
		PasswordHash:  bus.PasswordHash,
		Enabled:       bus.Enabled,
		DateHired:     bus.DateHired.Format(time.RFC3339),
		DateRequested: bus.DateRequested.Format(time.RFC3339),
		DateApproved:  bus.DateApproved.Format(time.RFC3339),
		CreatedDate:   bus.CreatedDate.Format(time.RFC3339),
		UpdatedDate:   bus.UpdatedDate.Format(time.RFC3339),
	}
}

func toAppUsers(users []userbus.User) []userapp.User {
	items := make([]userapp.User, len(users))
	for i, usr := range users {
		items[i] = toAppUser(usr)
	}

	return items
}

func toAppUserPtr(bus userbus.User) *userapp.User {
	appUsr := toAppUser(bus)
	return &appUsr
}
