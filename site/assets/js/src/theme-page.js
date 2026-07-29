// theme-page.js — Alpine behavior for the theme workbench.
(function () {
  function readStoredValue(key, fallback) {
    try {
      return localStorage.getItem(key) || fallback;
    } catch (error) {
      return fallback;
    }
  }

  function readStoredObject(key) {
    try {
      var value = JSON.parse(localStorage.getItem(key) || "{}");
      if (!value || typeof value !== "object" || Array.isArray(value)) return {};
      var result = {};
      Object.keys(value).forEach(function (entry) {
        if (typeof value[entry] === "string") result[entry] = value[entry];
      });
      return result;
    } catch (error) {
      return {};
    }
  }

  function sanitizeOverrides(value, tokens) {
    var allowed = new Set(tokens);
    var result = {};
    Object.entries(value).forEach(function (entry) {
      if (allowed.has(entry[0]) && typeof entry[1] === "string") result[entry[0]] = entry[1];
    });
    return result;
  }

  function storeValue(key, value) {
    try {
      localStorage.setItem(key, value);
    } catch (error) {}
  }

  function readPageData(root) {
    var element = root && root.querySelector("#theme-page-data");
    if (!element) return {};
    try {
      return JSON.parse(element.textContent || "{}");
    } catch (error) {
      return {};
    }
  }

  function register() {
    if (!window.Alpine || Alpine.__themePageRegistered) return;
    Alpine.__themePageRegistered = true;
    Alpine.data("themePage", function () {
      return {
        titleFont: readStoredValue("themeTitleFont", ""),
        bodyFont: readStoredValue("themeBodyFont", ""),
        radius: readStoredValue("themeRadius", ""),
        overrides: readStoredObject("themeOverrides"),
        resolved: {},
        cssMode: readStoredValue("themeCssMode", "single"),
        cssFilter: readStoredValue("themeCssFilter", "current"),
        contrastTab: "colors",
        contrastBase: "primary",
        contrastCustomFg: "#0f172a",
        contrastCustomBg: "#ffffff",
        copied: false,
        confirmingReset: false,
        resetDone: false,
        radiusMap: {},
        googleFontMap: {},
        allThemes: [],
        themeLabels: {},
        themeClassMap: {},
        blocks: {},
        cssOrder: [],
        allTokens: [],
        tokenLabels: {},
        _ctx: null,
        _themeObserver: null,
        _copyTimer: null,
        _resetTimer: null,
        init: function () {
          var data = readPageData(this.$root);
          this.radiusMap = data.radiusMap || {};
          this.googleFontMap = data.googleFontMap || {};
          this.allThemes = Array.isArray(data.allThemes) ? data.allThemes : [];
          this.themeLabels = data.themeLabels || {};
          this.themeClassMap = data.themeClassMap || {};
          this.blocks = data.blocks || {};
          this.cssOrder = Array.isArray(data.cssOrder) ? data.cssOrder : [];
          this.allTokens = Array.isArray(data.allTokens) ? data.allTokens : [];
          this.tokenLabels = data.tokenLabels || {};
          this.overrides = sanitizeOverrides(this.overrides, this.allTokens);

          try {
            this.applyAll();
          } catch (error) {
            console.error("themePage applyAll", error);
          }
          try {
            this.refreshResolved();
          } catch (error) {
            console.error("themePage refreshResolved", error);
          }

          var component = this;
          this.$watch("titleFont", function (value) {
            storeValue("themeTitleFont", value);
            component.applyFont("--font-title", value);
          });
          this.$watch("bodyFont", function (value) {
            storeValue("themeBodyFont", value);
            component.applyFont("--font-body", value);
          });
          this.$watch("radius", function (value) {
            storeValue("themeRadius", value);
            component.applyRadius(value);
          });
          this.$watch(
            "overrides",
            function (value) {
              storeValue("themeOverrides", JSON.stringify(value));
              component.applyColors();
            },
            { deep: true },
          );
          this.$watch("cssMode", function (value) {
            storeValue("themeCssMode", value);
          });
          this.$watch("cssFilter", function (value) {
            storeValue("themeCssFilter", value);
          });

          this._themeObserver = new MutationObserver(function () {
            component.applyAll();
            component.refreshResolved();
          });
          this._themeObserver.observe(document.documentElement, {
            attributes: true,
            attributeFilter: ["class", "data-theme"],
          });
        },
        destroy: function () {
          if (this._themeObserver) this._themeObserver.disconnect();
          if (this._copyTimer) clearTimeout(this._copyTimer);
          if (this._resetTimer) clearTimeout(this._resetTimer);
        },
        applyAll: function () {
          this.applyFont("--font-title", this.titleFont);
          this.applyFont("--font-body", this.bodyFont);
          this.applyRadius(this.radius);
          this.applyColors();
        },
        applyFont: function (variableName, label) {
          if (!label) {
            document.documentElement.style.removeProperty(variableName);
            return;
          }
          var googleFont = this.googleFontMap[label];
          if (googleFont) this.loadGoogleFont(label, googleFont);
          document.documentElement.style.setProperty(variableName, "'" + label + "', sans-serif");
        },
        applyRadius: function (key) {
          if (!key) {
            document.documentElement.style.removeProperty("--radius-radius");
            return;
          }
          var value = this.radiusMap[key];
          if (value != null) document.documentElement.style.setProperty("--radius-radius", value);
        },
        applyColors: function () {
          Object.entries(this.overrides).forEach(function (entry) {
            var key = entry[0];
            var value = entry[1];
            if (!value) {
              document.documentElement.style.removeProperty("--color-" + key);
              return;
            }
            var cssValue = value.startsWith("#") ? value : "var(--color-" + value + ")";
            document.documentElement.style.setProperty("--color-" + key, cssValue);
          });
        },
        setColor: function (token, value) {
          var next = Object.assign({}, this.overrides);
          next[token] = value;
          this.overrides = next;
        },
        pickColor: function (token, className) {
          var component = this;
          if (!className) this.clearColor(token);
          else this.setColor(token, className);
          requestAnimationFrame(function () {
            component.refreshResolved();
          });
        },
        currentClass: function (token) {
          var override = this.overrides[token];
          if (override) return override;
          var themeKey = this.theme || "goshtoso";
          return (this.themeClassMap[themeKey] || {})[token] || "";
        },
        classLabel: function (token) {
          var className = this.currentClass(token);
          if (!className) return "—";
          if (className.startsWith("#")) return className;
          var separator = className.indexOf("-");
          if (separator === -1) return className.charAt(0).toUpperCase() + className.slice(1);
          return className.charAt(0).toUpperCase() + className.slice(1, separator) + className.slice(separator);
        },
        clearColor: function (token) {
          document.documentElement.style.removeProperty("--color-" + token);
          var next = Object.assign({}, this.overrides);
          delete next[token];
          this.overrides = next;
        },
        resetAll: function () {
          var component = this;
          this.titleFont = "";
          this.bodyFont = "";
          this.radius = "";
          Object.keys(this.overrides).forEach(function (key) {
            document.documentElement.style.removeProperty("--color-" + key);
          });
          this.overrides = {};
          requestAnimationFrame(function () {
            component.refreshResolved();
          });
          this.confirmingReset = false;
          this.resetDone = true;
          if (this._resetTimer) clearTimeout(this._resetTimer);
          this._resetTimer = setTimeout(function () {
            component.resetDone = false;
          }, 2500);
        },
        loadGoogleFont: function (label, query) {
          var id = "gfont-" + label.replace(/\s+/g, "-");
          if (document.getElementById(id)) return;
          var link = document.createElement("link");
          link.id = id;
          link.rel = "stylesheet";
          link.href =
            "https://fonts.googleapis.com/css2?family=" +
            query +
            ":wght@400;500;600;700&display=swap";
          document.head.appendChild(link);
        },
        refreshResolved: function () {
          var component = this;
          var styles = getComputedStyle(document.documentElement);
          var next = {};
          this.allTokens.forEach(function (token) {
            var raw = styles.getPropertyValue("--color-" + token).trim();
            next[token] = component.normalizeColor(raw);
          });
          this.resolved = next;
        },
        normalizeColor: function (raw) {
          if (!raw) return "#000000";
          if (raw.startsWith("#")) {
            return raw.length === 4
              ? "#" +
                  raw
                    .slice(1)
                    .split("")
                    .map(function (character) {
                      return character + character;
                    })
                    .join("")
              : raw;
          }
          if (!this._ctx) {
            var canvas = document.createElement("canvas");
            canvas.width = 1;
            canvas.height = 1;
            this._ctx = canvas.getContext("2d", { willReadFrequently: true });
          }
          var context = this._ctx;
          context.clearRect(0, 0, 1, 1);
          context.fillStyle = "#000000";
          context.fillStyle = raw;
          context.fillRect(0, 0, 1, 1);
          var color = context.getImageData(0, 0, 1, 1).data;
          return (
            "#" +
            [color[0], color[1], color[2]]
              .map(function (channel) {
                return channel.toString(16).padStart(2, "0");
              })
              .join("")
          );
        },
        relLum: function (hex) {
          var red = parseInt(hex.slice(1, 3), 16) / 255;
          var green = parseInt(hex.slice(3, 5), 16) / 255;
          var blue = parseInt(hex.slice(5, 7), 16) / 255;
          var channel = function (value) {
            return value <= 0.03928 ? value / 12.92 : Math.pow((value + 0.055) / 1.055, 2.4);
          };
          return 0.2126 * channel(red) + 0.7152 * channel(green) + 0.0722 * channel(blue);
        },
        ratio: function (first, second) {
          var firstLum = this.relLum(first);
          var secondLum = this.relLum(second);
          var high = firstLum > secondLum ? firstLum : secondLum;
          var low = firstLum > secondLum ? secondLum : firstLum;
          return (high + 0.05) / (low + 0.05);
        },
        grade: function (ratio) {
          if (ratio >= 7) return "AAA";
          if (ratio >= 4.5) return "AA";
          if (ratio >= 3) return "AA Large";
          return "Fail";
        },
        gradeClass: function (ratio) {
          if (ratio >= 4.5) return "bg-green-500/15 text-green-700 dark:text-green-300";
          if (ratio >= 3) return "bg-amber-500/15 text-amber-700 dark:text-amber-300";
          return "bg-red-500/15 text-red-700 dark:text-red-300";
        },
        cssCode: function () {
          var filterKey = this.cssFilter === "current" ? this.theme || "goshtoso" : this.cssFilter;
          if (this.cssMode === "single") {
            return this.blocks[filterKey] || "/* theme " + filterKey + " not found */";
          }
          var keys = filterKey === "all" ? this.cssOrder : [filterKey];
          var component = this;
          var inner = keys
            .map(function (key) {
              return "    " + (component.blocks[key] || "").split("\n").join("\n    ");
            })
            .join("\n\n");
          return "@layer base {\n" + inner + "\n}\n";
        },
        cssExportLabel: function () {
          var filterKey = this.cssFilter === "current" ? this.theme || "goshtoso" : this.cssFilter;
          var label = filterKey === "all" ? "All themes" : this.themeLabels[filterKey] || filterKey;
          return label + (this.cssMode === "single" ? " — single" : " — @layer base");
        },
        copyCSS: function () {
          var component = this;
          navigator.clipboard.writeText(this.cssCode()).then(function () {
            component.copied = true;
            if (component._copyTimer) clearTimeout(component._copyTimer);
            component._copyTimer = setTimeout(function () {
              component.copied = false;
            }, 2000);
          });
        },
        contrastPairs: function () {
          var component = this;
          var base = this.resolved[this.contrastBase];
          if (!base) return [];
          return this.allTokens
            .filter(function (token) {
              return token !== component.contrastBase;
            })
            .map(function (token) {
              var color = component.resolved[token];
              if (!color) return null;
              var ratio = component.ratio(base, color);
              return { token: token, color: color, ratio: ratio, grade: component.grade(ratio) };
            })
            .filter(Boolean)
            .sort(function (first, second) {
              return second.ratio - first.ratio;
            });
        },
        base: function () {
          return this.resolved[this.contrastBase] || "#000000";
        },
        contrastMatrix: function () {
          var component = this;
          var baseColor = this.resolved[this.contrastBase];
          if (!baseColor) return [];
          return this.allTokens
            .filter(function (token) {
              return token !== component.contrastBase;
            })
            .map(function (token) {
              var color = component.resolved[token];
              if (!color) return null;
              var ratio = component.ratio(baseColor, color);
              return {
                token: token,
                label: component.tokenLabels[token] || token,
                color: color,
                ratio: ratio,
              };
            })
            .filter(Boolean);
        },
        customContrast: function () {
          var ratio = this.ratio(this.contrastCustomFg, this.contrastCustomBg);
          return { ratio: ratio, grade: this.grade(ratio) };
        },
      };
    });
  }

  if (window.Alpine) register();
  else document.addEventListener("alpine:init", register);
})();
