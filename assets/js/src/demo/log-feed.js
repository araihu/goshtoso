// log-feed.js — Alpine state for the streaming logs example.
(function () {
  var register = function () {
    if (!window.Alpine || Alpine.__logFeedRegistered) return;
    Alpine.__logFeedRegistered = true;
    Alpine.data("logFeed", function () {
      return {
        cap: 100,
        paused: false,
        autoScroll: true,
        connected: false,
        minLevel: "all",
        init: function () {
          this.handleAfterSwap = this.onSwap.bind(this);
          this.$root.addEventListener("htmx:afterSwap", this.handleAfterSwap);
          this.$watch("minLevel", this.applyFilter.bind(this));
          this.applyFilter(this.minLevel);
        },
        destroy: function () {
          if (this.handleAfterSwap) {
            this.$root.removeEventListener("htmx:afterSwap", this.handleAfterSwap);
          }
        },
        applyFilter: function (value) {
          var wrap = this.$refs.feedWrap;
          if (!wrap) return;
          wrap.classList.remove("flt-all", "flt-warn", "flt-error");
          wrap.classList.add("flt-" + value);
        },
        onSwap: function (event) {
          var swapTarget = event.detail && event.detail.target;
          var eventTarget = event.target instanceof Element ? event.target : null;
          var affectedFeed =
            (swapTarget && swapTarget.id === "log-feed") ||
            (eventTarget &&
              (eventTarget.id === "log-feed" || eventTarget.closest("#log-feed")));
          if (!affectedFeed) return;
          var feed = document.getElementById("log-feed");
          if (!feed) return;
          while (feed.children.length > this.cap) feed.removeChild(feed.firstElementChild);
          if (this.autoScroll) {
            requestAnimationFrame(function () {
              feed.scrollTop = feed.scrollHeight;
            });
          }
        },
        togglePause: function () {
          this.paused = !this.paused;
          if (this.paused) {
            this.connected = false;
            return;
          }
          this.$nextTick(function () {
            var element = document.querySelector("#logs-fragment [sse-connect]");
            if (element && window.htmx) window.htmx.process(element);
          });
        },
        clearFeed: function () {
          var feed = document.getElementById("log-feed");
          if (feed) feed.replaceChildren();
        },
        get statusText() {
          return this.paused ? "Paused" : this.connected ? "Connected" : "Connecting";
        },
      };
    });
  };

  if (window.Alpine) register();
  else document.addEventListener("alpine:init", register);
})();
