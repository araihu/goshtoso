# Structured Input Consolidation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace `keyvalue` and `triplet` with one public `structuredinput` component for repeatable structured form rows.

**Architecture:** Add `components/structuredinput` as the single repeatable-row editor, driven by column definitions and map-backed entries. Remove the old packages and route/docs surfaces, update `components/form` to expose one `StructuredInput` field, then regenerate templ output and the generated usage reference.

**Tech Stack:** Go 1.26+, templ, Alpine.js, Tailwind CSS v4, Playwright E2E, `go test`, `templ generate`, `go run ./scripts/skillgen`.

---

## File Structure

- Create `components/structuredinput/types.go`: public config types, defaults, column normalization, Alpine data literal builders, row defaults, class helpers.
- Create `components/structuredinput/structuredinput.templ`: templ renderer for rows, text columns, select columns, hidden structured submission inputs, add/remove controls.
- Create `components/structuredinput/types_test.go`: focused unit tests for normalization, JS escaping, nil entries, and row defaults.
- Modify `components/form/types.go`: replace `KeyValue` and `Triplet` config fields with `StructuredInput`.
- Modify `components/form/form.templ`: render `structuredinput.StructuredInput`.
- Replace `site/internal/pages/demo/components/keyvalue.templ` and `site/internal/pages/demo/components/triplet.templ` with one `site/internal/pages/demo/components/structuredinput.templ`.
- Modify `site/internal/pages/demo/components/registry.go`, `site/internal/pages/demo/layout.templ`, and `site/tests/e2e/sidebar_test.go`: remove old routes/nav and add `/components/structured-input`.
- Replace `site/tests/e2e/keyvalue_test.go` and `site/tests/e2e/triplet_test.go` with `site/tests/e2e/structuredinput_test.go`.
- Modify `docs/USAGE.md`: replace old component rows with `structuredinput`.
- Delete old source/generated files under `components/keyvalue`, `components/triplet`, `site/internal/pages/demo/components/keyvalue_templ.go`, and `site/internal/pages/demo/components/triplet_templ.go` after source migration.
- Regenerate `*_templ.go` files and `.agents/.claude/skills/using-goshtoso/components-reference.md`.

---

### Task 1: Add StructuredInput Types And Tests

**Files:**
- Create: `components/structuredinput/types.go`
- Create: `components/structuredinput/types_test.go`

- [ ] **Step 1: Create failing type tests**

Create `components/structuredinput/types_test.go`:

```go
package structuredinput

import (
	"strings"
	"testing"
)

func TestNormalizedColumnsDropsEmptyAndDuplicateKeys(t *testing.T) {
	cfg := Config{
		Columns: []Column{
			{Key: "", Placeholder: "skip"},
			{Key: "key", Placeholder: "first"},
			{Key: "key", Placeholder: "duplicate"},
			{Key: "effect", Type: ColumnSelect, Options: []Option{{Value: "NoSchedule", Label: "NoSchedule"}}},
		},
	}

	cols := cfg.NormalizedColumns()

	if len(cols) != 2 {
		t.Fatalf("len(cols) = %d, want 2", len(cols))
	}
	if cols[0].Key != "key" || cols[0].Type != ColumnText {
		t.Fatalf("first column = %#v, want key text column", cols[0])
	}
	if cols[1].Key != "effect" || cols[1].DefaultValue() != "NoSchedule" {
		t.Fatalf("second column = %#v, want effect select defaulting to first option", cols[1])
	}
}

func TestInitialEntriesNeverSerializesNull(t *testing.T) {
	cfg := Config{Name: "labels", Entries: nil}

	data := cfg.AlpineData()

	if strings.Contains(data, "entries: null") {
		t.Fatalf("AlpineData() = %s, must use [] for nil entries", data)
	}
	if !strings.Contains(data, "entries: []") {
		t.Fatalf("AlpineData() = %s, want entries: []", data)
	}
}

func TestAlpineDataEscapesSingleQuotedStrings(t *testing.T) {
	cfg := Config{
		Name: "labels",
		Columns: []Column{
			{Key: "key", Placeholder: "owner's key"},
		},
		Entries: []Entry{
			{"key": "team's app", "value": `web\api`},
		},
	}

	data := cfg.AlpineData()

	if !strings.Contains(data, `owner\'s key`) {
		t.Fatalf("AlpineData() = %s, want escaped placeholder", data)
	}
	if !strings.Contains(data, `team\'s app`) {
		t.Fatalf("AlpineData() = %s, want escaped entry value", data)
	}
	if !strings.Contains(data, `web\\api`) {
		t.Fatalf("AlpineData() = %s, want escaped backslash", data)
	}
}

