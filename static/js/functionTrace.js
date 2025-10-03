document.addEventListener('DOMContentLoaded', () => {
    (() => {
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

        const loadingHtml = `
            <div class="loader-container">
                <div class="bouncing-dots">
                    <div class="dot"></div>
                    <div class="dot"></div>
                    <div class="dot"></div>
                </div>
            </div>`;

        const uiElements = {
            healthMessageContainer: document.getElementById('health-message'),
            functionDetailsContainer: document.getElementById('function-details'),
            totalFunctionCount: document.getElementById('totalFNumber'),
        };

        Object.values(uiElements).forEach(el => el && (el.innerHTML = loadingHtml));

        function fetchAndDisplayFunctionMetrics() {
            authenticatedFetch(`/monigo/api/v1/function`)
                .then(response => response.json())
                .then(functionData => {
                    const { functionDetailsContainer, totalFunctionCount } = uiElements;
                    if (Object.keys(functionData).length === 0) {
                        functionDetailsContainer.innerHTML = `<p class='ml-4'>No function metrics are currently available. This may indicate that no functions have been instrumented for monitoring.</p>`;
                        totalFunctionCount.innerHTML = `<h3><strong>0</strong></h3>`;
                        return;
                    }

                    totalFunctionCount.innerHTML = `<h3><strong>${Object.keys(functionData).length}</strong></h3>`;
                    functionDetailsContainer.innerHTML = Object.keys(functionData).map((funcName) => {
                        const { function_last_ran_at: lastRanAt } = functionData[funcName];
                        return `
                            <div class="col-lg-4 col-md-4">
                                <div class="card card-block card-stretch card-height">
                                    <div class="card-body card-item-right">
                                        <div class="d-flex align-items-top">
                                            <div class="style-text text-left">
                                                <h5 class="mb-2">Function Name:</h5>
                                                <p class="mb-2">${funcName}</p>
                                                <p class="mb-0">Last Ran At: ${lastRanAt}</p>
                                            </div>
                                            <div class="card-header-toolbar d-flex align-items-center">
                                                <div><a href="#" class="btn btn-primary view-btn font-size-14" data-func-name="${funcName}">Detailed View</a></div>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        `;
                    }).join('');
                    document.querySelectorAll('.view-btn').forEach(button => {
                        button.addEventListener('click', (event) => {
                            const funcName = event.target.getAttribute('data-func-name');
                            openFunctionDetailModal(funcName);
                        });
                    });
                })
                .catch(error => {
                    console.error('Error fetching function metrics:', error);
                    uiElements.functionDetailsContainer.textContent = "An error occurred while fetching metrics. Please try again later.";
                });
        }

        function openFunctionDetailModal(funcName) {
            const modalHtml = `
                <div class="modal fade bd-example-modal-xl" id="functionDetailModal" tabindex="-1" role="dialog" aria-labelledby="functionDetailModalTitle" aria-hidden="true">
                    <div class="modal-dialog modal-xl modal-dialog-scrollable" role="document">
                        <div class="modal-content">
                            <div class="modal-header">
                                <h5 class="modal-title" id="functionDetailModalTitle">Details for "${funcName}"</h5>
                                <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                                    <span aria-hidden="true">&times;</span>
                                </button>
                            </div>
                            <div class="modal-body">
                                <div class="form-group">
                                    <label for="reportTypeSelect">Select Type</label>
                                    <div id="reportTypeButtons" class="btn-group ml-3" role="group">
                                        <button type="button" class="btn btn-outline-primary report-type-btn active" data-bs-toggle="tooltip" title="Prints the location entries, one per line, including the flat and cum values. For more info check 'pprof'" data-report-type="text" id="btn-text">Text<span style="color:red; font-size: 18px; vertical-align: top;">*</span></button>
                                        <button type="button" class="btn btn-outline-primary report-type-btn" data-bs-toggle="tooltip" title="Prints each location entry with its predecessors and successors. For more info check 'pprof'" data-report-type="traces" id="btn-traces">Traces<span style="color:red; font-size: 18px; vertical-align: top;">*</span></button>
                                        <button type="button" class="btn btn-outline-primary report-type-btn" data-bs-toggle="tooltip" title="Prints each sample with a location per line. For more info check 'pprof'" data-report-type="tree" id="btn-tree">Tree<span style="color:red; font-size: 18px; vertical-align: top;">*</span></button>
                                    </div>
                                </div>
                                <div id="function-details-content">Loading details...</div>
                            </div>
                            <div class="modal-footer">
                                <button type="button" class="btn btn-secondary" data-dismiss="modal">Close</button>
                            </div>
                        </div>
                    </div>
                </div>
            `;
            if (!document.getElementById('functionDetailModal')) {
                document.body.insertAdjacentHTML('beforeend', modalHtml);
            }
            setTimeout(() => {
                const modalElement = new bootstrap.Modal(document.getElementById('functionDetailModal'));
                modalElement.show();
            }, 100);

            const initializeTooltips = () => {
                const tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
                tooltipTriggerList.forEach(function (tooltipTriggerEl) {
                    new bootstrap.Tooltip(tooltipTriggerEl);
                });
            };
            initializeTooltips();

            const fetchFunctionDetails = (reportType) => {
                authenticatedFetch(`/monigo/api/v1/function-details?name=${funcName}&reportType=${reportType}`)
                    .then(response => response.json())
                    .then(details => {
                        const content = `
                            <h5>Code Trace</h5>
                            <pre>${details?.function_code_trace || 'No code trace data available.'}</pre>
                            <h5>Core Profile</h5>
                            <pre>${details?.core_profile?.cpu_profile || 'No core profile data available.'}</pre>
                            <h5>Memory Profile</h5>
                            <pre>${details?.core_profile?.mem_profile || 'No memory profile data available.'}</pre>
                        `;
                        document.getElementById('function-details-content').innerHTML = content;
                    })
                    .catch(error => {
                        console.error('Error fetching function details:', error);
                        document.getElementById('function-details-content').innerHTML = `
                            <div class="alert alert-danger" role="alert">
                                Error loading function details. Please try again later.
                            </div>
                        `;
                    });
            };
            fetchFunctionDetails('text');
            document.querySelectorAll('.report-type-btn').forEach(button => {
                button.addEventListener('click', (event) => {
                    const selectedReportType = event.target.getAttribute('data-report-type');
                    fetchFunctionDetails(selectedReportType);
                    document.querySelectorAll('.report-type-btn').forEach(btn => btn.classList.remove('active'));
                    event.target.classList.add('active');
                });
            });
            document.getElementById('functionDetailModal').addEventListener('hidden.bs.modal', function () {
                document.getElementById('functionDetailModal').remove();
            });
        }

        fetchAndDisplayFunctionMetrics();
    })();
});
