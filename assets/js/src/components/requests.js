// Latest-result cancellation for the pinned htmx 4.0.0 RequestQueue.
// Its replace strategy aborts A and starts B, but A's finally calls continue(),
// clearing B's queue slot. Keep the real context handle until its own completion.
// Remove only after the delayed three-request regression passes without it.
(function () {
  "use strict";
  if (window.goshtosoRequests) return;
  var current = new WeakMap();
  window.goshtosoRequests = {
    form: function (root, event) {
      var ctx = event.detail.ctx;
      if (ctx.sourceElement === root) {
        window.goshtosoRequests.track(root, ctx);
        root.querySelectorAll("[data-form-validation]").forEach(function (field) {
          window.goshtosoRequests.abort(field);
        });
      } else if (ctx.sourceElement.hasAttribute("data-form-validation") && window.goshtosoRequests.pending(root)) {
        event.preventDefault();
      }
    },
    combobox: function (root, event) {
      var ctx = event.detail.ctx;
      if (ctx.request.method === "POST") {
        window.goshtosoRequests.track(root, ctx);
        var search = root.querySelector("[data-combobox-search]");
        window.goshtosoRequests.abort(search);
        window.goshtosoRequests.lockInput(search, ctx);
      } else if (window.goshtosoRequests.pending(root)) {
        event.preventDefault();
      }
    },
    table: function (root, event) {
      var ctx = event.detail.ctx;
      if (ctx.sourceElement.closest("[data-table-requests]") !== root) return;
      if (ctx.request.method === "GET" && (ctx.sourceElement.matches("[data-table-filters],[data-table-sort-request]") || ctx.sourceElement.closest("[data-table-pages]"))) {
        window.goshtosoRequests.latest(root, ctx);
      }
    },
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
