# Create Plan Command

Generate a new build plan with complete structure, documentation, and commands.

## Overview

This command creates a standardized build plan that can be used across any project. It:
1. Asks you questions about your plan (project, description, tech stack, success criteria)
2. Has you provide a detailed description of the functionality
3. Analyzes the description to determine logical phases
4. Generates a complete plan structure in `.claude/plans/PLAN_NAME/`
5. Creates 10 project-specific commands in `.claude/commands/`
6. Uses global templates from `~/.claude/plan-system/templates/`

## Usage

```bash
/create-plan <plan-name>
```

Where `<plan-name>` is a short slug (e.g., `user-authentication`, `api-refactor`, `database-migration`)

## Your Task

### 1. Parse Plan Name

Get the plan name from the command parameter.

**Validation**:
- Must be lowercase with hyphens (kebab-case)
- No spaces or special characters
- Between 3-50 characters
- Example: `user-authentication`, `chart-integration`, `database-migration`

If invalid, show error and ask for correct format.

**Convert to Plan Directory Name**:
- Plan slug: `user-authentication`
- Plan directory: `USER_AUTHENTICATION_PLAN`
- Conversion: uppercase + underscores + `_PLAN` suffix

### 2. Check if Plan Already Exists

Check if `.claude/plans/PLAN_NAME/` already exists.

If it does:
```
❌ Plan already exists: .claude/plans/USER_AUTHENTICATION_PLAN/

Would you like to:
1. Choose a different name
2. Delete existing and recreate
3. Cancel

Enter choice (1/2/3):
```

Handle the user's choice appropriately.

### 3. Ask Initial Questions

Ask the user these questions to gather basic plan information:

#### Question 1: Project Name
```
What is the name of your project?
(e.g., "Ichor ERP System", "E-commerce Platform", "Analytics Dashboard")

Project name:
```

#### Question 2: Tech Stack - Backend
```
What backend technology are you using?
(e.g., "Go with Gin", "Python with Django", "Node.js with Express", "Ruby on Rails", "None")

Backend:
```

#### Question 3: Tech Stack - Frontend
```
What frontend technology are you using?
(e.g., "Vue 3 with TypeScript", "React with TypeScript", "Angular", "Svelte", "None")

Frontend:
```

#### Question 4: Success Criteria
```
What are 3-5 success criteria for this plan?
(Enter one per line, press Enter twice when done)

Success criteria:
```

Collect until user enters blank line.

#### Question 5: Dependencies
```
Does this plan depend on any other plans?
(Enter plan names separated by commas, or press Enter for none)

Dependencies:
```

Parse comma-separated list or accept empty for no dependencies.

### 4. Request Detailed Description

Ask the user to provide a comprehensive description of the functionality:

```
Please provide a detailed description of what this plan should accomplish.

Include:
- The problem you're solving
- Desired functionality and outcomes
- Any technical considerations you've thought about
- Integration points with existing systems
- Any constraints or requirements

You can paste a multi-paragraph description. This will be used to determine
the logical phases and structure of the plan.

Description:
```

Wait for the user to provide their detailed description.

### 5. Analyze and Determine Phases

Based on the user's description, analyze and determine the logical phases.

**Phase Breakdown Guidelines**:

1. **Large/Complex Components** should be broken into 3 phases:
   - **Research Phase**: Analyze existing patterns, evaluate options, design architecture
   - **Implementation Phase**: Build based on research findings
   - **Testing Phase**: Comprehensive tests defining patterns for future similar work

2. **Medium Components** may need 2 phases:
   - **Implementation Phase**: Build the functionality
   - **Testing Phase**: If testing patterns are important to establish

3. **Small/Straightforward Components** can be a single phase:
   - When the implementation is well-understood
   - When existing patterns clearly apply
   - When testing is straightforward

**When to Use Research-Implementation-Testing Pattern**:
- New technology or patterns being introduced to the codebase
- Complex integrations with multiple decision points
- Functionality that will set precedent for future similar features
- Areas where the "right" approach isn't immediately clear
- Components that require evaluating multiple options (e.g., library choices)

**When Unsure**: If you're uncertain whether a component needs the full 3-phase treatment or can be simpler, ASK THE USER:

