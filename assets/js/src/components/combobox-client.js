// Delegated client-mode selection and persistence for Combobox.
(function () {
  "use strict";

  if (window.__goshtosoComboboxClientInit) return;
  window.__goshtosoComboboxClientInit = true;

  function clientRoot(target) {
    if (!target || !target.closest) return null;
    return target.closest('[data-combobox][data-combobox-mode="client"]');
  }

  function readConfig(root) {
    return {
      id: root.getAttribute("id"),
      name: root.getAttribute("data-combobox-name") || root.getAttribute("id"),
      multi: root.getAttribute("data-combobox-multi") === "true",
      closeOnSelect: root.getAttribute("data-combobox-close-on-select") === "true",
      placeholder: root.getAttribute("data-combobox-placeholder") || "",
    };
  }

  function cssEscape(value) {
    return window.CSS.escape(String(value));
  }

  function readSelected(root, name) {
    return Array.from(
      root.querySelectorAll('input[type=hidden][name="' + cssEscape(name) + '"]'),
    ).map(function (element) {
      return element.value;
    });
  }

  function storageKey(config) {
    return "goshtoso:combobox:" + config.id + ":selected";
  }

  function saveSelected(config, selected) {
    try {
      window.sessionStorage.setItem(storageKey(config), JSON.stringify(selected));
    } catch (error) {
      // Storage can be unavailable under privacy policies. DOM state still works.
    }
  }

  function computeLabel(root, selected, config) {
    if (selected.length === 0) return config.placeholder || "Select…";
    if (selected.length === 1) {
      var option = root.querySelector(
        '[data-combobox-option][data-value="' + cssEscape(selected[0]) + '"]',
      );
      if (option) {
        var label = option.querySelector("[data-combobox-option-label]");
        if (label) return label.textContent;
      }
      return selected[0];
    }
    return selected.length + " selected";
  }

  function setHiddenInputs(root, name, selected) {
    var body = root.querySelector("[data-combobox-body]");
    if (!body) return;
    body
      .querySelectorAll('input[type=hidden][name="' + cssEscape(name) + '"]')
      .forEach(function (element) {
        element.remove();
      });
    selected.forEach(function (value) {
      var input = document.createElement("input");
      input.type = "hidden";
      input.name = name;
      input.value = value;
      body.insertBefore(input, body.firstChild);
    });
  }

  function selectedTriggerClasses(hasSelection) {
    if (hasSelection) {
      return {
        add: [
          "border-secondary",
          "bg-secondary/10",
          "text-on-surface-strong",
          "dark:border-secondary-dark",
          "dark:bg-secondary-dark/15",
          "dark:text-on-surface-dark-strong",
        ],
        remove: [
          "border-outline",
          "bg-surface-alt",
          "text-on-surface",
          "dark:border-outline-dark",
          "dark:bg-surface-dark-alt/50",
          "dark:text-on-surface-dark",
        ],
      };
    }
    return {
      add: [
        "border-outline",
        "bg-surface-alt",
        "text-on-surface",
        "dark:border-outline-dark",
        "dark:bg-surface-dark-alt/50",
        "dark:text-on-surface-dark",
      ],
      remove: [
        "border-secondary",
        "bg-secondary/10",
        "text-on-surface-strong",
        "dark:border-secondary-dark",
        "dark:bg-secondary-dark/15",
        "dark:text-on-surface-dark-strong",
      ],
    };
  }

  function updateUI(root) {
    var config = readConfig(root);
    var selected = readSelected(root, config.name);
    var labelText = computeLabel(root, selected, config);

    root.querySelectorAll("[data-combobox-trigger-label-outer]").forEach(function (element) {
      element.textContent = labelText;
    });

    var trigger = document.getElementById(config.id + "-trigger");
    if (trigger) {
      var groups = selectedTriggerClasses(selected.length > 0);
      groups.add.forEach(function (className) {
        trigger.classList.add(className);
      });
      groups.remove.forEach(function (className) {
        trigger.classList.remove(className);
      });
    }

    root.querySelectorAll("[data-combobox-option]").forEach(function (option) {
      var value = option.getAttribute("data-value");
      var isSelected = selected.indexOf(value) >= 0;
      option.setAttribute("aria-selected", String(isSelected));
      var checkbox = option.querySelector('input[type="checkbox"]');
      if (checkbox) checkbox.checked = isSelected;
    });

    root.querySelectorAll("[data-combobox-clear]").forEach(function (button) {
      if (selected.length === 0) button.setAttribute("hidden", "");
      else button.removeAttribute("hidden");
    });
  }

  function dispatchChange(root, selected, config) {
    root.dispatchEvent(
      new CustomEvent("combobox:change", {
        bubbles: true,
        detail: { id: config.id, values: selected },
      }),
    );
  }

  function closeAfterSelect(root, config) {
    if (!config.closeOnSelect || !window.Alpine || !window.Alpine.$data) return;
    var data = window.Alpine.$data(root);
    if (!data) return;
    data.isOpen = false;
    data.openedWithKeyboard = false;
    data.focusIndex = -1;
  }

  function toggleValue(root, value) {
    var config = readConfig(root);
    var selected = readSelected(root, config.name);
    if (!config.multi) {
      selected = [value];
    } else {
      var index = selected.indexOf(value);
      if (index >= 0) selected.splice(index, 1);
      else selected.push(value);
    }
    setHiddenInputs(root, config.name, selected);
    saveSelected(config, selected);
    updateUI(root);
    closeAfterSelect(root, config);
    dispatchChange(root, selected, config);
  }

  function clearAll(root) {
    var config = readConfig(root);
    setHiddenInputs(root, config.name, []);
    saveSelected(config, []);
    updateUI(root);
    dispatchChange(root, [], config);
  }

  function restoreSelected(root) {
    var config = readConfig(root);
    try {
      var raw = window.sessionStorage.getItem(storageKey(config));
      if (!raw) return;
      var selected = JSON.parse(raw);
      if (!Array.isArray(selected)) return;
      setHiddenInputs(root, config.name, selected.map(String));
      updateUI(root);
    } catch (error) {
      // Invalid or unavailable storage must not break the component.
    }
  }

  function restoreAllClientRoots() {
    document
      .querySelectorAll('[data-combobox][data-combobox-mode="client"]')
      .forEach(restoreSelected);
  }

  document.addEventListener(
    "click",
    function (event) {
      var root = clientRoot(event.target);
      if (!root) return;

      var clearButton = event.target.closest("[data-combobox-clear]");
      if (clearButton && root.contains(clearButton)) {
        event.preventDefault();
        clearAll(root);
        return;
      }

      var option = event.target.closest("[data-combobox-option]");
      if (!option || !root.contains(option)) return;
      if (option.getAttribute("aria-disabled") === "true") return;

      event.preventDefault();
      toggleValue(root, option.getAttribute("data-value"));
    },
    true,
  );

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", restoreAllClientRoots, { once: true });
  } else {
    restoreAllClientRoots();
  }
  window.addEventListener("pageshow", restoreAllClientRoots);
})();
