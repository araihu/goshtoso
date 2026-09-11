// State shared by the editor drawer and its nested resource picker.
(function () {
  function register() {
    Alpine.data("contentEditorDemo", function () {
      return {
        selectedItem: "Getting started",
        pendingItem: "Getting started",
        itemSearch: "",
        matches: function (label) {
          return label.toLowerCase().includes(this.itemSearch.toLowerCase());
        },
        get noMatches() {
          var self = this;
          return !Array.from(this.$refs.resources.querySelectorAll("label")).some(function (label) {
            return self.matches(label.textContent);
          });
        }
      };
    });
  }
  if (window.Alpine) register();
  else document.addEventListener("alpine:init", register, { once: true });
})();
