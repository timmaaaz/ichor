# Phase 3B: Remaining Import/Export Implementation

**Status**: In Progress
**Priority**: High
**Estimated Time**: 6-8 hours
**Depends On**: Phase 3A (Forms Export/Import) - ✅ COMPLETE

---

## Phase 3A Summary (COMPLETED)

✅ **Forms Export/Import** is fully implemented and compiling:
- Business Layer: `FormWithFields` struct, `ExportByIDs()`, `ImportForms()` with transaction support
- App Layer: `ExportPackage`, `ImportPackage`, `ImportResult` models with Encoder/Decoder
- API Layer: `exportForms()` and `importForms()` handlers
- Routes: `POST /v1/config/forms/export` and `/import`
- Dependencies: Wired in `all.go` and `dbtest.go`

**Reference Implementation**: Use Forms as the template for Table Configs and Page Configs.

---

## Overview

Complete export/import for the remaining two config domains:
1. **Table Configs** (simpler - no related records)
2. **Page Configs** (complex - two types of related records: content & actions)

Both follow the same pattern as Forms but with domain-specific differences.

---

## Part 1: Table Configs Export/Import

### Architecture

**Simplicity**: Table configs have NO related records - they're standalone entities with JSONB config.

**Pattern**: Direct export/import without child records (simpler than Forms).

### Database Schema
```sql
CREATE TABLE data.configs (
    id UUID PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    config JSONB NOT NULL
);
```

---

### Step 1: Business Layer

**Files to modify:**
- `business/domain/data/tableconfigbus/model.go`
- `business/domain/data/tableconfigbus/tableconfigbus.go`

#### 1a. Add Import Models to `model.go`

```go
// Add to model.go after UpdateTableConfig

// ImportStats represents statistics from an import operation.
type ImportStats struct {
	ImportedCount int
	SkippedCount  int
	UpdatedCount  int
}
```

**Note**: No need for `TableConfigWithRelations` since there are no related records!

#### 1b. Add Export Method to `tableconfigbus.go`

```go
// ExportByIDs exports table configs by IDs.
func (b *Business) ExportByIDs(ctx context.Context, configIDs []uuid.UUID) ([]TableConfig, error) {
	ctx, span := otel.AddSpan(ctx, "business.tableconfigbus.exportbyids")
	defer span.End()

	var results []TableConfig

	for _, configID := range configIDs {
		config, err := b.storer.QueryByID(ctx, configID)
		if err != nil {
			return nil, fmt.Errorf("query table config %s: %w", configID, err)
		}
		results = append(results, config)
	}

	return results, nil
}
```

#### 1c. Add Import Method to `tableconfigbus.go`

```go
// ImportTableConfigs imports table configs with conflict resolution.
func (b *Business) ImportTableConfigs(ctx context.Context, configs []TableConfig, mode string) (ImportStats, error) {
	ctx, span := otel.AddSpan(ctx, "business.tableconfigbus.importtableconfigs")
	defer span.End()

	stats := ImportStats{}

	for _, config := range configs {
		// Check if config exists by name
		existing, err := b.storer.QueryByName(ctx, config.Name)
		existsAlready := err == nil

		switch mode {
		case "skip":
			if existsAlready {
				stats.SkippedCount++
				continue
			}
			// Create new
			newConfig := NewTableConfig{
				Name:   config.Name,
				Config: config.Config,
			}
			if _, err := b.Create(ctx, newConfig); err != nil {
				return stats, fmt.Errorf("create config: %w", err)
			}
			stats.ImportedCount++

		case "replace":
			if existsAlready {
				// Delete existing and create new
				if err := b.Delete(ctx, existing); err != nil {
					return stats, fmt.Errorf("delete existing: %w", err)
				}
				stats.UpdatedCount++
			}
			newConfig := NewTableConfig{
				Name:   config.Name,
				Config: config.Config,
			}
			if _, err := b.Create(ctx, newConfig); err != nil {
				return stats, fmt.Errorf("create config: %w", err)
			}
			if !existsAlready {
				stats.ImportedCount++
			}

		case "merge":
			if existsAlready {
				// Update existing config
				updateConfig := UpdateTableConfig{
					Name:   &config.Name,
					Config: &config.Config,
				}
				if _, err := b.Update(ctx, existing, updateConfig); err != nil {
					return stats, fmt.Errorf("update config: %w", err)
				}
				stats.UpdatedCount++
			} else {
				// Create new
				newConfig := NewTableConfig{
					Name:   config.Name,
					Config: config.Config,
				}
				if _, err := b.Create(ctx, newConfig); err != nil {
					return stats, fmt.Errorf("create config: %w", err)
				}
				stats.ImportedCount++
			}
		}
	}

	return stats, nil
}
```

