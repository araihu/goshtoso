// Authored stateless combobox keyboard navigation and focus behavior.
// Reads and mutates DOM only. No selection state. Safe across HTMX swaps
// and bfcache restore because it queries fresh DOM on each event.
(function () {
  "use strict";

  function options(root) {
    return Array.from(root.querySelectorAll('[role="option"]:not([aria-disabled="true"]):not([data-disabled])')).filter(function (option) {
      return !option.hidden;
    });
  }

  function focusOption(opts, idx) {
    if (opts.length === 0) return -1;
    const i = (idx + opts.length) % opts.length;
    opts[i].focus({ preventScroll: true });
    opts[i].scrollIntoView({ block: "nearest" });
    return i;
  }

  function afterUpdate(callback) {
    // x-show applies transition visibility over animation frames after its
    // reactive update. Focus only once the newly opened panel can receive it.
    const painted = function () { requestAnimationFrame(function () { requestAnimationFrame(callback); }); };
    if (window.Alpine && window.Alpine.nextTick) window.Alpine.nextTick(painted);
    else queueMicrotask(painted);
  }

  function afterOwnedUpdate(root, callback) {
    let movedOutside = false;
    function moved(event) { if (!root.contains(event.target)) movedOutside = true; }
    document.addEventListener("focusin", moved, true);
    afterUpdate(function () {
      document.removeEventListener("focusin", moved, true);
      if (root.isConnected && !movedOutside && (root.contains(document.activeElement) || document.activeElement === document.body)) callback();
    });
  }

  const focusedClears = new WeakSet();
  document.addEventListener("click", function (event) {
    const clear = event.target.closest('[data-combobox-mode="server"] [data-combobox-clear-all]');
    if (clear && document.activeElement === clear) focusedClears.add(clear);
  }, true);

  // Server clear removes its focused button and temporarily disables the trigger.
  // Restore only after the request settles, unless focus moved to another control.
  document.addEventListener("htmx:before:request", function (event) {
    const ctx = event.detail.ctx;
    const source = ctx.sourceElement;
    const root = source.closest('[data-combobox-mode="server"]');
    if (!root || !source.matches("[data-combobox-clear-all]") || !focusedClears.has(source)) return;
    focusedClears.delete(source);
    let movedOutside = false;
    function moved(focusEvent) { if (!root.contains(focusEvent.target)) movedOutside = true; }
    document.addEventListener("focusin", moved, true);
    function settled(done) {
      if (done.detail.ctx !== ctx) return;
      document.removeEventListener("htmx:finally:request", settled, true);
      afterUpdate(function () {
        document.removeEventListener("focusin", moved, true);
        const trigger = root.querySelector('[role="combobox"]');
        if (root.isConnected && !movedOutside && trigger && !trigger.disabled && (root.contains(document.activeElement) || document.activeElement === document.body)) trigger.focus({ preventScroll: true });
      });
    }
    document.addEventListener("htmx:finally:request", settled, true);
  }, true);

  function move(root, key) {
    function focus() {
      if (!root.isConnected || !root.contains(document.activeElement)) return;
      const trigger = root.querySelector('[role="combobox"]');
      if (!trigger || trigger.getAttribute("aria-expanded") !== "true") return;
      const opts = options(root);
      const current = opts.indexOf(document.activeElement);
      if (key === "Home") focusOption(opts, 0);
      else if (key === "End") focusOption(opts, opts.length - 1);
      else if (key === "ArrowUp") focusOption(opts, current === -1 ? opts.length - 1 : current - 1);
      else focusOption(opts, current === -1 ? 0 : current + 1);
    }
    // Alpine opens the panel on the trigger's key handler. Wait for its DOM
    // update before focusing an option that was display:none at keydown time.
    afterOwnedUpdate(root, function () {
      const trigger = root.querySelector('[role="combobox"]');
      if (trigger && trigger.getAttribute("aria-expanded") !== "true") {
        trigger.click();
        afterOwnedUpdate(root, focus);
      } else focus();
    });
  }

  document.addEventListener("keydown", function (e) {
    const root = e.target.closest("[data-combobox]");
    if (!root) return;
    const opts = options(root);
    const current = opts.indexOf(document.activeElement);

    if ((e.key === "Home" || e.key === "End") && e.target.matches('input[type="search"], input[type="text"]')) return;

    switch (e.key) {
      case "ArrowDown":
      case "ArrowUp":
      case "Home":
      case "End":
        e.preventDefault();
        move(root, e.key);
        break;
      case "Enter":
      case " ":
        if (current !== -1) {
          e.preventDefault();
          opts[current].click();
          afterOwnedUpdate(root, function () {
            const trigger = root.querySelector('[role="combobox"]');
            if (trigger && trigger.getAttribute("aria-expanded") === "false") trigger.focus({ preventScroll: true });
          });
        } else if (e.target.matches('[role="combobox"]')) {
          e.preventDefault();
          move(root, "Home");
        }
        break;
      case "Escape":
        afterOwnedUpdate(root, function () {
          const trigger = root.querySelector('[role="combobox"]');
          if (trigger) trigger.focus({ preventScroll: true });
        });
        break;
    }
  });
})();
