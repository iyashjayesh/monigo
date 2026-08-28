/* Overview — everything on this page comes from GET /monigo/api/v1/metrics,
 * with the runtime-load chart from POST /service-metrics.
 */
(function () {
  'use strict';

  // Sparkline history is kept client-side. tstorage holds the authoritative
  // series, but it flushes on WithDataPointsSyncFrequency (5m by default), so
  // querying it every 5s would return the same points repeatedly. The tiles
  // want "the last few minutes as observed"; the chart below wants stored
  // history, and asks storage for it.
  var hist = { cpu: [], heap: [], gor: [], gc: [] };
  var CAP = 40;

  function push(key, value) {
    if (typeof value !== 'number' || isNaN(value)) { return; }
    hist[key].push(value);
    if (hist[key].length > CAP) { hist[key].shift(); }
  }

  function el(id) { return document.getElementById(id); }
  function set(id, text) { var n = el(id); if (n) { n.textContent = text; } }
  function html(id, markup) { var n = el(id); if (n) { n.innerHTML = markup; } }

  // health.icon_msg is prose with the numbers embedded:
  //   "System usage exceeds allowed limits: CPU Usage 198.33% / 90.00%, ..."
  // It is the only place the configured thresholds reach the client, so the
  // design's threshold bars read them from here. A parse failure hides the
  // bars rather than inventing a limit.
  function thresholds(iconMsg) {
    var out = [];
    var re = /(CPU|Memory|Goroutines)[^0-9]*([\d.]+)\s*%?\s*\/\s*([\d.]+)/gi;
    var m;
    while ((m = re.exec(iconMsg || '')) !== null) {
      out.push({ label: m[1].toLowerCase(), value: parseFloat(m[2]), limit: parseFloat(m[3]) });
    }
    return out;
  }

  function barRow(t) {
    var ratio = t.limit ? Math.min(1, t.value / t.limit) : 0;
    var over = t.value > t.limit;
    var near = !over && ratio > 0.85;
    var tone = over ? 'var(--bad)' : near ? 'var(--warn)' : 'var(--ok)';
    var unit = t.label === 'goroutines' ? '' : '%';
    return '<div style="display:flex;flex-direction:column;gap:5px">' +
      '<div style="display:flex;justify-content:space-between;font:500 var(--t-mono-s) var(--mono)">' +
      '<span class="is-dim">host ' + MG.esc(t.label) + '</span>' +
      '<span class="mg-num" style="color:' + tone + '">' +
      t.value.toFixed(2) + unit + ' / ' + t.limit.toFixed(2) + unit + '</span></div>' +
      '<div class="mg-bar"><span style="width:' + (ratio * 100).toFixed(1) + '%;background:' + tone + '"></span></div>' +
      '</div>';
  }

  function renderHealth(health) {
    var svc = (health && health.service_health) || {};
    var sys = (health && health.system_health) || {};

    var pct = typeof svc.percent === 'number' ? svc.percent : 0;
    var circ = 2 * Math.PI * 37;
    var ring = el('mg-health-ring');
    if (ring) {
      ring.setAttribute('stroke-dasharray', (circ * pct / 100).toFixed(1) + ' ' + circ.toFixed(1));
      ring.setAttribute('stroke', svc.healthy ? 'var(--ok)' : 'var(--bad)');
    }
    set('mg-health-pct', pct.toFixed(1));
    var verdict = el('mg-health-verdict');
    if (verdict) {
      // message is a long sentence with a bracketed grade at the front —
      // "[Outstanding] The Overall Health is rocking it! ...". The grade is
      // the readable part; the rest is padding at this size.
      var grade = /^\[([^\]]+)\]/.exec(svc.message || '');
      verdict.textContent = grade ? grade[1] : (svc.healthy ? 'Healthy' : 'Unhealthy');
      verdict.className = svc.healthy ? 'is-ok' : 'is-bad';
      verdict.style.fontWeight = '600';
    }
    // icon_msg is the actual measurement, so it becomes the supporting line.
    set('mg-health-detail', svc.icon_msg || '');

    var sysVerdict = el('mg-sys-verdict');
    if (sysVerdict) {
      sysVerdict.textContent = sys.healthy ? 'Healthy' : 'Degraded';
      sysVerdict.className = sys.healthy ? 'is-ok' : 'is-bad';
      sysVerdict.style.fontWeight = '600';
    }
    var bars = thresholds(sys.icon_msg);
    html('mg-sys-bars', bars.length
      ? bars.map(barRow).join('')
      : '<span class="mg-cell--micro">Thresholds unavailable</span>');
  }

  function renderHeap(mem) {
    var rec = MG.indexRecords(mem.mem_stats_records);
    var parts = [
      { key: 'HeapInuse', color: 'var(--vio)' },
      { key: 'Alloc', color: 'var(--acc)' },
      { key: 'HeapIdle', color: 'var(--pan2)' },
      { key: 'HeapReleased', color: 'var(--line2)' }
    ].map(function (p) {
      return { name: p.key, color: p.color, value: MG.recordBytes(rec[p.key]) || 0 };
    });

    var total = parts.reduce(function (a, p) { return a + p.value; }, 0) || 1;
    html('mg-heap-bar', parts.map(function (p) {
      return '<div style="width:' + ((p.value / total) * 100).toFixed(2) + '%;background:' + p.color + '"></div>';
    }).join(''));
    html('mg-heap-legend', parts.map(function (p) {
      return '<div style="display:flex;align-items:center;gap:8px">' +
        '<i style="width:7px;height:7px;border-radius:2px;background:' + p.color + ';display:block;flex:none"></i>' +
        '<span class="is-dim">' + p.name + '</span>' +
        '<span class="mg-num" style="margin-left:auto">' + MG.bytes(p.value) + '</span></div>';
    }).join(''));
  }

  function render(data) {
    var core = data.core_statistics || {};
    var load = data.load_statistics || {};
    var mem = data.memory_statistics || {};
    var cpu = data.cpu_statistics || {};
    var net = data.network_io || {};
    var rec = MG.indexRecords(mem.mem_stats_records);

    var cpuLoad = MG.pct(load.service_cpu_load);
    var heapInuse = MG.recordBytes(rec.HeapInuse);
    var gcPause = mem.gc_pause_duration_ms !== undefined
      ? mem.gc_pause_duration_ms
      : parseFloat(String(mem.gc_pause_duration || '').replace(/[^\d.]/g, ''));

    push('cpu', cpuLoad);
    push('heap', heapInuse);
    push('gor', core.goroutines);
    push('gc', gcPause);

    set('mg-cpu', cpuLoad === null ? '—' : cpuLoad.toFixed(3));
    set('mg-heap', heapInuse ? (heapInuse / 1048576).toFixed(2) : '—');
    set('mg-heap-of', 'of ' + (data.heap_alloc_by_system || '—'));
    set('mg-gor', core.goroutines === undefined ? '—' : core.goroutines);
    set('mg-gc', isNaN(gcPause) ? '—' : Number(gcPause).toFixed(2));
    set('mg-gc-cycles', (MG.recordBytes(rec.NumGC) ? '' : '') +
      (rec.NumGC ? Math.round(rec.NumGC.record_value) + ' cycles' : ''));

    var d = hist.cpu.length > 5 ? hist.cpu[hist.cpu.length - 1] - hist.cpu[hist.cpu.length - 6] : 0;
    set('mg-cpu-delta', (d >= 0 ? '+' : '') + d.toFixed(3));

    html('mg-cpu-spark', MG.sparkline(hist.cpu, { stroke: 'var(--acc)' }));
    html('mg-heap-spark', MG.sparkline(hist.heap, { stroke: 'var(--vio)' }));
    html('mg-gor-spark', MG.sparkline(hist.gor, { stroke: 'var(--warn)' }));
    html('mg-gc-spark', MG.sparkline(hist.gc, { stroke: 'var(--ok)' }));

    set('mg-uptime', core.uptime || '—');
    set('mg-cores', cpu.total_cores === undefined ? '—' : cpu.total_cores);
    set('mg-net-sent', MG.bytes(net.bytes_sent));
    set('mg-net-recv', MG.bytes(net.bytes_received));
    set('mg-stack-mem', mem.stack_memory_usage || '—');
    set('mg-swap-free', mem.free_swap_memory || '—');

    renderHealth(data.health);
    renderHeap(mem);

    // The leak verdict belongs on Overview too — it is the one thing on the
    // Goroutines page an operator must not have to go looking for.
    var leak = data.goroutine_leak_report;
    var box = el('mg-leak');
    if (box) {
      if (leak && leak.leak_suspected) {
        box.hidden = false;
        box.className = 'mg-note mg-note--warn';
        box.innerHTML = '<span>' + MG.esc(leak.message || 'Goroutine leak suspected.') + '</span>' +
          '<a class="mg-note__act" href="goroutines.html">Inspect</a>';
      } else {
        box.hidden = true;
      }
    }
  }

  /* Runtime load chart, from stored history rather than the live poll. */
  function loadChart() {
    var end = new Date();
    var start = new Date(end.getTime() - 60 * 60 * 1000);
    return MG.post(MG.API + '/service-metrics', {
      start_time: start.toISOString(),
      end_time: end.toISOString(),
      field_name: ['service_cpu_load', 'heap_inuse', 'goroutines']
    }).then(function (rows) {
      if (!rows || !rows.length) {
        // No stored points yet is the ordinary state for the first few
        // minutes of a process, so it gets an explanation, not an error.
        html('mg-chart', '<div class="mg-empty">No stored history yet. ' +
          'Points are written on each sync interval.</div>');
        return;
      }
      var series = { service_cpu_load: [], HeapInuse: [], goroutines: [] };
      rows.forEach(function (r) {
        var v = r.value || {};
        Object.keys(series).forEach(function (k) {
          if (typeof v[k] === 'number') { series[k].push(v[k]); }
        });
      });
      var W = 600, H = 262;
      var grid = [1, 65, 131, 196, 261].map(function (y) {
        return '<line x1="0" y1="' + y + '" x2="' + W + '" y2="' + y + '" stroke="var(--line)" stroke-width="1"></line>';
      }).join('');
      var paths = '';
      if (series.service_cpu_load.length > 1) {
        paths += '<path d="' + MG.areaPath(series.service_cpu_load, W, H, 12) + '" fill="var(--acc)" opacity=".10"></path>' +
          '<path d="' + MG.linePath(series.service_cpu_load, W, H, 12) + '" fill="none" stroke="var(--acc)" stroke-width="1.8" vector-effect="non-scaling-stroke"></path>';
      }
      if (series.HeapInuse.length > 1) {
        paths += '<path d="' + MG.linePath(series.HeapInuse, W, H, 12) + '" fill="none" stroke="var(--vio)" stroke-width="1.6" vector-effect="non-scaling-stroke"></path>';
      }
      if (series.goroutines.length > 1) {
        paths += '<path d="' + MG.linePath(series.goroutines, W, H, 12) + '" fill="none" stroke="var(--warn)" stroke-width="1.4" stroke-dasharray="4 3" vector-effect="non-scaling-stroke"></path>';
      }
      html('mg-chart', '<svg viewBox="0 0 ' + W + ' ' + H + '" preserveAspectRatio="none" ' +
        'style="width:100%;height:' + H + 'px;display:block">' + grid + paths + '</svg>');
    }).catch(function () {
      html('mg-chart', '<div class="mg-empty">History unavailable.</div>');
    });
  }

  document.addEventListener('DOMContentLoaded', function () {
    MG.init();
    MG.Poll.start(function () {
      return MG.request(MG.API + '/metrics').then(render);
    }, 5000);
    loadChart();
    setInterval(loadChart, 60000);
  });
}());