**Key Differences from Forms:**
- No child records to handle
- Simpler create/update logic
- Direct JSONB config field

---

### Step 2: Application Layer

**Files to modify:**
- `app/domain/data/tableconfigapp/model.go`
- `app/domain/data/tableconfigapp/tableconfigapp.go`

#### 2a. Add Export/Import Models to `model.go`

```go
// Add after existing models

// ExportPackage represents a JSON export package for table configs.
type ExportPackage struct {
	Version    string        `json:"version"`
	Type       string        `json:"type"`
	ExportedAt string        `json:"exportedAt"`
	Count      int           `json:"count"`
	Data       []TableConfig `json:"data"`
}

func (app ExportPackage) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ImportPackage represents a JSON import package for table configs.
type ImportPackage struct {
	Mode string        `json:"mode"` // "merge", "skip", "replace"
	Data []TableConfig `json:"data"`
}

func (app *ImportPackage) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app ImportPackage) Validate() error {
	if app.Mode != "merge" && app.Mode != "skip" && app.Mode != "replace" {
		return errs.Newf(errs.InvalidArgument, "mode must be 'merge', 'skip', or 'replace'")
	}

	if len(app.Data) == 0 {
		return errs.Newf(errs.InvalidArgument, "data cannot be empty")
	}

	// Validate each config
	for i, config := range app.Data {
		if config.Name == "" {
			return errs.Newf(errs.InvalidArgument, "config %d: name is required", i)
		}
	}

	return nil
}

// ImportResult represents the result of an import operation.
type ImportResult struct {
	ImportedCount int      `json:"importedCount"`
	SkippedCount  int      `json:"skippedCount"`
	UpdatedCount  int      `json:"updatedCount"`
	Errors        []string `json:"errors,omitempty"`
}

func (app ImportResult) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToBusTableConfig converts an app TableConfig to a business TableConfig.
func ToBusTableConfig(app TableConfig) (tableconfigbus.TableConfig, error) {
	configID, err := uuid.Parse(app.ID)
	if err != nil {
		// Generate new ID if parsing fails (import scenario)
		configID = uuid.New()
	}

	return tableconfigbus.TableConfig{
		ID:     configID,
		Name:   app.Name,
		Config: app.Config,
	}, nil
}
```

**Add missing imports:**
```go
import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/data/tableconfigbus"
)
```

#### 2b. Add Export/Import Methods to `tableconfigapp.go`

```go
// Add at end of file

// ExportByIDs exports table configs by IDs as a JSON package.
func (a *App) ExportByIDs(ctx context.Context, configIDs []string) (ExportPackage, error) {
	// Convert string IDs to UUIDs
	uuids := make([]uuid.UUID, len(configIDs))
	for i, id := range configIDs {
		uid, err := uuid.Parse(id)
		if err != nil {
			return ExportPackage{}, errs.Newf(errs.InvalidArgument, "invalid config ID %s: %s", id, err)
		}
		uuids[i] = uid
	}

	// Export from business layer
	results, err := a.tableconfigbus.ExportByIDs(ctx, uuids)
	if err != nil {
		return ExportPackage{}, errs.Newf(errs.Internal, "export: %s", err)
	}

	// Convert to app models
	configs := ToAppTableConfigs(results)

	return ExportPackage{
		Version:    "1.0",
		Type:       "table-configs",
		ExportedAt: time.Now().Format(time.RFC3339),
		Count:      len(configs),
		Data:       configs,
	}, nil
}

// ImportTableConfigs imports table configs from a JSON package.
func (a *App) ImportTableConfigs(ctx context.Context, pkg ImportPackage) (ImportResult, error) {
	// Validate package
	if err := pkg.Validate(); err != nil {
		return ImportResult{}, err
	}

	// Convert app models to business models
	var busConfigs []tableconfigbus.TableConfig
	for i, config := range pkg.Data {
		busConfig, err := ToBusTableConfig(config)
		if err != nil {
			return ImportResult{
				Errors: []string{err.Error()},
			}, errs.Newf(errs.InvalidArgument, "convert config %d: %s", i, err)
		}
		busConfigs = append(busConfigs, busConfig)
	}

	// Import via business layer
	stats, err := a.tableconfigbus.ImportTableConfigs(ctx, busConfigs, pkg.Mode)
	if err != nil {
		return ImportResult{
			Errors: []string{err.Error()},
		}, errs.Newf(errs.Internal, "import: %s", err)
	}

	return ImportResult{
		ImportedCount: stats.ImportedCount,
		SkippedCount:  stats.SkippedCount,
		UpdatedCount:  stats.UpdatedCount,
	}, nil
}
```

