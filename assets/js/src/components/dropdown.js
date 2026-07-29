// Dropdown Alpine factory. It reads only static root data and owns timers.
(function () {
  if (window.goshtosoDropdown) return;

  window.goshtosoDropdown = function (root) {
    return {
      isOpen: false,
      openedWithKeyboard: false,
      leaveTimeout: null,
      get triggerClass() {
        return this.isOpen || this.openedWithKeyboard
          ? "text-on-surface-strong dark:text-on-surface-dark-strong"
          : "text-on-surface dark:text-on-surface-dark";
      },
      closeAndFocus: function () {
        if (!this.isOpen && !this.openedWithKeyboard) return;
        this.isOpen = false;
        this.openedWithKeyboard = false;
        this.$nextTick(function () {
          if (this.$refs.trigger) this.$refs.trigger.focus();
        }.bind(this));
      },
      scheduleClose: function () {
        var state = this;
        clearTimeout(this.leaveTimeout);
        this.leaveTimeout = setTimeout(function () {
          state.isOpen = false;
        }, 250);
      },
      menuItems: function () {
        if (!this.$refs.menu) return [];
        return Array.from(
          this.$refs.menu.querySelectorAll('[role="menuitem"]:not([disabled]):not([hidden])'),
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
      },
    };
  };
})();
