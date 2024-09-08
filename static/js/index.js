document.addEventListener('DOMContentLoaded', () => {

    const DASHBOARD = document.getElementById('dashboard');
    const GOROUTINES_PAGE = document.getElementById('goroutines-page');

    const process_id = document.getElementById('process_id');
    const service_name = document.getElementById('service_name');
    const go_version = document.getElementById('go_version');
    const service_start_time = document.getElementById('service_start_time');

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
        healthTag: document.getElementById('health-tag')
    };

    Object.values(elements).forEach(el => el && (el.innerHTML = refreshHtml));

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
            .then(response => response.json())
            .then(data => {
                // serviceInfoContainer.innerHTML = '';
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

                service_name.innerHTML =`
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
                    <div class="d-flex align-items-center mb-4 card-total-sale">
                        
                        <div>
                            <p class="mb-2">Service Start Time: <h4>${formattedDate}<br/> ${formattedTime}</h4></p>
                        </div>
                    </div>`;
                process_id.innerHTML = `
                    <div class="d-flex align-items-center mb-4 card-total-sale">
                        
                        <div>
                            <p class="mb-2">Process ID: <h4>${data.process_id}</h4></p>
                        </div>
                    </div>`;
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
        fetch(`/monigo/api/v1/metrics`)
            .then(response => response.json())
            .then(data => {

                console.log('Metrics:', data);  
                const {
                    core_statistics,
                    load_statistics,
                    cpu_statistics,
                    memory_statistics,
                    overall_health
                } = data;



                updateGauge('g1', overall_health.overall_health_percent);


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
                padding: [40, 0, 0, 0],
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
                orient: 'horizontal',
                center: 0,
                padding: [40, 0, 0, 0],
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

    function fetchDataPointsFromServer(metricName, timeRange) {
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
        } else if (timeRange == "1month") {
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

        fetch(`/monigo/api/v1/service-metrics`, {
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


    function updateGauge(gaugeId, percentageStr) {

        percentageStr = percentageStr.replace('%', '');
        const percentage = parseFloat(percentageStr);

        const gauge = document.getElementById(gaugeId);
        const guageText = gauge.querySelector('text');
        // health-tag
        // const healthTag = document.getElementById('health-tag');

        // Update the text inside the gauge
        guageText.textContent = `${percentage}%`;

        // Determine the fill color based on the percentage
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
                                        // <!-- <p class="mb-0">Service health is at 70%</p> -->

        elements.healthTag.innerHTML = `
            <p class="mb-0">${tag}</p>
        `
        

        // Reset the --o property to 0 to restart the animation
        gauge.style.setProperty('--o', 0);

        // Trigger a reflow to reset the animation (forces a repaint)
        void gauge.offsetWidth;

        // Set the custom properties for the gauge
        gauge.style.setProperty('--fill-percentage', percentage);
        gauge.style.setProperty('--fill-color', fillColor);

        // Now update --o to the target percentage to animate
        gauge.style.setProperty('--o', percentage);
    }

});