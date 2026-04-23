package alertws

import (
	"testing"

	"github.com/timmaaaz/ichor/foundation/websocket"
)

func TestMessageTypeForAlert(t *testing.T) {
	tests := []struct {
		rabbitType string
		want       websocket.MessageType
	}{
		{rabbitType: "alert", want: websocket.MessageTypeAlert},
		{rabbitType: "alert_updated", want: websocket.MessageTypeAlertUpdated},
		{rabbitType: "approval_resolved", want: websocket.MessageTypeApprovalResolved},
		{rabbitType: "", want: websocket.MessageTypeAlert},
		{rabbitType: "unknown_type", want: websocket.MessageTypeAlert},
	}

	for _, tt := range tests {
		t.Run(tt.rabbitType, func(t *testing.T) {
			got := messageTypeForAlert(tt.rabbitType)
			if got != tt.want {
				t.Errorf("messageTypeForAlert(%q) = %q, want %q", tt.rabbitType, got, tt.want)
			}
		})
	}
}
