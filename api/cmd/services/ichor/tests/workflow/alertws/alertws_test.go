package alertws_test

import (
	"testing"

	"github.com/timmaaaz/ichor/api/sdk/http/apitest"
)

// Test_AlertWS runs WebSocket authentication tests through the full HTTP stack.
// These tests verify JWT-based authentication for WebSocket upgrade requests.
func Test_AlertWS(t *testing.T) {
	t.Parallel()

	test := apitest.StartWSTest(t, "Test_AlertWS")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// ==========================================================================
	// WebSocket Authentication Tests
	// ==========================================================================

	t.Run("ws-upgrade-authenticated", func(t *testing.T) {
		testWSUpgradeAuthenticated(t, test, sd)
	})
	t.Run("ws-upgrade-unauthenticated", func(t *testing.T) {
		testWSUpgradeUnauthenticated(t, test, sd)
	})
	t.Run("ws-upgrade-invalid-token", func(t *testing.T) {
		testWSUpgradeInvalidToken(t, test, sd)
	})
	t.Run("ws-upgrade-with-header-token", func(t *testing.T) {
		testWSUpgradeWithHeaderToken(t, test, sd)
	})
}

// Test_AlertWS_E2E runs end-to-end tests through the full pipeline.
// These tests require RabbitMQ and verify:
// - RabbitMQ message publishing
// - Consumer processing
// - AlertHub routing (user/role targeting)
// - WebSocket message delivery
//
// Note: Unit tests for AlertHub internals (connection counting, direct broadcast)
// remain in api/domain/http/workflow/alertws/alerthub_test.go
func Test_AlertWS_E2E(t *testing.T) {
	t.Parallel()

	test := apitest.StartWSTestWithRabbitMQ(t, "Test_AlertWS_E2E")

	sd, err := insertSeedData(test.DB, test.Auth)
	if err != nil {
		t.Fatalf("Seeding error: %s", err)
	}

	// ==========================================================================
	// End-to-End Tests (RabbitMQ → Consumer → AlertHub → WebSocket)
	// ==========================================================================

	t.Run("e2e-user-targeted-delivery", func(t *testing.T) {
		testE2EUserTargetedDelivery(t, test, sd)
	})
	t.Run("e2e-role-based-delivery", func(t *testing.T) {
		testE2ERoleBasedDelivery(t, test, sd)
	})
	t.Run("e2e-broadcast-delivery", func(t *testing.T) {
		testE2EBroadcastDelivery(t, test, sd)
	})
	t.Run("e2e-user-isolation", func(t *testing.T) {
		testE2EUserIsolation(t, test, sd)
	})
	t.Run("e2e-test-alert-endpoint", func(t *testing.T) {
		testE2ETestAlertEndpoint(t, test, sd)
	})
}
