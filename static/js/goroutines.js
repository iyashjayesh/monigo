document.addEventListener('DOMContentLoaded', () => {
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


        fetch(`/monigo/api/v1/service-metrics`, {
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
        fetch(`/monigo/api/v1/go-routines-stats`)
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