func TestNewRowLiteralUsesColumnDefaults(t *testing.T) {
	cfg := Config{
		Columns: []Column{
			{Key: "key"},
			{Key: "effect", Type: ColumnSelect, Options: []Option{{Value: "NoSchedule", Label: "NoSchedule"}}},
			{Key: "priority", Default: "high"},
		},
	}

	row := cfg.NewRowLiteral()

	for _, want := range []string{"'key': ''", "'effect': 'NoSchedule'", "'priority': 'high'"} {
		if !strings.Contains(row, want) {
			t.Fatalf("NewRowLiteral() = %s, missing %s", row, want)
		}
	}
}

func TestColumnAccessorsUseBracketNotation(t *testing.T) {
	col := Column{Key: "app.kubernetes.io/name"}

	if got := col.EntryAccessor(); got != "entry['app.kubernetes.io/name']" {
		t.Fatalf("EntryAccessor() = %s, want bracket notation", got)
	}
	if got := col.NameBinding(); got != "name + '[' + index + '][app.kubernetes.io/name]'" {
		t.Fatalf("NameBinding() = %s, want structured input name binding", got)
	}
}
```

- [ ] **Step 2: Run tests and verify they fail**

Run:

```bash
go test ./components/structuredinput
```

Expected: FAIL because `components/structuredinput` does not exist yet.

- [ ] **Step 3: Implement structuredinput types**

Create `components/structuredinput/types.go`:

```go
package structuredinput

import "strings"

// ColumnType is the rendered control kind for a structured input column.
type ColumnType string

const (
	ColumnText   ColumnType = "text"
	ColumnSelect ColumnType = "select"
)

// Option is one selectable value for a select column.
type Option struct {
	Value string
	Label string
}

// Column describes one input rendered in every structured row.
type Column struct {
	Key         string
	Label       string
	Type        ColumnType
	Placeholder string
	Options     []Option
	Default     string
}

// Entry holds initial values for one structured row.
type Entry map[string]string

// Config configures the StructuredInput component.
type Config struct {
	ID       string
	Name     string
	Columns  []Column
	Entries  []Entry
	AddLabel string
	Disabled bool
	Class    string
}

// NormalizedColumns returns usable columns with stable defaults.
func (c Config) NormalizedColumns() []Column {
	seen := map[string]bool{}
	cols := make([]Column, 0, len(c.Columns))
	for _, col := range c.Columns {
		if col.Key == "" || seen[col.Key] {
			continue
		}
		seen[col.Key] = true
		if col.Type == "" {
			col.Type = ColumnText
		}
		cols = append(cols, col)
	}
	return cols
}

// DefaultValue returns the value used when a new row is added.
func (c Column) DefaultValue() string {
	if c.Default != "" {
		return c.Default
	}
	if c.Type == ColumnSelect && len(c.Options) > 0 {
		return c.Options[0].Value
	}
	return ""
}

// OptionLabel returns the visible label for an option.
func (o Option) OptionLabel() string {
	if o.Label != "" {
		return o.Label
	}
	return o.Value
}

// GetAddLabel returns the add button label with default.
func (c Config) GetAddLabel() string {
	if c.AddLabel != "" {
		return c.AddLabel
	}
	return "Add row"
}

// ContainerClasses returns CSS classes for the outer container.
func (c Config) ContainerClasses() string {
	base := "flex flex-col gap-2"
	if c.Class != "" {
		return base + " " + c.Class
	}
	return base
}

// AlpineData returns the x-data object literal.
func (c Config) AlpineData() string {
	return "{ name: '" + jsEscapeSingle(c.Name) + "', columns: " + columnsLiteral(c.NormalizedColumns()) + ", entries: " + entriesLiteral(c.Entries, c.NormalizedColumns()) + " }"
}

