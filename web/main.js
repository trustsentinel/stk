// Browser glue: load the Go wasm client, wire xterm.js <-> the encrypted session.
(function () {
  "use strict";

  var term = new Terminal({ cursorBlink: true, fontFamily: "ui-monospace, Menlo, monospace", fontSize: 13, theme: { background: "#0d1117" } });
  term.open(document.getElementById("terminal"));
  term.writeln("stk browser client — set the agent key (required) and press Connect.");

  var statusEl = document.getElementById("status");
  var connectBtn = document.getElementById("connect");
  var roomInput = document.getElementById("room");
  var agentInput = document.getElementById("agent");

  // Prefill from URL: ?room=demo&agent=<base64>
  var params = new URLSearchParams(location.search);
  if (params.get("room")) roomInput.value = params.get("room");
  if (params.get("agent")) agentInput.value = params.get("agent");

  function setStatus(s) {
    statusEl.textContent = s;
    statusEl.className = /fail|error|denied/i.test(s) ? "err" : /connected/i.test(s) ? "ok" : "";
  }

  var session = null;

  function connect() {
    if (session) { session.close(); session = null; }
    var hubURL = (location.protocol === "https:" ? "wss://" : "ws://") + location.host + "/ws";
    setStatus("connecting");
    term.writeln("\r\n[connecting to " + hubURL + " room=" + (roomInput.value || "demo") + "]");
    session = window.stkConnect({
      hubURL: hubURL,
      room: roomInput.value || "demo",
      agentPub: agentInput.value.trim(),
      onData: function (u8) { term.write(u8); },
      onStatus: function (s) { setStatus(s); },
      onClose: function () { setStatus("disconnected"); session = null; }
    });
    term.focus();
  }

  term.onData(function (d) { if (session) session.send(d); });
  connectBtn.addEventListener("click", connect);

  // Load and start the wasm module.
  (async function () {
    try {
      var go = new Go();
      var resp = await fetch("stk.wasm");
      var bytes = await resp.arrayBuffer();
      var result = await WebAssembly.instantiate(bytes, go.importObject);
      go.run(result.instance); // sets window.stkConnect, then parks on select{}
      if (typeof window.stkConnect !== "function") throw new Error("wasm did not export stkConnect");
      connectBtn.disabled = false;
      connectBtn.textContent = "Connect";
      setStatus("ready");
    } catch (e) {
      setStatus("load failed");
      term.writeln("\r\n[wasm load error] " + e);
    }
  })();
})();
