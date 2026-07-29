// ticker-pane.js — Alpine state for the SSE-backed ticker example.
(function () {
  function register() {
    if (!window.Alpine || Alpine.__tickerPaneRegistered) return;
    Alpine.__tickerPaneRegistered = true;
    Alpine.data("tickerPane", function () {
      return {
        connected: false,
        paused: false,
        _sseElement: null,
        _beforeMessage: null,
        _onError: null,
        connect: function (element) {
          if (!element) return;
          this.disconnect();
          var component = this;
          this._sseElement = element;
          this._beforeMessage = function (event) {
            component.connected = true;
            if (component.paused) event.preventDefault();
          };
          this._onError = function () {
            component.connected = false;
          };
          element.addEventListener("htmx:sseBeforeMessage", this._beforeMessage);
          element.addEventListener("htmx:sseError", this._onError);
        },
        disconnect: function () {
          this.connected = false;
          if (!this._sseElement) return;
          if (this._beforeMessage) {
            this._sseElement.removeEventListener("htmx:sseBeforeMessage", this._beforeMessage);
          }
          if (this._onError) {
            this._sseElement.removeEventListener("htmx:sseError", this._onError);
          }
          this._sseElement = null;
          this._beforeMessage = null;
          this._onError = null;
        },
        destroy: function () {
          this.disconnect();
        },
      };
    });
  }

  if (window.Alpine) register();
  else document.addEventListener("alpine:init", register);
})();
