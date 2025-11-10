package pagecontentbus

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// =============================================================================
// Content Type Constants
// =============================================================================

// Content type constants
const (
	ContentTypeTable     = "table"
	ContentTypeForm      = "form"
	ContentTypeTabs      = "tabs"
	ContentTypeContainer = "container"
	ContentTypeText      = "text"
	ContentTypeChart     = "chart"
)

// Container type constants
const (
	ContainerTypeTab       = "tab"
	ContainerTypeAccordion = "accordion"
	ContainerTypeSection   = "section"
	ContainerTypeGrid      = "grid"
)

// =============================================================================
// Business Models
// =============================================================================

// PageContent represents a flexible content block on a page
type PageContent struct {
	ID            uuid.UUID
	PageConfigID  uuid.UUID
	ContentType   string
	Label         string
	TableConfigID uuid.UUID
	FormID        uuid.UUID
	OrderIndex    int
	ParentID      uuid.UUID
	Layout        json.RawMessage
	IsVisible     bool
	IsDefault     bool
	Children      []PageContent // Populated by queries, not stored in DB
}

// NewPageContent contains data required to create a new page content block
type NewPageContent struct {
	PageConfigID  uuid.UUID
	ContentType   string
	Label         string
	TableConfigID uuid.UUID
	FormID        uuid.UUID
	OrderIndex    int
	ParentID      uuid.UUID
	Layout        json.RawMessage
	IsVisible     bool
	IsDefault     bool
}

// UpdatePageContent contains data for updating an existing page content block
type UpdatePageContent struct {
	Label      *string
	OrderIndex *int
	Layout     *json.RawMessage
	IsVisible  *bool
	IsDefault  *bool
}

// LayoutConfig holds all layout/styling configuration (stored as JSONB)
type LayoutConfig struct {
	// Responsive column spans (Tailwind col-span-*)
	ColSpan *ResponsiveValue `json:"colSpan,omitempty"`

	// Row span
	RowSpan int `json:"rowSpan,omitempty"`

	// Explicit grid positioning
	ColStart *int `json:"colStart,omitempty"`
	RowStart *int `json:"rowStart,omitempty"`

	// Container-specific (if this content is a container)
	GridCols *ResponsiveValue `json:"gridCols,omitempty"`
	Gap      string           `json:"gap,omitempty"` // Tailwind gap class: "gap-4", "gap-x-4 gap-y-6"

	// Additional Tailwind classes
	ClassName string `json:"className,omitempty"`

	// Container behavior
	ContainerType string `json:"containerType,omitempty"` // "tab", "accordion", "section", "grid"

	// Display options
	Collapsible bool `json:"collapsible,omitempty"`
}

// ResponsiveValue holds mobile-first responsive values (Tailwind breakpoints)
type ResponsiveValue struct {
	Default int  `json:"default"`
	Sm      *int `json:"sm,omitempty"`
	Md      *int `json:"md,omitempty"`
	Lg      *int `json:"lg,omitempty"`
	Xl      *int `json:"xl,omitempty"`
	Xl2     *int `json:"2xl,omitempty"`
}

// =============================================================================
// Helper Methods
// =============================================================================

// GetTailwindClasses generates Tailwind classes from layout config
func (lc *LayoutConfig) GetTailwindClasses() string {
	classes := []string{}

	// Column span
	if lc.ColSpan != nil {
		classes = append(classes, lc.ColSpan.ToColSpanClasses()...)
	}

	// Row span
	if lc.RowSpan > 0 {
		classes = append(classes, fmt.Sprintf("row-span-%d", lc.RowSpan))
	}

	// Positioning
	if lc.ColStart != nil {
		classes = append(classes, fmt.Sprintf("col-start-%d", *lc.ColStart))
	}
	if lc.RowStart != nil {
		classes = append(classes, fmt.Sprintf("row-start-%d", *lc.RowStart))
	}

	// Custom classes
	if lc.ClassName != "" {
		classes = append(classes, lc.ClassName)
	}

	return strings.Join(classes, " ")
}

// GetContainerClasses generates container Tailwind classes
func (lc *LayoutConfig) GetContainerClasses() string {
	if lc.GridCols == nil {
		return ""
	}

	classes := []string{"grid"}
	classes = append(classes, lc.GridCols.ToGridColsClasses()...)

	if lc.Gap != "" {
		classes = append(classes, lc.Gap)
	}

	return strings.Join(classes, " ")
}

// ToColSpanClasses converts responsive values to col-span-* classes
func (rv *ResponsiveValue) ToColSpanClasses() []string {
	classes := []string{fmt.Sprintf("col-span-%d", rv.Default)}

	if rv.Sm != nil {
		classes = append(classes, fmt.Sprintf("sm:col-span-%d", *rv.Sm))
	}
	if rv.Md != nil {
		classes = append(classes, fmt.Sprintf("md:col-span-%d", *rv.Md))
	}
	if rv.Lg != nil {
		classes = append(classes, fmt.Sprintf("lg:col-span-%d", *rv.Lg))
	}
	if rv.Xl != nil {
		classes = append(classes, fmt.Sprintf("xl:col-span-%d", *rv.Xl))
	}
	if rv.Xl2 != nil {
		classes = append(classes, fmt.Sprintf("2xl:col-span-%d", *rv.Xl2))
	}

	return classes
}

// ToGridColsClasses converts responsive values to grid-cols-* classes
func (rv *ResponsiveValue) ToGridColsClasses() []string {
	classes := []string{fmt.Sprintf("grid-cols-%d", rv.Default)}

	if rv.Sm != nil {
		classes = append(classes, fmt.Sprintf("sm:grid-cols-%d", *rv.Sm))
	}
	if rv.Md != nil {
		classes = append(classes, fmt.Sprintf("md:grid-cols-%d", *rv.Md))
	}
	if rv.Lg != nil {
		classes = append(classes, fmt.Sprintf("lg:grid-cols-%d", *rv.Lg))
	}
	if rv.Xl != nil {
		classes = append(classes, fmt.Sprintf("xl:grid-cols-%d", *rv.Xl))
	}
	if rv.Xl2 != nil {
		classes = append(classes, fmt.Sprintf("2xl:grid-cols-%d", *rv.Xl2))
	}

	return classes
}
