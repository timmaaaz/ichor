package paymenttermdb

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/business/domain/core/paymenttermbus"
)

type paymentTerm struct {
	ID          uuid.UUID      `db:"id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
}

func toDBPaymentTerm(bus paymenttermbus.PaymentTerm) paymentTerm {
	pt := paymentTerm{
		ID:   bus.ID,
		Name: bus.Name,
	}
	if bus.Description != "" {
		pt.Description = sql.NullString{
			String: bus.Description,
			Valid:  true,
		}
	}
	return pt
}

func toBusPaymentTerm(dbPaymentTerm paymentTerm) paymenttermbus.PaymentTerm {
	return paymenttermbus.PaymentTerm{
		ID:          dbPaymentTerm.ID,
		Name:        dbPaymentTerm.Name,
		Description: dbPaymentTerm.Description.String,
	}
}

func toBusPaymentTerms(dbPaymentTerms []paymentTerm) []paymenttermbus.PaymentTerm {
	paymentTerms := make([]paymenttermbus.PaymentTerm, len(dbPaymentTerms))
	for i, pt := range dbPaymentTerms {
		paymentTerms[i] = toBusPaymentTerm(pt)
	}
	return paymentTerms
}
