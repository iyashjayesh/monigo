/*
 * Functions page renderer.
 *
 * Replaces functionTrace.js's list half. /function returns a map of
 * qualified name -> current totals, so this page shows totals: four data
 * columns, not the design's seven.
 *
 * CALLS, P50, P95 and TREND are absent because FunctionMetrics records
 * cumulative values and no per-function time series. Percentiles need a
 * histogram or reservoir per function; a trend needs samples over time.
 * Neither can be derived from what this endpoint returns, so the columns are
 * removed rather than rendered empty -- an empty column reads as "loading".
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

    // --------------------------------------------------------------- format

    function set(id, text) {
        const el = document.getElementById(id);
        if (el) {
            el.textContent = text;
        }
    }

    // execution_time is nanoseconds (time.Duration marshalled as an integer).
    function formatDuration(ns) {
        if (!isFinite(ns) || ns < 0) {
            return '—';
        }
        if (ns < 1000) {
            return ns + 'ns';
        }
        if (ns < 1e6) {
            return trim((ns / 1000).toFixed(2)) + 'µs';
        }
        if (ns < 1e9) {
            return trim((ns / 1e6).toFixed(2)) + 'ms';
        }
        return trim((ns / 1e9).toFixed(2)) + 's';
    }

    function formatBytes(bytes) {
        if (!isFinite(bytes)) {
            return '—';
        }
        if (bytes === 0) {
            return '0';
        }
        const units = ['B', 'KB', 'MB', 'GB'];
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

    function formatWhen(iso) {
        const t = Date.parse(iso);
        if (!isFinite(t)) {
            return '—';
        }
        const secs = Math.round((Date.now() - t) / 1000);
        if (secs < 0) {
            return 'just now';
        }
        if (secs < 60) {
            return secs + 's ago';
        }
        if (secs < 3600) {
            return Math.floor(secs / 60) + 'm ago';
        }
        if (secs < 86400) {
            return Math.floor(secs / 3600) + 'h ago';
        }
        return Math.floor(secs / 86400) + 'd ago';
    }

    // "main.highCPUUsage" -> package "main", name "highCPUUsage". The design
    // shows them as separate tones in one cell.
    function splitName(qualified) {
        const i = String(qualified).lastIndexOf('.');
        return i < 0
            ? { pkg: '', name: qualified }
            : { pkg: qualified.slice(0, i), name: qualified.slice(i + 1) };
    }

    // ---------------------------------------------------------------- state

    const state = {
        rows: [],
        filter: '',
        sort: 'exec',
        timer: null,
        openFor: null,
    };

    const SORTS = {
        exec: (a, b) => b.exec - a.exec,
        mem: (a, b) => b.mem - a.mem,
        gor: (a, b) => b.gor - a.gor,
        last: (a, b) => b.lastMs - a.lastMs,
        name: (a, b) => a.qualified.localeCompare(b.qualified),
    };

    function toRows(payload) {
        return Object.entries(payload || {}).map(([qualified, m]) => {
            const parts = splitName(qualified);
            return {
                qualified,
                pkg: parts.pkg,
                name: parts.name,
                exec: Number(m.execution_time) || 0,
                mem: Number(m.memory_usage) || 0,
                gor: Number(m.goroutine_count) || 0,
                last: m.function_last_ran_at,
                lastMs: Date.parse(m.function_last_ran_at) || 0,
                // Empty unless a profile was actually written for this call.
                hasProfile: Boolean(m.cpu_profile_file_path || m.mem_profile_file_path),
            };
        });
    }

    // ---------------------------------------------------------------- table

    function render() {
        const host = document.getElementById('mg-fn-rows');
        const tpl = document.getElementById('mg-fn-row-tpl');
        const empty = document.getElementById('mg-fn-empty');
        if (!host || !tpl) {
            return;
        }

        const q = state.filter.trim().toLowerCase();
        const visible = state.rows
            .filter(r => !q || r.qualified.toLowerCase().includes(q))
            .sort(SORTS[state.sort] || SORTS.exec);

        host.textContent = '';
        visible.forEach(r => {
            const node = tpl.content.firstElementChild.cloneNode(true);
            const row = node.querySelector('.mg-fn-row');
            const detail = node.querySelector('.mg-fn-detail');

            node.querySelector('.mg-fn-name').textContent = r.name;
            node.querySelector('.mg-fn-pkg').textContent = r.pkg;
            node.querySelector('.mg-fn-exec').textContent = formatDuration(r.exec);
            node.querySelector('.mg-fn-mem').textContent = formatBytes(r.mem);
            node.querySelector('.mg-fn-gor').textContent = r.gor;
            const last = node.querySelector('.mg-fn-last');
            last.textContent = formatWhen(r.last);
            last.title = r.last || '';

            row.addEventListener('click', () => toggle(r, row, detail));
            if (state.openFor === r.qualified) {
                openDetail(r, row, detail);
            }
            host.appendChild(node);
        });

        if (empty) {
            empty.hidden = visible.length > 0;
            empty.textContent = state.rows.length === 0
                ? 'No traced functions yet. Wrap a call with monigo.TraceFunction to see it here.'
                : `No function matches “${state.filter.trim()}”.`;
        }

        const total = state.rows.length;
        set('mg-fn-count', total
            ? (visible.length === total
                ? `${total} ${total === 1 ? 'function' : 'functions'}`
                : `${visible.length} of ${total}`)
            : '');
    }

    function toggle(r, row, detail) {
        if (!detail.hidden) {
            detail.hidden = true;
            row.setAttribute('aria-expanded', 'false');
            state.openFor = null;
            return;
        }
        state.openFor = r.qualified;
        openDetail(r, row, detail);
    }

    function openDetail(r, row, detail) {
        detail.hidden = false;
        row.setAttribute('aria-expanded', 'true');
        detail.innerHTML = '';

        const grid = document.createElement('div');
        grid.className = 'mg-fn-detailgrid';
        [
            ['EXEC TIME', formatDuration(r.exec), 'cumulative'],
            ['MEMORY Δ', formatBytes(r.mem), 'cumulative'],
            ['GOROUTINE Δ', String(r.gor), r.gor ? 'net change' : 'no change'],
            ['LAST RAN', formatWhen(r.last), new Date(r.lastMs).toLocaleString()],
        ].forEach(([label, value, note]) => {
            const cell = document.createElement('div');
            cell.className = 'mg-fn-detailcell';
            cell.innerHTML = '<span class="mg-eyebrow"></span>' +
                '<span class="mg-fn-detailvalue"></span>' +
                '<span class="mg-fn-detailnote"></span>';
            cell.querySelector('.mg-eyebrow').textContent = label;
            cell.querySelector('.mg-fn-detailvalue').textContent = value;
            cell.querySelector('.mg-fn-detailnote').textContent = note;
            grid.appendChild(cell);
        });
        detail.appendChild(grid);

        /*
         * The profile view shells out to `go tool pprof` (issue #43), so this
         * links to the existing text report rather than embedding a flame graph
         * the dashboard cannot draw offline.
         *
         * Only when a profile was actually written. Without the guard the link
         * is always present and always opens
         *   {"cpu_profile":"Error: Profile file path is empty"}
         * which is worse than saying there is nothing to show.
         */
        if (r.hasProfile) {
            const link = document.createElement('a');
            link.className = 'mg-fn-profile';
            link.href = `/monigo/api/v1/function-details?name=${encodeURIComponent(r.qualified)}&reportType=text`;
            link.target = '_blank';
            link.rel = 'noopener';
            link.textContent = 'Open pprof text report';
            detail.appendChild(link);
        } else {
            const note = document.createElement('span');
            note.className = 'mg-fn-noprofile';
            note.textContent = 'No pprof profile was captured for this call.';
            detail.appendChild(note);
        }
    }

    // ----------------------------------------------------------------- csv

    function exportCsv() {
        const q = state.filter.trim().toLowerCase();
        const rows = state.rows
            .filter(r => !q || r.qualified.toLowerCase().includes(q))
            .sort(SORTS[state.sort] || SORTS.exec);

        // Raw values, not the formatted ones: a CSV is for a spreadsheet, and
        // "22.56ms" is not a number to anything that opens it.
        const head = 'function,execution_time_ns,memory_usage_bytes,goroutine_count,last_ran_at';
        const body = rows.map(r => [
            csvCell(r.qualified), r.exec, r.mem, r.gor, csvCell(r.last || ''),
        ].join(','));

        const blob = new Blob([[head, ...body].join('\n') + '\n'], { type: 'text/csv' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = 'monigo-functions.csv';
        document.body.appendChild(a);
        a.click();
        a.remove();
        URL.revokeObjectURL(url);
    }

    function csvCell(v) {
        const s = String(v);
        return /[",\n]/.test(s) ? '"' + s.replace(/"/g, '""') + '"' : s;
    }

    // ---------------------------------------------------------------- poll

    function poll() {
        authenticatedFetch('/monigo/api/v1/function')
            .then(r => {
                if (!r.ok) {
                    throw new Error('function responded ' + r.status);
                }
                return r.json();
            })
            .then(payload => {
                state.rows = toRows(payload);
                markConnected(true);
                render();
            })
            .catch(err => {
                console.error('[monigo] function metrics failed:', err);
                markConnected(false);
            });
    }

    function markConnected(ok) {
        const dot = document.getElementById('mg-service-dot');
        if (dot) {
            dot.classList.toggle('is-unhealthy', !ok);
        }
        set('mg-topbar-state', ok ? 'Connected' : 'Disconnected');
    }

    function wire() {
        const filter = document.getElementById('mg-fn-filter');
        if (filter) {
            filter.addEventListener('input', () => {
                state.filter = filter.value;
                render();
            });
        }
        const sort = document.getElementById('mg-fn-sort');
        if (sort) {
            sort.addEventListener('change', () => {
                state.sort = sort.value;
                render();
            });
        }
        const csv = document.getElementById('mg-fn-export');
        if (csv) {
            csv.addEventListener('click', exportCsv);
        }

        const live = document.getElementById('mg-live');
        if (live) {
            live.addEventListener('click', () => {
                const on = !live.classList.contains('is-live');
                live.classList.toggle('is-live', on);
                live.setAttribute('aria-pressed', String(on));
                set('mg-live-label', on ? 'LIVE' : 'PAUSED');
                clearInterval(state.timer);
                if (on) {
                    poll();
                    state.timer = setInterval(poll, 15000);
                }
            });
        }
    }

    function init() {
        authenticatedFetch('/monigo/api/v1/service-info')
            .then(r => r.ok ? r.json() : null)
            .then(info => {
                if (info && info.service_name) {
                    set('mg-service-name', info.service_name);
                    set('mg-page-subtitle', 'traced calls · ' + info.service_name);
                }
            })
            .catch(() => { /* the status dot carries the failure */ });

        const year = document.getElementById('mg-footer-year');
        if (year) {
            year.textContent = new Date().getFullYear();
        }

        wire();
        poll();
        state.timer = setInterval(poll, 15000);
    }

    init();
});
