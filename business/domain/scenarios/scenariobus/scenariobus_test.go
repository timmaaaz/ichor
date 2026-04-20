package scenariobus_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/timmaaaz/ichor/business/domain/scenarios/scenariobus"
	"github.com/timmaaaz/ichor/business/sdk/delegate"
	"github.com/timmaaaz/ichor/foundation/logger"
	"github.com/timmaaaz/ichor/foundation/otel"
)

func Test_NewBusiness(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf, logger.LevelInfo, "TEST", func(context.Context) string { return otel.GetTraceID(context.Background()) })
	bus := scenariobus.NewBusiness(log, delegate.New(log), nil, nil)
	if bus == nil {
		t.Fatalf("NewBusiness returned nil")
	}
}
