(function () {
  "use strict";
  var selector = '[data-goshtoso-layout="container"]';
  var itemSelector = '[data-goshtoso-layout="item"]';
  var owners = new Map();
  var styles = new WeakMap();
  var visualRoots = new WeakSet();
  var observer, frame = 0, pickup = null;
  var props = ["--layout-columns", "--layout-row-span", "--layout-masonry", "--layout-track-min", "--layout-drag-width", "--layout-origin-height"];

  function items(root) {
    return Array.from(root.children).filter(function (node) { return node.matches(itemSelector); });
  }
  function local(root, marker) {
    var candidates = Array.from(root.children);
    items(root).forEach(function (item) { var button = handle(item); if (button) candidates.push(button); });
    return candidates.filter(function (node) { return node.matches(marker); });
  }
  function handle(item) {
    return item.querySelector(":scope > [data-layout-handle-row] > [data-layout-handle]");
  }
  function emit(root, name, detail, cancelable) {
    return root.dispatchEvent(new CustomEvent(name, { bubbles: true, cancelable: !!cancelable, detail: detail }));
  }
  function error(root, code, itemId) { emit(root, "layout:error", { containerId: root.dataset.layoutId, code: code, ...(itemId ? {itemId: itemId} : {}) }); }
  function nativeDisabled(root) {
    for (var node = root.parentElement; node; node = node.parentElement) {
      if (node.tagName !== "FIELDSET" || !node.disabled) continue;
      var legend = Array.from(node.children).find(function (child) { return child.tagName === "LEGEND"; });
      if (!legend || !legend.contains(root)) return true;
    }
    return false;
  }
  function available(root) {
    var state = owners.get(root);
    return state && root.isConnected && root.dataset.layoutReorder === "true" && !state.blocked && !nativeDisabled(root) && root.getClientRects().length > 0 && getComputedStyle(root).visibility !== "hidden";
  }
  function compatible(source, target) {
    return source === target || (source.dataset.layoutGroup !== "" && source.dataset.layoutGroup === target.dataset.layoutGroup);
  }
  function setStyle(node, key, value) {
    var owned = styles.get(node) || {};
    owned[key] = value;
    styles.set(node, owned);
    if (node.style.getPropertyValue(key) !== value) node.style.setProperty(key, value);
  }
  function stateAttr(root, key, value) { if (root.dataset[key] !== value) root.dataset[key] = value; }
  function focusItem(item, fallback) {
    var button = item && handle(item);
    if (button && button.isConnected && !button.matches(":disabled") && !button.hidden) button.focus({preventScroll: true});
    else if (fallback && fallback.isConnected) fallback.focus({preventScroll: true});
  }
  function announce(root, text) { var status = local(root,"[data-layout-status]")[0]; if (status) status.textContent = text; }
  function cancel(restoreFocus) {
    if (!pickup) return;
    var old = pickup; pickup = null;
    clearTarget(); cleanupVisuals(old);
    if (old.active) announce(old.source, old.item.dataset.layoutCancel || "");
    if (old.active && (restoreFocus !== false || document.activeElement === old.button)) focusItem(old.item, old.source);
  }
  function clearTarget() {
    owners.forEach(function (_, root) { local(root, "[data-layout-drop-target]").forEach(function (node) { node.removeAttribute("data-layout-drop-target"); }); root.removeAttribute("data-layout-drop-target"); });
  }
  function toggleHandles(root, enabled) {
    local(root, "[data-layout-handle]").forEach(function (button) {
      if (!enabled && document.activeElement === button) root.focus({preventScroll:true});
      if (button.hidden === enabled) button.hidden = !enabled;
      if (button.disabled === enabled) button.disabled = !enabled;
    });
    local(root,"[data-layout-instructions]").forEach(function (node) { if (node.hidden === enabled) node.hidden = !enabled; });
  }
  function addCollision(map, key, value) { if (!map.has(key)) map.set(key, []); map.get(key).push(value); }
  function validateRegistry() {
    var dom = new Map(), containers = new Map(), groupItems = new Map(), blocked = new Map();
    document.querySelectorAll("[id]").forEach(function (node) { addCollision(dom, node.id, node); });
    owners.forEach(function (_, root) {
      addCollision(containers,root.dataset.layoutId,root);
      var scope = root.dataset.layoutGroup || root;
      if (!groupItems.has(scope)) groupItems.set(scope,new Map());
      items(root).forEach(function (item) { addCollision(groupItems.get(scope),item.dataset.layoutItemId,root); });
      [root].concat(items(root),local(root,"[data-layout-instructions]")).forEach(function (node) {
        if (node.id && (dom.get(node.id) || []).length !== 1) blocked.set(root,"duplicate_dom_id");
      });
    });
    groupItems.forEach(function (ids) { ids.forEach(function (roots) { if (roots.length > 1) roots.forEach(function (root) { blocked.set(root,"duplicate_item_id"); }); }); });
    containers.forEach(function (roots) { if (roots.length > 1) roots.forEach(function (root) { blocked.set(root,"duplicate_container_id"); }); });
    owners.forEach(function (state,root) {
      var code = blocked.get(root) || "";
      if (code && state.blocked !== code) error(root,code);
      state.blocked = code;
      stateAttr(root,"layoutReorderState",code ? "blocked" : root.dataset.layoutReorder === "true" ? "ready" : "disabled");
      var enabled = root.dataset.layoutReorder === "true";
      if (enabled !== state.listening) {
        var action = enabled ? "addEventListener" : "removeEventListener";
        root[action]("keydown",state.key); root[action]("pointerdown",state.down); root[action]("layout:request-move",state.request);
        state.listening = enabled;
      }
      local(root,"[data-layout-handle]").forEach(function (button) { if (button.getAttribute("aria-describedby") !== root.id+"--layout-instructions") button.setAttribute("aria-describedby",root.id+"--layout-instructions"); });
      toggleHandles(root,!code && enabled);
    });
    if (pickup && (!available(pickup.source) || !validPickup())) cancel(false);
  }
  function resolve(source, detail) {
    if (!detail || typeof detail !== "object" || Array.isArray(detail)) return null;
    var keys = Object.keys(detail).sort().join(",");
    if (keys !== "beforeItemId,fromContainerId,itemId,toContainerId") return null;
    if ([detail.itemId,detail.fromContainerId,detail.toContainerId].some(function (id) { return typeof id !== "string" || id === ""; })) return null;
    if (detail.beforeItemId !== null && typeof detail.beforeItemId !== "string") return null;
    if (source.dataset.layoutId !== detail.fromContainerId) return null;
    var target = Array.from(owners.keys()).find(function (root) { return root.dataset.layoutId === detail.toContainerId; });
    var item = items(source).find(function (node) { return node.dataset.layoutItemId === detail.itemId; });
    if (!target || !item || item.contains(target) || !compatible(source,target)) return null;
    var anchor = detail.beforeItemId === null ? null : items(target).find(function (node) { return node.dataset.layoutItemId === detail.beforeItemId; });
    if (detail.beforeItemId !== null && (!anchor || anchor === item)) return null;
    return {target:target,item:item,anchor:anchor};
  }
  function permitted(source, move) { var button = handle(move.item); return available(source) && available(move.target) && button && !button.hidden && !button.matches(":disabled"); }
  function move(source, detail) {
    validateRegistry();
    var resolved = resolve(source,detail);
    if (!resolved) { error(source,"invalid_move_request"); return; }
    if (!permitted(source,resolved)) { error(source,"reorder_unavailable",detail.itemId); return; }
    var current = items(source), next = current[current.indexOf(resolved.item)+1] || null;
    if (source === resolved.target && next === resolved.anchor) return;
    // Consumers may mutate workflow state synchronously in this cancelable callback.
    if (!emit(source,"layout:before-move",{...detail},true)) return;
    validateRegistry();
    var checked = resolve(source,detail);
    if (!checked || checked.item !== resolved.item || checked.target !== resolved.target || checked.anchor !== resolved.anchor) { error(source,"invalid_move_request"); return; }
    if (!permitted(source,checked)) { error(source,"reorder_unavailable",detail.itemId); return; }
    var old = pickup; pickup = null; clearTarget(); cleanupVisuals(old);
    var reference = checked.anchor || local(checked.target,"[data-layout-instructions]")[0] || null;
    if (checked.target.moveBefore) checked.target.moveBefore(checked.item,reference);
    else checked.target.insertBefore(checked.item,reference);
    validateRegistry(); invalidate();
    var values = {item: checked.item.getAttribute("aria-label"),container:checked.target.getAttribute("aria-label"),position:String(items(checked.target).indexOf(checked.item)+1)};
    var parts = JSON.parse(checked.target.dataset.layoutAnnouncement);
    announce(checked.target,parts.map(function (part) { return part.slot ? values[part.slot] : part.text; }).join(""));
    focusItem(checked.item,checked.target);
    emit(checked.target,"layout:move",{...detail});
  }
  function detailFor(source,item,target,anchor) { return {itemId:item.dataset.layoutItemId,fromContainerId:source.dataset.layoutId,toContainerId:target.dataset.layoutId,beforeItemId:anchor ? anchor.dataset.layoutItemId : null}; }
  function keyboard(root,event) {
    if (event.key === "Escape" && pickup) { if (pickup.active) event.preventDefault(); cancel(); return; }
    var button = event.target.closest("[data-layout-handle]");
    if (!button || button.closest(selector) !== root || !event.altKey || handle(button.closest(itemSelector)) !== button || !available(root) || button.matches(":disabled")) return;
    var item = button.closest(itemSelector), members = items(root), index = members.indexOf(item), anchor, target = root;
    if (event.key === "ArrowUp") { if (index === 0) return; anchor = members[index-1]; }
    else if (event.key === "ArrowDown") { if (index === members.length-1) return; anchor = members[index+2] || null; }
    else if (event.key === "Home") { if (index === 0) return; anchor = members[0]; }
    else if (event.key === "End") { if (index === members.length-1) return; anchor = null; }
    else if (event.key === "PageUp" || event.key === "PageDown") {
      var roots = Array.from(document.querySelectorAll(selector)).filter(function (candidate) { return available(candidate) && compatible(root,candidate); });
      target = roots[roots.indexOf(root)+(event.key === "PageUp" ? -1 : 1)];
      if (!target) return;
      anchor = null;
    } else return;
    event.preventDefault(); move(root,detailFor(root,item,target,anchor));
  }
  var staticData = ["data-theme","data-state","data-orientation","data-layout-mode","data-layout-gap","data-layout-span","data-layout-geometry-state"];
  function staticGhost(node) {
    if (node.nodeType === 3) return node.cloneNode();
    if (node.nodeType !== 1) return null;
    if (node.localName.includes("-") || ["script","style","link","iframe","object","embed","audio","video"].includes(node.localName)) {
      var box = node.getBoundingClientRect();
      if (!box.width || !box.height) return null;
      var replacement = document.createElement("div");
      replacement.style.width = box.width+"px"; replacement.style.height = box.height+"px";
      return replacement;
    }
    // Creating an ordinary element avoids customized built-in lifecycle callbacks,
    // even when the consumer removed its original `is` attribute after upgrade.
    var copy = document.createElementNS(node.namespaceURI,node.localName);
    Array.from(node.attributes).forEach(function (attr) {
      if (/^(id$|is$|name$|for$|form$|autofocus$|on|x-|hx-|data-hx-|@|:)/i.test(attr.name) || attr.name.startsWith("data-") && !staticData.includes(attr.name)) return;
      copy.setAttributeNS(attr.namespaceURI,attr.name,attr.value);
    });
    Array.from(node.childNodes).forEach(function (child) { var cloned = staticGhost(child); if (cloned) copy.append(cloned); });
    if (node.localName === "canvas") { try { copy.getContext("2d").drawImage(node,0,0); } catch (_) { /* A lost canvas has no snapshot. */ } }
    return copy;
  }
  function blockMargins(style) { return Math.max(0,parseFloat(style.marginBlockStart)||0)+Math.max(0,parseFloat(style.marginBlockEnd)||0); }
  function scrollPosition(node) {
    var x = window.scrollX, y = window.scrollY;
    for (var parent = node.parentElement; parent && parent !== document.documentElement; parent = parent.parentElement) { x += parent.scrollLeft; y += parent.scrollTop; }
    return {x:x,y:y};
  }
  function captureBounds(node) { return {box:node.getBoundingClientRect(),scroll:scrollPosition(node)}; }
  function originalBox(node,entry) {
    var scroll = scrollPosition(node), box = entry.box;
    return {left:box.left+entry.scroll.x-scroll.x,top:box.top+entry.scroll.y-scroll.y,width:box.width,height:box.height};
  }
  function containsPoint(box,event) { return event.clientX >= box.left && event.clientX <= box.left+box.width && event.clientY >= box.top && event.clientY <= box.top+box.height; }
  function captureTargets() {
    var bounds = new Map();
    owners.forEach(function (_,root) {
      if (!available(root) || !compatible(pickup.source,root) || pickup.item.contains(root)) return;
      var entry = captureBounds(root); entry.items = new Map(items(root).map(function (item) { return [item,captureBounds(item)]; })); bounds.set(root,entry);
    });
    return bounds;
  }
  function createVisuals() {
    var item = pickup.item, box = item.getBoundingClientRect(), computed = getComputedStyle(item);
    pickup.bounds = captureTargets();
    var ghost = staticGhost(item);
    ghost.classList.add("goshtoso-layout-ghost");
    ghost.setAttribute("data-layout-ghost",""); ghost.setAttribute("aria-hidden","true"); ghost.setAttribute("inert",""); ghost.setAttribute("x-ignore","");
    Array.from(computed).forEach(function (name) { if (name.startsWith("--")) ghost.style.setProperty(name,computed.getPropertyValue(name)); });
    ["fontFamily","fontSize","fontWeight","fontStyle","lineHeight","letterSpacing","direction","color"].forEach(function (name) { ghost.style[name] = computed[name]; });
    ghost.style.width = box.width+"px"; ghost.style.height = box.height+"px";
    var placeholder = document.createElement("div");
    placeholder.className = "goshtoso-layout-placeholder"; placeholder.setAttribute("data-layout-placeholder",""); placeholder.setAttribute("aria-hidden","true"); placeholder.setAttribute("x-ignore","");
    placeholder.dataset.layoutSpan = item.dataset.layoutSpan;
    visualRoots.add(ghost); visualRoots.add(placeholder);
    pickup.ghost = ghost; pickup.placeholder = placeholder; pickup.height = box.height+blockMargins(computed);
    pickup.offsetX = pickup.x-box.left; pickup.offsetY = pickup.y-box.top; pickup.animations = new Map();
    positionGhost({clientX:pickup.x,clientY:pickup.y});
    document.body.append(ghost);
    setStyle(item,"--layout-drag-width",box.width+"px");
    setStyle(pickup.source,"--layout-origin-height",getComputedStyle(pickup.source).height);
    pickup.source.setAttribute("data-layout-drag-origin",""); item.setAttribute("data-layout-drag-source","");
    pickup.target = {root:pickup.source,anchor:items(pickup.source)[items(pickup.source).indexOf(item)+1] || null};
    sizePlaceholder(pickup.source,pickup.source.dataset.layoutMode === "masonry" && !!window.ResizeObserver && !owners.get(pickup.source).fallback);
    pickup.source.insertBefore(placeholder,item);
  }
  function positionGhost(event) {
    if (pickup.ghost) pickup.ghost.style.transform = "translate3d("+(event.clientX-pickup.offsetX)+"px,"+(event.clientY-pickup.offsetY)+"px,0)";
  }
  function cleanupVisuals(old) {
    if (!old) return;
    old.animations?.forEach(function (animation) { animation.cancel(); });
    old.ghost?.remove(); old.placeholder?.remove(); old.item.removeAttribute("data-layout-drag-source"); old.source.removeAttribute("data-layout-drag-origin");
    old.item.style.removeProperty("--layout-drag-width"); old.source.style.removeProperty("--layout-origin-height");
    delete (styles.get(old.item)||{})["--layout-drag-width"]; delete (styles.get(old.source)||{})["--layout-origin-height"];
    if (old.button && old.button.hasPointerCapture(old.pointer)) old.button.releasePointerCapture(old.pointer);
  }
  function sizePlaceholder(root,masonry) {
    var placeholder = pickup.placeholder;
    placeholder.style.height = pickup.height+"px";
    placeholder.style.gridRowEnd = masonry ? "span "+Math.max(1,Math.ceil(pickup.height+Number(root.dataset.layoutGap))) : "auto";
  }
  function animateShift(before) {
    if (window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;
    before.forEach(function (box,node) {
      var next = node.getBoundingClientRect(), x = box.left-next.left, y = box.top-next.top;
      if (Math.abs(x)+Math.abs(y)<1 || !node.animate) return;
      var base = getComputedStyle(node).transform;
      var animation = node.animate([{transform:"translate("+x+"px,"+y+"px) "+(base==="none"?"":base)},{transform:base}],{duration:140,easing:"ease-out"});
      pickup.animations.set(node,animation);
    });
  }
  function showInsertion(target) {
    var old = pickup.target;
    if (old?.root === target?.root && old?.anchor === target?.anchor) return;
    var before = new Map();
    [old?.root,target?.root].filter(Boolean).forEach(function (root) { items(root).forEach(function (item) { if (item!==pickup.item) before.set(item,item.getBoundingClientRect()); }); });
    pickup.animations.forEach(function (animation) { animation.cancel(); }); pickup.animations.clear();
    pickup.target = target;
    if (target) {
      sizePlaceholder(target.root,target.root.dataset.layoutMode === "masonry" && !!window.ResizeObserver && !owners.get(target.root).fallback);
      target.root.insertBefore(pickup.placeholder,target.anchor || local(target.root,"[data-layout-instructions]")[0] || null);
    } else pickup.placeholder.remove();
    animateShift(before); invalidate();
  }
  var interactiveContent = 'a,area,button,input,select,textarea,summary,label,audio[controls],video[controls],iframe,object,embed,img[usemap],[tabindex],[draggable="true"],[contenteditable]:not([contenteditable="false"])' +
    ["button","checkbox","combobox","grid","gridcell","link","listbox","menu","menubar","menuitem","menuitemcheckbox","menuitemradio","option","radio","radiogroup","scrollbar","searchbox","slider","spinbutton","switch","tab","tablist","textbox","tree","treegrid","treeitem"].map(function (role) { return ',[role~="'+role+'"]'; }).join("");
  function pointerItem(root,event) {
    if (event.target.closest(selector) !== root) return null;
    var item = event.target.closest(itemSelector);
    if (!item || item.parentElement !== root) return null;
    var button = handle(item);
    if (event.target.closest("[data-layout-handle]") === button) return {item:item,button:button,active:true};
    var content = item.querySelector(":scope > [data-layout-content]");
    if (root.dataset.layoutContentDrag !== "true" || event.pointerType === "touch" || !content?.contains(event.target)) return null;
    var control = event.target.closest(interactiveContent);
    if (control && content.contains(control)) return null;
    return {item:item,button:button,active:false};
  }
  function pointerDown(root,event) {
    if (event.button !== 0) return;
    var candidate = pointerItem(root,event);
    if (!candidate || !candidate.button) return;
    validateRegistry();
    if (!available(root) || candidate.button.matches(":disabled")) return;
    cancel();
    pickup = {source:root,item:candidate.item,button:candidate.button,row:candidate.button.parentElement,pointer:event.pointerId,active:candidate.active,x:event.clientX,y:event.clientY,config:pickupConfig(root),members:items(root).map(function (item) { return {node:item,id:item.dataset.layoutItemId}; })};
    if (pickup.active) activatePickup();
    event.preventDefault();
  }
  function activatePickup() {
    pickup.active = true;
    pickup.button.focus({preventScroll:true});
    pickup.button.setPointerCapture(pickup.pointer);
    createVisuals();
  }
  function pickupConfig(root) {
    return JSON.stringify([root.id, root.dataset.layoutId, root.dataset.layoutMode, root.dataset.layoutColumns, root.dataset.layoutMinWidth, root.dataset.layoutGap, root.dataset.layoutGroup, root.dataset.layoutReorder, root.dataset.layoutContentDrag]);
  }
  function validPickup() {
    if (!pickup || !pickup.source.isConnected || !pickup.button.isConnected || handle(pickup.item) !== pickup.button || pickup.button.parentElement !== pickup.row || pickup.row.parentElement !== pickup.item || pickupConfig(pickup.source) !== pickup.config) return false;
    var current = items(pickup.source);
    return current.length === pickup.members.length && current.every(function (item, n) { return item === pickup.members[n].node && item.dataset.layoutItemId === pickup.members[n].id; });
  }
  function directItem(hit,root) {
    var item = hit?.closest(itemSelector);
    while (item && item.parentElement !== root) item = item.parentElement?.closest(itemSelector);
    return item;
  }
  function displacedTarget(root,hit,event) {
    if (root !== pickup.source || directItem(hit,root)) return null;
    var source = pickup.bounds.get(root);
    if (!source || containsPoint(originalBox(root,source),event)) return null;
    // A source gap can extend past its original bounds and cover the next container.
    var target = null;
    pickup.bounds.forEach(function (entry,candidate) {
      if (candidate !== root && available(candidate) && compatible(root,candidate) && containsPoint(originalBox(candidate,entry),event) && (!target || target.contains(candidate))) target = candidate;
    });
    return target;
  }
  function pointerTarget(event) {
    if (!pickup) return null;
    var hit = document.elementFromPoint(event.clientX,event.clientY), root = hit?.closest(selector);
    // An outer gesture may hit an inner renderer; ownership stays at its level.
    while (root && (!available(root) || !compatible(pickup.source,root) || pickup.item.contains(root))) root = root.parentElement?.closest(selector);
    if (!root) return null;
    var displaced = displacedTarget(root,hit,event), anchor, box;
    if (displaced) {
      root = displaced;
      pickup.bounds.get(root).items.forEach(function (entry,item) { var original = originalBox(item,entry); if (containsPoint(original,event)) { anchor = item; box = original; } });
    } else {
      if (hit?.closest("[data-layout-placeholder]") === pickup.placeholder && pickup.target?.root === root) return pickup.target;
      anchor = directItem(hit,root); box = anchor?.getBoundingClientRect();
    }
    if (anchor === pickup.item) return null;
    if (anchor && event.clientY > box.top+box.height/2) anchor = items(root)[items(root).indexOf(anchor)+1] || null;
    if (anchor === pickup.item) anchor = items(root)[items(root).indexOf(anchor)+1] || null;
    return {root:root,anchor:anchor || null};
  }
  function pointerMove(event) {
    if (!pickup || pickup.pointer !== event.pointerId) return;
    if (!validPickup() || !available(pickup.source)) { cancel(false); return; }
    if (!pickup.active) {
      if (Math.hypot(event.clientX-pickup.x,event.clientY-pickup.y) < 6) return;
      activatePickup();
    }
    positionGhost(event);
    var target = pointerTarget(event); clearTarget(); showInsertion(target);
    if (target) (target.anchor || target.root).setAttribute("data-layout-drop-target","");
  }
  function pointerUp(event) {
    if (!pickup || pickup.pointer !== event.pointerId) return;
    if (!validPickup() || !available(pickup.source)) { cancel(false); return; }
    if (!pickup.active) { pickup = null; return; }
    var old = pickup, target = pointerTarget(event);
    if (!target) { cancel(); return; }
    pickup = null; clearTarget(); cleanupVisuals(old); move(old.source,detailFor(old.source,old.item,target.root,target.anchor));
  }

  function invalidate() {
    owners.forEach(function (state,root) { state.passes = 0; state.last = null; state.fallback = false; stateAttr(root,"layoutGeometryState","pending"); });
    schedule();
  }
  function schedule() { if (!frame && owners.size) frame = requestAnimationFrame(measure); }
  function columns(root) {
    var style = getComputedStyle(root), width = root.clientWidth-parseFloat(style.paddingLeft)-parseFloat(style.paddingRight);
    var gap = Number(root.dataset.layoutGap), min = Number(root.dataset.layoutMinWidth)*parseFloat(getComputedStyle(document.documentElement).fontSize);
    var count = root.dataset.layoutMode === "" ? 1 : Math.max(1,Math.min(Number(root.dataset.layoutColumns),Math.floor((width+gap)/(min+gap))));
    setStyle(root,"--layout-columns",String(count)); setStyle(root,"--layout-track-min","0px"); return count;
  }
  function measure() {
    frame = 0;
    var roots = Array.from(document.querySelectorAll(selector)).filter(function (root) { return owners.has(root) && root.getClientRects().length; });
    roots.forEach(columns);
    var again = false;
    roots.reverse().forEach(function (root) {
      var state = owners.get(root), masonry = root.dataset.layoutMode === "masonry" && !!window.ResizeObserver && !state.fallback;
      setStyle(root,"--layout-masonry",masonry ? "1px" : "auto");
      if (pickup?.placeholder?.parentElement === root) sizePlaceholder(root,masonry);
      var heights = [], spans = [];
      items(root).forEach(function (item) {
        var style = getComputedStyle(item), height = item.getBoundingClientRect().height+blockMargins(style);
        var span = masonry ? String(Math.max(1,Math.ceil(height+Number(root.dataset.layoutGap)))) : "auto";
        setStyle(item,"--layout-row-span",span); heights.push(height); spans.push(span);
      });
      var next = {columns:root.style.getPropertyValue("--layout-columns"),heights:heights,spans:spans};
      var settled = state.last && state.last.columns === next.columns && JSON.stringify(state.last.spans) === JSON.stringify(spans) && state.last.heights.length === heights.length && heights.every(function (h,n) { return Math.abs(h-state.last.heights[n])<=0.5; });
      state.last = next; state.passes += 1;
      if (settled || !masonry) stateAttr(root,"layoutGeometryState",state.fallback || (root.dataset.layoutMode === "masonry" && !window.ResizeObserver) ? "fallback" : "settled");
      else if (state.passes >= 4) { state.fallback = true; setStyle(root,"--layout-masonry","auto"); items(root).forEach(function (item) { setStyle(item,"--layout-row-span","auto"); }); stateAttr(root,"layoutGeometryState","fallback"); error(root,"geometry_unstable"); }
      else again = true;
    });
    if (again) schedule();
  }
  function observeSizes(state,root) {
    if (!state.resize) return;
    state.resize.disconnect(); state.sizes = new WeakMap();
    [root].concat(items(root)).forEach(function (node) { state.resize.observe(node); });
  }
  function connect(root) {
    if (owners.has(root) || !root.isConnected) return;
    var state = {blocked:"",passes:0,last:null,fallback:false,sizes:new WeakMap()};
    state.key = function (event) { keyboard(root,event); };
    state.down = function (event) { pointerDown(root,event); };
    state.request = function (event) { if (event.target === root) { event.stopPropagation(); move(root,event.detail); } };
    if (window.ResizeObserver) state.resize = new ResizeObserver(function (entries) {
      var changed = false;
      entries.forEach(function (entry) {
        var size = entry.target === root ? [entry.contentRect.width] : [entry.contentRect.width,entry.contentRect.height];
        var previous = state.sizes.get(entry.target);
        if (previous && JSON.stringify(previous) !== JSON.stringify(size)) changed = true;
        state.sizes.set(entry.target,size);
      });
      if (changed) invalidate(); else schedule();
    });
    owners.set(root,state); observeSizes(state,root);

    startObserver(); validateRegistry(); invalidate();
  }
  function disconnect(root) {
    var state = owners.get(root); if (!state) return;
    if (pickup && (pickup.source === root || root.contains(pickup.item))) cancel(false);
    toggleHandles(root,false);
    if (state.resize) state.resize.disconnect();
    root.removeEventListener("keydown",state.key); root.removeEventListener("pointerdown",state.down); root.removeEventListener("layout:request-move",state.request);
    owners.delete(root);
    if (!owners.size) { if (frame) cancelAnimationFrame(frame); frame = 0; if (observer) observer.disconnect(); observer = null; }
  }
  function scan(scope) { if (scope.matches && scope.matches(selector)) connect(scope); if (scope.querySelectorAll) scope.querySelectorAll(selector).forEach(connect); }
  function strippedStyle(text) { var node = document.createElement("div"); node.style.cssText = text || ""; props.forEach(function (prop) { node.style.removeProperty(prop); }); return node.style.cssText; }
  function visualNode(node) {
    for (; node; node = node.parentElement) if (visualRoots.has(node)) return true;
    return false;
  }
  function internal(record) {
    if (visualNode(record.target)) return true;
    if (record.type === "childList" && Array.from(record.addedNodes).concat(Array.from(record.removedNodes)).every(function (node) { return visualRoots.has(node); })) return true;
    if (record.type !== "attributes") return false;
    var key = record.attributeName;
    if (key === "style") return strippedStyle(record.oldValue) === strippedStyle(record.target.getAttribute("style"));
    if (["data-layout-geometry-state","data-layout-reorder-state","data-layout-drop-target","data-layout-drag-source","data-layout-drag-origin"].includes(key)) return true;
    if (key === "hidden" && record.target.matches("[data-layout-handle],[data-layout-instructions]")) return true;
    return key === "disabled" && record.target.matches("[data-layout-handle]");
  }
  function startObserver() {
    if (observer) return;
    observer = new MutationObserver(function (records) {
      var relevant = records.filter(function (record) { return !internal(record) && !record.target.closest?.("[data-layout-status]"); });
      if (!relevant.length) return;
      if (pickup && !validPickup()) cancel(false);
      Array.from(owners.keys()).forEach(function (root) { if (!root.isConnected) disconnect(root); });
      scan(document);
      owners.forEach(function (state,root) { if (relevant.some(function (record) { return record.type === "childList" && record.target === root; })) observeSizes(state,root); });
      validateRegistry(); invalidate();
    });
    observer.observe(document.documentElement,{childList:true,subtree:true,attributes:true,attributeOldValue:true});
  }
  function register() { window.Alpine.directive("layout",function (root,binding,utilities) { connect(root); utilities.cleanup(function () { disconnect(root); }); }); }
  if (window.Alpine) register(); else document.addEventListener("alpine:init",register,{once:true});
  document.addEventListener("DOMContentLoaded",function () { scan(document); });
  document.addEventListener("htmx:after:process",function (event) { scan(event.target); validateRegistry(); invalidate(); });
  document.addEventListener("htmx:before:cleanup",function (event) { Array.from(owners.keys()).forEach(function (root) { if (event.target === root || event.target.contains(root)) disconnect(root); }); });
  document.addEventListener("htmx:before:morph:node",function (event) {
    var old = event.detail.oldNode, next = event.detail.newNode, owned = styles.get(old);
    if (!owned || !next.style) return;
    Object.keys(owned).forEach(function (key) { next.style.setProperty(key,owned[key]); });
  });
  document.addEventListener("pointermove",pointerMove); document.addEventListener("pointerup",pointerUp);
  document.addEventListener("pointercancel",cancel); document.addEventListener("keydown",function (event) { if (event.key === "Escape" && pickup) { if (pickup.active) event.preventDefault(); cancel(); } });
  window.addEventListener("blur",function () { cancel(false); });
  window.addEventListener("resize",function () { cancel(false); invalidate(); });
  if (document.fonts) document.fonts.addEventListener("loadingdone",invalidate);
  if (document.readyState !== "loading") scan(document);
})();
