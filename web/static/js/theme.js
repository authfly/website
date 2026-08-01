(function () {
  var root = document.documentElement;
  var themeStorageKey = "ui8kit-theme";

  function ready(fn) {
    if (document.readyState === "loading") {
      document.addEventListener("DOMContentLoaded", fn);
      return;
    }
    fn();
  }

  function readStoredTheme() {
    try {
      return localStorage.getItem(themeStorageKey);
    } catch (_) {
      return null;
    }
  }

  function writeStoredTheme(value) {
    try {
      localStorage.setItem(themeStorageKey, value);
    } catch (_) {}
  }

  function resolvePreferredTheme() {
    var storedTheme = readStoredTheme();
    if (storedTheme === "dark" || storedTheme === "light") {
      return storedTheme;
    }
    var prefersDark =
      window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches;
    return prefersDark ? "dark" : "light";
  }

  function applyTheme(theme) {
    root.classList.toggle("dark", theme === "dark");
  }

  function applyThemeButtonState() {
    var button = document.getElementById("ui8kit-theme-toggle");
    if (!button) return;
    var icon = button.querySelector("#theme-toggle-icon");
    var dark = root.classList.contains("dark");
    var switchToDark = button.dataset.switchToDarkLabel || "Dark theme";
    var switchToLight = button.dataset.switchToLightLabel || "Light theme";

    if (icon) {
      icon.textContent = dark ? "☀" : "☾";
    }
    button.setAttribute("aria-pressed", dark ? "true" : "false");
    button.setAttribute("title", dark ? switchToLight : switchToDark);
    button.setAttribute("aria-label", dark ? switchToLight : switchToDark);
  }

  applyTheme(resolvePreferredTheme());

  ready(function () {
    var themeButton = document.getElementById("ui8kit-theme-toggle");
    if (themeButton) {
      themeButton.addEventListener("click", function () {
        var nextTheme = root.classList.contains("dark") ? "light" : "dark";
        applyTheme(nextTheme);
        writeStoredTheme(nextTheme);
        applyThemeButtonState();
      });
    }
    applyThemeButtonState();
  });
})();
