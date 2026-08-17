// Dropdown keeps its historical Alpine factory name as a compatibility alias.
// Shared lifecycle and menu focus behavior now live in popover.js.
(function () {
  if (window.goshtosoDropdown || !window.goshtosoPopover) return;
  window.goshtosoDropdown = window.goshtosoPopover;
})();
