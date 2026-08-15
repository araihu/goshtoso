(() => {
  const config = __GOSHTOSO_TGS011_CAPTURE_CONFIG__;
  const viewport = document.querySelector(window.__goshtosoTGS011Selector || '#scroll-region-default [data-goshtoso-scroll-viewport]');
  if (!viewport) throw new Error("ScrollRegion viewport missing");
  const rect = viewport.getBoundingClientRect();
  return JSON.stringify({
    schema: config.schema,
    pair: config.pair,
    route: location.pathname,
    challenge: new URLSearchParams(location.search).get(config.queryKey),
    candidate_tree: document.querySelector('meta[name="goshtoso-t-gs-011-candidate-tree"]')?.content,
    manifest_sha256: document.querySelector('meta[name="goshtoso-t-gs-011-manifest-sha256"]')?.content,
    window: { x: window.screenX, y: window.screenY, width: window.outerWidth, height: window.outerHeight },
    candidate_region: { x: rect.x, y: rect.y, width: rect.width, height: rect.height }
  });
})()
