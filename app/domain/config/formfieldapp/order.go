package formfieldapp

import (
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy(formfieldbus.OrderByFieldOrder, order.ASC)

var orderByFields = map[string]string{
	formfieldbus.OrderByID:         formfieldbus.OrderByID,
	formfieldbus.OrderByFormID:     formfieldbus.OrderByFormID,
	formfieldbus.OrderByName:       formfieldbus.OrderByName,
	formfieldbus.OrderByFieldOrder: formfieldbus.OrderByFieldOrder,
	formfieldbus.OrderByFieldType:  formfieldbus.OrderByFieldType,
}