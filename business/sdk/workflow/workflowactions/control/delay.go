package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/timmaaaz/ichor/business/sdk/workflow"
	"github.com/timmaaaz/ichor/foundation/logger"
)

// MaxDelayDuration is the maximum allowed delay (30 days).
const MaxDelayDuration = 720 * time.Hour

// DelayConfig represents configuration for delay actions.
type DelayConfig struct {
	Duration string `json:"duration"`
}

// DelayHandler handles delay actions. In production, the delay is intercepted
// at the Temporal workflow level using workflow.Sleep() for durable timers.
// This handler only provides validation and a fallback Execute.
type DelayHandler struct {
	log *logger.Logger
}

// NewDelayHandler creates a new delay handler.
func NewDelayHandler(log *logger.Logger) *DelayHandler {
	return &DelayHandler{log: log}
}

// GetType returns the action type.
func (h *DelayHandler) GetType() string {
	return "delay"
}

// SupportsManualExecution returns false - delays only make sense in automated workflows.
func (h *DelayHandler) SupportsManualExecution() bool {
	return false
}

// IsAsync returns false - delay is handled at the workflow level, not the activity level.
func (h *DelayHandler) IsAsync() bool {
	return false
}

// GetDescription returns a human-readable description.
func (h *DelayHandler) GetDescription() string {
	return "Pause workflow execution for a specified duration using durable timers"
}

// ParseDuration parses and validates a duration string from a delay config.
func ParseDuration(durationStr string) (time.Duration, error) {
	if durationStr == "" {
		return 0, errors.New("duration is required")
	}

	d, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0, fmt.Errorf("invalid duration format: %w", err)
	}

	if d <= 0 {
		return 0, errors.New("duration must be positive")
	}

	if d > MaxDelayDuration {
		return 0, fmt.Errorf("duration %s exceeds maximum of %s", d, MaxDelayDuration)
	}

	return d, nil
}

// Validate validates the delay configuration.
func (h *DelayHandler) Validate(config json.RawMessage) error {
	var cfg DelayConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return fmt.Errorf("invalid configuration format: %w", err)
	}

	_, err := ParseDuration(cfg.Duration)
	return err
}

// Execute is a fallback that should not be called in production.
// In production, the delay is intercepted at the Temporal workflow level
// and uses workflow.Sleep() for durable timers.
func (h *DelayHandler) Execute(ctx context.Context, config json.RawMessage, execContext workflow.ActionExecutionContext) (any, error) {
	var cfg DelayConfig
	if err := json.Unmarshal(config, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	d, err := ParseDuration(cfg.Duration)
	if err != nil {
		return nil, err
	}

	h.log.Info(ctx, "delay action executed (fallback - should be intercepted at workflow level)",
		"duration", d.String())

	return map[string]any{
		"delayed":  true,
		"duration": cfg.Duration,
	}, nil
}
