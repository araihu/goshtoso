// select-demo.js — named actions used only by Select documentation previews.
(function () {
  if (window.goshtosoRestoreSelectDraft) return;
  window.goshtosoRestoreSelectDraft = function (id, value) {
    var input = document.getElementById(id);
    if (!input) return false;
    input.value = value;
    input.dispatchEvent(new Event("change", { bubbles: true }));
    return true;
  };
})();
