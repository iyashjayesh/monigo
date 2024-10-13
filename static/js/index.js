document.addEventListener('DOMContentLoaded', () => {
    const DASHBOARD = document.getElementById('dashboard');
    const GOROUTINES_PAGE = document.getElementById('goroutines-page');

    const process_id = document.getElementById('process_id');
    const service_name = document.getElementById('service_name');
    const go_version = document.getElementById('go_version');
    const service_start_time = document.getElementById('service_start_time');

    const dlMonigoDashboardBtn = getDlBtn('download-image-btn');
    const dlLoadStatisticsBtn = getDlBtn('download-load-statistics-btn');
    const dlCpuStatisticsBtn = getDlBtn('download-cpu-statistics-btn');

    const dlMemoryDistributionBtn = getDlBtn('download-memory-statistics-btn');
    const dlHeapMemoryUsageBtn = getDlBtn('download-heap-statistics-btn');
    const dlHistoricChartBtn = getDlBtn('download-history-statistics-btn');

    const dlCpuUsageMetricsBtn = getDlBtn('download-cpu-usage-btn');
    const dlGoRoutinesMetricsBtn = getDlBtn(
        'download-go-routines-statistics-btn'
    );
    const dlLoadMemoryMetricsBtn = getDlBtn(
        'download-load-memory-statistics-btn'
    );
    const dlHealthMetricsBtn = getDlBtn('download-health-statistics-btn');

    const refreshHtml = `
        <div class="loader-container mt-3">
            <div class="bouncing-dots">
                <div class="dot"></div>
                <div class="dot"></div>
                <div class="dot"></div>
            </div>
        </div>`;

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
        heapUsageChart: document.getElementById('heap-memory-chart'),
        serviceHealthTag: document.getElementById('service-health-tag'),
        systemHealthTag: document.getElementById('system-health-tag'),
        memoryDetailButton: document.getElementById('memory-detail-button'),
        coreUsageDetailButton: document.getElementById(
            'core-usage-detail-button'
        ),
        loadUsageDetailButton: document.getElementById(
            'load-usage-detail-button'
        )
    };

    Object.values(elements).forEach((el) => el && (el.innerHTML = refreshHtml));

    if (GOROUTINES_PAGE) {
        fetchServiceInfo();
    } else if (DASHBOARD) {
        fetchMetrics();
        fetchServiceInfo();
    } else {
        console.warn('No valid page found');
    }

    function fetchServiceInfo() {
        fetch(`/monigo/api/v1/service-info`)
            .then((response) => response.json())
            .then((data) => {
                service_name.innerHTML = '';
                go_version.innerHTML = '';
                service_start_time.innerHTML = '';
                process_id.innerHTML = '';
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

                service_name.innerHTML = `
                    <div class="d-flex align-items-center mb-4 card-total-sale">
                        <div>
                            <p class="mb-2">Service Name: <h4>${data.service_name}</h4></p>
                        </div>
                    </div>`;
                go_version.innerHTML = `
                    <div class="d-flex align-items-center mb-4 card-total-sale">
                        <div>
                            <p class="mb-2">Go Version: <h4>${data.go_version}</h4></p>
                        </div>
                    </div>`;
                service_start_time.innerHTML = `
                 <div>
                    <p class="mb-2">Service Start Time: <h4>${formattedDate}<br/> ${formattedTime}</h4></p>
                </div>`;

                process_id.innerHTML = `
                 <div>
                    <p class="mb-2">Process ID: <h4>${data.process_id}</h4></p>
                </div>`;
            });
    }

    function updateElement(element, label, value, info = '', obj) {
        if (element) {
            element.innerHTML = `
                <div>
                    <p class="mb-2">${label} <span class="info-icon" data-tooltip="${info}">i</span></p>
                    <h4>${value}</h4>
                </div>`;

            if (label == 'Memory:') {
                elements.memoryDetailButton.innerHTML = '';
                for (let i = 0; i < obj.mem_stats_records.length; i++) {
                    const record = obj.mem_stats_records[i];
                    const recordName = record.record_name;
                    const recordValue = record.record_value;
                    const recordUnit = record.record_unit;
                    const recordDescription = record.record_description;

                    elements.memoryDetailButton.innerHTML += `
                        <div class="d-flex align-items-center mb-1 card-total-sale">
                            <div>
                                <p class="mb-2">${recordName} <span class="info-icon" data-tooltip="${recordDescription}">i</span></p>
                                <h4>${recordValue} ${recordUnit}</h4>
                            </div>
                        </div>`;
                }
            }

            if (label == 'Cores:') {
                elements.coreUsageDetailButton.innerHTML = '';
                elements.coreUsageDetailButton.innerHTML = `
                    <div class="d-flex align-items-center mb-1 card-total-sale">
                        <div>
                            <p class="mb-2">Total Cores: <h4>${obj.total_cores}</h4></p>
                        </div>
                    </div>
                    <div class="d-flex align-items-center mb-1 card-total-sale">
                        <div>
                            <p class="mb-2">Total Logical Cores: <h4>${obj.total_logical_cores}</h4></p>
                        </div>
                    </div>
                    <div class="d-flex align-items-center mb-1 card-total-sale">
                        <div>
                            <p class="mb-2">Cores Used by System: <h4>${obj.cores_used_by_system}</h4></p>
                        </div>
                    </div>
                    <div class="d-flex align-items-center mb-1 card-total-sale">
                        <div>
                            <p class="mb-2">Cores Used by Service: <h4>${obj.cores_used_by_service}</h4></p>
                        </div>
                    </div>
                    <div class="d-flex align-items-center mb-1 card-total-sale">
                        <div>
                            <p class="mb-2">Cores Used by Service in Percent: <h4>${obj.cores_used_by_service_in_percent}</h4></p>
                        </div>
                    </div>
                    <div class="d-flex align-items-center mb-1 card-total-sale">
                        <div>
                            <p class="mb-2">Cores Used by System in Percent: <h4>${obj.cores_used_by_system_in_percent}</h4></p>
                        </div>
                    </div>`;
            }

            if (label == 'Load:') {
                elements.loadUsageDetailButton.innerHTML = '';
                elements.loadUsageDetailButton.innerHTML = `
                    <div class="d-flex align-items-center mb-1 card-total-sale">
                        <div>
                            <p class="mb-2">Service CPU Load: <h4>${obj.service_cpu_load}</h4></p>
                        </div>
                    </div>
                    <div class="d-flex align-items-center mb-1 card-total-sale">
                        <div>
                            <p class="mb-2">System CPU Load: <h4>${obj.system_cpu_load}</h4></p>
                        </div>
                    </div>
                    <div class="d-flex align-items-center mb-1 card-total-sale">
                        <div>
                            <p class="mb-2">Total CPU Load: <h4>${obj.total_cpu_load}</h4></p>
                        </div>
                    </div>
                    <div class="d-flex align-items-center mb-1 card-total-sale">
                        <div>
                            <p class="mb-2">Service Memory Load: <h4>${obj.service_memory_load}</h4></p>
                        </div>
                    </div>
                    <div class="d-flex align-items-center mb-1 card-total-sale">
                        <div>
                            <p class="mb-2">System Memory Load: <h4>${obj.system_memory_load}</h4></p>
                        </div>
                    </div>`;
            }
        } else {
            console.warn(`Element for ${label} not found`);
        }
    }

    function fetchMetrics() {
        fetch(`/monigo/api/v1/metrics`)
            .then((response) => response.json())
            .then((data) => {
                const {
                    core_statistics,
                    load_statistics,
                    cpu_statistics,
                    memory_statistics,
                    health
                } = data;

                updateGauge('g1', health);
                updateElement(
                    elements.goroutines,
                    'Go Routines:',
                    core_statistics?.goroutines ?? 'N/A',
                    'Number of goroutines that are currently running',
                    core_statistics
                );
                updateElement(
                    elements.serviceLoad,
                    'Load:',
                    `${load_statistics?.overall_load_of_service ?? 'N/A'}`,
                    'The load average of the system',
                    load_statistics
                );
                updateElement(
                    elements.cores,
                    'Cores:',
                    `${cpu_statistics?.cores_used_by_service ?? 'N/A'} / ${
                        cpu_statistics?.total_cores ?? 'N/A'
                    }`,
                    'Number of CPU cores',
                    cpu_statistics
                );
                updateElement(
                    elements.memory,
                    'Memory:',
                    `${memory_statistics?.memory_used_by_service ?? 'N/A'}`,
                    'Memory used by the service',
                    memory_statistics
                );
                updateElement(
                    elements.cpuUsage,
                    'CPU Usage:',
                    `${
                        cpu_statistics?.cores_used_by_service_in_percent ??
                        'N/A'
                    }`,
                    'CPU usage of the service',
                    cpu_statistics
                );
                updateElement(
                    elements.uptime,
                    'Uptime:',
                    core_statistics?.uptime ?? 'N/A',
                    'Uptime of the service',
                    core_statistics
                );

                const healthIndicator =
                    document.getElementById('health-indicator');
                if (health.service_health.healthy) {
                    healthIndicator.classList.add('healthy');
                    document.getElementById('health-message').textContent =
                        health.service_health.message;
                } else {
                    healthIndicator.classList.add('unhealthy');
                    document.getElementById('health-message').textContent =
                        health.service_health.message;
                }
                renderCharts(data);
                downloadDivAsImage(
                    dlMonigoDashboardBtn,
                    'content-page',
                    'Monigo'
                );
                downloadDivAsImage(
                    dlLoadStatisticsBtn,
                    'load-chart',
                    'load-statiistics'
                );
                downloadDivAsImage(
                    dlCpuStatisticsBtn,
                    'cpu-statistics',
                    'cpu-statistics'
                );
                downloadDivAsImage(
                    dlMemoryDistributionBtn,
                    'memory-pie-chart',
                    'memory-statistics'
                );
                downloadDivAsImage(
                    dlHeapMemoryUsageBtn,
                    'heap-memory-chart',
                    'heap-memory'
                );
                downloadDivAsImage(
                    dlHistoricChartBtn,
                    'chart',
                    'historic-statistics-overview'
                );
                downloadDivAsImage(
                    dlCpuUsageMetricsBtn,
                    'cpu-usage-chart',
                    'cpu-usage'
                );
                downloadDivAsImage(
                    dlGoRoutinesMetricsBtn,
                    'goroutines-chart',
                    'goroutines-metrics'
                );
                downloadDivAsImage(
                    dlLoadMemoryMetricsBtn,
                    'load-memory-chart',
                    'load-memory-metrics'
                );
                downloadDivAsImage(
                    dlHealthMetricsBtn,
                    'health-chart',
                    'health-metrics'
                );
            })
            .catch((error) => {
                console.error('Error fetching metrics:', error);
            });
    }
    function downloadDivAsImage(downloadBtn, divID, fileName) {
        const targetDiv = document.getElementById(divID);
        downloadBtn.style.display = 'block';
        downloadBtn.addEventListener('click', function () {
            html2canvas(targetDiv).then(function (canvas) {
                const link = document.createElement('a');
                link.href = canvas.toDataURL('image/png');
                link.download = fileName;
                link.click();
            });
        });
    }
    // function chartsDownload() {
    //   downloadChartImgBtn.style.display = 'block'
    //   downloadChartImgBtn.addEventListener('click', function () {
    //     html2canvas(document.querySelector('.content-page')).then(function (
    //       canvas
    //     ) {
    //       let link = document.createElement('a')
    //       link.href = canvas.toDataURL('image/png')
    //       link.download = 'content-page.png'
    //       link.click()
    //     })
    //   })
    // }

    // KB
    function renderCharts(data) {
        const charts = {
            loadChart: echarts.init(elements.loadChart),
            cpuChart: echarts.init(elements.cpuChart),
            memoryPieChart: echarts.init(elements.memoryPieChart),
            heapUsageChart: echarts.init(elements.heapUsageChart)
        };

        Object.values(charts).forEach((chart) =>
            chart.setOption({
                title: {
                    text: 'Loading...'
                },
                tooltip: {}
            })
        );

        const { load_statistics, cpu_statistics, memory_statistics } = data;

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
                formatter: (params) => {
                    return params
                        .map((param) => {
                            const { axisValueLabel, value } = param;
                            let info = '';
                            if (value > 90) info = '[Critical Load]';
                            else if (value > 80) info = '[High Load]';
                            else if (value > 50) info = '[Moderate Load]';
                            else if (value <= 30) info = '[Healthy]';
                            return `${axisValueLabel}: ${value} %<br/><span>${info}</span>`;
                        })
                        .join('<br/>');
                }
            },
            xAxis: {
                type: 'category',
                data: [
                    'Service CPU Load',
                    'System CPU Load',
                    'Total CPU Load',
                    'Service Memory Load',
                    'System Memory Load'
                ]
            },
            yAxis: {
                type: 'value',
                max: 100
            },
            series: [
                {
                    data: [
                        parseFloat(load_statistics.service_cpu_load),
                        parseFloat(load_statistics.system_cpu_load),
                        parseFloat(load_statistics.total_cpu_load),
                        parseFloat(load_statistics.service_memory_load),
                        parseFloat(load_statistics.system_memory_load)
                    ],
                    type: 'bar',
                    itemStyle: {
                        color: (params) => {
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
                }
            ]
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
                data: [
                    {
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
            series: [
                {
                    name: 'CPU Usage',
                    type: 'pie',
                    radius: '55%',
                    center: ['50%', '60%'],
                    data: [
                        {
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
                }
            ]
        });

        // Memory Pie Chart
        charts.memoryPieChart.setOption({
            title: {
                text: 'Memory Distribution'
            },
            tooltip: {
                trigger: 'item',
                formatter: function (params) {
                    return `${params.seriesName}<br/>${params.name}: ${params.value} (${params.percent}%) []`;
                }
            },
            legend: {
                orient: 'horizontal',
                center: 0,
                padding: [40, 0, 0, 0],
                data: [
                    'Memory Used by Service',
                    'Memory Used by System',
                    'Available Memory'
                ]
            },
            series: [
                {
                    name: 'Memory Usage',
                    type: 'pie',
                    radius: '55%',
                    center: ['50%', '60%'],
                    data: [
                        {
                            value: parseFloat(
                                data.memory_statistics.memory_used_by_service
                            ),
                            name: 'Memory Used by Service'
                        },
                        {
                            value: parseFloat(
                                data.memory_statistics.memory_used_by_system
                            ),
                            name: 'Memory Used by System'
                        },
                        {
                            value: parseFloat(
                                data.memory_statistics.available_memory
                            ),
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
                }
            ]
        });

        let values = [];
        data.memory_statistics.mem_stats_records.forEach((record) => {
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
                data: [
                    'HeapAlloc',
                    'HeapSys',
                    'HeapIdle',
                    'HeapInuse',
                    'HeapReleased'
                ]
            },
            yAxis: {
                type: 'value',
                name: 'MB'
            },
            series: [
                {
                    name: 'Memory (MB)',
                    type: 'bar',
                    data: [
                        values[0],
                        values[1],
                        values[2],
                        values[3],
                        values[4]
                    ]
                }
            ]
        });
    }

    const chart = echarts.init(document.getElementById('chart'));

    // Function to get the local ISO string with timezone offset
    function toLocalISOString(date) {
        const tzOffset = -date.getTimezoneOffset(); // in minutes
        const diff = tzOffset >= 0 ? '+' : '-';
        const pad = (num) => `${Math.floor(Math.abs(num))}`.padStart(2, '0');

        const offsetHours = pad(tzOffset / 60);
        const offsetMinutes = pad(tzOffset % 60);

        return (
            date.getFullYear() +
            '-' +
            pad(date.getMonth() + 1) +
            '-' +
            pad(date.getDate()) +
            'T' +
            pad(date.getHours()) +
            ':' +
            pad(date.getMinutes()) +
            ':' +
            pad(date.getSeconds()) +
            '.' +
            String((date.getMilliseconds() / 1000).toFixed(3)).slice(2, 5) +
            diff +
            offsetHours +
            ':' +
            offsetMinutes
        );
    }

    function fetchDataPointsFromServer(metricName, timeRange) {
        let StartTime = new Date();
        let EndTime = new Date();

        if (timeRange == '5m') {
            StartTime = new Date(new Date().getTime() - 5 * 60000); // Subtract 5 minutes
        } else if (timeRange == '15m') {
            StartTime = new Date(new Date().getTime() - 15 * 60000); // Subtract 15 minutes
        } else if (timeRange == '30m') {
            StartTime = new Date(new Date().getTime() - 30 * 60000); // Subtract 30 minutes
        } else if (timeRange == '1h') {
            StartTime = new Date(new Date().getTime() - 60 * 60000); // Subtract 1 hour
        } else if (timeRange == '6h') {
            StartTime = new Date(new Date().getTime() - 360 * 60000); // Subtract 6 hours
        } else if (timeRange == '1d') {
            StartTime = new Date(new Date().getTime() - 1440 * 60000); // Subtract 1 day
        } else if (timeRange == '3d') {
            StartTime = new Date(new Date().getTime() - 4320 * 60000); // Subtract 3 days
        } else if (timeRange == '7d') {
            StartTime = new Date(new Date().getTime() - 10080 * 60000); // Subtract 7 days
        }

        // else if (timeRange == "7d") {
        //     StartTime = new Date(new Date().getTime() - 10080 * 60000); // Subtract 7 days
        // } else if (timeRange == "1month") {
        //     StartTime = new Date(new Date().getTime() - 43200 * 60000); // Subtract 1 month
        // }

        let metricList = [];
        if (metricName == 'heap') {
            metricList = [
                'heap_alloc',
                'heap_sys',
                'heap_inuse',
                'heap_idle',
                'heap_released'
            ];
        } else if (metricName == 'stack') {
            metricList = ['stack_inuse', 'stack_sys'];
        } else if (metricName == 'gc') {
            metricList = ['pause_total_ns', 'num_gc', 'gc_cpu_fraction'];
        } else if (metricName == 'misc') {
            metricList = [
                'm_span_inuse',
                'm_span_sys',
                'm_cache_inuse',
                'm_cache_sys',
                'buck_hash_sys',
                'gc_sys',
                'other_sys'
            ];
        }

        let data = {
            field_name: metricList,
            timerange: timeRange,
            start_time: toLocalISOString(StartTime),
            end_time: toLocalISOString(EndTime)
        };

        fetch(`/monigo/api/v1/service-metrics`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        })
            .then((response) => response.json())
            .then((data) => {
                let rawData = [];
                for (let i = 0; i < data.length; i++) {
                    const timestamp = new Date(data[i].time);
                    rawData.push({
                        time: timestamp,
                        value: data[i].value
                    });
                }

                const seriesData = {};
                rawData.forEach((dataPoint) => {
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
                        data: rawData.map((d) => d.time.toLocaleString()),
                        axisLabel: {
                            formatter: function (value) {
                                return value.split(' ')[1]; // Show only time
                            }
                        }
                    },
                    yAxis: {
                        type: 'value',
                        axisLabel: {
                            formatter: function (value) {
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
    document
        .getElementById('metric-select')
        .addEventListener('change', updateChart);
    document
        .getElementById('time-select')
        .addEventListener('change', updateChart);
    updateChart();

    // Responsive behavior
    window.addEventListener('resize', function () {
        chart.resize();
    });

    function getStatus(percentage) {
        let fillColor;
        let tag;
        if (percentage >= 80) {
            fillColor = 'var(--green)';
            tag = 'Smooth Sailing 🛳️';
        } else if (percentage >= 60) {
            fillColor = 'var(--lightgreen)';
            tag = 'System Looking Good 👍';
        } else if (percentage >= 50) {
            fillColor = 'var(--yellow)';
            tag = 'Fairly Balanced 🌟';
        } else if (percentage >= 40) {
            fillColor = 'var(--orange)';
            tag = 'System Under Stress 😟';
        } else {
            fillColor = 'var(--red)';
            tag = 'Critical Condition 🚨';
        }
        return [fillColor, tag]; // Return an array
    }

    function updateGauge(gaugeId, health) {
        const srevPercentage = health.service_health.percent;
        const sysPercentage = health.system_health.percent;
        const iconSysMsg = health.system_health.icon_msg;
        const iconServMsg = health.service_health.icon_msg;
        const gauge = document.getElementById(gaugeId);
        const gaugeText = gauge.querySelector('text'); // Correct typo here
        gaugeText.textContent = `${srevPercentage}%`;

        if (elements.healthMessage.textContent === '') {
            elements.healthMessage.textContent = health.service_health.message;
        }

        let fillColorServ, tagServ;
        [fillColorServ, tagServ] = getStatus(srevPercentage);

        let fillColorSys, tagSys;
        [fillColorSys, tagSys] = getStatus(sysPercentage);

        if (srevPercentage == 0) {
            elements.serviceHealthTag.innerHTML = `
            <h6 class="mb-0 mt-1">Service: ${tagServ} <br> Not in a Good State, Kindly check! <span class="info-icon" data-tooltip="${iconServMsg}">i</span></h6>
        `;
        } else {
            elements.serviceHealthTag.innerHTML = `
            <h6 class="mb-0">Service: ${tagServ} : ${srevPercentage}% <span class="info-icon" data-tooltip="${iconServMsg}">i</span></h6>
        `;
        }

        if (sysPercentage == 0) {
            elements.systemHealthTag.innerHTML = `
            <h6 class="mb-0">System: ${tagSys} <br> Not in a Good State, Kindly check! <span class="info-icon" data-tooltip="${iconSysMsg}">i</span></h6>
        `;
        } else {
            elements.systemHealthTag.innerHTML = `
            <h6 class="mb-0">System: ${tagSys} : ${sysPercentage}% <span class="info-icon" data-tooltip="${iconSysMsg}">i</span></h6>
        `;
        }

        // Reset the --o property to 0 to restart the animation
        gauge.style.setProperty('--o', 0);

        // Trigger a reflow to reset the animation (forces a repaint)
        void gauge.offsetWidth;

        // Set the custom properties for the gauge
        gauge.style.setProperty('--fill-percentage', srevPercentage); // Use percentage for fill
        gauge.style.setProperty('--fill-color', fillColorServ); // Use color for fill

        // Now update --o to the target percentage to animate
        gauge.style.setProperty('--o', srevPercentage);
    }
});

function getDlBtn(id) {
    btn = document.getElementById(id);
    btn.style.display = 'none';
    return btn;
}
