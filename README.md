<p align="center">
  <img src="./static/assets/monigo-icon.png" width="200" title="Monigo Icon" alt="monigo-icon"/>
</p>

# MoniGo - Performance Monitoring for Go Applications

[![Go Report Card](https://goreportcard.com/badge/github.com/iyashjayesh/monigo)](https://goreportcard.com/report/github.com/iyashjayesh/monigo)
[![GoDoc](https://godoc.org/github.com/iyashjayesh/monigo?status.svg)](https://pkg.go.dev/github.com/iyashjayesh/monigo)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![Visitors](https://api.visitorbadge.io/api/visitors?path=iyashjayesh%2Fmonigo%20&countColor=%23263759&style=flat)
![GitHub last commit](https://img.shields.io/github/last-commit/iyashjayesh/monigo)

**MoniGo** is a performance monitoring library for Go applications. It provides real-time insights into application performance with an intuitive user interface, enabling developers to track and optimize both service-level and function-level metrics.

<div align="center">
  <img src="monigo.gif" width="100%" alt="monigo-gif">
</div>

## Features

- **Asynchronous Telemetry Pipeline** — Zero impact on application latency
- **Adaptive Sampling** — Intelligent function tracing with configurable sampling rates
- **Pluggable Storage** — Persistent disk or volatile in-memory storage
- **Headless Mode** — Run as a background telemetry agent without the dashboard
- **Real-Time Dashboard** — Intuitive UI with graphs, charts, and downloadable reports
- **Prometheus & OpenTelemetry** — Built-in `/metrics` endpoint and OTel Collector export
- **Router Integration** — Works with standard mux, Gin, Echo, Chi, Fiber, and more
- **Dashboard Security** — Basic Auth, API Key, IP Whitelist, and Rate Limiting middleware

> 📖 **Full documentation**: [iyashjayesh.github.io/monigo-website](https://iyashjayesh.github.io/monigo-website)

## Installation

```bash
go get github.com/iyashjayesh/monigo@latest
```

## Quick Start

```go
package main

import (
    "context"
    "log"
    "math"
    "net/http"

    "github.com/iyashjayesh/monigo"
)

func main() {
    monigoInstance := monigo.NewBuilder().
        WithServiceName("data-api").
        WithPort(8080).
        WithRetentionPeriod("4d").
        WithDataPointsSyncFrequency("5s").
        WithSamplingRate(100).
        WithStorageType("memory").
        Build()

    go func() {
        if err := monigoInstance.Start(); err != nil {
            log.Fatalf("Failed to start MoniGo: %v", err)
        }
    }()
    log.Printf("Monigo dashboard started at port %d\n", monigoInstance.GetRunningPort())

    http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
        monigo.TraceFunction(r.Context(), highCPUUsage)
        w.Write([]byte("done"))
    })

    log.Fatal(http.ListenAndServe(":8000", nil))
}

func highCPUUsage() {
    var sum float64
    for i := 0; i < 1e8; i++ {
        sum += math.Sqrt(float64(i))
    }
}
```

By default, the dashboard will be available at `http://localhost:8080/`.

## Documentation

For detailed guides and API reference, visit the **[MoniGo Documentation Site](https://iyashjayesh.github.io/monigo-website)**:

| Topic | Link |
|-------|------|
| Introduction & Features | [Docs →](https://iyashjayesh.github.io/monigo-website/guides/introduction/) |
| Configuration (Builder API) | [Docs →](https://iyashjayesh.github.io/monigo-website/guides/configuration/) |
| Function Tracing | [Docs →](https://iyashjayesh.github.io/monigo-website/guides/function-tracing/) |
| Router Integration | [Docs →](https://iyashjayesh.github.io/monigo-website/guides/router-integration/) |
| Dashboard Security | [Docs →](https://iyashjayesh.github.io/monigo-website/guides/security/) |
| Examples | [Docs →](https://iyashjayesh.github.io/monigo-website/guides/examples/) |
| Benchmarks | [Docs →](https://iyashjayesh.github.io/monigo-website/reference/benchmarks/) |
| API Reference | [Docs →](https://iyashjayesh.github.io/monigo-website/reference/api-reference/) |
| Migration (v1 → v2) | [Docs →](https://iyashjayesh.github.io/monigo-website/reference/migration-v1-to-v2/) |

## Contributing

We welcome contributions! If you encounter any issues or have suggestions, please submit a pull request or open an issue.

**If you find MoniGo useful, consider giving it a star! ⭐**

## Contact

For questions or feedback, please open an issue or contact me at `iyashjayesh@gmail.com` or at [LinkedIn](https://www.linkedin.com/in/iyashjayesh/)

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=iyashjayesh/monigo&type=Date)](https://star-history.com/#iyashjayesh/monigo&Date)

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE file](https://github.com/iyashjayesh/monigo?tab=Apache-2.0-1-ov-file) for details.
