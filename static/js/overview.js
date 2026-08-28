/*
 * Overview page renderer.
 *
 * Replaces the metric-tile and chart half of index.js, and the 300-second
 * `location.reload()` in refresh.js. The design's LIVE indicator implies
 * in-place updates, so a full page reload would make it a lie -- and reloading
 * also threw away expanded state and navigated to the browser error page at
 * exactly the moment the monitored service died.
 *
 * Every value drawn here is traced to a real API field. Where the design shows
 * something MoniGo does not measure, the element is omitted rather than filled
 * with a plausible number: see DESIGN-ADOPTION.md.
 */
document.addEventListener('DOMContentLoaded', () => {
    'use strict';

    // ------------------------------------------------------------ transport

    /*
     * common.js defines these inside its own DOMContentLoaded closure, so they
     * are not visible here. Duplicated rather than exported, because making
     * them global would put the API key on `window` where any script on the
     * page could read it.
     */
    function getApiKey() {
        const params = new URLSearchParams(window.location.search);
        return params.get('api_key') || window.MONIGO_API_KEY || null;
    }

    function authenticatedFetch(url, options = {}) {
        const key = getApiKey();
        if (!key) {
            return fetch(url, options);
        }
        const headers = Object.assign({}, options.headers, { 'X-API-Key': key });
        return fetch(url, Object.assign({}, options, { headers }));
    }

    // ---------------------------------------------------------------- guards

    /*
     * Several fields on this page are unit-suffixed display strings that sit
     * next to a near-identically named numeric twin -- memory_used_by_service
     * beside memory_used_by_service_bytes, and so on. parseFloat("1.2 GB")
     * returns 1.2 and silently plots gigabytes against megabytes. That bug
     * shipped once already.
     *
     * num() makes it impossible by construction rather than by review: a
     * string carrying anything other than digits is refused outright.
     */
    function num(value, where) {
        if (typeof value === 'number') {
            return isFinite(value) ? value : NaN;
        }
        if (typeof value !== 'string') {
            return NaN;
        }
        if (!/^-?\d+(\.\d+)?$/.test(value.trim())) {
            console.error(
                `[monigo] refusing to parse ${JSON.stringify(value)} as a number` +
                (where ? ` for ${where}` : '') +
                '. It carries a unit suffix; use the *_bytes or raw field instead.'
            );
            return NaN;
        }
        return Number(value);
    }

    // A percent string ("0.26%") has no numeric twin -- ServiceCPULoadRaw is
    // json:"-" -- so this is the one suffix we strip, and only this one.
    function percent(value) {
        return num(String(value == null ? '' : value).replace(/%\s*$/, ''), 'percent');
    }

    // -------------------------------------------------------------- format

    function formatBytes(bytes) {
        if (!isFinite(bytes) || bytes < 0) {
            return '—';
        }
        const units = ['B', 'KB', 'MB', 'GB', 'TB'];
        let v = bytes;
        let i = 0;
        while (v >= 1024 && i < units.length - 1) {
            v /= 1024;
            i++;
        }
        return trimZeros(v.toFixed(2)) + ' ' + units[i];
    }

    // raw_mem_stats_records are base-1024 KB, not bytes (core.go newRawRecord).
    function formatKB(kb) {
        if (!isFinite(kb)) {
            return '—';
        }
        return kb >= 1024
            ? trimZeros((kb / 1024).toFixed(2)) + ' MB'
            : trimZeros(kb.toFixed(2)) + ' KB';
    }

    function trimZeros(s) {
        return String(s).replace(/\.00$/, '');
    }

    function set(id, text) {
        const el = document.getElementById(id);
        if (el) {
            el.textContent = text;
        }
    }

    function severityClass(el, level) {
        if (!el) {
            return;
        }
        el.classList.remove('mg-sev-ok', 'mg-sev-warn', 'mg-sev-crit');
        el.classList.add('mg-sev-' + level);
    }

    // --------------------------------------------------------------- state

    const RING_CIRCUMFERENCE = 232.5;   // 2 * PI * r, r = 37
    const state = {
        live: true,
        range: '15m',
        timer: null,
        lastGood: null,
        syncFrequency: '5m',
    };

    // -------------------------------------------------------------- health

    function renderHealth(m) {
        const svc = (m.health && m.health.service_health) || {};
        const sys = (m.health && m.health.system_health) || {};

        const pct = num(svc.percent, 'service_health.percent');
        if (isFinite(pct)) {
            set('mg-ring-value', pct.toFixed(1));
            const arc = document.getElementById('mg-ring-arc');
            if (arc) {
                const filled = Math.max(0, Math.min(100, pct)) / 100 * RING_CIRCUMFERENCE;
                arc.setAttribute('stroke-dasharray', filled.toFixed(1) + ' ' + RING_CIRCUMFERENCE);
                severityClass(arc, pct >= 70 ? 'ok' : pct >= 50 ? 'warn' : 'crit');
            }
        }

        renderState('mg-svc-state', svc);
        renderState('mg-sys-state', sys);
        renderState('mg-topbar-state', svc);

        const stateEl = document.getElementById('mg-svc-state');
        if (stateEl && svc.message) {
            stateEl.title = svc.message;
        }

        // The service detail line and both threshold bars are scraped from
        // icon_msg, which is the only place the configured thresholds surface:
        // MaxCPUUsage and friends are never serialised on their own.
        const svcMsg = svc.icon_msg || '';
        const cpu = /CPU Usage ([\d.]+)% \/ ([\d.]+)%/.exec(svcMsg);
        const mem = /Memory Usage ([\d.]+)% \/ ([\d.]+)%/.exec(svcMsg);
        const gor = /Goroutines ([\d.]+) \/ (\d+)/.exec(svcMsg);
        const parts = [];
        if (cpu) {
            parts.push(`CPU ${(+cpu[1]).toFixed(2)}% / ${(+cpu[2]).toFixed(0)}%`);
        }
        if (mem) {
            parts.push(`Memory ${(+mem[1]).toFixed(2)}% / ${(+mem[2]).toFixed(0)}%`);
        }
        if (gor) {
            parts.push(`Goroutines ${Math.round(+gor[1])} / ${gor[2]}`);
        }
        set('mg-svc-detail', parts.join('  ·  '));

        const sysMsg = sys.icon_msg || '';
        bar('cpu', /CPU Usage ([\d.]+)% \/ ([\d.]+)%/.exec(sysMsg));
        bar('mem', /Memory Usage ([\d.]+)% \/ ([\d.]+)%/.exec(sysMsg));

        const dot = document.getElementById('mg-service-dot');
        if (dot) {
            dot.classList.remove('is-unhealthy', 'is-degraded');
            if (svc.healthy === false) {
                dot.classList.add('is-unhealthy');
            } else if (isFinite(pct) && pct < 90) {
                dot.classList.add('is-degraded');
            }
        }
    }

    function renderState(id, health) {
        const el = document.getElementById(id);
        if (!el) {
            return;
        }
        const healthy = health.healthy !== false;
        el.textContent = healthy ? 'Healthy' : 'Degraded';
        severityClass(el, healthy ? 'ok' : 'crit');
    }

    // Draws one threshold bar. The fill is the usage as a proportion of the
    // limit, so a bar that reaches the end IS the breach -- no separate marker
    // needed, and a value over 100% clamps visually while still reading true.
    function bar(which, match) {
        const valueEl = document.getElementById(`mg-bar-${which}-value`);
        const fillEl = document.getElementById(`mg-bar-${which}-fill`);
        if (!match) {
            if (valueEl) {
                valueEl.textContent = '—';
            }
            if (fillEl) {
                fillEl.style.width = '0%';
            }
            return;
        }
        const used = +match[1];
        const limit = +match[2];
        if (valueEl) {
            valueEl.textContent = `${used.toFixed(2)}% / ${limit.toFixed(2)}%`;
            severityClass(valueEl, used >= limit ? 'crit' : used >= limit * 0.8 ? 'warn' : 'ok');
        }
        if (fillEl) {
            const ratio = limit > 0 ? used / limit : 0;
            fillEl.style.width = Math.max(0, Math.min(1, ratio)) * 100 + '%';
            severityClass(fillEl, ratio >= 1 ? 'crit' : ratio >= 0.8 ? 'warn' : 'ok');
        }
    }

    // ----------------------------------------------------------- KPI tiles

    function rawRecords(m) {
        const out = {};
        const list = (m.memory_statistics && m.memory_statistics.raw_mem_stats_records) || [];
        list.forEach(r => {
            out[r.record_name] = r.record_value;
        });
        return out;
    }

    function renderKPIs(m) {
        const raw = rawRecords(m);

        const cpu = percent(m.load_statistics && m.load_statistics.service_cpu_load);
        set('mg-kpi-cpu-value', isFinite(cpu) ? cpu.toFixed(3) : '—');

        // heap_inuse / heap_sys are base-1024 KB.
        const inuse = num(raw.heap_inuse, 'heap_inuse');
        const sys = num(raw.heap_sys, 'heap_sys');
        set('mg-kpi-heap-value', isFinite(inuse) ? (inuse / 1024).toFixed(2) : '—');
        set('mg-kpi-heap-total', isFinite(sys) ? 'of ' + (sys / 1024).toFixed(2) : '');

        const goroutines = m.core_statistics && m.core_statistics.goroutines;
        set('mg-kpi-gor-value', goroutines == null ? '—' : String(goroutines));

        // Stale goroutines are only worth a word when there are some.
        const stale = (m.goroutine_leak_report && m.goroutine_leak_report.stale_goroutines) || 0;
        const staleEl = document.getElementById('mg-kpi-gor-stale');
        if (staleEl) {
            staleEl.textContent = stale > 0 ? stale + ' stale' : '';
            staleEl.classList.toggle('is-warn', stale > 0);
        }

        // gc_pause_duration_ms is PauseTotalNs/1e6 -- cumulative since start,
        // which is why the tile is labelled GC PAUSE TOTAL rather than
        // presenting a running total as a latest pause.
        const gc = num(m.memory_statistics && m.memory_statistics.gc_pause_duration_ms, 'gc_pause_duration_ms');
        set('mg-kpi-gc-value', isFinite(gc) ? gc.toFixed(2) : '—');
        // num_gc is a count, not a byte metric, so it is not scaled by 1024.
        const cycles = num(raw.num_gc, 'num_gc');
        set('mg-kpi-gc-cycles', isFinite(cycles) ? Math.round(cycles) + ' cycles' : '');
    }

    // -------------------------------------------------------- heap panel

    function renderHeap(m) {
        const raw = rawRecords(m);
        const inuse = num(raw.heap_inuse, 'heap_inuse');
        const alloc = num(raw.alloc, 'alloc');
        const idle = num(raw.heap_idle, 'heap_idle');
        const released = num(raw.heap_released, 'heap_released');

        // heap_released is a subset of heap_idle, so it is legend-only and must
        // not enter the denominator or the bar would sum past 100%.
        const total = [inuse, alloc, idle].reduce((a, b) => a + (isFinite(b) ? b : 0), 0);
        const seg = (id, v) => {
            const el = document.getElementById(id);
            if (el) {
                el.style.width = total > 0 && isFinite(v) ? (v / total * 100).toFixed(2) + '%' : '0%';
            }
        };
        seg('mg-heap-seg-inuse', inuse);
        seg('mg-heap-seg-alloc', alloc);
        seg('mg-heap-seg-idle', idle);

        set('mg-heap-val-inuse', formatKB(inuse));
        set('mg-heap-val-alloc', formatKB(alloc));
        set('mg-heap-val-idle', formatKB(idle));
        set('mg-heap-val-released', formatKB(released));

        const io = m.network_io || {};
        set('mg-readout-net-sent', formatBytes(num(io.bytes_sent, 'bytes_sent')));
        set('mg-readout-net-recv', formatBytes(num(io.bytes_received, 'bytes_received')));

        const mem = m.memory_statistics || {};
        set('mg-readout-stack', formatBytes(num(mem.stack_memory_usage_bytes, 'stack_memory_usage_bytes')));
        // free_swap_memory is the one field with no numeric twin. Printed
        // verbatim; never parsed.
        set('mg-readout-swap-free', mem.free_swap_memory ? trimZeros(mem.free_swap_memory) : '—');
    }

    function renderReadouts(m, info) {
        set('mg-readout-uptime', (m.core_statistics && m.core_statistics.uptime) || '—');
        const cores = num(m.cpu_statistics && m.cpu_statistics.total_cores, 'total_cores');
        set('mg-readout-cores', isFinite(cores) ? String(Math.round(cores)) : '—');
        if (info) {
            set('mg-readout-pid', info.process_id == null ? '—' : String(info.process_id));
            set('mg-readout-go', (info.go_version || '—').replace(/^go/, ''));
        }
    }

    // ------------------------------------------------------------ freshness

    // The stamp is the time of the last SUCCESSFUL poll, frozen on failure, so
    // a dead service shows an ageing timestamp rather than a ticking clock
    // that implies the data is current.
    function stampNow() {
        state.lastGood = new Date();
        set('mg-svc-stamp', 'last check ' + state.lastGood.toTimeString().slice(0, 8));
    }

    function markStale(err) {
        const dot = document.getElementById('mg-service-dot');
        if (dot) {
            dot.classList.add('is-unhealthy');
        }
        const stamp = document.getElementById('mg-svc-stamp');
        if (stamp && state.lastGood) {
            const age = Math.round((Date.now() - state.lastGood.getTime()) / 1000);
            stamp.textContent = `last check ${state.lastGood.toTimeString().slice(0, 8)} · ${age}s ago`;
            stamp.classList.add('is-stale');
        } else if (stamp) {
            stamp.textContent = 'no data';
        }
        console.error('[monigo] metrics poll failed:', err);
    }

    // -------------------------------------------------------------- polling

    function poll() {
        Promise.all([
            authenticatedFetch('/monigo/api/v1/metrics').then(r => {
                if (!r.ok) {
                    throw new Error('metrics HTTP ' + r.status);
                }
                return r.json();
            }),
            authenticatedFetch('/monigo/api/v1/service-info').then(r => r.ok ? r.json() : null).catch(() => null),
        ])
            .then(([m, info]) => {
                renderHealth(m);
                renderKPIs(m);
                renderHeap(m);
                renderReadouts(m, info);
                const stamp = document.getElementById('mg-svc-stamp');
                if (stamp) {
                    stamp.classList.remove('is-stale');
                }
                stampNow();
            })
            .catch(markStale);
    }

    function startPolling() {
        stopPolling();
        // 15s: fast enough that the LIVE badge is honest, slow enough that the
        // observer effect stays negligible. The old behaviour was a full page
        // reload every 300s.
        state.timer = setInterval(poll, 15000);
    }

    function stopPolling() {
        if (state.timer) {
            clearInterval(state.timer);
            state.timer = null;
        }
    }

    // ---------------------------------------------------------- top bar UI

    function wireLive() {
        const btn = document.getElementById('mg-live');
        if (!btn) {
            return;
        }
        btn.addEventListener('click', () => {
            state.live = !state.live;
            btn.classList.toggle('is-live', state.live);
            btn.setAttribute('aria-pressed', String(state.live));
            set('mg-live-label', state.live ? 'LIVE' : 'PAUSED');
            if (state.live) {
                poll();
                startPolling();
            } else {
                stopPolling();
            }
        });
    }

    function wireRange() {
        const group = document.getElementById('mg-range');
        if (!group) {
            return;
        }
        group.addEventListener('click', (e) => {
            const seg = e.target.closest('.mg-range-seg');
            if (!seg) {
                return;
            }
            state.range = seg.dataset.range;
            [...group.querySelectorAll('.mg-range-seg')].forEach(s => {
                const on = s === seg;
                s.classList.toggle('is-active', on);
                s.setAttribute('aria-pressed', String(on));
            });
            loadSeries();
        });
    }

    // ------------------------------------------------------ history series

    const RANGE_MINUTES = { '5m': 5, '15m': 15, '1h': 60, '24h': 1440 };

    function loadSeries() {
        const mins = RANGE_MINUTES[state.range] || 15;
        const end = new Date();
        const start = new Date(end.getTime() - mins * 60000);
        const body = {
            field_name: ['service_cpu_load', 'heap_inuse', 'goroutines'],
            timeframe: state.range,
            start_time: toUTCISO(start),
            end_time: toUTCISO(end),
        };

        authenticatedFetch('/monigo/api/v1/service-metrics', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body),
        })
            .then(r => {
                if (!r.ok) {
                    throw new Error('service-metrics responded ' + r.status);
                }
                return r.json();
            })
            .then(rows => drawSeries(Array.isArray(rows) ? rows : []))
            .catch(err => {
                console.error('[monigo] runtime load series failed:', err);
                drawSeries(null);
            });
    }

    /*
     * Datapoints are stored against real Unix timestamps, and the handler
     * parses this string as RFC3339. index.js formatted *local* time and then
     * appended "Z", which claims UTC -- so the server searched a window offset
     * by the machine's timezone and returned 500 for every non-UTC host. The
     * chart has to send true UTC.
     */
    function toUTCISO(d) {
        return d.toISOString().replace(/\.\d{3}Z$/, 'Z');
    }

    /*
     * The API re-keys field names on the way out (api.NameMap), so heap_inuse
     * comes back as HeapInuse. Reading the request-side name yields undefined,
     * which becomes NaN, which draws an empty chart that looks like "no data
     * yet" rather than a bug. Check both spellings.
     */
    function pick(value, ...names) {
        for (const n of names) {
            if (value && value[n] != null) {
                return num(value[n], n);
            }
        }
        return NaN;
    }

    // rows === null means the fetch failed. That is a different message from
    // "the window is genuinely short", and conflating the two hid a real bug.
    function drawSeries(rows) {
        const empty = document.getElementById('mg-runtime-load-empty');
        const canvas = document.getElementById('mg-runtime-load-chart');
        const short = document.getElementById('mg-runtime-load-short');
        const failed = document.getElementById('mg-runtime-load-failed');
        set('mg-sync-freq', state.syncFrequency);

        if (rows === null) {
            if (empty) {
                empty.hidden = false;
            }
            if (short) {
                short.hidden = true;
            }
            if (failed) {
                failed.hidden = false;
            }
            if (canvas) {
                canvas.style.visibility = 'hidden';
            }
            return;
        }
        if (failed) {
            failed.hidden = true;
        }
        if (short) {
            short.hidden = false;
        }

        // Two points cannot show a trend. Samples are written every
        // DataPointsSyncFrequency (5m by default), so a 15m window holds three
        // -- saying so is better than drawing a straight line and letting the
        // reader infer stability that was never measured.
        if (rows.length < 2) {
            if (empty) {
                empty.hidden = false;
            }
            if (canvas) {
                canvas.style.visibility = 'hidden';
            }
            return;
        }
        if (empty) {
            empty.hidden = true;
        }
        if (canvas) {
            canvas.style.visibility = '';
        }

        const times = rows.map(r => new Date(r.time).toTimeString().slice(0, 5));
        const cpu = rows.map(r => pick(r.value, 'service_cpu_load', 'ServiceCPULoad'));
        const heap = rows.map(r => {
            const kb = pick(r.value, 'heap_inuse', 'HeapInuse');
            return isFinite(kb) ? +(kb / 1024).toFixed(2) : null;
        });
        const gor = rows.map(r => pick(r.value, 'goroutines', 'Goroutines'));

        renderChart(canvas, times, cpu, heap, gor, rows.length);
    }

    function cssVar(name) {
        return getComputedStyle(document.documentElement).getPropertyValue(name).trim();
    }

    let chart = null;

    function renderChart(el, times, cpu, heap, gor, sampleCount) {
        if (!el || typeof echarts === 'undefined') {
            return;
        }
        if (!chart) {
            chart = echarts.init(el);
            window.addEventListener('resize', () => chart && chart.resize());
        }

        const ink = cssVar('--ink-subtle');
        const grid = cssVar('--grid-line');
        // Fewer than ~6 samples reads as a polyline, not a trend, so the points
        // are shown explicitly rather than implied by a smooth curve.
        const sparse = sampleCount < 6;

        const line = (name, data, colorVar) => ({
            name,
            type: 'line',
            data,
            smooth: false,
            showSymbol: sparse,
            symbolSize: 5,
            lineStyle: { width: 1.6, color: cssVar(colorVar) },
            itemStyle: { color: cssVar(colorVar) },
            areaStyle: sparse ? undefined : { opacity: 0.12, color: cssVar(colorVar) },
            connectNulls: true,
        });

        chart.setOption({
            backgroundColor: 'transparent',
            animation: false,
            // containLabel reserves room for the axis labels but not for the
            // last category label, which boundaryGap:false centres on the final
            // point at the very edge -- so it loses its trailing digit without
            // this inset.
            grid: { left: 8, right: 24, top: 8, bottom: 4, containLabel: true },
            tooltip: { trigger: 'axis' },
            xAxis: {
                type: 'category',
                data: times,
                boundaryGap: false,
                axisLine: { lineStyle: { color: grid } },
                axisTick: { show: false },
                axisLabel: { color: ink, fontSize: 10 },
            },
            yAxis: {
                type: 'value',
                splitLine: { lineStyle: { color: grid } },
                axisLabel: { color: ink, fontSize: 10 },
            },
            series: [
                line('service cpu', cpu, '--series-1'),
                line('heap inuse', heap, '--series-2'),
                line('goroutines', gor, '--series-3'),
            ],
        }, { notMerge: true });
    }

    // --------------------------------------------------- traced functions

    function loadFunctions() {
        authenticatedFetch('/monigo/api/v1/function')
            .then(r => r.ok ? r.json() : {})
            .then(renderFunctions)
            .catch(() => renderFunctions({}));
    }

    function renderFunctions(fns) {
        const rows = document.getElementById('mg-hot-rows');
        const empty = document.getElementById('mg-hot-empty');
        const tpl = document.getElementById('mg-hot-row-tpl');
        if (!rows || !tpl) {
            return;
        }
        const names = Object.keys(fns || {});
        set('mg-hot-count', String(names.length));
        rows.innerHTML = '';

        if (!names.length) {
            if (empty) {
                empty.hidden = false;
            }
            return;
        }
        if (empty) {
            empty.hidden = true;
        }

        // Sorted by the most recent execution time. There is no latency
        // distribution to rank by -- FunctionMetrics keeps only the latest
        // sample -- so the column is labelled LAST EXEC, not P95.
        names
            .map(name => ({ name, m: fns[name] || {} }))
            .sort((a, b) => execMs(b.m) - execMs(a.m))
            .slice(0, 8)
            .forEach(({ name, m }) => {
                const node = tpl.content.cloneNode(true);
                const short = name.split('/').pop();
                node.querySelector('.mg-fn-name').textContent = short;
                const pkg = node.querySelector('.mg-fn-pkg');
                if (pkg) {
                    pkg.textContent = short === name ? '' : name.slice(0, name.length - short.length);
                }
                node.querySelector('.mg-hot__exec').textContent = formatDuration(execMs(m));
                node.querySelector('.mg-hot__mem').textContent =
                    m.memory_usage ? String(m.memory_usage) : '—';
                node.querySelector('.mg-hot__last').textContent = m.function_last_ran || '—';
                rows.appendChild(node);
            });
    }

    function execMs(m) {
        const raw = m && (m.execution_time || m.ExecutionTime);
        if (typeof raw === 'number') {
            return raw / 1e6;   // Go duration, nanoseconds
        }
        const parsed = num(raw, 'execution_time');
        return isFinite(parsed) ? parsed / 1e6 : 0;
    }

    function formatDuration(ms) {
        if (!isFinite(ms) || ms <= 0) {
            return '—';
        }
        if (ms >= 1000) {
            return trimZeros((ms / 1000).toFixed(2)) + 's';
        }
        if (ms >= 1) {
            return trimZeros(ms.toFixed(2)) + 'ms';
        }
        return trimZeros((ms * 1000).toFixed(0)) + 'µs';
    }

    // ----------------------------------------------------------------- init

    function init() {
        // The service name appears in the top bar as well as the sidebar.
        authenticatedFetch('/monigo/api/v1/service-info')
            .then(r => r.ok ? r.json() : null)
            .then(info => {
                if (!info) {
                    return;
                }
                set('mg-service-name', info.service_name || '—');
                set('mg-page-subtitle', info.service_name
                    ? 'runtime · ' + info.service_name
                    : 'runtime');
            })
            .catch(() => { /* the status dot carries the failure */ });

        const year = document.getElementById('mg-footer-year');
        if (year) {
            year.textContent = new Date().getFullYear();
        }

        wireLive();
        wireRange();
        poll();
        startPolling();
        loadSeries();
        loadFunctions();

        // Charts read their colours from the tokens, so a theme change needs a
        // redraw rather than a reload.
        document.addEventListener('monigoThemeChanged', () => {
            loadSeries();
        });
    }

    init();
});
