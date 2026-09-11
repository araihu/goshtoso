// chat.js — global htmx hooks for the chat example.
(function () {
  if (window.__gtChatInit) return;
  window.__gtChatInit = true;

  document.addEventListener("htmx:ws:after:message:outgoing", function () {
    var textarea = document.getElementById("chat-message");
    if (!textarea) return;
    textarea.value = "";
    textarea.focus();
  });

  document.addEventListener("htmx:ws:after:message:incoming", function (event) {
    var log = document.getElementById("chat-log");
    if (!log) return;
    log.scrollTop = log.scrollHeight;
  });
})();
