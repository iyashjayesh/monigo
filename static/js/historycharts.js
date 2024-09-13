document.addEventListener('DOMContentLoaded', () => {

     const refreshHtml = `
        <div class="loader-container mt-3">
            <div class="bouncing-dots">
                <div class="dot"></div>
                <div class="dot"></div>
                <div class="dot"></div>
            </div>
        </div>`;

     const elements = {
        cpuUsageChart: document.getElementById('cpu-usage-chart'),
        goroutinesChart: document.getElementById('goroutines-chart'),
        loadMemoryChart: document.getElementById('load-memory-chart'),
        healthChart: document.getElementById('health-chart')
    };

    Object.values(elements).forEach(el => el && (el.innerHTML = refreshHtml));

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
        } 
        
        // else if (timeRange == "7d") {
        //     StartTime = new Date(new Date().getTime() - 10080 * 60000); // Subtract 7 days
        // } else if (timeRange == "1month") {
        //     StartTime = new Date(new Date().getTime() - 43200 * 60000); // Subtract 1 month
        // }

        let metricList = [];
        if (metricName == "cpu-usage") {
            metricList = ["total_cores", "cores_used_by_service", "cores_used_by_system"];
        } else if (metricName == "goroutines") {
            metricList = ["goroutines"];
        } else if (metricName == "load-memory") {
            metricList = ["overall_load_of_service", "service_cpu_load", "service_memory_load", "system_cpu_load", "system_memory_load"];
        } else if (metricName == "health") {
            metricList = ["service_health_percent", "system_health_percent", "overall_health_percent"];
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
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(data),
            }).then(response => response.json())
            .then(data => {
                // console.log('Fetching data for metric:', metricName, 'and time range:', timeRange);
                // console.log('API REQ:', data);

                let rawData = [];
                for (let i = 0; i < data.length; i++) {
                    const timestamp = new Date(data[i].time);
                    rawData.push({
                        time: timestamp,
                        value: data[i].value
                    });
                }

                if (metricName == "health") {
                    renderHealthChart(rawData);
                } 

                if (metricName == "cpu-usage") {
                    renderCpuUsageChart(rawData);
                }

                if (metricName == "goroutines") {
                    renderGoroutinesChart(rawData);
                }

                if (metricName == "load-memory") {
                    renderLoadMemoryChart(rawData);
                }
            })
            .catch((error) => {
                console.error('Error:', error);
            });
    }

    function renderHealthChart(data) {
        const healthChart = echarts.init(elements.healthChart);
        const time = data.map(entry => entry.time);
        const overallHealthPercent = data.map(entry => entry.value.overall_health_percent);
        const serviceHealthPercent = data.map(entry => entry.value.service_health_percent);
        const systemHealthPercent = data.map(entry => entry.value.system_health_percent);

        const option = {
            title: {
                text: 'Health Metrics',
                left: 'center'
            },
            tooltip: {
                trigger: 'axis'
            },
            legend: {
                data: ['Overall Health', 'Service Health', 'System Health'],
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
            series: [
                {
                    name: 'Overall Health',
                    type: 'line',
                    data: overallHealthPercent
                },
                {
                    name: 'Service Health',
                    type: 'line',
                    data: serviceHealthPercent
                },
                {
                    name: 'System Health',
                    type: 'line',
                    data: systemHealthPercent
                }
            ]
        };

        healthChart.setOption(option);
    }

    function renderCpuUsageChart(data) {
        const cpuUsageChart = echarts.init(elements.cpuUsageChart);
        const time = data.map(entry => entry.time);
        const totalCores = data.map(entry => entry.value.total_cores);
        const coresUsedByService = data.map(entry => entry.value.cores_used_by_service);
        const coresUsedBySystem = data.map(entry => entry.value.cores_used_by_system);

        const option = {
            title: {
                text: 'CPU Usage Metrics',
                left: 'center'
            },
            tooltip: {
                trigger: 'axis'
            },
            legend: {
                data: ['Total Cores', 'Cores Used by Service', 'Cores Used by System'],
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
            series: [
                {
                    name: 'Total Cores',
                    type: 'line',
                    data: totalCores
                },
                {
                    name: 'Cores Used by Service',
                    type: 'line',
                    data: coresUsedByService
                },
                {
                    name: 'Cores Used by System',
                    type: 'line',
                    data: coresUsedBySystem
                }
            ]
        };

        cpuUsageChart.setOption(option);
    }

    function renderGoroutinesChart(data) {
        const goroutinesChart = echarts.init(elements.goroutinesChart);
        const time = data.map(entry => entry.time);
        const goroutines = data.map(entry => entry.value.goroutines);

        const option = {
            title: {
                text: 'Goroutines Metrics',
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
            series: [
                {
                    name: 'Goroutines',
                    type: 'line',
                    data: goroutines
                }
            ]
        };

        goroutinesChart.setOption(option);
    }

    function renderLoadMemoryChart(data) {
        const loadMemoryChart = echarts.init(elements.loadMemoryChart);
        const time = data.map(entry => entry.time);
        const overallLoadOfService = data.map(entry => entry.value.overall_load_of_service);
        const serviceCpuLoad = data.map(entry => entry.value.service_cpu_load);
        const serviceMemoryLoad = data.map(entry => entry.value.service_memory_load);
        const systemCpuLoad = data.map(entry => entry.value.system_cpu_load);
        const systemMemoryLoad = data.map(entry => entry.value.system_memory_load);

        const option = {
            title: {
                text: 'Load Memory Metrics',
                left: 'center',
                padding: [0, 0, 10, 0]
            },
            tooltip: {
                trigger: 'axis'
            },
            legend: {
                data: ['Overall Load of Service', 'Service CPU Load', 'Service Memory Load', 'System CPU Load', 'System Memory Load'],
                top: 10
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
            series: [
                {
                    name: 'Overall Load of Service',
                    type: 'line',
                    data: overallLoadOfService
                },
                {
                    name: 'Service CPU Load',
                    type: 'line',
                    data: serviceCpuLoad
                },
                {
                    name: 'Service Memory Load',
                    type: 'line',
                    data: serviceMemoryLoad
                },
                {
                    name: 'System CPU Load',
                    type: 'line',
                    data: systemCpuLoad
                },
                {
                    name: 'System Memory Load',
                    type: 'line',
                    data: systemMemoryLoad
                }
            ]
        };

        loadMemoryChart.setOption(option);
    } 

    function updateHistoryChart(metricName) {
        const metricSelect = metricName;
        const timeSelect = document.getElementById(`${metricName}-time-select`).value;
        fetchDataPointsFromServer(metricSelect, timeSelect);
    }

    document.getElementById('cpu-usage-time-select').addEventListener('change', () => updateHistoryChart("cpu-usage"));
    document.getElementById('goroutines-time-select').addEventListener('change', () => updateHistoryChart("goroutines"));
    document.getElementById('load-memory-time-select').addEventListener('change', () => updateHistoryChart("load-memory"));
    document.getElementById('health-time-select').addEventListener('change', () => updateHistoryChart("health"));

    updateHistoryChart("cpu-usage");
    updateHistoryChart("goroutines");
    updateHistoryChart("load-memory");
    updateHistoryChart("health");
});