// Table filters, linked-row navigation, and infinite-scroll sentinels.
(function () {
  "use strict";

  if (window.__goshtosoTableRuntimeInit) return;
  window.__goshtosoTableRuntimeInit = true;

  function readInitialFilters(root) {
    var filters = Object.create(null);
    root.querySelectorAll("[data-table-filter-key]").forEach(function (control) {
      var key = control.dataset.tableFilterKey;
      filters[key] = control.dataset.tableFilterDefault || "";
    });
    return filters;
  }

  function activeFilterEntries(filters) {
    return Object.entries(filters).filter(function (entry) {
      return entry[1] !== "" && entry[1] !== "false";
    });
  }

  window.goshtosoTableFilters = function (root) {
    return {
      filtersExpanded: root.dataset.tableFiltersExpanded === "true",
      filters: readInitialFilters(root),
      configRequestListener: null,
      buildFilterURL: function () {
        var url = (root.dataset.tableFilterEndpoint || "") + "?_filter=1";
        var perPage = root.dataset.tableFilterPerPage || "";
        if (perPage) url += "&per_page=" + encodeURIComponent(perPage);
        url += root.dataset.tableFilterExtraQuery || "";
        activeFilterEntries(this.filters).forEach(function (entry) {
          url += "&" + encodeURIComponent(entry[0]) + "=" + encodeURIComponent(entry[1]);
        });
        return url;
      },
      applyFilters: function () {
        if (!window.htmx) return;
        window.htmx.ajax("GET", this.buildFilterURL(), {
          target: root.dataset.tableFilterTarget || "",
          swap: root.dataset.tableFilterSwap || "innerHTML",
        });
      },
      init: function () {
        var state = this;
        this.configRequestListener = function (event) {
          var element = event.detail && event.detail.elt;
          if (!element || !element.closest || element.closest("[data-table-filters]") !== root) {
            return;
          }
          activeFilterEntries(state.filters).forEach(function (entry) {
            event.detail.parameters[entry[0]] = entry[1];
          });
        };
        document.addEventListener("htmx:configRequest", this.configRequestListener);
      },
      destroy: function () {
        if (this.configRequestListener) {
          document.removeEventListener("htmx:configRequest", this.configRequestListener);
        }
        this.configRequestListener = null;
      },
    };
  };

  function linkedRow(event) {
    if (!event.target || !event.target.closest) return null;
    return event.target.closest("[data-table-row-link]");
  }

  document.addEventListener("click", function (event) {
    var row = linkedRow(event);
    if (!row || row.dataset.tableRowLinkMode !== "full" || event.defaultPrevented) return;
    var target = window.goshtosoSafeNavigationTarget(row.dataset.tableRowLink);
    if (target) window.location.href = target;
  });

  document.addEventListener("auxclick", function (event) {
    if (event.button !== 1) return;
    var row = linkedRow(event);
    if (!row) return;
    var target = window.goshtosoSafeNavigationTarget(row.dataset.tableRowLink);
    if (!target) return;
    event.preventDefault();
    var opened = window.open(target, "_blank");
    if (opened) opened.opener = null;
  });

  var sentinelObservers = new WeakMap();

  function requestSentinel(sentinel) {
    var url = sentinel.getAttribute("data-hx-get");
    if (!url || !window.htmx) return false;
    window.htmx.ajax("GET", url, {
      source: sentinel,
      target: sentinel,
      swap: "outerHTML settle:200ms",
    });
    return true;
  }

  function cleanupSentinel(sentinel) {
    var observer = sentinelObservers.get(sentinel);
    if (observer) observer.disconnect();
    sentinelObservers.delete(sentinel);
  }

  function initializeSentinel(sentinel) {
    if (sentinelObservers.has(sentinel)) return;
    if (typeof IntersectionObserver === "undefined") {
      requestSentinel(sentinel);
      return;
    }

    var scrollRoot =
      sentinel.closest(".overflow-y-auto") ||
      sentinel.closest('[style*="overflow-y"]') ||
      null;
    var observer = new IntersectionObserver(
      function (entries) {
        for (var index = 0; index < entries.length; index += 1) {
          if (!entries[index].isIntersecting) continue;
          if (requestSentinel(sentinel)) cleanupSentinel(sentinel);
          return;
        }
      },
      { root: scrollRoot, rootMargin: "400px 0px" },
    );
    sentinelObservers.set(sentinel, observer);
    observer.observe(sentinel);
  }

  function sentinelNodes(root) {
    var nodes = [];
    if (root && root.matches && root.matches("[data-table-scroll-sentinel]")) nodes.push(root);
    if (root && root.querySelectorAll) {
      nodes = nodes.concat(Array.from(root.querySelectorAll("[data-table-scroll-sentinel]")));
    }
    return nodes;
  }

  function initializeSentinels(root) {
    sentinelNodes(root).forEach(initializeSentinel);
  }

  function restartSentinels() {
    sentinelNodes(document).forEach(function (sentinel) {
      cleanupSentinel(sentinel);
      initializeSentinel(sentinel);
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener(
      "DOMContentLoaded",
      function () {
        initializeSentinels(document);
      },
      { once: true },
    );
  } else {
    initializeSentinels(document);
  }

  document.addEventListener("htmx:load", function (event) {
    initializeSentinels((event.detail && event.detail.elt) || event.target);
  });
  document.addEventListener("htmx:beforeCleanupElement", function (event) {
    sentinelNodes((event.detail && event.detail.elt) || event.target).forEach(cleanupSentinel);
  });
  window.addEventListener("goshtoso:dependencies-ready", restartSentinels);
})();
