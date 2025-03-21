package userbus

import (
	"context"
	"fmt"
	"math/rand"
	"net/mail"

	"github.com/google/uuid"
)

// TestNewUsers is a helper method for testing.
func TestNewUsers(n int, role Role, requestedBy, titleID, officeID *uuid.UUIDs) []NewUser {
	newUsrs := make([]NewUser, n)

	if requestedBy == nil {
		idx := rand.Intn(10000)
		for i := 0; i < n; i++ {
			idx++

			nu := NewUser{
				Username:    Name{fmt.Sprintf("Username%d", idx)},
				FirstName:   Name{fmt.Sprintf("FirstName%d", idx)},
				LastName:    Name{fmt.Sprintf("LastName%d", idx)},
				Email:       mail.Address{Address: fmt.Sprintf("Email%d@gmail.com", idx)},
				Roles:       []Role{role},
				SystemRoles: []Role{role},
				Password:    fmt.Sprintf("Password%d", idx),
				Enabled:     true,
			}

			newUsrs[i] = nu
		}
	} else {
		idx := rand.Intn(10000)
		for i := 0; i < n; i++ {
			idx++

			nu := NewUser{
				Username:    Name{fmt.Sprintf("Username%d", idx)},
				FirstName:   Name{fmt.Sprintf("FirstName%d", idx)},
				LastName:    Name{fmt.Sprintf("LastName%d", idx)},
				Email:       mail.Address{Address: fmt.Sprintf("Email%d@gmail.com", idx)},
				Roles:       []Role{role},
				SystemRoles: []Role{role},
				Password:    fmt.Sprintf("Password%d", idx),
				Enabled:     true,
				RequestedBy: (*requestedBy)[i%len(*requestedBy)],
				TitleID:     (*titleID)[i%len(*titleID)],
				OfficeID:    (*officeID)[i%len(*officeID)],
			}

			newUsrs[i] = nu
		}
	}

	return newUsrs
}

// TestSeedUsers is a helper method for testing.
func TestSeedUsers(ctx context.Context, n int, role Role, requestedBy, titleID, officeID uuid.UUIDs, api *Business) ([]User, error) {
	newUsrs := TestNewUsers(n, role, &requestedBy, &titleID, &officeID)

	usrs := make([]User, len(newUsrs))
	for i, nu := range newUsrs {
		usr, err := api.Create(ctx, nu)
		if err != nil {
			return nil, fmt.Errorf("seeding user: idx: %d : %w", i, err)
		}

		usrs[i] = usr
	}

	return usrs, nil
}

func TestSeedUsersWithNoFKs(ctx context.Context, n int, role Role, api *Business) ([]User, error) {
	newUsrs := TestNewUsers(n, role, nil, nil, nil)

	usrs := make([]User, len(newUsrs))
	for i, nu := range newUsrs {
		usr, err := api.Create(ctx, nu)
		if err != nil {
			return nil, fmt.Errorf("seeding user: idx: %d : %w", i, err)
		}

		usrs[i] = usr
	}

	return usrs, nil
}