---

### Step 3: API Layer

**Files to modify:**
- `api/domain/http/data/tableconfigapi/tableconfigapi.go`
- `api/domain/http/data/tableconfigapi/routes.go`

#### 3a. Add Handlers to `tableconfigapi.go`

```go
// Add at end of file

func (api *api) exportTableConfigs(ctx context.Context, r *http.Request) web.Encoder {
	var req struct {
		IDs []string `json:"ids"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if len(req.IDs) == 0 {
		return errs.New(errs.InvalidArgument, errs.Newf(errs.InvalidArgument, "ids cannot be empty"))
	}

	pkg, err := api.tableconfigapp.ExportByIDs(ctx, req.IDs)
	if err != nil {
		return errs.NewError(err)
	}

	return pkg
}

func (api *api) importTableConfigs(ctx context.Context, r *http.Request) web.Encoder {
	var pkg tableconfigapp.ImportPackage
	if err := web.Decode(r, &pkg); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	result, err := api.tableconfigapp.ImportTableConfigs(ctx, pkg)
	if err != nil {
		return errs.NewError(err)
	}

	return result
}
```

**Add missing imports:**
```go
import (
	"encoding/json"
	"io"
)
```

#### 3b. Register Routes in `routes.go`

```go
// Add before closing brace of Routes function

app.HandlerFunc(http.MethodPost, version, "/data/configs/export", api.exportTableConfigs, authen,
	mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

app.HandlerFunc(http.MethodPost, version, "/data/configs/import", api.importTableConfigs, authen,
	mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAny))
```

---

## Part 2: Page Configs Export/Import

### Architecture

**Complexity**: Page configs have TWO types of related records:
- `page_content` (sections/tabs, with parent/child relationships)
- `page_actions` (buttons/dropdowns)

**Challenge**: Nested content requires ID remapping during import.

### Database Schema
```sql
CREATE TABLE config.page_configs (
    id UUID PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    path TEXT,
    icon TEXT
);

CREATE TABLE config.page_content (
    id UUID PRIMARY KEY,
    page_config_id UUID REFERENCES config.page_configs(id),
    parent_id UUID REFERENCES config.page_content(id), -- Self-referencing!
    content_type TEXT,
    config JSONB,
    sort_order INT
);

CREATE TABLE config.page_actions (
    id UUID PRIMARY KEY,
    page_config_id UUID REFERENCES config.page_configs(id),
    action_type TEXT,
    label TEXT,
    config JSONB,
    sort_order INT
);
```

---

### Step 1: Business Layer

**Files to modify:**
- `business/domain/config/pageconfigbus/model.go`
- `business/domain/config/pageconfigbus/pageconfigbus.go`

#### 1a. Add Models to `model.go`

```go
// Add after UpdatePageConfig

// PageConfigWithRelations represents a page config with its content and actions.
type PageConfigWithRelations struct {
	PageConfig PageConfig
	Contents   []pagecontentbus.PageContent
	Actions    []pageactionbus.PageAction
}

// ImportStats represents statistics from an import operation.
type ImportStats struct {
	ImportedCount int
	SkippedCount  int
	UpdatedCount  int
}
```

**Add imports:**
```go
import (
	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus"
	"github.com/timmaaaz/ichor/business/domain/config/pagecontentbus"
)
```

#### 1b. Update Business Struct in `pageconfigbus.go`

```go
// Update Business struct to include child business layers
type Business struct {
	log            *logger.Logger
	storer         Storer
	delegate       *delegate.Delegate
	pageContentBus *pagecontentbus.Business
	pageActionBus  *pageactionbus.Business
}

// Update NewBusiness constructor
func NewBusiness(log *logger.Logger, delegate *delegate.Delegate, storer Storer, pageContentBus *pagecontentbus.Business, pageActionBus *pageactionbus.Business) *Business {
	return &Business{
		log:            log,
		delegate:       delegate,
		storer:         storer,
		pageContentBus: pageContentBus,
		pageActionBus:  pageActionBus,
	}
}

