// A toast owns its timer and waits for x-show's public DOM hide completion.
(function () {
  if (window.goshtosoToast) return;
  window.goshtosoToast = function (root, duration, persistent, complete) {
    var timer = null;
    var observer = null;
    var destroyed = false;
    var closing = false;
    var owner = root.closest('[data-toast-container]');
    return {
      isVisible: Boolean(persistent),
      init: function () {
        this.$nextTick(() => {
          if (destroyed || closing) return;
          this.isVisible = true;
          this.resume();
        });
      },
      pause: function (event) {
        if (event && event.detail.container !== owner) return;
        clearTimeout(timer);
        timer = null;
      },
      resume: function (event) {
        if (event && event.detail.container !== owner) return;
        this.pause();
        if (persistent || closing || destroyed) return;
        timer = setTimeout(() => this.dismiss(), duration);
      },
      dismiss: function () {
        if (closing || destroyed) return;
        closing = true;
        this.pause();
        // x-show applies display:none after its leave transition completes,
        // including zero-duration/reduced-motion and canceled transitions.
        var finish = function () {
          if (destroyed || root.style.display !== 'none') return;
          observer.disconnect();
          observer = null;
          destroyed = true;
          complete();
        };
        observer = new MutationObserver(finish);
        observer.observe(root, { attributes: true, attributeFilter: ['style'] });
        this.isVisible = false;
        this.$nextTick(finish);
      },
      destroy: function () {
        destroyed = true;
        this.pause();
        if (observer) observer.disconnect();
        observer = null;
      },
    };
  };
})();
