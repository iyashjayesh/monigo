document.addEventListener('DOMContentLoaded', function() {
    
    const serviceInfoContainer = document.getElementById('service-container');
    const runtimeMetricsContainer = document.getElementById('runtime-metrics-container');
    const goRoutinesNumber = document.getElementById('goroutine-count');
    const memValue = document.getElementById('mem-value');
    const cpuValue = document.getElementById('cpu-value');
    const serviceHealth = document.getElementById('load-service-health-guage');
    

    const memoryStatsContainer = document.getElementById('memory-stats');
    const cpuStatsContainer = document.getElementById('cpu-stats');


    if (serviceInfoContainer) {
        fetchServiceInfo();
    } else {
        console.error('Element with ID "service-container" not found.');
    }

    if (runtimeMetricsContainer) {
        fetchMetrics("MB");
    } else {
        console.error('Element with ID "runtime-metrics-container" not found.');
    }

    if (goRoutinesNumber) {
        fetchGoRoutines();
    } else {
        console.error('Element with ID "go-routines" not found.');
    }

    if (memoryStatsContainer){
        console.log('Memory and CPU stats found');
        updateStats();
    } else {
        console.error('Element with ID "memory-stats" or "cpu-stats" not found.');
    }
    
    // if (serviceHealth) {
    //     fetchServiceHealth();
    // } else {
    //     console.error('Element with ID "load-service-health" not found.');
    // }

    

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

                const serviceName = data.service_name;
                const goVersion = data.go_version;
                const serviceStartTime = data.service_start_time;

                const date = new Date(serviceStartTime);
                const options = { year: 'numeric', month: 'long', day: 'numeric' };
                const formattedDate = date.toLocaleDateString('en-US', options);
                const timeOptions = { hour: 'numeric', minute: 'numeric', hour12: true };
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
                                            <p class="mb-2">Service Name:</p>
                                            <h4>${serviceName}</h4>

                                            <p class="mb-2">Go Version:</p>
                                            <h4>${goVersion}</h4>
                                            <p class="mb-2">Service Start Time:</p>
                                            <h4>${formattedDate}<br/> ${formattedTime}</h4>
                                        </div>
                                    </div>
                                    <div class="iq-progress-bar mt-2">
                                        <span class="bg-info iq-progress progress-1" data-percent="85"></span>
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

    function fetchMetrics(unit) {
        runtimeMetricsContainer.innerHTML = `<div class="service-info">Fetching the data...</div>`;
        
        fetch(`/metrics?unit=${unit}`)
            .then(response => response.json())  // Assume the server returns JSON directly
            .then(metrics => {
                let cpuLoopDetails = '';

                for (let [key, value] of Object.entries(metrics.cpu)) {
                    let name = '';
                    if (key === "total_cores") {
                        name = "Total Cores";   
                    } else if (key === "total_logical_cores") {
                        name = "Total Logical Cores";
                    } else if (key === "system_used_cores") {
                        name = "System Used Cores";
                    } else if (key === "process_used_cores") {
                        name = "Process Used Cores";
                    } else if (key === "cores") {
                        name = "Cores";
                    } else if (key === "used_in_percent") {
                        name = "Used in Percent";
                    }

                    cpuLoopDetails += `
                        <div>
                            <p class="mb-2">${name} : </p>
                            <h4>${value}</h4>
                        </div>
                    `;
                }


                cpuValue.innerHTML = `
                            <div class="card card-block card-stretch card-height">
                                <div class="card-body">
                                    <div class="d-flex align-items-center justify-content-between mb-4 card-total-sale">
                                        <div class="d-flex align-items-center">
                                            <div class="icon iq-icon-box-2 bg-info-light">
                                                <img src="../assets/images/product/1.png" class="img-fluid" alt="CPU Usage IMG">
                                            </div>
                                            <div class="ml-3">
                                                <p class="mb-2">CPU usage by service:</p>
                                                <h4>${metrics.cpu.used_in_percent}</h4>
                                            </div>
                                        </div>
                                        <div>
                                            <button type="button" class="btn btn-primary mt-2" data-toggle="modal" data-target="#exampleModalCenteredScrollable">
                                                View Details<i class="fa fa-external-link pl-2" aria-hidden="true"></i>
                                            </button>
                                            <div id="exampleModalCenteredScrollable" class="modal fade" tabindex="-1" role="dialog"
                                                aria-labelledby="exampleModalCenteredScrollableTitle" aria-hidden="true">
                                                <div class="modal-dialog modal-dialog-scrollable modal-dialog-centered" role="document">
                                                    <div class="modal-content">
                                                        <div class="modal-header">
                                                            <h5 class="modal-title" id="exampleModalCenteredScrollableTitle">Modal title</h5>
                                                            <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                                                                <span aria-hidden="true">×</span>
                                                            </button>
                                                        </div>
                                                        <div class="modal-body">
                                                            ${cpuLoopDetails}
                                                        </div>
                                                        <div class="modal-footer">
                                                            <button type="button" class="btn btn-secondary" data-dismiss="modal">Close</button>
                                                            <button type="button" class="btn btn-primary">Save changes</button>
                                                        </div>
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                    <div class="iq-progress-bar mt-2">
                                        <span class="bg-info iq-progress progress-1" style="width: 0%;" data-percent="65"></span>
                                    </div>
                                </div>
                            </div>`;

                let memLoopDetails = '';

                for (let [key, value] of Object.entries(metrics.memory)) {
                    if (key === "mem_stats_records") {
                        console.log('Value', value);
                        console.log('Value Length', value.records.length);

                        value.records.forEach(item => {
                            memLoopDetails += `
                                <div>
                                    <p class="mb-2">${item.record_name} : </p>
                                    <h4>${item.record_value} ${item.record_unit}</h4>
                                </div>
                            `;
                        });
                    }
                }
               

                memValue.innerHTML = `
                            <div class="card card-block card-stretch card-height">
                                <div class="card-body">
                                    <div class="d-flex align-items-center justify-content-between mb-4 card-total-sale">
                                        <div class="d-flex align-items-center">
                                            <div class="icon iq-icon-box-2 bg-info-light">
                                                <img src="../assets/images/product/1.png" class="img-fluid" alt="CPU Usage IMG">
                                            </div>
                                            <div class="ml-3">
                                                <p class="mb-2">Memory usage by service:</p>
                                                <h4>${metrics.memory.used_in_percent}</h4>
                                            </div>
                                        </div>
                                        <div>
                                            <!-- Button trigger modal -->
                                            <button type="button" class="btn btn-primary mt-2" data-toggle="modal" data-target="#mem-statstics">
                                                View Details<i class="fa fa-external-link pl-2" aria-hidden="true"></i>
                                            </button>
                                            <!-- Modal -->
                                            <div id="mem-statstics" class="modal fade" tabindex="-1" role="dialog"
                                                aria-labelledby="mem-statsticsTitle" aria-hidden="true">
                                                <div class="modal-dialog modal-dialog-scrollable modal-dialog-centered" role="document">
                                                    <div class="modal-content">
                                                        <div class="modal-header">
                                                            <h5 class="modal-title" id="mem-statsticsTitle">Modal title</h5>
                                                            <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                                                                <span aria-hidden="true">×</span>
                                                            </button>
                                                        </div>
                                                        <div class="modal-body">
                                                            ${memLoopDetails}
                                                        </div>
                                                        <div class="modal-footer">
                                                            <button type="button" class="btn btn-secondary" data-dismiss="modal">Close</button>
                                                            <button type="button" class="btn btn-primary"><i class="fa fa-download" aria-hidden="true"></i>Download Excel</a></button>
                                                        </div>
                                                    </div>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                    <div class="iq-progress-bar mt-2">
                                        <span class="bg-info iq-progress progress-1" style="width: 0%;" data-percent="65"></span>
                                    </div>
                                </div>
                            </div>`;


                // alloc : "5.18" cores : "0.01PC / 4.00SC / 10LC / 10C" goroutines : 9 heap_alloc : "5.18" heap_sys : "11.41" load : "0.15%" memory_usage : "MB" memory_used : "0.14%" requests : 0 sys : "18.16" total_alloc : "6.62" total_duration : "0s" uptime : "12.83 s"

                const listCont = [
                    {
                        name: "Load",
                        image: "../assets/images/product/1.png",
                        value: metrics.load,
                        dataPercent: 85
                    },
                    {
                        name: "Cores",
                        image: "../assets/images/product/2.png",
                        value: metrics.cores,
                        dataPercent: 70
                    },
                    {
                        name: "Memory Used",
                        image: "../assets/images/product/3.png",
                        value: metrics.memory_used,
                        dataPercent: 75
                    },
                    {
                        name: "Requests",
                        image: "../assets/images/product/3.png",
                        value: metrics.requests,
                        dataPercent: 60
                    },  
                    {
                        name: "Goroutines",
                        image: "../assets/images/product/3.png",
                        value: metrics.goroutines,
                        dataPercent: 60
                    },
                    {
                        name: "Heap Alloc",
                        image: "../assets/images/product/3.png",
                        value: metrics.heap_alloc,
                        dataPercent: 50
                    },
                    {
                        name: "Heap Sys",
                        image: "../assets/images/product/3.png",
                        value: metrics.heap_sys,
                        dataPercent: 40
                    },
                    {
                        name: "Memory Usage",
                        image: "../assets/images/product/3.png",
                        value: metrics.memory_usage,
                        dataPercent: 30
                    },
                    {
                        name: "Sys",
                        image: "../assets/images/product/3.png",
                        value: metrics.sys,
                        dataPercent: 20
                    },
                    {
                        name: "Total Alloc",
                        image: "../assets/images/product/3.png",
                        value: metrics.total_alloc,
                        dataPercent: 10
                    },
                    {
                        name: "Total Duration",
                        image: "../assets/images/product/3.png",
                        value: metrics.total_duration,
                        dataPercent: 5
                    },
                    {
                        name: "Uptime",
                        image: "../assets/images/product/3.png",
                        value: metrics.uptime,
                        dataPercent: 2
                    }
                ];

                runtimeMetricsContainer.innerHTML = ''; 
                let rowHTML = `<div class="row">`;

                listCont.forEach(item => {
                    const cardHTML = `
                        <div class="col-lg-3 col-md-3">
                            <div class="card card-block card-stretch card-height">
                                <div class="card-body">
                                    <div class="d-flex align-items-center mb-4 card-total-sale">
                                        <div class="icon iq-icon-box-2 bg-info-light">
                                            <img src="${item.image}" class="img-fluid" alt="${item.name} Image">
                                        </div>
                                        <div>
                                            <p class="mb-2">${item.name}</p>
                                            <h4>${item.value}</h4>
                                        </div>
                                    </div>
                                    <div class="iq-progress-bar mt-2">
                                        <span class="bg-info iq-progress progress-1" style="width: 0%;" data-percent="${item.dataPercent}">
                                        </span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    `;

                    rowHTML += cardHTML;
                });

                rowHTML += `</div>`;

                runtimeMetricsContainer.innerHTML = rowHTML;


                const progressBars = runtimeMetricsContainer.querySelectorAll('.iq-progress');
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
                goRoutinesNumber.innerHTML = data.go_routines;
                const container = document.getElementById('goroutines-container');
                const countElement = document.getElementById('goroutine-count');

                let goroutines = [];
                data.list.forEach((item, index) => {
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

    function updateGauge(gaugeId, percentage) {
                const gauge = document.getElementById(gaugeId);
                const text = gauge.querySelector('text');

                // Update the text inside the gauge
                text.textContent = `${percentage}%`;

                // Determine the fill color based on the percentage
                let fillColor;
                if (percentage >= 80) {
                    fillColor = 'var(--red)';
                } else if (percentage >= 60) {
                    fillColor = 'var(--yellow)';
                } else if (percentage >= 50) {
                    fillColor = 'var(--orange)';
                } else if (percentage >= 40) {
                    fillColor = 'var(--lightgreen)';
                } else {
                    fillColor = 'var(--green)';
                }

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

    function updateStats() {
        fetch(`/metrics`)
            .then(response => response.json())
            .then(data => {

                console.log(data);
                let preHtml = `
                    <div class="card card-block card-stretch card-height mb-4">
                        <div class="card-body">
                            <div class="d-flex align-items-center mb-4 card-total-sale">
                                <div class="icon iq-icon-box-2 bg-info-light">
                                    <!-- Example image, replace as needed -->
                                    <img src="https://via.placeholder.com/40" class="img-fluid" alt="Memory Image">
                                </div>
                `
                
                for (let [key, value] of Object.entries(data.memory)) {
                    let name = '';
                    if (key === "mem_stats_records") {
                        console.log('Skipping mem_stats_records');
                        continue;
                    }

                    if (key === "used_in_percent") {
                        name = "Memory Usage";
                    } else if (key === "total_memory") {
                        name = "Total Memory";
                        value = `${value} bytes`;
                    } else if (key === "used_by_system") {
                        name = "Used by System";
                        value = `${value} bytes`;
                    } else if (key === "used_by_process") {
                        name = "Used by Process";
                        value = `${value} bytes`;
                    } else if (key === "free_memory") {
                        name = "Free Memory";
                        value = `${value} bytes`;
                    } else if (key === "heap_alloc_by_process") {
                        name = "Heap Alloc by Process";
                        value = `${value} bytes`;
                    } else if (key === "heap_sys_by_process") {
                        name = "Heap Sys by Process";
                        value = `${value} bytes`;
                    } else if (key === "total_alloc_by_process") {
                        name = "Total Alloc by Process";
                        value = `${value} bytes`;
                    } else if (key === "total_sys_by_process") {
                        name = "Total Sys by Process";
                        value = `${value} bytes`;
                    } else {
                        name = key;
                    }

                    // <div>
                    //                         <p class="mb-2">Service Name:</p>
                    //                         <h4>${serviceName}</h4>

                    //                         <p class="mb-2">Go Version:</p>
                    //                         <h4>${goVersion}</h4>
                    //                         <p class="mb-2">Service Start Time:</p>
                    //                         <h4>${formattedDate}<br/> ${formattedTime}</h4>
                    //                     </div>

                    memoryStatsContainer.innerHTML += `
                                    <div>
                                        <p class="mb-2">${name} : </p>
                                        <h4>${value}</h4>
                                    </div>
                                    `;
                                
                }

                let postHtml = `</div>
                            </div>
                        </div>
                    `;

                preHtml + memoryStatsContainer.innerHTML + postHtml;


            // CPU Stats
            cpuStatsContainer.innerHTML = `
                <div class="card card-block card-stretch card-height mb-4">
                    <div class="card-body">
                        <div class="d-flex align-items-center mb-4 card-total-sale">
                            <div class="icon iq-icon-box-2 bg-info-light">
                                <!-- Example image, replace as needed -->
                                <img src="https://via.placeholder.com/40" class="img-fluid" alt="CPU Image">
                            </div>
                            <div>
                                <p class="mb-2">CPU Usage</p>
                                <h4>${data.cpu.used_in_percent}</h4>
                            </div>
                        </div>
                        <div class="iq-progress-bar mt-2">
                            <span class="bg-info iq-progress progress-1" style="width: ${data.cpu.used_in_percent};"
                                data-percent="${data.cpu.used_in_percent}">
                            </span>
                        </div>
                    </div>
                </div>
            `;
        })
        .catch(error => {
            console.error(error);
        });
    }


            // Dynamically set the values
            updateGauge('g1', 20); // Example for the first gauge
            updateGauge('g2', 70); // Example for the second gauge
            updateGauge('g3', 30); // Example for the second gauge
            updateGauge('g4', 40); // Example for the second gauge
            updateGauge('g5', 50); // Example for the second gauge
});
