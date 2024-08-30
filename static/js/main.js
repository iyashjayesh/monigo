document.addEventListener('DOMContentLoaded', function() {
    const unitInputs = document.querySelectorAll('input[name="unit"]');
    // const countdownElement = document.getElementById('countdown');
    const refreshBtn = document.getElementById('refresh-btn');

    const countdownElement = document.getElementById('countdown');
    const intervalInput = document.getElementById('interval-input');
    const setIntervalBtn = document.getElementById('set-interval-btn');

    const intervalInMinutes = parseFloat(intervalInput.value);

    let refreshInterval = intervalInMinutes * 60 * 1000;

    // Function to convert minutes to seconds
    function convertMinutesToSeconds(minutes) {
        return minutes * 60;
    }

    // let refreshInterval = parseInt(intervalInput.value, 10) * 1000; // default 60 seconds
    let refreshTimeout;
    let countdown = refreshInterval / 1000; // countdown in seconds

    const serviceInfoContainer = document.getElementById('service-container');
    const functionsContainer = document.getElementById('functions-container');
    const serviceMetricsContainer = document.getElementById('service-metrics-container');
    const memoryContainer = document.getElementById('runtime-metrics-container');

    function fetchServiceInfo() {
        fetch(`/service-info`)
            .then(response => response.json())
            .then(data => {
                serviceInfoContainer.innerHTML = '';

                const serviceName = data.service_name;
                const goVersion = data.go_version;
                serviceStartTime = data.service_start_time;

                serviceInfoContainer.innerHTML = `
                            <div class="service-info">
                                <div class="info-item">
                                    <span class="label">Service Name:</span> ${serviceName}
                                </div>
                                <div class="info-item">
                                    <span class="label">Go Version:</span> ${goVersion}
                                </div>
                                <div class="info-item">
                                    <span class="label">Service Started at:</span> ${serviceStartTime}
                                </div>
                            </div>
                        `;
            })
    }

    function fetchMetrics(unit) {
        serviceMetricsContainer.innerHTML = `<div class="service-info">Fetching the data...</div>`;
        memoryContainer.innerHTML = `<div class="service-info">Fetching the data...</div>`;

        fetch(`/metrics?unit=${unit}`)
            .then(response => response.text())
            .then(data => {
                serviceMetricsContainer.innerHTML = '';
                memoryContainer.innerHTML = '';

                const metrics = JSON.parse(data);
                const rcount = metrics.requests;
                let reqCountStr = metrics.requests;
                let totalDurationStr = metrics.total_duration;
                const load = metrics.load;
                const cores = metrics.cores;
                const memoryUsed = metrics.memory_used;
                let uptime = metrics.uptime; // 2.700655333s

                serviceMetricsContainer.innerHTML = `
                        <div class="metrics-overview">
                            <div class="metric-box">
                                <div class="metric-title">Load</div>
                                <div class="metric-value">${load}</div>
                            </div>
                            <div class="metric-box">
                                <div class="metric-title">Cores</div>
                                <div class="metric-value">${cores}</div>
                            </div>
                            <div class="metric-box">
                                <div class="metric-title">Memory</div>
                                <div class="metric-value">${memoryUsed}</div>
                            </div>
                            <div class="metric-box">
                                <div class="metric-title">Uptime</div>
                                <div class="metric-value">${uptime}</div>
                            </div>
                            <div class="metric-box">
                                <div class="metric-title">Request Served <span class="info-icon"
                                        data-tooltip="The total number of requests handled by the service.">i</span></div>
                                <p class="metric-value"> ${rcount}</p>
                            </div>
                            <div class="metric-box">
                                <div class="metric-title">Total Duration <span class="info-icon"
                                        data-tooltip="The total time spent handling requests.">i</span></div>
                                <p class="metric-value">${totalDurationStr}</p>
                            </div>
                        </div>`;

                const goroutiuneCount = metrics.goroutines;
                const totalAlloc = metrics.total_alloc;
                const memoryAllocatedBySystem = metrics.sys;
                const heapAlloc = metrics.heap_alloc;
                const heapSys = metrics.heap_sys;


                let reqCount = parseInt(reqCountStr, 10);
                let totalDuration = parseFloat(totalDurationStr.replace('ms', ''));

                let avgAPIResponseTime = reqCount > 0 ? (totalDuration / reqCount) : 0;
                let avgAPIResponseTimeStr;

                if (avgAPIResponseTime >= 3600000) { // More than or equal to 1 hour
                    avgAPIResponseTimeStr = (avgAPIResponseTime / 3600000).toFixed(3) + 'h';
                } else if (avgAPIResponseTime >= 60000) { // More than or equal to 1 minute
                    avgAPIResponseTimeStr = (avgAPIResponseTime / 60000).toFixed(3) + 'm';
                } else if (avgAPIResponseTime >= 1000) { // More than or equal to 1 second
                    avgAPIResponseTimeStr = (avgAPIResponseTime / 1000).toFixed(3) + 's';
                } else { // Less than 1 second
                    avgAPIResponseTimeStr = avgAPIResponseTime.toFixed(3) + 'ms';
                }

                memoryContainer.innerHTML = `
                        <div class="metrics-overview">
                            <div class="metric-box">
                                <div class="metric-title">Goroutines<span class="info-icon" data-tooltip="The number of goroutines currently running in the application.">i</span></div>
                                <p class="metric-value">${goroutiuneCount}</p>
                            </div>
                            <div class="metric-box">
                                <div class="metric-title">TotalAlloc <span class="info-icon" data-tooltip="Total memory allocated by the application since it started.">i</span>
                                </div>
                                <p class="metric-value">${totalAlloc}</p>
                            </div>
                            <div class="metric-box">
                                <div class="metric-title">Memory Allocated by System <span class="info-icon" data-tooltip="Total memory obtained from the OS.">i</span></div>
                                <p class="metric-value">${memoryAllocatedBySystem}</p>
                            </div>
                            <div class="metric-box">
                                <div class="metric-title">HeapAlloc <span class="info-icon" data-tooltip="Memory allocated to the heap.">i</span></div>
                                <p class="metric-value">${heapAlloc}</p>
                            </div>
                            <div class="metric-box">
                                <div class="metric-title">HeapSys: <span class="info-icon" data-tooltip="Total memory obtained from the OS for the heap.">i</span></div>
                                <p class="metric-value">${heapSys}</p>
                            </div>
                            <div class="metric-box">
                                <div class="metric-title">Average API Response Time <span class="info-icon" data-tooltip="The average time taken to respond to an API request.">i</span></div>
                                <p class="metric-value">${avgAPIResponseTimeStr}</p>
                            </div>
                        </div>`;
            })
            .catch(error => {
                console.error('Error:', error);
                serviceMetricsContainer.innerHTML = `<div class="error">Failed to load service metrics.</div>`;
                memoryContainer.innerHTML = `<div class="error">Failed to load memory metrics.</div>`;
            });

        // Reset the countdown
        countdown = refreshInterval / 1000;
    }

    function startCountdown() {
        countdownElement.textContent = `Refreshing in ${countdown} seconds`;

        refreshTimeout = setInterval(() => {
            countdown--;
            countdownElement.textContent = `Refreshing in ${countdown} seconds`;

            if (countdown <= 0) {
                clearInterval(refreshTimeout);
                fetchAllMetrics(); // Fetch all metrics when countdown reaches 0
                startCountdown(); // Restart countdown
            }
        }, 1000); // Update countdown every second
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
                    button.addEventListener('click', function() {
                        const functionName = button.getAttribute('data-function-name');
                        const functionBlock = [...document.querySelectorAll('.function-metric')]
                            .find(metric => metric.querySelector('h3').textContent.includes(functionName));

                        console.log(functionBlock);
                        viewCPUMetrics(functionBlock, 'cpu');
                    });
                });

                // Add event listeners to the download buttons
                document.querySelectorAll('.download-btn-mem').forEach(button => {
                    button.addEventListener('click', function() {
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
        profileCanvas.src = `/generate-function-metrics?name=${encodeURIComponent(functionName)}`;
    }

    function viewMEMmetrics(functionBlock, metricType) {
        const functionName = functionBlock.querySelector('h3').textContent.replace('Function: ', '') + '_' + metricType;
        const profileCanvas = document.getElementById('profile-canvas');
        profileCanvas.src = `/generate-function-metrics?name=${encodeURIComponent(functionName)}`;
    }

    function fetchAllMetrics() {
        const selectedUnit = Array.from(unitInputs).find(input => input.checked).value;
        fetchServiceInfo();
        fetchMetrics(selectedUnit);
        // fetchFunctionMetrics(selectedUnit); // Fetch function metrics in parallel
    }

    function onUnitChange() {
        const selectedUnit = Array.from(unitInputs).find(input => input.checked).value;
        localStorage.setItem('selectedUnit', selectedUnit);
        fetchAllMetrics();
        clearInterval(refreshTimeout); // Stoping the current countdown
    }

    function startAutoRefresh() {
        const savedUnit = localStorage.getItem('selectedUnit') || 'KB';
        unitInputs.forEach(input => {
            if (input.value === savedUnit) {
                input.checked = true;
            }
        });
        fetchAllMetrics();
        startCountdown();
    }

    function refreshMetrics() {
        clearInterval(refreshTimeout); // Stopping the current countdown
        fetchAllMetrics(); // Fetching metrics immediately
        startCountdown(); // Restarting the countdown
    }

    unitInputs.forEach(input => {
        input.addEventListener('change', onUnitChange);
    });

    refreshBtn.addEventListener('click', refreshMetrics);

    setIntervalBtn.addEventListener('click', () => {
        clearInterval(refreshTimeout); // Stop the current countdown
        refreshInterval = parseInt(intervalInput.value, 10) * 1000; // Update interval
        fetchServiceInfo(); // Fetch service info immediately
        fetchMetrics(); // Fetch metrics immediately
        startCountdown(); // Restart countdown with new interval
    });

    // Echart chart
    var myChart = echarts.init(document.getElementById('chartContainer'));
    var option = {
        xAxis: {
            type: 'category',
            data: [],
        },
        yAxis: {
            type: 'value',
        },
        series: [{
            data: [],
            type: 'line',
        }]
    };

    myChart.setOption(option);

    function updateChart(data) {
        const values = data.map(item => item.Value);
        const timestamps = data.map(item => new Date(item.Timestamp * 1000).toLocaleString());

        myChart.setOption({
            xAxis: {
                data: timestamps,
            },
            series: [{
                data: values,
                type: 'line',
                label: {
                    show: true,
                    position: 'top',
                    color: '#333',
                    fontSize: 18,
                },
            }]
        });
    }

    function fetchAndUpdateChart(startTime, endTime) {
        fetch('/service-metrics', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    field_name: 'load_metrics',
                    start_time: startTime,
                    end_time: endTime,
                }),
            })
            .then(response => response.json())
            .then(data => {
                updateChart(data);
            });
    }

    document.getElementById('timeFilter').addEventListener('change', function() {
        const filter = this.value;
        let startTime, endTime;

        const now = new Date();
        if (filter === 'custom') {
            document.getElementById('customTimeRange').style.display = 'block';
        } else {
            document.getElementById('customTimeRange').style.display = 'none';
            switch (filter) {
                case 'seconds':
                    startTime = new Date(now.getTime() - 1000 * 60).toISOString(); // Last 60 seconds
                    break;
                case 'minutes':
                    startTime = new Date(now.getTime() - 1000 * 60 * 60).toISOString(); // Last 60 minutes
                    break;
                case 'hours':
                    startTime = new Date(now.getTime() - 1000 * 60 * 60 * 24).toISOString(); // Last 24 hours
                    break;
                case 'days':
                    startTime = new Date(now.getTime() - 1000 * 60 * 60 * 24 * 7).toISOString(); // Last 7 days
                    break;
                case 'months':
                    startTime = new Date(now.getTime() - 1000 * 60 * 60 * 24 * 30).toISOString(); // Last 30 days
                    break;
                default:
                    startTime = now.toISOString();
            }

            endTime = now.toISOString();
            fetchAndUpdateChart(startTime, endTime);
        }
    });

    document.getElementById('applyCustomRange').addEventListener('click', function() {
        const startTime = new Date(document.getElementById('startTime').value).toISOString();
        const endTime = new Date(document.getElementById('endTime').value).toISOString();

        fetchAndUpdateChart(startTime, endTime);
    });

    document.getElementById('timeFilter').dispatchEvent(new Event('change'));

    startAutoRefresh();
});