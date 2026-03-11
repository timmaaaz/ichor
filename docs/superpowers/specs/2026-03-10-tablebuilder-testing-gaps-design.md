# Tablebuilder Testing Gaps — Design Spec

**Date:** 2026-03-10
**Branch:** feature/tablebuilder-testing-gaps
**Scope:** Full coverage pass across `business/sdk/tablebuilder`

---

## Problem

The tablebuilder package has significant testing gaps:

1. `builder.go` — SQL generation has zero unit tests; only exercised implicitly by integration tests that don't assert correctness.
2. `chart.go` (`ChartTransformer`) — entirely untested.
3. `tablebuilder_test.go` integration tests — all example functions use `log.Printf` on errors and `fmt.Printf` on results; they never call `t.Fatal`/`t.Error`, so they cannot catch regressions.
4. `multi_groupby_test.go` — only tests struct construction, not actual SQL output.

---

## Approach

Option 2: Full test restructure.

- Add `builder_test.go` — pure unit tests for SQL generation (no DB).
- Add `chart_test.go` — pure unit tests for ChartTransformer (no DB).
- Rewrite `tablebuilder_test.go` — convert example functions into proper `t.Run` subtests with `t.Fatal` and `cmp.Diff` assertions derived from known seed data.
- Upgrade `multi_groupby_test.go` — add SQL string assertions alongside existing struct tests.

---

## Section 1: `builder_test.go`

**File:** `business/sdk/tablebuilder/builder_test.go`
**Package:** `tablebuilder_test`
**No DB required.**

### Test Groups

**`TestBuildQuery_Filters`**
One subtest per operator: `eq`, `neq`, `gt`, `gte`, `lt`, `lte`, `in`, `like`, `ilike`, `is_null`, `is_not_null`.
Assert the returned SQL `WHERE` clause contains the expected expression.
Also test: dynamic filter from `QueryParams.Dynamic` overrides static filter value.

**`TestBuildQuery_Sorting`**
- Config sort used when `params.Sort` is empty.
- `params.Sort` overrides config sort.
- `asc` and `desc` directions produce correct `ORDER BY`.
- Multiple sort columns all appear in `ORDER BY`.

**`TestBuildQuery_Pagination`**
- `Page=1, Rows=10` → `LIMIT 10 OFFSET 0`.
- `Page=3, Rows=10` → `LIMIT 10 OFFSET 20`.
- Non-primary source skips pagination.
- `ds.Rows > 0` with no page params applies row limit only.

**`TestBuildQuery_Joins`**
- `inner`, `left`, `right`, `full` join types produce correct SQL JOIN keyword.
- Join condition string `"table1.id = table2.foreign_id"` parsed correctly.
- Schema prefix applied when `join.Schema != ""`.

**`TestBuildQuery_ForeignTableJoins`**
- Schema + alias: `schema.table AS alias`.
- No schema + alias: `table AS alias`.
- Schema, no alias: `schema.table`.
- All 4 relationship parse patterns (table.col/table.col, col/table.col, table.col/col, col/col).
- Nested foreign tables recurse correctly.
- Join type (`left`, `inner`, etc.) respected.

**`TestBuildCountQuery`**
- Produces `SELECT COUNT(*) AS "count"`.
- No `LIMIT`/`OFFSET` in output.
- Filters still applied to `WHERE` clause.

**`TestBuildMetricQuery`**
- Each aggregate function: `sum`, `count`, `count_distinct`, `avg`, `min`, `max`.
- `count_distinct` → `COUNT(DISTINCT col)`.
- Single GroupBy with `Interval: "month"` → `DATE_TRUNC('month', ...)` in SELECT and GROUP BY.
- Multi GroupBy → all columns in GROUP BY clause.
- Raw expression GroupBy (`Expression: true`) → literal SQL in both SELECT and GROUP BY.
- Invalid metric function → error returned.
- Metric with `Expression` (arithmetic) → operator appears in SQL.

**`TestBuildArithmeticExpression`**
- `multiply` → `*`, `add` → `+`, `subtract` → `-`, `divide` → `/`.
- Invalid operator → error.
- Invalid column reference → error.
- Two-column and three-column expressions.

---

## Section 2: `chart_test.go`

**File:** `business/sdk/tablebuilder/chart_test.go`
**Package:** `tablebuilder_test`
**No DB required.** Constructs `TableData` directly and calls `NewChartTransformer().Transform(...)`.

### Test Groups

**`TestChartTransformer_KPI`**
- Single metric row → `KPIData.Value` matches input.
- `Change` and `Trend` (`"up"`/`"down"`) computed correctly when `PreviousValue` present.
- Nil data → error returned.

