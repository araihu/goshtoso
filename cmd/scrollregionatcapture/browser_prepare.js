(() => {
  const config = __GOSHTOSO_TGS011_CAPTURE_CONFIG__;
  const viewport = document.querySelector(config.selector);
  if (!viewport) throw new Error("ScrollRegion viewport missing");
  const boundary = () => viewport.scrollHeight <= viewport.clientHeight + 1 ? "no-overflow" : viewport.scrollTop <= 1 ? "start" : viewport.scrollTop + viewport.clientHeight >= viewport.scrollHeight - 1 ? "end" : "middle";
  if (location.pathname !== config.route || new URLSearchParams(location.search).get(config.queryKey) !== config.challenge || new URLSearchParams(location.search).get("t-gs-011-at-state") !== config.state || new URLSearchParams(location.search).get("t-gs-011-at-action-token") !== config.actionToken || document.querySelector('meta[name="goshtoso-t-gs-011-at-action-state"]')?.content !== config.state || document.querySelector('meta[name="goshtoso-t-gs-011-at-action-token"]')?.content !== config.actionToken || document.querySelector('meta[name="goshtoso-t-gs-011-at-challenge"]')?.content !== config.challenge || document.querySelector('meta[name="goshtoso-t-gs-011-candidate-tree"]')?.content !== config.candidateTree || document.querySelector('meta[name="goshtoso-t-gs-011-manifest-sha256"]')?.content !== config.manifestSHA256) throw new Error("candidate action binding mismatch");
  switch (config.setup) {
    case "preceding-focus": {
      const focusable = Array.from(document.querySelectorAll('a,button,input,select,textarea,[tabindex]:not([tabindex="-1"])'));
      const index = focusable.indexOf(viewport);
      if (index <= 0) throw new Error("no preceding focus target");
      focusable[index - 1].focus();
      break;
    }
    case "start":
      viewport.focus();
      viewport.scrollTop = 0;
      break;
    case "middle":
      viewport.focus();
      viewport.scrollTop = Math.max(1, Math.floor((viewport.scrollHeight - viewport.clientHeight) / 3));
      break;
    default:
      viewport.blur();
  }
  window.__goshtosoTGS011BeforeBoundary = boundary();
  window.__goshtosoTGS011Selector = config.selector;
})()
