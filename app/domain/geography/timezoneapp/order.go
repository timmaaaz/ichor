package timezoneapp

import (
	"github.com/timmaaaz/ichor/business/domain/geography/timezonebus"
	"github.com/timmaaaz/ichor/business/sdk/order"
)

var defaultOrderBy = order.NewBy("display_name", order.ASC)

var orderByFields = map[string]string{
	"id":          timezonebus.OrderByID,
	"name":        timezonebus.OrderByName,
	"displayName": timezonebus.OrderByDisplayName,
	"utcOffset":   timezonebus.OrderByUTCOffset,
	"isActive":    timezonebus.OrderByIsActive,
}
