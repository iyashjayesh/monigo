document.addEventListener('DOMContentLoaded', function () {
    const serviceNameElement = document.getElementById('servicename');
    const serviceStartTimeElement = document.getElementById('servicestarttime');
    const gcountElement = document.getElementById('gcount');
    const rcountElement = document.getElementById('rcount');
    const tdurElement = document.getElementById('tdur');
    const allocElement = document.getElementById('alloc');
    const totalallocElement = document.getElementById('totalalloc');
    const sysElement = document.getElementById('sys');
    const heapallocElement = document.getElementById('heapalloc');
    const heapsysElement = document.getElementById('heapsys');
    const unitInputs = document.querySelectorAll('input[name="unit"]');
    const countdownElement = document.getElementById('countdown');
    const refreshBtn = document.getElementById('refresh-btn');
    const functionsContainer = document.getElementById('functions-container');

    let refreshInterval = 60000; // 60 seconds
    let refreshTimeout;

    function fetchMetrics(unit) {
        fetch(`/metrics?unit=${unit}`)
            .then(response => response.text())
            .then(data => {
                const lines = data.split('\n');
                serviceNameElement.textContent = 'Service Name: ' + lines[0].split(': ')[1];
                serviceStartTimeElement.textContent = 'Service Start Time: ' + lines[1].split(': ')[1];
                gcountElement.textContent = 'Goroutines: ' + lines[2].split(': ')[1];
                rcountElement.textContent = 'Requests: ' + lines[3].split(': ')[1];
                tdurElement.textContent = 'Total Duration: ' + lines[4].split(': ')[1];
                allocElement.textContent = 'Alloc: ' + lines[6].split(': ')[1];
                totalallocElement.textContent = 'TotalAlloc: ' + lines[7].split(': ')[1];
                sysElement.textContent = 'Sys: ' + lines[8].split(': ')[1];
                heapallocElement.textContent = 'HeapAlloc: ' + lines[9].split(': ')[1];
                heapsysElement.textContent = 'HeapSys: ' + lines[10].split(': ')[1];
            });
    }

    function fetchFunctionMetrics(unit) {
        fetch(`/function-metrics?unit=${unit}`)
            .then(response => response.text())
            .then(data => {
                // Clear the previous function metrics
                functionsContainer.innerHTML = '';

                // Split the response into blocks of function metrics
                const functionBlocks = data.trim().split('\n\n');

                functionBlocks.forEach(block => {
                    const lines = block.split('\n');
                    if (lines.length > 1) {
                        const functionName = lines[0].split(': ')[1];
                        const functionLastRanAT = lines[1].split(': ')[1];
                        const cpuProfile = lines[2].split(': ')[1];
                        const execTime = lines[3].split(': ')[1];
                        const memoryUsage = lines[4].split(': ')[1];
                        const goroutineCount = lines[5].split(': ')[1];

                        const functionMetricHtml = `
                            <div class="function-metric">
                                <h3>Function: ${functionName}</h3>
                                <p>Function Last Ran At: ${functionLastRanAT}</p>
                                <p>CPU Profile: ${cpuProfile}</p>
                                <p>Execution Time: ${execTime}</p>
                                <p>Memory Usage: ${memoryUsage} bytes</p>
                                <p>Goroutines: ${goroutineCount}</p>

                                <button class="download-btn-cpu" data-function-name="${functionName}">View CPU Metrics</button>
                                <button class="download-btn-mem" data-function-name="${functionName}">View MEM Metrics</button>
                            </div>
                        `;

                        // Append the function metric to the container
                        functionsContainer.innerHTML += functionMetricHtml;
                    }
                });

                // Add event listeners to the download buttons
                document.querySelectorAll('.download-btn-cpu').forEach(button => {
                    button.addEventListener('click', function () {
                        const functionName = button.getAttribute('data-function-name');
                        const functionBlock = [...document.querySelectorAll('.function-metric')]
                            .find(metric => metric.querySelector('h3').textContent.includes(functionName));

                        console.log(functionBlock);
                        viewCPUMetrics(functionBlock, 'cpu');
                    });
                });

                 // Add event listeners to the download buttons
                document.querySelectorAll('.download-btn-mem').forEach(button => {
                    button.addEventListener('click', function () {
                        const functionName = button.getAttribute('data-function-name');
                        const functionBlock = [...document.querySelectorAll('.function-metric')]
                            .find(metric => metric.querySelector('h3').textContent.includes(functionName));

                        console.log(functionBlock);
                        viewMEMmetrics(functionBlock, 'mem');
                    });
                });
            });
    }

function viewCPUMetrics(functionBlock, metricType) {
    const functionName = functionBlock.querySelector('h3').textContent.replace('Function: ', '') + '_' + metricType;
    const profileCanvas = document.getElementById('profile-canvas');
    profileCanvas.src = `/cpu-metrics?name=${encodeURIComponent(functionName)}`;
}

function viewMEMmetrics(functionBlock, metricType) {
    const functionName = functionBlock.querySelector('h3').textContent.replace('Function: ', '') + '_' + metricType;
    const profileCanvas = document.getElementById('profile-canvas');
    profileCanvas.src = `/mem-metrics?name=${encodeURIComponent(functionName)}`;
}





//    function viewCPUMetrics(functionBlock, metricType) {
//         const functionName = functionBlock.querySelector('h3').textContent.replace('Function: ', '') + '_'+metricType;
//         fetch(`/cpu-metrics?name=${encodeURIComponent(functionName)}`, {
//             method: 'GET',
//         })
//         .then(response => {
//             console.log(response);
//         })
//         .catch(err => {
//             console.error('Failed to fetch profile:', err);
//         });
//     }

//     function viewMEMmetrics(functionBlock, metricType) {
//         const functionName = functionBlock.querySelector('h3').textContent.replace('Function: ', '') + '_'+metricType;
//         fetch(`/mem-metrics?name=${encodeURIComponent(functionName)}`, {
//             method: 'GET',
//         })
//         .then(response => {
//             console.log(response);
//         })
//         .catch(err => {
//             console.error('Failed to fetch profile:', err);
//         });
//     }

    function fetchAllMetrics() {
        const selectedUnit = Array.from(unitInputs).find(input => input.checked).value;
        fetchMetrics(selectedUnit);
        fetchFunctionMetrics(selectedUnit); // Fetch function metrics in parallel
    }

    function onUnitChange() {
        const selectedUnit = Array.from(unitInputs).find(input => input.checked).value;
        localStorage.setItem('selectedUnit', selectedUnit);
        fetchAllMetrics();
        resetRefreshTimer();
    }

    function startAutoRefresh() {
        const savedUnit = localStorage.getItem('selectedUnit') || 'KB';
        unitInputs.forEach(input => {
            if (input.value === savedUnit) {
                input.checked = true;
            }
        });
        fetchAllMetrics();
        startRefreshTimer();
    }

    function startRefreshTimer() {
        let timeLeft = refreshInterval / 1000; // Convert milliseconds to seconds

        function updateCountdown() {
            countdownElement.textContent = `Refreshing in: ${timeLeft}s`;
            if (timeLeft > 0) {
                timeLeft -= 1;
                refreshTimeout = setTimeout(updateCountdown, 1000); // Update every second
            } else {
                fetchAllMetrics();
                startRefreshTimer(); // Restart the timer after fetching metrics
            }
        }

        updateCountdown();
    }

    function resetRefreshTimer() {
        clearTimeout(refreshTimeout);
        startRefreshTimer();
    }

    function refreshMetrics() {
        fetchAllMetrics();
        resetRefreshTimer();
    }

    unitInputs.forEach(input => {
        input.addEventListener('change', onUnitChange);
    });

    refreshBtn.addEventListener('click', refreshMetrics);

    startAutoRefresh();
});