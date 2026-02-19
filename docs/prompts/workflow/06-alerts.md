# Step 06: Working with Alerts

**Goal**: Test the alert-related tools — listing your own alerts, getting alert details, checking if a rule has fired, and understanding the difference between alert inbox and configured recipients.

---

## Context Setup

Alert queries don't require a workflow context (you're querying fired events, not workflow structure):

```json
{
  "message": "<prompt>",
  "context_type": "workflow"
}
```

To check alerts specifically linked to a known workflow:

```json
{
  "message": "<prompt>",
  "context_type": "workflow",
  "context": {
    "workflow_id": "<uuid>",
    "rule_name": "Simple Test Workflow"
  }
}
```

---

## Prompt 6A — Check Your Alert Inbox

Use this to test listing the current user's alerts:

```
Do I have any active alerts? Show me what's in my inbox.
```

**Expected agent behavior:**
1. Calls `list_my_alerts` with default parameters (status defaults to `active`)
2. Returns list of alerts addressed to the current user (direct or via role)
3. Presents a summary: count, severities, titles

**What to verify:**
- Agent uses `list_my_alerts` (not `list_alerts_for_rule` or `get_alert_detail`)
- Response shows only alerts for the current user's inbox
- Recipient names/emails are shown (not raw UUIDs)
- If inbox is empty, agent says so clearly

---

## Prompt 6B — Filter Alerts by Severity

Use this to test alert filtering:

```
Show me only my high or critical alerts.
```

**Expected agent behavior:**
1. Calls `list_my_alerts` with `severity: "high"` (and possibly separate call for critical)
2. Returns filtered list

**What to verify:**
- Agent passes the severity filter in the tool call
- Response only includes high-severity alerts
- Agent handles the case where `severity` only supports one value per call (may need two calls for high + critical)

---

## Prompt 6C — Check if a Workflow Has Fired

Use this to test whether a specific rule has triggered any alerts:

```
Has the "Simple Test Workflow" actually fired any alerts? How many times has it triggered?
```

**Expected agent behavior:**
1. Calls `list_alerts_for_rule` with `workflow_id: "Simple Test Workflow"` (name resolves automatically)
2. Returns all alerts created by that rule (not just current user's)
3. Shows count, recent alerts, statuses

**What to verify:**
- Agent uses `list_alerts_for_rule` (not `list_my_alerts`)
- Response shows ALL alerts from that rule (not just user's inbox)
- Shows alert count, statuses, timestamps
- If the rule hasn't fired yet, agent says "no alerts have been fired by this rule"

---

## Prompt 6D — Get Alert Details

Use this to test drilling into a specific alert:

```
Show me the details of my most recent alert.
```

**Expected agent behavior:**
1. Calls `list_my_alerts` to find the most recent one
2. Calls `get_alert_detail` with the alert's UUID
3. Returns enriched details including recipient names, source rule, severity

**What to verify:**
- Agent makes two calls: `list_my_alerts` then `get_alert_detail`
- `get_alert_detail` is called with the UUID from the list result (not a made-up UUID)
- Recipients in the detail are shown as names/emails (enriched, not raw UUIDs)
- Agent presents the details clearly in plain language

---

## Prompt 6E — Pagination Test

Use this to test large alert sets:

```
Show me ALL my alerts, not just the first page. I have a lot of them.
```

**Expected agent behavior:**
1. Calls `list_my_alerts` with `rows: 50` (first page)
2. Checks `has_more` in response
3. If `has_more: true`, calls again with `page: 2`
4. Continues until all alerts are fetched or summarizes with total count

**What to verify:**
- Agent uses the `page` parameter to paginate
- Response mentions total count
- Agent doesn't just dump all alerts — summarizes them

---

## Prompt 6F — Who Receives Alerts (Configured vs Fired)

This tests the critical distinction between configured recipients and alert inbox:

**Part 1 — Configured recipients (who the workflow is SET UP to alert):**

```
Who is configured to receive alerts from the "Branching Test Workflow"?
```

**Expected agent behavior:**
1. Calls `explain_workflow_node` with `node_name: "High Value Alert"` (or asks which action)
2. Returns the configured recipients: users and roles in `action_config.recipients`
3. Resolves UUIDs to names

**What to verify:**
- Agent uses `explain_workflow_node` (NOT `list_my_alerts` or `list_alerts_for_rule`)
- Recipients are resolved to names/emails
- Agent explains these are the CONFIGURED recipients (who will receive alerts when it fires)

---

**Part 2 — Alert history (who HAS RECEIVED alerts):**

```
Who has actually received alerts from the "Simple Test Workflow" in the past?
```

**Expected agent behavior:**
1. Calls `list_alerts_for_rule` for the rule
2. Returns fired alert instances with recipient data
3. Explains these are historical fired alerts

**What to verify:**
- Agent uses `list_alerts_for_rule` (NOT `explain_workflow_node`)
- Response shows actual fired alert instances
- Agent distinguishes between "configured to receive" vs "has received"

---

## Alert Status Values

| Status | Meaning |
|--------|---------|
| `active` | Alert is in the recipient's inbox, unacknowledged |
| `acknowledged` | Recipient has acknowledged the alert |
| `dismissed` | Recipient dismissed without resolving |
| `resolved` | Alert has been resolved |

---

## Alert Severity Values

| Severity | Use Case |
|----------|---------|
| `low` | Informational, no immediate action needed |
| `medium` | Warrants attention but not urgent |
| `high` | Requires prompt action |
| `critical` | Immediate action required |

---

## Common Errors to Watch For

| Error | Cause | Fix |
|-------|-------|-----|
| Agent uses `list_my_alerts` for "who does the workflow alert?" | Confusing inbox with config | Config recipients → `explain_workflow_node`; Fired alerts inbox → `list_my_alerts` |
| Agent uses `list_alerts_for_rule` for current user's inbox | Wrong tool for inbox | User inbox → `list_my_alerts`; All rule alerts → `list_alerts_for_rule` |
| Alert recipients shown as raw UUIDs | Missing enrichment | `list_my_alerts` and `get_alert_detail` both return enriched recipient data |
| Agent only fetches page 1 when user asked for all | Missing pagination | Check `has_more` and paginate |
| Agent calls `get_alert_detail` with made-up UUID | Hallucination | `get_alert_detail` requires a real UUID from a previous list call |

---

## Notes

- `list_my_alerts` shows ONLY alerts where the current user is a recipient (directly or via role).
- `list_alerts_for_rule` shows ALL alerts fired by a rule — useful for debugging whether a rule is working.
- `get_alert_detail` returns enriched recipient data (names, emails) for a single alert.
- `explain_workflow_node` resolves configured recipients for an action — not fired alert history.
- The `workflow_id` in context is auto-injected into `list_alerts_for_rule` — no need to specify it explicitly when it's in context.
