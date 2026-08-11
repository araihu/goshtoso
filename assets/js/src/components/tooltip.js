// tooltip.js — custom-trigger accessibility wiring for Tooltip.
(function () {
  if (window.goshtosoInitTooltipTrigger) return;

  window.goshtosoInitTooltipTrigger = function (root, expanded) {
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
    var setOwnedAttribute = function (element, name, value) {
      var snapshot = {
        present: element.hasAttribute(name),
        value: element.getAttribute(name),
        owned: true,
        appliedValue: value,
      };
      element.setAttribute(name, value);
      return snapshot;
    };
    var id = root.dataset.tooltipContentId;
    var isPersistent = root.dataset.tooltipActivation === "click";
    var expandedValue =
      expanded === true || expanded === "true" ? "true" : "false";
    var previousId = root.__goshtosoTooltipContentID;
    var previousTarget = root.__goshtosoTooltipTarget;
    if (previousTarget && previousId && root.__goshtosoTooltipTargetDescriptionAdded) {
      removeDescription(previousTarget, previousId);
    }
    if (previousTarget) {
      restoreAttribute(
        previousTarget,
        "aria-controls",
        root.__goshtosoTooltipTargetControls,
      );
      restoreAttribute(
        previousTarget,
        "aria-expanded",
        root.__goshtosoTooltipTargetExpanded,
      );
      if (root.__goshtosoTooltipTargetKeydown) {
        previousTarget.removeEventListener(
          "keydown",
          root.__goshtosoTooltipTargetKeydown,
        );
      }
    }
    root.__goshtosoTooltipTarget = null;
    root.__goshtosoTooltipTargetDescriptionAdded = false;
    root.__goshtosoTooltipTargetControls = null;
    root.__goshtosoTooltipTargetExpanded = null;
    root.__goshtosoTooltipTargetKeydown = null;
    if (root.__goshtosoTooltipFallbackKeydown) {
      root.removeEventListener("keydown", root.__goshtosoTooltipFallbackKeydown);
      root.__goshtosoTooltipFallbackKeydown = null;
    }
    var previousFallback = root.__goshtosoTooltipFallbackState;
    if (previousFallback) {
      restoreAttribute(root, "tabindex", previousFallback.tabindex);
      restoreAttribute(root, "role", previousFallback.role);
      restoreAttribute(root, "aria-controls", previousFallback.controls);
      restoreAttribute(root, "aria-expanded", previousFallback.expanded);
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
      if (isPersistent) {
        root.__goshtosoTooltipTargetControls = setOwnedAttribute(
          target,
          "aria-controls",
          id,
        );
        root.__goshtosoTooltipTargetExpanded = setOwnedAttribute(
          target,
          "aria-expanded",
          expandedValue,
        );
        var roleTokens = (target.getAttribute("role") || "")
          .split(/\s+/)
          .filter(Boolean);
        var isNativeButton =
          target.localName === "button" ||
          (target.localName === "input" &&
            ["button", "submit", "reset", "image"].indexOf(
              (target.getAttribute("type") || "text").toLowerCase(),
            ) !== -1);
        if (roleTokens.indexOf("button") !== -1 && !isNativeButton) {
          var targetKeydown = function (event) {
            if (
              event.target !== target ||
              (event.key !== "Enter" && event.key !== " ")
            ) {
              return;
            }
            event.preventDefault();
            target.click();
          };
          root.__goshtosoTooltipTargetKeydown = targetKeydown;
          target.addEventListener("keydown", targetKeydown);
        }
      }
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
      controls: null,
      expanded: null,
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
    if (isPersistent) {
      fallbackState.controls = setOwnedAttribute(root, "aria-controls", id);
      fallbackState.expanded = setOwnedAttribute(
        root,
        "aria-expanded",
        expandedValue,
      );
      var fallbackKeydown = function (event) {
        if (event.target !== root || (event.key !== "Enter" && event.key !== " ")) return;
        event.preventDefault();
        root.click();
      };
      root.__goshtosoTooltipFallbackKeydown = fallbackKeydown;
      root.addEventListener("keydown", fallbackKeydown);
    }
  };

  window.goshtosoSetTooltipExpanded = function (root, expanded) {
    if (root.dataset.tooltipActivation !== "click") return;
    var target = root.__goshtosoTooltipTarget;
    var snapshot = root.__goshtosoTooltipTargetExpanded;
    if (!target && root.__goshtosoTooltipFallbackState) {
      target = root;
      snapshot = root.__goshtosoTooltipFallbackState.expanded;
    }
    if (!target || !snapshot || !snapshot.owned) {
      window.goshtosoInitTooltipTrigger(root, expanded);
      return;
    }
    var value = expanded === true || expanded === "true" ? "true" : "false";
    var current = target.getAttribute("aria-expanded");
    if (current !== snapshot.appliedValue) {
      snapshot.present = target.hasAttribute("aria-expanded");
      snapshot.value = current;
    }
    target.setAttribute("aria-expanded", value);
    snapshot.appliedValue = value;
  };
})();
