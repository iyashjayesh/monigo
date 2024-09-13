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

    function fetchFunctionTrace() {
        fetch(`/monigo/api/v1/function-trace`)
            .then(response => response.json())
            .then(data => {
                console.log(data);
                const container = document.getElementById('metricsContainer');
                const emptyMessage = document.getElementById('emptyMessage');

                const functionNumber = document.getElementById('totalFNumber');

                if (Object.keys(data).length === 0) {
                    emptyMessage.textContent = "Oops! Looks like there are no metrics to display. Maybe the functions are taking a coffee break?";
                    functionNumber.innerHTML = `<h3><strong>0</strong></h3>`;
                    return;
                }

                functionNumber.innerHTML = `<h3><strong>${Object.keys(data).length}</strong></h3>`;

                for (const [key, value] of Object.entries(data)) {
                    // Create section
                    const section = document.createElement('div');
                    section.className = 'fms-section';

                    // Extract function name and stack trace
                    const functionName = key.split('.')[1];
                    const [stackTrace, metrics] = value.split('System Metrics Before Execution:');

                    // Create and append title
                    const title = document.createElement('h2');
                    title.textContent = "Function Name: " + functionName;
                    section.appendChild(title);

                    // Create card container
                    const cardContainer = document.createElement('div');
                    cardContainer.className = 'fms-card-container';

                    // Create and append stack trace card
                    const stackTraceCard = document.createElement('div');
                    stackTraceCard.className = 'fms-card fms-stack-trace';
                    stackTraceCard.textContent = stackTrace.trim();
                    cardContainer.appendChild(stackTraceCard);

                    // Create and append metrics card
                    const metricsCard = document.createElement('div');
                    metricsCard.className = 'fms-card fms-metrics';

                    const [before, after] = metrics.split('System Metrics After Execution:');
                    const beforeMetrics = document.createElement('div');
                    beforeMetrics.className = 'fms-before';
                    // beforeMetrics.textContent = 'System Metrics Before Execution:\n' + before.trim();
                    beforeMetrics.innerHTML = '<strong>System Metrics Before Execution:</strong>\n' + before.trim();
                    metricsCard.appendChild(beforeMetrics);

                    const afterMetrics = document.createElement('div');
                    afterMetrics.className = 'fms-after';
                    // afterMetrics.textContent = 'System Metrics After Execution:\n' + after.trim();
                    afterMetrics.innerHTML = '<strong>System Metrics After Execution:</strong>\n' + before.trim();
                    metricsCard.appendChild(afterMetrics);

                    cardContainer.appendChild(metricsCard);

                    section.appendChild(cardContainer);

                    // Append section to container
                    container.appendChild(section);
                }
            })
            .catch(error => {
                console.error('Error fetching metrics:', error);
            });


    }

    // on page refresh
    fetchFunctionTrace();
});