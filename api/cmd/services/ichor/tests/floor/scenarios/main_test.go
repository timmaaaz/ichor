// Package scenarios_test walks 21 scenario YAMLs through floor workflow
// endpoints with ScenariosEnabled: true on the mux, providing regression
// coverage for the GB-006/007/012/014/015 bug class.
//
// This package is intentionally named `scenarios` (no `api` suffix) because
// it walks scenarios across multiple API domains (transferorderapi, picktaskapi,
// lottrackingsapi, inventoryitemapi) rather than targeting one handler package.
// See docs/arch/floor-scenarios-harness.md.
package scenarios_test
