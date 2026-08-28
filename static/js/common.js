document.addEventListener('DOMContentLoaded', () => {
    const refreshHtml = `
        <div class="loader-container">
            <div class="bouncing-dots">
                <div class="dot"></div>
                <div class="dot"></div>
                <div class="dot"></div>
            </div>
        </div>`;

    // Function to get API key from URL parameters
    function getApiKey() {
        const urlParams = new URLSearchParams(window.location.search);
        return urlParams.get('api_key');
    }

    // Function to add API key to fetch URL (only for API key auth)
    function addApiKeyToUrl(url) {
        const apiKey = getApiKey();
        if (apiKey) {
            const separator = url.includes('?') ? '&' : '?';
            return `${url}${separator}api_key=${encodeURIComponent(apiKey)}`;
        }
        return url;
    }

    // Function to make authenticated fetch request
    function authenticatedFetch(url, options = {}) {
        const apiKey = getApiKey();
        if (apiKey) {
            // API key authentication - add to URL
            const separator = url.includes('?') ? '&' : '?';
            url = `${url}${separator}api_key=${encodeURIComponent(apiKey)}`;
        } else {
            // Check for custom authentication methods
            const urlParams = new URLSearchParams(window.location.search);
            const secret = urlParams.get('secret');

            if (secret === 'monigo-admin-secret') {
                // Custom query parameter authentication
                const separator = url.includes('?') ? '&' : '?';
                url = `${url}${separator}secret=${encodeURIComponent(secret)}`;
            } else {
                // Check for custom header authentication
                // For custom auth, we need to add headers
                if (!options.headers) {
                    options.headers = {};
                }

                // Add custom header for admin access
                options.headers['X-User-Role'] = 'admin';

                // Set custom user agent for automated access
                options.headers['User-Agent'] = 'MoniGo-Admin/1.0';
            }
        }
        // For basic auth, the browser handles credentials automatically
        return fetch(url, options);
    }

    // 1. Theme Configuration & Toggle Initialization (Synchronous, prioritized)
    const savedTheme = localStorage.getItem('monigo-theme') || 'light';
    
    // Inject theme toggle button into the correct, visible top right navbar-list
    const navbarList = document.querySelector('#navbarSupportedContent .navbar-list') || document.querySelector('.navbar-nav.navbar-list');
    if (navbarList) {
        const toggleLi = document.createElement('li');
        toggleLi.className = 'nav-item nav-icon ml-3';
        
        const toggleLink = document.createElement('a');
        toggleLink.id = 'theme-toggle-btn';
        toggleLink.className = 'cursor-pointer';
        toggleLink.innerHTML =
            '<svg class="icon" aria-hidden="true" style="width:16px;height:16px;"><use href="#i-moon"></use></svg>';
        
        toggleLi.appendChild(toggleLink);
        navbarList.appendChild(toggleLi);
        
        // Add event listener to toggle theme
        toggleLink.addEventListener('click', () => {
            const currentTheme = document.body.classList.contains('dark-theme') ? 'dark' : 'light';
            const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
            setTheme(newTheme);
        });
    }

    function setTheme(theme) {
        if (theme === 'dark') {
            document.body.classList.add('dark-theme');
            localStorage.setItem('monigo-theme', 'dark');
            const icon = document.querySelector('#theme-toggle-btn use');
            if (icon) {
                icon.setAttribute('href', '#i-sun');
            }
        } else {
            document.body.classList.remove('dark-theme');
            localStorage.setItem('monigo-theme', 'light');
            const icon = document.querySelector('#theme-toggle-btn use');
            if (icon) {
                icon.setAttribute('href', '#i-moon');
            }
        }
        // Emit event to notify other scripts (like charts) to update colors
        document.dispatchEvent(new CustomEvent('monigoThemeChanged', { detail: { theme } }));
    }

    // Apply saved theme preference immediately
    setTheme(savedTheme);

    // 2. Pre-populate all chart containers with a stylized loading spinner
    document.querySelectorAll('.chart-container').forEach(el => {
        el.innerHTML = `
            <div class="d-flex flex-column align-items-center justify-content-center h-100 text-muted" style="min-height: 200px;">
                <svg class="icon icon-2x icon-spin mb-3" aria-hidden="true" style="color: #ff5c35;"><use href="#i-refresh"></use></svg>
                <span class="small font-weight-bold">Awaiting MoniGo server connection...</span>
            </div>
        `;
    });

    const elements = {
        healthMessage: document.getElementById('health-message'),
    };

    Object.values(elements).forEach(el => el && (el.innerHTML = refreshHtml));

    function fetchMetrics() {
        authenticatedFetch(`/monigo/api/v1/metrics`)
            .then(response => response.json())
            .then(data => {
                const { health } = data;
                const healthIndicator = document.getElementById('health-indicator');
                if (healthIndicator) {
                    if (health.service_health.healthy) {
                        healthIndicator.classList.remove('unhealthy');
                        healthIndicator.classList.add('healthy');
                    } else {
                        healthIndicator.classList.remove('healthy');
                        healthIndicator.classList.add('unhealthy');
                    }
                }
                const healthMsgEl = document.getElementById('health-message');
                if (healthMsgEl) {
                    healthMsgEl.textContent = health.service_health.message;
                }
            })
            .catch(error => {
                console.error('Error fetching metrics:', error);
            });
    }

    // --- Shell: service identity, nav badges, storage footer ---------------
    //
    // These live in the sidebar chrome because they are true of the whole
    // instrument rather than of any one page, so every page shows them and no
    // page has to fetch them twice.

    function setStatus(state) {
        const dot = document.getElementById('mg-status-dot');
        if (!dot) {
            return;
        }
        dot.classList.remove('is-stale', 'is-down');
        if (state !== 'ok') {
            dot.classList.add(state === 'down' ? 'is-down' : 'is-stale');
        }
        dot.title = state === 'ok' ? 'Connected'
            : state === 'down' ? 'Lost connection to the service'
            : 'Data is stale';
    }

    function fillShell() {
        authenticatedFetch('/monigo/api/v1/service-info')
            .then(r => {
                if (!r.ok) {
                    throw new Error('HTTP ' + r.status);
                }
                return r.json();
            })
            .then(info => {
                const name = info.service_name || 'service';
                const nameEl = document.getElementById('mg-service-name');
                const metaEl = document.getElementById('mg-service-meta');
                const avatar = document.getElementById('mg-avatar');
                if (nameEl) {
                    nameEl.textContent = name;
                    nameEl.title = name;
                }
                if (avatar) {
                    avatar.textContent = name.slice(0, 2);
                }
                if (metaEl) {
                    // The design's `pid 94917 · go1.21`: the two facts you need
                    // to be sure you are looking at the right process.
                    const go = (info.go_version || '').replace(/^go/, '');
                    metaEl.textContent = `pid ${info.process_id || '-'}${go ? ' \u00b7 go' + go : ''}`;
                }
                // Retention and footprint are properties of the instrument, so
                // they sit in the chrome. Both are omitted by the API when
                // unset, so absent means "not configured", not "zero".
                const mode = document.getElementById('mg-storage-mode');
                const store = document.getElementById('mg-storage-value');
                if (mode) {
                    const parts = [info.storage_type, info.retention_period].filter(Boolean);
                    mode.textContent = parts.join(' \u00b7 ');
                }
                if (store) {
                    store.textContent = info.storage_on_disk
                        ? `${info.storage_on_disk} retained`
                        : (info.storage_type === 'memory' ? 'in memory' : '');
                }
                setStatus('ok');
            })
            .catch(() => setStatus('down'));
    }

    function fillNavBadges() {
        authenticatedFetch('/monigo/api/v1/metrics')
            .then(r => r.json())
            .then(data => {
                const g = document.getElementById('mg-count-goroutines');
                if (g && data.core_statistics) {
                    g.textContent = data.core_statistics.goroutines;
                }

                // Traced-function count. The endpoint returns a map keyed by
                // function name, so its size is the count. Left blank when
                // nothing is traced yet -- an empty badge is hidden by CSS
                // rather than showing a zero that reads like a measurement.
                authenticatedFetch('/monigo/api/v1/function')
                    .then(r => r.json())
                    .then(fns => {
                        const el = document.getElementById('mg-count-functions');
                        const n = fns && typeof fns === 'object' ? Object.keys(fns).length : 0;
                        if (el) {
                            el.textContent = n > 0 ? String(n) : '';
                        }
                    })
                    .catch(() => { /* badge simply stays empty */ });
                // storage footer is filled from /service-info, not here
            })
            .catch(() => { /* the status dot already carries the failure */ });
    }

    fillShell();
    fillNavBadges();

    // Fetch metrics on load
    fetchMetrics();
});