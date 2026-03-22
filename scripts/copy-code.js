// Adds copy-to-clipboard buttons to all <pre><code> blocks.

(function () {
  document.addEventListener("DOMContentLoaded", function () {
    document.querySelectorAll("pre").forEach(function (pre) {
      var code = pre.querySelector("code");
      if (!code) return;

      var btn = document.createElement("button");
      btn.className = "copy-btn";
      btn.textContent = "Copy";
      btn.addEventListener("click", function () {
        navigator.clipboard.writeText(code.textContent).then(function () {
          btn.textContent = "Copied!";
          setTimeout(function () {
            btn.textContent = "Copy";
          }, 2000);
        });
      });
      pre.appendChild(btn);
    });
  });
})();
