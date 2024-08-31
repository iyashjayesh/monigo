document.addEventListener('DOMContentLoaded', function() {
    
    const greeting = document.getElementById('greeting');
    const serviceInfoContainer = document.getElementById('service-container');
    const runtimeMetricsContainer = document.getElementById('runtime-metrics-container');
    const goRoutinesNumber = document.getElementById('goroutine-count');

    if (serviceInfoContainer) {
        fetchServiceInfo();
    } else {
        console.error('Element with ID "service-container" not found.');
    }

    if (greeting) {
        fetchGreetings();
    } else {
        console.error('Element with ID "greeting" not found.');
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
                // greeting.innerHTML = '';

                const serviceName = data.service_name;
                const goVersion = data.go_version;
                const serviceStartTime = data.service_start_time;
                // const currentHour = new Date().getHours();
                // let greetingText = '';
                // if (currentHour >= 0 && currentHour < 12) {
                //     greetingText = 'Hello, Good Morning';
                // } else if (currentHour >= 12 && currentHour < 18) {
                //     greetingText = 'Hello, Good Afternoon';
                // } else {
                //     greetingText = 'Hello, Good Evening';
                // }

                // greeting.innerHTML = `${greetingText}!`;

                const date = new Date(serviceStartTime);
                const options = { year: 'numeric', month: 'long', day: 'numeric' };
                const formattedDate = date.toLocaleDateString('en-US', options);
                const timeOptions = { hour: 'numeric', minute: 'numeric', hour12: true };
                const formattedTime = date.toLocaleTimeString('en-US', timeOptions);

                serviceInfoContainer.innerHTML = `
                    <div class="row">
                        <div class="col-lg-4 col-md-4">
                            <div class="card card-block card-stretch card-height">
                                <div class="card-body">
                                    <div class="d-flex align-items-center mb-4 card-total-sale">
                                        <div class="icon iq-icon-box-2 bg-info-light">
                                            <img src="../assets/images/product/1.png" class="img-fluid" alt="image">
                                        </div>
                                        <div>
                                            <p class="mb-2">Service Name:</p>
                                            <h4>${serviceName}</h4>
                                        </div>
                                    </div>
                                    <div class="iq-progress-bar mt-2">
                                        <span class="bg-info iq-progress progress-1" data-percent="85"></span>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <div class="col-lg-4 col-md-4">
                            <div class="card card-block card-stretch card-height">
                                <div class="card-body">
                                    <div class="d-flex align-items-center mb-4 card-total-sale">
                                        <div class="icon iq-icon-box-2 bg-danger-light">
                                            <img src="../assets/images/product/2.png" class="img-fluid" alt="image">
                                        </div>
                                        <div>
                                            <p class="mb-2">Go Version:</p>
                                            <h4>${goVersion}</h4>
                                        </div>
                                    </div>
                                    <div class="iq-progress-bar mt-2">
                                        <span class="bg-danger iq-progress progress-1" data-percent="70"></span>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <div class="col-lg-4 col-md-4">
                            <div class="card card-block card-stretch card-height">
                                <div class="card-body">
                                    <div class="d-flex align-items-center mb-4 card-total-sale">
                                        <div class="icon iq-icon-box-2 bg-success-light">
                                            <img src="../assets/images/product/3.png" class="img-fluid" alt="image">
                                        </div>
                                        <div>
                                            <p class="mb-2">Service Start Time:</p>
                                            <h4>${formattedDate}<br/> ${formattedTime}</h4>
                                        </div>
                                    </div>
                                    <div class="iq-progress-bar mt-2">
                                        <span class="bg-success iq-progress progress-1" data-percent="75"></span>
                                    </div>
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

                console.log(metrics);
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

    function fetchGreetings() {
        greeting.innerHTML = '';

        const currentHour = new Date().getHours();
        let greetingText = '';
        if (currentHour >= 0 && currentHour < 12) {
            greetingText = 'Hello, Good Morning';
        } else if (currentHour >= 12 && currentHour < 18) {
            greetingText = 'Hello, Good Afternoon';
        } else {
            greetingText = 'Hello, Good Evening';
        }

        greeting.innerHTML = `${greetingText}!`;
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
});
