package charts

import "strings"

func tocID(title string) string {
	return strings.ToLower(strings.ReplaceAll(title, " ", "-"))
}
