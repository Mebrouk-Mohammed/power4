(function(){
  const maxBackoff = 30 * 1000;
  let backoff = 1000;
  let ws = null;

  function connect() {
    const proto = (location.protocol === 'https:') ? 'wss' : 'ws';
    const url = proto + '://' + location.host + '/ws';
    ws = new WebSocket(url);

    ws.onopen = function() {
      backoff = 1000;
      console.log('WS connected');
    };

    ws.onmessage = function(ev) {
      console.log('WS message received:', ev.data);
      // For simplicity, refresh the page on any state update so UI stays in sync
      try {
        const data = JSON.parse(ev.data);
        // optional: you could update DOM in-place instead of reload
        location.reload();
      } catch (e) {
        location.reload();
      }
    };

    ws.onclose = function(evt) {
      console.log('WS closed, reconnecting in', backoff);
      setTimeout(() => {
        backoff = Math.min(maxBackoff, backoff * 2);
        connect();
      }, backoff);
    };

    ws.onerror = function(err) {
      console.error('WS error', err);
      ws.close();
    };
  }

  // Start connection after page loads
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', connect);
  } else {
    connect();
  }
})();
