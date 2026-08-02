// Package logspage owns the Live Log Feed runnable example page.
package logspage

import "github.com/araihu/goshtoso/site/internal/pages/demo"

// Definition is the Live Log Feed example's neutral registry entry.
var Definition = demo.PageDefinition{
	Key:     "examples/logs",
	Title:   "Live Log Feed",
	Active:  "logs",
	Type:    "SoftwareSourceCode",
	Content: LogsContent,
}
