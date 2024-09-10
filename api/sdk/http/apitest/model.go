package apitest

import (
	"bitbucket.org/superiortechnologies/ichor/business/domain/homebus"
	"bitbucket.org/superiortechnologies/ichor/business/domain/location/countrybus"
	"bitbucket.org/superiortechnologies/ichor/business/domain/productbus"
	"bitbucket.org/superiortechnologies/ichor/business/domain/userbus"
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
