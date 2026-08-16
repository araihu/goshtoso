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

  function firstTrigger(root) {
    var marker = root.querySelector('[data-popover-trigger]');
    if (!marker) return null;
    return marker.querySelector(focusableSelector) || marker;
  }

  function panel(root) {
    return root.querySelector('[data-popover-panel]');
  }

  window.goshtosoPopover = function (root) {
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
      triggerElement: function () {
        return this.$refs.trigger || firstTrigger(root);
      },
      panelElement: function () {
        return this.$refs.panel || panel(root);
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
        this.syncTrigger();
        this.$watch("isOpen", this.syncTrigger.bind(this));
        this.$watch("openedWithKeyboard", this.syncTrigger.bind(this));
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
        this.cancelFocusRestore();
        clearTimeout(this.leaveTimeout);
        this.leaveTimeout = null;
      },
    };
  };
})();
