/* Goroutines — GET /monigo/api/v1/go-routines-stats
 *
 * The endpoint returns { number_of_goroutines, stack_view[], leak_report? }.
 * leak_report is the grouped, deduplicated view with signatures, blocked
 * minutes and growth; it is nil on builds where leak detection has not run.
 *
 * When it is absent this file derives the same grouping client-side from
 * stack_view, so the page works against today's API and simply becomes more
 * accurate — growth in particular needs history only the server has — once
 * the server-side report is present.
 */
(function () {
  'use strict';

  var open = {};   // signature -> expanded

  // "goroutine 43 [select, 11 minutes]:" — the header carries the state and,
  // when the runtime timestamps the wait, how long it has been blocked.
  var HEADER = /^goroutine\s+(\d+)\s+\[([^,\]]+)(?:,\s*(\d+)\s*minutes?)?\]:/;

  function hash(str) {
    var h = 5381;
    for (var i = 0; i < str.length; i++) { h = ((h << 5) + h + str.charCodeAt(i)) >>> 0; }
    return h.toString(16).padStart(8, '0');
  }

  function topFrame(lines) {
    // First non-header, non-indented line is the deepest frame.
    for (var i = 1; i < lines.length; i++) {
      var l = lines[i];
      if (l && l.charAt(0) !== '\t' && l.trim()) {
        return l.replace(/\(.*$/, '').replace(/^github\.com\/[^/]+\/[^/]+\//, '');
      }
    }
    return 'unknown';
  }

  function groupFromStacks(stacks) {
    var by = {};
    (stacks || []).forEach(function (raw) {
      var lines = String(raw).split('\n');
      var m = HEADER.exec(lines[0] || '');
      var state = m ? m[2] : 'unknown';
      var blocked = m && m[3] ? parseInt(m[3], 10) : 0;
      // Signature over the stack with the per-goroutine header and all
      // hex addresses removed, so two goroutines in the same place agree.
      var body = lines.slice(1).join('\n').replace(/0x[0-9a-f]+/g, '');
      var sig = hash(body);
      if (!by[sig]) {
        by[sig] = {
          signature: sig, state: state, count: 0, blocked_minutes: blocked,
          growth: 0, stale: false, growing: false,
          call_stack: lines.slice(1).join('\n').trim(),
          frame: topFrame(lines)
        };
      }
      by[sig].count++;
      by[sig].blocked_minutes = Math.max(by[sig].blocked_minutes, blocked);
    });
    return Object.keys(by).map(function (k) { return by[k]; });
  }

  function normalise(g) {
    // Server groups carry call_stack but no pre-split top frame.
    if (!g.frame) {
      var first = String(g.call_stack || '').split('\n').find(function (l) {
        return l && l.charAt(0) !== '\t' && l.trim();
      });
      g.frame = first
        ? first.replace(/\(.*$/, '').replace(/^github\.com\/[^/]+\/[^/]+\//, '')
        : 'unknown';
    }
    return g;
  }

  function stateTone(g) {
    if (g.state === 'running' || g.state === 'runnable') { return 'is-ok'; }
    return g.stale || g.blocked_minutes >= 10 ? 'is-warn' : 'is-dim';
  }

  function renderSummary(total, report, groups) {
    var el = document.getElementById('mg-gor-total');
    if (el) { el.textContent = total === undefined ? '—' : total; }
    document.getElementById('mg-gor-groups').textContent = groups.length + ' groups';

    // BY STATE — proportional, derived from the groups either way.
    var byState = {};
    groups.forEach(function (g) { byState[g.state] = (byState[g.state] || 0) + g.count; });
    var states = Object.keys(byState).sort(function (a, b) { return byState[b] - byState[a]; });
    var sum = states.reduce(function (a, k) { return a + byState[k]; }, 0) || 1;
    var palette = ['var(--warn)', 'var(--acc)', 'var(--ok)', 'var(--vio)', 'var(--line2)'];

    document.getElementById('mg-state-bar').innerHTML = states.map(function (k, i) {
      return '<div style="width:' + ((byState[k] / sum) * 100).toFixed(2) + '%;background:' +
        palette[i % palette.length] + '"></div>';
    }).join('');
    document.getElementById('mg-state-legend').innerHTML = states.slice(0, 5).map(function (k, i) {
      return '<div style="display:flex;align-items:center;gap:8px">' +
        '<i style="width:6px;height:6px;border-radius:2px;background:' + palette[i % palette.length] +
        ';display:block;flex:none"></i><span class="is-dim">' + MG.esc(k) + '</span>' +
        '<span class="mg-num" style="margin-left:auto">' + byState[k] + '</span></div>';
    }).join('');

    // LONGEST BLOCKED
    var blocked = groups.filter(function (g) { return g.blocked_minutes > 0; })
      .sort(function (a, b) { return b.blocked_minutes - a.blocked_minutes; }).slice(0, 3);
    document.getElementById('mg-blocked').innerHTML = blocked.length
      ? blocked.map(function (g) {
        return '<div style="display:flex;align-items:center;gap:10px">' +
          '<span class="mg-cell" style="flex:1">' + MG.esc(g.frame) + '</span>' +
          '<span class="mg-num is-warn" style="font:600 var(--t-mono) var(--mono)">' +
          g.blocked_minutes + 'm</span></div>';
      }).join('')
      : '<span class="mg-cell--micro">Nothing blocked over a minute.</span>';

    var thresholdEl = document.getElementById('mg-threshold');
    thresholdEl.textContent = report && report.stale_threshold_minutes
      ? 'stale threshold ' + report.stale_threshold_minutes + 'm'
      : 'stale threshold not reported';

    // Verdict card. Without a server report there is no growth history, so
    // the card says what it can see rather than implying a verdict.
    var verdict = document.getElementById('mg-verdict');
    if (report) {
      verdict.hidden = false;
      verdict.innerHTML =
        '<div style="display:flex;align-items:center;gap:9px">' +
        '<svg width="15" height="15" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.6" style="flex:none">' +
        '<path d="M8 6v3.2M8 11.4h.01"></path><path d="M7 2.3 1.5 12.2a1.15 1.15 0 0 0 1 1.7h11a1.15 1.15 0 0 0 1-1.7L9 2.3a1.15 1.15 0 0 0-2 0z"></path></svg>' +
        '<span style="font-weight:600;font-size:var(--t-body)">' +
        (report.leak_suspected ? 'Leak suspected' : 'No leak detected') + '</span></div>' +
        '<span style="font-size:var(--t-mono);line-height:1.55;color:var(--dim);text-wrap:pretty">' +
        MG.esc(report.message || '') + '</span>' +
        '<span class="mg-cell--micro">' + report.snapshots_retained + ' of ' +
        report.snapshots_required + ' snapshots retained</span>';
      verdict.style.borderColor = report.leak_suspected
        ? 'color-mix(in srgb, var(--warn) 36%, transparent)'
        : 'var(--line)';
    } else {
      verdict.hidden = false;
      verdict.className = 'mg-card--warn';
      verdict.style.background = 'var(--pan)';
      verdict.style.borderColor = 'var(--line)';
      verdict.innerHTML = '<span style="font-weight:600;font-size:var(--t-body)">Grouped locally</span>' +
        '<span style="font-size:var(--t-mono);line-height:1.55;color:var(--dim);text-wrap:pretty">' +
        'This build does not return a leak report, so groups are derived from the live stack dump. ' +
        'Counts, states and blocked time are accurate; growth needs server-side history.</span>';
    }
  }

  function renderTable(groups) {
    var rows = groups.sort(function (a, b) {
      return (b.growth - a.growth) || (b.blocked_minutes - a.blocked_minutes) || (b.count - a.count);
    }).map(function (g) {
      var body = open[g.signature]
        ? '<pre class="mg-pre">' + MG.esc(g.call_stack) + '</pre>'
        : '';
      return '<div>' +
        '<div class="mg-tr mg-tr--click mg-gor-row" data-sig="' + g.signature + '" ' +
        'style="grid-template-columns:104px 1fr 130px 70px 90px 90px">' +
        '<span class="mg-cell mg-cell--micro">' + MG.esc(g.signature.slice(0, 8)) + '</span>' +
        '<span class="mg-cell">' + MG.esc(g.frame) + '</span>' +
        '<span class="mg-cell ' + stateTone(g) + '">' + MG.esc(g.state) + '</span>' +
        '<span class="mg-cell mg-cell--r mg-num" style="font-weight:600">' + g.count + '</span>' +
        '<span class="mg-cell mg-cell--r mg-num ' + (g.blocked_minutes >= 10 ? 'is-warn' : 'is-dim') + '">' +
        (g.blocked_minutes ? g.blocked_minutes + 'm' : '—') + '</span>' +
        '<span class="mg-cell mg-cell--r mg-num ' + (g.growth > 0 ? 'is-bad' : 'is-dim') + '">' +
        (g.growth > 0 ? '+' + g.growth : g.growth || '0') + '</span>' +
        '</div>' + body + '</div>';
    }).join('');

    document.getElementById('mg-gor-rows').innerHTML = rows ||
      '<div class="mg-empty">No goroutines reported.</div>';

    Array.prototype.forEach.call(document.querySelectorAll('.mg-gor-row'), function (row) {
      row.addEventListener('click', function () {
        var sig = row.getAttribute('data-sig');
        open[sig] = !open[sig];
        renderTable(groups);
      });
    });
  }

  document.addEventListener('DOMContentLoaded', function () {
    MG.init();
    MG.Poll.start(function () {
      return MG.request(MG.API + '/go-routines-stats').then(function (data) {
        var report = data.leak_report || null;
        var groups = report && report.suspicious_groups && report.suspicious_groups.length
          ? report.suspicious_groups.map(normalise)
          : groupFromStacks(data.stack_view);

        // A report lists only the offenders. Everything else still has to be
        // visible, so locally-derived groups fill in the rest.
        if (report && report.suspicious_groups && report.suspicious_groups.length) {
          var seen = {};
          groups.forEach(function (g) { seen[g.signature] = true; });
          groupFromStacks(data.stack_view).forEach(function (g) {
            if (!seen[g.signature]) { groups.push(g); }
          });
        }

        renderSummary(data.number_of_goroutines, report, groups);
        renderTable(groups);
      });
    }, 5000);
  });
}());
