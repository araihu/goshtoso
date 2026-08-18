package demo

import "strings"

func componentDocsFamily(active string) string {
	switch active {
	case "agents":
		return "agents"
	case "icon", "icon-catalog", "iconpack":
		return "icon-packs"
	case "module-charts":
		return "charts"
	case "module-app-shells":
		return "app-shells"
	case "examples", "todo", "expense", "chat", "logs", "profile", "ticker", "wizard":
		return "examples"
	default:
		if strings.HasPrefix(active, "app-shells-") {
			return "app-shells"
		}
		if strings.HasPrefix(active, "charts-") {
			return "charts"
		}
		return "core"
	}
}
