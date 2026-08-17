// Carousel Alpine factory. Configuration is JSON data on each root so this
// global exists before Alpine's first scan and also works in HTMX fragments.
(function () {
  if (window.goshtosoCarousel) return;

  window.goshtosoCarousel = function (root) {
    var autoplayInterval = Number(root.dataset.carouselAutoplayInterval) || 0;
    var motionQuery = window.matchMedia("(prefers-reduced-motion: reduce)");
    var motionChangeHandler = null;
    var state = {
      slides: window.goshtosoParseData(root.dataset.carouselSlides, []),
      currentSlideIndex: 1,
      autoplayIntervalTime: autoplayInterval || 4000,
      isPaused: false,
      autoplayInterval: null,
      reducedMotion: motionQuery.matches,
      touchStartX: null,
      touchEndX: null,
      swipeThreshold: 50,
      init: function () {
        motionChangeHandler = function (event) {
          state.handleReducedMotionChange(event);
        };
        if (motionQuery.addEventListener) {
          motionQuery.addEventListener("change", motionChangeHandler);
        } else {
          motionQuery.addListener(motionChangeHandler);
        }
      },
      previous: function () {
        if (this.currentSlideIndex > 1) this.currentSlideIndex -= 1;
        else this.currentSlideIndex = this.slides.length;
      },
      next: function () {
        if (this.currentSlideIndex < this.slides.length) this.currentSlideIndex += 1;
        else this.currentSlideIndex = 1;
      },
      autoplay: function () {
        clearInterval(this.autoplayInterval);
        this.autoplayInterval = null;
        if (!autoplayInterval || this.reducedMotion) return;
        this.autoplayInterval = setInterval(function () {
          if (!state.isPaused) state.next();
        }, this.autoplayIntervalTime);
      },
      setAutoplayInterval: function (interval) {
        clearInterval(this.autoplayInterval);
        this.autoplayIntervalTime = interval;
        this.autoplay();
      },
      handleReducedMotionChange: function (event) {
        this.reducedMotion = event.matches;
        clearInterval(this.autoplayInterval);
        this.autoplayInterval = null;
        if (!this.reducedMotion) this.autoplay();
      },
      handleTouchStart: function (event) {
        this.touchStartX = event.touches[0].clientX;
      },
      handleTouchMove: function (event) {
        this.touchEndX = event.touches[0].clientX;
      },
      handleTouchEnd: function () {
        if (this.touchEndX) {
          if (this.touchStartX - this.touchEndX > this.swipeThreshold) this.next();
          if (this.touchStartX - this.touchEndX < -this.swipeThreshold) this.previous();
          this.touchStartX = null;
          this.touchEndX = null;
        }
      },
      destroy: function () {
        clearInterval(this.autoplayInterval);
        this.autoplayInterval = null;
        if (!motionChangeHandler) return;
        if (motionQuery.removeEventListener) {
          motionQuery.removeEventListener("change", motionChangeHandler);
        } else {
          motionQuery.removeListener(motionChangeHandler);
        }
        motionChangeHandler = null;
      },
    };
    return state;
  };
})();
