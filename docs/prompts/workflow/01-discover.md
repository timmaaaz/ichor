# Step 01: Discover Workflow Building Blocks

**Goal**: Use the `discover` tool to learn what action types, trigger types, and entities are available in the system. This is the first step before building any workflow.

---

## Context Setup

For discovery, no workflow context is needed. Omit `context` or send a blank one:

```json
{
  "message": "<prompt>",
  "context_type": "workflow"
}
```

---

## Prompt 1A — Discover All Action Types

Use this to learn what action types exist and what config each requires:

```
What workflow action types are available? I want to understand what each one does.
```

**Expected agent behavior:**
1. Calls `discover` with `category: "action_types"`
2. Returns a list of all 17 action types with their config schemas and output ports
3. Explains each action in plain language

**What to verify:**
- Agent calls `discover`, not some other tool
- Response lists multiple action types (create_alert, evaluate_condition, send_email, etc.)
- Each action type shows its required config fields
- Actions with branching show their output ports (e.g., `evaluate_condition` → `output-true` / `output-false`)

---

## Prompt 1B — Discover Trigger Types

Use this to learn when workflows can fire:

```
What trigger types are available for workflows? When does each one fire?
```

**Expected agent behavior:**
1. Calls `discover` with `category: "trigger_types"`
2. Returns trigger types: on_create, on_update, on_delete, scheduled
3. Explains when each fires

**What to verify:**
- Response includes at least: `on_create`, `on_update`, `on_delete`, `scheduled`
- Agent explains the difference between them in plain language

---

## Prompt 1C — Discover Available Entities

Use this to find what tables/entities workflows can trigger on:

```
What entities can workflows be triggered on? Show me all the available tables.
```

**Expected agent behavior:**
1. Calls `discover` with `category: "entities"`
2. Returns a list of schema.table pairs (e.g., `sales.orders`, `inventory.inventory_items`)
3. Groups them by schema if possible

**What to verify:**
- Response includes entities across multiple schemas (inventory, sales, core, etc.)
- Each entity is listed in `schema.table` format
- Agent can identify which entities are available for a given domain

---

## Prompt 1D — Targeted Action Type Discovery

Use this when you need to know the specific config schema for one action:

```
What config does the send_email action type require? What fields are needed?
```

**Expected agent behavior:**
1. Calls `discover` with `category: "action_types"`
2. Finds and explains the `send_email` action's config schema
3. Lists required fields and their types

**What to verify:**
- Agent explains the required config for send_email (recipients, subject, body, template)
- Does NOT hallucinate fields — shows only what the schema defines

---

## Prompt 1E — Discover Output Ports (for branching)

Use this to understand how conditional branching works:

```
What output ports does evaluate_condition have? How do I use them to branch a workflow?
```

**Expected agent behavior:**
1. Calls `discover` with `category: "action_types"`
2. Finds the `evaluate_condition` action and explains its output ports
3. Explains how `output-true` and `output-false` ports are used in edges

**What to verify:**
- Response mentions `output-true` and `output-false` as the output ports
- Agent explains that the `after: "ActionName:output-true"` shorthand connects to one branch
- Does NOT mention ports that don't exist for this action type

---

## Common Errors to Watch For

| Error | Cause | Fix |
|-------|-------|-----|
| Agent calls `list_workflow_rules` instead of `discover` | Wrong tool for discovery | Agent should use `discover` with category param |
| Agent makes up action type names | No `discover` call | Verify agent always calls `discover` before listing options |
| Missing output ports in response | Incomplete discovery | Ensure agent calls `discover` with `action_types` (not just listing names) |

---

## Notes

- `discover` is a read-only tool — it queries the live API, so results reflect the current system configuration.
- `discover` with `action_types` returns the full config schemas including required/optional fields.
- The `output ports` are listed per action type — only `evaluate_condition` and similar branching actions have multiple ports; linear actions like `create_alert` have a single default port.
- For new workflow creation, always call `discover` first so the LLM knows the correct config schemas before adding actions.
