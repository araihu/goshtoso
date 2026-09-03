// Authored CodeBlock progressive enhancement. The control stays hidden until
// this runtime is present, so server-only output never presents an inert button.
(function () {
  "use strict";

  var buttonSelector = "[data-code-block-copy]";
  var resetDelay = 2000;

  function setState(button, state) {
    var status = button.querySelector("[data-code-block-copy-status]");
    var copyIcon = button.querySelector("[data-code-block-copy-icon]");
    var successIcon = button.querySelector("[data-code-block-success-icon]");
    var initialLabel = button.dataset.codeBlockCopyLabel || button.getAttribute("aria-label");
    button.dataset.codeBlockCopyLabel = initialLabel;

    var successful = state === "success";
    if (status) status.textContent = successful ? "Copied!" : state === "error" ? "Unable to copy" : "Copy";
    if (copyIcon) copyIcon.hidden = successful;
    if (successIcon) successIcon.hidden = !successful;
    button.setAttribute("aria-label", initialLabel);
    button.dataset.codeBlockCopyState = state;
  }

  function enableWithin(node) {
    if (!node || node.nodeType !== Node.ELEMENT_NODE) return;
    var buttons = node.matches(buttonSelector) ? [node] : Array.from(node.querySelectorAll(buttonSelector));
    buttons.forEach(function (button) {
      button.hidden = false;
      if (!button.dataset.codeBlockCopyState) setState(button, "idle");
    });
  }

  async function copy(button) {
    var target = document.getElementById(button.dataset.codeBlockTarget || "");
    if (!target || !navigator.clipboard || typeof navigator.clipboard.writeText !== "function") {
      setState(button, "error");
      return;
    }
    try {
      await navigator.clipboard.writeText(target.textContent);
      setState(button, "success");
    } catch (_error) {
      setState(button, "error");
    }
    window.clearTimeout(button._goshtosoCodeBlockReset);
    button._goshtosoCodeBlockReset = window.setTimeout(function () {
      setState(button, "idle");
    }, resetDelay);
  }

  document.addEventListener("click", function (event) {
    var button = event.target.closest && event.target.closest(buttonSelector);
    if (!button) return;
    event.preventDefault();
    copy(button);
  });

  function initializeDocument() {
    enableWithin(document.documentElement);
  }
  document.readyState === "loading"
    ? document.addEventListener("DOMContentLoaded", initializeDocument, { once: true })
    : initializeDocument();
  document.addEventListener("htmx:afterSwap", function (event) {
    enableWithin(event.target);
  });
})();
