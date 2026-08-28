/* MoniGo — shared dashboard runtime.
 *
 * Supersedes the shell half of common.js: same element ids (mg-service-name,
 * mg-avatar, mg-status-dot, mg-count-*, mg-storage-*), so either file can
 * drive the chrome. Load ONE of them, not both, or the shell is fetched twice.
 *
 * What this adds over common.js:
 *   - one poll loop for the whole page instead of per-widget fetches
 *   - a connection state machine (ok / stale / down) that every page renders
 *     the same way, replacing the 300s location.reload()
 *   - in-place updates, so investigation state survives a refresh
 */
(function (global) {
  'use strict';

  var API = '/monigo/api/v1';

  /* ---------------------------------------------------------------- auth */

  function apiKey() {
    return new URLSearchParams(global.location.search).get('api_key');
  }

  // Mirrors common.js authenticatedFetch so a migrated page keeps working
  // under every auth mode the middleware supports.
  function request(url, options) {
    options = options || {};
    var key = apiKey();
    if (key) {
      url += (url.indexOf('?') === -1 ? '?' : '&') + 'api_key=' + encodeURIComponent(key);
    } else {
      var secret = new URLSearchParams(global.location.search).get('secret');
      if (secret === 'monigo-admin-secret') {
        url += (url.indexOf('?') === -1 ? '?' : '&') + 'secret=' + encodeURIComponent(secret);
      } else {
        options.headers = options.headers || {};
        options.headers['X-User-Role'] = 'admin';
      }
    }
    return fetch(url, options).then(function (r) {
      if (!r.ok) { throw new Error('HTTP ' + r.status); }
      return r.json();
    });
  }

  function post(url, body) {
    return request(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });
  }

  /* ---------------------------------------------------------- formatting */

  // The API returns two shapes for the same quantity: a pre-formatted string
  // carrying its unit ("865.34 KB") and a raw number. Never parse the former —
  // a value in MB and a value in GB would land on the same axis.
  function pct(str) {
    var n = parseFloat(String(str || '').replace('%', ''));
    return isNaN(n) ? null : n;
  }

  function bytes(n) {
    if (n === null || n === undefined || isNaN(n)) { return '—'; }
    var u = ['B', 'KB', 'MB', 'GB', 'TB'], i = 0;
    while (n >= 1024 && i < u.length - 1) { n /= 1024; i++; }
    return (n < 10 ? n.toFixed(2) : n.toFixed(1)) + ' ' + u[i];
  }

  // mem_stats_records carry an explicit record_unit, which makes them the safe
  // source for anything plotted or compared.
  function recordBytes(rec) {
    if (!rec) { return null; }
    var mult = { bytes: 1, KB: 1024, MB: 1048576, GB: 1073741824 }[rec.record_unit] || 1;
    return Number(rec.record_value) * mult;
  }

  function indexRecords(list) {
    var out = {};
    (list || []).forEach(function (r) { out[r.record_name] = r; });
    return out;
  }

  // execution_time arrives as a Go time.Duration: an integer count of ns.
  function duration(ns) {
    if (!ns && ns !== 0) { return '—'; }
    if (ns < 1000) { return ns + ' ns'; }
    if (ns < 1e6) { return (ns / 1000).toFixed(1) + ' µs'; }
    if (ns < 1e9) { return (ns / 1e6).toFixed(1) + ' ms'; }
    return (ns / 1e9).toFixed(2) + ' s';
  }

  function ago(iso) {
    if (!iso) { return '—'; }
    var t = new Date(iso).getTime();
    if (isNaN(t)) { return '—'; }
    var s = Math.max(0, Math.round((Date.now() - t) / 1000));
    if (s < 60) { return s + 's ago'; }
    if (s < 3600) { return Math.round(s / 60) + 'm ago'; }
    if (s < 86400) { return Math.round(s / 3600) + 'h ago'; }
    return Math.round(s / 86400) + 'd ago';
  }

  function clock(d) {
    return (d || new Date()).toISOString().slice(11, 19) + 'Z';
  }

  function esc(s) {
    return String(s === null || s === undefined ? '' : s)
      .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  }

  /* -------------------------------------------------------------- charts */

  // Inline SVG rather than ECharts: these are single-series lines, and the
  // dashboard already carries a 1 MB chart library for four of them.
  function linePath(values, w, h, pad) {
    var v = (values || []).filter(function (n) { return typeof n === 'number' && !isNaN(n); });
    if (v.length < 2) { return ''; }
    pad = pad === undefined ? 3 : pad;
    var lo = Math.min.apply(null, v), hi = Math.max.apply(null, v), span = (hi - lo) || 1;
    return v.map(function (n, i) {
      var x = (i / (v.length - 1)) * w;
      var y = h - pad - ((n - lo) / span) * (h - pad * 2);
      return (i ? 'L' : 'M') + x.toFixed(1) + ' ' + y.toFixed(1);
    }).join(' ');
  }

  function areaPath(values, w, h, pad) {
    var d = linePath(values, w, h, pad);
    return d ? d + ' L' + w + ' ' + h + ' L0 ' + h + ' Z' : '';
  }

  function sparkline(values, opts) {
    opts = opts || {};
    var w = opts.width || 240, h = opts.height || 44, stroke = opts.stroke || 'var(--acc)';
    var d = linePath(values, w, h, 4);
    if (!d) {
      return '<svg class="mg-spark" viewBox="0 0 ' + w + ' ' + h + '" preserveAspectRatio="none"></svg>';
    }
    return '<svg class="mg-spark" viewBox="0 0 ' + w + ' ' + h + '" preserveAspectRatio="none">' +
      '<path d="' + areaPath(values, w, h, 4) + '" fill="' + stroke + '" opacity=".14"></path>' +
      '<path d="' + d + '" fill="none" stroke="' + stroke + '" stroke-width="1.6" vector-effect="non-scaling-stroke"></path>' +
      '</svg>';
  }

  /* --------------------------------------------------------------- theme */

  function setTheme(theme) {
    document.body.setAttribute('data-theme', theme);
    try { localStorage.setItem('monigo-theme', theme); } catch (e) { /* private mode */ }
    document.dispatchEvent(new CustomEvent('monigoThemeChanged', { detail: { theme: theme } }));
  }

  function initTheme() {
    var saved = 'dark';
    try { saved = localStorage.getItem('monigo-theme') || 'dark'; } catch (e) { /* ignore */ }
    setTheme(saved);
    var btn = document.getElementById('mg-theme');
    if (btn) {
      btn.addEventListener('click', function () {
        setTheme(document.body.getAttribute('data-theme') === 'dark' ? 'light' : 'dark');
      });
    }
  }

  /* --------------------------------------------------- connection status */

  var status = { state: 'ok', fails: 0, lastGood: null };

  function renderStatus() {
    var dot = document.getElementById('mg-status-dot');
    if (dot) {
      dot.classList.remove('is-stale', 'is-down');
      if (status.state !== 'ok') { dot.classList.add(status.state === 'down' ? 'is-down' : 'is-stale'); }
      dot.title = status.state === 'ok' ? 'Connected'
        : status.state === 'down' ? 'Lost connection to the service' : 'Data is stale';
    }

    var content = document.getElementById('mg-content');
    if (content) { content.classList.toggle('is-stale', status.state !== 'ok'); }

    var banner = document.getElementById('mg-banner');
    if (!banner) { return; }
    if (status.state === 'ok') {
      banner.innerHTML = '';
      banner.hidden = true;
      return;
    }
    banner.hidden = false;
    var since = status.lastGood ? clock(status.lastGood) : 'never';
    banner.className = 'mg-note ' + (status.state === 'down' ? 'mg-note--bad' : 'mg-note--warn');
    banner.innerHTML =
      '<svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.6">' +
      '<circle cx="8" cy="8" r="6.3"></circle><path d="M8 5.4v3.4M8 11.4h.01"></path></svg>' +
      '<span><b>' + (status.state === 'down' ? 'No response from the service.' : 'Data is stale.') +
      '</b> Showing the last snapshot from ' + since + '. ' +
      (status.fails > 1 ? status.fails + ' failed polls.' : 'Retrying…') + '</span>' +
      '<button class="mg-note__act" id="mg-retry">Retry now</button>';
    var retry = document.getElementById('mg-retry');
    if (retry) { retry.addEventListener('click', function () { Poll.tick(true); }); }
  }

  function markOk() {
    status.fails = 0;
    status.lastGood = new Date();
    if (status.state !== 'ok') { status.state = 'ok'; renderStatus(); }
    var el = document.getElementById('mg-asof');
    if (el) { el.textContent = 'as of ' + clock(status.lastGood); }
  }

  function markFail() {
    status.fails++;
    // One miss is a blip; the design only escalates to "down" once a retry has
    // also failed, so a single dropped poll does not flash an alarm.
    var next = status.fails >= 2 ? 'down' : 'stale';
    if (next !== status.state) { status.state = next; }
    renderStatus();
  }

  /* ---------------------------------------------------------- poll loop */

  var Poll = {
    fn: null, interval: 5000, timer: null, running: true,

    start: function (fn, interval) {
      this.fn = fn;
      this.interval = interval || 5000;
      this.tick(true);
      this.timer = setInterval(function () { Poll.tick(false); }, this.interval);

      var btn = document.getElementById('mg-live');
      if (btn) {
        btn.addEventListener('click', function () {
          Poll.running = !Poll.running;
          btn.classList.toggle('is-paused', !Poll.running);
          var label = btn.querySelector('span');
          if (label) { label.textContent = Poll.running ? 'LIVE' : 'PAUSED'; }
          if (Poll.running) { Poll.tick(true); }
        });
      }
      // Polling a hidden tab burns the profiled process's CPU for nothing.
      document.addEventListener('visibilitychange', function () {
        if (!document.hidden && Poll.running) { Poll.tick(false); }
      });
    },

    tick: function (force) {
      if (!this.fn) { return; }
      if (!force && (!this.running || document.hidden)) { return; }
      this.fn().then(markOk).catch(markFail);
    }
  };

  /* ----------------------------------------------------------- the shell */

  function fillShell() {
    return request(API + '/service-info').then(function (info) {
      var name = info.service_name || 'service';
      var set = function (id, text, title) {
        var el = document.getElementById(id);
        if (el) { el.textContent = text; if (title) { el.title = title; } }
      };
      set('mg-service-name', name, name);
      set('mg-avatar', name.slice(0, 2));
      var go = String(info.go_version || '').replace(/^go/, '');
      set('mg-service-meta', 'pid ' + (info.process_id || '—') + (go ? ' · go' + go : ''));

      // Omitted by the API when unset, so absent means "not configured",
      // never zero.
      var parts = [info.storage_type, info.retention_period].filter(Boolean);
      set('mg-storage-mode', parts.join(' · '));
      set('mg-storage-value', info.storage_on_disk
        ? info.storage_on_disk + ' retained'
        : (info.storage_type === 'memory' ? 'in memory' : ''));
      return info;
    });
  }

  function markActiveNav() {
    var here = location.pathname.split('/').pop() || 'index.html';
    Array.prototype.forEach.call(document.querySelectorAll('.mg-nav a'), function (a) {
      var href = a.getAttribute('href');
      if (href && href.split('/').pop() === here) { a.classList.add('is-active'); }
    });
  }

  function init() {
    initTheme();
    markActiveNav();
    fillShell().catch(function () { status.state = 'down'; renderStatus(); });
  }

  global.MG = {
    API: API,
    request: request, post: post, init: init, Poll: Poll,
    pct: pct, bytes: bytes, recordBytes: recordBytes, indexRecords: indexRecords,
    duration: duration, ago: ago, clock: clock, esc: esc,
    linePath: linePath, areaPath: areaPath, sparkline: sparkline,
    status: status
  };
}(window));
