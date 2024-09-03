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
    } else if (DASHBOARD) {
        fetchMetrics();
        fetchServiceInfo();
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
                console.log("Dashboard metrics: ", data);

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
});