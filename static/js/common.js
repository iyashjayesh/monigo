document.addEventListener('DOMContentLoaded', () => {
    const refreshHtml = `
        <div class="loader-container">
            <div class="bouncing-dots">
                <div class="dot"></div>
                <div class="dot"></div>
                <div class="dot"></div>
            </div>
        </div>`;


    const elements = {
        healthMessage: document.getElementById('health-message'),
    };

    Object.values(elements).forEach(el => el && (el.innerHTML = refreshHtml));
    
    function fetchMetrics() {
        fetch(`/monigo/api/v1/metrics`)
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
                    document.getElementById('health-message').textContent =health.service_health.message;
                }
            })
            .catch(error => {
                console.error('Error fetching metrics:', error);
            });
    }

    // on page refresh
    fetchMetrics();
});