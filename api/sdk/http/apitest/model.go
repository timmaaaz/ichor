package apitest

import (
	"github.com/timmaaaz/ichor/app/domain/location/cityapp"
	"github.com/timmaaaz/ichor/app/domain/location/streetapp"
	"github.com/timmaaaz/ichor/business/domain/homebus"
	"github.com/timmaaaz/ichor/business/domain/location/countrybus"
	"github.com/timmaaaz/ichor/business/domain/location/regionbus"
	"github.com/timmaaaz/ichor/business/domain/productbus"
	"github.com/timmaaaz/ichor/business/domain/userbus"
)

// User extends the dbtest user for api test support.
type User struct {
	userbus.User
	Products []productbus.Product
	Homes    []homebus.Home
	Token    string
}

// SeedData represents users for api tests.
type SeedData struct {
	Users     []User
	Admins    []User
	Countries []countrybus.Country
	Regions   []regionbus.Region
	Cities    []cityapp.City
	Streets   []streetapp.Street
}

// Table represent fields needed for running an api test.
type Table struct {
	Name       string
	URL        string
	Token      string
	Method     string
	StatusCode int
	Input      any
	GotResp    any
	ExpResp    any
	CmpFunc    func(got any, exp any) string
}
