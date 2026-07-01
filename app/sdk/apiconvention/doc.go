// Package apiconvention holds static, source-scanning guards that enforce
// cross-cutting API conventions documented in docs/layer-patterns.md.
//
// These are not runtime code; they exist so convention drift is caught by the
// test suite rather than in review. See the _test.go files for the enforced
// rules.
package apiconvention
