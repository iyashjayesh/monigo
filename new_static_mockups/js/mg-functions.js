/* Functions — GET /monigo/api/v1/function returns a map of function name to
 * models.FunctionMetrics, and GET /function-details?name=&reportType=text
 * returns the pprof text for one of them.
 *
 * The canvas design showed P50 / P95 / call-count / memory-delta columns.
 * FunctionMetrics records a single last-run sample per function — no
 * histogram, no call counter, no per-function series — so those columns are
 * not drawn here. Showing a last-run figure under a "P95" heading would be a
 * lie the operator cannot detect. They come back when the collector does.
 */
(function () {
  'use strict';

  var selected = null;

  function sortRows(map) {
    return Object.keys(map || {}).map(function (name) {
      var m = map[name] || {};
      return {
        name: name,
        exec: Number(m.execution_time) || 0,
        mem: Number(m.memory_usage) || 0,
        gor: m.goroutine_count,
        last: m.function_last_ran_at
      };
    }).sort(function (a, b) { return b.exec - a.exec; });
  }

  function renderTable(rows) {
    document.getElementById('mg-fn-count').textContent =
      rows.length + (rows.length === 1 ? ' function traced' : ' functions traced');

    if (!rows.length) {
      document.getElementById('mg-fn-rows').innerHTML =
        '<div class="mg-empty">No traced functions yet.<br>' +
        'Wrap a call with <code>monigo.TraceFunction(ctx, fn)</code> and it appears here.</div>';
      return;
    }

    var slowest = rows[0].exec || 1;
    document.getElementById('mg-fn-rows').innerHTML = rows.map(function (r) {
      var share = Math.min(1, r.exec / slowest);
      var tone = share > 0.66 ? 'is-bad' : share > 0.33 ? 'is-warn' : '';
      return '<div class="mg-tr mg-tr--click mg-fn-row" data-name="' + MG.esc(r.name) + '" ' +
        'style="grid-template-columns:1fr 110px 110px 100px 120px">' +
        '<span class="mg-cell">' + MG.esc(r.name) + '</span>' +
        '<span class="mg-cell mg-cell--r mg-num ' + tone + '" style="font-weight:600">' +
        MG.duration(r.exec) + '</span>' +
        '<span class="mg-cell mg-cell--r mg-num mg-cell--dim">' + MG.bytes(r.mem) + '</span>' +
        '<span class="mg-cell mg-cell--r mg-num mg-cell--dim">' +
        (r.gor === undefined ? '—' : r.gor) + '</span>' +
        '<span class="mg-cell mg-cell--r mg-cell--micro">' + MG.ago(r.last) + '</span>' +
        '</div>';
    }).join('');

    Array.prototype.forEach.call(document.querySelectorAll('.mg-fn-row'), function (row) {
      row.addEventListener('click', function () { openDetail(row.getAttribute('data-name')); });
    });
  }

  function openDetail(name) {
    selected = name;
    var panel = document.getElementById('mg-fn-detail');
    panel.hidden = false;
    document.getElementById('mg-fn-name').textContent = name;
    document.getElementById('mg-fn-cpu').innerHTML = '<span class="mg-skel">loading profile…</span>';
    document.getElementById('mg-fn-mem').innerHTML = '';
    document.getElementById('mg-fn-code').textContent = '';

    MG.request(MG.API + '/function-details?name=' + encodeURIComponent(name) + '&reportType=text')
      .then(function (d) {
        if (selected !== name) { return; }
        var p = d.core_profile || {};
        document.getElementById('mg-fn-cpu').textContent = p.cpu_profile || 'No CPU profile captured.';
        document.getElementById('mg-fn-mem').textContent = p.mem_profile || 'No memory profile captured.';
        document.getElementById('mg-fn-code').textContent = d.function_code_trace || '';
      })
      .catch(function () {
        if (selected !== name) { return; }
        document.getElementById('mg-fn-cpu').textContent = 'Profile unavailable.';
      });
  }

  document.addEventListener('DOMContentLoaded', function () {
    MG.init();
    document.getElementById('mg-fn-close').addEventListener('click', function () {
      selected = null;
      document.getElementById('mg-fn-detail').hidden = true;
    });
    MG.Poll.start(function () {
      return MG.request(MG.API + '/function').then(function (map) {
        renderTable(sortRows(map));
      });
    }, 10000);
  });
}());