// Update NewWithTx
func (b *Business) NewWithTx(tx sqldb.CommitRollbacker) (*Business, error) {
	storer, err := b.storer.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	pageContentBus, err := b.pageContentBus.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	pageActionBus, err := b.pageActionBus.NewWithTx(tx)
	if err != nil {
		return nil, err
	}

	return &Business{
		log:            b.log,
		delegate:       b.delegate,
		storer:         storer,
		pageContentBus: pageContentBus,
		pageActionBus:  pageActionBus,
	}, nil
}
```

#### 1c. Add Export Method

```go
// ExportByIDs exports page configs with their content and actions by IDs.
func (b *Business) ExportByIDs(ctx context.Context, configIDs []uuid.UUID) ([]PageConfigWithRelations, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.exportbyids")
	defer span.End()

	var results []PageConfigWithRelations

	for _, configID := range configIDs {
		config, err := b.storer.QueryByID(ctx, configID)
		if err != nil {
			return nil, fmt.Errorf("query page config %s: %w", configID, err)
		}

		contents, err := b.pageContentBus.QueryByPageConfigID(ctx, configID)
		if err != nil {
			return nil, fmt.Errorf("query contents for page config %s: %w", configID, err)
		}

		actions, err := b.pageActionBus.QueryByPageConfigID(ctx, configID)
		if err != nil {
			return nil, fmt.Errorf("query actions for page config %s: %w", configID, err)
		}

		results = append(results, PageConfigWithRelations{
			PageConfig: config,
			Contents:   contents,
			Actions:    actions,
		})
	}

	return results, nil
}
```

#### 1d. Add Import Method with Nested Content Handling

```go
// ImportPageConfigs imports page configs with conflict resolution and nested content handling.
func (b *Business) ImportPageConfigs(ctx context.Context, packages []PageConfigWithRelations, mode string) (ImportStats, error) {
	ctx, span := otel.AddSpan(ctx, "business.pageconfigbus.importpageconfigs")
	defer span.End()

	stats := ImportStats{}

	for _, pkg := range packages {
		// Check if config exists by name
		existing, err := b.storer.QueryByName(ctx, pkg.PageConfig.Name)
		existsAlready := err == nil

		switch mode {
		case "skip":
			if existsAlready {
				stats.SkippedCount++
				continue
			}
			if err := b.createPageConfigWithRelations(ctx, pkg); err != nil {
				return stats, err
			}
			stats.ImportedCount++

		case "replace":
			if existsAlready {
				if err := b.Delete(ctx, existing); err != nil {
					return stats, fmt.Errorf("delete existing: %w", err)
				}
				stats.UpdatedCount++
			}
			if err := b.createPageConfigWithRelations(ctx, pkg); err != nil {
				return stats, err
			}
			if !existsAlready {
				stats.ImportedCount++
			}

		case "merge":
			if existsAlready {
				if err := b.updatePageConfigWithRelations(ctx, existing.ID, pkg); err != nil {
					return stats, fmt.Errorf("update page config: %w", err)
				}
				stats.UpdatedCount++
			} else {
				if err := b.createPageConfigWithRelations(ctx, pkg); err != nil {
					return stats, err
				}
				stats.ImportedCount++
			}
		}
	}

	return stats, nil
}

func (b *Business) createPageConfigWithRelations(ctx context.Context, pkg PageConfigWithRelations) error {
	// Create page config
	newConfig := NewPageConfig{
		Name: pkg.PageConfig.Name,
		Path: pkg.PageConfig.Path,
		Icon: pkg.PageConfig.Icon,
	}

	config, err := b.Create(ctx, newConfig)
	if err != nil {
		return fmt.Errorf("create page config: %w", err)
	}

	// Create contents with ID remapping for parent/child relationships
	if err := b.createContentsWithRemapping(ctx, config.ID, pkg.Contents); err != nil {
		return err
	}

	// Create actions
	for _, action := range pkg.Actions {
		newAction := pageactionbus.NewPageAction{
			PageConfigID: config.ID,
			ActionType:   action.ActionType,
			Label:        action.Label,
			Config:       action.Config,
			SortOrder:    action.SortOrder,
		}
		if _, err := b.pageActionBus.Create(ctx, newAction); err != nil {
			return fmt.Errorf("create action %s: %w", action.Label, err)
		}
	}

	return nil
}

