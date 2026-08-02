// Package todopage owns the Todo List runnable example page.
package todopage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Todo List example's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "examples/todo",
	Title:   "Todo List",
	Active:  "todo",
	Type:    "SoftwareSourceCode",
	Content: TodoContent,
}
