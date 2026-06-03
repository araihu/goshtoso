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
	cols := c.NormalizedColumns()
	return "{ name: '" + jsEscapeSingle(c.Name) + "', columns: " + columnsLiteral(cols) + ", entries: " + entriesLiteral(c.Entries, cols) + " }"
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