func (b *Business) createContentsWithRemapping(ctx context.Context, pageConfigID uuid.UUID, contents []pagecontentbus.PageContent) error {
	// Map old IDs to new IDs for parent/child relationships
	idMap := make(map[uuid.UUID]uuid.UUID)

	// First pass: Create all parent contents (where ParentID is nil)
	for _, content := range contents {
		if content.ParentID == nil {
			newContent := pagecontentbus.NewPageContent{
				PageConfigID: pageConfigID,
				ParentID:     nil,
				ContentType:  content.ContentType,
				Config:       content.Config,
				SortOrder:    content.SortOrder,
			}
			created, err := b.pageContentBus.Create(ctx, newContent)
			if err != nil {
				return fmt.Errorf("create parent content: %w", err)
			}
			idMap[content.ID] = created.ID
		}
	}

	// Second pass: Create child contents with remapped ParentID
	for _, content := range contents {
		if content.ParentID != nil {
			// Remap parent ID
			newParentID, ok := idMap[*content.ParentID]
			if !ok {
				return fmt.Errorf("parent content %s not found in id map", *content.ParentID)
			}

			newContent := pagecontentbus.NewPageContent{
				PageConfigID: pageConfigID,
				ParentID:     &newParentID,
				ContentType:  content.ContentType,
				Config:       content.Config,
				SortOrder:    content.SortOrder,
			}
			created, err := b.pageContentBus.Create(ctx, newContent)
			if err != nil {
				return fmt.Errorf("create child content: %w", err)
			}
			idMap[content.ID] = created.ID
		}
	}

	return nil
}

func (b *Business) updatePageConfigWithRelations(ctx context.Context, configID uuid.UUID, pkg PageConfigWithRelations) error {
	// Get current config
	config, err := b.storer.QueryByID(ctx, configID)
	if err != nil {
		return fmt.Errorf("query config: %w", err)
	}

	// Update page config
	updateConfig := UpdatePageConfig{
		Name: &pkg.PageConfig.Name,
		Path: &pkg.PageConfig.Path,
		Icon: &pkg.PageConfig.Icon,
	}

	if _, err := b.Update(ctx, config, updateConfig); err != nil {
		return fmt.Errorf("update config: %w", err)
	}

	// Delete existing contents and actions
	existingContents, err := b.pageContentBus.QueryByPageConfigID(ctx, configID)
	if err != nil {
		return fmt.Errorf("query existing contents: %w", err)
	}

	for _, content := range existingContents {
		if err := b.pageContentBus.Delete(ctx, content); err != nil {
			return fmt.Errorf("delete content %s: %w", content.ID, err)
		}
	}

	existingActions, err := b.pageActionBus.QueryByPageConfigID(ctx, configID)
	if err != nil {
		return fmt.Errorf("query existing actions: %w", err)
	}

	for _, action := range existingActions {
		if err := b.pageActionBus.Delete(ctx, action); err != nil {
			return fmt.Errorf("delete action %s: %w", action.ID, err)
		}
	}

	// Recreate contents and actions
	if err := b.createContentsWithRemapping(ctx, configID, pkg.Contents); err != nil {
		return err
	}

	for _, action := range pkg.Actions {
		newAction := pageactionbus.NewPageAction{
			PageConfigID: configID,
			ActionType:   action.ActionType,
			Label:        action.Label,
			Config:       action.Config,
			SortOrder:    action.SortOrder,
		}
		if _, err := b.pageActionBus.Create(ctx, newAction); err != nil {
			return fmt.Errorf("create action %s: %w", action.Label, err)
		}
	}

	return nil
}
```

**Key Points:**
- **ID Remapping**: Track old → new IDs for parent/child relationships
- **Two-Pass Creation**: Parents first, then children with remapped ParentIDs
- **Complete Replacement**: Delete all existing content/actions before recreating

---

### Step 2: Application Layer

**Files to modify:**
- `app/domain/config/pageconfigapp/model.go`
- `app/domain/config/pageconfigapp/pageconfigapp.go`

#### 2a. Add Models to `model.go`

```go
// Add after existing models

// ExportPackage represents a JSON export package for page configs.
type ExportPackage struct {
	Version    string              `json:"version"`
	Type       string              `json:"type"`
	ExportedAt string              `json:"exportedAt"`
	Count      int                 `json:"count"`
	Data       []PageConfigPackage `json:"data"`
}

func (app ExportPackage) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// PageConfigPackage represents a page config with its content and actions.
type PageConfigPackage struct {
	PageConfig PageConfig                 `json:"pageConfig"`
	Contents   []pagecontentapp.PageContent `json:"contents"`
	Actions    []pageactionapp.PageAction   `json:"actions"`
}

// ImportPackage represents a JSON import package for page configs.
type ImportPackage struct {
	Mode string              `json:"mode"` // "merge", "skip", "replace"
	Data []PageConfigPackage `json:"data"`
}

