// Site-level navigation synchronization for the component documentation shell.
(function () {
  "use strict";

  function currentFamily(pathname) {
    if (pathname === "/docs/agents" || pathname.indexOf("/docs/agents/") === 0) return "agents";
    if (pathname === "/components/icon" || pathname === "/docs/icon-catalog" || pathname === "/docs/iconpack") return "icon-packs";
    if (pathname === "/modules/charts" || pathname.indexOf("/modules/charts/") === 0) return "charts";
    if (pathname === "/modules/app-shells" || pathname.indexOf("/modules/app-shells/") === 0) return "app-shells";
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
