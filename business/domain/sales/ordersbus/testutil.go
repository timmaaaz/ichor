package ordersbus

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

func TestNewOrders(n int, userIDs uuid.UUIDs, customerIDs uuid.UUIDs, ofIDs uuid.UUIDs) []NewOrder {
	// Use actual Orders
	orders := make([]NewOrder, 0, n)
	for i := 0; i < n; i++ {
		orders = append(orders, NewOrder{
			Number:              fmt.Sprintf("TST-%d", i+1),
			CustomerID:          customerIDs[i%len(customerIDs)],
			DueDate:             time.Now().AddDate(0, 0, i+1),
			FulfillmentStatusID: ofIDs[i%len(ofIDs)],
			CreatedBy:           userIDs[i%len(userIDs)],
		})
	}
	return orders
}

func TestSeedOrders(ctx context.Context, n int, userIDs uuid.UUIDs, customerIDs uuid.UUIDs, ofIDs uuid.UUIDs, api *Business) ([]Order, error) {
	newOrders := TestNewOrders(n, userIDs, customerIDs, ofIDs)
	orders := make([]Order, len(newOrders))
	for i, ns := range newOrders {
		s, err := api.Create(ctx, ns)
		if err != nil {
			return []Order{}, err
		}
		orders[i] = s
	}
	return orders, nil
}

// TestNewOrdersHistorical creates orders distributed across a time range for seeding.
// daysBack specifies how many days of history to generate (e.g., 30, 90, 365).
// Orders are evenly distributed across the time range.
func TestNewOrdersHistorical(n int, daysBack int, userIDs uuid.UUIDs, customerIDs uuid.UUIDs, ofIDs uuid.UUIDs) []NewOrder {
	orders := make([]NewOrder, 0, n)
	now := time.Now()

	for i := 0; i < n; i++ {
		// Distribute evenly across the time range
		daysAgo := (i * daysBack) / n
		createdDate := now.AddDate(0, 0, -daysAgo)

		orders = append(orders, NewOrder{
			Number:              fmt.Sprintf("SEED-%d", i+1),
			CustomerID:          customerIDs[i%len(customerIDs)],
			DueDate:             createdDate.AddDate(0, 0, 7), // Due 7 days after creation
			FulfillmentStatusID: ofIDs[i%len(ofIDs)],
			CreatedBy:           userIDs[i%len(userIDs)],
			CreatedDate:         &createdDate, // Explicit historical date
		})
	}
	return orders
}

// TestSeedOrdersHistorical seeds orders with historical date distribution.
func TestSeedOrdersHistorical(ctx context.Context, n int, daysBack int, userIDs uuid.UUIDs, customerIDs uuid.UUIDs, ofIDs uuid.UUIDs, api *Business) ([]Order, error) {
	newOrders := TestNewOrdersHistorical(n, daysBack, userIDs, customerIDs, ofIDs)
	orders := make([]Order, len(newOrders))
	for i, no := range newOrders {
		order, err := api.Create(ctx, no)
		if err != nil {
			return []Order{}, err
		}
		orders[i] = order
	}
	return orders, nil
}

// getWeightedHour returns an hour (0-23) with weighted probabilities favoring business hours.
// Business hours (9-17): 60% probability
// Early/late hours (6-8, 18-20): 25% probability
// Night hours (21-5): 15% probability
func getWeightedHour() int {
	// Generate a random number between 0-99 for percentage-based selection
	r := rand.Intn(100)

	switch {
	case r < 60: // 60% - Business hours (9 AM - 5 PM)
		return 9 + rand.Intn(9) // Returns 9-17
	case r < 85: // 25% - Early morning/evening (6-8 AM, 6-8 PM)
		if rand.Intn(2) == 0 {
			return 6 + rand.Intn(3) // Returns 6-8 (morning)
		}
		return 18 + rand.Intn(3) // Returns 18-20 (evening)
	default: // 15% - Night hours (9 PM - 5 AM)
		night := 21 + rand.Intn(9) // Returns 21-29
		if night >= 24 {
			return night - 24 // Wrap to 0-5 (early morning)
		}
		return night
	}
}

// isWeekday returns true if the given time is a weekday (Monday-Friday).
func isWeekday(t time.Time) bool {
	weekday := t.Weekday()
	return weekday >= time.Monday && weekday <= time.Friday
}

// TestNewOrdersFrontendWeighted creates orders with weighted random distribution
// across business hours and days for realistic heatmap visualization.
// This function is specifically designed for frontend seed data and should NOT
// be used in unit/integration tests where predictable data is required.
func TestNewOrdersFrontendWeighted(n int, daysBack int, userIDs uuid.UUIDs, customerIDs uuid.UUIDs, ofIDs uuid.UUIDs) []NewOrder {
	orders := make([]NewOrder, 0, n)
	now := time.Now()

	// Track how many orders we've placed on weekdays vs weekends
	weekdayTarget := int(float64(n) * 0.75)  // 75% on weekdays
	weekdayCount := 0

	for i := 0; i < n; i++ {
		// Generate a random date within the time range
		var createdDate time.Time

		// If we haven't hit our weekday target and we're not at the end, prefer weekdays
		// Otherwise, just pick a random day
		maxAttempts := 10
		for attempt := 0; attempt < maxAttempts; attempt++ {
			daysAgo := rand.Intn(daysBack + 1)
			candidateDate := now.AddDate(0, 0, -daysAgo)

			needWeekday := weekdayCount < weekdayTarget && (n-i) > (weekdayTarget-weekdayCount)

			if needWeekday && isWeekday(candidateDate) {
				createdDate = candidateDate
				weekdayCount++
				break
			} else if !needWeekday && !isWeekday(candidateDate) {
				createdDate = candidateDate
				break
			} else if attempt == maxAttempts-1 {
				// Just use whatever we got on the last attempt
				createdDate = candidateDate
				if isWeekday(candidateDate) {
					weekdayCount++
				}
				break
			}
		}

		// Get weighted random hour
		hour := getWeightedHour()
		minute := rand.Intn(60)
		second := rand.Intn(60)

		// Combine date and time
		finalDate := time.Date(
			createdDate.Year(),
			createdDate.Month(),
			createdDate.Day(),
			hour,
			minute,
			second,
			0,
			createdDate.Location(),
		)

		orders = append(orders, NewOrder{
			Number:              fmt.Sprintf("DEMO-%d", i+1),
			CustomerID:          customerIDs[i%len(customerIDs)],
			DueDate:             finalDate.AddDate(0, 0, 7), // Due 7 days after creation
			FulfillmentStatusID: ofIDs[i%len(ofIDs)],
			CreatedBy:           userIDs[i%len(userIDs)],
			CreatedDate:         &finalDate,
		})
	}

	return orders
}

// TestSeedOrdersFrontendWeighted seeds orders with weighted random distribution
// for realistic frontend visualization. Uses business-hour and weekday weighting
// to create patterns that show well in heatmap charts.
func TestSeedOrdersFrontendWeighted(ctx context.Context, n int, daysBack int, userIDs uuid.UUIDs, customerIDs uuid.UUIDs, ofIDs uuid.UUIDs, api *Business) ([]Order, error) {
	newOrders := TestNewOrdersFrontendWeighted(n, daysBack, userIDs, customerIDs, ofIDs)
	orders := make([]Order, len(newOrders))
	for i, no := range newOrders {
		order, err := api.Create(ctx, no)
		if err != nil {
			return []Order{}, err
		}
		orders[i] = order
	}
	return orders, nil
}
