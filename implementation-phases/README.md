# Backend API Implementation Phases

**Purpose**: Implement 11 missing endpoints to unblock frontend development phases 2, 4, and 8.

**Based on**: Frontend Gap Analysis (2025-01-19)

---

## Overview

This directory contains detailed, step-by-step implementation guides for three phases of backend API development. Each phase file is a self-contained prompt that can be given to Claude Code to implement that specific set of endpoints.

---

## Phase Summary

| Phase | Endpoints | Estimated Time | Unblocks Frontend Phase | Priority |
|-------|-----------|----------------|------------------------|----------|
| [Phase 1: QueryAll](phase-1-queryall.md) | 3 | 3-4 hours | Phase 2 (Admin Lists) | Medium |
| [Phase 2: Introspection](phase-2-introspection.md) | 4 | 8-10 hours | Phase 4 (Table Builder) | **CRITICAL** |
| [Phase 3: Import/Export](phase-3-import-export.md) | 6 | 6-8 hours | Phase 8 (Templates) | **CRITICAL** |

**Total**: 11 endpoints, 17-22 hours (3-4 days)

---

## Recommended Implementation Order

### Week 1 (Critical Path)
1. **Phase 2: Introspection** (Days 1-2)
   - Creates new `introspectionbus` domain
   - 4 endpoints for database schema metadata
   - **BLOCKS**: Frontend Phase 4 (Table Builder UI)
   - Start here to unblock frontend ASAP

2. **Phase 3: Import/Export - Page Configs** (Day 3)
   - Add export/import to `pageconfigapi`
   - 2 endpoints (export + import)
   - **BLOCKS**: Frontend Phase 8 (Templates)

3. **Testing & Integration** (Day 4)
   - Integration tests for Phase 2 and partial Phase 3
   - Verify frontend can consume new endpoints

### Week 2 (Completion)
4. **Phase 3: Import/Export - Forms & Table Configs** (Days 1-2)
   - Complete remaining 4 endpoints
   - Forms export/import
   - Table configs export/import

5. **Phase 1: QueryAll** (Day 3)
   - Add list endpoints to existing domains
   - 3 endpoints (forms, page-configs, table-configs)
   - **ENHANCES**: Frontend Phase 2 (Admin Lists)

6. **Testing & Documentation** (Day 4)
   - Complete integration test suite
   - Update API documentation
   - Verify all 11 endpoints working

---

## Phase Details

### Phase 1: QueryAll Endpoints (Optional but Recommended)

**File**: [phase-1-queryall.md](phase-1-queryall.md)

**What**: Add "list all" endpoints for config entities
**Why**: Enables admin UI to display all configs without database queries
**Pattern**: Follow `purchaseorderstatusapi` QueryAll pattern

**Endpoints**:
- `GET /v1/config/forms/all`
- `GET /v1/config/page-configs/all`
- `GET /v1/data/configs/all`

**Frontend Impact**: Phase 2 admin list views work cleanly without workarounds

**Can Be Skipped**: Yes (frontend can query database directly as workaround)

---

### Phase 2: Introspection Domain (Critical)

**File**: [phase-2-introspection.md](phase-2-introspection.md)

**What**: New domain for PostgreSQL schema introspection
**Why**: Table Builder UI needs to know database structure
**Pattern**: Create full domain (bus/app/api) following Ardan Labs architecture

**Endpoints**:
- `GET /v1/introspection/schemas` - List all schemas
- `GET /v1/introspection/schemas/{schema}/tables` - List tables in schema
- `GET /v1/introspection/tables/{schema}/{table}/columns` - Column metadata
- `GET /v1/introspection/tables/{schema}/{table}/relationships` - Foreign keys

**Frontend Impact**: Phase 4 (Table Builder) **CANNOT PROCEED** without these

**Can Be Skipped**: No - this is a hard blocker

---

### Phase 3: Import/Export Endpoints (Critical)

**File**: [phase-3-import-export.md](phase-3-import-export.md)

