# 06 — Working with Alerts

**Goal**: Test the chatbot's ability to show your alert inbox, look up alert details, and check whether a workflow rule has actually fired.

---

## Setup

Alert queries don't need a workflow context:

```json
{
  "message": "your prompt here",
  "context_type": "workflow"
}
```

---

## Prompts to Try

### 6A — Check your inbox

```
Do I have any active alerts? Show me what's in my inbox.
```

**What to check:**
- Response shows alerts addressed to you (either directly or via a role you belong to)
- Each alert shows: title, severity, status
- If your inbox is empty, chatbot says so clearly — not "I couldn't find any alerts" (which implies an error)

---

### 6B — Filter by severity

```
Show me only my high or critical alerts.
```

**What to check:**
- Response only shows high-severity alerts
- Lower-severity alerts (low, medium) are not included

---

### 6C — Check if a workflow has actually fired

```
Has the "Simple Test Workflow" sent any alerts? How many times has it triggered?
```

**What to check:**
- Chatbot shows alerts created by that rule (not just ones in your inbox — all of them)
- Shows count and recent alert titles
- If the rule hasn't fired yet, chatbot says "No alerts have been created by this rule yet"

---

### 6D — Alert details

```
Show me the details of my most recent alert.
```

**What to check:**
- Chatbot first lists your alerts, then fetches detail on the most recent one
- Detail includes: title, message, severity, who sent it, what workflow triggered it
- Recipients shown as names/emails (not UUIDs)

---

### 6E — Understand the difference: configured vs. fired

This is a key distinction — make sure the chatbot gets it right.

**Part 1 — who is SET UP to receive alerts:**
```
Who is configured to receive alerts from the "High Value Alert" action in the Branching Test Workflow?
```

**What to check:**
- Chatbot explains the **configured** recipients (the users/roles set up in the action)
- It should look up the action config, not the alert history
- Recipients shown as names

**Part 2 — who HAS received alerts:**
```
Who has actually received alerts from the Simple Test Workflow in the past?
```

**What to check:**
- Chatbot shows **fired alert instances** (historical events, not config)
- Different information from Part 1
- Chatbot distinguishes between "configured to receive" vs "has received"

---

## Common Issues

| Problem | What to look for |
|---------|-----------------|
| Chatbot shows alert history when asked "who receives alerts from this action?" | Should show configured recipients instead — different question |
| Chatbot shows only your inbox when asked about a rule's alert history | Rule history includes everyone's alerts, not just yours |
| Alert recipients shown as UUIDs | Should be resolved to names/emails |
| Chatbot says "no alerts" when the workflow has fired | Check if it's looking at the right scope (your inbox vs. all alerts for the rule) |

---

## Alert Statuses

| Status | Meaning |
|--------|---------|
| active | Unacknowledged, in your inbox |
| acknowledged | You've seen it and acknowledged it |
| dismissed | Closed without resolving |
| resolved | Marked as resolved |

---

## Alert Severities

| Severity | Meaning |
|----------|---------|
| low | Informational |
| medium | Warrants attention |
| high | Needs prompt action |
| critical | Needs immediate action |
