// chat.js — global htmx hooks for the chat example.
(function () {
  if (window.__gtChatInit) return;
  window.__gtChatInit = true;

  document.addEventListener("htmx:wsAfterSend", function () {
    var textarea = document.getElementById("chat-message");
    if (!textarea) return;
    textarea.value = "";
    textarea.focus();
  });

  document.addEventListener("htmx:oobAfterSwap", function (event) {
    var log = document.getElementById("chat-log");
    if (!log || !event.detail || event.detail.target !== log) return;
    log.scrollTop = log.scrollHeight;
  });
})();
