// charts-showcase.js — reveal the isolated preview only after Charts paints.
(function () {
  window.addEventListener("load", function () {
    var fallback = document.querySelector("[data-chart-fallback]");
    if (fallback && document.querySelector("canvas")) fallback.hidden = true;
  });
})();