**What**: Bulk export/import for config entities
**Why**: Template library, config sharing, backup/restore
**Pattern**: Extend existing domains with export/import methods

**Endpoints**:
- `POST /v1/config/forms/export` + `POST /v1/config/forms/import`
- `POST /v1/config/page-configs/export` + `POST /v1/config/page-configs/import`
- `POST /v1/data/configs/export` + `POST /v1/data/configs/import`

**Frontend Impact**: Phase 8 (Import/Export UI) **CANNOT PROCEED** without these

**Can Be Skipped**: No - this is a hard blocker for Phase 8

---

## How to Use These Guides

Each phase file is designed to be given to Claude Code as a complete prompt:

### Option 1: Sequential Implementation
```bash
# Copy contents of phase-2-introspection.md and give to Claude
# Then copy phase-3-import-export.md
# Then copy phase-1-queryall.md
```

### Option 2: Parallel Implementation (Multiple Developers)
- Developer 1: Phase 2 (Introspection)
- Developer 2: Phase 3 (Import/Export)
- Developer 3: Phase 1 (QueryAll)

All phases are independent and can be implemented in parallel.

---

## Testing Strategy

Each phase includes:
1. **Unit Tests**: Business logic validation
2. **Integration Tests**: HTTP endpoint testing
3. **Manual Testing**: Verification via API calls

**Test Location**: `api/cmd/services/ichor/tests/{domain}/{api}/`

**Run Tests**:
```bash
# All tests
make test

# Specific domain
go test -v ./api/cmd/services/ichor/tests/introspectionapi
```

---

## Success Criteria

### Phase 1 Complete
- [ ] 3 QueryAll endpoints return all records
- [ ] Frontend can list all forms, page-configs, table-configs
- [ ] Integration tests pass

### Phase 2 Complete
- [ ] 4 introspection endpoints return PostgreSQL metadata
- [ ] Frontend can browse schemas/tables/columns
- [ ] Table Builder UI can suggest joins based on relationships
- [ ] Integration tests pass
- [ ] Admin-only authorization enforced

### Phase 3 Complete
- [ ] 6 export/import endpoints handle JSON packages
- [ ] Round-trip test: export → import → verify identical
- [ ] Conflict resolution works (merge/skip/replace)
- [ ] Related records exported (forms + fields, page-configs + content)
- [ ] Transactions rollback on import failure
- [ ] Integration tests pass

---

## Common Pitfalls

1. **Naming Conflicts**: Avoid naming structs same as packages (e.g., `page` conflicts with `business/sdk/page`)
   - Solution: Use prefixes like `dbPage`, `busPage`

2. **Import Paths**: Always use full module path
   - Correct: `github.com/timmaaaz/ichor/business/domain/introspectionbus`
   - Wrong: `business/domain/introspectionbus`

3. **Encoder Interface**: All response types must implement `Encode()`
   - For slices: Create wrapper type (e.g., `type Entities []Entity`)
   - For paginated: Use `query.Result[T]` (already implements Encode)

4. **Decoder Interface**: All request types must implement `Decode()`
   - Define request models in app layer, not API layer

5. **Foreign Key Resolution**: Import/export must handle ID remapping
   - Export: Include related records by ID
   - Import: Create dependencies first, remap IDs

---

## Additional Resources

- **CLAUDE.md**: Project architecture and patterns
- **FORMDATA_IMPLEMENTATION.md**: Multi-entity transaction pattern
- **Ardan Labs Service**: https://github.com/ardanlabs/service

---

## Questions?

If you encounter issues:
1. Review the specific phase file for detailed steps
2. Check CLAUDE.md for architecture patterns
3. Look at reference implementations:
   - QueryAll: `purchaseorderstatusapi`
   - Domain Creation: `rolebus`, `roleapp`, `roleapi`
   - Import/Export: (no existing example - new pattern)

---

**Ready to start?** Begin with [Phase 2: Introspection](phase-2-introspection.md) to unblock the frontend immediately.