func (app *ImportPackage) Decode(data []byte) error {
	return json.Unmarshal(data, &app)
}

func (app ImportPackage) Validate() error {
	if app.Mode != "merge" && app.Mode != "skip" && app.Mode != "replace" {
		return errs.Newf(errs.InvalidArgument, "mode must be 'merge', 'skip', or 'replace'")
	}

	if len(app.Data) == 0 {
		return errs.Newf(errs.InvalidArgument, "data cannot be empty")
	}

	for i, pkg := range app.Data {
		if pkg.PageConfig.Name == "" {
			return errs.Newf(errs.InvalidArgument, "page config %d: name is required", i)
		}
	}

	return nil
}

// ImportResult represents the result of an import operation.
type ImportResult struct {
	ImportedCount int      `json:"importedCount"`
	SkippedCount  int      `json:"skippedCount"`
	UpdatedCount  int      `json:"updatedCount"`
	Errors        []string `json:"errors,omitempty"`
}

func (app ImportResult) Encode() ([]byte, string, error) {
	data, err := json.Marshal(app)
	return data, "application/json", err
}

// ToBusPageConfigWithRelations converts an app PageConfigPackage to business PageConfigWithRelations.
func ToBusPageConfigWithRelations(app PageConfigPackage) (pageconfigbus.PageConfigWithRelations, error) {
	configID, err := uuid.Parse(app.PageConfig.ID)
	if err != nil {
		configID = uuid.New()
	}

	config := pageconfigbus.PageConfig{
		ID:   configID,
		Name: app.PageConfig.Name,
		Path: app.PageConfig.Path,
		Icon: app.PageConfig.Icon,
	}

	// Convert contents
	contents := make([]pagecontentbus.PageContent, len(app.Contents))
	for i, appContent := range app.Contents {
		contentID, _ := uuid.Parse(appContent.ID)
		if contentID == uuid.Nil {
			contentID = uuid.New()
		}

		var parentID *uuid.UUID
		if appContent.ParentID != "" {
			pid, _ := uuid.Parse(appContent.ParentID)
			if pid != uuid.Nil {
				parentID = &pid
			}
		}

		contents[i] = pagecontentbus.PageContent{
			ID:           contentID,
			PageConfigID: configID,
			ParentID:     parentID,
			ContentType:  appContent.ContentType,
			Config:       appContent.Config,
			SortOrder:    appContent.SortOrder,
		}
	}

	// Convert actions
	actions := make([]pageactionbus.PageAction, len(app.Actions))
	for i, appAction := range app.Actions {
		actionID, _ := uuid.Parse(appAction.ID)
		if actionID == uuid.Nil {
			actionID = uuid.New()
		}

		actions[i] = pageactionbus.PageAction{
			ID:           actionID,
			PageConfigID: configID,
			ActionType:   appAction.ActionType,
			Label:        appAction.Label,
			Config:       appAction.Config,
			SortOrder:    appAction.SortOrder,
		}
	}

	return pageconfigbus.PageConfigWithRelations{
		PageConfig: config,
		Contents:   contents,
		Actions:    actions,
	}, nil
}
```

**Add imports:**
```go
import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/timmaaaz/ichor/app/domain/config/pageactionapp"
	"github.com/timmaaaz/ichor/app/domain/config/pagecontentapp"
	"github.com/timmaaaz/ichor/app/sdk/errs"
	"github.com/timmaaaz/ichor/business/domain/config/pageactionbus"
	"github.com/timmaaaz/ichor/business/domain/config/pageconfigbus"
	"github.com/timmaaaz/ichor/business/domain/config/pagecontentbus"
)
```

#### 2b. Add Methods to `pageconfigapp.go`

```go
// Add at end of file

// ExportByIDs exports page configs by IDs as a JSON package.
func (a *App) ExportByIDs(ctx context.Context, configIDs []string) (ExportPackage, error) {
	// Convert string IDs to UUIDs
	uuids := make([]uuid.UUID, len(configIDs))
	for i, id := range configIDs {
		uid, err := uuid.Parse(id)
		if err != nil {
			return ExportPackage{}, errs.Newf(errs.InvalidArgument, "invalid config ID %s: %s", id, err)
		}
		uuids[i] = uid
	}

	// Export from business layer
	results, err := a.pageconfigbus.ExportByIDs(ctx, uuids)
	if err != nil {
		return ExportPackage{}, errs.Newf(errs.Internal, "export: %s", err)
	}

	// Convert to app models
	var packages []PageConfigPackage
	for _, result := range results {
		packages = append(packages, PageConfigPackage{
			PageConfig: ToAppPageConfig(result.PageConfig),
			Contents:   pagecontentapp.ToAppPageContents(result.Contents),
			Actions:    pageactionapp.ToAppPageActions(result.Actions),
		})
	}

	return ExportPackage{
		Version:    "1.0",
		Type:       "page-configs",
		ExportedAt: time.Now().Format(time.RFC3339),
		Count:      len(packages),
		Data:       packages,
	}, nil
}

