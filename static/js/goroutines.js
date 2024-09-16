document.addEventListener('DOMContentLoaded', () => {
    const goRoutinesNumber = document.getElementById('goroutine-count');

    if (goRoutinesNumber) {
        fetchGoRoutines();
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


                if (goroutines.length > 0) {
                    const downloadBtn = document.getElementById('download-stack-view');
                    if (downloadBtn) {
                        downloadBtn.style.display = 'block';
                        downloadBtn.addEventListener('click', () => {
                            const blob = new Blob([goroutines.map(g => g.stackTrace).join('\n')], { type: 'text/plain' });
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