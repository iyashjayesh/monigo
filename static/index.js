document.addEventListener('DOMContentLoaded', function() {

    const DASHBOARD = document.getElementById('dashboard');
    const serviceInfoContainer = document.getElementById('service-container');
    const refreshHtml = `<div class="loader-container">
                            <div class="bouncing-dots">
                                <div class="dot"></div>
                                <div class="dot"></div>
                                <div class="dot"></div>
                            </div>
                        </div>`;
    

            //             const loadChart = echarts.init(document.getElementById('load-chart'));
            // const cpuChart = echarts.init(document.getElementById('cpu-chart'));
            // const memoryPieChart = echarts.init(document.getElementById('memory-pie-chart'));
            // const heapUsageChart = echarts.init(document.getElementById('heap-memory-chart'));
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

    for (let key in elements) {
        if (elements[key]) {
            elements[key].innerHTML = refreshHtml;
        }
    }

    if (DASHBOARD) {
        console.log('Dashboard found');
        fetchMetrics();
        fetchServiceInfo();
    } else {
        console.log('No dashboard found');
    }
    

        function animateProgressBar(bar, targetWidth, duration) {
        let start = null;
        const initialWidth = 0;

        function step(timestamp) {
            if (!start) start = timestamp;
            const progress = timestamp - start;
            const width = Math.min(initialWidth + (progress / duration) * targetWidth, targetWidth);
            bar.style.width = width + '%';

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
                const options = {
                    year: 'numeric',
                    month: 'long',
                    day: 'numeric'
                };
                const formattedDate = date.toLocaleDateString('en-US', options);
                const timeOptions = {
                    hour: 'numeric',
                    minute: 'numeric',
                    hour12: true
                };
                const formattedTime = date.toLocaleTimeString('en-US', timeOptions);

                serviceInfoContainer.innerHTML = `
                    <div class="row pl-3 pr-3">
                            <div class="card card-block card-stretch card-height">
                                <div class="card-body">
                                    <div class="d-flex align-items-center mb-4 card-total-sale">
                                        <div class="icon iq-icon-box-2 bg-info-light">
                                            <img src="../assets/images/product/1.png" class="img-fluid" alt="image">
                                        </div>
                                        <div>
                                            <p class="mb-2">Service Name: 
                                                <h4>${data.service_name}</h4> 
                                            </p>
                                            <p class="mb-2">Go Version: 
                                                <h4>${data.go_version}</h4>
                                            </p>
                                            <p class="mb-2">Service Start Time:
                                                <h4>${formattedDate}<br/> ${formattedTime}</h4>
                                            </p>
                                            <p class="mb-2">Process ID:
                                                <h4>${data.process_id}</h4>
                                            </p>
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

   

    function updateElement(element, label, value, info = '') {
        if (element) {
            element.innerHTML = `<div>
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
                updateElement(elements.goroutines, 'Go Routines:', data?.core_statistics?.goroutines ?? 'N/A', 'Number of goroutines that are currently running');
                updateElement(elements.serviceLoad, 'Load:', `${data?.load_statistics?.overall_load_of_service ?? 'N/A'}`, 'The load average of the system');
                updateElement(elements.cores, 'Cores:', `${data?.cpu_statistics?.cores_used_by_service ?? 'N/A'} / ${data?.cpu_statistics?.total_cores ?? 'N/A'}`, 'Number of CPU cores');
                updateElement(elements.memory, 'Memory:', `${data?.memory_statistics?.memory_used_by_service ?? 'N/A'}`, 'Memory used by the service');
                updateElement(elements.cpuUsage, 'CPU Usage:', `${data?.cpu_statistics?.cores_used_by_service_in_percent ?? 'N/A'}`, 'CPU usage of the service');
                updateElement(elements.uptime, 'Uptime:', data?.core_statistics?.uptime ?? 'N/A', 'Uptime of the service');

                const healthIndicator = document.getElementById('health-indicator');
                if (data.overall_health.health.healthy) {
                    healthIndicator.classList.add('healthy');
                    document.getElementById('health-message').textContent = data.overall_health.health.message;
                } else {
                    healthIndicator.classList.add('unhealthy');
                    document.getElementById('health-message').textContent = data.overall_health.health.message;
                }

                renderCharts(data);
            })
            .catch(error => {
                console.error('Error fetching metrics:', error);
            });
    }

    function renderCharts(data) {
            const loadChart = echarts.init(document.getElementById('load-chart'));
            const cpuChart = echarts.init(document.getElementById('cpu-chart'));
            const memoryPieChart = echarts.init(document.getElementById('memory-pie-chart'));
            const heapUsageChart = echarts.init(document.getElementById('heap-memory-chart'));

            loadChart.innerHTML = refreshHtml;
            cpuChart.innerHTML = refreshHtml;
            memoryPieChart.innerHTML = refreshHtml;
            heapUsageChart.innerHTML = refreshHtml;

            // Load Chart
            loadChart.setOption({
                title: { text: 'Load Statistics' },
                tooltip: {
                    trigger: 'axis',
                    axisPointer: { type: 'shadow' },
                    formatter: function (params) {
                        let tooltipContent = '';
                        let info = '';
                        params.forEach(param => {
                            if (param.value > 90) {
                                info = '[Critical Load]'; // Custom info for values above 90
                            } else if (param.value > 80) {
                                info = '[High Load]'; // Custom info for values between 80 and 90
                            } else if (param.value > 50) {
                                info = '[Moderate Load]'; // Custom info for values between 50 and 80
                            } else if (param.value <= 30) {
                                info = '[Healthy]'; // Custom info for values below or equal to 30
                            } 
                            tooltipContent += `${param.axisValueLabel}: ${param.value} %<br/><span>${info}</span>`;
                        });
                        return tooltipContent;
                    }
                },
                xAxis: {
                    type: 'category',
                    data: ['Service CPU Load', 'System CPU Load', 'Total CPU Load', 'Service Memory Load', 'System Memory Load']
                },
                yAxis: {
                    type: 'value',
                    max: 100 // Set max value to 100
                },
                series: [{
                    data: [
                        parseFloat(data.load_statistics.service_cpu_load),
                        parseFloat(data.load_statistics.system_cpu_load),
                        parseFloat(data.load_statistics.total_cpu_load),
                        parseFloat(data.load_statistics.service_memory_load),
                        parseFloat(data.load_statistics.system_memory_load)
                    ],
                    type: 'bar',
                    itemStyle: {
                        color: function (params) {
                            const value = params.value;
                            if (value > 90) {
                                return 'red'; // Color for values above 90
                            } else if (value > 80) {
                                return 'orange'; // Color for values between 80 and 90
                            } else if (value > 50) {
                                return 'yellow'; // Color for values between 50 and 80
                            } else if (value <= 30) {
                                return 'green'; // Color for values below or equal to 30
                            }
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
            cpuChart.setOption({
                title: { text: 'CPU Statistics' },
                tooltip: {
                    trigger: 'item',
                    formatter: '{a} <br/>{b} : {c} ({d}%)'
                },
                legend: {
                    orient: 'horizontal',
                    center: 0,
                    padding: [30, 0, 0, 0],
                    data: [
                        { name: 'Cores Used by Service', icon: 'rect', itemStyle: { color: '#00A1E4' } },
                        { name: 'Cores Used by System', icon: 'rect', itemStyle: { color: '#FF6F61' } },
                        { name: 'Total Cores', icon: 'rect', itemStyle: { color: '#FFD166' } }
                    ]
                },
                series: [{
                    name: 'CPU Usage',
                    type: 'pie',
                    radius: '50%',
                    center: ['50%', '50%'],
                    data: [
                        { value: data.cpu_statistics.cores_used_by_service, name: 'Cores Used by Service' },
                        { value: data.cpu_statistics.cores_used_by_system, name: 'Cores Used by System' },
                        { value: parseFloat(data.cpu_statistics.total_cores), name: 'Total Cores' }
                    ],
                    itemStyle: {
                        emphasis: {
                            shadowBlur: 10,
                            shadowOffsetX: 0,
                            shadowColor: 'rgba(0, 0, 0, 0.5)'
                        }
                    }
                }]
            });

            memoryPieChart.setOption({
                title: {
                    text: 'Memory Distribution',
                },
                tooltip: {
                    trigger: 'item',
                    formatter: function (params) {
                        return `${params.seriesName}<br/>${params.name}: ${params.value} (${params.percent}%) []`;
                    }
                },
                legend: {
                    orient: 'vertical',
                    left: 'left',
                    padding: [30, 0, 0, 0],
                    data: ['Memory Used by Service', 'Memory Used by System', 'Available Memory']
                },
                series: [
                    {
                        name: 'Memory Usage',
                        type: 'pie',
                        radius: '55%',
                        center: ['50%', '60%'],
                        data: [
                            { value: parseFloat(data.memory_statistics.memory_used_by_service), name: 'Memory Used by Service' },
                            { value: parseFloat(data.memory_statistics.memory_used_by_system), name: 'Memory Used by System' },
                            { value: parseFloat(data.memory_statistics.available_memory), name: 'Available Memory' }
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

            heapUsageChart.setOption({
                title: { text: 'Heap Memory Usage' },
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
