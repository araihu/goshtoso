// Authored stateless combobox keyboard navigation and focus behavior.
// Reads and mutates DOM only. No component state. Safe across HTMX swaps
// and bfcache restore because it queries fresh DOM on each event.
(function () {
  "use strict";

  function options(root) {
    return Array.from(root.querySelectorAll('[role="option"]:not([data-disabled])'));
  }

  function focusOption(opts, idx) {
    if (opts.length === 0) return -1;
    const i = (idx + opts.length) % opts.length;
    opts[i].focus();
    return i;
  }

  document.addEventListener("keydown", function (e) {
    const root = e.target.closest("[data-combobox]");
    if (!root) return;
    const opts = options(root);
    const current = opts.indexOf(document.activeElement);

    switch (e.key) {
      case "ArrowDown":
        e.preventDefault();
        focusOption(opts, current === -1 ? 0 : current + 1);
        break;
      case "ArrowUp":
        e.preventDefault();
        focusOption(opts, current === -1 ? opts.length - 1 : current - 1);
        break;
      case "Home":
        e.preventDefault();
        focusOption(opts, 0);
        break;
      case "End":
        e.preventDefault();
        focusOption(opts, opts.length - 1);
        break;
      case "Enter":
        if (current !== -1) {
          e.preventDefault();
          opts[current].click();
        }
        break;
    }
  });
})();
