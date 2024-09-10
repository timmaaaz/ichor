package vproduct_test

import (
	"time"

	"bitbucket.org/superiortechnologies/ichor/app/domain/vproductapp"
	"bitbucket.org/superiortechnologies/ichor/business/domain/productbus"
	"bitbucket.org/superiortechnologies/ichor/business/domain/userbus"
)

func toAppVProduct(usr userbus.User, prd productbus.Product) vproductapp.Product {
	return vproductapp.Product{
		ID:          prd.ID.String(),
		UserID:      prd.UserID.String(),
		Name:        prd.Name.String(),
		Cost:        prd.Cost,
		Quantity:    prd.Quantity,
		DateCreated: prd.DateCreated.Format(time.RFC3339),
		DateUpdated: prd.DateUpdated.Format(time.RFC3339),
		UserName:    usr.Username.String(),
	}
}

func toAppVProducts(usr userbus.User, prds []productbus.Product) []vproductapp.Product {
	items := make([]vproductapp.Product, len(prds))
	for i, prd := range prds {
		items[i] = toAppVProduct(usr, prd)
	}

	return items
}