// ImportPageConfigs imports page configs from a JSON package.
func (a *App) ImportPageConfigs(ctx context.Context, pkg ImportPackage) (ImportResult, error) {
	// Validate package
	if err := pkg.Validate(); err != nil {
		return ImportResult{}, err
	}

	// Convert app models to business models
	var busPackages []pageconfigbus.PageConfigWithRelations
	for i, configPkg := range pkg.Data {
		busPkg, err := ToBusPageConfigWithRelations(configPkg)
		if err != nil {
			return ImportResult{
				Errors: []string{err.Error()},
			}, errs.Newf(errs.InvalidArgument, "convert page config %d: %s", i, err)
		}
		busPackages = append(busPackages, busPkg)
	}

	// Import via business layer
	stats, err := a.pageconfigbus.ImportPageConfigs(ctx, busPackages, pkg.Mode)
	if err != nil {
		return ImportResult{
			Errors: []string{err.Error()},
		}, errs.Newf(errs.Internal, "import: %s", err)
	}

	return ImportResult{
		ImportedCount: stats.ImportedCount,
		SkippedCount:  stats.SkippedCount,
		UpdatedCount:  stats.UpdatedCount,
	}, nil
}
```

---

### Step 3: API Layer

**Files to modify:**
- `api/domain/http/config/pageconfigapi/pageconfigapi.go`
- `api/domain/http/config/pageconfigapi/routes.go`

#### 3a. Add Handlers

```go
// Add to pageconfigapi.go

