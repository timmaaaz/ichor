# 01 — Discover What's Available

**Goal**: Ask the chatbot what it knows about workflow building blocks. These are read-only — nothing gets created.

---

## Setup

No context needed. Just send:

```json
{
  "message": "your prompt here",
  "context_type": "workflow"
}
```

---

## Prompts to Try

### 1A — What action types exist?

```
What workflow action types are available? Give me a quick overview of what each one does.
```

**What to check:**
- The chatbot calls `discover` behind the scenes (you may see this in the tool call events)
- The response lists action types in plain language — things like create_alert, send_email, evaluate_condition, delay, etc.
- The chatbot explains what each one does without dumping a wall of technical data

---

### 1B — What triggers are available?

```
What triggers can I use for a workflow? When does each one fire?
```

**What to check:**
- Response mentions: on_create, on_update, on_delete, and scheduled
- Chatbot explains when each fires in plain terms (e.g., "on_create fires whenever a new record is added")

---

### 1C — What entities can I watch?

```
What tables or entities can a workflow be triggered on?
```

**What to check:**
- Response lists entities across different areas of the system (inventory, sales, core, etc.)
- Entities are shown in a readable format (not raw UUIDs)
- If there are many, the chatbot should group or summarize rather than listing hundreds

---

### 1D — How does a specific action work?

```
What fields does a send_email action need? What's required?
```

**What to check:**
- Chatbot explains the required config for send_email (recipients, subject, body/template)
- Does not make up fields — sticks to what's actually defined
- Explains it conversationally, not as a JSON schema dump

---

### 1E — How does branching work?

```
I want to have a workflow do different things based on a condition. How does that work?
```

**What to check:**
- Chatbot explains the `evaluate_condition` action type
- Mentions that it has two output paths: one if the condition is true, one if false
- Explains in plain language without heavy jargon

---

## Common Issues

| Problem | What to look for |
|---------|-----------------|
| Chatbot just says "I don't know" | Should call `discover` — check if tool calls are showing |
| Response is a wall of raw JSON | Chatbot should summarize in plain language |
| Chatbot invents action types that don't exist | Check against the real list: create_alert, send_email, evaluate_condition, delay, seek_approval, etc. |
