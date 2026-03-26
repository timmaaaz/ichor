package cyclecountsessionbus

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// TestNewCycleCountSessions generates n new cycle count sessions for testing.
func TestNewCycleCountSessions(n int, createdByIDs []uuid.UUID) []NewCycleCountSession {
	sessions := make([]NewCycleCountSession, n)

	for i := range n {
		sessions[i] = NewCycleCountSession{
			Name:      fmt.Sprintf("Cycle Count Session %d", i+1),
			CreatedBy: createdByIDs[i%len(createdByIDs)],
		}
	}

	return sessions
}

// TestSeedCycleCountSessions creates n cycle count sessions in the database for testing.
func TestSeedCycleCountSessions(ctx context.Context, n int, createdByIDs []uuid.UUID, api *Business) ([]CycleCountSession, error) {
	newSessions := TestNewCycleCountSessions(n, createdByIDs)

	sessions := make([]CycleCountSession, len(newSessions))
	for i, ncs := range newSessions {
		session, err := api.Create(ctx, ncs)
		if err != nil {
			return nil, err
		}
		sessions[i] = session
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].ID.String() < sessions[j].ID.String()
	})

	return sessions, nil
}