// NewRowLiteral returns the JavaScript object literal pushed by the add button.
func (c Config) NewRowLiteral() string {
	parts := make([]string, 0, len(c.NormalizedColumns()))
	for _, col := range c.NormalizedColumns() {
		parts = append(parts, "'"+jsEscapeSingle(col.Key)+"': '"+jsEscapeSingle(col.DefaultValue())+"'")
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

// EntryAccessor returns a JavaScript expression for this column's entry value.
func (c Column) EntryAccessor() string {
	return "entry['" + jsEscapeSingle(c.Key) + "']"
}

// NameBinding returns a JavaScript expression for this column's hidden input name.
func (c Column) NameBinding() string {
	return "name + '[' + index + '][" + jsEscapeSingle(c.Key) + "]'"
}

func columnsLiteral(cols []Column) string {
	if len(cols) == 0 {
		return "[]"
	}
	items := make([]string, 0, len(cols))
	for _, col := range cols {
		items = append(items, "{ key: '"+jsEscapeSingle(col.Key)+"', type: '"+jsEscapeSingle(string(col.Type))+"', placeholder: '"+jsEscapeSingle(col.Placeholder)+"' }")
	}
	return "[" + strings.Join(items, ", ") + "]"
}

func entriesLiteral(entries []Entry, cols []Column) string {
	if len(entries) == 0 {
		return "[]"
	}
	items := make([]string, 0, len(entries))
	for _, entry := range entries {
		values := make([]string, 0, len(cols))
		for _, col := range cols {
			values = append(values, "'"+jsEscapeSingle(col.Key)+"': '"+jsEscapeSingle(entry[col.Key])+"'")
		}
		items = append(items, "{ "+strings.Join(values, ", ")+" }")
	}
	return "[" + strings.Join(items, ", ") + "]"
}

func jsEscapeSingle(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '\'':
			b.WriteString(`\'`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
```

- [ ] **Step 4: Run type tests and verify they pass**

Run:

```bash
go test ./components/structuredinput
```

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```bash
git add components/structuredinput/types.go components/structuredinput/types_test.go
git commit -m "feat: add structured input types"
```

Expected: commit succeeds.

---

### Task 2: Render StructuredInput Component

**Files:**
- Create: `components/structuredinput/structuredinput.templ`
- Modify after generation: `components/structuredinput/structuredinput_templ.go`

- [ ] **Step 1: Add render tests for generated markup**

Extend `components/structuredinput/types_test.go` with:

```go
func TestStructuredInputRendersHiddenStructuredNames(t *testing.T) {
	var buf strings.Builder
	err := StructuredInput(Config{
		ID:   "labelsDemo",
		Name: "labels",
		Columns: []Column{
			{Key: "key", Placeholder: "key"},
			{Key: "value", Placeholder: "value"},
		},
		Entries: []Entry{{"key": "app", "value": "web"}},
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	html := buf.String()
	for _, want := range []string{
		`id="labelsDemo"`,
		`x-bind:name="name + '[' + index + '][key]'"`,
		`x-bind:name="name + '[' + index + '][value]'"`,
		`x-bind:value="entry['key']"`,
		`x-bind:value="entry['value']"`,
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered HTML missing %s:\n%s", want, html)
		}
	}
}

func TestStructuredInputRendersSelectColumn(t *testing.T) {
	var buf strings.Builder
	err := StructuredInput(Config{
		ID:   "taintsDemo",
		Name: "taints",
		Columns: []Column{
			{Key: "key"},
			{Key: "effect", Type: ColumnSelect, Options: []Option{{Value: "NoSchedule", Label: "NoSchedule"}}},
		},
		Entries: []Entry{{"key": "node", "effect": "NoSchedule"}},
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	html := buf.String()
	for _, want := range []string{`<select`, `x-model="entry['effect']"`, `<option value="NoSchedule">NoSchedule</option>`} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered HTML missing %s:\n%s", want, html)
		}
	}
}
```

Also add `context` to the test imports:

```go
import (
	"context"
	"strings"
	"testing"
)
```

- [ ] **Step 2: Run tests and verify they fail**

Run:

```bash
go test ./components/structuredinput
```

Expected: FAIL because `StructuredInput` is undefined.

- [ ] **Step 3: Implement the templ renderer**

Create `components/structuredinput/structuredinput.templ`:

```go
package structuredinput

// StructuredInput renders repeatable structured form rows powered by Alpine.js.
//
// Submitted hidden inputs use name[index][columnKey]=value.
templ StructuredInput(cfg Config) {
	{{ columns := cfg.NormalizedColumns() }}
	<div
		if cfg.ID != "" {
			id={ cfg.ID }
		}
		class={ cfg.ContainerClasses() }
		x-data={ cfg.AlpineData() }
	>
		<template x-for="(entry, index) in entries" x-bind:key="index">
			<div class="flex gap-2 items-center flex-wrap">
				for _, col := range columns {
					if col.Type == ColumnSelect {
						@selectColumn(col, cfg.Disabled)
					} else {
						@textColumn(col, cfg.Disabled)
					}
					<input
						type="hidden"
						x-bind:name={ col.NameBinding() }
						x-bind:value={ col.EntryAccessor() }
					/>
				}
				if !cfg.Disabled {
					<button
						type="button"
						x-on:click="entries.splice(index, 1)"
						class="shrink-0 p-2 rounded-radius text-on-surface-muted dark:text-on-surface-dark-muted hover:bg-outline/20 dark:hover:bg-outline-dark/20 hover:text-on-surface-strong dark:hover:text-on-surface-dark-strong transition"
						aria-label="Remove row"
					>
						<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M18 6 6 18"></path><path d="m6 6 12 12"></path></svg>
					</button>
				}
			</div>
		</template>
		if !cfg.Disabled {
			<button
				type="button"
				x-on:click={ "entries.push(" + cfg.NewRowLiteral() + ")" }
				class="self-start px-3 py-2 text-sm rounded-radius border border-dashed border-outline dark:border-outline-dark text-on-surface dark:text-on-surface-dark hover:bg-outline/10 dark:hover:bg-outline-dark/20 hover:border-outline-strong dark:hover:border-outline-dark-strong transition"
				data-add-row
			>
				{ cfg.GetAddLabel() }
			</button>
		}
	</div>
}

templ textColumn(col Column, disabled bool) {
	<input
		type="text"
		x-model={ col.EntryAccessor() }
		placeholder={ col.Placeholder }
		aria-label={ col.Label }
		class="flex-1 min-w-24 px-3 py-2 text-sm rounded-radius border border-outline dark:border-outline-dark bg-surface-alt dark:bg-surface-dark-alt/50 text-on-surface dark:text-on-surface-dark placeholder:text-on-surface-muted dark:placeholder:text-on-surface-dark-muted focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary dark:focus-visible:outline-primary-dark"
		if disabled {
			disabled
		}
	/>
}

templ selectColumn(col Column, disabled bool) {
	<div class="relative shrink-0">
		<select
			x-model={ col.EntryAccessor() }
			aria-label={ col.Label }
			class="appearance-none min-w-32 px-3 py-2 pr-9 text-sm rounded-radius border border-outline dark:border-outline-dark bg-surface-alt dark:bg-surface-dark-alt/50 text-on-surface dark:text-on-surface-dark focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary dark:focus-visible:outline-primary-dark"
			if disabled {
				disabled
			}
		>
			for _, opt := range col.Options {
				<option value={ opt.Value }>{ opt.OptionLabel() }</option>
			}
		</select>
		<div class="pointer-events-none absolute inset-y-0 right-3 flex items-center text-on-surface dark:text-on-surface-dark">
			<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="size-4" aria-hidden="true">
				<path fill-rule="evenodd" d="M5.22 8.22a.75.75 0 0 1 1.06 0L10 11.94l3.72-3.72a.75.75 0 1 1 1.06 1.06l-4.25 4.25a.75.75 0 0 1-1.06 0L5.22 9.28a.75.75 0 0 1 0-1.06Z" clip-rule="evenodd"></path>
			</svg>
		</div>
	</div>
}
```

- [ ] **Step 4: Generate templ output**

Run:

```bash
templ generate
```

Expected: creates `components/structuredinput/structuredinput_templ.go`.

- [ ] **Step 5: Run component tests**

Run:

```bash
go test ./components/structuredinput
```

Expected: PASS.

- [ ] **Step 6: Commit**

Run:

```bash
git add components/structuredinput
git commit -m "feat: render structured input"
```

Expected: commit succeeds.

---

### Task 3: Replace Form Integration

**Files:**
- Modify: `components/form/types.go`
- Modify: `components/form/form.templ`
- Generated: `components/form/form_templ.go`

- [ ] **Step 1: Add a form integration test**

Create `components/form/structuredinput_test.go`:

```go
package form

import (
	"context"
	"strings"
	"testing"

	"github.com/araihu/goshtoso/components/structuredinput"
)

func TestFieldGroupRendersStructuredInput(t *testing.T) {
	var buf strings.Builder
	err := FieldGroup(FieldGroupConfig{
		ID:    "labels",
		Label: "Labels",
		StructuredInput: &structuredinput.Config{
			ID:   "labelsInput",
			Name: "labels",
			Columns: []structuredinput.Column{
				{Key: "key"},
				{Key: "value"},
			},
			Entries: []structuredinput.Entry{{"key": "app", "value": "web"}},
		},
	}).Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}

	html := buf.String()
	for _, want := range []string{`Labels`, `id="labelsInput"`, `x-bind:name="name + '[' + index + '][key]'"`} {
		if !strings.Contains(html, want) {
			t.Fatalf("rendered HTML missing %s:\n%s", want, html)
		}
	}
}
```

- [ ] **Step 2: Run test and verify it fails**

Run:

```bash
go test ./components/form -run TestFieldGroupRendersStructuredInput
```

Expected: FAIL because `FieldGroupConfig.StructuredInput` is undefined.

- [ ] **Step 3: Update form types**

In `components/form/types.go`, replace:

```go
	"github.com/araihu/goshtoso/components/keyvalue"
	"github.com/araihu/goshtoso/components/triplet"
```

with:

```go
	"github.com/araihu/goshtoso/components/structuredinput"
```

Replace the field config section:

```go
	TagsList  *tagslist.Config
	KeyValue  *keyvalue.Config
	Triplet   *triplet.Config
	FileInput *fileinput.Config
```

with:

```go
	TagsList        *tagslist.Config
	StructuredInput *structuredinput.Config
	FileInput       *fileinput.Config
```

- [ ] **Step 4: Update form renderer**

In `components/form/form.templ`, replace imports for `keyvalue` and `triplet` with:

```go
	"github.com/araihu/goshtoso/components/structuredinput"
```

Replace:

```go
		} else if cfg.TagsList != nil {
			@tagslist.TagsList(*cfg.TagsList)
		} else if cfg.KeyValue != nil {
			@keyvalue.KeyValue(*cfg.KeyValue)
		} else if cfg.Triplet != nil {
			@triplet.Triplet(*cfg.Triplet)
		} else if cfg.FileInput != nil {
```

with:

```go
		} else if cfg.TagsList != nil {
			@tagslist.TagsList(*cfg.TagsList)
		} else if cfg.StructuredInput != nil {
			@structuredinput.StructuredInput(*cfg.StructuredInput)
		} else if cfg.FileInput != nil {
```

- [ ] **Step 5: Generate templ output**

Run:

```bash
templ generate
```

Expected: updates `components/form/form_templ.go`.

- [ ] **Step 6: Run form tests**

Run:

```bash
go test ./components/form
```

Expected: PASS.

- [ ] **Step 7: Commit**

Run:

```bash
git add components/form
git commit -m "feat: use structured input in forms"
```

Expected: commit succeeds.

---

### Task 4: Replace Demo Page And Navigation

**Files:**
- Create: `site/internal/pages/demo/components/structuredinput.templ`
- Modify: `site/internal/pages/demo/components/registry.go`
- Modify: `site/internal/pages/demo/layout.templ`
- Delete: `site/internal/pages/demo/components/keyvalue.templ`
- Delete: `site/internal/pages/demo/components/triplet.templ`
- Generated by templ: `site/internal/pages/demo/components/structuredinput_templ.go`, `site/internal/pages/demo/components/registry_templ.go` if present, `site/internal/pages/demo/layout_templ.go`

- [ ] **Step 1: Add a failing sidebar expectation**

In `site/tests/e2e/sidebar_test.go`, replace:

```go
		{"/components/key-value", "Key Value"},
		{"/components/triplet", "Triplet"},
```

with:

```go
		{"/components/structured-input", "Structured Input"},
```

Run:

```bash
go test ./site/tests/e2e/... -run TestSidebarContainsAllComponents -count=1 -timeout 5m
```

Expected: FAIL because navigation still lacks `Structured Input`.

- [ ] **Step 2: Create the structured input demo source**

Create `site/internal/pages/demo/components/structuredinput.templ`:

```go
package components

import (
	"github.com/araihu/goshtoso/components/structuredinput"
	"github.com/araihu/goshtoso/site/internal/pages/demo"
)

// StructuredInputDemoPage renders the Structured Input component demo.
templ StructuredInputDemoPage() {
	@demo.Layout("Structured Input", "structured-input", structuredInputDemoContent())
}

templ structuredInputDemoContent() {
	<div id="structured-input-fragment">
		@demo.ComponentDemo(
			demo.ComponentDemoProps{
				Title:       "Structured Input",
				Description: "Repeatable rows of structured form inputs for metadata, taints, rules, and similar list-shaped data.",
			},
			structuredInputKeyValuePreview(),
			`@structuredinput.StructuredInput(structuredinput.Config{
    ID:   "labels",
    Name: "labels",
    Columns: []structuredinput.Column{
        {Key: "key", Label: "Key", Placeholder: "key"},
        {Key: "value", Label: "Value", Placeholder: "value"},
    },
    Entries: []structuredinput.Entry{
        {"key": "app", "value": "web"},
        {"key": "env", "value": "prod"},
    },
})`,
		)

		@demo.DemoSection(
			demo.DemoSectionProps{
				Title:       "Select Column",
				Description: "Rows can combine text inputs with select columns for enumerated values like Kubernetes taint effects.",
			},
			structuredInputTaintsPreview(),
			`@structuredinput.StructuredInput(structuredinput.Config{
    ID:   "taints",
    Name: "taints",
    Columns: []structuredinput.Column{
        {Key: "key", Label: "Key", Placeholder: "key"},
        {Key: "value", Label: "Value", Placeholder: "value"},
        {
            Key: "effect",
            Label: "Effect",
            Type: structuredinput.ColumnSelect,
            Options: []structuredinput.Option{
                {Value: "NoSchedule", Label: "NoSchedule"},
                {Value: "PreferNoSchedule", Label: "PreferNoSchedule"},
                {Value: "NoExecute", Label: "NoExecute"},
            },
        },
    },
    Entries: []structuredinput.Entry{
        {"key": "node-role.kubernetes.io/control-plane", "value": "true", "effect": "NoSchedule"},
    },
    AddLabel: "Add taint",
})`,
		)

		@demo.DemoSection(
			demo.DemoSectionProps{
				Title:       "Empty Starter",
				Description: "With no entries the list starts empty and adds rows from the configured column defaults.",
			},
			structuredInputEmptyPreview(),
			`@structuredinput.StructuredInput(structuredinput.Config{
    ID:   "rules",
    Name: "rules",
    Columns: []structuredinput.Column{
        {Key: "name", Label: "Name", Placeholder: "rule name"},
        {Key: "condition", Label: "Condition", Placeholder: "condition"},
        {
            Key: "priority",
            Label: "Priority",
            Type: structuredinput.ColumnSelect,
            Options: []structuredinput.Option{
                {Value: "high", Label: "High"},
                {Value: "medium", Label: "Medium"},
                {Value: "low", Label: "Low"},
            },
        },
    },
    AddLabel: "Add rule",
})`,
		)
	</div>
	@demo.APIReference([]demo.PropDoc{
		{Name: "ID", Type: "string", Default: `""`, Description: "Unique id for the structured input root."},
		{Name: "Name", Type: "string", Default: `""`, Description: "Base form field name; submitted inputs use name[index][columnKey]."},
		{Name: "Columns", Type: "[]Column", Default: "nil", Description: "Column schema for every repeatable row."},
		{Name: "Entries", Type: "[]Entry", Default: "nil", Description: "Initial row values keyed by column key."},
		{Name: "AddLabel", Type: "string", Default: `"Add row"`, Description: "Label of the add-row button."},
		{Name: "Disabled", Type: "bool", Default: "false", Description: "Disable controls and hide add/remove actions."},
		{Name: "Class", Type: "string", Default: `""`, Description: "Extra classes on the container."},
	})
}

templ structuredInputKeyValuePreview() {
	<div id="structured-input-key-value" class="w-full max-w-lg mx-auto">
		@structuredinput.StructuredInput(structuredinput.Config{
			ID:   "labelsDemo",
			Name: "labels",
			Columns: []structuredinput.Column{
				{Key: "key", Label: "Key", Placeholder: "key"},
				{Key: "value", Label: "Value", Placeholder: "value"},
			},
			Entries: []structuredinput.Entry{
				{"key": "app", "value": "web"},
				{"key": "env", "value": "prod"},
			},
		})
	</div>
}

templ structuredInputTaintsPreview() {
	<div id="structured-input-taints" class="w-full max-w-2xl mx-auto">
		@structuredinput.StructuredInput(structuredinput.Config{
			ID:   "taintsDemo",
			Name: "taints",
			Columns: []structuredinput.Column{
				{Key: "key", Label: "Key", Placeholder: "key (e.g. node-role.kubernetes.io/control-plane)"},
				{Key: "value", Label: "Value", Placeholder: "value (e.g. true)"},
				{
					Key:   "effect",
					Label: "Effect",
					Type:  structuredinput.ColumnSelect,
					Options: []structuredinput.Option{
						{Value: "NoSchedule", Label: "NoSchedule"},
						{Value: "PreferNoSchedule", Label: "PreferNoSchedule"},
						{Value: "NoExecute", Label: "NoExecute"},
					},
				},
			},
			Entries: []structuredinput.Entry{
				{"key": "node-role.kubernetes.io/control-plane", "value": "true", "effect": "NoSchedule"},
			},
			AddLabel: "Add taint",
		})
	</div>
}

templ structuredInputEmptyPreview() {
	<div id="structured-input-empty" class="w-full max-w-2xl mx-auto">
		@structuredinput.StructuredInput(structuredinput.Config{
			ID:   "rulesDemo",
			Name: "rules",
			Columns: []structuredinput.Column{
				{Key: "name", Label: "Name", Placeholder: "rule name"},
				{Key: "condition", Label: "Condition", Placeholder: "condition"},
				{
					Key:   "priority",
					Label: "Priority",
					Type:  structuredinput.ColumnSelect,
					Options: []structuredinput.Option{
						{Value: "high", Label: "High"},
						{Value: "medium", Label: "Medium"},
						{Value: "low", Label: "Low"},
					},
				},
			},
			AddLabel: "Add rule",
		})
	</div>
}
```

- [ ] **Step 3: Update registry**

In `site/internal/pages/demo/components/registry.go`, replace:

```go
	"components/key-value":       {"Key Value", "key-value", keyValueDemoContent},
```

and:

```go
	"components/triplet":         {"Triplet", "triplet", tripletDemoContent},
```

with:

```go
	"components/structured-input": {"Structured Input", "structured-input", structuredInputDemoContent},
```

- [ ] **Step 4: Update sidebar and ordered component nav**

In `site/internal/pages/demo/layout.templ`, replace the form sidebar entries:

```go
				sItem("key-value", "Key Value", "/components/key-value", activeComponent),
				sItem("triplet", "Triplet", "/components/triplet", activeComponent),
```

with:

```go
				sItem("structured-input", "Structured Input", "/components/structured-input", activeComponent),
```

In `orderedComponents`, replace:

```go
	{"Key Value", "/components/key-value"},
	{"Triplet", "/components/triplet"},
```

with:

```go
	{"Structured Input", "/components/structured-input"},
```

- [ ] **Step 5: Delete old demo sources**

Run:

```bash
rm site/internal/pages/demo/components/keyvalue.templ site/internal/pages/demo/components/triplet.templ
```

Expected: old demo source files are removed.

- [ ] **Step 6: Generate templ output**

Run:

```bash
templ generate
```

Expected: creates `site/internal/pages/demo/components/structuredinput_templ.go` and updates `site/internal/pages/demo/layout_templ.go`. If generated files for deleted demos remain, remove `site/internal/pages/demo/components/keyvalue_templ.go` and `site/internal/pages/demo/components/triplet_templ.go`.

- [ ] **Step 7: Run sidebar test**

Run:

```bash
go test ./site/tests/e2e/... -run TestSidebarContainsAllComponents -count=1 -timeout 5m
```

Expected: PASS.

- [ ] **Step 8: Commit**

Run:

```bash
git add site/internal/pages/demo/components site/internal/pages/demo/layout.templ site/internal/pages/demo/layout_templ.go site/tests/e2e/sidebar_test.go
git commit -m "docs: add structured input demo"
```

Expected: commit succeeds.

---

### Task 5: Replace E2E Coverage For The Component

**Files:**
- Create: `site/tests/e2e/structuredinput_test.go`
- Delete: `site/tests/e2e/keyvalue_test.go`
- Delete: `site/tests/e2e/triplet_test.go`

- [ ] **Step 1: Create structured input E2E tests**

Create `site/tests/e2e/structuredinput_test.go`:

```go
package e2e

import (
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStructuredInput_AddAndRemoveKeyValueRows(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	_, err := page.Goto(baseURL+"/components/structured-input", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	container := page.Locator("#labelsDemo")
	require.NoError(t, container.WaitFor())

	hiddenInputs := container.Locator("input[type='hidden']")
	count, err := hiddenInputs.Count()
	require.NoError(t, err)
	assert.Equal(t, 4, count, "two rows with two columns should render four hidden inputs")

	name, err := hiddenInputs.First().GetAttribute("name")
	require.NoError(t, err)
	assert.Equal(t, "labels[0][key]", name)

	addBtn := container.Locator("[data-add-row]")
	require.NoError(t, addBtn.Click())
	count, err = hiddenInputs.Count()
	require.NoError(t, err)
	assert.Equal(t, 6, count, "adding one two-column row should add two hidden inputs")

	removeBtn := container.Locator("button[aria-label='Remove row']").First()
	require.NoError(t, removeBtn.Click())
	count, err = hiddenInputs.Count()
	require.NoError(t, err)
	assert.Equal(t, 4, count, "removing one two-column row should remove two hidden inputs")
}

func TestStructuredInput_SelectColumnDefaults(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}
	cleanupServer := setupServer(t)
	defer cleanupServer()
	_, browser, cleanupPW := setupPlaywright(t)
	defer cleanupPW()

	page := newPage(t, browser)
	_, err := page.Goto(baseURL+"/components/structured-input", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	})
	require.NoError(t, err)
	_, err = page.WaitForFunction("() => typeof Alpine !== 'undefined'", nil)
	require.NoError(t, err)

	container := page.Locator("#rulesDemo")
	require.NoError(t, container.WaitFor())

	hiddenInputs := container.Locator("input[type='hidden']")
	count, err := hiddenInputs.Count()
	require.NoError(t, err)
	assert.Equal(t, 0, count, "empty starter should start with no rows")

	require.NoError(t, container.Locator("[data-add-row]").Click())

	selects := container.Locator("select")
	selectCount, err := selects.Count()
	require.NoError(t, err)
	assert.Equal(t, 1, selectCount, "new row should include the select column")

	value, err := selects.First().InputValue()
	require.NoError(t, err)
	assert.Equal(t, "high", value, "select should default to first option")

	count, err = hiddenInputs.Count()
	require.NoError(t, err)
	assert.Equal(t, 3, count, "one three-column row should render three hidden inputs")
}
```

- [ ] **Step 2: Delete old E2E tests**

Run:

```bash
rm site/tests/e2e/keyvalue_test.go site/tests/e2e/triplet_test.go
```

Expected: old route-specific tests are removed.

- [ ] **Step 3: Run focused E2E tests**

Run:

```bash
go test ./site/tests/e2e/... -run 'TestStructuredInput|TestSidebarContainsAllComponents' -count=1 -timeout 5m
```

Expected: PASS.

- [ ] **Step 4: Commit**

Run:

```bash
git add site/tests/e2e
git commit -m "test: cover structured input"
```

Expected: commit succeeds.

---

### Task 6: Remove Old Component Packages And Usage Docs

**Files:**
- Delete: `components/keyvalue/keyvalue.templ`
- Delete: `components/keyvalue/keyvalue_templ.go`
- Delete: `components/keyvalue/types.go`
- Delete: `components/triplet/triplet.templ`
- Delete: `components/triplet/triplet_templ.go`
- Delete: `components/triplet/types.go`
- Modify: `docs/USAGE.md`
- Generated: `.agents/skills/using-goshtoso/components-reference.md`
- Generated: `.claude/skills/using-goshtoso/components-reference.md`

- [ ] **Step 1: Remove old packages**

Run:

```bash
rm -rf components/keyvalue components/triplet
```

Expected: old component packages are deleted.

- [ ] **Step 2: Update docs usage component list**

In `docs/USAGE.md`, replace:

```markdown
| `keyvalue` | `components/keyvalue` | Dynamic key-value pair editor (for labels, env vars) |
| `triplet` | `components/triplet` | Key-value-effect editor (for Kubernetes taints) |
```

with:

```markdown
| `structuredinput` | `components/structuredinput` | Repeatable structured row editor (for labels, taints, rules) |
```

- [ ] **Step 3: Regenerate usage skill reference**

Run:

```bash
go run ./scripts/skillgen
```

Expected: `.agents/skills/using-goshtoso/components-reference.md` and `.claude/skills/using-goshtoso/components-reference.md` mention `structuredinput` and no longer mention `keyvalue` or `triplet`.

- [ ] **Step 4: Search for stale references**

Run:

```bash
rg -n "keyvalue|KeyValue|triplet|Triplet|key-value|components/triplet|components/keyvalue" components site docs .agents .claude
```

Expected: no stale references except historical text in committed plan/spec files under `docs/superpowers`.

- [ ] **Step 5: Run root package tests**

Run:

```bash
go test ./components/structuredinput ./components/form
```

Expected: PASS.

- [ ] **Step 6: Commit**

Run:

```bash
git add components docs/USAGE.md .agents/skills/using-goshtoso/components-reference.md .claude/skills/using-goshtoso/components-reference.md
git commit -m "refactor: remove keyvalue and triplet"
```

Expected: commit succeeds.

---

### Task 7: Final Generation, Build, And Verification

**Files:**
- Verify all touched generated files.
- No new source files unless a previous task exposed a small compile fix.

- [ ] **Step 1: Run final templ generation**

Run:

```bash
templ generate
```

Expected: completes without errors and leaves no unexpected generated drift after review.

- [ ] **Step 2: Run root tests**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 3: Run site tests excluding full E2E**

Run:

```bash
(cd site && go test ./internal/... ./cmd/...)
```

Expected: PASS.

- [ ] **Step 4: Build server**

Run:

```bash
go build -o bin/server ./site/cmd/server
```

Expected: PASS and writes `bin/server`.

- [ ] **Step 5: Run focused E2E**

Run:

```bash
go test ./site/tests/e2e/... -run 'TestStructuredInput|TestSidebarContainsAllComponents' -count=1 -timeout 5m
```

Expected: PASS.

- [ ] **Step 6: Run stale-reference search**

Run:

```bash
rg -n "components/keyvalue|components/triplet|KeyValue|Triplet|key-value|/components/triplet|/components/key-value" components site docs .agents .claude
```

Expected: no stale references outside `docs/superpowers/specs/2026-06-03-structured-input-consolidation-design.md` and this plan.

- [ ] **Step 7: Inspect git diff**

Run:

```bash
git status --short
git diff --stat HEAD
```

Expected: only intended source, generated, docs, and test changes are present.

- [ ] **Step 8: Commit any final verification drift**

If Step 1 or Step 3 produced intended changes, run:

```bash
git add .
git commit -m "chore: verify structured input consolidation"
```

Expected: commit succeeds, or no commit is needed because there is no drift.

---

## Self-Review

- Spec coverage: the plan adds the new component, removes old public packages, updates form integration, replaces docs/navigation, changes submission shape to `name[index][columnKey]`, regenerates usage docs, and verifies stale references.
- Placeholder scan: no task uses vague placeholder wording; every code-changing step includes exact code or exact replacement text.
- Type consistency: the plan consistently uses package `structuredinput`, entry point `StructuredInput`, config field `StructuredInput`, type `Entry map[string]string`, `ColumnText`, and `ColumnSelect`.
