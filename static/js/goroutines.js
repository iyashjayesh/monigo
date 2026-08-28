document.addEventListener('DOMContentLoaded', () => {
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

    const goRoutinesNumber = document.getElementById('goroutine-count');

    if (goRoutinesNumber) {
        fetchGoRoutines();
    }

    // Function to get the local ISO string with timezone offset
    function toLocalISOString(date) {
        const tzOffset = -date.getTimezoneOffset(); // in minutes
        const diff = tzOffset >= 0 ? '+' : '-';
        const pad = (num) => `${Math.floor(Math.abs(num))}`.padStart(2, '0');

        const offsetHours = pad(tzOffset / 60);
        const offsetMinutes = pad(tzOffset % 60);

        return date.getFullYear() +
            '-' + pad(date.getMonth() + 1) +
            '-' + pad(date.getDate()) +
            'T' + pad(date.getHours()) +
            ':' + pad(date.getMinutes()) +
            ':' + pad(date.getSeconds()) +
            '.' + String((date.getMilliseconds() / 1000).toFixed(3)).slice(2, 5) +
            diff + offsetHours + ':' + offsetMinutes;
    }

    const goroutinesChart = document.getElementById('goroutines-chart');
    let goroutinesChartChartObj = null;

    function fetchDataPointsFromServer() {
        let StartTime = new Date();
        let EndTime = new Date();

        StartTime = new Date(new Date().getTime() - 60 * 60000); // Subtract 1 hour
        EndTime = new Date(); // Current time
        metricList = ["goroutines"];

        let data = {
            field_name: metricList,
            timerange: "1h",
            start_time: toLocalISOString(StartTime),
            end_time: toLocalISOString(EndTime)
        };

        authenticatedFetch(`/monigo/api/v1/service-metrics`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
        }).then(response => response.json())
            .then(data => {
                let rawData = [];
                for (let i = 0; i < data.length; i++) {
                    const timestamp = new Date(data[i].time);
                    rawData.push({
                        time: timestamp,
                        value: data[i].value
                    });
                }

                window.lastGoroutinesRawData = rawData;

                const isDark = document.body.classList.contains('dark-theme');
                const textColor = isDark ? '#9ca3af' : '#333';
                const gridLineColor = isDark ? '#1f2937' : '#eee';

                if (!goroutinesChartChartObj && goroutinesChart) {
                    goroutinesChartChartObj = echarts.init(goroutinesChart);
                }
                const time = rawData.map(entry => entry.time.toLocaleTimeString());
                const goroutines = rawData.map(entry => entry.value.goroutines);

                const option = {
                    backgroundColor: 'transparent',
                    title: {
                        text: 'Goroutines Metrics for last 1 hour',
                        left: 'center',
                        textStyle: { color: isDark ? '#f3f4f6' : '#333' }
                    },
                    tooltip: {
                        trigger: 'axis'
                    },
                    legend: {
                        data: ['Goroutines'],
                        top: 30,
                        textStyle: { color: textColor }
                    },
                    grid: {
                        left: '3%',
                        right: '4%',
                        bottom: '3%',
                        containLabel: true
                    },
                    xAxis: {
                        type: 'category',
                        boundaryGap: false,
                        data: time,
                        axisLabel: { color: textColor },
                        axisLine: { lineStyle: { color: gridLineColor } }
                    },
                    yAxis: {
                        type: 'value',
                        axisLabel: { color: textColor },
                        splitLine: { lineStyle: { color: gridLineColor } }
                    },
                    series: [{
                        name: 'Goroutines',
                        type: 'line',
                        data: goroutines,
                        itemStyle: { color: '#ff5c35' }
                    }]
                };

                goroutinesChartChartObj.setOption(option);
            })
            .catch((error) => {
                console.error('Error:', error);
            });
    }

    function fetchGoRoutines() {
        authenticatedFetch(`/monigo/api/v1/go-routines-stats`)
            .then(response => response.json())
            .then(data => {
                goRoutinesNumber.innerHTML = data.number_of_goroutines;
                const container = document.getElementById('goroutines-container');
                const countElement = document.getElementById('goroutine-count');

                let goroutines = [];
                data.stack_view.forEach((item, index) => {
                    const goroutine = {
                        id: index + 1,
                        stackTrace: item
                    };
                    goroutines.push(goroutine);
                });

                fetchDataPointsFromServer();

                if (goroutines.length > 0) {
                    const downloadBtn = document.getElementById('download-stack-view');
                    if (downloadBtn) {
                        downloadBtn.style.display = 'block';
                        downloadBtn.addEventListener('click', () => {
                            const blob = new Blob([goroutines.map(g => g.stackTrace).join('\n')], {
                                type: 'text/plain'
                            });
                            const url = URL.createObjectURL(blob);
                            const a = document.createElement('a');
                            a.href = url;
                            a.download = 'go-routines-stack-view.txt';
                            a.click();
                            URL.revokeObjectURL(url);
                        });
                    }
                }

                countElement.textContent = goroutines.length;
                container.innerHTML = '';

                // Group identical stack traces (ignoring the specific goroutine ID / state line)
                const groups = {};
                goroutines.forEach(g => {
                    const lines = g.stackTrace.trim().split('\n');
                    const header = lines[0] || '';
                    const callStack = lines.slice(1).join('\n');

                    let state = "unknown";
                    const stateMatch = header.match(/\[(.*)\]/);
                    if (stateMatch) {
                        state = stateMatch[1];
                    }

                    if (!groups[callStack]) {
                        groups[callStack] = {
                            callStack: callStack,
                            state: state,
                            headers: [],
                            count: 0
                        };
                    }
                    groups[callStack].headers.push(header);
                    groups[callStack].count++;
                });

                // Sort groups by count descending
                const sortedGroups = Object.values(groups).sort((a, b) => b.count - a.count);

                // Render with virtualization / safety limit of 50 groups
                const displayLimit = 50;
                const displayedGroups = sortedGroups.slice(0, displayLimit);

                displayedGroups.forEach((group, index) => {
                    const groupDiv = document.createElement('div');
                    groupDiv.className = 'goroutine-card card mb-3';

                    const headerDiv = document.createElement('div');
                    headerDiv.className = 'card-header d-flex justify-content-between align-items-center cursor-pointer';
                    headerDiv.style.userSelect = 'none';
                    headerDiv.innerHTML = `
                        <h6 class="mb-0">
                            <i class="fa fa-chevron-right mr-2 transition-transform" style="transition: transform 0.2s;"></i>
                            <strong>${group.count} Goroutines</strong> in state [${group.state}]
                        </h6>
                        <span class="badge badge-primary">${group.count}</span>
                    `;

                    const bodyDiv = document.createElement('div');
                    bodyDiv.className = 'card-body d-none';
                    bodyDiv.innerHTML = `
                        <p class="text-muted small mb-2">Individual Goroutines: ${group.headers.join(', ')}</p>
                        <pre style="margin-bottom: 0; padding: 10px; border-radius: 4px;">${group.callStack}</pre>
                    `;

                    headerDiv.addEventListener('click', () => {
                        const isHidden = bodyDiv.classList.contains('d-none');
                        const icon = headerDiv.querySelector('i');
                        if (isHidden) {
                            bodyDiv.classList.remove('d-none');
                            icon.style.transform = 'rotate(90deg)';
                        } else {
                            bodyDiv.classList.add('d-none');
                            icon.style.transform = 'rotate(0deg)';
                        }
                    });

                    groupDiv.appendChild(headerDiv);
                    groupDiv.appendChild(bodyDiv);
                    container.appendChild(groupDiv);
                });

                if (sortedGroups.length > displayLimit) {
                    const alertDiv = document.createElement('div');
                    alertDiv.className = 'alert alert-warning mt-3';
                    alertDiv.role = 'alert';
                    alertDiv.textContent = `Showing top ${displayLimit} goroutine groups to prevent performance degradation. There are ${sortedGroups.length - displayLimit} more groups not displayed.`;
                    container.appendChild(alertDiv);
                }

            }).catch(error => {
                console.error(error);
            });
    }

    document.addEventListener('monigoThemeChanged', () => {
        if (window.lastGoroutinesRawData && goroutinesChart) {
            if (!goroutinesChartChartObj) {
                goroutinesChartChartObj = echarts.init(goroutinesChart);
            }
            const isDark = document.body.classList.contains('dark-theme');
            const textColor = isDark ? '#9ca3af' : '#333';
            const gridLineColor = isDark ? '#1f2937' : '#eee';
            const time = window.lastGoroutinesRawData.map(entry => entry.time.toLocaleTimeString());
            const goroutines = window.lastGoroutinesRawData.map(entry => entry.value.goroutines);
            
            goroutinesChartChartObj.setOption({
                title: {
                    textStyle: { color: isDark ? '#f3f4f6' : '#333' }
                },
                legend: {
                    textStyle: { color: textColor }
                },
                xAxis: {
                    data: time,
                    axisLabel: { color: textColor },
                    axisLine: { lineStyle: { color: gridLineColor } }
                },
                yAxis: {
                    axisLabel: { color: textColor },
                    splitLine: { lineStyle: { color: gridLineColor } }
                }
            });
        }
    });
});