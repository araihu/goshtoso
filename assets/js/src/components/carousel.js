// Carousel Alpine factory. Configuration is JSON data on each root so this
// global exists before Alpine's first scan and also works in HTMX fragments.
(function () {
  if (window.goshtosoCarousel) return;

  window.goshtosoCarousel = function (root) {
    var autoplayInterval = Number(root.dataset.carouselAutoplayInterval) || 0;
    var state = {
      slides: window.goshtosoParseData(root.dataset.carouselSlides, []),
      currentSlideIndex: 1,
      autoplayIntervalTime: autoplayInterval || 4000,
      isPaused: false,
      autoplayInterval: null,
      touchStartX: null,
      touchEndX: null,
      swipeThreshold: 50,
      previous: function () {
        if (this.currentSlideIndex > 1) this.currentSlideIndex -= 1;
        else this.currentSlideIndex = this.slides.length;
      },
      next: function () {
        if (this.currentSlideIndex < this.slides.length) this.currentSlideIndex += 1;
        else this.currentSlideIndex = 1;
      },
      autoplay: function () {
        if (!autoplayInterval) return;
        clearInterval(this.autoplayInterval);
        this.autoplayInterval = setInterval(function () {
          if (!state.isPaused) state.next();
        }, this.autoplayIntervalTime);
      },
      setAutoplayInterval: function (interval) {
        clearInterval(this.autoplayInterval);
        this.autoplayIntervalTime = interval;
        this.autoplay();
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
      },
    };
    return state;
  };
})();
