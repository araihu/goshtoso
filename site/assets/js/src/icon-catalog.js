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
  // function so the visible source cannot drift from the clipboard source.
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
      "package main",
      "",
      "import (",
      '    "context"',
      '    "log"',
      '    "os"',
      '',
      '    "github.com/araihu/goshtoso/components/icon"',
      '    "github.com/araihu/goshtoso/components/icon/heroicons"',
      ")",
      "",
      "func main() {",
      "    if err := icon.Icon(icon.Config{",
    ].concat(fields.map(function (field) { return "        " + field + ","; }), [
      "    }).Render(context.Background(), os.Stdout); err != nil {",
      "        log.Fatal(err)",
      "    }",
      "}",
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
        size: "",
        label: "",
        decorative: false,
        rootClass: "",
        copied: false,
        copyAnnouncement: "",
        lastTrigger: null,
        _copyTimer: 0,
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
          this.copied = false;
          this.copyAnnouncement = "";
          this.open = true;
          this.$nextTick(function () {
            var control = document.getElementById("icon-size");
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
          return classes[this.size] + (root ? " " + root : "");
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
        copyCode: function () {
          var component = this;
          if (!navigator.clipboard || !navigator.clipboard.writeText) {
            this.copyAnnouncement = "Clipboard unavailable; select the code to copy it.";
            return;
          }
          navigator.clipboard.writeText(this.code).then(function () {
            component.copied = true;
            component.copyAnnouncement = "Compilable Go source copied to clipboard.";
            if (component._copyTimer) window.clearTimeout(component._copyTimer);
            component._copyTimer = window.setTimeout(function () {
              component.copied = false;
              component._copyTimer = 0;
            }, 2000);
          }).catch(function () {
            component.copyAnnouncement = "Could not copy code; select it manually.";
          });
        },
        destroy: function () {
          if (this._copyTimer) window.clearTimeout(this._copyTimer);
          this._copyTimer = 0;
        },
      };
    });
  }

  if (window.Alpine) register();
  else document.addEventListener("alpine:init", register, { once: true });
})();
