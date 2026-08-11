// Toast client-message dismissal lives in the public bundle so templ event
// attributes remain small. Keep the existing two-stage removal lifecycle:
// begin the leave state, then ask the container to remove the notification.
(function () {
  if (window.goshtosoDismissClientMessageToast) return;

  window.goshtosoDismissClientMessageToast = function (state, notificationID) {
    state.isVisible = false;
    window.setTimeout(function () {
      state.removeNotification(notificationID);
    }, state.reducedMotion ? 0 : 400);
  };
})();
