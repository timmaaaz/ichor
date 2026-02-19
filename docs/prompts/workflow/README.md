# Workflow Chat Prompts

This directory contains test prompts for the workflow automation AI assistant. Use these to verify the chatbot is working correctly.

## How It Works

The workflow assistant lives at `POST /v1/agent/chat` with `context_type: "workflow"`. You type a message, it calls tools behind the scenes, and responds in plain language. For any create or update operation, it always sends a **preview** first — you accept or reject it in the UI without the chatbot saving anything directly.

### Basic Request Format

```json
{
  "message": "your message here",
  "context_type": "workflow"
}
```

To open a specific existing workflow for editing:

```json
{
  "message": "your message here",
  "context_type": "workflow",
  "context": {
    "workflow_id": "<uuid>",
    "rule_name": "Simple Test Workflow"
  }
}
```

To start building a brand new workflow:

```json
{
  "message": "your message here",
  "context_type": "workflow",
  "context": { "is_new": true }
}
```

### What You Should See

When you send a message, the UI shows events in real time:
- **Tool calls** — the chatbot explains what it's looking up (e.g., "I'll check the available action types…")
- **Streamed response** — the chatbot explains what it found or did
- **Preview card** — appears when a workflow is ready for your review (create/update operations only)

---

## Test Files

| File | What it tests |
|------|--------------|
| [01-discover.md](01-discover.md) | Asking what action types, triggers, and entities are available |
| [02-list-and-read.md](02-list-and-read.md) | Listing workflows, reading summaries, asking about specific actions |
| [03-create-simple.md](03-create-simple.md) | Building a new workflow step by step |
| [04-create-branching.md](04-create-branching.md) | Workflows with "if this, do that / if not, do this" logic |
| [05-modify-existing.md](05-modify-existing.md) | Changing an existing workflow |
| [06-alerts.md](06-alerts.md) | Checking your alert inbox and workflow alert history |
| [07-complex-workflow.md](07-complex-workflow.md) | Multi-step real-world automation scenarios |
| [complete-walkthrough.md](complete-walkthrough.md) | Full end-to-end test covering everything |

---

## Seeded Workflows for Testing

After `make seed`, these three workflows exist and can be used in tests:

| Name | Trigger | What it does |
|------|---------|-------------|
| **Simple Test Workflow** | on_create | 1 step: creates an alert |
| **Sequence Test Workflow** | on_update | 3 steps in a row: each creates an alert |
| **Branching Test Workflow** | on_create | Evaluates a condition, then takes one of two paths |

Use these by name in prompts — the chatbot resolves names to IDs automatically.

---

## Quick Reference: Things You Can Ask

| Goal | Example prompt |
|------|---------------|
| Discover what's available | "What action types are there?" |
| List all workflows | "Show me all the workflow rules" |
| Understand a workflow | "Explain the Branching Test Workflow" |
| Drill into one action | "What does the Evaluate Amount action do?" |
| Check alert recipients | "Who gets alerts from the High Value Alert action?" |
| Check your inbox | "Do I have any alerts?" |
| Check if a rule fired | "Has the Simple Test Workflow sent any alerts?" |
| Create new workflow | "Build a new workflow called X on Y triggered when Z" |
| Edit existing workflow | "Change the alert severity to critical" |
