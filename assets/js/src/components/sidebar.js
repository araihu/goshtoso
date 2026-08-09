// Sidebar overlay Alpine provider. Each rendered root owns independent state.
(function () {
  "use strict";

  window.goshtosoSidebarOverlay = window.goshtosoSidebarOverlay || function () {
    return {
      open: false,
      closeAndFocus: function () {
        if (!this.open) return;
        this.open = false;
        this.$nextTick(function () {
          if (this.$refs.trigger) this.$refs.trigger.focus();
        }.bind(this));
      },
    };
  };

  function registerSidebarOverlay() {
    if (!window.Alpine || window.Alpine.__goshtosoSidebarOverlayRegistered) return;
    window.Alpine.__goshtosoSidebarOverlayRegistered = true;
    window.Alpine.data("goshtosoSidebarOverlay", window.goshtosoSidebarOverlay);
  }

  if (window.Alpine) registerSidebarOverlay();
  else document.addEventListener("alpine:init", registerSidebarOverlay, { once: true });
})();
