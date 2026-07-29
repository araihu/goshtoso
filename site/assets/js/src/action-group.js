// action-group.js — Alpine state for Action Group demo actions.
(function () {
  var register = function () {
    if (!window.Alpine || Alpine.__actionGroupDemoRegistered) return;
    Alpine.__actionGroupDemoRegistered = true;
    Alpine.data("actionGroupDemo", function () {
      return { lastAction: "none" };
    });
  };

  if (window.Alpine) register();
  else document.addEventListener("alpine:init", register);
})();
