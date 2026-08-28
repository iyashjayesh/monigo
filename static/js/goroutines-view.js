/*
 * Goroutines page renderer.
 *
 * Replaces goroutines.js, which drew the template's stack list. Every panel
 * here is backed by /go-routines-stats: the leak report now carries every
 * distinct call stack, not just the offending ones, because a breakdown by
 * state cannot be computed from the offenders alone.
 *
 * Where the design shows something MoniGo does not measure -- the live-count
 * sparkline needs a goroutine time series this endpoint does not return -- the
 * element is left out rather than filled with a plausible shape.
 */
document.addEventListener('DOMContentLoaded', () => {
    'use strict';

    // ------------------------------------------------------------ transport

    // Duplicated from common.js for the same reason overview.js duplicates it:
    // exporting would put the API key on `window`.
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

    // The API reports blocked time in whole minutes; zero means "under a
    // minute, or a state the runtime does not timestamp", which is not the
    // same as "not blocked" and must not render as 0m.
    function blockedLabel(minutes) {
        if (!minutes) {
            return '—';
        }
        if (minutes < 60) {
            return minutes + 'm';
        }
        const h = Math.floor(minutes / 60);
        const m = minutes % 60;
        return m ? `${h}h ${m}m` : `${h}h`;
    }

    function growthLabel(growth) {
        if (!growth) {
            return '0';
        }
        return (growth > 0 ? '+' : '') + growth;
    }

    // The runtime's wait reasons are an open set, so states are coloured by
    // what they imply rather than by an exhaustive list: anything running or
    // runnable is healthy, a blocking wait is worth noticing, and a syscall or
    // lock wait sits between.
    function stateTone(state) {
        const s = String(state || '').toLowerCase();
        if (s.includes('running') || s.includes('runnable')) {
            return 'ok';
        }
        if (s.includes('syscall') || s.includes('io wait')) {
            return 'info';
        }
        if (s.includes('chan') || s.includes('select') || s.includes('semacquire') ||
            s.includes('sleep') || s.includes('wait')) {
            return 'warn';
        }
        return 'info';
    }

    /*
     * The group carries the whole shared stack; the design's TOP FRAME column
     * wants only the function the goroutine is sitting in, without arguments.
     *
     * Cutting at the first "(" is wrong: a method frame parenthesises its
     * receiver, so
     *   monigo.(*Monigo).registerShutdownHandler.func1()
     * becomes "monigo." -- which is what shipped for a moment and told the
     * reader nothing. Only the trailing argument list is removed, matched as a
     * paren group at the end of the line containing no nested parens.
     */
    function topFrame(callStack) {
        const first = String(callStack || '').split('\n').find(l => l.trim());
        if (!first) {
            return '—';
        }
        return first.trim().replace(/\([^()]*\)$/, '');
    }

    // ---------------------------------------------------------------- panels

    function renderSummary(data) {
        const report = data.leak_report || {};
        const groups = report.groups || [];

        set('mg-gor-count', data.number_of_goroutines != null
            ? data.number_of_goroutines
            : (report.total_goroutines != null ? report.total_goroutines : '—'));

        // The design shows "/ 100" against a configured ceiling. MoniGo's
        // goroutine threshold is not serialised on this endpoint, so the limit
        // is omitted rather than invented.
        set('mg-gor-limit', '');

        const total = report.groups_total || groups.length;
        set('mg-gor-groups', total
            ? total + (total === 1 ? ' group' : ' groups')
            : '');

        const stale = report.stale_goroutines || 0;
        const growing = report.growing_groups || 0;
        const detail = document.getElementById('mg-gor-live-detail');
        if (detail) {
            detail.textContent =
                `${stale} stale · ${growing} growing`;
            detail.classList.toggle('mg-tone-warn', stale > 0 || growing > 0);
            detail.classList.toggle('mg-tone-dim', stale === 0 && growing === 0);
        }

        renderStates(groups);
        renderBlocked(groups, report);
        renderVerdict(report);
        renderRows(groups, report);
    }

    function renderStates(groups) {
        const bar = document.getElementById('mg-gor-statebar');
        const list = document.getElementById('mg-gor-statelist');
        if (!bar || !list) {
            return;
        }
        bar.textContent = '';
        list.textContent = '';

        const byState = new Map();
        let total = 0;
        groups.forEach(g => {
            const key = g.state || 'unknown';
            byState.set(key, (byState.get(key) || 0) + g.count);
            total += g.count;
        });
        if (!total) {
            list.innerHTML = '<li class="mg-gor-emptyline">No goroutine groups reported.</li>';
            return;
        }

        const ordered = [...byState.entries()].sort((a, b) => b[1] - a[1]);

        // The bar keeps every state -- it is proportional and costs no height.
        // The legend is capped so the card cannot grow past its neighbours.
        const MAX_LEGEND = 5;
        const shown = ordered.slice(0, MAX_LEGEND);
        const rest = ordered.slice(MAX_LEGEND);

        ordered.forEach(([state, count]) => {
            const tone = stateTone(state);

            const seg = document.createElement('span');
            seg.className = 'mg-gor-seg mg-tone-' + tone;
            seg.style.width = (count / total * 100).toFixed(2) + '%';
            bar.appendChild(seg);
        });

        shown.forEach(([state, count]) => {
            const tone = stateTone(state);
            const li = document.createElement('li');
            li.innerHTML =
                `<i class="mg-gor-swatch mg-tone-${tone}" aria-hidden="true"></i>` +
                `<span class="mg-gor-statename"></span>` +
                `<span class="mg-gor-statecount mg-num"></span>`;
            li.querySelector('.mg-gor-statename').textContent = state;
            li.querySelector('.mg-gor-statecount').textContent = count;
            list.appendChild(li);
        });

        if (rest.length) {
            const li = document.createElement('li');
            li.className = 'mg-gor-morestates';
            const n = rest.reduce((sum, [, c]) => sum + c, 0);
            li.textContent = `+${rest.length} more ${rest.length === 1 ? 'state' : 'states'} · ${n}`;
            li.title = rest.map(([st, c]) => `${st}: ${c}`).join(', ');
            list.appendChild(li);
        }

        bar.setAttribute('aria-label',
            ordered.map(([s, c]) => `${s}: ${c}`).join(', '));
    }

    function renderBlocked(groups, report) {
        const list = document.getElementById('mg-gor-blocked');
        if (!list) {
            return;
        }
        list.textContent = '';

        const blocked = groups
            .filter(g => g.blocked_minutes > 0)
            .sort((a, b) => b.blocked_minutes - a.blocked_minutes)
            .slice(0, 3);

        if (!blocked.length) {
            list.innerHTML =
                '<li class="mg-gor-emptyline">Nothing blocked long enough for the ' +
                'runtime to timestamp it.</li>';
        } else {
            blocked.forEach(g => {
                const li = document.createElement('li');
                li.innerHTML = '<span class="mg-gor-blockedname"></span>' +
                    '<span class="mg-gor-blockedtime"></span>';
                li.querySelector('.mg-gor-blockedname').textContent = topFrame(g.call_stack);
                const t = li.querySelector('.mg-gor-blockedtime');
                t.textContent = blockedLabel(g.blocked_minutes);
                t.classList.add(g.stale ? 'mg-tone-warn' : 'mg-tone-dim');
                list.appendChild(li);
            });
        }

        const mins = report.stale_threshold_minutes;
        set('mg-gor-threshold', mins
            ? 'stale threshold ' + blockedLabel(mins)
            : '');
    }

    function renderVerdict(report) {
        const card = document.getElementById('mg-gor-verdict');
        const icon = document.getElementById('mg-gor-verdict-icon');
        if (!card) {
            return;
        }
        const suspected = report.leak_suspected === true;

        card.classList.toggle('is-suspect', suspected);
        if (icon) {
            icon.setAttribute('href', suspected ? '#i-alert-triangle' : '#i-check-circle');
        }
        set('mg-gor-verdict-title', suspected ? 'Leak suspected' : 'No leak detected');

        // The server writes this sentence, including the snapshot counts the
        // design shows ("1 growing across 6 of 6 retained snapshots") and the
        // warming-up notice while the window fills. It is rendered verbatim:
        // appending the same facts here said everything twice.
        set('mg-gor-verdict-msg', report.message || '');
    }

    function renderRows(groups, report) {
        const host = document.getElementById('mg-gor-rows');
        const tpl = document.getElementById('mg-gor-row-tpl');
        if (!host || !tpl) {
            return;
        }
        host.textContent = '';

        if (!groups.length) {
            const p = document.createElement('p');
            p.className = 'mg-gor-emptyline';
            p.textContent = 'No goroutine groups reported yet.';
            host.appendChild(p);
        }

        groups.forEach(g => {
            const node = tpl.content.firstElementChild.cloneNode(true);
            const row = node.querySelector('.mg-gor-row');
            const stack = node.querySelector('.mg-gor-stack');

            node.querySelector('.mg-gor-sig').textContent = g.signature || '—';
            const frame = node.querySelector('.mg-gor-frame');
            frame.textContent = topFrame(g.call_stack);
            frame.title = frame.textContent;

            const state = node.querySelector('.mg-gor-state');
            state.textContent = g.state || '—';
            state.classList.add('mg-tone-' + stateTone(g.state));

            node.querySelector('.mg-gor-n').textContent = g.count;

            const bl = node.querySelector('.mg-gor-blockedcell');
            bl.textContent = blockedLabel(g.blocked_minutes);
            if (g.stale) {
                bl.classList.add('mg-tone-warn');
            }

            const gr = node.querySelector('.mg-gor-growth');
            gr.textContent = growthLabel(g.growth);
            if (g.growing) {
                gr.classList.add('mg-tone-warn');
            }

            if (g.stale || g.growing) {
                row.classList.add('is-suspect');
            }

            stack.textContent = g.call_stack || '(no stack recorded)';
            row.addEventListener('click', () => {
                const open = !stack.hidden;
                stack.hidden = open;
                row.setAttribute('aria-expanded', String(!open));
            });

            host.appendChild(node);
        });

        // Saying how many are hidden matters more than hiding them: a list that
        // silently stops at the cap reads as "this is all of them".
        const shown = groups.length;
        const total = report.groups_total || shown;
        const note = document.getElementById('mg-gor-truncated');
        if (note) {
            note.hidden = total <= shown;
            note.textContent = total > shown
                ? `Showing the ${shown} most significant of ${total} groups.`
                : '';
        }
    }

    // -------------------------------------------------------------- stacks

    function wireDownload(getData) {
        const btn = document.getElementById('mg-gor-snapshot');
        if (!btn) {
            return;
        }
        btn.addEventListener('click', () => {
            const data = getData();
            if (!data) {
                return;
            }
            const stacks = (data.stack_view || []).join('\n\n');
            const blob = new Blob([stacks], { type: 'text/plain' });
            const url = URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'goroutine-stacks.txt';
            document.body.appendChild(a);
            a.click();
            a.remove();
            URL.revokeObjectURL(url);
        });
    }

    // ---------------------------------------------------------------- poll

    const state = { live: true, timer: null, last: null };

    function poll() {
        authenticatedFetch('/monigo/api/v1/go-routines-stats')
            .then(r => {
                if (!r.ok) {
                    throw new Error('go-routines-stats responded ' + r.status);
                }
                return r.json();
            })
            .then(data => {
                state.last = data;
                markConnected(true);
                renderSummary(data);
            })
            .catch(err => {
                console.error('[monigo] goroutine stats failed:', err);
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
                state.timer = setInterval(poll, 15000);
            } else {
                clearInterval(state.timer);
            }
        });
    }

    function init() {
        authenticatedFetch('/monigo/api/v1/service-info')
            .then(r => r.ok ? r.json() : null)
            .then(info => {
                if (info && info.service_name) {
                    set('mg-service-name', info.service_name);
                    set('mg-page-subtitle', 'leak detection · ' + info.service_name);
                }
            })
            .catch(() => { /* the status dot carries the failure */ });

        const year = document.getElementById('mg-footer-year');
        if (year) {
            year.textContent = new Date().getFullYear();
        }

        wireDownload(() => state.last);
        wireLive();
        poll();
        state.timer = setInterval(poll, 15000);
    }

    init();
});
