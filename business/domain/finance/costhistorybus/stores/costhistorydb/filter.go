package costhistorydb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/finance/costhistorybus"
)

func applyFilter(filter costhistorybus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.CostHistoryID != nil {
		data["history_id"] = *filter.CostHistoryID
		wc = append(wc, "history_id = :history_id")
	}

	if filter.ProductID != nil {
		data["product_id"] = *filter.ProductID
		wc = append(wc, "product_id = :product_id")
	}

	if filter.CostType != nil {
		data["cost_type"] = *filter.CostType
		wc = append(wc, "cost_type = :cost_type")
	}

	if filter.Amount != nil {
		data["amount"] = *filter.Amount
		wc = append(wc, "amount = :amount")
	}

	if filter.Currency != nil {
		data["currency"] = *filter.Currency
		wc = append(wc, "currency = :currency")
	}

	if filter.EndDate != nil {
		data["end_date"] = *filter.EndDate
		wc = append(wc, "end_date = :end_date")
	}

	if filter.EffectiveDate != nil {
		data["effective_date"] = *filter.EffectiveDate
		wc = append(wc, "effective_date = :effective_date")
	}

	if filter.CreatedDate != nil {
		data["created_date"] = *filter.CreatedDate
		wc = append(wc, "created_date = :created_date")
	}

	if filter.UpdatedDate != nil {
		data["updated_date"] = *filter.UpdatedDate
		wc = append(wc, "updated_date = :updated_date")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
