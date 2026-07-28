// action-group.js — container-aware progressive enhancement for ActionGroup.
(function () {
  "use strict";

  var rootSelector = "[data-goshtoso-action-group]";

  function numericGap(root) {
    var style = window.getComputedStyle(root);
    return parseFloat(style.columnGap || style.gap || "0") || 0;
  }

  function contentWidth(root) {
    var style = window.getComputedStyle(root);
    return Math.max(
      0,
      root.clientWidth -
        (parseFloat(style.paddingLeft) || 0) -
        (parseFloat(style.paddingRight) || 0),
    );
  }

  function overflowCounts(root) {
    var value = root.dataset.actionGroupOverflowCounts || "";
    if (!value) return [];
    return value.split(",").map(function (entry) {
      return parseInt(entry, 10) || 0;
    });
  }

  function setHidden(element, hidden) {
    element.hidden = hidden;
    element.style.display = hidden ? "none" : "";
  }

  function setOverflowItems(overflow, counts, collapsed) {
    var menu = overflow.querySelector('[role="menu"]');
    if (!menu) return;

    var items = Array.from(menu.querySelectorAll('[role="menuitem"]'));
    var offset = 0;
    counts.forEach(function (count, index) {
      for (var itemIndex = offset; itemIndex < offset + count; itemIndex += 1) {
        if (items[itemIndex]) setHidden(items[itemIndex], !collapsed[index]);
      }
      offset += count;
    });

    if (menu.children.length === counts.length) {
      Array.from(menu.children).forEach(function (section, index) {
        setHidden(section, !collapsed[index]);
      });
    }
  }

  function closeHiddenDropdown(wrapper) {
    var trigger = wrapper.querySelector('button[aria-haspopup="true"]');
    if (
      trigger &&
      trigger.getAttribute("aria-expanded") === "true" &&
      typeof trigger.click === "function"
    ) {
      trigger.click();
    }
  }

  function initialize(root) {
    if (root.dataset.actionGroupInitialized === "true") return;
    if (typeof window.ResizeObserver !== "function") return;
    root.dataset.actionGroupInitialized = "true";
    root.style.flexWrap = "nowrap";

    var primary = root.querySelector("[data-action-group-primary]");
    var secondary = Array.from(
      root.querySelectorAll(":scope > [data-action-group-secondary]"),
    );
    var overflow = root.querySelector(":scope > [data-action-group-overflow]");
    var counts = overflowCounts(root);
    if (!primary || !overflow || secondary.length === 0) return;

    var queued = false;
    function measure() {
      queued = false;
      if (!root.isConnected || root.getBoundingClientRect().width <= 0) return;

      var activeInOverflow = overflow.contains(document.activeElement);
      secondary.forEach(function (wrapper) {
        setHidden(wrapper, false);
      });

      var gap = numericGap(root);
      var available = contentWidth(root);
      var primaryWidth = primary.getBoundingClientRect().width;
      var widths = secondary.map(function (wrapper) {
        return wrapper.getBoundingClientRect().width;
      });
      var allWidth =
        primaryWidth +
        widths.reduce(function (sum, width) {
          return sum + width;
        }, 0) +
        gap * secondary.length;

      var collapsed = secondary.map(function () {
        return false;
      });
      if (allWidth <= available) {
        setOverflowItems(overflow, counts, collapsed);
        closeHiddenDropdown(overflow);
        if (activeInOverflow) {
          var restoredAction = secondary[0].querySelector("button, a[href]");
          if (restoredAction) restoredAction.focus();
        }
        setHidden(overflow, true);
        return;
      }

      setHidden(overflow, false);
      var used = primaryWidth + overflow.getBoundingClientRect().width + gap;
      var collapseRest = false;
      secondary.forEach(function (wrapper, index) {
        var nextWidth = widths[index] + gap;
        if (!collapseRest && used + nextWidth <= available) {
          used += nextWidth;
          return;
        }
        collapseRest = true;

        if (wrapper.contains(document.activeElement)) {
          var overflowTrigger = overflow.querySelector(
            'button[aria-haspopup="true"]',
          );
          if (overflowTrigger) overflowTrigger.focus();
        }
        closeHiddenDropdown(wrapper);
        setHidden(wrapper, true);
        collapsed[index] = true;
      });
      setOverflowItems(overflow, counts, collapsed);
    }

    function schedule() {
      if (queued) return;
      queued = true;
      window.requestAnimationFrame(measure);
    }

    var observer = new ResizeObserver(schedule);
    observer.observe(root);
    root._goshtosoActionGroupObserver = observer;
    if (document.fonts && document.fonts.ready) {
      document.fonts.ready.then(schedule);
    }
    schedule();
  }

  function initializeWithin(node) {
    if (!node || node.nodeType !== Node.ELEMENT_NODE) return;
    if (node.matches(rootSelector)) initialize(node);
    node.querySelectorAll(rootSelector).forEach(initialize);
  }

  function initializeDocument() {
    document.querySelectorAll(rootSelector).forEach(initialize);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initializeDocument, {
      once: true,
    });
  } else {
    initializeDocument();
  }

  document.addEventListener("htmx:afterSwap", function (event) {
    initializeWithin(event.target);
  });
})();
