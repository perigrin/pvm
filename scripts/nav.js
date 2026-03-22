// ABOUTME: Handles nav dropdown close-on-outside-click and mobile hamburger toggle.
// ABOUTME: Loaded on all pages to provide consistent nav behavior.

(function () {
  document.addEventListener("DOMContentLoaded", function () {
    // Close nav dropdowns when clicking outside or when another opens
    var allDetails = document.querySelectorAll("nav details");
    allDetails.forEach(function (details) {
      details.addEventListener("toggle", function () {
        if (details.open) {
          allDetails.forEach(function (other) {
            if (other !== details) other.removeAttribute("open");
          });
        }
      });
    });

    document.addEventListener("click", function (e) {
      if (!e.target.closest("nav details")) {
        allDetails.forEach(function (d) { d.removeAttribute("open"); });
      }
    });

    // Mobile nav toggle
    var toggle = document.querySelector(".nav-toggle");
    if (toggle) {
      var menu = toggle.closest("nav").querySelector("ul:last-child");
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
    }
  });
})();
