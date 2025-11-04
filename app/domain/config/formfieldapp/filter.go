package formfieldapp

import (
	"strconv"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/config/formfieldbus"
)

func parseFilter(qp QueryParams) (formfieldbus.QueryFilter, error) {
	var filter formfieldbus.QueryFilter

	if qp.ID != "" {
		id, err := uuid.Parse(qp.ID)
		if err != nil {
			return formfieldbus.QueryFilter{}, errs.NewFieldsError("id", err)
		}
		filter.ID = &id
	}

	if qp.FormID != "" {
		formID, err := uuid.Parse(qp.FormID)
		if err != nil {
			return formfieldbus.QueryFilter{}, errs.NewFieldsError("form_id", err)
		}
		filter.FormID = &formID
	}

	if qp.EntitySchema != "" {
		filter.EntitySchema = &qp.EntitySchema
	}

	if qp.EntityTable != "" {
		filter.EntityTable = &qp.EntityTable
	}

	if qp.Name != "" {
		filter.Name = &qp.Name
	}

	if qp.FieldType != "" {
		filter.FieldType = &qp.FieldType
	}

	if qp.Required != "" {
		required, err := strconv.ParseBool(qp.Required)
		if err != nil {
			return formfieldbus.QueryFilter{}, errs.NewFieldsError("required", err)
		}
		filter.Required = &required
	}

	return filter, nil
}