// demo-layout.js — Alpine providers plus HTMX navigation and TOC lifecycle.
(function () {
  function storageAllowed() {
    return !window.goshtosoStorageConsent || window.goshtosoStorageConsent.allowed();
  }

  function readTheme() {
    if (!storageAllowed()) return "araihu";
    try {
      return localStorage.getItem("theme") || "araihu";
    } catch (error) {
      return "araihu";
    }
  }

  function persistTheme(value) {
    try {
      if (storageAllowed()) localStorage.setItem("theme", value);
      else localStorage.removeItem("theme");
    } catch (error) {}
    document.documentElement.setAttribute("data-theme", value);
  }

  function registerAlpineProviders() {
    if (!window.Alpine || Alpine.__demoLayoutProvidersRegistered) return;
    Alpine.__demoLayoutProvidersRegistered = true;

    Alpine.data("demoStorageConsent", function () {
      return {
        show:
          !window.goshtosoStorageConsent || window.goshtosoStorageConsent.shouldShow(),
        allow: function () {
          if (window.goshtosoStorageConsent) window.goshtosoStorageConsent.allow();
          this.show = false;
        },
        deny: function () {
          if (window.goshtosoStorageConsent) window.goshtosoStorageConsent.deny();
          this.show = false;
        },
      };
    });

    Alpine.data("demoLayout", function () {
      return {
        theme: readTheme(),
        sidebarOpen: false,
        showThemeDropdown: false,
        _stopThemeWatch: null,
        init: function () {
          this._stopThemeWatch = this.$watch("theme", persistTheme);
        },
        destroy: function () {
          if (typeof this._stopThemeWatch === "function") this._stopThemeWatch();
          this._stopThemeWatch = null;
        },
        setTheme: function (name) {
          this.theme = name;
          document.documentElement.setAttribute("data-theme", name);
        },
      };
    });

    var nav = Alpine.store("nav");
    if (nav) nav.path = window.location.pathname;
    else Alpine.store("nav", { path: window.location.pathname });
  }

  if (window.Alpine) registerAlpineProviders();
  else document.addEventListener("alpine:init", registerAlpineProviders, { once: true });

  if (window.__goshtosoDemoLayoutRuntimeInstalled) return;
  window.__goshtosoDemoLayoutRuntimeInstalled = true;

  var sidebarScrollTop = 0;
  var tocObserver = null;

  function updateNavPath() {
    if (!window.Alpine) return;
    var nav = Alpine.store("nav");
    if (nav) nav.path = window.location.pathname;
  }

  function rememberSidebarScroll() {
    var sidebar = document.querySelector(".sidebar-scroll");
    if (sidebar) sidebarScrollTop = sidebar.scrollTop;
  }

  function handleAfterSwap(event) {
    var target = event && event.detail && event.detail.target;
    if (target && window.Alpine && Alpine.initTree) Alpine.initTree(target);
    if (!target || target.id !== "main-content") return;

    var sidebarContent = document.getElementById("sidebar-nav-content");
    if (sidebarContent && window.Alpine && Alpine.initTree) Alpine.initTree(sidebarContent);
    if (sidebarContent && window.htmx && htmx.process) htmx.process(sidebarContent);
    var sidebar = document.querySelector(".sidebar-scroll");
    if (sidebar) sidebar.scrollTop = sidebarScrollTop;
    var pageScroll = document.getElementById("page-scroll");
    if (pageScroll) pageScroll.scrollTo({ top: 0 });
    buildTOC();
  }

  function disconnectTOC() {
    if (!tocObserver) return;
    tocObserver.disconnect();
    tocObserver = null;
  }

  function clearActive(nav) {
    nav.querySelectorAll("[data-toc-link]").forEach(function (link) {
      link.classList.remove(
        "border-primary",
        "dark:border-primary-dark",
        "text-on-surface-strong",
        "dark:text-on-surface-dark-strong",
        "font-medium",
      );
      link.classList.add("border-transparent");
    });
  }

  function setActive(nav, id) {
    var link = Array.prototype.find.call(
      nav.querySelectorAll("[data-toc-link]"),
      function (candidate) {
        return candidate.getAttribute("data-toc-link") === id;
      },
    );
    if (!link) return;
    clearActive(nav);
    link.classList.remove("border-transparent");
    link.classList.add(
      "border-primary",
      "dark:border-primary-dark",
      "text-on-surface-strong",
      "dark:text-on-surface-dark-strong",
      "font-medium",
    );
  }

  function buildTOC() {
    var rail = document.getElementById("toc-rail");
    var nav = document.getElementById("toc-list");
    var pageScroll = document.getElementById("page-scroll");
    var content = document.getElementById("main-content");
    if (!rail || !nav || !pageScroll || !content) return;

    disconnectTOC();
    var headings = Array.prototype.slice.call(content.querySelectorAll("[data-toc-heading]"));
    nav.replaceChildren();
    if (headings.length < 2) {
      rail.style.display = "none";
      return;
    }
    rail.style.display = "";

    headings.forEach(function (heading) {
      if (!heading.id) return;
      var link = document.createElement("a");
      link.href = "#" + heading.id;
      link.textContent = (heading.textContent || "").trim();
      link.setAttribute("data-toc-link", heading.id);
      link.className =
        "block border-l border-transparent py-1.5 pl-4 -ml-px text-sm text-on-surface-muted transition-colors hover:text-on-surface-strong dark:text-on-surface-dark-muted dark:hover:text-on-surface-dark-strong";
      link.addEventListener("click", function (event) {
        event.preventDefault();
        heading.scrollIntoView({ behavior: "smooth", block: "start" });
        history.replaceState(null, "", "#" + heading.id);
        setActive(nav, heading.id);
      });
      nav.appendChild(link);
    });

    tocObserver = new IntersectionObserver(
      function (entries) {
        entries.forEach(function (entry) {
          if (entry.isIntersecting) setActive(nav, entry.target.id);
        });
      },
      { root: pageScroll, rootMargin: "0px 0px -75% 0px", threshold: 0 },
    );
    headings.forEach(function (heading) {
      if (heading.id) tocObserver.observe(heading);
    });

    var hashID = "";
    try {
      hashID = decodeURIComponent((window.location.hash || "").replace(/^#/, ""));
    } catch (error) {}
    var active = headings.find(function (heading) {
      return heading.id === hashID;
    });
    setActive(nav, (active || headings[0]).id);
  }

  function teardownRuntime(event) {
    if (event && event.persisted) return;
    disconnectTOC();
    document.removeEventListener("htmx:pushedIntoHistory", updateNavPath);
    document.removeEventListener("htmx:beforeSwap", rememberSidebarScroll);
    document.removeEventListener("htmx:afterSwap", handleAfterSwap);
  }

  window.buildTOC = buildTOC;
  document.addEventListener("htmx:pushedIntoHistory", updateNavPath);
  document.addEventListener("htmx:beforeSwap", rememberSidebarScroll);
  document.addEventListener("htmx:afterSwap", handleAfterSwap);
  window.addEventListener("pagehide", teardownRuntime, { once: true });
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", buildTOC, { once: true });
  } else {
    buildTOC();
  }
})();
