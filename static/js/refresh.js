document.addEventListener('DOMContentLoaded', () => {

    const refreshBtn = document.getElementById('refresh-btn');
    const refreshCountdown = document.getElementById('refresh-countdown');
    const refreshTime = 300; // 5 minutes

    function startCountdown() {
        let timeLeft = refreshTime;
        refreshCountdown.textContent = `Refreshing in ${Math.floor(timeLeft / 60)}m ${timeLeft % 60}s`;
        const interval = setInterval(() => {
            timeLeft--;
            refreshCountdown.textContent = `Refreshing in ${Math.floor(timeLeft / 60)}m ${timeLeft % 60}s`;
            if (timeLeft <= 0) {
                clearInterval(interval);
                location.reload();
                startCountdown();
            }
        }, 1000);
    }

    refreshBtn.addEventListener('click', () => {
        location.reload();
    });
    startCountdown();
});