// Palette Alpine factory. Swatch names are JSON data, never executable markup.
(function () {
  if (window.goshtosoPalette) return;

  window.goshtosoPalette = function (root) {
    var config = window.goshtosoParseData(root.dataset.paletteSwatches, { hues: [], shades: [], hideNeutral: false });
    return {
      hovered: "",
      selectedHex: "#000000",
      hexInput: "",
      hexInvalid: false,
      swatchHues: config.hues || [],
      swatchShades: config.shades || [],
      hideNeutral: Boolean(config.hideNeutral),
      pick: function (value, element) {
        this.syncHex(value, element);
        this.$dispatch("select-close", value);
      },
      commitHex: function (value) {
        var hex = this.normalizeHex(value);
        if (!hex) {
          this.hexInvalid = true;
          return;
        }
        this.hexInvalid = false;
        this.selectedHex = hex;
        this.hexInput = hex;
        this.$dispatch("select-close", hex);
      },
      previewHex: function (value) {
        this.hexInput = value;
        var hex = this.normalizeHex(value);
        this.hexInvalid = this.hexInput !== "" && !hex;
        if (hex) {
          this.selectedHex = hex;
          this.hexInput = hex;
          this.$dispatch("select-close", hex);
        }
      },
      syncHex: function (value, element) {
        this.hexInvalid = false;
        if (!value) {
          this.selectedHex = "#000000";
          this.hexInput = "";
          return;
        }
        if (value[0] === "#") {
          var hex = this.normalizeHex(value);
          if (hex) {
            this.selectedHex = hex;
            this.hexInput = hex;
          }
          return;
        }
        if (value === "white" || value === "black") {
          this.selectedHex = value === "white" ? "#ffffff" : "#000000";
          this.hexInput = this.selectedHex;
          return;
        }
        if (element) {
          var resolved = this.colorToHex(getComputedStyle(element).backgroundColor);
          this.selectedHex = resolved;
          this.hexInput = resolved;
        }
      },
      normalizeHex: function (value) {
        var raw = String(value || "").trim();
        var short = raw.match(/^#([0-9a-fA-F]{3})$/);
        if (short) return "#" + short[1].split("").map(function (character) { return character + character; }).join("").toLowerCase();
        var full = raw.match(/^#([0-9a-fA-F]{6})$/);
        return full ? "#" + full[1].toLowerCase() : "";
      },
      colorToHex: function (color) {
        if (!this.context) {
          var canvas = document.createElement("canvas");
          canvas.width = 1;
          canvas.height = 1;
          this.context = canvas.getContext("2d", { willReadFrequently: true });
        }
        this.context.clearRect(0, 0, 1, 1);
        this.context.fillStyle = "#000000";
        this.context.fillStyle = color;
        this.context.fillRect(0, 0, 1, 1);
        var channels = this.context.getImageData(0, 0, 1, 1).data;
        return "#" + [channels[0], channels[1], channels[2]].map(function (channel) { return channel.toString(16).padStart(2, "0"); }).join("");
      },
      escapeAttribute: function (value) {
        return String(value).replace(/&/g, "&amp;").replace(/"/g, "&quot;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
      },
      swatchButton: function (className, classes, style) {
        var safeClassName = this.escapeAttribute(className);
        var styleAttribute = style ? ' style="' + this.escapeAttribute(style) + '"' : "";
        return '<button type="button" data-cls="' + safeClassName + '" class="' + classes + '"' + styleAttribute + ' title="' + safeClassName + '"></button>';
      },
      swatchGridHTML: function () {
        var standard = "h-5 w-full rounded-sm border border-outline/30 dark:border-outline-dark/30 transition-transform hover:scale-125 hover:ring-2 hover:ring-primary focus:scale-125 dark:hover:ring-primary-dark";
        var neutral = "h-5 w-full rounded-sm border border-outline/60 transition-transform hover:scale-125 hover:ring-2 hover:ring-primary focus:scale-125 dark:border-outline-dark/60 dark:hover:ring-primary-dark";
        var html = "";
        if (!this.hideNeutral) {
          html += this.swatchButton("white", neutral + " bg-white", "");
          html += this.swatchButton("black", neutral + " bg-black", "");
        }
        this.swatchHues.forEach(function (hue) {
          this.swatchShades.forEach(function (shade) {
            var className = hue + "-" + shade;
            html += this.swatchButton(className, standard, "background-color: var(--color-" + className + ")");
          }, this);
        }, this);
        return html;
      },
      handleSwatchEvent: function (event, action) {
        var button = event.target.closest("button[data-cls]");
        if (!button || !event.currentTarget.contains(button)) return;
        if (action === "pick") {
          this.pick(button.dataset.cls, button);
          return;
        }
        this.hovered = button.dataset.cls;
      },
    };
  };
})();
