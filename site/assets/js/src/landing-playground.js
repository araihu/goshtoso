// landing-playground.js — keep the isolated homepage preview flush with its content.
(function () {
  var messageType = "goshtoso:landing-playground-height";
  var frameID = "theme-playground-frame";
  var root = document.documentElement;

  if (!root.hasAttribute("data-landing-playground")) {
    window.addEventListener("message", function (event) {
      if (event.origin !== window.location.origin || !event.data || event.data.type !== messageType) {
        return;
      }
      var frame = document.getElementById(frameID);
      if (!frame || event.source !== frame.contentWindow) return;
      var height = Math.ceil(Number(event.data.height));
      if (!Number.isFinite(height) || height < 1) return;
      frame.style.height = height + "px";
    });
    return;
  }

  try {
    root.classList.toggle("dark", window.parent.document.documentElement.classList.contains("dark"));
  } catch (error) {}

  function reportHeight() {
    var body = document.body;
    var bodyHeight = body
      ? Math.max(body.scrollHeight, body.getBoundingClientRect().height)
      : root.scrollHeight;
    // A small allowance absorbs fractional CSS-pixel rounding at zoom levels
    // where scrollHeight and the painted border land on opposite integers.
    var height = Math.ceil(bodyHeight) + 2;
    window.parent.postMessage({ type: messageType, height: height }, window.location.origin);
  }

  function observeHeight() {
    reportHeight();
    if (typeof ResizeObserver === "function") {
      var observer = new ResizeObserver(reportHeight);
      observer.observe(root);
      if (document.body) observer.observe(document.body);
      window.addEventListener("pagehide", function () {
        observer.disconnect();
      }, { once: true });
    }
    window.addEventListener("load", reportHeight, { once: true });
    document.addEventListener("htmx:afterSettle", reportHeight);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", observeHeight, { once: true });
  } else {
    observeHeight();
  }
})();
