(() => {
  const copy = (value) => JSON.parse(JSON.stringify(value));
  const only = (object, keys) => object && typeof object === "object" && Object.keys(object).length === keys.length && keys.every((k) => Object.hasOwn(object, k));
  const finite = (n) => typeof n === "number" && Number.isFinite(n);
  const colorValid = (c) => only(c, ["space", "channels", "alpha"]) && ["srgb", "oklch"].includes(c.space) && Array.isArray(c.channels) && c.channels.length === 3 && c.channels.every(finite) && finite(c.alpha) && c.alpha >= 0 && c.alpha <= 1 && (c.space === "srgb" ? c.channels.every((n) => n >= 0 && n <= 1) : c.channels[0] >= 0 && c.channels[0] <= 1 && c.channels[1] >= 0 && c.channels[1] <= 0.5 && c.channels[2] >= 0 && c.channels[2] < 360);
  const byteLength = (s) => new TextEncoder().encode(s).length;
  const fontValid = (f) => only(f, ["source", "id"]) && ["system", "google", "bundled", "upload"].includes(f.source) && /^[A-Za-z0-9][A-Za-z0-9:._/-]{0,199}$/.test(f.id);
  function valid(d, keys) {
    return only(d, ["schemaVersion", "tokens", "typography", "radiusRem", "provenance"]) && only(d.typography, ["title", "body"]) && only(d.provenance, ["presetKey", "goshtosoVersion", "catalogRevision"]) && Object.entries({presetKey: 100, goshtosoVersion: 200, catalogRevision: 128}).every(([k, max]) => typeof d.provenance[k] === "string" && byteLength(d.provenance[k]) <= max) && d.schemaVersion === 1 && finite(d.radiusRem) && d.radiusRem >= 0 && d.radiusRem <= 4 && fontValid(d.typography?.title) && fontValid(d.typography?.body) && d.tokens && Object.keys(d.tokens).length === keys.length && keys.length <= 256 && keys.every((k) => colorValid(d.tokens[k])) && byteLength(JSON.stringify(d)) <= 65536;
  }
  const colorCSS = (c) => c.space === "oklch" ? `oklch(${c.channels.join(" ")} / ${c.alpha})` : `color(srgb ${c.channels.join(" ")} / ${c.alpha})`;
  function parseColor(text) {
    text = text.trim();
    let match = text.match(/^#([0-9a-f]{3,4}|[0-9a-f]{6}|[0-9a-f]{8})$/i);
    if (match) {
      let hex = match[1];
      if (hex.length < 5) hex = [...hex].map((x) => x + x).join("");
      return { space: "srgb", channels: [0, 2, 4].map((i) => parseInt(hex.slice(i, i + 2), 16) / 255), alpha: hex.length === 8 ? parseInt(hex.slice(6, 8), 16) / 255 : 1 };
    }
    match = text.match(/^(oklch\(|color\(srgb\s+)([^()]+)\)$/);
    if (!match) throw Error("Use hex, color(srgb \u2026), or oklch(\u2026).");
    const fields = match[2].trim().split(/\s+/);
    if (fields.length !== 3 && (fields.length !== 5 || fields[3] !== "/")) throw Error("Invalid color.");
    const number = (s) => s.endsWith("%") ? Number(s.slice(0, -1)) / 100 : Number(s);
    const c = { space: match[1] === "oklch(" ? "oklch" : "srgb", channels: fields.slice(0, 3).map(number), alpha: fields.length === 5 ? number(fields[4]) : 1 };
    if (!colorValid(c)) throw Error("Color outside supported bounds.");
    return c;
  }
  function paletteColor(value, baseline, style) {
    if (typeof value !== "string") return null;
    if (!value) return copy(baseline);
    if (value === "white") return parseColor("#ffffff");
    if (value === "black") return parseColor("#000000");
    try {
      if (value.startsWith("#")) return parseColor(value);
      if (!/^[a-z]+-(50|[1-9]00|950)$/.test(value)) return null;
      return parseColor(style.getPropertyValue("--color-" + value).trim());
    } catch {
      return null;
    }
  }
  function rgb(c) {
    if (c.space === "srgb") return c.channels;
    const [L, C, H] = c.channels, h = H * Math.PI / 180, a = C * Math.cos(h), b = C * Math.sin(h);
    const l = (L + 0.3963377774 * a + 0.2158037573 * b) ** 3, m = (L - 0.1055613458 * a - 0.0638541728 * b) ** 3, s = (L - 0.0894841775 * a - 1.291485548 * b) ** 3;
    return [4.0767416621 * l - 3.3077115913 * m + 0.2309699292 * s, -1.2684380046 * l + 2.6097574011 * m - 0.3413193965 * s, -0.0041960863 * l - 0.7034186147 * m + 1.707614701 * s].map((v) => v <= 31308e-7 ? 12.92 * v : 1.055 * Math.pow(v, 1 / 2.4) - 0.055);
  }
  const clamp = (v) => Math.min(1, Math.max(0, v));
  function derive(d, specs, base, changedKey) {
    const derived = new Map(specs.filter((s) => !s.editable).map((s) => [s.key, s.expression]));
    const get = (key, seen = [], tokens = d.tokens, expressions = derived) => {
      if (seen.includes(key)) throw Error("Cyclic derived token.");
      const expression = expressions.get(key);
      if (!expression) return tokens[key];
      const alias = expression.match(/^var\(--color-([\w-]+)\)$/);
      if (alias) return get(alias[1], [...seen, key], tokens, expressions);
      const match = expression.match(/^color-mix\(in srgb, var\(--color-([\w-]+)\) ([\d.]+)%, var\(--color-([\w-]+)\) ([\d.]+)%\)$/);
      if (!match) throw Error("Unsupported derived token.");
      const a = get(match[1], [...seen, key], tokens, expressions), b = get(match[3], [...seen, key], tokens, expressions), p = Number(match[2]) / 100, alpha = a.alpha * p + b.alpha * (1 - p), av = rgb(a), bv = rgb(b);
      return { space: "srgb", channels: av.map((v, i) => alpha ? clamp((v * a.alpha * p + bv[i] * b.alpha * (1 - p)) / alpha) : 0), alpha };
    };
    // Keep preset overrides and imported literals independent of default derivations.
    for (const [key, expression] of derived) {
      const expected = get(key, [], base.tokens, new Map([[key, expression]])), actual = base.tokens[key];
      if (expected.space !== actual.space || Math.abs(expected.alpha - actual.alpha) > 1e-9 || expected.channels.some((v, i) => Math.abs(v - actual.channels[i]) > 1e-9)) derived.delete(key);
    }
    const affected = new Set([changedKey]);
    for (let i = 0; i < derived.size; i++) {
      for (const [key, expression] of derived) {
        if ([...affected].some((input) => expression.includes("var(--color-" + input + ")"))) affected.add(key);
      }
    }
    for (const key of derived.keys()) if (!affected.has(key)) derived.delete(key);
    for (const key of derived.keys()) d.tokens[key] = get(key);
  }
  const safeURL = (value) => {
    try {
      const u = new URL(value, location.origin);
      return !/[<>"'\\\r\n\0]/.test(value) && !u.username && !u.password && (value.startsWith("https://") && u.protocol === "https:" || value.startsWith("/") && !value.startsWith("//") && u.origin === location.origin) && !u.hash;
    } catch {
      return false;
    }
  };
  function validWeight(value) {
    if (value === "normal") return true;
    if (!/^[0-9]+(?:\.[0-9]+)?(?: [0-9]+(?:\.[0-9]+)?)?$/.test(value)) return false;
    const weights = value.split(" ").map(Number);
    return weights.every((n) => Number.isFinite(n) && n >= 1 && n <= 1e3) && (weights.length === 1 || weights[0] <= weights[1]);
  }
  function fontAssets(ref, fonts) {
    const font = fonts.find((f) => f.source === ref.source && f.id === ref.id);
    if (!font || (!font.url || font.source === "google") && (!Array.isArray(font.families) || !font.families.length || font.families.some((name) => typeof name !== "string" || !name || name.length > 200 || /[<>\r\n\0]/.test(name)))) throw Error("Font unavailable: " + ref.id);
    let stylesheet = "", face = "";
    let family = (font.url ? [] : font.families).map((name) => /^(serif|sans-serif|monospace|system-ui|ui-serif|ui-sans-serif|ui-monospace|cursive|fantasy)$/.test(name) ? name : JSON.stringify(name)).join(",");
    if (font.source === "google") stylesheet = "https://fonts.googleapis.com/css2?" + new URLSearchParams({ family: font.families[0], display: "swap" });
    if (font.url) {
      const weight = font.weight === "unknown" ? "normal" : font.weight, style = font.style === "unknown" ? "normal" : font.style;
      if (!safeURL(font.url) || !["woff2", "woff", "truetype", "opentype"].includes(font.format) || !validWeight(weight) || !["normal", "italic", "oblique"].includes(style)) throw Error("Invalid font catalog.");
      const alias = "goshtoso-font-" + ref.id.replace(/[:/.]/g, "-");
      family = JSON.stringify(alias) + ",system-ui,sans-serif";
      face = `@font-face{font-family:${JSON.stringify(alias)};src:url(${JSON.stringify(font.url)}) format(${JSON.stringify(font.format)});font-weight:${weight};font-style:${style};font-display:swap;}`;
    }
    return { font, family, face, stylesheet };
  }
  const googlePreviewAlias = (ref) => "goshtoso-preview-google-" + ref.id.replace(/[:/.]/g, "-");
  function editorFontFamily(ref, fonts) {
    return ref.source === "google" ? JSON.stringify(googlePreviewAlias(ref)) + ",system-ui,sans-serif" : fontAssets(ref, fonts).family;
  }
  async function loadGooglePreview(ref, fonts, signal) {
    if (Array.from(document.fonts).some((face) => face.family === googlePreviewAlias(ref))) return;
    const assets = fontAssets(ref, fonts), response = await fetch(assets.stylesheet, { signal });
    if (!response.ok) throw Error("Google font preview unavailable.");
    const sheet = await response.text(), faces = [];
    for (const match of sheet.matchAll(/@font-face\s*\{([^{}]*)\}/g)) {
      const block = match[1], source = block.match(/src:\s*url\(([^)]+)\)/), range = block.match(/unicode-range:\s*([^;]+);/), weight = block.match(/font-weight:\s*([^;]+);/), style = block.match(/font-style:\s*([^;]+);/);
      if (!source) continue;
      const url = new URL(source[1].replace(/["']/g, ""));
      if (url.protocol !== "https:" || url.hostname !== "fonts.gstatic.com" || url.username || url.password) throw Error("Invalid Google font source.");
      const descriptors = { display: "swap" };
      if (range) descriptors.unicodeRange = range[1];
      if (weight && validWeight(weight[1].trim())) descriptors.weight = weight[1].trim();
      if (style && ["normal", "italic", "oblique"].includes(style[1].trim())) descriptors.style = style[1].trim();
      const face = new FontFace(googlePreviewAlias(ref), "url(" + JSON.stringify(url.href) + ")", descriptors);
      faces.push(face);
    }
    if (!faces.length) throw Error("Google font preview unavailable.");
    for (const face of faces) document.fonts.add(face);
  }
  function css(d, fonts) {
    const title = fontAssets(d.typography.title, fonts), body = fontAssets(d.typography.body, fonts), assets = [title, body];
    return "/* Goshtoso theme fonts: " + JSON.stringify(d.typography) + " */\n" + assets.filter((f) => f.stylesheet).map((f) => "@import url(" + JSON.stringify(f.stylesheet) + ");").join("\n") + "\n" + assets.map((f) => f.face).join("\n") + '\n:root[data-theme="custom"]{' + Object.entries(d.tokens).map(([key, value]) => "--color-" + key + ":" + colorCSS(value) + ";").join("") + `--font-title:${title.family};--font-body:${body.family};--font-paragraph:${body.family};--radius-radius:${d.radiusRem}rem;color-scheme:light;}
:root[data-theme="custom"].dark{color-scheme:dark;}`;
  }
  function importThemeCSS(text, base, fonts, specs) {
    if (typeof text !== "string" || !text.trim() || new TextEncoder().encode(text).length > 1048576) throw Error("Choose a theme CSS file up to 1 MiB.");
    const sheet = new CSSStyleSheet();
    sheet.replaceSync(text);
    const variables = /* @__PURE__ */ new Map(), themeNames = /* @__PURE__ */ new Set();
    const collect = (rules) => {
      for (const rule of rules) {
        if (rule.type === CSSRule.STYLE_RULE) {
          const properties = Array.from(rule.style).filter((key) => key.startsWith("--"));
          if (!properties.length) continue;
          const selectors = rule.selectorText.split(",");
          for (const selector of selectors) {
            const match = selector.trim().match(/^(?::root)?(?:\[data-theme=["']?([a-z0-9-]+)["']?\])?$/);
            if (!match || !selector.trim()) throw Error("Theme variables must use :root or [data-theme] selectors.");
            if (match[1]) themeNames.add(match[1]);
          }
          if (themeNames.size > 1) throw Error("Import one theme at a time. Choose a single-theme CSS export.");
          for (const key of properties) {
            if (rule.style.getPropertyPriority(key)) throw Error("Theme variables cannot use !important.");
            variables.set(key, rule.style.getPropertyValue(key).trim());
          }
        } else if (rule.type === CSSRule.FONT_FACE_RULE) {
        } else if (rule.cssRules && rule.constructor.name === "CSSLayerBlockRule") collect(rule.cssRules);
        else throw Error("Conditional CSS is not supported. Use a Goshtoso theme export.");
      }
    };
    collect(sheet.cssRules);
    const variable = (value, seen = []) => {
      const match = value?.match(/^var\((--[\w-]+)\)$/);
      if (!match) return value;
      if (seen.includes(match[1]) || seen.length > 64 || !variables.has(match[1])) throw Error("Unresolved or cyclic theme variable: " + match[1]);
      return variable(variables.get(match[1]), [...seen, match[1]]);
    };
    const definition = copy(base);
    for (const spec of specs) {
      const value = variable(variables.get("--color-" + spec.key));
      if (!value) throw Error("Missing theme variable: --color-" + spec.key);
      try {
        definition.tokens[spec.key] = parseColor(value);
      } catch {
        throw Error("Unsupported color for --color-" + spec.key + ". Use a Goshtoso CSS export with literal colors.");
      }
    }
    const radius = variable(variables.get("--radius-radius"));
    if (!/^(?:0|(?:\d+(?:\.\d+)?|\.\d+)rem)$/.test(radius || "")) throw Error("The theme radius must be expressed in rem.");
    definition.radiusRem = Number(radius === "0" ? radius : radius.slice(0, -3));
    const family = (value) => {
      const style = document.createElement("span").style;
      style.fontFamily = value || "";
      return style.fontFamily.toLowerCase().replace(/["']/g, "").replace(/\s+/g, " ").trim();
    };
    const metadata = text.match(/\/\* Goshtoso theme fonts: ([^\r\n]*?) \*\//);
    const fontReferences = metadata ? JSON.parse(metadata[1]) : null;
    if (fontReferences && (!only(fontReferences, ["title", "body"]) || !fontValid(fontReferences.title) || !fontValid(fontReferences.body))) throw Error("Invalid font references in the CSS export.");
    for (const role of ["title", "body"]) {
      const wanted = family(variable(variables.get("--font-" + role)));
      if (!wanted) throw Error("Missing or invalid --font-" + role + ".");
      const matching = fonts.filter((font) => {
        try {
          return family(fontAssets({ source: font.source, id: font.id }, fonts).family) === wanted;
        } catch {
          return false;
        }
      });
      const reference = fontReferences?.[role];
      const found = reference ? matching.find((font) => font.source === reference.source && font.id === reference.id) : matching.find((font) => font.source === "google" && text.includes(fontAssets({ source: font.source, id: font.id }, fonts).stylesheet)) || matching.find((font) => font.source === "system") || matching[0];
      if (!found) throw Error("The " + role + " font is not available. Add the font before importing this CSS.");
      definition.typography[role] = { source: found.source, id: found.id };
    }
    definition.provenance = { presetKey: "", goshtosoVersion: base.provenance.goshtosoVersion, catalogRevision: base.provenance.catalogRevision };
    if (!valid(definition, specs.map((spec) => spec.key))) throw Error("The CSS contains an invalid theme value.");
    return definition;
  }
  function contrast(a, b) {
    const lum = (c) => rgb(c).map(clamp).map((v) => v <= 0.04045 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4).reduce((n, v, i) => n + v * [0.2126, 0.7152, 0.0722][i], 0);
    const x = lum(a), y = lum(b);
    return (Math.max(x, y) + 0.05) / (Math.min(x, y) + 0.05);
  }
  function editor(root) {
    const config = JSON.parse(root.dataset.themeEditor), abort = new AbortController(), opts = { signal: abort.signal };
    const invalidInputs = /* @__PURE__ */ new Set();
    let definition = copy(config.initial), baseline = copy(definition), mode = config.previewMode === "dark" ? "dark" : "light", fonts = config.fonts || [], dirty = false, sequence = 0;
    const specs = config.tokens, keys = specs.map((t) => t.key), hidden = root.querySelector("[data-theme-definition]"), feedback = root.querySelector("[data-theme-feedback]");
    const emit = (type, detail) => root.dispatchEvent(new CustomEvent("themeeditor:" + type, { bubbles: true, detail }));
    const fontLoads = /* @__PURE__ */ new Map();
    function loadFontPreview(value) {
      const [source, ...id] = value.split("/"), ref = { source, id: id.join("/") };
      if (source === "system" || fontLoads.has(value)) return;
      const load = async () => {
        if (source === "google") await loadGooglePreview(ref, fonts, abort.signal);
        const family = source === "google" ? JSON.stringify(googlePreviewAlias(ref)) : fontAssets(ref, fonts).family.split(",")[0];
        const faces = await document.fonts.load("16px " + family);
        if (!faces.length) throw Error("Font preview unavailable.");
      };
      fontLoads.set(value, load().catch((error) => {
        if (!abort.signal.aborted) feedback.textContent = error.message + " Using system fallback for this preview.";
      }));
    }
    const fontObserver = typeof IntersectionObserver === "undefined" ? null : new IntersectionObserver((entries) => {
      for (const entry of entries) if (entry.isIntersecting) loadFontPreview(entry.target.dataset.themeFontOption);
    });
    function observeFonts() {
      fontObserver?.disconnect();
      document.querySelectorAll("[data-theme-font-option]").forEach((option) => {
        if (root.contains(option) || option.dataset.themeEditorFor === config.id) fontObserver?.observe(option);
      });
    }
    function syncFontPickers(replaceOptions = false) {
      for (const role of ["title", "body"]) {
        const ref = definition.typography[role], picker = document.getElementById(config.id + "-" + role), value = ref.source + "/" + ref.id;
        const detail = { id: picker.id, values: [value] };
        if (replaceOptions) {
          detail.options = fonts.flatMap((font) => {
            try {
              const reference = { source: font.source, id: font.id };
              return [{ value: font.source + "/" + font.id, label: font.label + " \xB7 " + font.weight + " \xB7 " + font.style, labelAttrs: { style: "font-family:" + editorFontFamily(reference, fonts), "data-theme-font-option": font.source + "/" + font.id, title: font.label + " \xB7 " + font.weight + " \xB7 " + font.style } }];
            } catch {
              return [];
            }
          });
          if (!detail.options.some((option) => option.value === value)) detail.options.push({ value, label: "Unavailable: " + ref.id, disabled: true });
        }
        picker.dispatchEvent(new CustomEvent("combobox:load", { bubbles: true, detail }));
      }
      observeFonts();
    }
    function refreshFontFaces() {
      root.querySelector("[data-theme-font-faces]").textContent = fonts.flatMap((font) => {
        try {
          return [fontAssets({ source: font.source, id: font.id }, fonts).face];
        } catch {
          return [];
        }
      }).join("\n");
    }
    const dialog = document.getElementById(config.id + "-confirm");
    let confirmation = null, accepted = false;
    const closeConfirmation = () => window.dispatchEvent(new CustomEvent("modal:close", { detail: { id: dialog.id } }));
    dialog.querySelector("[data-theme-confirm-accept]").addEventListener("click", () => {
      accepted = true;
      closeConfirmation();
    }, opts);
    dialog.querySelector("[data-theme-confirm-cancel]").addEventListener("click", closeConfirmation, opts);
    dialog.addEventListener("close", () => {
      const resolve = confirmation;
      confirmation = null;
      resolve?.(accepted);
    }, opts);
    const confirmDraft = (message) => new Promise((resolve) => {
      if (confirmation) {
        resolve(false);
        return;
      }
      accepted = false;
      confirmation = resolve;
      dialog.querySelector("[data-theme-confirm-message]").textContent = message;
      window.dispatchEvent(new CustomEvent("modal:open", { detail: { id: dialog.id } }));
    });
    const preview = () => {
      const frame2 = document.getElementById(config.previewID);
      if (frame2?.contentWindow) frame2.contentWindow.postMessage({ type: "goshtoso:theme-state", version: 1, instanceId: config.previewID, sequence: ++sequence, definition: copy(definition), previewMode: mode }, location.origin);
    };
    function update(sync = false, event = "change") {
      const isValid = !invalidInputs.size && valid(definition, keys) && Object.values(definition.typography).every((ref) => fonts.some((font) => font.source === ref.source && font.id === ref.id));
      hidden.value = JSON.stringify(definition);
      root.querySelectorAll("[data-theme-mode]").forEach((input) => {
        input.checked = input.value === mode;
      });
      if (sync) {
        root.querySelectorAll("[data-theme-color]").forEach((el) => {
          el.value = colorCSS(definition.tokens[el.dataset.themeColor]);
          el.dispatchEvent(new Event("input", { bubbles: true }));
        });
        root.querySelector("[data-theme-radius]").value = definition.radiusRem;
        syncFontPickers();
      }
      for (const role of ["title", "body"]) {
        const ref = definition.typography[role], trigger = document.getElementById(config.id + "-" + role + "-trigger");
        try {
          trigger.style.fontFamily = editorFontFamily(ref, fonts);
          loadFontPreview(ref.source + "/" + ref.id);
        } catch {
          trigger.style.fontFamily = "system-ui";
          feedback.textContent = "Selected font unavailable. Choose another font before saving.";
        }
      }
      root.querySelectorAll("[data-theme-color-row]").forEach((el) => {
        const key = el.dataset.themeColorRow, dark = key.includes("-dark"), hasDark = specs.some((t) => t.key.includes("-dark") && t.key.replace("-dark", "") === key);
        el.hidden = dark ? mode !== "dark" : hasDark && mode === "dark";
      });
      (config.contrastID ? document.getElementById(config.contrastID) : root)?.querySelectorAll("[data-theme-contrast-pair]").forEach((panel) => {
        const bg = panel.dataset.themeContrastPair + (mode === "dark" ? "-dark" : ""), fg = "on-" + bg;
        const foreground = definition.tokens[fg], background = definition.tokens[bg];
        const opaque = foreground.alpha === 1 && background.alpha === 1;
        const ratio = contrast(foreground, background);
        const grade = !opaque ? "unrated" : ratio >= 7 ? "aaa" : ratio >= 4.5 ? "aa" : ratio >= 3 ? "aa-large" : "fail";
        panel.dataset.themeContrastRating = grade;
        panel.querySelector("[data-theme-contrast-ratio]").textContent = opaque ? ratio.toFixed(2) + ":1" : "Not rated";
        Object.assign(panel.querySelector("[data-theme-contrast-sample]").style, { color: colorCSS(foreground), backgroundColor: colorCSS(background) });
        Object.assign(panel.querySelector("[data-theme-contrast-reverse]").style, { color: colorCSS(background), backgroundColor: colorCSS(foreground) });
        panel.querySelectorAll("[data-theme-contrast-grade]").forEach((badge) => {
          badge.hidden = badge.dataset.themeContrastGrade !== grade;
        });
        panel.querySelector("[data-theme-contrast-transparency]").hidden = opaque;
      });
      if (isValid) preview();
      emit(event, { id: config.id, definition: copy(definition), previewMode: mode, dirty, valid: isValid });
    }
    const updateFromControl = async (e) => {
      try {
        const el = e.target;
        feedback.textContent = "";
        if (el.hasAttribute("data-theme-color")) {
          const before = copy(definition);
          definition.tokens[el.dataset.themeColor] = parseColor(el.value);
          derive(definition, specs, before, el.dataset.themeColor);
          dirty = true;
        } else if (el.hasAttribute("data-theme-radius")) {
          definition.radiusRem = Number(el.value);
          dirty = true;
        } else if (el.hasAttribute("data-theme-mode")) mode = el.value;
        else if (el.hasAttribute("data-theme-preset")) {
          const p = config.catalog.find((p2) => p2.id === el.value);
          if (!p) return;
          if (dirty && !await confirmDraft("Discard unsaved theme changes?")) {
            el.value = "";
            el.dispatchEvent(new Event("input", { bubbles: true }));
            return;
          }
          definition = copy(p.definition);
          baseline = copy(definition);
          for (const el2 of invalidInputs) el2.removeAttribute("aria-invalid");
          invalidInputs.clear();
          dirty = true;
          update(true);
          return;
        } else return;
        invalidInputs.delete(el);
        el.removeAttribute("aria-invalid");
        update();
      } catch (err) {
        invalidInputs.add(e.target);
        e.target.setAttribute("aria-invalid", "true");
        dirty = true;
        feedback.textContent = err.message;
        emit("change", { id: config.id, definition: copy(definition), previewMode: mode, dirty: true, valid: false });
      }
    };
    root.addEventListener("change", updateFromControl, opts);
    root.addEventListener("input", (event) => {
      if (event.target.hasAttribute("data-theme-radius")) updateFromControl(event);
    }, opts);
    function focusColorTrigger(row) {
      requestAnimationFrame(() => {
        if (row.isConnected && (document.activeElement === document.body || row.contains(document.activeElement))) {
          row.querySelector("[data-theme-color-trigger]").focus({ preventScroll: true });
        }
      });
    }
    root.addEventListener("keydown", (event) => {
      if (event.key !== "Escape") return;
      const row = event.target.closest("[data-theme-color-row]");
      if (row) focusColorTrigger(row);
    }, opts);
    root.addEventListener("select-close", (event) => {
      const row = event.target.closest("[data-theme-color-row]");
      if (!row || !event.target.closest("[data-palette]")) return;
      const key = row.dataset.themeColorRow, input = row.querySelector("[data-theme-color]");
      const color = paletteColor(event.detail, baseline.tokens[key], getComputedStyle(root));
      if (!color) return;
      input.value = colorCSS(color);
      input.dispatchEvent(new Event("input", { bubbles: true }));
      input.dispatchEvent(new Event("change", { bubbles: true }));
      focusColorTrigger(row);
    }, opts);
    root.addEventListener("combobox:change", (event) => {
      const role = event.detail.id === config.id + "-title" ? "title" : event.detail.id === config.id + "-body" ? "body" : null;
      if (!role || event.detail.values.length !== 1) return;
      const [source, ...id] = event.detail.values[0].split("/");
      definition.typography[role] = { source, id: id.join("/") };
      dirty = true;
      update();
    }, opts);
    const importFile = root.querySelector("[data-theme-import-file]"), importFeedback = root.querySelector("[data-theme-import-feedback]");
    let importSequence = 0;
    importFile.addEventListener("change", async () => {
      const file = importFile.files[0];
      if (!file) return;
      const sequence2 = ++importSequence;
      importFeedback.hidden = true;
      importFile.removeAttribute("aria-invalid");
      importFeedback.classList.remove("text-danger");
      importFile.disabled = true;
      try {
        if (!/\.css$/i.test(file.name) || !file.size || file.size > 1048576) throw Error("Choose a .css theme file up to 1 MiB.");
        const candidate = importThemeCSS(await file.text(), definition, fonts, specs);
        if (abort.signal.aborted || sequence2 !== importSequence) return;
        if (dirty && !await confirmDraft("Replace unsaved theme changes with the imported CSS?")) return;
        if (abort.signal.aborted || sequence2 !== importSequence) return;
        definition = candidate;
        for (const input of invalidInputs) input.removeAttribute("aria-invalid");
        invalidInputs.clear();
        feedback.textContent = "";
        dirty = true;
        update(true);
        importFeedback.textContent = "CSS imported into the draft.";
        importFeedback.hidden = false;
      } catch (error) {
        if (abort.signal.aborted || sequence2 !== importSequence) return;
        importFile.setAttribute("aria-invalid", "true");
        importFeedback.textContent = error.message || "The CSS could not be imported. Your draft was preserved.";
        importFeedback.classList.add("text-danger");
        importFeedback.hidden = false;
        window.dispatchEvent(new CustomEvent("notify", { detail: { kind: "toast", tone: "danger", title: "Theme import failed", message: importFeedback.textContent } }));
      } finally {
        importFile.value = "";
        importFile.dispatchEvent(new Event("change", { bubbles: true }));
        importFile.disabled = false;
      }
    }, opts);
    root.addEventListener("click", async (e) => {
      if (e.target.closest("[data-theme-reset]")) {
        if (dirty && !await confirmDraft("Reset this draft?")) return;
        definition = copy(baseline);
        for (const el of invalidInputs) el.removeAttribute("aria-invalid");
        invalidInputs.clear();
        dirty = true;
        update(true);
      }
      if (e.target.closest("[data-theme-export]")) {
        try {
          const blob = new Blob([css(definition, fonts)], { type: "text/css" }), url = URL.createObjectURL(blob), a = document.createElement("a");
          a.href = url;
          a.download = "theme.css";
          a.click();
          URL.revokeObjectURL(url);
        } catch (err) {
          feedback.textContent = err.message;
        }
      }
    }, opts);
    root.addEventListener("themeeditor:load", (e) => {
      if (e.detail.id !== config.id) return;
      if (e.detail.definition) {
        if (!valid(e.detail.definition, keys)) return;
        definition = copy(e.detail.definition);
        baseline = copy(definition);
        for (const el of invalidInputs) el.removeAttribute("aria-invalid");
        invalidInputs.clear();
        dirty = !!e.detail.dirty;
      }
      if (["light", "dark"].includes(e.detail.previewMode)) {
        mode = e.detail.previewMode;
      }
      if (e.detail.fonts) {
        fonts = copy(e.detail.fonts);
        refreshFontFaces();
        syncFontPickers(true);
      }
      update(!!e.detail.definition);
    }, opts);
    root.addEventListener("themeeditor:committed", (e) => {
      if (e.detail.id !== config.id) return;
      dirty = false;
      baseline = copy(definition);
      update();
    }, opts);
    window.addEventListener("message", (e) => {
      const frame2 = document.getElementById(config.previewID);
      if (e.origin === location.origin && e.source === frame2?.contentWindow && e.data?.type === "goshtoso:theme-ready" && e.data.version === 1 && e.data.instanceId === config.previewID) preview();
    }, opts);
    update(true, "ready");
    return () => {
      fontObserver?.disconnect();
      confirmation?.(false);
      confirmation = null;
      abort.abort();
    };
  }
  async function frame(root) {
    const config = JSON.parse(root.dataset.themeFrame), abort = new AbortController();
    let sequence = 0, fonts = config.fonts || [], style;
    window.addEventListener("message", async (e) => {
      const m = e.data;
      if (e.origin !== location.origin || e.source !== parent || m?.type !== "goshtoso:theme-state" || m.version !== 1 || m.instanceId !== config.id || !Number.isSafeInteger(m.sequence) || m.sequence <= sequence) return;
      const keys = config.tokens.map((t) => t.key);
      if (!valid(m.definition, keys) || !keys.every((k) => /^[a-z][a-z0-9-]*$/.test(k)) || !["light", "dark"].includes(m.previewMode)) return;
      sequence = m.sequence;
      const current = m.sequence, warning = document.querySelector("[data-theme-font-warning]");
      try {
        if (config.fontCatalogURL && [m.definition.typography.title, m.definition.typography.body].some((ref) => !fonts.some((f) => f.source === ref.source && f.id === ref.id))) {
          const url = new URL(config.fontCatalogURL, location.origin);
          if (url.origin !== location.origin) throw Error("Invalid catalog origin.");
          const response = await fetch(url, { signal: abort.signal });
          if (!response.ok) throw Error("Font catalog unavailable.");
          const refreshed = await response.json();
          if (current !== sequence) return;
          fonts = refreshed;
        }
        const sheet = css(m.definition, fonts);
        if (current !== sequence) return;
        if (!style) {
          style = document.createElement("style");
          document.head.append(style);
        }
        style.textContent = sheet;
        root.dataset.theme = "custom";
        root.classList.toggle("dark", m.previewMode === "dark");
        warning.textContent = "";
        document.fonts.ready.then(async () => {
          for (const ref of Object.values(m.definition.typography)) {
            if (current !== sequence) return;
            if (!["upload", "bundled"].includes(ref.source)) continue;
            const family = fontAssets(ref, fonts).family.split(",")[0];
            const faces = await document.fonts.load("16px " + family);
            if (current === sequence && !faces.length) warning.textContent = "Font unavailable. Using system fallback.";
          }
        }).catch(() => {
          if (current === sequence) warning.textContent = "Font unavailable. Using system fallback.";
        });
      } catch (err) {
        if (current !== sequence) return;
        warning.textContent = err.message + " Using Goshtoso/System temporarily.";
        root.dataset.theme = "goshtoso";
        root.classList.toggle("dark", matchMedia("(prefers-color-scheme: dark)").matches);
      }
    }, { signal: abort.signal });
    parent.postMessage({ type: "goshtoso:theme-ready", version: 1, instanceId: config.id }, location.origin);
    return () => abort.abort();
  }
  const optionFonts = /* @__PURE__ */ new Map(), observedOptions = /* @__PURE__ */ new Set();
  let optionObserver;
  function loadStandaloneOption(option) {
    const font = JSON.parse(option.dataset.themeFontPreview), key = font.source + "/" + font.id;
    optionObserver?.unobserve(option);
    if (font.source === "system" || optionFonts.has(key)) return;
    const load = async () => {
      const ref = { source: font.source, id: font.id }, fonts = [font], assets = fontAssets(ref, fonts);
      if (font.source === "google") await loadGooglePreview(ref, fonts);
      else if (assets.face) {
        const style = document.createElement("style");
        style.textContent = assets.face;
        document.head.append(style);
      }
      await document.fonts.load("16px " + editorFontFamily(ref, fonts));
    };
    optionFonts.set(key, load().catch(() => {
      option.title = "Font preview unavailable. Using system fallback.";
    }));
  }
  function observeStandaloneOptions() {
    if (typeof IntersectionObserver === "undefined") return;
    optionObserver ||= new IntersectionObserver((entries) => {
      for (const entry of entries) if (entry.isIntersecting) loadStandaloneOption(entry.target);
    });
    document.querySelectorAll("[data-theme-font-preview]").forEach((option) => {
      if (document.getElementById(option.dataset.themeEditorFor)?.hasAttribute("data-theme-editor")) return;
      if (option.closest('[role="option"]')?.getAttribute("aria-selected") === "true") loadStandaloneOption(option);
      if (observedOptions.has(option)) return;
      observedOptions.add(option);
      optionObserver.observe(option);
    });
    for (const option of observedOptions) if (!option.isConnected) {
      optionObserver.unobserve(option);
      observedOptions.delete(option);
    }
  }
  const instances = /* @__PURE__ */ new Map();
  function mount() {
    observeStandaloneOptions();
    document.querySelectorAll("[data-theme-editor],[data-theme-frame]").forEach((root) => {
      if (instances.has(root)) return;
      const cleanup = root.hasAttribute("data-theme-editor") ? editor(root) : frame(root);
      instances.set(root, cleanup);
    });
    for (const [root, cleanup] of instances) if (!root.isConnected) {
      Promise.resolve(cleanup).then((fn) => fn());
      instances.delete(root);
    }
  }
  function start() {
    mount();
    if (document.documentElement) new MutationObserver(mount).observe(document.documentElement, { childList: true, subtree: true });
  }
  if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", start, { once: true });
  else start();
  document.addEventListener("htmx:afterSwap", mount);
  document.addEventListener("htmx:after:swap", mount);
})();
