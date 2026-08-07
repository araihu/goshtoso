(function () {
  "use strict";

  var regions = new WeakMap();

  function setVisible(element, visible) {
    if (!element) return;
    element.hidden = !visible;
  }

  function connect(root) {
    if (!root || regions.has(root)) return;

    var viewport = root.querySelector("[data-goshtoso-scroll-viewport]");
    var start = root.querySelector("[data-goshtoso-scroll-start]");
    var end = root.querySelector("[data-goshtoso-scroll-end]");
    var startIndicator = root.querySelector("[data-goshtoso-scroll-start-indicator]");
    var endIndicator = root.querySelector("[data-goshtoso-scroll-end-indicator]");
    if (!viewport || !start || !end) return;

    var frame = 0;
    var resizeObserver;
    var mutationObserver;

    function observeContent() {
      if (!resizeObserver) return;
      Array.prototype.forEach.call(viewport.children, function (child) {
        resizeObserver.observe(child);
      });
    }

    function update() {
      frame = 0;
      var viewportRect = viewport.getBoundingClientRect();
      var startRect = start.getBoundingClientRect();
      var endRect = end.getBoundingClientRect();
      var tolerance = 1;

      setVisible(startIndicator, startRect.top < viewportRect.top - tolerance);
      setVisible(endIndicator, endRect.bottom > viewportRect.bottom + tolerance);
    }

    function schedule() {
      if (frame) return;
      frame = window.requestAnimationFrame(update);
    }

    viewport.addEventListener("scroll", schedule, { passive: true });

    if (window.ResizeObserver) {
      resizeObserver = new ResizeObserver(schedule);
      resizeObserver.observe(viewport);
      resizeObserver.observe(end);
      observeContent();
    }

    if (window.MutationObserver) {
      mutationObserver = new MutationObserver(function () {
        observeContent();
        schedule();
      });
      mutationObserver.observe(viewport, { childList: true, subtree: true });
    }

    regions.set(root, {
      disconnect: function () {
        viewport.removeEventListener("scroll", schedule);
        if (resizeObserver) resizeObserver.disconnect();
        if (mutationObserver) mutationObserver.disconnect();
        if (frame) window.cancelAnimationFrame(frame);
        regions.delete(root);
      },
    });

    schedule();
  }

  function scan(scope) {
    if (!scope || !scope.querySelectorAll) return;
    if (scope.matches && scope.matches("[data-goshtoso-scroll-region]")) connect(scope);
    scope.querySelectorAll("[data-goshtoso-scroll-region]").forEach(connect);
  }

  function refresh() {
    scan(document);
  }

  function scanHTMXTarget(event) {
    var detail = event.detail || {};
    scan(detail.elt || detail.target || document);
  }

  document.addEventListener("DOMContentLoaded", refresh);
  document.addEventListener("htmx:load", scanHTMXTarget);
  document.addEventListener("htmx:afterSwap", scanHTMXTarget);
  document.addEventListener("htmx:beforeCleanupElement", function (event) {
    var element = event.detail && event.detail.elt;
    if (!element) return;
    var state = regions.get(element);
    if (state) state.disconnect();
    if (element.querySelectorAll) {
      element.querySelectorAll("[data-goshtoso-scroll-region]").forEach(function (region) {
        var nestedState = regions.get(region);
        if (nestedState) nestedState.disconnect();
      });
    }
  });
  window.addEventListener("load", refresh);
})();
