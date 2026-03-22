// Theme toggle for Pico CSS dark/light mode.
// Persists preference to localStorage.

(function () {
  const STORAGE_KEY = "pvm-theme";
  const html = document.documentElement;

  function getPreferred() {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored === "light" || stored === "dark") return stored;
    return window.matchMedia("(prefers-color-scheme: dark)").matches
      ? "dark"
      : "light";
  }

  function apply(theme) {
    html.setAttribute("data-theme", theme);
    localStorage.setItem(STORAGE_KEY, theme);
    const btn = document.getElementById("theme-toggle");
    if (btn) btn.textContent = theme === "dark" ? "Light" : "Dark";
  }

  apply(getPreferred());

  document.addEventListener("click", function (e) {
    if (e.target && e.target.id === "theme-toggle") {
      const current = html.getAttribute("data-theme");
      apply(current === "dark" ? "light" : "dark");
    }
  });
})();
