package demo

import "strings"

func componentDocsFamily(active string) string {
	switch active {
	case "internationalization", "internationalization-examples", "internationalization-files":
		return "internationalization"
	case "agents", "icon", "icon-catalog", "iconpack":
		return "core"
	case "module-charts":
		return "charts"
	case "module-app-shells":
		return "app-shells"
	case "examples", "deployments", "todo", "expense", "chat", "logs", "profile", "ticker", "wizard":
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
