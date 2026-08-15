(() => {
  const config = __GOSHTOSO_TGS011_CAPTURE_CONFIG__;
  if (!new Set(["before", "after", "exit"]).has(config.phase)) throw new Error("invalid direct browser snapshot phase");
  const viewport = document.querySelector(config.selector);
  if (!viewport) throw new Error("ScrollRegion viewport missing");
  const actionToken = document.querySelector('meta[name="goshtoso-t-gs-011-at-action-token"]')?.content;
  const actionState = document.querySelector('meta[name="goshtoso-t-gs-011-at-action-state"]')?.content;
  const announcer = document.querySelector('#goshtoso-t-gs-011-at-action-token');
  if (location.pathname !== config.route || new URLSearchParams(location.search).get(config.queryKey) !== config.challenge || new URLSearchParams(location.search).get("t-gs-011-at-state") !== config.state || new URLSearchParams(location.search).get("t-gs-011-at-action-token") !== config.actionToken || actionState !== config.state || actionToken !== config.actionToken || announcer?.textContent?.includes(config.actionToken) !== true || document.querySelector('meta[name="goshtoso-t-gs-011-at-challenge"]')?.content !== config.challenge || document.querySelector('meta[name="goshtoso-t-gs-011-candidate-tree"]')?.content !== config.candidateTree || document.querySelector('meta[name="goshtoso-t-gs-011-manifest-sha256"]')?.content !== config.manifestSHA256) throw new Error("candidate action binding mismatch");
  const boundary = () => viewport.scrollHeight <= viewport.clientHeight + 1 ? "no-overflow" : viewport.scrollTop <= 1 ? "start" : viewport.scrollTop + viewport.clientHeight >= viewport.scrollHeight - 1 ? "end" : "middle";
  const accessibleName = element => {
    if (!element) return "document body";
    const direct = element.getAttribute?.("aria-label");
    if (direct) return direct.trim();
    const labelled = element.getAttribute?.("aria-labelledby");
    if (labelled) {
      const text = labelled.split(/\s+/).map(id => document.getElementById(id)?.textContent?.trim() || "").filter(Boolean).join(" ");
      if (text) return text;
    }
    const text = element.textContent?.trim().replace(/\s+/g, " ").slice(0, 80);
    return text || (element.tagName || "document").toLowerCase();
  };
  const active = document.activeElement;
  const rect = viewport.getBoundingClientRect();
  const snapshot = {
    active_role: active?.getAttribute?.("role") || active?.tagName?.toLowerCase() || "document",
    active_name: accessibleName(active),
    region_role: viewport.getAttribute("role"),
    region_name: accessibleName(viewport),
    region_focused: active === viewport,
    boundary: boundary(),
    scroll_top: viewport.scrollTop,
    client_height: viewport.clientHeight,
    scroll_height: viewport.scrollHeight,
    start_cue_visible: viewport.scrollHeight > viewport.clientHeight + 1 && viewport.scrollTop > 1,
    end_cue_visible: viewport.scrollHeight > viewport.clientHeight + 1 && viewport.scrollTop + viewport.clientHeight < viewport.scrollHeight - 1,
  };
  return JSON.stringify({
    schema: config.schema,
    pair: config.pair,
    state: config.state,
    command: config.command,
    route: location.pathname,
    challenge: new URLSearchParams(location.search).get(config.queryKey),
    candidate_tree: document.querySelector('meta[name="goshtoso-t-gs-011-candidate-tree"]')?.content,
    manifest_sha256: document.querySelector('meta[name="goshtoso-t-gs-011-manifest-sha256"]')?.content,
    action_token: actionToken,
    phase: config.phase,
    observed_at: new Date().toISOString(),
    snapshot,
    candidate_region: { x: rect.x, y: rect.y, width: rect.width, height: rect.height }
  });
})()
