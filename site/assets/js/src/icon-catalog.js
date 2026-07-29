// icon-catalog.js — local state for the generated Heroicons workbench.
(function () {
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
          var fields = [
            "SpriteURL: heroicons.SpriteURL",
            "Symbol:    heroicons." + this.selected.name,
          ];
          var sizes = { xs: "SizeXS", sm: "SizeSM", lg: "SizeLG", xl: "SizeXL" };
          var label = String(this.label || "").trim();
          if (sizes[this.size]) fields.push("Size:      icon." + sizes[this.size]);
          if (this.decorative) fields.push("Decorative: true");
          else if (label) fields.push("Label:     " + JSON.stringify(label));
          var root = String(this.rootClass || "").trim();
          if (root) fields.push("RootClass: " + JSON.stringify(root));
          return "@icon.Icon(icon.Config{\n" + fields.map(function (field) { return "    " + field + ","; }).join("\n") + "\n})";
        },
        copyCode: function () {
          var component = this;
          if (!navigator.clipboard || !navigator.clipboard.writeText) {
            this.copyAnnouncement = "Clipboard unavailable; select the code to copy it.";
            return;
          }
          navigator.clipboard.writeText(this.code).then(function () {
            component.copied = true;
            component.copyAnnouncement = "Go code copied to clipboard.";
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
