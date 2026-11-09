// Popover Alpine factory. It owns shared open/close and focus lifecycle.
(function () {
  if (window.goshtosoPopover) return;

  var focusableSelector = [
    'button:not([disabled])',
    'a[href]',
    'area[href]',
    'input:not([disabled]):not([type="hidden"])',
    'select:not([disabled])',
    'textarea:not([disabled])',
    'summary',
    '[contenteditable]:not([contenteditable="false"])',
    '[tabindex]:not([tabindex="-1"])',
  ].join(',');

  // Alpine refs may resolve through ancestor scopes. Only use refs owned by
  // this popover, never another component's trigger or panel.
  function ownedRef(root, element) {
    return element && element.closest('[data-popover-root]') === root ? element : null;
  }

  function firstTrigger(root) {
    var marker = root.querySelector('[data-popover-trigger]');
    if (!marker) return null;
    return marker.querySelector(focusableSelector) || marker;
  }

  function panel(root) {
    return root.querySelector('[data-popover-panel]');
  }

  // Keep the panel in its Alpine/DOM scope. Only clipping or transformed
  // ancestors require the native top layer; ordinary panels stay absolute.
  function viewportBounds() {
    var viewport = window.visualViewport;
    return {
      left: (viewport ? viewport.offsetLeft : 0) + 8,
      top: (viewport ? viewport.offsetTop : 0) + 8,
      right: (viewport ? viewport.offsetLeft + viewport.width : document.documentElement.clientWidth) - 8,
      bottom: (viewport ? viewport.offsetTop + viewport.height : window.innerHeight) - 8,
    };
  }

  function clippingAncestors(trigger) {
    var ancestors = [];
    for (var node = trigger.parentElement; node; node = node.parentElement) {
      var style = getComputedStyle(node);
      var clips = node !== document.body && node !== document.documentElement &&
        /(auto|scroll|hidden|clip)/.test(style.overflowX + style.overflowY);
      if (clips ||
          style.transform !== "none" || style.perspective !== "none" ||
          style.filter !== "none" || /(paint|strict|content)/.test(style.contain)) {
        ancestors.push(node);
      }
    }
    return ancestors;
  }

  function setPanelStyle(panel, name, value) {
    if (panel.style.getPropertyValue(name) !== value) panel.style.setProperty(name, value);
  }

  function createPositioner(root, state) {
    var active = false;
    var frame = null;
    var resize = null;
    var closing = null;
    var menu = null;
    var saved = null;
    var portaled = false;
    var limits = null;
    var viewport = window.visualViewport;
    var properties = ["position", "inset", "left", "top", "right", "bottom", "margin", "max-width", "max-height", "min-width", "min-height", "overflow"];

    function restore() {
      if (closing) closing.disconnect();
      closing = null;
      if (!menu || !saved) return;
      if (portaled && menu.matches(":popover-open")) menu.hidePopover();
      if (portaled) menu.removeAttribute("popover");
      properties.forEach(function (name) {
        if (saved[name]) menu.style.setProperty(name, saved[name].value, saved[name].priority);
        else menu.style.removeProperty(name);
      });
      menu.removeAttribute("data-popover-placement");
      saved = null;
      portaled = false;
    }

    function update() {
      frame = null;
      var trigger = state.triggerElement();
      if (!active || state.destroyed || !root.isConnected || !trigger || !menu) return;
      var bounds = viewportBounds();
      var anchor = trigger.getBoundingClientRect();
      var ancestors = clippingAncestors(trigger);
      if (!saved) {
        saved = {};
        properties.forEach(function (name) {
          var value = menu.style.getPropertyValue(name);
          if (value) saved[name] = { value: value, priority: menu.style.getPropertyPriority(name) };
        });
        var style = getComputedStyle(menu);
        limits = {
          minWidth: parseFloat(style.minWidth) || 0,
          minHeight: parseFloat(style.minHeight) || 0,
          maxWidth: parseFloat(style.maxWidth) || Infinity,
          maxHeight: parseFloat(style.maxHeight) || Infinity,
        };
        portaled = ancestors.length > 0 && typeof menu.showPopover === "function";
        if (portaled) {
          menu.setAttribute("popover", "manual");
          setPanelStyle(menu, "position", "fixed");
          menu.showPopover();
        }
        setPanelStyle(menu, "inset", "auto");
        setPanelStyle(menu, "margin", "0px");
      }
      // Without the native top layer, respect the available clipping rectangle.
      if (!portaled) ancestors.forEach(function (ancestor) {
        var style = getComputedStyle(ancestor);
        var rect = ancestor.getBoundingClientRect();
        if (/(auto|scroll|hidden|clip)/.test(style.overflowX)) {
          bounds.left = Math.max(bounds.left, rect.left + ancestor.clientLeft);
          bounds.right = Math.min(bounds.right, rect.left + ancestor.clientLeft + ancestor.clientWidth);
        }
        if (/(auto|scroll|hidden|clip)/.test(style.overflowY)) {
          bounds.top = Math.max(bounds.top, rect.top + ancestor.clientTop);
          bounds.bottom = Math.min(bounds.bottom, rect.top + ancestor.clientTop + ancestor.clientHeight);
        }
      });
      var width = Math.min(limits.maxWidth, Math.max(0, bounds.right - bounds.left));
      var height = Math.min(limits.maxHeight, Math.max(0, bounds.bottom - bounds.top));
      setPanelStyle(menu, "max-width", width + "px");
      setPanelStyle(menu, "min-width", Math.min(limits.minWidth, width) + "px");
      setPanelStyle(menu, "overflow", "auto");
      var preferred = root.dataset.popoverPlacement || "bottom-start";
      var topPreferred = preferred.indexOf("top") === 0;
      var above = Math.max(0, anchor.top - bounds.top - 8);
      var below = Math.max(0, bounds.bottom - anchor.bottom - 8);
      var naturalHeight = menu.scrollHeight + menu.offsetHeight - menu.clientHeight;
      var top = topPreferred ? (naturalHeight <= above || above >= below) : !(naturalHeight <= below || below >= above);
      var available = top ? above : below;
      // Keep content on the selected side, with scrolling when it cannot fit.
      var constrainedHeight = Math.min(height, available || height);
      setPanelStyle(menu, "min-height", Math.min(limits.minHeight, constrainedHeight) + "px");
      setPanelStyle(menu, "max-height", constrainedHeight + "px");
      var panelWidth = menu.offsetWidth;
      var panelHeight = menu.offsetHeight;
      var left = preferred.endsWith("end") ? anchor.right - panelWidth : anchor.left;
      var y = top ? anchor.top - panelHeight - 8 : anchor.bottom + 8;
      left = Math.max(bounds.left, Math.min(left, bounds.right - panelWidth));
      y = Math.max(bounds.top, Math.min(y, bounds.bottom - panelHeight));
      if (!portaled) {
        var parent = menu.offsetParent;
        var rect = parent ? parent.getBoundingClientRect() : { left: 0, top: 0 };
        left -= rect.left + (parent ? parent.clientLeft - parent.scrollLeft : -window.scrollX);
        y -= rect.top + (parent ? parent.clientTop - parent.scrollTop : -window.scrollY);
      }
      setPanelStyle(menu, "left", left + "px");
      setPanelStyle(menu, "top", y + "px");
      menu.dataset.popoverPlacement = (top ? "top" : "bottom") + (preferred.endsWith("end") ? "-end" : "-start");
    }

    function schedule() {
      if (active && frame === null) frame = requestAnimationFrame(update);
    }
    function stop() {
      active = false;
      if (frame !== null) cancelAnimationFrame(frame);
      frame = null;
      if (resize) resize.disconnect();
      resize = null;
      window.removeEventListener("resize", schedule);
      document.removeEventListener("scroll", schedule, true);
      if (viewport) {
        viewport.removeEventListener("resize", schedule);
        viewport.removeEventListener("scroll", schedule);
      }
    }
    return {
      open: function () {
        if (closing) closing.disconnect();
        closing = null;
        if (active) { schedule(); return; }
        menu = state.panelElement();
        var trigger = state.triggerElement();
        if (!menu || !trigger) return;
        active = true;
        window.addEventListener("resize", schedule);
        document.addEventListener("scroll", schedule, true);
        if (viewport) {
          viewport.addEventListener("resize", schedule);
          viewport.addEventListener("scroll", schedule);
        }
        if (typeof ResizeObserver === "function") {
          resize = new ResizeObserver(schedule);
          resize.observe(menu);
          resize.observe(trigger);
        }
        update();
      },
      close: function () {
        stop();
        if (!menu || !saved || closing) return;
        if (getComputedStyle(menu).display === "none") { restore(); return; }
        var hidden = function () {
          if (!active && getComputedStyle(menu).display === "none") restore();
        };
        if (typeof MutationObserver === "function") {
          closing = new MutationObserver(hidden);
          closing.observe(menu, { attributes: true, attributeFilter: ["style", "class"] });
        } else {
          var pending;
          closing = { disconnect: function () { cancelAnimationFrame(pending); } };
          var poll = function () {
            hidden();
            if (closing && !active) pending = requestAnimationFrame(poll);
          };
          pending = requestAnimationFrame(poll);
        }
      },
      destroy: function () { stop(); restore(); },
    };
  }

  window.goshtosoPopover = function (root) {
    var positioner;
    return {
      isOpen: false,
      openedWithKeyboard: false,
      leaveTimeout: null,
      focusRestoreObserver: null,
      focusRestoreFallbackCleanup: null,
      focusRestoreGeneration: 0,
      destroyed: false,
      get triggerClass() {
        return this.isOpen || this.openedWithKeyboard
          ? "text-on-surface-strong dark:text-on-surface-dark-strong"
          : "text-on-surface dark:text-on-surface-dark";
      },
      handleEscape: function (event) {
        if (event.defaultPrevented || !(this.isOpen || this.openedWithKeyboard)) return;
        var owner = event.target && event.target.closest && event.target.closest('[data-popover-root]');
        while (owner) {
          var trigger = firstTrigger(owner);
          if (trigger && trigger.getAttribute('aria-expanded') === 'true') {
            if (owner !== root) return;
            break;
          }
          owner = owner.parentElement && owner.parentElement.closest('[data-popover-root]');
        }
        event.preventDefault();
        this.closeAndFocus();
      },
      triggerElement: function () {
        return ownedRef(root, this.$refs.trigger) || firstTrigger(root);
      },
      panelElement: function () {
        return ownedRef(root, this.$refs.panel) || panel(root);
      },
      syncTrigger: function () {
        var trigger = this.triggerElement();
        if (!trigger) return;
        var content = this.panelElement();
        var role = content && content.getAttribute("role");
        trigger.setAttribute("aria-haspopup", role || "true");
        if (content && content.id) {
          trigger.setAttribute("aria-controls", content.id);
        }
        trigger.setAttribute("aria-expanded", String(this.isOpen || this.openedWithKeyboard));
      },
      cancelFocusRestore: function () {
        this.focusRestoreGeneration += 1;
        if (this.focusRestoreObserver) this.focusRestoreObserver.disconnect();
        this.focusRestoreObserver = null;
        if (this.focusRestoreFallbackCleanup) this.focusRestoreFallbackCleanup();
        this.focusRestoreFallbackCleanup = null;
      },
      focusTriggerIfOwned: function (trigger, menu, closingFocus) {
        if (this.destroyed || !trigger) return false;
        var active = document.activeElement;
        var closingFocusOwned =
          closingFocus === trigger ||
          closingFocus === document.body ||
          closingFocus === document.documentElement ||
          (menu && menu.contains(closingFocus));
        if (
          !closingFocusOwned ||
          !(active === closingFocus ||
            active === trigger ||
            active === document.body ||
            active === document.documentElement ||
            (menu && menu.contains(active)))
        ) {
          return false;
        }
        trigger.focus();
        return true;
      },
      deferFocusRestoreUntilMenuHidden: function (trigger, menu, closingFocus, generation) {
        var state = this;
        var animationFrame = null;
        var timeout = null;
        var settled = false;
        var cleanup = function () {
          if (settled) return;
          settled = true;
          if (animationFrame !== null && typeof window.cancelAnimationFrame === "function") {
            window.cancelAnimationFrame(animationFrame);
            animationFrame = null;
          }
          if (timeout !== null) {
            window.clearTimeout(timeout);
            timeout = null;
          }
          menu.removeEventListener("transitionend", onTransitionEnd);
          menu.removeEventListener("animationend", onTransitionEnd);
          if (state.focusRestoreFallbackCleanup === cleanup) {
            state.focusRestoreFallbackCleanup = null;
          }
        };
        var schedule = function () {
          if (settled) return;
          if (typeof window.requestAnimationFrame === "function") {
            animationFrame = window.requestAnimationFrame(attempt);
          } else {
            timeout = window.setTimeout(attempt, 16);
          }
        };
        var attempt = function () {
          if (state.destroyed || state.focusRestoreGeneration !== generation) {
            cleanup();
            return;
          }
          if (state.restoreTriggerAfterMenuHidden(trigger, menu, closingFocus)) {
            cleanup();
            return;
          }
          schedule();
        };
        var onTransitionEnd = function (event) {
          if (event.target === menu) attempt();
        };
        this.focusRestoreFallbackCleanup = cleanup;
        menu.addEventListener("transitionend", onTransitionEnd);
        menu.addEventListener("animationend", onTransitionEnd);
        attempt();
      },
      restoreTriggerAfterMenuHidden: function (trigger, menu, closingFocus) {
        if (this.destroyed) return true;
        if (this.isOpen || this.openedWithKeyboard) {
          this.cancelFocusRestore();
          return true;
        }
        if (menu && window.getComputedStyle(menu).display !== "none") return false;

        this.cancelFocusRestore();
        this.focusTriggerIfOwned(trigger, menu, closingFocus);
        return true;
      },
      init: function () {
        positioner = createPositioner(root, this);
        var sync = function () {
          this.syncTrigger();
          this.$nextTick(function () {
            if (this.destroyed) return;
            if (this.isOpen || this.openedWithKeyboard) positioner.open();
            else positioner.close();
          }.bind(this));
        }.bind(this);
        sync();
        this.$watch("isOpen", sync);
        this.$watch("openedWithKeyboard", sync);
      },
      open: function () {
        this.cancelFocusRestore();
        clearTimeout(this.leaveTimeout);
        this.leaveTimeout = null;
        this.isOpen = true;
        this.syncTrigger();
      },
      close: function () {
        this.isOpen = false;
        this.openedWithKeyboard = false;
        this.syncTrigger();
      },
      toggle: function () {
        if (this.isOpen || this.openedWithKeyboard) {
          this.close();
          return;
        }
        this.open();
      },
      openFromKeyboard: function () {
        this.openedWithKeyboard = true;
        this.open();
        this.focusFirstItem();
      },
      closeAndFocus: function () {
        if (this.destroyed || (!this.isOpen && !this.openedWithKeyboard)) return;
        this.cancelFocusRestore();
        var trigger = this.triggerElement();
        var menu = this.panelElement();
        var closingFocus = document.activeElement;
        var state = this;
        var generation = this.focusRestoreGeneration;
        this.close();
        this.$nextTick(function () {
          if (state.destroyed || state.focusRestoreGeneration !== generation) return;
          if (state.restoreTriggerAfterMenuHidden(trigger, menu, closingFocus)) return;
          if (typeof window.MutationObserver !== "function") {
            if (menu) {
              state.deferFocusRestoreUntilMenuHidden(trigger, menu, closingFocus, generation);
              return;
            }
            state.focusTriggerIfOwned(trigger, menu, closingFocus);
            state.cancelFocusRestore();
            return;
          }
          if (state.destroyed || state.focusRestoreGeneration !== generation) return;
          state.focusRestoreObserver = new MutationObserver(function () {
            if (state.destroyed || state.focusRestoreGeneration !== generation) return;
            state.restoreTriggerAfterMenuHidden(trigger, menu, closingFocus);
          });
          state.focusRestoreObserver.observe(menu, {
            attributes: true,
            attributeFilter: ["class", "hidden", "style"],
          });
          state.restoreTriggerAfterMenuHidden(trigger, menu, closingFocus);
        });
      },
      scheduleClose: function () {
        var state = this;
        clearTimeout(this.leaveTimeout);
        this.leaveTimeout = setTimeout(function () {
          state.close();
        }, 250);
      },
      clearScheduledClose: function () {
        clearTimeout(this.leaveTimeout);
        this.leaveTimeout = null;
      },
      menuItems: function () {
        var content = this.panelElement();
        if (!content) return [];
        return Array.from(
          content.querySelectorAll('[role="menuitem"]:not([disabled]):not([hidden])'),
        ).filter(function (item) {
          return item.getClientRects().length > 0 && getComputedStyle(item).visibility !== "hidden";
        });
      },
      focusFirstItem: function () {
        this.cancelFocusRestore();
        this.$nextTick(function () {
          var first = this.menuItems()[0];
          if (first) first.focus();
        }.bind(this));
      },
      focusAdjacentItem: function (step) {
        var items = this.menuItems();
        if (items.length === 0) return;
        var current = items.indexOf(document.activeElement);
        var next = current < 0 ? (step > 0 ? 0 : items.length - 1) : (current + step + items.length) % items.length;
        items[next].focus();
      },
      destroy: function () {
        this.destroyed = true;
        if (positioner) positioner.destroy();
        this.cancelFocusRestore();
        clearTimeout(this.leaveTimeout);
        this.leaveTimeout = null;
      },
    };
  };
})();
