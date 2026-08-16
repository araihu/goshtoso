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
      init: function () {
        this.syncTrigger();
        this.$watch("isOpen", this.syncTrigger.bind(this));
        this.$watch("openedWithKeyboard", this.syncTrigger.bind(this));
      },
      open: function () {
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
        if (!this.isOpen && !this.openedWithKeyboard) return;
        this.close();
        this.$nextTick(function () {
          var trigger = this.triggerElement();
          if (trigger) trigger.focus();
        }.bind(this));
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
        clearTimeout(this.leaveTimeout);
        this.leaveTimeout = null;
      },
    };
  };
})();
