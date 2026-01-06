package tablebuilder_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/timmaaaz/ichor/business/sdk/tablebuilder"
)

func Test_EvaluatorFunctions(t *testing.T) {
	t.Parallel()

	eval := tablebuilder.NewEvaluator()

	t.Run("ceil function", func(t *testing.T) {
		row := tablebuilder.TableRow{"value": 3.2}
		result, err := eval.Evaluate("ceil(value)", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != 4.0 {
			t.Errorf("ceil(3.2) = %v, want 4", result)
		}
	})

	t.Run("floor function", func(t *testing.T) {
		row := tablebuilder.TableRow{"value": 3.8}
		result, err := eval.Evaluate("floor(value)", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != 3.0 {
			t.Errorf("floor(3.8) = %v, want 3", result)
		}
	})

	t.Run("round function", func(t *testing.T) {
		row := tablebuilder.TableRow{"value": 3.5}
		result, err := eval.Evaluate("round(value)", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != 4.0 {
			t.Errorf("round(3.5) = %v, want 4", result)
		}
	})

	t.Run("now function", func(t *testing.T) {
		row := tablebuilder.TableRow{}
		result, err := eval.Evaluate("now()", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		nowUnix := float64(time.Now().Unix())
		resultFloat := result.(float64)
		// Allow 2 second tolerance
		if resultFloat < nowUnix-2 || resultFloat > nowUnix+2 {
			t.Errorf("now() = %v, want ~%v", resultFloat, nowUnix)
		}
	})

	t.Run("daysUntil with future date", func(t *testing.T) {
		futureDate := time.Now().AddDate(0, 0, 5).Format("2006-01-02")
		row := tablebuilder.TableRow{"due_date": futureDate}
		result, err := eval.Evaluate("daysUntil(due_date)", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		resultFloat := result.(float64)
		if resultFloat < 4 || resultFloat > 6 {
			t.Errorf("daysUntil(5 days from now) = %v, want ~5", resultFloat)
		}
	})

	t.Run("daysUntil with past date returns negative", func(t *testing.T) {
		pastDate := time.Now().AddDate(0, 0, -3).Format("2006-01-02")
		row := tablebuilder.TableRow{"due_date": pastDate}
		result, err := eval.Evaluate("daysUntil(due_date)", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		resultFloat := result.(float64)
		if resultFloat > -2 || resultFloat < -4 {
			t.Errorf("daysUntil(3 days ago) = %v, want ~-3", resultFloat)
		}
	})

	t.Run("daysSince with past date", func(t *testing.T) {
		pastDate := time.Now().AddDate(0, 0, -5).Format("2006-01-02")
		row := tablebuilder.TableRow{"order_date": pastDate}
		result, err := eval.Evaluate("daysSince(order_date)", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		resultFloat := result.(float64)
		if resultFloat < 4 || resultFloat > 6 {
			t.Errorf("daysSince(5 days ago) = %v, want ~5", resultFloat)
		}
	})

	t.Run("isOverdue with past date returns true", func(t *testing.T) {
		pastDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		row := tablebuilder.TableRow{"due_date": pastDate}
		result, err := eval.Evaluate("isOverdue(due_date)", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != true {
			t.Errorf("isOverdue(yesterday) = %v, want true", result)
		}
	})

	t.Run("isOverdue with future date returns false", func(t *testing.T) {
		futureDate := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
		row := tablebuilder.TableRow{"due_date": futureDate}
		result, err := eval.Evaluate("isOverdue(due_date)", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != false {
			t.Errorf("isOverdue(tomorrow) = %v, want false", result)
		}
	})

	t.Run("hasValue with non-nil value returns true", func(t *testing.T) {
		row := tablebuilder.TableRow{"delivery_date": "2025-01-15"}
		result, err := eval.Evaluate("hasValue(delivery_date)", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != true {
			t.Errorf("hasValue('2025-01-15') = %v, want true", result)
		}
	})

	t.Run("hasValue with nil returns false", func(t *testing.T) {
		row := tablebuilder.TableRow{"delivery_date": nil}
		result, err := eval.Evaluate("hasValue(delivery_date)", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != false {
			t.Errorf("hasValue(nil) = %v, want false", result)
		}
	})

	t.Run("hasValue with empty string returns false", func(t *testing.T) {
		row := tablebuilder.TableRow{"delivery_date": ""}
		result, err := eval.Evaluate("hasValue(delivery_date)", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != false {
			t.Errorf("hasValue('') = %v, want false", result)
		}
	})
}

func Test_EvaluatorTransformations(t *testing.T) {
	t.Parallel()

	eval := tablebuilder.NewEvaluator()

	t.Run("transforms !== to !=", func(t *testing.T) {
		row := tablebuilder.TableRow{"status": "Pending"}
		result, err := eval.Evaluate("status !== 'Delivered'", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != true {
			t.Errorf("status !== 'Delivered' with status='Pending' = %v, want true", result)
		}
	})

	t.Run("transforms === to ==", func(t *testing.T) {
		row := tablebuilder.TableRow{"status": "Delivered"}
		result, err := eval.Evaluate("status === 'Delivered'", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != true {
			t.Errorf("status === 'Delivered' with status='Delivered' = %v, want true", result)
		}
	})

	t.Run("transforms Math.ceil to ceil", func(t *testing.T) {
		row := tablebuilder.TableRow{"value": 3.2}
		result, err := eval.Evaluate("Math.ceil(value)", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != 4.0 {
			t.Errorf("Math.ceil(3.2) = %v, want 4", result)
		}
	})

	t.Run("transforms Math.floor to floor", func(t *testing.T) {
		row := tablebuilder.TableRow{"value": 3.8}
		result, err := eval.Evaluate("Math.floor(value)", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != 3.0 {
			t.Errorf("Math.floor(3.8) = %v, want 3", result)
		}
	})

	t.Run("transforms Math.round to round", func(t *testing.T) {
		row := tablebuilder.TableRow{"value": 3.5}
		result, err := eval.Evaluate("Math.round(value)", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != 4.0 {
			t.Errorf("Math.round(3.5) = %v, want 4", result)
		}
	})
}

func Test_EvaluatorComplexExpressions(t *testing.T) {
	t.Parallel()

	eval := tablebuilder.NewEvaluator()

	t.Run("isOverdue combined with status check", func(t *testing.T) {
		pastDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

		// Overdue and not delivered -> true
		row := tablebuilder.TableRow{
			"due_date":    pastDate,
			"status_name": "Pending",
		}
		result, err := eval.Evaluate("isOverdue(due_date) && status_name != 'Delivered'", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != true {
			t.Errorf("isOverdue && not delivered = %v, want true", result)
		}

		// Overdue but delivered -> false
		row["status_name"] = "Delivered"
		result, err = eval.Evaluate("isOverdue(due_date) && status_name != 'Delivered'", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != false {
			t.Errorf("isOverdue && delivered = %v, want false", result)
		}
	})

	t.Run("ternary with hasValue for delivery status", func(t *testing.T) {
		pastDate := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

		// Has delivery date -> delivered
		row := tablebuilder.TableRow{
			"actual_delivery_date":   "2025-01-10",
			"expected_delivery_date": pastDate,
		}
		result, err := eval.Evaluate("hasValue(actual_delivery_date) ? 'delivered' : (isOverdue(expected_delivery_date) ? 'overdue' : 'pending')", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != "delivered" {
			t.Errorf("with delivery date = %v, want 'delivered'", result)
		}

		// No delivery date, past expected -> overdue
		row["actual_delivery_date"] = nil
		result, err = eval.Evaluate("hasValue(actual_delivery_date) ? 'delivered' : (isOverdue(expected_delivery_date) ? 'overdue' : 'pending')", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != "overdue" {
			t.Errorf("no delivery, past expected = %v, want 'overdue'", result)
		}

		// No delivery date, future expected -> pending
		futureDate := time.Now().AddDate(0, 0, 5).Format("2006-01-02")
		row["expected_delivery_date"] = futureDate
		result, err = eval.Evaluate("hasValue(actual_delivery_date) ? 'delivered' : (isOverdue(expected_delivery_date) ? 'overdue' : 'pending')", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != "pending" {
			t.Errorf("no delivery, future expected = %v, want 'pending'", result)
		}
	})

	t.Run("round rating expression", func(t *testing.T) {
		row := tablebuilder.TableRow{"rating": 3.7}
		result, err := eval.Evaluate("round(rating * 2) / 2", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		// 3.7 * 2 = 7.4, round = 7, / 2 = 3.5
		if result != 3.5 {
			t.Errorf("round(3.7 * 2) / 2 = %v, want 3.5", result)
		}
	})

	t.Run("daysSince with ternary for tenure", func(t *testing.T) {
		pastDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")

		// Has date_hired -> show days
		row := tablebuilder.TableRow{"date_hired": pastDate}
		result, err := eval.Evaluate("hasValue(date_hired) ? daysSince(date_hired) : nil", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		resultFloat := result.(float64)
		if resultFloat < 29 || resultFloat > 31 {
			t.Errorf("daysSince(30 days ago) = %v, want ~30", resultFloat)
		}

		// No date_hired -> nil
		row["date_hired"] = nil
		result, err = eval.Evaluate("hasValue(date_hired) ? daysSince(date_hired) : nil", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != nil {
			t.Errorf("daysSince with nil date_hired = %v, want nil", result)
		}
	})
}

func Test_EvaluatorDateFormats(t *testing.T) {
	t.Parallel()

	eval := tablebuilder.NewEvaluator()

	testCases := []struct {
		name   string
		format string
	}{
		{"RFC3339", time.Now().AddDate(0, 0, 5).Format(time.RFC3339)},
		{"RFC3339 with Z", time.Now().AddDate(0, 0, 5).Format("2006-01-02T15:04:05Z")},
		{"datetime with space", time.Now().AddDate(0, 0, 5).Format("2006-01-02 15:04:05")},
		{"date only", time.Now().AddDate(0, 0, 5).Format("2006-01-02")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			row := tablebuilder.TableRow{"date": tc.format}
			result, err := eval.Evaluate("daysUntil(date)", row)
			if err != nil {
				t.Fatalf("evaluate failed for format %s: %v", tc.name, err)
			}
			resultFloat := result.(float64)
			if resultFloat < 4 || resultFloat > 6 {
				t.Errorf("daysUntil with format %s = %v, want ~5", tc.name, resultFloat)
			}
		})
	}
}

func Test_EvaluatorTypeErrors(t *testing.T) {
	t.Parallel()

	eval := tablebuilder.NewEvaluator()

	t.Run("ceil with non-numeric returns ErrInvalidNumericArg", func(t *testing.T) {
		row := tablebuilder.TableRow{"value": "not-a-number"}
		_, err := eval.Evaluate("ceil(value)", row)
		if err == nil {
			t.Fatal("expected error for non-numeric ceil argument")
		}
		if !errors.Is(err, tablebuilder.ErrInvalidNumericArg) {
			t.Errorf("expected ErrInvalidNumericArg, got: %v", err)
		}
	})

	t.Run("ceil with nil returns ErrInvalidNumericArg", func(t *testing.T) {
		row := tablebuilder.TableRow{"value": nil}
		_, err := eval.Evaluate("ceil(value)", row)
		if err == nil {
			t.Fatal("expected error for nil ceil argument")
		}
		if !errors.Is(err, tablebuilder.ErrInvalidNumericArg) {
			t.Errorf("expected ErrInvalidNumericArg, got: %v", err)
		}
	})

	t.Run("daysUntil with nil returns ErrNilArgument", func(t *testing.T) {
		row := tablebuilder.TableRow{"date": nil}
		_, err := eval.Evaluate("daysUntil(date)", row)
		if err == nil {
			t.Fatal("expected error for nil date")
		}
		if !errors.Is(err, tablebuilder.ErrNilArgument) {
			t.Errorf("expected ErrNilArgument, got: %v", err)
		}
	})

	t.Run("daysUntil with invalid date returns ErrInvalidDateFormat", func(t *testing.T) {
		row := tablebuilder.TableRow{"date": "not-a-date"}
		_, err := eval.Evaluate("daysUntil(date)", row)
		if err == nil {
			t.Fatal("expected error for invalid date format")
		}
		if !errors.Is(err, tablebuilder.ErrInvalidDateFormat) {
			t.Errorf("expected ErrInvalidDateFormat, got: %v", err)
		}
	})

	t.Run("isOverdue with invalid date returns ErrInvalidDateFormat", func(t *testing.T) {
		row := tablebuilder.TableRow{"date": "January 5th, 2025"}
		_, err := eval.Evaluate("isOverdue(date)", row)
		if err == nil {
			t.Fatal("expected error for invalid date format")
		}
		if !errors.Is(err, tablebuilder.ErrInvalidDateFormat) {
			t.Errorf("expected ErrInvalidDateFormat, got: %v", err)
		}
	})

	t.Run("daysSince with empty string returns ErrInvalidDateFormat", func(t *testing.T) {
		row := tablebuilder.TableRow{"date": ""}
		_, err := eval.Evaluate("daysSince(date)", row)
		if err == nil {
			t.Fatal("expected error for empty date string")
		}
		if !errors.Is(err, tablebuilder.ErrInvalidDateFormat) {
			t.Errorf("expected ErrInvalidDateFormat, got: %v", err)
		}
	})

	t.Run("error message contains helpful context", func(t *testing.T) {
		row := tablebuilder.TableRow{"date": "bad-date"}
		_, err := eval.Evaluate("daysUntil(date)", row)
		if err == nil {
			t.Fatal("expected error")
		}
		errMsg := err.Error()
		if !strings.Contains(errMsg, "RFC3339") && !strings.Contains(errMsg, "YYYY-MM-DD") {
			t.Errorf("error message should mention expected formats: %v", errMsg)
		}
	})
}

func Test_EvaluatorEdgeCases(t *testing.T) {
	t.Parallel()

	eval := tablebuilder.NewEvaluator()

	t.Run("arithmetic with functions", func(t *testing.T) {
		row := tablebuilder.TableRow{"value": 100.0}
		result, err := eval.Evaluate("value * 2 + ceil(value / 3)", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		// 100 * 2 + ceil(33.33) = 200 + 34 = 234
		if result != 234.0 {
			t.Errorf("100 * 2 + ceil(100/3) = %v, want 234", result)
		}
	})

	t.Run("nested ternary", func(t *testing.T) {
		row := tablebuilder.TableRow{"status": "pending", "priority": "high"}
		result, err := eval.Evaluate("status == 'complete' ? 'done' : (priority == 'high' ? 'urgent' : 'normal')", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != "urgent" {
			t.Errorf("nested ternary = %v, want 'urgent'", result)
		}
	})

	t.Run("comparison operators", func(t *testing.T) {
		row := tablebuilder.TableRow{"value": 5.0}
		result, err := eval.Evaluate("value > 3 && value < 10", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != true {
			t.Errorf("5 > 3 && 5 < 10 = %v, want true", result)
		}
	})

	t.Run("string concatenation in expression", func(t *testing.T) {
		row := tablebuilder.TableRow{"first": "John", "last": "Doe"}
		result, err := eval.Evaluate("first + ' ' + last", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != "John Doe" {
			t.Errorf("first + ' ' + last = %v, want 'John Doe'", result)
		}
	})

	t.Run("hasValue for null coalescing pattern", func(t *testing.T) {
		// govaluate doesn't support ?? operator, use hasValue() pattern instead
		row := tablebuilder.TableRow{"value": nil}
		result, err := eval.Evaluate("hasValue(value) ? value : 'default'", row)
		if err != nil {
			t.Fatalf("evaluate failed: %v", err)
		}
		if result != "default" {
			t.Errorf("hasValue pattern with nil = %v, want 'default'", result)
		}
	})
}
