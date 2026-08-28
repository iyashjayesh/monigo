/*
 * Reports page renderer.
 *
 * Replaces reports.js. One topic at a time, one card per field, drawn from
 * the stored series that /reports already returns.
 *
 * The design's RESOLUTION control and Generate PDF button are absent: storage
 * has no downsampling to resolve to, and there is no PDF renderer that works
 * on an airgapped host. JSON and CSV export the data actually on screen.
 */
document.addEventListener('DOMContentLoaded', () => {
    'use strict';

    // ------------------------------------------------------------ transport

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

    function set(id, text) {
        const el = document.getElementById(id);
        if (el) {
            el.textContent = text;
        }
    }

    /*
     * The handler parses this with time.Parse(time.RFC3339). reports.js built
     * a correct offset-bearing string here and this keeps that behaviour --
     * unlike index.js, which appended a literal "Z" to local time and so asked
     * the server for a window shifted by the machine's timezone.
     *
     * toISOString() is used rather than a hand-built offset because it cannot
     * disagree with the actual instant.
     */
    function toUTCISO(d) {
        return d.toISOString().replace(/\.\d{3}Z$/, 'Z');
    }

    // ---------------------------------------------------------------- state

    const RANGE_MINUTES = { '15m': 15, '1h': 60, '6h': 360, '24h': 1440 };

    const state = {
        range: '1h',
        topic: 'LoadStatistics',
        rows: [],
        window: null,
    };

    // Field names come back in the API's own spelling; these are the labels a
    // reader recognises. An unlisted field falls back to its raw name rather
    // than being dropped, so a new server field still shows up.
    const LABELS = {
        overall_load_of_service: 'Overall service load',
        service_cpu_load: 'Service CPU load',
        service_memory_load: 'Service memory load',
        system_cpu_load: 'System CPU load',
        system_memory_load: 'System memory load',
        total_cores: 'Total cores',
        cores_used_by_service: 'Cores used by service',
        cores_used_by_system: 'Cores used by system',
        total_system_memory: 'Total system memory',
        memory_used_by_system: 'Memory used by system',
        memory_used_by_service: 'Memory used by service',
        available_memory: 'Available memory',
        gc_pause_duration: 'GC pause duration',
        stack_memory_usage: 'Stack memory',
        heap_alloc_by_service: 'Heap alloc by service',
        heap_alloc_by_system: 'Heap alloc by system',
        total_alloc_by_service: 'Total alloc by service',
        total_memory_by_os: 'Total memory by OS',
        bytes_sent: 'Bytes sent',
        bytes_received: 'Bytes received',
        service_health_percent: 'Service health',
        system_health_percent: 'System health',
    };

    // Units are not carried on the series, so they are declared here from what
    // each field is known to be. A field with no entry is printed raw rather
    // than guessed at.
    const UNITS = {
        overall_load_of_service: '%', service_cpu_load: '%', service_memory_load: '%',
        system_cpu_load: '%', system_memory_load: '%',
        service_health_percent: '%', system_health_percent: '%',
        total_system_memory: 'bytes', memory_used_by_system: 'bytes',
        memory_used_by_service: 'bytes', available_memory: 'bytes',
        stack_memory_usage: 'bytes', heap_alloc_by_service: 'bytes',
        heap_alloc_by_system: 'bytes', total_alloc_by_service: 'bytes',
        total_memory_by_os: 'bytes', bytes_sent: 'bytes', bytes_received: 'bytes',
    };

    function formatValue(field, v) {
        if (!isFinite(v)) {
            return '—';
        }
        const unit = UNITS[field];
        if (unit === '%') {
            return v.toFixed(2) + '%';
        }
        if (unit === 'bytes') {
            return formatBytes(v);
        }
        return trim(v.toFixed(2));
    }

    function formatBytes(bytes) {
        const units = ['B', 'KB', 'MB', 'GB', 'TB'];
        let v = Math.abs(bytes);
        let i = 0;
        while (v >= 1024 && i < units.length - 1) {
            v /= 1024;
            i++;
        }
        return (bytes < 0 ? '-' : '') + trim(v.toFixed(2)) + ' ' + units[i];
    }

    function trim(s) {
        return String(s).replace(/\.00$/, '');
    }

    // ---------------------------------------------------------------- fetch

    function load() {
        const end = new Date();
        const start = new Date(end.getTime() - RANGE_MINUTES[state.range] * 60000);
        state.window = { start, end };

        set('mg-rep-window',
            `${start.toLocaleString()} → ${end.toLocaleTimeString()}`);

        authenticatedFetch('/monigo/api/v1/reports', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                topic: state.topic,
                start_time: toUTCISO(start),
                end_time: toUTCISO(end),
                time_frame: state.range,
            }),
        })
            .then(r => {
                if (!r.ok) {
                    throw new Error('reports responded ' + r.status);
                }
                return r.json();
            })
            .then(rows => {
                state.rows = Array.isArray(rows) ? rows : [];
                markConnected(true);
                render();
            })
            .catch(err => {
                console.error('[monigo] report failed:', err);
                state.rows = [];
                markConnected(false);
                render(err);
            });
    }

    function markConnected(ok) {
        const dot = document.getElementById('mg-service-dot');
        if (dot) {
            dot.classList.toggle('is-unhealthy', !ok);
        }
        set('mg-topbar-state', ok ? 'Connected' : 'Disconnected');
    }

    // --------------------------------------------------------------- render

    function render(err) {
        const grid = document.getElementById('mg-rep-grid');
        const tpl = document.getElementById('mg-rep-card-tpl');
        const empty = document.getElementById('mg-rep-empty');
        if (!grid || !tpl) {
            return;
        }
        grid.textContent = '';

        set('mg-rep-samples', state.rows.length
            ? state.rows.length + (state.rows.length === 1 ? ' sample' : ' samples')
            : '—');

        if (!state.rows.length) {
            if (empty) {
                empty.hidden = false;
                empty.textContent = err
                    ? 'Could not load this report. Check the console for the response status.'
                    : 'No samples in this window yet. Samples are written every 5m by default.';
            }
            return;
        }
        if (empty) {
            empty.hidden = true;
        }

        // Field order is whatever the first row declares, so the cards follow
        // the server's own grouping rather than an order invented here.
        const fields = Object.keys(state.rows[0].value || {});

        fields.forEach(field => {
            const series = state.rows.map(r => Number(r.value[field]));
            if (!series.some(isFinite)) {
                return;
            }
            const node = tpl.content.firstElementChild.cloneNode(true);
            node.querySelector('.mg-rep-title').textContent = LABELS[field] || field;

            const finite = series.filter(isFinite);
            const peak = Math.max(...finite);
            const mean = finite.reduce((a, b) => a + b, 0) / finite.length;
            node.querySelector('.mg-rep-stats').textContent =
                `peak ${formatValue(field, peak)} · mean ${formatValue(field, mean)}`;

            drawSpark(node.querySelector('.mg-rep-spark'), series);
            grid.appendChild(node);
        });
    }

    /*
     * A single point has no shape to draw, so the path is left empty and the
     * peak/mean line carries the value instead. Drawing a flat line across the
     * card would claim the value held steady over the window, which is exactly
     * what one sample cannot tell you.
     */
    function drawSpark(svg, series) {
        if (!svg) {
            return;
        }
        const area = svg.querySelector('.mg-rep-area');
        const line = svg.querySelector('.mg-rep-line');
        const pts = series.filter(isFinite);
        if (pts.length < 2) {
            area.removeAttribute('d');
            line.removeAttribute('d');
            svg.classList.add('is-flat');
            return;
        }
        svg.classList.remove('is-flat');

        const W = 280;
        const H = 90;
        const min = Math.min(...pts);
        const max = Math.max(...pts);
        const span = max - min || 1;
        const x = i => (i / (pts.length - 1)) * W;
        const y = v => H - ((v - min) / span) * (H - 8) - 4;

        const d = pts.map((v, i) => `${i ? 'L' : 'M'}${x(i).toFixed(1)},${y(v).toFixed(1)}`).join(' ');
        line.setAttribute('d', d);
        area.setAttribute('d', `${d} L${W},${H} L0,${H} Z`);
    }

    // -------------------------------------------------------------- export

    function download(name, text, mime) {
        const blob = new Blob([text], { type: mime });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = name;
        document.body.appendChild(a);
        a.click();
        a.remove();
        URL.revokeObjectURL(url);
    }

    function stamp() {
        return `${state.topic}-${state.range}`;
    }

    function exportJson() {
        download(`monigo-${stamp()}.json`,
            JSON.stringify({
                topic: state.topic,
                range: state.range,
                start_time: state.window && toUTCISO(state.window.start),
                end_time: state.window && toUTCISO(state.window.end),
                samples: state.rows,
            }, null, 2),
            'application/json');
    }

    function exportCsv() {
        if (!state.rows.length) {
            return;
        }
        const fields = Object.keys(state.rows[0].value || {});
        // Raw values, not the formatted ones: a spreadsheet cannot add "1.2 MB".
        const head = ['time', ...fields].join(',');
        const body = state.rows.map(r =>
            [r.time, ...fields.map(f => r.value[f])].join(','));
        download(`monigo-${stamp()}.csv`, [head, ...body].join('\n') + '\n', 'text/csv');
    }

    // ----------------------------------------------------------------- wire

    function wire() {
        const range = document.getElementById('mg-range');
        if (range) {
            range.addEventListener('click', e => {
                const btn = e.target.closest('[data-range]');
                if (!btn) {
                    return;
                }
                state.range = btn.dataset.range;
                [...range.querySelectorAll('[data-range]')].forEach(b => {
                    const on = b === btn;
                    b.classList.toggle('is-active', on);
                    b.setAttribute('aria-pressed', String(on));
                });
                load();
            });
        }

        const topic = document.getElementById('mg-rep-topic');
        if (topic) {
            topic.addEventListener('change', () => {
                state.topic = topic.value;
                load();
            });
        }

        const json = document.getElementById('mg-rep-json');
        if (json) {
            json.addEventListener('click', exportJson);
        }
        const csv = document.getElementById('mg-rep-csv');
        if (csv) {
            csv.addEventListener('click', exportCsv);
        }
    }

    function init() {
        authenticatedFetch('/monigo/api/v1/service-info')
            .then(r => r.ok ? r.json() : null)
            .then(info => {
                if (info && info.service_name) {
                    set('mg-service-name', info.service_name);
                    set('mg-page-subtitle', 'stored series · ' + info.service_name);
                }
            })
            .catch(() => { /* the status dot carries the failure */ });

        const year = document.getElementById('mg-footer-year');
        if (year) {
            year.textContent = new Date().getFullYear();
        }

        wire();
        load();
    }

    init();
});