```
I've identified [Component X] as a significant piece of functionality.

Would you like this broken into:
1. Research → Implementation → Testing (3 phases)
   - Best for: new patterns, complex decisions, setting precedents

2. Implementation → Testing (2 phases)
   - Best for: moderately complex, patterns exist but need validation

3. Single Implementation phase
   - Best for: straightforward, well-understood patterns

Which approach for [Component X]?
```

### 6. Present Phase Structure for Approval

After analyzing, present the proposed phase structure:

```
📋 Proposed Phase Structure

Based on your description, I've identified the following logical phases:

══════════════════════════════════════════════════════════════

BACKEND FOUNDATION (Phases 1-3)
  Phase 1: [Name] Research
    → [Why research is needed]
  Phase 2: [Name] Implementation
    → [What will be built]
  Phase 3: [Name] Testing
    → [What testing patterns will be established]

FRONTEND FOUNDATION (Phases 4-5)
  Phase 4: [Name] Implementation
    → [What will be built]
  Phase 5: [Name] Testing
    → [What will be tested]

UI LAYER (Phases 6-7)
  Phase 6: [Component Name]
    → [What will be created]
  Phase 7: [Page Name]
    → [What will be created]

══════════════════════════════════════════════════════════════

Total Phases: 7

Does this structure look correct?
- Type 'yes' to proceed
- Type 'no' to discuss adjustments
- Suggest specific changes (e.g., "combine phases 6 and 7")
```

Wait for user approval or adjustments.

### 7. Confirm Final Plan Details

Show summary and ask for confirmation:

```
📋 Plan Summary

══════════════════════════════════════════════════════════════

Plan Name: user-authentication
Directory: .claude/plans/USER_AUTHENTICATION_PLAN/
Project: Ichor ERP System

Description:
[First 2-3 sentences of their description]

Tech Stack:
  Backend: Go with Gin
  Frontend: Vue 3 with TypeScript

Phases: [N] (determined from analysis)

Success Criteria:
  ✓ [Criterion 1]
  ✓ [Criterion 2]
  ✓ [Criterion 3]

Dependencies: None

══════════════════════════════════════════════════════════════

This will create:
  - Plan directory: .claude/plans/USER_AUTHENTICATION_PLAN/
  - README.md (master plan document)
  - PROGRESS.yaml (progress tracking)
  - phases/ directory (for phase documentation)
  - 10 commands in .claude/commands/

Proceed with plan creation? (yes/no):
```

If no, abort. If yes, continue.

### 8. Create Directory Structure

Create the following directories:

```bash
mkdir -p .claude/plans/USER_AUTHENTICATION_PLAN/phases
mkdir -p .claude/commands
```

### 9. Generate README.md

**Steps**:
1. Read template from `~/.claude/plan-system/templates/plans/README.md.template`
2. Replace template variables:
   - `{{PLAN_NAME}}` → "User Authentication System"
   - `{{PROJECT_NAME}}` → User's project name
   - `{{PLAN_DESCRIPTION}}` → User's description
   - `{{PHASE_COUNT}}` → Number of phases (determined from analysis)
   - `{{PLAN_SLUG}}` → Plan slug (user-authentication)
   - `{{BACKEND_LANGUAGE}}` / `{{BACKEND_FRAMEWORK}}` → Parse from backend answer
   - `{{FRONTEND_FRAMEWORK}}` / `{{FRONTEND_LANGUAGE}}` → Parse from frontend answer
   - `{{FUNCTIONAL_REQ_1,2,3}}` → From success criteria
   - `{{LAST_UPDATED}}` → Current date (YYYY-MM-DD)
   - `{{CURRENT_STATUS}}` → "Planning"
   - For phase table, create rows based on the analyzed phases
   - For unspecified variables, use helpful placeholders like "TODO: Add details"

3. Write to `.claude/plans/USER_AUTHENTICATION_PLAN/README.md`

### 10. Generate PROGRESS.yaml

**Steps**:
1. Read template from `~/.claude/plan-system/templates/plans/PROGRESS.yaml.template`
2. Replace template variables:
   - `{{PLAN_NAME}}` → Display name
   - `{{PLAN_DESCRIPTION}}` → User's description
   - `{{START_DATE}}` → Current date (YYYY-MM-DD)
   - `{{PHASE_COUNT}}` → Number of phases
   - `{{PLAN_SLUG}}` → Plan slug
   - Create `phases` array based on analyzed structure
   - Add `phase_N_doc_created: false` for each phase in planning_status
   - Parse dependencies and add to `dependencies.external`

