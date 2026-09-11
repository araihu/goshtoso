// Package expressionutil contains shared component rendering helpers.
package expressionutil

import "encoding/json"

// CountLabels renders complete messages for counts 0 through maxCount as JSON.
// Components use this in escaped data attributes so browser interactions can
// select application-formatted messages without executing translation code.
// label must be non-nil; component callers obtain it from expressions.From, which fills defaults.
func CountLabels(maxCount int, label func(int) string) string {
	labels := make([]string, max(0, maxCount)+1)
	for n := range labels {
		labels[n] = label(n)
	}
	data, _ := json.Marshal(labels)
	return string(data)
}
