// Table filter presentation/payload normalization and linked-row navigation.
(function () {
  "use strict";

  if (window.__goshtosoTableRuntimeInit) return;
  window.__goshtosoTableRuntimeInit = true;

  function readInitialFilters(root) {
    var filters = Object.create(null);
    root.querySelectorAll("[data-table-filter-key]").forEach(function (control) {
      var key = control.dataset.tableFilterKey;
      var value = control.dataset.tableFilterDefault || "";
      filters[key] = control.type === "checkbox" ? value === "true" : value;
    });
    return filters;
  }

  window.goshtosoTableFilters = function (root) {
    return {
      filtersExpanded: root.dataset.tableFiltersExpanded === "true",
      filters: readInitialFilters(root),
      applyFilters: function () {
        root.dispatchEvent(new Event("table-filter", { bubbles: true }));
      },
      configureRequest: function (event) {
        var ctx = event.detail && event.detail.ctx;
        if (!ctx || ctx.request.method !== "GET" || ctx.sourceElement.closest("[data-table-filters]") !== root) return;
        // Option loaders have their own endpoint and must not become row filters.
        if (ctx.sourceElement.matches("select[data-table-filter-key]")) return;
        var url = new URL(ctx.request.action, window.location.href);
        // Native inclusion owns values. Clear stale values in server-generated
        // sort/page URLs, retaining the existing omission contract for empty/false.
        root.querySelectorAll("[data-table-filter-key]").forEach(function (input) {
          var key = input.name;
          url.searchParams.delete(key);
          var value = ctx.request.body.get(key);
          if (value === "" || value === "false") ctx.request.body.delete(key);
        });
        if (ctx.sourceElement === root) {
          var head = root.querySelector("thead[data-table-sort-by]");
          ["order_by", "order_dir"].forEach(function (key) { url.searchParams.delete(key); });
          if (head && head.dataset.tableSortBy && head.dataset.tableSortDir) {
            ctx.request.body.set("order_by", head.dataset.tableSortBy);
            ctx.request.body.set("order_dir", head.dataset.tableSortDir);
          }
        }
        ctx.request.action = url.origin === location.origin ? url.pathname + url.search + url.hash : url.toString();
      },
    };
  };

  function linkedRow(event) {
    if (!event.target || !event.target.closest) return null;
    return event.target.closest("[data-table-row-link]");
  }

  // Shared by native navigation and HTMX's row click filter. Nested controls
  // retain their own default actions and event handlers.
  window.goshtosoTableRowLinkEvent = function (event) {
    var row = linkedRow(event);
    if (!row || event.defaultPrevented) return false;
    var control = event.target.closest("a,button,input,label,select,textarea,summary,[contenteditable]:not([contenteditable='false']),[role='button'],[role='link'],[role='checkbox'],[role='switch'],[role='combobox']");
    return !control || control === row || !row.contains(control);
  };

  document.addEventListener("click", function (event) {
    var row = linkedRow(event);
    if (!window.goshtosoTableRowLinkEvent(event)) return;

    if (!row || row.dataset.tableRowLinkMode !== "full") return;
    var target = window.goshtosoSafeNavigationTarget(row.dataset.tableRowLink);
    if (target) window.location.href = target;
  });

  document.addEventListener("auxclick", function (event) {
    if (event.button !== 1) return;
    var row = linkedRow(event);
    if (!row || !window.goshtosoTableRowLinkEvent(event)) return;
    var target = window.goshtosoSafeNavigationTarget(row.dataset.tableRowLink);
    if (!target) return;
    event.preventDefault();
    var opened = window.open(target, "_blank");
    if (opened) opened.opener = null;
  });

})();
