// Each preview runs a real shell document in its own responsive viewport.
(function () {
  function register() {
    Alpine.data("shellPreview", function () {
      return {
        mobile: false,
        available: 0,
        observer: null,
        init: function () {
          var self = this;
          this.observer = new ResizeObserver(function (entries) {
            self.available = entries[0].contentRect.width;
          });
          this.observer.observe(this.$refs.stage);
        },
        destroy: function () { this.observer.disconnect(); },
        get width() { return this.mobile ? 390 : 1280; },
        get scale() { return Math.min(1, (this.available || this.width) / this.width); },
        get stageStyle() { return "height:" + (800 * this.scale) + "px"; },
        get frameStyle() {
          return "width:" + this.width + "px;height:800px;transform:scale(" + this.scale + ");transform-origin:top left;position:absolute;left:50%;margin-left:-" + (this.width * this.scale / 2) + "px";
        }
      };
    });
  }
  if (window.Alpine) register();
  else document.addEventListener("alpine:init", register, { once: true });
})();
