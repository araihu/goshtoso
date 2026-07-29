// tab-view.js — Alpine state for the legacy demo comparison tabs.
(function () {
  function register() {
    if (!window.Alpine || Alpine.__demoTabViewRegistered) return;
    Alpine.__demoTabViewRegistered = true;
    Alpine.data("demoTabView", function () {
      return {
        activeTab: "gottha",
        size: "md",
        disabled: false,
        copiedCode: false,
        _copyTimer: 0,
        _destroyed: false,
        copyToClipboard: function (text) {
          var component = this;
          if (!navigator.clipboard || !navigator.clipboard.writeText) return;
          navigator.clipboard
            .writeText(text)
            .then(function () {
              if (component._destroyed) return;
              component.copiedCode = true;
              if (component._copyTimer) window.clearTimeout(component._copyTimer);
              component._copyTimer = window.setTimeout(function () {
                component.copiedCode = false;
                component._copyTimer = 0;
              }, 2000);
            })
            .catch(function () {});
        },
        destroy: function () {
          this._destroyed = true;
          if (this._copyTimer) window.clearTimeout(this._copyTimer);
          this._copyTimer = 0;
        },
      };
    });
  }

  if (window.Alpine) register();
  else document.addEventListener("alpine:init", register, { once: true });
})();
