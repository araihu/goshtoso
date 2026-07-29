// Authored ordered CDN loader with version-matched local fallback.
(function () {
  "use strict";

  var loader = document.currentScript;
  if (!loader) return;

  var existing = window.goshtosoDependencies;
  if (existing && existing.ready) return;

  var config;
  try {
    config = JSON.parse(loader.dataset.goshtosoDependencies || "{}");
  } catch (error) {
    window.goshtosoDependencies = { ready: Promise.reject(error) };
    return;
  }

  var dependencies = Array.isArray(config.dependencies) ? config.dependencies : [];
  var state = { loaded: {}, sources: {} };
  window.goshtosoDependencies = state;

  function dispatch(name, detail) {
    window.dispatchEvent(new CustomEvent(name, { detail: detail }));
  }

  function whenWindowLoaded() {
    if (document.readyState === "complete") return Promise.resolve();
    return new Promise(function (resolve) {
      window.addEventListener("load", resolve, { once: true });
    });
  }

  function appendScript(entry, url, source) {
    return new Promise(function (resolve, reject) {
      var script = document.createElement("script");
      script.src = url;
      script.async = false;
      script.dataset.goshtosoDependency = entry.name;
      script.dataset.goshtosoSource = source;
      if (entry.integrity) {
        script.integrity = entry.integrity;
        script.crossOrigin = "anonymous";
      }
      if (loader.nonce) script.nonce = loader.nonce;
      script.addEventListener("load", function () {
        state.loaded[entry.name] = true;
        state.sources[entry.name] = source;
        resolve();
      }, { once: true });
      script.addEventListener("error", function () {
        script.remove();
        reject(new Error("failed to load " + entry.name + " from " + url));
      }, { once: true });
      document.head.appendChild(script);
    });
  }

  async function load(entry) {
    if (entry.wait_for_window_loaded) await whenWindowLoaded();
    try {
      await appendScript(entry, entry.primary_url, "primary");
    } catch (primaryError) {
      if (!entry.fallback_url || entry.fallback_url === entry.primary_url) {
        throw primaryError;
      }
      dispatch("goshtoso:dependency-fallback", {
        dependency: entry.name,
        primaryURL: entry.primary_url,
        fallbackURL: entry.fallback_url,
      });
      await appendScript(entry, entry.fallback_url, "fallback");
    }
  }

  state.ready = (async function () {
    try {
      for (var i = 0; i < dependencies.length; i += 1) {
        await load(dependencies[i]);
      }
      dispatch("goshtoso:dependencies-ready", {
        loaded: Object.assign({}, state.loaded),
        sources: Object.assign({}, state.sources),
      });
      return state;
    } catch (error) {
      state.error = error;
      dispatch("goshtoso:dependency-error", { error: error });
      throw error;
    }
  })();
})();
