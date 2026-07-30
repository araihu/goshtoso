// icon-catalog.js — local state for the generated Heroicons workbench.
(function () {
  function goString(value) {
    var output = '"';
    for (var _i = 0, characters = Array.from(String(value || "")); _i < characters.length; _i += 1) {
      var character = characters[_i];
      switch (character) {
        case "\\": output += "\\\\"; break;
        case '"': output += '\\"'; break;
        case "\n": output += "\\n"; break;
        case "\r": output += "\\r"; break;
        case "\t": output += "\\t"; break;
        case "\b": output += "\\b"; break;
        case "\f": output += "\\f"; break;
        default: {
          var codePoint = character.codePointAt(0);
          output += codePoint < 32 || codePoint === 127
            ? "\\x" + codePoint.toString(16).padStart(2, "0")
            : character;
        }
      }
    }
    return output + '"';
  }

  // goshtosoIconCode is the one copy encoder. Alpine and tests call this exact
  // function so visible and copied .templ source stay in lockstep.
  window.goshtosoIconCode = function (input) {
    var fields = [
      "SpriteURL: heroicons.SpriteURL",
      "Symbol:    heroicons." + input.glyph,
    ];
    var sizes = { xs: "SizeXS", sm: "SizeSM", lg: "SizeLG", xl: "SizeXL" };
    var label = String(input.label || "").trim();
    if (sizes[input.size]) fields.push("Size:      icon." + sizes[input.size]);
    if (input.decorative) fields.push("Decorative: true");
    else if (label) fields.push("Label:     " + goString(label));
    var root = String(input.rootClass || "").trim();
    if (root) fields.push("RootClass: " + goString(root));

    return [
      "@icon.Icon(icon.Config{",
    ].concat(fields.map(function (field) { return "    " + field + ","; }), [
      "})",
      "",
    ]).join("\n");
  };

  function register() {
    if (!window.Alpine || Alpine.__iconCatalogRegistered) return;
    Alpine.__iconCatalogRegistered = true;
    Alpine.data("iconCatalog", function () {
      return {
        open: false,
        selected: { name: "Icon16SolidCheck", symbol: "hi-16-solid-check" },
        size: "xl",
        label: "",
        decorative: false,
        rootClass: "text-primary",
        lastTrigger: null,
        _syncing: false,
        init: function () {
          var component = this;
          this.$watch("decorative", function (value) {
            if (value) {
              component._syncing = true;
              component.label = "";
              component._syncing = false;
            }
          });
          this.$watch("label", function (value) {
            if (!component._syncing && String(value || "").trim()) {
              component.decorative = false;
            }
          });
        },
        selectGlyph: function (card) {
          this.selected = {
            name: card.dataset.iconName || "Icon16SolidCheck",
            symbol: card.dataset.iconSymbol || "hi-16-solid-check",
          };
          this.lastTrigger = card;
          this.open = true;
          this.$nextTick(function () {
            var control = document.getElementById("icon-size-xl");
            if (control) control.focus();
          });
        },
        close: function () {
          this.open = false;
          var trigger = this.lastTrigger;
          this.$nextTick(function () {
            if (trigger) trigger.focus();
          });
        },
        get isDecorative() {
          return this.decorative || !String(this.label || "").trim();
        },
        get spriteHref() {
          return "/assets/icons/heroicons.svg#" + this.selected.symbol;
        },
        get previewClass() {
          var classes = { xs: "size-3", sm: "size-4", "": "size-5", lg: "size-6", xl: "size-8" };
          var root = String(this.rootClass || "").trim();
          var darkClasses = {
            "text-primary": "dark:text-primary-dark",
            "text-secondary": "dark:text-secondary-dark",
            "text-success": "dark:text-success-dark",
            "text-warning": "dark:text-warning-dark",
            "text-danger": "dark:text-danger-dark",
          };
          return classes[this.size] + (root ? " " + root : "") + (darkClasses[root] ? " " + darkClasses[root] : "");
        },
        get code() {
          return window.goshtosoIconCode({
            glyph: this.selected.name,
            size: this.size,
            label: this.label,
            decorative: this.decorative,
            rootClass: this.rootClass,
          });
        },
      };
    });
  }

  if (window.Alpine) register();
  else document.addEventListener("alpine:init", register, { once: true });
})();
