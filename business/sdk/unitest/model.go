package unitest

import (
	"context"

	"bitbucket.org/superiortechnologies/ichor/business/domain/homebus"
	"bitbucket.org/superiortechnologies/ichor/business/domain/location/countrybus"
	"bitbucket.org/superiortechnologies/ichor/business/domain/location/regionbus"
	"bitbucket.org/superiortechnologies/ichor/business/domain/productbus"
	"bitbucket.org/superiortechnologies/ichor/business/domain/userbus"
)

// User represents an app user specified for the test.
type User struct {
	userbus.User
	Products []productbus.Product
	Homes    []homebus.Home
}

// SeedData represents data that was seeded for the test.
type SeedData struct {
	Users     []User
	Admins    []User
	Countries []countrybus.Country
	Regions   []regionbus.Region
}

// Table represent fields needed for running an unit test.
type Table struct {
	Name    string
	ExpResp any
	ExcFunc func(ctx context.Context) any
	CmpFunc func(got any, exp any) string
}
