(function() {
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
      persist: root.getAttribute("data-combobox-persist") !== "false",
      closeOnSelect: root.getAttribute("data-combobox-close-on-select") === "true",
      placeholder: root.getAttribute("data-combobox-placeholder") || "",
      selectedLabels: JSON.parse(root.getAttribute("data-combobox-selected-labels") || "null")
    };
  }
  function cssEscape(value) {
    return window.CSS.escape(String(value));
  }
  function readSelected(root, name) {
    return Array.from(
      root.querySelectorAll('input[type=hidden][name="' + cssEscape(name) + '"]')
    ).map(function(element) {
      return element.value;
    });
  }
  function storageKey(config) {
    return "goshtoso:combobox:" + config.id + ":selected";
  }
  function saveSelected(config, selected) {
    if (!config.persist) return;
    try {
      window.sessionStorage.setItem(storageKey(config), JSON.stringify(selected));
    } catch (error) {
    }
  }
  function computeLabel(root, selected, config) {
    if (selected.length === 0) return config.placeholder || "Select\u2026";
    if (selected.length === 1) {
      var option = root.querySelector(
        '[data-combobox-option][data-value="' + cssEscape(selected[0]) + '"]'
      );
      if (option) {
        var label = option.querySelector("[data-combobox-option-label]");
        if (label) return label.textContent;
      }
      return selected[0];
    }
    return config.selectedLabels && config.selectedLabels[selected.length] !== void 0 ? config.selectedLabels[selected.length] : selected.length + " selected";
  }
  function setHiddenInputs(root, name, selected) {
    var body = root.querySelector("[data-combobox-body]");
    if (!body) return;
    body.querySelectorAll('input[type=hidden][name="' + cssEscape(name) + '"]').forEach(function(element) {
      element.remove();
    });
    selected.forEach(function(value) {
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
          "dark:text-on-surface-dark-strong"
        ],
        remove: [
          "border-outline",
          "bg-surface-alt",
          "text-on-surface",
          "dark:border-outline-dark",
          "dark:bg-surface-dark-alt/50",
          "dark:text-on-surface-dark"
        ]
      };
    }
    return {
      add: [
        "border-outline",
        "bg-surface-alt",
        "text-on-surface",
        "dark:border-outline-dark",
        "dark:bg-surface-dark-alt/50",
        "dark:text-on-surface-dark"
      ],
      remove: [
        "border-secondary",
        "bg-secondary/10",
        "text-on-surface-strong",
        "dark:border-secondary-dark",
        "dark:bg-secondary-dark/15",
        "dark:text-on-surface-dark-strong"
      ]
    };
  }
  function updateUI(root) {
    var config = readConfig(root);
    var selected = readSelected(root, config.name);
    var labelText = computeLabel(root, selected, config);
    root.querySelectorAll("[data-combobox-trigger-label-outer]").forEach(function(element) {
      element.textContent = labelText;
    });
    var trigger = document.getElementById(config.id + "-trigger");
    if (trigger) {
      var groups = selectedTriggerClasses(selected.length > 0);
      groups.add.forEach(function(className) {
        trigger.classList.add(className);
      });
      groups.remove.forEach(function(className) {
        trigger.classList.remove(className);
      });
    }
    root.querySelectorAll("[data-combobox-option]").forEach(function(option) {
      var value = option.getAttribute("data-value");
      var isSelected = selected.indexOf(value) >= 0;
      option.setAttribute("aria-selected", String(isSelected));
      var checkbox = option.querySelector('input[type="checkbox"]');
      if (checkbox) checkbox.checked = isSelected;
    });
    root.querySelectorAll("[data-combobox-clear]").forEach(function(button) {
      if (selected.length === 0) button.setAttribute("hidden", "");
      else button.removeAttribute("hidden");
    });
  }
  function dispatchChange(root, selected, config) {
    root.dispatchEvent(
      new CustomEvent("combobox:change", {
        bubbles: true,
        detail: { id: config.id, values: selected }
      })
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
    if (!config.persist) return;
    try {
      var raw = window.sessionStorage.getItem(storageKey(config));
      if (!raw) return;
      var selected = JSON.parse(raw);
      if (!Array.isArray(selected)) return;
      var allowed = new Set(readSelected(root, config.name));
      root.querySelectorAll("[data-combobox-option]").forEach(function(option) {
        allowed.add(option.getAttribute("data-value"));
      });
      selected = Array.from(new Set(selected.map(String))).filter(function(value) {
        return allowed.has(value);
      });
      if (!config.multi) selected = selected.slice(0, 1);
      setHiddenInputs(root, config.name, selected);
      updateUI(root);
    } catch (error) {
    }
  }
  function filterClientOptions(root) {
    var search = root.querySelector("[data-combobox-search]");
    var query = search ? search.value.trim().toLowerCase() : "";
    root.querySelectorAll("[data-combobox-option]").forEach(function(option) {
      option.hidden = !option.textContent.toLowerCase().includes(query);
    });
  }
  document.addEventListener("input", function(event) {
    if (!event.target.matches("[data-combobox-search]") || event.target.hasAttribute("hx-get")) return;
    var root = clientRoot(event.target);
    if (root) filterClientOptions(root);
  });
  document.addEventListener("combobox:load", function(event) {
    var detail = event.detail;
    if (!detail || typeof detail.id !== "string") return;
    var root = document.getElementById(detail.id);
    if (!root || clientRoot(root) !== root) return;
    var config = readConfig(root);
    var options = detail.options;
    var values = detail.values === void 0 ? readSelected(root, config.name) : detail.values;
    if (!Array.isArray(values) || values.some(function(value) {
      return typeof value !== "string";
    }) || !config.multi && values.length > 1) return;
    if (options !== void 0 && (!Array.isArray(options) || options.some(function(option) {
      return !option || typeof option.value !== "string" || typeof option.label !== "string";
    }))) return;
    var allowed = new Set(options === void 0 ? Array.from(root.querySelectorAll("[data-combobox-option]")).map(function(option) {
      return option.getAttribute("data-value");
    }) : options.map(function(option) {
      return option.value;
    }));
    if (values.some(function(value) {
      return !allowed.has(value);
    })) return;
    if (options !== void 0) {
      var list = root.querySelector("[role=listbox]");
      if (!list) return;
      var rows = options.map(function(option) {
        var row = document.createElement("li");
        row.setAttribute("role", "option");
        row.setAttribute("data-combobox-option", "");
        row.setAttribute("data-value", option.value);
        row.tabIndex = -1;
        row.className = "flex items-center gap-2.5 px-4 py-2.5 text-sm transition-colors motion-reduce:transition-none dark:text-on-surface-dark " + (option.disabled ? "cursor-not-allowed opacity-50" : "cursor-pointer hover:bg-surface-dark-alt/5 dark:hover:bg-surface-alt/5");
        if (option.disabled) row.setAttribute("aria-disabled", "true");
        if (config.multi) {
          var checkbox = document.createElement("input");
          checkbox.type = "checkbox";
 checkbox.className="pointer-events-none size-4 accent-primary dark:accent-primary-dark";
          checkbox.tabIndex = -1;
          checkbox.setAttribute("aria-hidden", "true");
          row.append(checkbox);
        }
        var label = document.createElement("span");
        label.setAttribute("data-combobox-option-label", "");
        label.className = "truncate";
        label.textContent = option.label;
        for (var name of Object.keys(option.labelAttrs || {})) {
          if (name === "style" || name === "title" || name.startsWith("data-") && name !== "data-combobox-option-label" || name.startsWith("aria-")) label.setAttribute(name, String(option.labelAttrs[name]));
        }
        row.append(label);
        return row;
      });
      list.replaceChildren(...rows);
    }
    setHiddenInputs(root, config.name, Array.from(new Set(values)));
    updateUI(root);
    filterClientOptions(root);
  });
  window.goshtosoComboboxClient = function(root) {
    return {
      isOpen: false,
      openedWithKeyboard: false,
      focusIndex: -1,
      init: function() {
        restoreSelected(root);
      },
      restoreSelection: function() {
        restoreSelected(root);
      }
    };
  };
  document.addEventListener(
    "click",
    function(event) {
      var root = clientRoot(event.target);
      if (!root) return;
      var clearButton = event.target.closest("[data-combobox-clear]");
      if (clearButton && root.contains(clearButton)) {
        event.preventDefault();
        var ownedFocus = document.activeElement === clearButton;
        clearAll(root);
        if (ownedFocus) {
          var trigger = root.querySelector('[role="combobox"]');
          if (trigger) trigger.focus({ preventScroll: true });
        }
        return;
      }
      var option = event.target.closest("[data-combobox-option]");
      if (!option || !root.contains(option)) return;
      if (option.getAttribute("aria-disabled") === "true") return;
      event.preventDefault();
      toggleValue(root, option.getAttribute("data-value"));
    },
    true
  );
})();
