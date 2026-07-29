// Select Alpine factories. Data comes from JSON attributes, not executable Go.
(function () {
  if (window.goshtosoSelect && window.goshtosoSelectShell) return;

  window.goshtosoSelectShell = function () {
    return { isOpen: false, openedWithKeyboard: false };
  };

  window.goshtosoSelect = function (root) {
    var config = window.goshtosoParseData(root.dataset.selectConfig, {
      placeholder: "Please Select",
      options: [],
      selectedValues: [],
      activeIndex: 0,
    });
    return {
      placeholder: config.placeholder || "Please Select",
      allOptions: Array.isArray(config.options) ? config.options : [],
      selectedValues: Array.isArray(config.selectedValues) ? config.selectedValues : [],
      isOpen: false,
      openedWithKeyboard: false,
      activeIndex: Number(config.activeIndex) || 0,
      alpineModel: config.alpineModel || "",
      modelUnwatches: [],
      get selectedOption() {
        if (this.selectedValues.length === 0) return null;
        var value = this.selectedValues[0];
        return this.allOptions.find(function (option) { return option.value === value; }) || null;
      },
      init: function () {
        if (!this.alpineModel || !window.Alpine) return;
        var state = this;
        this.modelUnwatches.push(this.$watch("selectedOption", function (option) {
          window.Alpine.evaluate(root, state.alpineModel + " = value", { scope: { value: option ? option.value : "" } });
        }));
        this.modelUnwatches.push(this.$watch(this.alpineModel, function (value) {
          var current = state.selectedValues.length ? state.selectedValues[0] : "";
          if (value === current) return;
          state.syncFromInput(value || "");
        }));
        var initial = window.Alpine.evaluate(root, this.alpineModel);
        if (initial) this.syncFromInput(initial);
      },
      destroy: function () {
        this.modelUnwatches.forEach(function (unwatch) {
          if (typeof unwatch === "function") unwatch();
        });
        this.modelUnwatches = [];
      },
      selectOption: function (option) {
        this.selectedValues = [option.value];
        this.activeIndex = this.allOptions.findIndex(function (item) { return item.value === option.value; });
        this.isOpen = false;
        this.openedWithKeyboard = false;
        this.$nextTick(function () {
          if (this.$refs.trigger) this.$refs.trigger.focus();
        }.bind(this));
      },
      optionElements: function () {
        return Array.from(this.$root.querySelectorAll('[role="option"]'));
      },
      openFromTrigger: function (direction) {
        this.openedWithKeyboard = true;
        this.$nextTick(function () {
          var options = this.optionElements();
          if (options.length === 0) return;
          var selectedIndex = this.selectedOption
            ? this.allOptions.findIndex(function (item) { return item.value === this.selectedOption.value; }, this)
            : -1;
          this.activeIndex = selectedIndex >= 0
            ? (selectedIndex + direction + options.length) % options.length
            : (direction < 0 ? options.length - 1 : 0);
          this.$nextTick(function () { options[this.activeIndex].focus(); }.bind(this));
        }.bind(this));
      },
      moveActiveOption: function (current, direction) {
        var options = this.optionElements();
        if (options.length === 0) return;
        var currentIndex = options.indexOf(current);
        this.activeIndex = (currentIndex + direction + options.length) % options.length;
        this.$nextTick(function () { options[this.activeIndex].focus(); }.bind(this));
      },
      syncFromInput: function (value) {
        var option = this.allOptions.find(function (item) { return item.value === value; });
        this.selectedValues = option ? [option.value] : [];
        this.activeIndex = option ? this.allOptions.indexOf(option) : 0;
      },
    };
  };
})();
