// ABOUTME: Handles mobile hamburger toggle for the nav-links list.
// ABOUTME: Loaded on all pages to provide consistent nav behavior.

(function () {
  document.addEventListener("DOMContentLoaded", function () {
    var toggle = document.querySelector(".nav-toggle");
    if (!toggle) return;

    var menu = document.querySelector(".nav-links");
    if (!menu) return;

    toggle.addEventListener("click", function () {
      var expanded = toggle.getAttribute("aria-expanded") === "true";
      toggle.setAttribute("aria-expanded", String(!expanded));
      menu.classList.toggle("open");
    });

    // Close mobile menu when a link is clicked
    menu.querySelectorAll("a").forEach(function (link) {
      link.addEventListener("click", function () {
        menu.classList.remove("open");
        toggle.setAttribute("aria-expanded", "false");
      });
    });
  });
})();
