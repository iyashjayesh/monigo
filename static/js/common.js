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

    const elements = {
        healthMessage: document.getElementById('health-message'),
    };

    Object.values(elements).forEach(el => el && (el.innerHTML = refreshHtml));

    function fetchMetrics() {
        authenticatedFetch(`/monigo/api/v1/metrics`)
            .then(response => response.json())
            .then(data => {
                const {
                    // core_statistics,
                    // load_statistics,
                    // cpu_statistics,
                    // memory_statistics,
                    health
                } = data;
                const healthIndicator = document.getElementById('health-indicator');
                if (health.service_health.healthy) {
                    healthIndicator.classList.add('healthy');
                    document.getElementById('health-message').textContent = health.service_health.message;
                } else {
                    healthIndicator.classList.add('unhealthy');
                    document.getElementById('health-message').textContent = health.service_health.message;
                }
            })
            .catch(error => {
                console.error('Error fetching metrics:', error);
            });
    }

    // on page refresh
    fetchMetrics();
});