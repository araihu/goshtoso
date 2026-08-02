// Package expensepage owns the Expense Tracker runnable example page.
package expensepage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Expense Tracker example's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "examples/expense",
	Title:   "Expense Tracker",
	Active:  "expense",
	Type:    "SoftwareSourceCode",
	Content: ExpenseContent,
}
