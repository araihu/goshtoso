// tooltip.js — custom-trigger accessibility wiring for Tooltip.
(function () {
  if (window.goshtosoInitTooltipTrigger) return;

  window.goshtosoInitTooltipTrigger = function (root) {
    var removeDescription = function (element, token) {
      var ids = (element.getAttribute("aria-describedby") || "")
        .split(/\s+/)
        .filter(Boolean);
      ids = ids.filter(function (candidate) {
        return candidate !== token;
      });
      if (ids.length) element.setAttribute("aria-describedby", ids.join(" "));
      else element.removeAttribute("aria-describedby");
    };
    var restoreAttribute = function (element, name, snapshot) {
      if (!snapshot || !snapshot.owned) return;
      if (element.getAttribute(name) !== snapshot.appliedValue) return;
      if (snapshot.present) element.setAttribute(name, snapshot.value);
      else element.removeAttribute(name);
    };
    var id = root.dataset.tooltipContentId;
    var previousId = root.__goshtosoTooltipContentID;
    var previousTarget = root.__goshtosoTooltipTarget;
    if (previousTarget && previousId && root.__goshtosoTooltipTargetDescriptionAdded) {
      removeDescription(previousTarget, previousId);
    }
    root.__goshtosoTooltipTarget = null;
    root.__goshtosoTooltipTargetDescriptionAdded = false;
    if (root.__goshtosoTooltipFallbackKeydown) {
      root.removeEventListener("keydown", root.__goshtosoTooltipFallbackKeydown);
      root.__goshtosoTooltipFallbackKeydown = null;
    }
    var previousFallback = root.__goshtosoTooltipFallbackState;
    if (previousFallback) {
      restoreAttribute(root, "tabindex", previousFallback.tabindex);
      restoreAttribute(root, "role", previousFallback.role);
      if (previousFallback.descriptionAdded) {
        removeDescription(root, previousFallback.contentID);
      }
      root.__goshtosoTooltipFallbackState = null;
    }
    root.__goshtosoTooltipContentID = id || "";
    if (!id) return;
    var focusableSelector =
      'button:not([disabled]),a[href],area[href],input:not([disabled]):not([type="hidden"]),select:not([disabled]),textarea:not([disabled]),audio[controls],video[controls],summary,[contenteditable]:not([contenteditable="false"]),[tabindex]:not([tabindex="-1"])';
    var candidates = root.querySelectorAll(focusableSelector);
    var target = Array.prototype.find.call(candidates, function (candidate) {
      return candidate.tabIndex >= 0 && !candidate.closest("[hidden],[inert]");
    });
    if (target) {
      var ids = (target.getAttribute("aria-describedby") || "")
        .split(/\s+/)
        .filter(Boolean);
      var descriptionAdded = ids.indexOf(id) === -1;
      if (descriptionAdded) {
        ids.push(id);
        target.setAttribute("aria-describedby", ids.join(" "));
      }
      root.__goshtosoTooltipTarget = target;
      root.__goshtosoTooltipTargetDescriptionAdded = descriptionAdded;
      return;
    }
    var interactiveSelector =
      'button,a[href],area[href],input,select,textarea,audio[controls],video[controls],summary,[contenteditable],[tabindex],[role~="button"],[role~="link"],[role~="checkbox"],[role~="radio"],[role~="switch"],[role~="tab"],[role~="menuitem"],[role~="option"],[role~="textbox"],[role~="slider"],[role~="spinbutton"],[role~="combobox"],[role~="treeitem"]';
    if (root.querySelector(interactiveSelector)) return;
    var fallbackState = {
      contentID: id,
      tabindex: {
        present: root.hasAttribute("tabindex"),
        value: root.getAttribute("tabindex"),
        owned: false,
        appliedValue: null,
      },
      role: {
        present: root.hasAttribute("role"),
        value: root.getAttribute("role"),
        owned: false,
        appliedValue: null,
      },
      descriptionAdded: false,
    };
    if (!fallbackState.tabindex.present) {
      root.setAttribute("tabindex", "0");
      fallbackState.tabindex.owned = true;
      fallbackState.tabindex.appliedValue = "0";
    }
    if (!fallbackState.role.present) {
      root.setAttribute("role", "button");
      fallbackState.role.owned = true;
      fallbackState.role.appliedValue = "button";
    }
    var fallbackIDs = (root.getAttribute("aria-describedby") || "")
      .split(/\s+/)
      .filter(Boolean);
    if (fallbackIDs.indexOf(id) === -1) {
      fallbackIDs.push(id);
      root.setAttribute("aria-describedby", fallbackIDs.join(" "));
      fallbackState.descriptionAdded = true;
    }
    root.__goshtosoTooltipFallbackState = fallbackState;
    if (root.dataset.tooltipActivation === "click") {
      var fallbackKeydown = function (event) {
        if (event.target !== root || (event.key !== "Enter" && event.key !== " ")) return;
        event.preventDefault();
        root.click();
      };
      root.__goshtosoTooltipFallbackKeydown = fallbackKeydown;
      root.addEventListener("keydown", fallbackKeydown);
    }
  };
})();