3. Write to `.claude/plans/USER_AUTHENTICATION_PLAN/PROGRESS.yaml`

### 11. Generate Commands

For each of the 10 commands, generate from templates:

**Commands to generate**:
1. `user-authentication-status.md`
2. `user-authentication-next.md`
3. `user-authentication-phase.md`
4. `user-authentication-validate.md`
5. `user-authentication-review.md`        # Code review (AFTER implementation)
6. `user-authentication-plan-review.md`   # Plan review (BEFORE implementation)
7. `user-authentication-build-phase.md`
8. `user-authentication-summary.md`
9. `user-authentication-dependencies.md`
10. `user-authentication-quick-status.md`  # Compact phase overview table

**Steps for each command**:
1. Read template from `~/.claude/plan-system/templates/commands/{command}.md.template`
2. Replace template variables:
   - `{{PLAN_NAME}}` → Display name ("User Authentication System")
   - `{{PLAN_SLUG}}` → Plan slug ("user-authentication")
   - `{{PHASE_COUNT}}` → Number of phases
3. Write to `.claude/commands/{plan-slug}-{command}.md`

### 12. Create Phase Entries in PROGRESS.yaml

Based on the analyzed phase structure, create detailed phase entries:

```yaml
phases:
  - phase: 1
    name: "Backend Architecture Research"
    description: "Analyze patterns and design WebSocket architecture"
    status: "pending"
    category: "backend"
    tasks:
      - task: "Define tasks using /{plan-slug}-build-phase"
        status: "pending"
        notes: []
        files: []
    validation:
      - check: "Architecture document created"
        status: "pending"
    deliverables: []
    blockers: []
```

Repeat for each phase with appropriate names and categories.

### 13. Show Success Summary

Display what was created:

```
✅ Plan Created Successfully!

══════════════════════════════════════════════════════════════

PLAN: USER_AUTHENTICATION_PLAN
Location: .claude/plans/USER_AUTHENTICATION_PLAN/

FILES CREATED:
  ✓ README.md (master plan document)
  ✓ PROGRESS.yaml (progress tracking initialized)
  ✓ phases/ (directory for phase documentation)

COMMANDS CREATED (10):
  ✓ /user-authentication-status
  ✓ /user-authentication-next
  ✓ /user-authentication-phase
  ✓ /user-authentication-validate
  ✓ /user-authentication-review        (code review - AFTER implementation)
  ✓ /user-authentication-plan-review   (plan review - BEFORE implementation)
  ✓ /user-authentication-build-phase
  ✓ /user-authentication-summary
  ✓ /user-authentication-dependencies
  ✓ /user-authentication-quick-status  (compact phase overview table)

══════════════════════════════════════════════════════════════

PHASE STRUCTURE:

[Group 1 Name]:
  Phase 1: [Name] (research)
  Phase 2: [Name] (implementation)
  Phase 3: [Name] (testing)

[Group 2 Name]:
  Phase 4: [Name] (implementation)
  ...

══════════════════════════════════════════════════════════════

NEXT STEPS:

1. Review and refine the generated README.md:
   .claude/plans/USER_AUTHENTICATION_PLAN/README.md

2. Start documenting phases:
   /user-authentication-build-phase

   This will guide you through creating detailed documentation
   for Phase 1 (and subsequent phases).

3. Review the phase plan BEFORE implementing:
   /user-authentication-plan-review

   This reviews the plan document and grades it.
   Target grade: B+ or higher before implementing.

4. Once plan is approved, begin execution:
   /user-authentication-next

   This will execute the first phase step-by-step.

5. Review code AFTER implementing:
   /user-authentication-review

   This reviews the implementation and grades it.
   Target grade: B+ or higher before next phase.

6. Check progress anytime:
   /user-authentication-status

7. After ALL phases are complete and tested:
   mv .claude/commands/user-authentication-*.md .claude/completed-commands/

   This archives the commands to keep .claude/commands/ clean.

══════════════════════════════════════════════════════════════

TIP: Run /user-authentication-build-phase now to document your
first phase while the plan context is fresh!
```

