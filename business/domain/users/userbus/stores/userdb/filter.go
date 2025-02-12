package userdb

import (
	"bytes"
	"strings"

	"github.com/timmaaaz/ichor/business/domain/users/userbus"
)

func applyFilter(filter userbus.QueryFilter, data map[string]any, buf *bytes.Buffer) {
	var wc []string

	if filter.ID != nil {
		data["user_id"] = *filter.ID
		wc = append(wc, "user_id = :user_id")
	}

	if filter.RequestedBy != nil {
		data["requested_by"] = *filter.RequestedBy
		wc = append(wc, "requested_by = :requested_by")
	}

	if filter.ApprovedBy != nil {
		data["approved_by"] = *filter.ApprovedBy
		wc = append(wc, "approved_by = :approved_by")
	}

	if filter.TitleID != nil {
		data["title_id"] = *filter.TitleID
		wc = append(wc, "title_id = :title_id")
	}

	if filter.OfficeID != nil {
		data["office_id"] = *filter.OfficeID
		wc = append(wc, "office_id = :office_id")
	}

	if filter.Username != nil {
		data["username"] = "%" + filter.Username.String() + "%"
		wc = append(wc, "username ILIKE :username")
	}

	if filter.FirstName != nil {
		data["first_name"] = "%" + filter.FirstName.String() + "%"
		wc = append(wc, "first_name ILIKE :first_name")
	}

	if filter.LastName != nil {
		data["last_name"] = "%" + filter.LastName.String() + "%"
		wc = append(wc, "last_name ILIKE :last_name")
	}

	if filter.Email != nil {
		data["email"] = (*filter.Email).String()
		wc = append(wc, "email = :email")
	}

	if filter.StartBirthday != nil {
		data["start_birthday"] = filter.StartBirthday.UTC()
		wc = append(wc, "birthday >= :start_birthday")
	}

	if filter.EndBirthday != nil {
		data["end_birthday"] = filter.EndBirthday.UTC()
		wc = append(wc, "birthday <= :end_birthday")
	}

	if filter.StartDateHired != nil {
		data["start_date_hired"] = filter.StartDateHired.UTC()
		wc = append(wc, "date_hired >= :start_date_hired")
	}

	if filter.EndDateHired != nil {
		data["end_date_hired"] = filter.EndDateHired.UTC()
		wc = append(wc, "date_hired <= :end_date_hired")
	}

	if filter.StartDateRequested != nil {
		data["start_date_requested"] = filter.StartDateRequested.UTC()
		wc = append(wc, "date_requested >= :start_date_requested")
	}

	if filter.EndDateRequested != nil {
		data["end_date_requested"] = filter.EndDateRequested.UTC()
		wc = append(wc, "date_requested <= :end_date_requested")
	}

	if filter.StartDateApproved != nil {
		data["start_date_approved"] = filter.StartDateApproved.UTC()
		wc = append(wc, "date_approved >= :start_date_approved")
	}

	if filter.StartCreatedDate != nil {
		data["start_date_created"] = filter.StartCreatedDate.UTC()
		wc = append(wc, "date_created >= :start_date_created")
	}

	if filter.EndCreatedDate != nil {
		data["end_date_created"] = filter.EndCreatedDate.UTC()
		wc = append(wc, "date_created <= :end_date_created")
	}

	if len(wc) > 0 {
		buf.WriteString(" WHERE ")
		buf.WriteString(strings.Join(wc, " AND "))
	}
}
