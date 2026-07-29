// Shared decoder for Base64-encoded JSON data attributes used by factories.
(function () {
  if (window.goshtosoParseData) return;

  window.goshtosoParseData = function (value, fallback) {
    try {
      var bytes = Uint8Array.from(atob(value || ""), function (character) {
        return character.charCodeAt(0);
      });
      return JSON.parse(new TextDecoder().decode(bytes));
    } catch (error) {
      return fallback;
    }
  };
})();