func (api *api) exportPageConfigs(ctx context.Context, r *http.Request) web.Encoder {
	var req struct {
		IDs []string `json:"ids"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return errs.New(errs.InvalidArgument, err)
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &req); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	if len(req.IDs) == 0 {
		return errs.New(errs.InvalidArgument, errs.Newf(errs.InvalidArgument, "ids cannot be empty"))
	}

	pkg, err := api.pageconfigapp.ExportByIDs(ctx, req.IDs)
	if err != nil {
		return errs.NewError(err)
	}

	return pkg
}

func (api *api) importPageConfigs(ctx context.Context, r *http.Request) web.Encoder {
	var pkg pageconfigapp.ImportPackage
	if err := web.Decode(r, &pkg); err != nil {
		return errs.New(errs.InvalidArgument, err)
	}

	result, err := api.pageconfigapp.ImportPageConfigs(ctx, pkg)
	if err != nil {
		return errs.NewError(err)
	}

	return result
}
```

#### 3b. Register Routes

```go
// Add to routes.go

app.HandlerFunc(http.MethodPost, version, "/config/page-configs/export", api.exportPageConfigs, authen,
	mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Read, auth.RuleAny))

app.HandlerFunc(http.MethodPost, version, "/config/page-configs/import", api.importPageConfigs, authen,
	mid.Authorize(cfg.AuthClient, cfg.PermissionsBus, RouteTable, permissionsbus.Actions.Create, auth.RuleAny))
```

---

### Step 4: Wire Dependencies in all.go

**File**: `api/cmd/services/ichor/build/all/all.go`

Find where pageConfigBus is instantiated and update it:

```go
// Before (around line 378):
pageConfigBus := pageconfigbus.NewBusiness(cfg.Log, delegate, pageconfigdb.NewStore(cfg.Log, cfg.DB))

// After:
pageContentBus := pagecontentbus.NewBusiness(cfg.Log, delegate, pagecontentdb.NewStore(cfg.Log, cfg.DB))
pageActionBus := pageactionbus.NewBusiness(cfg.Log, delegate, pageactiondb.NewStore(cfg.Log, cfg.DB))
pageConfigBus := pageconfigbus.NewBusiness(cfg.Log, delegate, pageconfigdb.NewStore(cfg.Log, cfg.DB), pageContentBus, pageActionBus)
```

**Also update** `business/sdk/dbtest/dbtest.go` in the same way (around line 361).

---

## Part 3: Testing

### Build and Compile Test

```bash
# Test compilation
go test ./api/cmd/services/ichor/tests/config/formapi/... -v
go test ./api/cmd/services/ichor/tests/data/tableconfigapi/... -v
go test ./api/cmd/services/ichor/tests/config/pageconfigapi/... -v
```

### Manual Testing

#### Get Authentication Token
```bash
make token
export TOKEN=<your-token>
```

#### Test Table Configs Export
```bash
# Export table configs
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"ids": ["<table-config-id>"]}' \
  http://localhost:3000/v1/data/configs/export | jq
```

#### Test Table Configs Import
```bash
# Import with skip mode
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "mode": "skip",
    "data": [
      {
        "name": "test_table_config",
        "config": {"columns": ["id", "name"]}
      }
    ]
  }' \
  http://localhost:3000/v1/data/configs/import | jq
```

#### Test Page Configs Export
```bash
# Export page configs
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"ids": ["<page-config-id>"]}' \
  http://localhost:3000/v1/config/page-configs/export | jq
```

#### Test Page Configs Import
```bash
# Import page configs
curl -X POST \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "mode": "merge",
    "data": [
      {
        "pageConfig": {
          "name": "test_page",
          "path": "/test",
          "icon": "test-icon"
        },
        "contents": [
          {
            "contentType": "section",
            "config": {"title": "Test Section"},
            "sortOrder": 1
          }
        ],
        "actions": [
          {
            "actionType": "button",
            "label": "Test Action",
            "config": {},
            "sortOrder": 1
          }
        ]
      }
    ]
  }' \
  http://localhost:3000/v1/config/page-configs/import | jq
```

### Round-Trip Tests

For each domain, test export → import → verify:

```bash
# 1. Export
EXPORT=$(curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"ids": ["<id>"]}' \
  http://localhost:3000/v1/config/forms/export)

# 2. Modify name to avoid conflict
echo $EXPORT | jq '.data[0].form.name = "imported_test"' > import.json

# 3. Import
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d @import.json \
  http://localhost:3000/v1/config/forms/import | jq

# 4. Verify imported entity exists
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:3000/v1/config/forms | jq '.items[] | select(.name == "imported_test")'
```

---

## Success Criteria

### Compilation
- [ ] All files compile without errors
- [ ] Integration tests run successfully

### Functional Testing
- [ ] Table Configs export returns valid JSON
- [ ] Table Configs import creates/updates records
- [ ] Page Configs export includes content and actions
- [ ] Page Configs import preserves nested content relationships
- [ ] All three conflict modes work (skip, replace, merge)

### Data Integrity
- [ ] Round-trip tests pass for all domains
- [ ] Foreign key relationships maintained
- [ ] Parent/child content relationships preserved in page configs
- [ ] No duplicate names after import

---

## Common Issues & Solutions

### Issue: Missing QueryByPageConfigID Method
**Error**: `undefined: b.pageContentBus.QueryByPageConfigID`
**Solution**: Ensure pagecontentbus and pageactionbus have `QueryByPageConfigID` methods.

### Issue: Parent/Child Content Import Fails
**Error**: "parent content not found in id map"
**Solution**: Ensure two-pass creation: parents first (ParentID == nil), then children.

### Issue: Transaction Not Rolling Back
**Error**: Partial imports on error
**Solution**: Ensure all operations use the same transaction context via `NewWithTx`.

### Issue: Circular Dependency
**Error**: Content references parent that hasn't been created
**Solution**: Sort contents by ParentID (nil first) before creating.

---

## Estimated Time Breakdown

- **Table Configs**: 1-2 hours (simple, no related records)
- **Page Configs**: 3-4 hours (complex nested content)
- **Testing**: 2-3 hours (manual + round-trip tests)
- **Debugging**: 1-2 hours (edge cases)

**Total**: 7-11 hours

---

## Next Steps

After completing Phase 3B:
1. Run full test suite: `make test`
2. Manual test all 6 endpoints
3. Perform round-trip testing
4. Commit changes: `git commit -m "feat: add import/export for table-configs and page-configs"`
5. **Update Phase 3 status to COMPLETE**
6. Notify frontend team that all import/export endpoints are ready

---

## Reference: Forms Implementation

All Forms implementation is complete and serves as the reference pattern:
- [business/domain/config/formbus/formbus.go](business/domain/config/formbus/formbus.go:174-356)
- [app/domain/config/formapp/model.go](app/domain/config/formapp/model.go:137-249)
- [api/domain/http/config/formapi/formapi.go](api/domain/http/config/formapi/formapi.go:149-188)

Use these files as templates for Table Configs and Page Configs implementation.
