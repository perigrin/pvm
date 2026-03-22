// ABOUTME: Platform tab switcher for the install block on the homepage.
// ABOUTME: Auto-detects visitor platform and activates the matching tab.

(function () {
  document.addEventListener("DOMContentLoaded", function () {
    var tablist = document.querySelector(".platform-tabs");
    if (!tablist) return;

    var tabs = tablist.querySelectorAll("[role='tab']");
    var panels = [];
    tabs.forEach(function (tab) {
      panels.push(document.getElementById(tab.getAttribute("aria-controls")));
    });

    function activate(index) {
      tabs.forEach(function (t, i) {
        var selected = i === index;
        t.setAttribute("aria-selected", String(selected));
        t.tabIndex = selected ? 0 : -1;
        panels[i].hidden = !selected;
      });
    }

    tabs.forEach(function (tab, i) {
      tab.addEventListener("click", function () { activate(i); });
      tab.addEventListener("keydown", function (e) {
        var dir = 0;
        if (e.key === "ArrowRight") dir = 1;
        if (e.key === "ArrowLeft") dir = -1;
        if (dir) {
          e.preventDefault();
          var next = (i + dir + tabs.length) % tabs.length;
          tabs[next].focus();
          activate(next);
        }
      });
    });

    // Auto-detect platform and select the matching tab
    var ua = navigator.userAgent.toLowerCase();
    if (ua.indexOf("mac") !== -1) {
      activate(2); // macOS ARM64
    } else if (ua.indexOf("aarch64") !== -1 || ua.indexOf("arm64") !== -1) {
      activate(1); // Linux ARM64
    }
    // Default is Linux AMD64 (index 0), already selected
  });
})();
