package structuredinput

import (
	"encoding/json"
	"strconv"
)

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
	Separator   string
	Options     []Option
	Default     string
}

// Entry holds initial values for one structured row.
type Entry map[string]string

// Config configures the StructuredInput component.
type Config struct {
	ID             string
	Name           string
	Columns        []Column
	Entries        []Entry
	AddActionLabel string
	Disabled       bool
	RootClass      string
}

// NormalizedColumns returns usable columns with stable defaults.
func (c Config) normalizedColumns() []Column {
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
func (c Column) defaultValue() string {
	if c.Default != "" {
		return c.Default
	}
	if c.Type == ColumnSelect && len(c.Options) > 0 {
		return c.Options[0].Value
	}
	return ""
}

// OptionLabel returns the visible label for an option.
func (o Option) optionLabel() string {
	if o.Label != "" {
		return o.Label
	}
	return o.Value
}

// GetAddLabel returns the add button label with default.
func (c Config) getAddLabel() string {
	if c.AddActionLabel != "" {
		return c.AddActionLabel
	}
	return "Add row"
}

// ContainerClasses returns CSS classes for the outer container.
func (c Config) containerClasses() string {
	base := "flex flex-col gap-2"
	if c.RootClass != "" {
		return base + " " + c.RootClass
	}
	return base
}

// EntriesJSON returns ordered row values for Alpine initialization.
func (c Config) entriesJSON() string {
	rows := c.entryRows()
	b, err := json.Marshal(rows)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// NewRowJSON returns ordered default values for new rows.
func (c Config) newRowJSON() string {
	row := make([]string, 0, len(c.normalizedColumns()))
	for _, col := range c.normalizedColumns() {
		row = append(row, col.defaultValue())
	}
	b, err := json.Marshal(row)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// EntryAccessor returns a JavaScript expression for this column's entry value.
func (c Column) entryAccessor(index int) string {
	return "entry[" + strconv.Itoa(index) + "]"
}

// NameBinding returns a JavaScript expression for this column's hidden input name.
func (c Column) nameBinding() string {
	return "inputName(index, $el.dataset.columnKey)"
}

func (c Config) entryRows() [][]string {
	cols := c.normalizedColumns()
	if len(c.Entries) == 0 {
		return [][]string{}
	}
	rows := make([][]string, 0, len(c.Entries))
	for _, entry := range c.Entries {
		row := make([]string, 0, len(cols))
		for _, col := range cols {
			row = append(row, entry[col.Key])
		}
		rows = append(rows, row)
	}
	return rows
}
