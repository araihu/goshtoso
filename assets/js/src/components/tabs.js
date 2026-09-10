// Tabs Alpine factory. Hash synchronization is opt-in root data.
(function () {
  if (window.goshtosoTabs) return;

  window.goshtosoTabs = function (root) {
    var config = window.goshtosoParseData(root.dataset.tabsConfig, { default: "", ids: [], syncHash: false });
    var ids = Array.isArray(config.ids) ? config.ids : [];
    var syncHash = Boolean(config.syncHash);
    return {
      selectedTab: config.default || "",
      init: function () {
        if (!syncHash) return;
        var hash = window.location.hash.slice(1);
        if (hash && ids.includes(hash)) this.selectedTab = hash;
        this.$watch("selectedTab", function (tab) {
          history.replaceState(null, "", "#" + tab);
        });
      },
      loadPanel: function (panel) {
        if (panel.dataset.loaded || panel.dataset.loading) return;
        panel.dataset.loading = "true";
        this.$nextTick(function () {
          if (!panel.isConnected) return;
          // Alpine may select an initial/hash panel before htmx's first scan.
          htmx.process(panel);
          panel.dispatchEvent(new CustomEvent("goshtoso-tab-load", { bubbles: true }));
        });
      },
      finishPanel: function (panel, event) {
        var ctx = event.detail && event.detail.ctx;
        if (!ctx || ctx.sourceElement !== panel) return;
        delete panel.dataset.loading;
        if (ctx.response && ctx.response.status >= 200 && ctx.response.status < 400) panel.dataset.loaded = "true";
      },
      moveFocus: function (event, direction) {
        var tabs = Array.from(event.currentTarget.querySelectorAll('[role="tab"]'));
        var index = tabs.indexOf(document.activeElement);
        if (index < 0) index = tabs.indexOf(event.target);
        if (index < 0 || tabs.length === 0) return;
        tabs[(index + direction + tabs.length) % tabs.length].focus();
      },
    };
  };
})();