**`TestChartTransformer_Bar` / `TestChartTransformer_Line`**
- Grouped rows → `SeriesData` labels and values match input rows in order.
- Multi-series config → multiple series entries produced.

**`TestChartTransformer_StackedBar` / `TestChartTransformer_StackedArea`**
- Stacked variants produce correct series structure with stacking metadata.

**`TestChartTransformer_Pie`**
- Label/value pairs extracted correctly.
- Percentages sum to 100 (within float tolerance).

**`TestChartTransformer_Heatmap`**
- x/y/value triple extracted from rows correctly.

**`TestChartTransformer_Gantt`**
- task, start, end fields mapped to `GanttData` correctly.

**`TestChartTransformer_Treemap`**
- name/value/parent hierarchy preserved in `TreemapData`.

**`TestChartTransformer_Waterfall`**
- Sequential step values extracted correctly.

**`TestChartTransformer_ErrorCases`**
- Nil data → error.
- Missing required fields for chart type → error.
- Unknown chart type → error.

---

## Section 3: `tablebuilder_test.go` rewrite

**File:** `business/sdk/tablebuilder/tablebuilder_test.go` — rewritten.
**Requires DB.** Uses `dbtest.NewDatabase` + seed data.

Example functions become proper `t.Run` subtests inside `Test_TableBuilder`, each receiving `t *testing.T`. All `log.Printf` → `t.Fatal`. Assertions use `cmp.Diff`.

### Test Cases

**`simple_inventory_items`**
- Row count matches seeded items with `quantity > 0`.
- Every row contains `inventory_items.id` and `current_stock` fields.
- Rows sorted descending by quantity (first row has highest value).

**`orders_view`**
- Row count matches seeded orders.
- Required column names present in every row.

**`inventory_with_joins`**
- `product_name` field is non-empty in every row (join worked).
- `stock_status` computed column is `"low"` or `"normal"` for every row (evaluator ran).

**`stored_config_roundtrip`**
- Create config → load by ID → fetch data.
- Loaded config title matches saved title (`cmp.Diff`).
- Fetched rows > 0.

**`pagination_correctness`**
- Page 1: exactly 10 rows returned.
- Page 2: correct remaining count.
- No ID appears on both pages (no duplicates across pages).

**`inventory_adjustments_deep_join`**
- 3-level nested foreign table join (adjustment → location → warehouse).
- `warehouse_name` field present in result rows.
- `location_code` computed column present in result rows.

**`configstore_crud`**
- Create → QueryByID → QueryByName → Update → Delete lifecycle.
- `cmp.Diff` at each read step.

**`page_config_crud`**
- Create → QueryByID → QueryByName → Update → Delete lifecycle.
- `cmp.Diff` at each read step.
- (This test is currently in `pageConfigsExample` which is never called.)

**`dynamic_filters`**
- Pass `QueryParams.Filters` at call time.
- Filtered row count < unfiltered count.

---

## Section 4: `multi_groupby_test.go` upgrade

**File:** `business/sdk/tablebuilder/multi_groupby_test.go` — additive (existing struct tests preserved).
**No DB required.**

### New Test Groups

**`TestMultiGroupBy_SQLGeneration/single_groupby_date_trunc`**
- `Interval: "month"` → SQL contains `DATE_TRUNC('month', orders.created_date)` and `GROUP BY`.

**`TestMultiGroupBy_SQLGeneration/multiple_groupby_categorical`**
- Two simple column GroupBys → SQL `GROUP BY` clause contains both columns.

**`TestMultiGroupBy_SQLGeneration/expression_groupby`**
- `Expression: true` with `EXTRACT(DOW FROM ...)` → raw expression in both SELECT and GROUP BY.

**`TestMultiGroupBy_SQLGeneration/mixed_interval_and_expression`**
- One date interval + one raw expression → both appear correctly in generated SQL.

**`TestMultiGroupBy_SQLGeneration/no_groupby_produces_no_group_clause`**
- Empty `GroupBy` slice → generated SQL contains no `GROUP BY`.

**`TestMultiGroupBy_ErrorCases/invalid_interval_returns_error`**
- `Interval: "invalid"` → `BuildQuery` returns error.

**`TestMultiGroupBy_ErrorCases/expression_missing_alias_returns_error`**
- `Expression: true`, `Alias: ""` → `BuildQuery` returns error.

---

## Files Changed

| File | Action |
|------|--------|
| `business/sdk/tablebuilder/builder_test.go` | Create |
| `business/sdk/tablebuilder/chart_test.go` | Create |
| `business/sdk/tablebuilder/tablebuilder_test.go` | Rewrite |
| `business/sdk/tablebuilder/multi_groupby_test.go` | Add new test groups |

No production code changes required.
