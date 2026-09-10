// ticker-pane.js — Alpine state for the SSE-backed ticker example.
(function () {
  function register() {
    if (!window.Alpine || Alpine.__tickerPaneRegistered) return;
    Alpine.__tickerPaneRegistered = true;
    Alpine.data("tickerPane", function () {
      return {
        connected: false,
        paused: false,
      };
    });
  }

  if (window.Alpine) register();
  else document.addEventListener("alpine:init", register);
})();
