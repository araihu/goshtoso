(function () {
  if (window.__goshtosoComboboxV2Init) return;
  window.__goshtosoComboboxV2Init = true;

  function clientRoot(target) {
    if (!target || !target.closest) return null;
    return target.closest('[data-combobox][data-combobox-mode="client"]');
  }

  function readCfg(root) {
    return {
      id: root.getAttribute('id'),
      name: root.getAttribute('data-combobox-name') || root.getAttribute('id'),
      multi: root.getAttribute('data-combobox-multi') === 'true',
      closeOnSelect: root.getAttribute('data-combobox-close-on-select') === 'true',
      placeholder: root.getAttribute('data-combobox-placeholder') || ''
    };
  }

  function readSelected(root, name) {
    return Array.from(root.querySelectorAll('input[type=hidden][name="' + cssEscape(name) + '"]'))
      .map(function (el) { return el.value; });
  }

  function storageKey(cfg) {
    return 'goshtoso:combobox:v2:' + cfg.id + ':selected';
  }

  function saveSelected(cfg, selected) {
    try {
      window.sessionStorage.setItem(storageKey(cfg), JSON.stringify(selected));
    } catch (_) {}
  }

  function restoreSelected(root) {
    var cfg = readCfg(root);
    try {
      var raw = window.sessionStorage.getItem(storageKey(cfg));
      if (!raw) return;
      var selected = JSON.parse(raw);
      if (!Array.isArray(selected)) return;
      setHiddenInputs(root, cfg.name, selected.map(String));
      updateUI(root);
    } catch (_) {}
  }

  function restoreAllClientRoots() {
    document.querySelectorAll('[data-combobox][data-combobox-mode="client"]').forEach(restoreSelected);
  }

  function cssEscape(s) {
    return window.CSS.escape(String(s));
  }

  function computeLabel(root, selected, cfg) {
    if (selected.length === 0) return cfg.placeholder || 'Select…';
    if (selected.length === 1) {
      var li = root.querySelector('[data-combobox-option][data-value="' + cssEscape(selected[0]) + '"]');
      if (li) {
        var label = li.querySelector('[data-combobox-option-label]');
        if (label) return label.textContent;
      }
      return selected[0];
    }
    return selected.length + ' selected';
  }

  function setHiddenInputs(root, name, selected) {
    var body = root.querySelector('[data-combobox-body]');
    if (!body) return;
    body.querySelectorAll('input[type=hidden][name="' + cssEscape(name) + '"]').forEach(function (el) {
      el.remove();
    });
    selected.forEach(function (v) {
      var input = document.createElement('input');
      input.type = 'hidden';
      input.name = name;
      input.value = v;
      body.insertBefore(input, body.firstChild);
    });
  }

  function selectedTriggerClasses(hasSelection) {
    if (hasSelection) {
      return {
        add: ['border-secondary', 'bg-secondary/10', 'text-on-surface-strong', 'dark:border-secondary-dark', 'dark:bg-secondary-dark/15', 'dark:text-on-surface-dark-strong'],
        remove: ['border-outline', 'bg-surface-alt', 'text-on-surface', 'dark:border-outline-dark', 'dark:bg-surface-dark-alt/50', 'dark:text-on-surface-dark']
      };
    }
    return {
      add: ['border-outline', 'bg-surface-alt', 'text-on-surface', 'dark:border-outline-dark', 'dark:bg-surface-dark-alt/50', 'dark:text-on-surface-dark'],
      remove: ['border-secondary', 'bg-secondary/10', 'text-on-surface-strong', 'dark:border-secondary-dark', 'dark:bg-secondary-dark/15', 'dark:text-on-surface-dark-strong']
    };
  }

  function updateUI(root) {
    var cfg = readCfg(root);
    var selected = readSelected(root, cfg.name);
    var labelText = computeLabel(root, selected, cfg);

    root.querySelectorAll('[data-combobox-trigger-label-outer]').forEach(function (el) {
      el.textContent = labelText;
    });

    var trigger = document.getElementById(cfg.id + '-trigger');
    if (trigger) {
      var groups = selectedTriggerClasses(selected.length > 0);
      groups.add.forEach(function (c) { trigger.classList.add(c); });
      groups.remove.forEach(function (c) { trigger.classList.remove(c); });
    }

    root.querySelectorAll('[data-combobox-option]').forEach(function (li) {
      var val = li.getAttribute('data-value');
      var sel = selected.indexOf(val) >= 0;
      li.setAttribute('aria-selected', String(sel));
      var cb = li.querySelector('input[type=checkbox]');
      if (cb) cb.checked = sel;
    });

    // Clear-all visibility: hide when no selection.
    root.querySelectorAll('[data-combobox-clear]').forEach(function (btn) {
      if (selected.length === 0) {
        btn.setAttribute('hidden', '');
      } else {
        btn.removeAttribute('hidden');
      }
    });
  }

  function dispatchChange(root, selected, cfg) {
    root.dispatchEvent(new CustomEvent('combobox:change', {
      bubbles: true,
      detail: { id: cfg.id, values: selected }
    }));
  }

  function closeAfterSelect(root, cfg) {
    if (!cfg.closeOnSelect || !window.Alpine || !window.Alpine.$data) return;
    var data = window.Alpine.$data(root);
    if (!data) return;
    data.isOpen = false;
    data.openedWithKeyboard = false;
    data.focusIndex = -1;
  }

  function toggleValue(root, value) {
    var cfg = readCfg(root);
    var selected = readSelected(root, cfg.name);
    if (!cfg.multi) {
      selected = [value];
    } else {
      var idx = selected.indexOf(value);
      if (idx >= 0) selected.splice(idx, 1);
      else selected.push(value);
    }
    setHiddenInputs(root, cfg.name, selected);
    saveSelected(cfg, selected);
    updateUI(root);
    closeAfterSelect(root, cfg);
    dispatchChange(root, selected, cfg);
  }

  function clearAll(root) {
    var cfg = readCfg(root);
    setHiddenInputs(root, cfg.name, []);
    saveSelected(cfg, []);
    updateUI(root);
    dispatchChange(root, [], cfg);
  }

  document.addEventListener('click', function (e) {
    var root = clientRoot(e.target);
    if (!root) return;

    var clearBtn = e.target.closest('[data-combobox-clear]');
    if (clearBtn && root.contains(clearBtn)) {
      e.preventDefault();
      clearAll(root);
      return;
    }

    var li = e.target.closest('[data-combobox-option]');
    if (!li || !root.contains(li)) return;
    if (li.getAttribute('aria-disabled') === 'true') return;

    e.preventDefault();
    toggleValue(root, li.getAttribute('data-value'));
  }, true);

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', restoreAllClientRoots);
  } else {
    restoreAllClientRoots();
  }
  window.addEventListener('pageshow', restoreAllClientRoots);
})();
