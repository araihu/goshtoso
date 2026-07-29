// structured-input.js — Alpine factory for repeatable structured rows.
(function () {
  if (window.structuredInput) return;

  window.structuredInput = function (root) {
    function parseJSON(value, fallback) {
      try {
        return JSON.parse(value || "");
      } catch (error) {
        return fallback;
      }
    }

    return {
      name: root.dataset.name || "",
      entries: parseJSON(root.dataset.entries, []),
      newRow: parseJSON(root.dataset.newRow, []),
      addRow: function () {
        this.entries.push(this.newRow.slice());
      },
      inputName: function (index, key) {
        return this.name + "[" + index + "][" + key + "]";
      },
    };
  };
})();
