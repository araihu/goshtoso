// profile-images.js — IndexedDB-backed avatar and banner state for Profile.
(function () {
  var databaseName = "gt_profile";
  var storeName = "images";
  var maxBytes = 1024 * 1024;
  var allowedTypes = ["image/png", "image/jpeg", "image/webp", "image/gif"];

  function openDatabase() {
    return new Promise(function (resolve, reject) {
      var request = indexedDB.open(databaseName, 1);
      request.onupgradeneeded = function () {
        request.result.createObjectStore(storeName);
      };
      request.onsuccess = function () {
        resolve(request.result);
      };
      request.onerror = function () {
        reject(request.error);
      };
    });
  }

  function transact(mode, operation) {
    return openDatabase().then(function (database) {
      return new Promise(function (resolve, reject) {
        var transaction = database.transaction(storeName, mode);
        var request = operation(transaction.objectStore(storeName));
        transaction.oncomplete = function () {
          database.close();
          resolve(request && request.result);
        };
        var fail = function () {
          database.close();
          reject(transaction.error);
        };
        transaction.onerror = fail;
        transaction.onabort = fail;
      });
    });
  }

  function getImage(key) {
    return transact("readonly", function (store) {
      return store.get(key);
    });
  }

  function putImage(key, value) {
    return transact("readwrite", function (store) {
      return store.put(value, key);
    });
  }

  function deleteImage(key) {
    return transact("readwrite", function (store) {
      return store.delete(key);
    });
  }

  function register() {
    if (!window.Alpine || Alpine.__profileImagesRegistered) return;
    Alpine.__profileImagesRegistered = true;
    Alpine.data("profileImages", function () {
      return {
        avatarSrc: "",
        bannerSrc: "",
        _supported: typeof indexedDB !== "undefined",
        _destroyed: false,
        _revisions: { avatar: 0, banner: 0 },
        init: function () {
          var component = this;
          if (!this._supported) return;
          ["avatar", "banner"].forEach(function (kind) {
            var revision = component._revisions[kind];
            getImage(kind)
              .then(function (blob) {
                if (!blob) return;
                var objectURL = URL.createObjectURL(blob);
                if (component._destroyed || revision !== component._revisions[kind]) {
                  URL.revokeObjectURL(objectURL);
                  return;
                }
                var oldURL = component[kind + "Src"];
                if (oldURL) URL.revokeObjectURL(oldURL);
                component[kind + "Src"] = objectURL;
              })
              .catch(function () {});
          });
        },
        pick: function (kind) {
          var input = document.getElementById("profile-" + kind + "-input");
          if (input) input.click();
        },
        onFile: function (kind, event) {
          var file = event.target.files && event.target.files[0];
          if (!file) return;
          if (!allowedTypes.includes(file.type)) {
            this._toast("danger", "Unsupported type", "Use PNG, JPG, WebP, or GIF.");
            event.target.value = "";
            return;
          }
          if (file.size > maxBytes) {
            this._toast("danger", "Too large", "Images must be 1 MB or smaller.");
            event.target.value = "";
            return;
          }
          this._revisions[kind] += 1;
          var oldURL = this[kind + "Src"];
          if (oldURL) URL.revokeObjectURL(oldURL);
          this[kind + "Src"] = URL.createObjectURL(file);
          if (this._supported) {
            var component = this;
            putImage(kind, file).catch(function () {
              if (!component._destroyed) {
                component._toast("warning", "Won't persist", "Saved for this session only.");
              }
            });
          } else {
            this._toast("warning", "Won't persist", "IndexedDB unavailable in this browser.");
          }
          event.target.value = "";
        },
        remove: function (kind) {
          this._revisions[kind] += 1;
          var oldURL = this[kind + "Src"];
          if (oldURL) URL.revokeObjectURL(oldURL);
          this[kind + "Src"] = "";
          if (this._supported) deleteImage(kind).catch(function () {});
        },
        destroy: function () {
          this._destroyed = true;
          this._revisions.avatar += 1;
          this._revisions.banner += 1;
          if (this.avatarSrc) URL.revokeObjectURL(this.avatarSrc);
          if (this.bannerSrc) URL.revokeObjectURL(this.bannerSrc);
        },
        _toast: function (tone, title, message) {
          try {
            window.dispatchEvent(
              new CustomEvent("notify", {
                detail: { kind: "toast", tone: tone, title: title, message: message },
              }),
            );
          } catch (error) {
            console.warn("[profileImages] toast unavailable:", title, "-", message);
          }
        },
      };
    });
  }

  if (window.Alpine) register();
  else document.addEventListener("alpine:init", register);
})();
