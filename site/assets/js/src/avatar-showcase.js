// avatar-showcase.js — shared responsive state for Avatar demo previews.
(function () {
  var register = function () {
    if (!window.Alpine || Alpine.__avatarShowcaseRegistered) return;
    Alpine.__avatarShowcaseRegistered = true;
    Alpine.data("avatarShowcase", function () {
      return {
        selected: "md",
        radius: "md",
        sizes: ["xs", "sm", "md", "lg", "xl", "2xl"],
        radii: ["none", "xs", "sm", "md", "lg"],
        sizeMap: {
          xs: "size-8 text-xs",
          sm: "size-10 text-sm",
          md: "size-14 text-2xl",
          lg: "size-20 text-3xl",
          xl: "size-24 text-4xl",
          "2xl": "size-32 text-5xl",
        },
        radiusMap: {
          none: "rounded-none",
          xs: "rounded-xs",
          sm: "rounded-sm",
          md: "rounded-md",
          lg: "rounded-lg",
        },
        statusSizeMap: {
          xs: "size-2",
          sm: "size-2.5",
          md: "size-4",
          lg: "size-5",
          xl: "size-6",
          "2xl": "size-7",
        },
        get avatarSizeClass() {
          return this.sizeMap[this.selected];
        },
        get avatarRadiusClass() {
          return this.radiusMap[this.radius];
        },
        get avatarStatusSizeClass() {
          return this.statusSizeMap[this.selected];
        },
      };
    });
  };

  if (window.Alpine) register();
  else document.addEventListener("alpine:init", register);
})();
