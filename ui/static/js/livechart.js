function getTorrentCountUrl() {
    return window.location.origin + "/api/torrent/count";
}

function getTorrentMetricsUrl(secondsFromNow, count, timeAxisFormat) {
    return window.location.origin  + "/api/torrent/metrics" + "?SecondsFromNow=" + secondsFromNow + "&Count=" + count + "&TimeAxisFormat=" + timeAxisFormat;
}

function getTorrentCategoriesUrl() {
    return window.location.origin  + "/api/torrent/categories";
}

function LiveTorrentChart(container, label, timeDelta, count, timeAxisFormat, refresh_rate, height = 64) {
    let timer = null;
    let canvas = document.createElement("canvas");
    canvas.id = container + "_" + label
    canvas.height = height

    const chart = new Chart(canvas.getContext("2d"), {
        type: "line",
        data: {
            labels: [],
            datasets: [
                {
                    label: label,
                    data: [],
                    fill: false,
                    tension: 0.42,
                    borderColor: 'rgba(84,178,77, 0.8)'
                }
            ],
            options: {
                scales: {
                    y: {
                        beginAtZero: true
                    },
                    x: {
                        beginAtZero: true
                    }
                },
            }
        }
    });

    document.getElementById(container).appendChild(canvas);

    function onResponse(data) {
        chart.data.labels = data["labels"];
        chart.data.datasets[0].data = data["values"];
        chart.update();
    }

    function reload() {
        fetch(getTorrentMetricsUrl(timeDelta, count, timeAxisFormat)).then((rsp) => rsp.json()).then((data) => onResponse(data));
    }

    reload();

    this.stop = function() {
        if(timer !== null) {
            clearInterval(timer);
            timer = null;
        }
        return this;
    }

    this.start = function () {
        if(timer === null) {
            this.stop();
            timer = setInterval(reload, refresh_rate);
        }
        return this;
    }
}

function TorrentCategoryChart(container, label, height = 32) {
    let canvas = document.createElement("canvas");
    canvas.id = container + "_" + label
    canvas.height = height

    const chart = new Chart(canvas.getContext("2d"), {
        type: "doughnut",
        data: {
            labels: [],
            datasets: [
                {
                    label: label,
                    data: [],
                    fill: false,
                    tension: 0.42,
                    backgroundColor: [
                        'rgb(255,179,186)',
                        'rgb(255,223,186)',
                        'rgb(255,255,186)',
                        'rgb(186,255,201)',
                        'rgb(186,225,255)',
                        'rgb(255,110,120)',
                        'rgb(216,186,255)',
                    ]
                }
            ],
            options: {
                scales: {
                    y: {
                        beginAtZero: true
                    },
                    x: {
                        beginAtZero: true
                    }
                },
            }
        }
    });

    document.getElementById(container).appendChild(canvas);

    function onResponse(data) {
        chart.data.labels = data["labels"];
        chart.data.datasets[0].data = data["values"];
        chart.update();
    }

    function reload() {
        fetch(getTorrentCategoriesUrl()).then((rsp) => rsp.json()).then((data) => onResponse(data));
    }

    reload();
}