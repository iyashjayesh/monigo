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

                const goroutinesChartChartObj = echarts.init(goroutinesChart);
                const time = rawData.map(entry => entry.time);
                const goroutines = rawData.map(entry => entry.value.goroutines);

                const option = {
                    title: {
                        text: 'Goroutines Metrics for last 1 hour',
                        left: 'center'
                    },
                    tooltip: {
                        trigger: 'axis'
                    },
                    legend: {
                        data: ['Goroutines'],
                        top: 30
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
                        data: time
                    },
                    yAxis: {
                        type: 'value'
                    },
                    series: [{
                        name: 'Goroutines',
                        type: 'line',
                        data: goroutines
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
                    } else {
                        downloadBtn.style.display = 'none';
                    }
                }

                countElement.textContent = goroutines.length;
                container.innerHTML = '';

                // Iterate over each goroutine and create HTML content
                goroutines.forEach(goroutine => {
                    const div = document.createElement('div');
                    div.className = 'goroutine';
                    div.innerHTML = `
                        <div class="goroutine-header">Goroutine ${goroutine.id}:</div>
                        <pre>${goroutine.stackTrace}</pre>
                    `;
                    container.appendChild(div);
                });

            }).catch(error => {
                console.error(error);
            });
    }
});