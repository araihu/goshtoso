// Search factories. Per-instance values come from inert data attributes.
(function () {
  "use strict";

  if (window.goshtosoSearchField && window.goshtosoSearchModal) return;

  window.goshtosoSearchField = function (root) {
    var globalShortcut = root.dataset.searchGlobalShortcut === "true";
    return {
      searchId: root.dataset.searchId || "search",
      open: false,
      openSearch: function () {
        window.dispatchEvent(
          new CustomEvent("goshtoso-search-open", {
            detail: { id: this.searchId },
          }),
        );
      },
      handleWindowKey: function (event) {
        if (!globalShortcut) return;
        if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === "k") {
          event.preventDefault();
          this.openSearch();
        }
      },
    };
  };

  window.goshtosoSearchModal = function (root) {
    return {
      searchId: root.dataset.searchId || "search",
      open: false,
      query: "",
      activeIndex: 0,
      maxResults: Number(root.dataset.searchMaxResults) || 4,
      descriptionMaxLength: Number(root.dataset.searchDescriptionMaxLength) || 120,
		matchMode: root.dataset.searchMatchMode === "fuzzy" ? "fuzzy" : "substring",
      cachedAllResults: null,
      cachedDOMTerm: null,
      cachedDOMResults: [],
      clientItems: [],
      clientItemsLoaded: false,
      clientItemsLoading: false,
      cachedClientTerm: null,
      cachedClientResults: [],
      requestController: null,
      openSearch: function () {
        this.open = true;
        this.activeIndex = 0;
        if (!this.usesClientSource()) this.resetDOMCache();
        this.ensureClientItems();
        this.$nextTick(function () {
          if (this.$refs.input) this.$refs.input.focus();
        }.bind(this));
      },
      closeSearch: function () {
        this.open = false;
        this.query = "";
        this.activeIndex = 0;
        window.dispatchEvent(
          new CustomEvent("goshtoso-search-close", {
            detail: { id: this.searchId },
          }),
        );
      },
      destroy: function () {
        if (this.requestController) this.requestController.abort();
        this.requestController = null;
      },
      handleWindowKey: function (event) {
        if (this.open && event.key === "Escape") {
          event.preventDefault();
          this.closeSearch();
        }
      },
		compactSearchValue: function (value) {
			return this.stringValue(value).toLowerCase().replace(/[^a-z0-9]+/g, "");
		},
		fuzzyScore: function (query, value) {
			var needle = this.compactSearchValue(query);
			var haystack = this.compactSearchValue(value);
			if (!needle || !haystack) return -1;
			var direct = haystack.indexOf(needle);
			if (direct >= 0) return direct;
			var cursor = 0;
			var first = -1;
			var last = -1;
			var gaps = 0;
			for (var i = 0; i < haystack.length && cursor < needle.length; i++) {
				if (haystack[i] !== needle[cursor]) continue;
				if (first < 0) first = i;
				if (last >= 0) gaps += i - last - 1;
				last = i;
				cursor++;
			}
			if (cursor !== needle.length) return -1;
			return 20 + first + gaps + Math.max(0, haystack.length - needle.length) / 10;
		},
		resultScore: function (query, title, text, priority) {
			var term = this.stringValue(query).trim().toLowerCase();
			title = this.stringValue(title);
			text = this.stringValue(text);
			priority = Number(priority) || 0;
			var contextScore = -priority;
			if (!term) return null;
			if (title.toLowerCase().indexOf(term) !== -1) return [0, contextScore, 0];
			if (text.toLowerCase().indexOf(term) !== -1) return [1, contextScore, 0];
			if (this.matchMode !== "fuzzy") return null;
			var titleScore = this.fuzzyScore(term, title);
			if (titleScore >= 0) return [2, contextScore, titleScore];
			var textScore = this.fuzzyScore(term, text);
			if (textScore >= 0) return [3, contextScore, textScore];
			return null;
		},
		rankedMatches: function (values, scoreFor) {
			var matches = [];
			values.forEach(function (value, index) {
				var score = scoreFor(value);
				if (score !== null) matches.push({ value: value, index: index, score: score });
			});
			matches.sort(function (left, right) {
				for (var scoreIndex = 0; scoreIndex < Math.max(left.score.length, right.score.length); scoreIndex++) {
					var leftScore = left.score[scoreIndex] || 0;
					var rightScore = right.score[scoreIndex] || 0;
					if (leftScore !== rightScore) return leftScore - rightScore;
				}
				return left.index - right.index;
			});
			return matches.map(function (match) { return match.value; });
		},
      sourceURL: function () {
        return root.dataset.searchSourceUrl || "";
      },
      usesClientSource: function () {
        return this.sourceURL() !== "";
      },
      resetClientCache: function () {
        this.cachedClientTerm = null;
        this.cachedClientResults = [];
      },
      resetDOMCache: function () {
        this.cachedAllResults = null;
        this.cachedDOMTerm = null;
        this.cachedDOMResults = [];
      },
      stringValue: function (value) {
        if (value === null || value === undefined) return "";
        return String(value);
      },
      searchTextForItem: function (item) {
        return [item.title, item.description, item.kind, item.method, item.path, item.section]
          .concat(item.keywords || [])
          .filter(Boolean)
          .join(" ");
      },
      methodBadgeClasses: function (method) {
        switch (this.stringValue(method).trim().toUpperCase()) {
          case "GET":
            return "border border-primary bg-primary text-on-primary dark:border-primary-dark dark:bg-primary-dark dark:text-on-primary-dark";
          case "POST":
            return "border border-success bg-success text-on-success";
          case "PUT":
          case "PATCH":
            return "border border-warning bg-warning text-on-warning";
          case "DELETE":
            return "border border-danger bg-danger text-on-danger";
          default:
            return "border border-outline bg-surface-alt text-on-surface dark:border-outline-dark dark:bg-surface-dark-alt dark:text-on-surface-dark";
        }
      },
      safeHref: function (value) {
        return window.goshtosoSafeNavigationTarget(this.stringValue(value));
      },
      normalizeItems: function (payload) {
        var values = Array.isArray(payload)
          ? payload
          : Array.isArray(payload && payload.items)
            ? payload.items
            : [];
        var state = this;
        return values.map(function (raw, index) {
          raw = raw || {};
          var keywords = raw.keywords !== undefined ? raw.keywords : raw.Keywords;
          if (!Array.isArray(keywords)) keywords = [];
          var item = {
            id: state.stringValue(raw.id !== undefined ? raw.id : raw.ID),
            title: state.stringValue(raw.title !== undefined ? raw.title : raw.Title),
            description: state.stringValue(
              raw.description !== undefined ? raw.description : raw.Description,
            ),
            href: state.safeHref(raw.href !== undefined ? raw.href : raw.Href),
            kind: state.stringValue(raw.kind !== undefined ? raw.kind : raw.Kind),
            method: state
              .stringValue(raw.method !== undefined ? raw.method : raw.Method)
              .trim()
              .toUpperCase(),
				path: state.stringValue(raw.path !== undefined ? raw.path : raw.Path),
				section: state.stringValue(raw.section !== undefined ? raw.section : raw.Section),
				priority: Number(raw.priority !== undefined ? raw.priority : raw.Priority) || 0,
				keywords: keywords.map(function (keyword) {
              return state.stringValue(keyword);
            }),
            index: index,
          };
          item.key = item.id || item.href || item.title || String(index);
          item.searchText = state.searchTextForItem(item);
          return item;
        });
      },
      ensureClientItems: function () {
        if (this.usesClientSource() === false || this.clientItemsLoaded || this.clientItemsLoading) {
          return;
        }
        this.clientItemsLoading = true;
        var controller = typeof AbortController === "function" ? new AbortController() : null;
        this.requestController = controller;
        var requestOptions = { headers: { Accept: "application/json" } };
        if (controller) requestOptions.signal = controller.signal;
        fetch(this.sourceURL(), requestOptions)
          .then(function (response) {
            if (!response.ok) throw new Error("Search items request failed");
            return response.json();
          })
          .then(function (payload) {
            this.clientItems = this.normalizeItems(payload);
            this.clientItemsLoaded = true;
            this.resetClientCache();
          }.bind(this))
          .catch(function (error) {
            if (error && error.name === "AbortError") return;
            this.clientItems = [];
            this.clientItemsLoaded = true;
            this.resetClientCache();
          }.bind(this))
          .finally(function () {
            if (this.requestController === controller) this.requestController = null;
            this.clientItemsLoading = false;
          }.bind(this));
      },
      allResults: function () {
        if (!this.cachedAllResults) {
          this.cachedAllResults = Array.prototype.slice.call(
            this.$root.querySelectorAll("[data-search-result]"),
          );
        }
        return this.cachedAllResults;
      },
      matchedDOMResults: function () {
        var term = this.query.trim().toLowerCase();
        if (!term) return [];
        if (term === this.cachedDOMTerm) return this.cachedDOMResults;
		this.cachedDOMTerm = term;
			this.cachedDOMResults = this.rankedMatches(this.allResults(), function (element) {
				return this.resultScore(term, element.dataset.searchTitle, element.dataset.searchText, element.dataset.searchPriority);
		}.bind(this)).slice(0, this.maxResults);
        return this.cachedDOMResults;
      },
      matchedClientResults: function () {
        var term = this.query.trim().toLowerCase();
        if (!term) return [];
        if (term === this.cachedClientTerm) return this.cachedClientResults;
		this.cachedClientTerm = term;
			this.cachedClientResults = this.rankedMatches(this.clientItems, function (item) {
				return this.resultScore(term, item.title, item.searchText, item.priority);
		}.bind(this)).slice(0, this.maxResults);
        return this.cachedClientResults;
      },
      matchedResults: function () {
        if (this.usesClientSource()) return this.matchedClientResults();
        return this.matchedDOMResults();
      },
      visibleResults: function () {
        return this.matchedResults();
      },
      isResultVisible: function (element) {
        return this.matchedResults().indexOf(element) !== -1;
      },
      resultOrder: function (element) {
        var index = this.matchedResults().indexOf(element);
        return index < 0 ? 999 : index;
      },
      move: function (delta) {
        var results = this.visibleResults();
        if (!results.length) return;
        this.activeIndex = (this.activeIndex + delta + results.length) % results.length;
        this.$nextTick(function () {
          var element = this.usesClientSource()
            ? this.$root.querySelectorAll("[data-search-result]")[this.activeIndex]
            : results[this.activeIndex];
          if (element) element.scrollIntoView({ block: "nearest" });
        }.bind(this));
      },
      setActive: function (element) {
        var results = this.visibleResults();
        var index = results.indexOf(element);
        if (index >= 0) this.activeIndex = index;
      },
      setActiveIndex: function (index) {
        this.activeIndex = index;
      },
      isActiveIndex: function (index) {
        return index === this.activeIndex;
      },
      isActive: function (element) {
        return this.visibleResults().indexOf(element) === this.activeIndex;
      },
      choose: function () {
        var results = this.visibleResults();
        if (!results.length) return;
        if (this.usesClientSource()) this.selectResult(results[this.activeIndex]);
        else results[this.activeIndex].click();
      },
      selectResult: function (result) {
        this.closeSearch();
        if (!result) return;
        if (result.dataset) {
          if (
            !result.hasAttribute("hx-get") &&
            !result.hasAttribute("hx-post") &&
            result.dataset.searchHref
          ) {
            var elementHref = this.safeHref(result.dataset.searchHref);
            if (elementHref) window.location.href = elementHref;
          }
          return;
        }
        if (result.href) {
          var itemHref = this.safeHref(result.href);
          if (itemHref) window.location.href = itemHref;
        }
      },
      truncate: function (value) {
        value = value || "";
        if (value.length <= this.descriptionMaxLength) return value;
        return value.substring(0, this.descriptionMaxLength) + "...";
      },
      escapeHTML: function (value) {
        return (value || "").replace(/[&<>"']/g, function (character) {
          return {
            "&": "&amp;",
            "<": "&lt;",
            ">": "&gt;",
            '"': "&quot;",
            "'": "&#39;",
          }[character];
        });
      },
      highlight: function (value) {
        var escaped = this.escapeHTML(value);
        var term = this.query.trim();
        if (!term) return escaped;
        var pattern = term.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
        return escaped.replace(new RegExp(pattern, "gi"), function (match) {
          return '<span class="search-highlight">' + match + "</span>";
        });
      },
    };
  };
})();
