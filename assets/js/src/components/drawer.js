(function () {
  "use strict";

  window.goshtosoDrawer = function (openState, nonModalMinWidth) {
    let mediaQuery;
    let update;
    return {
      [openState]: false,
      wide: false,
      init() {
        if (nonModalMinWidth <= 0) return;
        mediaQuery = window.matchMedia("(min-width: " + nonModalMinWidth + "px)");
        update = () => { this.wide = mediaQuery.matches; };
        mediaQuery.addEventListener("change", update);
        update();
      },
      destroy() {
        mediaQuery?.removeEventListener("change", update);
      },
      canDismissDrawer(event, panel) {
        const menu = "[data-popover-panel], [data-popover-trigger] [aria-expanded=true]";
        return this[openState] && !(panel.contains(event.target) && event.target.closest?.(menu));
      },
    };
  };
})();