### 14. Archive Commands After Plan Completion

**IMPORTANT**: After ALL phases of the plan have been completed AND tested:

1. Move the plan's commands to the completed-commands directory:
   ```bash
   mv .claude/commands/{plan-slug}-*.md .claude/completed-commands/
   ```

   Example for `user-authentication`:
   ```bash
   mv .claude/commands/user-authentication-*.md .claude/completed-commands/
   ```

2. This keeps the `.claude/commands/` directory clean and manageable
3. Completed commands are gitignored but preserved locally for reference

**When to archive**:
- ✅ All phases marked as "completed" in PROGRESS.yaml
- ✅ All validation checks passing
- ✅ Final testing completed successfully
- ✅ Plan is considered DONE

**Do NOT archive** if there are any pending phases or untested changes.

## Phase Breakdown Decision Tree

Use this to determine how to structure phases:

```
Is this component...

├─ Introducing NEW technology/patterns to the codebase?
│   └─ YES → Research → Implementation → Testing (3 phases)
│
├─ Requiring evaluation of multiple options (libraries, approaches)?
│   └─ YES → Research → Implementation → Testing (3 phases)
│
├─ Setting precedent for future similar features?
│   └─ YES → Research → Implementation → Testing (3 phases)
│
├─ Complex with multiple integration points?
│   └─ YES → Consider 3 phases, ASK USER if unsure
│
├─ Moderately complex but patterns exist?
│   └─ Implementation → Testing (2 phases)
│
├─ Straightforward with well-understood patterns?
│   └─ Single Implementation phase
│
└─ Unsure?
    └─ ASK THE USER which approach they prefer
```

## Template Variable Reference

### Common Variables

| Variable | Example Value | Source |
|----------|---------------|--------|
| `{{PLAN_NAME}}` | "User Authentication System" | Derived from slug |
| `{{PLAN_SLUG}}` | "user-authentication" | User input |
| `{{PLAN_DESCRIPTION}}` | "Implement JWT auth..." | User description |
| `{{PROJECT_NAME}}` | "Ichor ERP System" | User Q&A |
| `{{PHASE_COUNT}}` | 8 | Determined from analysis |
| `{{START_DATE}}` | "2025-12-02" | Current date |
| `{{BACKEND_FRAMEWORK}}` | "Go with Gin" | Parsed from Q&A |
| `{{FRONTEND_FRAMEWORK}}` | "Vue 3 with TypeScript" | Parsed from Q&A |

### Slug to Display Name Conversion

```
user-authentication → User Authentication System
chart-integration → Chart Integration System
database-migration → Database Migration System
api-refactor → API Refactor System
```

Rules:
1. Split on hyphens
2. Capitalize each word
3. Append "System" or "Plan" (user preference)

## Error Handling

### Global Templates Not Found

If `~/.claude/plan-system/templates/` doesn't exist:

```
❌ Global plan system not found at ~/.claude/plan-system/

This system requires the global plan templates to be installed.

Would you like me to install them now? (yes/no):
```

If yes, create the directory structure and copy templates from the current project (if they exist) or ask user to run the formalization plan first.

### Invalid Plan Name

```
❌ Invalid plan name: "User Authentication"

Plan names must be:
  - Lowercase letters and hyphens only
  - No spaces or special characters
  - Format: kebab-case (e.g., user-authentication)

Please try again with a valid name.
```

### Command Generation Failure

If a command template is missing:

```
⚠️  Warning: Template not found for command: status
Skipping this command. You may need to create it manually.

Continuing with remaining commands...
```

## Success Criteria

Plan creation is successful when:
- ✅ All files created in correct locations
- ✅ README.md has no unfilled `{{VARIABLES}}`
- ✅ PROGRESS.yaml is valid YAML syntax
- ✅ PROGRESS.yaml includes both `plan_reviewed`/`plan_review_grade` AND `reviewed`/`review_grade` fields
- ✅ All 10 commands generated and functional
- ✅ Phase structure logically derived from user's description
- ✅ User approved the phase breakdown
- ✅ Generated files are ready to use immediately

---

**Note**: This command should work from any project directory. It discovers the global template system automatically and creates plans in the current project's `.claude/plans/` directory.
