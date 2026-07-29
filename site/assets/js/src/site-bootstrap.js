// site-bootstrap.js — synchronous storage policy and first-paint theme setup.
(function () {
  if (window.__goshtosoSiteBootstrapInstalled) return;
  window.__goshtosoSiteBootstrapInstalled = true;

  var root = document.documentElement;
  var preferenceCookie = "gt_storage";
  var maxAge = 15552000;
  var demoCookies = ["gt_todo", "gt_profile", "gt_chat"];
  var localKeys = [
    "theme",
    "darkMode",
    "themeOverrides",
    "themeTitleFont",
    "themeBodyFont",
    "themeRadius",
    "themeCssMode",
    "themeCssFilter",
    "cookieConsent",
  ];

  function cookieValue(name) {
    var parts = document.cookie ? document.cookie.split("; ") : [];
    for (var index = 0; index < parts.length; index++) {
      var part = parts[index];
      var equals = part.indexOf("=");
      var key = equals >= 0 ? part.slice(0, equals) : part;
      if (key !== name) continue;
      if (equals < 0) return "";
      try {
        return decodeURIComponent(part.slice(equals + 1));
      } catch (error) {
        return "";
      }
    }
    return "";
  }

  function setCookie(name, value, age) {
    document.cookie =
      name + "=" + encodeURIComponent(value) + "; Path=/; Max-Age=" + age + "; SameSite=Lax";
  }

  function clearLocalStorage() {
    try {
      localKeys.forEach(function (key) {
        localStorage.removeItem(key);
      });
    } catch (error) {}
  }

  function clearIndexedDB() {
    try {
      if (window.indexedDB && indexedDB.deleteDatabase) indexedDB.deleteDatabase("gt_profile");
    } catch (error) {}
  }

  function clearDemoStorage() {
    demoCookies.forEach(function (name) {
      setCookie(name, "", 0);
    });
    clearLocalStorage();
    clearIndexedDB();
  }

  var strict = root.getAttribute("data-demo-storage-policy") === "strict";
  window.goshtosoStorageConsent = window.goshtosoStorageConsent || {
    shouldShow: function () {
      var value = cookieValue(preferenceCookie);
      return value !== "allowed" && value !== "denied";
    },
    allowed: function () {
      var value = cookieValue(preferenceCookie);
      return strict ? value === "allowed" : value !== "denied";
    },
    allow: function () {
      setCookie(preferenceCookie, "allowed", maxAge);
      try {
        localStorage.setItem("cookieConsent", "v2");
      } catch (error) {}
    },
    deny: function () {
      setCookie(preferenceCookie, "denied", maxAge);
      clearDemoStorage();
    },
  };

  if (!window.goshtosoStorageConsent.allowed()) clearDemoStorage();
  if (!root.hasAttribute("data-demo-theme-bootstrap")) return;

  try {
    var canStore = window.goshtosoStorageConsent.allowed();
    var theme = canStore ? localStorage.getItem("theme") || "araihu" : "araihu";
    root.setAttribute("data-theme", theme);
    var savedDarkMode = canStore ? localStorage.getItem("darkMode") : null;
    var darkMode =
      savedDarkMode !== null
        ? savedDarkMode === "true"
        : window.matchMedia("(prefers-color-scheme: dark)").matches;
    if (darkMode) root.classList.add("dark");
    root.classList.add("boot");

    var bootTimer = 0;
    var clearBoot = function () {
      bootTimer = window.setTimeout(function () {
        root.classList.remove("boot");
      }, 600);
    };
    if (document.readyState === "loading") {
      document.addEventListener("DOMContentLoaded", clearBoot, { once: true });
    } else {
      clearBoot();
    }
    window.addEventListener(
      "pagehide",
      function () {
        if (bootTimer) window.clearTimeout(bootTimer);
      },
      { once: true },
    );
  } catch (error) {}
})();
