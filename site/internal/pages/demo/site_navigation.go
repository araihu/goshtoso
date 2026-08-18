package demo

func componentDocsSecondaryNavigationScript() string {
	return `<script>
(function () {
  "use strict";
  function currentFamily(pathname) {
    if (pathname === "/docs/iconpack") return "icon-packs";
    if (pathname === "/modules/charts") return "charts";
    if (pathname === "/modules/app-shells") return "app-shells";
    if (pathname === "/examples" || pathname.indexOf("/examples/") === 0) return "examples";
    return "core";
  }
  function sync() {
    var family = currentFamily(window.location.pathname);
    document.querySelectorAll("#goshtoso-site-secondary-navigation [data-site-secondary-family]").forEach(function (link) {
      if (link.getAttribute("data-site-secondary-family") === family) {
        link.setAttribute("aria-current", "location");
      } else {
        link.removeAttribute("aria-current");
      }
    });
  }
  window.addEventListener("componentdocshell:navigated", sync);
  sync();
})();
</script>`
}

func componentDocsFamily(active string) string {
	switch active {
	case "iconpack":
		return "icon-packs"
	case "module-charts":
		return "charts"
	case "module-app-shells":
		return "app-shells"
	case "examples", "todo", "expense", "chat", "logs", "profile", "ticker", "wizard":
		return "examples"
	default:
		return "core"
	}
}
