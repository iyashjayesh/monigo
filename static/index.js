document.addEventListener('DOMContentLoaded', () => {
    const DASHBOARD = document.getElementById('dashboard');
    const GOROUTINES_PAGE = document.getElementById('goroutines-page');
    const goRoutinesNumber = document.getElementById('goroutine-count');
    const serviceInfoContainer = document.getElementById('service-container');
    const refreshHtml = `
        <div class="loader-container">
            <div class="bouncing-dots">
                <div class="dot"></div>
                <div class="dot"></div>
                <div class="dot"></div>
            </div>
        </div>`;

         let countdownInterval;
        let refreshIntervalId;
        let refreshInterval; // To store the refresh interval in minutes
        let remainingTime;
    const elements = {
        goroutines: document.getElementById('goroutines'),
        serviceLoad: document.getElementById('service-load'),
        cores: document.getElementById('cores'),
        memory: document.getElementById('memory'),
        cpuUsage: document.getElementById('cpu-usage'),
        uptime: document.getElementById('uptime'),
        healthMessage: document.getElementById('health-message'),
        loadChart: document.getElementById('load-chart'),
        cpuChart: document.getElementById('cpu-chart'),
        memoryPieChart: document.getElementById('memory-pie-chart'),
        heapUsageChart: document.getElementById('heap-memory-chart')
    };

    Object.values(elements).forEach(el => el && (el.innerHTML = refreshHtml));

    if (GOROUTINES_PAGE) {
        fetchServiceInfo();
        fetchGoRoutines();
        startCountdown();
    } else if (DASHBOARD) {
        fetchMetrics();
        fetchServiceInfo();
        startCountdown();
    } else {
        console.warn('No valid page found');
    }

    

    function animateProgressBar(bar, targetWidth, duration) {
        let start = null;

        function step(timestamp) {
            if (!start) start = timestamp;
            const progress = timestamp - start;
            const width = Math.min((progress / duration) * targetWidth, targetWidth);
            bar.style.width = `${width}%`;

            if (width < targetWidth) {
                requestAnimationFrame(step);
            }
        }

        requestAnimationFrame(step);
    }

    function fetchServiceInfo() {
        fetch(`/service-info`)
            .then(response => response.json())
            .then(data => {
                serviceInfoContainer.innerHTML = '';
                const date = new Date(data.service_start_time);
                const formattedDate = date.toLocaleDateString('en-US', {
                    year: 'numeric',
                    month: 'long',
                    day: 'numeric'
                });
                const formattedTime = date.toLocaleTimeString('en-US', {
                    hour: 'numeric',
                    minute: 'numeric',
                    hour12: true
                });

                serviceInfoContainer.innerHTML = `
                    <div class="row pl-3 pr-3">
                        <div class="card card-block card-stretch card-height">
                            <div class="card-body">
                                <div class="d-flex align-items-center mb-4 card-total-sale">
                                    <div class="icon iq-icon-box-2 bg-info-light">
                                        <img src="../assets/images/product/1.png" class="img-fluid" alt="image">
                                    </div>
                                    <div>
                                        <p class="mb-2">Service Name: <h4>${data.service_name}</h4></p>
                                        <p class="mb-2">Go Version: <h4>${data.go_version}</h4></p>
                                        <p class="mb-2">Service Start Time: <h4>${formattedDate}<br/> ${formattedTime}</h4></p>
                                        <p class="mb-2">Process ID: <h4>${data.process_id}</h4></p>
                                    </div>
                                </div>
                                <div class="iq-progress-bar mt-2">
                                    <span class="bg-info iq-progress progress-1" data-percent="100"></span>
                                </div>
                            </div>
                        </div>
                    </div>`;

                const progressBars = serviceInfoContainer.querySelectorAll('.iq-progress');
                progressBars.forEach(bar => {
                    const percent = bar.getAttribute('data-percent');
                    animateProgressBar(bar, percent, 2000); // 2 seconds duration
                });
            });
    }

    function fetchGoRoutines() {
        fetch(`/go-routines-stats`)
            .then(response => response.json())
            .then(data => {
                console.log(data);
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

    function updateElement(element, label, value, info = '') {
        if (element) {
            element.innerHTML = `
                <div>
                    <p class="mb-2">${label} <span class="info-icon" data-tooltip="${info}">i</span></p>
                    <h4>${value}</h4>
                </div>`;
        } else {
            console.warn(`Element for ${label} not found`);
        }
    }

    function fetchMetrics() {
        fetch(`/metrics`)
            .then(response => response.json())
            .then(data => {
                const {
                    core_statistics,
                    load_statistics,
                    cpu_statistics,
                    memory_statistics,
                    overall_health
                } = data;

                updateElement(elements.goroutines, 'Go Routines:', core_statistics?.goroutines ?? 'N/A', 'Number of goroutines that are currently running');
                updateElement(elements.serviceLoad, 'Load:', `${load_statistics?.overall_load_of_service ?? 'N/A'}`, 'The load average of the system');
                updateElement(elements.cores, 'Cores:', `${cpu_statistics?.cores_used_by_service ?? 'N/A'} / ${cpu_statistics?.total_cores ?? 'N/A'}`, 'Number of CPU cores');
                updateElement(elements.memory, 'Memory:', `${memory_statistics?.memory_used_by_service ?? 'N/A'}`, 'Memory used by the service');
                updateElement(elements.cpuUsage, 'CPU Usage:', `${cpu_statistics?.cores_used_by_service_in_percent ?? 'N/A'}`, 'CPU usage of the service');
                updateElement(elements.uptime, 'Uptime:', core_statistics?.uptime ?? 'N/A', 'Uptime of the service');

                const healthIndicator = document.getElementById('health-indicator');
                if (overall_health.health.healthy) {
                    healthIndicator.classList.add('healthy');
                    document.getElementById('health-message').textContent = overall_health.health.message;
                } else {
                    healthIndicator.classList.add('unhealthy');
                    document.getElementById('health-message').textContent = overall_health.health.message;
                }

                renderCharts(data);
            })
            .catch(error => {
                console.error('Error fetching metrics:', error);
            });
    }

    // KB
    function renderCharts(data) {
        const charts = {
            loadChart: echarts.init(elements.loadChart),
            cpuChart: echarts.init(elements.cpuChart),
            memoryPieChart: echarts.init(elements.memoryPieChart),
            heapUsageChart: echarts.init(elements.heapUsageChart)
        };

        Object.values(charts).forEach(chart => chart.setOption({
            title: {
                text: 'Loading...'
            },
            tooltip: {}
        }));

        const {
            load_statistics,
            cpu_statistics,
            memory_statistics
        } = data;

        // Load Chart
        charts.loadChart.setOption({
            title: {
                text: 'Load Statistics'
            },
            tooltip: {
                trigger: 'axis',
                axisPointer: {
                    type: 'shadow'
                },
                formatter: params => {
                    return params.map(param => {
                        const {
                            axisValueLabel,
                            value
                        } = param;
                        let info = '';
                        if (value > 90) info = '[Critical Load]';
                        else if (value > 80) info = '[High Load]';
                        else if (value > 50) info = '[Moderate Load]';
                        else if (value <= 30) info = '[Healthy]';
                        return `${axisValueLabel}: ${value} %<br/><span>${info}</span>`;
                    }).join('<br/>');
                }
            },
            xAxis: {
                type: 'category',
                data: ['Service CPU Load', 'System CPU Load', 'Total CPU Load', 'Service Memory Load', 'System Memory Load']
            },
            yAxis: {
                type: 'value',
                max: 100
            },
            series: [{
                data: [
                    parseFloat(load_statistics.service_cpu_load),
                    parseFloat(load_statistics.system_cpu_load),
                    parseFloat(load_statistics.total_cpu_load),
                    parseFloat(load_statistics.service_memory_load),
                    parseFloat(load_statistics.system_memory_load)
                ],
                type: 'bar',
                itemStyle: {
                    color: params => {
                        const value = params.value;
                        if (value > 90) return 'red';
                        if (value > 80) return 'orange';
                        if (value > 50) return 'yellow';
                        return 'green';
                    }
                },
                emphasis: {
                    itemStyle: {
                        shadowBlur: 10,
                        shadowOffsetX: 0,
                        shadowColor: 'rgba(0, 0, 0, 0.5)'
                    }
                }
            }]
        });

        // CPU Chart
        charts.cpuChart.setOption({
            title: {
                text: 'CPU Statistics'
            },
            tooltip: {
                trigger: 'item',
                formatter: '{a} <br/>{b} : {c} ({d}%)'
            },
            legend: {
                orient: 'horizontal',
                center: 0,
                padding: [30, 0, 0, 0],
                data: [{
                        name: 'Cores Used by Service',
                        icon: 'rect',
                        itemStyle: {
                            color: '#00A1E4'
                        }
                    },
                    {
                        name: 'Cores Used by System',
                        icon: 'rect',
                        itemStyle: {
                            color: '#FF6F61'
                        }
                    },
                    {
                        name: 'Total Cores',
                        icon: 'rect',
                        itemStyle: {
                            color: '#FFD166'
                        }
                    }
                ]
            },
            series: [{
                name: 'CPU Usage',
                type: 'pie',
                radius: '55%',
                center: ['50%', '50%'],
                data: [{
                        value: cpu_statistics.cores_used_by_service,
                        name: 'Cores Used by Service'
                    },
                    {
                        value: cpu_statistics.cores_used_by_system,
                        name: 'Cores Used by System'
                    },
                    {
                        value: cpu_statistics.total_cores,
                        name: 'Total Cores'
                    }
                ],
                emphasis: {
                    itemStyle: {
                        shadowBlur: 10,
                        shadowOffsetX: 0,
                        shadowColor: 'rgba(0, 0, 0, 0.5)'
                    }
                }
            }]
        });

        // Memory Pie Chart
        charts.memoryPieChart.setOption({
            title: {
                text: 'Memory Distribution',
            },
            tooltip: {
                trigger: 'item',
                formatter: function(params) {
                    return `${params.seriesName}<br/>${params.name}: ${params.value} (${params.percent}%) []`;
                }
            },
            legend: {
                orient: 'vertical',
                left: 'left',
                padding: [30, 0, 0, 0],
                data: ['Memory Used by Service', 'Memory Used by System', 'Available Memory']
            },
            series: [{
                name: 'Memory Usage',
                type: 'pie',
                radius: '55%',
                center: ['50%', '60%'],
                data: [{
                        value: parseFloat(data.memory_statistics.memory_used_by_service),
                        name: 'Memory Used by Service'
                    },
                    {
                        value: parseFloat(data.memory_statistics.memory_used_by_system),
                        name: 'Memory Used by System'
                    },
                    {
                        value: parseFloat(data.memory_statistics.available_memory),
                        name: 'Available Memory'
                    }
                ],
                emphasis: {
                    itemStyle: {
                        shadowBlur: 10,
                        shadowOffsetX: 0,
                        shadowColor: 'rgba(0, 0, 0, 0.5)'
                    }
                }
            }]
        });

        let values = [];
        data.memory_statistics.mem_stats_records.forEach(record => {
            if (record.record_name === 'HeapAlloc') {
                values.push(record.record_value);
            } else if (record.record_name === 'HeapSys') {
                values.push(record.record_value);
            } else if (record.record_name === 'HeapIdle') {
                values.push(record.record_value);
            } else if (record.record_name === 'HeapInuse') {
                values.push(record.record_value);
            } else if (record.record_name === 'HeapReleased') {
                values.push(record.record_value);
            }

            if (values.length === 5) {
                return;
            }
        });

        // Heap Usage Chart
        charts.heapUsageChart.setOption({
            title: {
                text: 'Heap Memory Usage'
            },
            tooltip: {},
            xAxis: {
                type: 'category',
                data: ['HeapAlloc', 'HeapSys', 'HeapIdle', 'HeapInuse', 'HeapReleased']
            },
            yAxis: {
                type: 'value',
                name: 'MB'
            },
            series: [{
                name: 'Memory (MB)',
                type: 'bar',
                data: [
                    values[0],
                    values[1],
                    values[2],
                    values[3],
                    values[4],
                ]
            }]
        });
    }


    const chart = echarts.init(document.getElementById('chart'));

    const timeRanges = {
        '15m': 15,
        '30m': 30,
        '1h': 60,
        '6h': 360,
        '1d': 1440,
        '3d': 4320,
        '1m': 43200 // Approximate minutes in a month
    };

    // Generate mock time-series data
    function generateMockData(metric, durationInMinutes) {
        const dataPoints = [];
        const now = new Date();
        const interval = Math.floor(durationInMinutes / 60); // Generate data points every minute
        const totalPoints = durationInMinutes;

        for (let i = totalPoints; i >= 0; i--) {
            const timestamp = new Date(now.getTime() - i * 60000); // Subtract i minutes
            dataPoints.push({
                time: timestamp,
                value: getRandomValue(metric)
            });
        }
        return dataPoints;
    }

    // Function to get random values based on metric type
    function getRandomValue(metric) {
        switch (metric) {
            case 'heap':
                return {
                    HeapAlloc: getRandomArbitrary(4, 8),
                        HeapSys: getRandomArbitrary(10, 15),
                        HeapInuse: getRandomArbitrary(5, 10),
                        HeapIdle: getRandomArbitrary(3, 7),
                        HeapReleased: getRandomArbitrary(2, 5)
                };
            case 'stack':
                return {
                    StackInuse: getRandomArbitrary(600, 800),
                        StackSys: getRandomArbitrary(600, 800)
                };
            case 'gc':
                return {
                    PauseTotalNs: getRandomArbitrary(50, 150),
                        NumGC: getRandomInt(1, 10),
                        GCCPUFraction: getRandomArbitrary(0.0001, 0.005)
                };
            case 'misc':
                return {
                    MSpanInuse: getRandomArbitrary(80, 120),
                        MSpanSys: getRandomArbitrary(100, 130),
                        MCacheInuse: getRandomArbitrary(10, 20),
                        MCacheSys: getRandomArbitrary(12, 25),
                        BuckHashSys: getRandomArbitrary(1, 2),
                        GCSys: getRandomArbitrary(2, 5),
                        OtherSys: getRandomArbitrary(1, 3)
                };
            default:
                return {};
        }
    }

    // Utility functions to generate random numbers
    function getRandomArbitrary(min, max) {
        return +(Math.random() * (max - min) + min).toFixed(2);
    }

    function getRandomInt(min, max) {
        return Math.floor(Math.random() * (max - min + 1)) + min;
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

    function fetchDataPointsFromServer(metricName, timeRange){
        // const dataPoints = [];
        // let StartTime = new Date();
        // let EndTime = new Date();
        // if (timeRange == "15m") {
        //     StartTime = new Date(new Date().getTime() - 15 * 60000); // Subtract i minutes
        // } else if (timeRange == "30m") {
        //     StartTime = new Date(new Date().getTime() - 30 * 60000); // Subtract i minutes
        // } else if (timeRange == "1h") {
        //     StartTime = new Date(new Date().getTime() - 60 * 60000); // Subtract i minutes
        // }

        // let metricList = [];
        // if (metricName == "heap") {
        //     metricList = ["heap_alloc", "heap_sys", "heap_inuse", "heap_idle", "heap_released"];
        // } else if (metricName == "stack") {
        //     metricList = ["stack_inuse", "stack_sys"];
        // } else if (metricName == "gc") {
        //     metricList = ["pause_total_ns", "num_gc", "gc_cpu_fraction"];
        // } else if (metricName == "misc") {
        //     metricList = ["m_span_inuse", "m_span_sys", "m_cache_inuse", "m_cache_sys", "buck_hash_sys", "gc_sys", "other_sys"];
        // }

        // let data = {
        //     field_name: metricList,
        //     start_time: StartTime.toISOString(),
        //     end_time: EndTime.toISOString()
        // };

        const dataPoints = [];
        let StartTime = new Date();
        let EndTime = new Date();

        if (timeRange == "5m") {
            StartTime = new Date(new Date().getTime() - 5 * 60000); // Subtract 5 minutes
        } else if (timeRange == "15m") {
            StartTime = new Date(new Date().getTime() - 15 * 60000); // Subtract 15 minutes
        } else if (timeRange == "30m") {
            StartTime = new Date(new Date().getTime() - 30 * 60000); // Subtract 30 minutes
        } else if (timeRange == "1h") {
            StartTime = new Date(new Date().getTime() - 60 * 60000); // Subtract 1 hour
        } else if (timeRange == "6h") {
            StartTime = new Date(new Date().getTime() - 360 * 60000); // Subtract 6 hours
        } else if (timeRange == "1d") {
            StartTime = new Date(new Date().getTime() - 1440 * 60000); // Subtract 1 day
        } else if (timeRange == "3d") {
            StartTime = new Date(new Date().getTime() - 4320 * 60000); // Subtract 3 days
        } else if (timeRange == "7d") {
            StartTime = new Date(new Date().getTime() - 10080 * 60000); // Subtract 7 days
        } else if (timeRange == "1m") {
            StartTime = new Date(new Date().getTime() - 43200 * 60000); // Subtract 1 month
        }
        

        let metricList = [];
        if (metricName == "heap") {
            metricList = ["heap_alloc", "heap_sys", "heap_inuse", "heap_idle", "heap_released"];
        } else if (metricName == "stack") {
            metricList = ["stack_inuse", "stack_sys"];
        } else if (metricName == "gc") {
            metricList = ["pause_total_ns", "num_gc", "gc_cpu_fraction"];
        } else if (metricName == "misc") {
            metricList = ["m_span_inuse", "m_span_sys", "m_cache_inuse", "m_cache_sys", "buck_hash_sys", "gc_sys", "other_sys"];
        }

        let data = {
            field_name: metricList,
            timerange: timeRange,
            start_time: toLocalISOString(StartTime),
            end_time: toLocalISOString(EndTime)
        };
        console.log('Fetching data for metric:', metricName, 'and time range:', timeRange);
        console.log('API REQ:', data);

        fetch(`/service-metrics`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
        }).then(response => response.json())
            .then(data => {
                console.log('API RES:', data);

                let rawData = [];
                for (let i = 0; i < data.length; i++) {
                    const timestamp = new Date(data[i].time); 
                    rawData.push({
                        time: timestamp,
                        value: data[i].value
                    });
                }


                console.log('New Constructed DATA:', rawData);
                const seriesData = {};
                rawData.forEach(dataPoint => {
                    const timeLabel = dataPoint.time.toLocaleString();
                    for (let key in dataPoint.value) {
                        if (!seriesData[key]) {
                            seriesData[key] = [];
                        }
                        seriesData[key].push([timeLabel, dataPoint.value[key]]);
                    }
                });

                const series = [];
                for (let key in seriesData) {
                    series.push({
                        name: key,
                        type: 'line',
                        data: seriesData[key],
                        smooth: true
                    });
                }

                chart.setOption({
                    title: {
                        text: getMetricTitle(metricName),
                        left: 'center'
                    },
                    tooltip: {
                        trigger: 'axis'
                    },
                    legend: {
                        top: 30,
                        data: Object.keys(seriesData)
                    },
                    xAxis: {
                        type: 'category',
                        boundaryGap: false,
                        data: rawData.map(d => d.time.toLocaleString()),
                        axisLabel: {
                            formatter: function(value) {
                                return value.split(' ')[1]; // Show only time
                            }
                        }
                    },
                    yAxis: {
                        type: 'value',
                        axisLabel: {
                            formatter: function(value) {
                                return formatYAxisLabel(metricName, value);
                            }
                        }
                    },
                    series: series
                });
            })
            .catch((error) => {
                console.error('Error:', error);
            });
    }

    // // Function to prepare chart options based on selected metric and time range
    // function getChartOptions(metric, timeRange) {
    //     fetchDataPointsFromServer(metric, timeRange);
    // }

    // Helper function to get chart title based on metric
    function getMetricTitle(metric) {
        switch (metric) {
            case 'heap':
                return 'Heap Memory Usage Over Time';
            case 'stack':
                return 'Stack Memory Usage Over Time';
            case 'gc':
                return 'Garbage Collection Over Time';
            case 'misc':
                return 'Miscellaneous System Memory Over Time';
            default:
                return '';
        }
    }

    // Helper function to format Y-axis labels
    function formatYAxisLabel(metric, value) {
        if (metric === 'heap' || metric === 'misc') {
            return `${value} KB`;
        } else if (metric === 'stack' || (metric === 'gc' && value > 1)) {
            return `${value} KB`;
        } else if (metric === 'gc' && value <= 1) {
            return value.toFixed(4);
        } else {
            return value;
        }
    }

    // Function to update chart based on selections
    function updateChart() {
        const metricSelect = document.getElementById('metric-select').value;
        const timeSelect = document.getElementById('time-select').value;
        fetchDataPointsFromServer(metricSelect, timeSelect);   
    }

    // Event listeners for dropdown changes
    document.getElementById('metric-select').addEventListener('change', updateChart);
    document.getElementById('time-select').addEventListener('change', updateChart);
    updateChart();

    // Responsive behavior
    window.addEventListener('resize', function() {
        chart.resize();
    });


    //////// Refreesh 
   

        // // Function to fetch data and update the chart
        // function updateChart() {
        //     startCountdown(); // Restart countdown after updating chart
        // }

        // Function to start the countdown
        function startCountdown() {
            clearInterval(countdownInterval); // Clear any existing countdown
            const countdownDisplay = document.getElementById('refresh-countdown');
            remainingTime = refreshInterval * 60; // Convert minutes to seconds

            countdownInterval = setInterval(() => {
                const minutes = Math.floor(remainingTime / 60);
                const seconds = remainingTime % 60;
                countdownDisplay.textContent = `Refreshing in ${minutes}m ${seconds}s`;
                remainingTime--;

                if (remainingTime < 0) {
                    clearInterval(countdownInterval);
                }
            }, 1000);
        }

        // Function to start auto-refresh with the set interval
        function startAutoRefresh() {
            clearInterval(refreshIntervalId); // Clear any existing interval
            refreshIntervalId = setInterval(updateChart, refreshInterval * 60 * 1000); // Convert minutes to milliseconds
            startCountdown(); // Start the countdown immediately after setting the interval
        }

        // Enable the "Set" button when the input is changed
        document.getElementById('refresh-interval').addEventListener('input', function () {
            document.getElementById('set-interval').disabled = false;
        });

        // Apply the refresh interval when "Set" is clicked
        document.getElementById('set-interval').addEventListener('click', function () {
            refreshInterval = parseInt(document.getElementById('refresh-interval').value, 10) || 5;
            refreshInterval = Math.max(1, Math.min(60, refreshInterval)); // Enforce min 1, max 60
            localStorage.setItem('refreshInterval', refreshInterval); // Store the value in localStorage
            startAutoRefresh();
            this.disabled = true; // Disable the button again until the next input change
        });

        // Load the refresh interval from localStorage or use default
        function loadRefreshInterval() {
            const storedInterval = localStorage.getItem('refreshInterval');
            if (storedInterval) {
                refreshInterval = parseInt(storedInterval, 10);
            } else {
                refreshInterval = 5; // Default value if nothing is stored
            }
            document.getElementById('refresh-interval').value = refreshInterval; // Update the input field
        }

        loadRefreshInterval();
        startAutoRefresh();
});