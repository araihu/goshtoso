// Latest-result cancellation for the pinned htmx 4.0.0 RequestQueue.
// Its replace strategy aborts A and starts B, but A's finally calls continue(),
// clearing B's queue slot. Keep the real context handle until its own completion.
// Remove only after the delayed three-request regression passes without it.
(function () {
  "use strict";
  if (window.goshtosoRequests) return;
  var current = new WeakMap();
  window.goshtosoRequests = {
    abort: function (owner) {
      var ctx = owner && current.get(owner);
      if (ctx) {
        current.delete(owner);
        ctx.request.abort();
      }
    },
    lockInput: function (input, ctx) {
      if (!input) return;
      var readOnly = input.readOnly;
      input.readOnly = true;
      function finished(event) {
        if (event.detail.ctx !== ctx) return;
        input.readOnly = readOnly;
        ctx.sourceElement.removeEventListener("htmx:finally:request", finished);
      }
      ctx.sourceElement.addEventListener("htmx:finally:request", finished);
    },
    pending: function (owner) { return !!owner && current.has(owner); },
    latest: function (owner, ctx) {
      window.goshtosoRequests.abort(owner);
      window.goshtosoRequests.track(owner, ctx);
    },
    track: function (owner, ctx) {
      if (!owner) return;
      current.set(owner, ctx);
      var source = ctx.sourceElement;
      function cleanup(event) {
        if (event.target === source && current.get(owner) === ctx) window.goshtosoRequests.abort(owner);
      }
      function finished(event) {
        if (event.detail.ctx !== ctx) return;
        if (current.get(owner) === ctx) current.delete(owner);
        source.removeEventListener("htmx:before:cleanup", cleanup);
        source.removeEventListener("htmx:finally:request", finished);
      }
      source.addEventListener("htmx:before:cleanup", cleanup);
      source.addEventListener("htmx:finally:request", finished);
    },
  };
})();
