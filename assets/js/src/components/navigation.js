// Shared validation for component-owned client navigation sinks.
(function () {
  "use strict";

  if (window.goshtosoSafeNavigationTarget) return;

  window.goshtosoSafeNavigationTarget = function (value) {
    var target = String(value || "").trim();
    if (!target) return "";
    try {
      var parsed = new URL(target, window.location.href);
      var protocol = parsed.protocol.toLowerCase();
      if (
        protocol === "http:" ||
        protocol === "https:" ||
        protocol === "mailto:" ||
        protocol === "tel:"
      ) {
        return target;
      }
    } catch (error) {
      return "";
    }
    return "";
  };
})();
