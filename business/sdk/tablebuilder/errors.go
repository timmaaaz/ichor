package tablebuilder

import "errors"

// Package-level errors for the tablebuilder package
var (
	// Configuration errors
	ErrInvalidConfig     = errors.New("invalid table configuration")
	ErrNoDataSource      = errors.New("no data source specified")
	ErrInvalidDataSource = errors.New("invalid data source configuration")

	// Column errors
	ErrColumnNotFound        = errors.New("column not found")
	ErrInvalidColumn         = errors.New("invalid column configuration")
	ErrDuplicateColumn       = errors.New("duplicate column name")
	ErrMissingColumnType     = errors.New("column missing type in visual settings")
	ErrMissingDatetimeFormat = errors.New("datetime column missing format configuration")

	// Date format errors
	ErrGoDateFormatDetected = errors.New("Go date format detected, use date-fns format instead")

	// Query errors
	ErrInvalidQuery  = errors.New("invalid query")
	ErrQueryFailed   = errors.New("query execution failed")
	ErrInvalidFilter = errors.New("invalid filter configuration")
	ErrInvalidSort   = errors.New("invalid sort configuration")

	// Join errors
	ErrInvalidJoin = errors.New("invalid join configuration")
	ErrJoinFailed  = errors.New("join operation failed")

	// Expression errors
	ErrInvalidExpression = errors.New("invalid expression")
	ErrEvaluationFailed  = errors.New("expression evaluation failed")

	// Type conversion errors
	ErrInvalidType      = errors.New("invalid type")
	ErrConversionFailed = errors.New("type conversion failed")

	// Database errors
	ErrNotFound      = errors.New("record not found")
	ErrDatabaseError = errors.New("database error")

	// Permission errors
	ErrUnauthorized     = errors.New("unauthorized access")
	ErrInsufficientRole = errors.New("insufficient role permissions")
)
