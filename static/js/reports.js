document.addEventListener('DOMContentLoaded', () => {

    const refreshHtml = `
        <div class="loader-container">
            <div class="bouncing-dots">
                <div class="dot"></div>
                <div class="dot"></div>
                <div class="dot"></div>
            </div>
        </div>`;


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


    function updateTable(metric, timeframe) {

        let StartTime = new Date();
        let EndTime = new Date();

        if (timeframe == "5m") {
            StartTime = new Date(new Date().getTime() - 5 * 60000); // Subtract 5 minutes
        } else if (timeframe == "10m") {
            StartTime = new Date(new Date().getTime() - 10 * 60000); // Subtract 10 minutes
        } else if (timeframe == "30m") {
            StartTime = new Date(new Date().getTime() - 30 * 60000); // Subtract 30 minutes
        } else if (timeframe == "1h") {
            StartTime = new Date(new Date().getTime() - 60 * 60000); // Subtract 1 hour
        } else if (timeframe == "6h") {
            StartTime = new Date(new Date().getTime() - 360 * 60000); // Subtract 6 hours
        } else if (timeframe == "1d") {
            StartTime = new Date(new Date().getTime() - 1440 * 60000); // Subtract 1 day
        } else if (timeframe == "3d") {
            StartTime = new Date(new Date().getTime() - 4320 * 60000); // Subtract 3 days
        } else if (timeframe == "1week") {
            StartTime = new Date(new Date().getTime() - 10080 * 60000); // Subtract 1 week
        } else if (timeframe == "1month") {
            StartTime = new Date(new Date().getTime() - 43200 * 60000); // Subtract 1 month
        }

        let reqObj = {
            topic: metric,
            start_time: toLocalISOString(StartTime),
            end_time: toLocalISOString(EndTime),
            time_frame: timeframe
        };

        fetch('/monigo/api/v1/reports', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(reqObj)
        })
        .then(response => response.json())
        .then(data => {
            // const sectionTitle = document.getElementById('sectionTitle');
            const tablesContainer = document.getElementById('tablesContainer');

            if (data.length > 0) {
                // sectionTitle.textContent = `${metric} - ${timeframe}`;
                const table = createTable(topic, data);
                tablesContainer.appendChild(table);

                // Show the download button
                const downloadBtn = document.getElementById('downloadBtn');
                if (downloadBtn) {
                    downloadBtn.style.display = 'block';
                    downloadBtn.addEventListener('click', () => downloadCSV(data, metric));
                }
            } else {
                // sectionTitle.textContent = 'No Data Available';
                tablesContainer.innerHTML = '';
                const downloadBtn = document.getElementById('downloadBtn');
                if (downloadBtn) {
                    downloadBtn.style.display = 'none';
                }
            }
        }).catch((error) => {
            console.error('Error:', error);
        });
    }

    function downloadCSV(data, metric) {
        const directHeaders = data.length > 0 ? Object.keys(data[0]).filter(header => header !== 'value') : [];
        const valueHeaders = data.length > 0 && data[0].value ? Object.keys(data[0].value) : [];
        const headers = [...directHeaders, ...valueHeaders];

        // headers in toUpperCase
        const headersUpperCase = headers.map(header => header.replace(/_/g, ' ').toUpperCase());

        const csvContent = [
            headersUpperCase.join(','),
            ...data.map(item => {
                const directValues = directHeaders.map(header => item[header]);
                const valueValues = valueHeaders.map(header => item.value[header]);
                return [...directValues, ...valueValues].join(',');
            })
        ].join('\n');

        const blob = new Blob([csvContent], { type: 'text/csv' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${metric}.csv`;
        a.click();
        URL.revokeObjectURL(url);
    }


    function createTable(topic, data) {    
        const table = document.createElement('div');
        table.classList.add('table-responsive', 'rounded', 'mb-3');
        const tableElement = document.createElement('table');
        tableElement.classList.add('data-table', 'table', 'mb-0', 'tbl-server-info');
        const thead = document.createElement('thead');
        thead.classList.add('bg-white', 'text-uppercase');
        const tbody = document.createElement('tbody');
        tbody.classList.add('ligth-body');
        const directHeaders = data.length > 0 ? Object.keys(data[0]).filter(header => header !== 'value') : [];
        const valueHeaders = data.length > 0 && data[0].value ? Object.keys(data[0].value) : [];
        const headers = [...directHeaders, ...valueHeaders];

        const headerRow = document.createElement('tr');
        headerRow.classList.add('ligth', 'ligth-data');
        headers.forEach(header => {
            const th = document.createElement('th');
            th.textContent = header.replace(/_/g, ' ').toUpperCase();
            headerRow.appendChild(th);
        });
        thead.appendChild(headerRow);

        data.forEach(item => {
            const row = document.createElement('tr');
            headers.forEach(header => {
                const td = document.createElement('td');
                if (header in item) {
                    td.textContent = item[header];
                } else if (header in item.value) {
                    td.textContent = item.value[header];
                } else {
                    td.textContent = '';
                }
                row.appendChild(td);
            });
            tbody.appendChild(row);
        });

        tableElement.appendChild(thead);
        tableElement.appendChild(tbody);
        table.appendChild(tableElement);
        document.body.appendChild(table);
        return table;
    }

     // Function to update chart based on selections
    function updateTableCompo() {

        // const sectionTitle = document.getElementById('sectionTitle');
        const tablesContainer = document.getElementById('tablesContainer');
        // sectionTitle.textContent = 'Loading...';
        tablesContainer.innerHTML = '';
        const metricSelect = document.getElementById('topic').value;
        const timeSelect = document.getElementById('timeframe').value;
        updateTable(metricSelect, timeSelect);
    }

    document.getElementById('topic').addEventListener('change', updateTableCompo);
    document.getElementById('timeframe').addEventListener('change', updateTableCompo);
    updateTableCompo();
});