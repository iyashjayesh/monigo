### Status: In Development 🚧

# MoniGo - Performance Monitoring for Go Applications

[![Go Report Card](https://goreportcard.com/badge/github.com/iyashjayesh/monigo)](https://goreportcard.com/report/github.com/iyashjayesh/monigo)
[![GoDoc](https://godoc.org/github.com/iyashjayesh/monigo?status.svg)](https://pkg.go.dev/github.com/iyashjayesh/monigo)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

**MoniGo** is a performance monitoring library for Go applications. It provides real-time insights into application performance with an intuitive user interface, enabling developers to track and optimize both service-level and function-level metrics.

## Features

- **Real-Time Monitoring**: Access up-to-date performance metrics for your Go applications.
- **Detailed Insights**: Track and analyze both service and function-level performance.
- **Customizable Dashboard**: Manage performance data with an easy-to-use UI.
- **Visualizations**: Utilize graphs and charts to interpret performance trends.

## Installation

To install MoniGo, use the following command:

```bash
go get github.com/iyashjayesh/monigo
```

#### For Linux

Install Graphviz:

```bash
sudo apt-get install graphviz
```

Or, if you use Homebrew:

```bash
brew install graphviz
```

#### For Windows

1. Download Graphviz from the following link: https://graphviz.gitlab.io/_pages/Download/Download_windows.html
2. Install Graphviz and set the path in the environment variables.

## Usage

To begin monitoring your Go application, import the monigo package and call the monigo.Start function:

```go
package main

import (
    "github.com/iyashjayesh/monigo"
)

func main() {

    monigoInstance := &monigo.Monigo{
		ServiceName:   "service_name", // Default service name is "service_name"
		DashboardPort: 8080, // Default port is 8080
	}

	monigoInstance.PurgeMonigoStorage()
	monigoInstance.SetDbSyncFrequency("1m") // Default is 5m
	monigoInstance.StartDashboard()

    select {} // To keep the program running
}
```

For more detailed usage instructions, refer to the documentation.

By default, the dashboard will be available at http://localhost:8080/.

You can access the dashboard by visiting the following URL: http://localhost:8080/

Reports need to be generated by the user by clicking on the "Generate Report" button on the Reports page.

## Bellow Reports are available:

1. **Load Statistics**: Provides an overview of the overall load of the service, CPU load, memory load, and system load.

| Field Name                | Value (Datatype) |
| ------------------------- | ---------------- |
| `overall_load_of_service` | `float64`        |
| `service_cpu_load`        | `float64`        |
| `service_memory_load`     | `float64`        |
| `system_cpu_load`         | `float64`        |
| `system_memory_load`      | `float64`        |

2. **CPU Statistics**: Displays the total number of cores, cores used by the service, and cores used by the system.

| Field Name              | Value (Datatype) |
| ----------------------- | ---------------- |
| `total_cores`           | `int`            |
| `cores_used_by_service` | `int`            |
| `cores_used_by_system`  | `int`            |

3. **Memory Statistics**: Shows the total system memory, memory used by the system, memory used by the service, available memory, GC pause duration, and stack memory usage.

| Field Name               | Value (Datatype) |
| ------------------------ | ---------------- |
| `total_system_memory`    | `float64`        |
| `memory_used_by_system`  | `float64`        |
| `memory_used_by_service` | `float64`        |
| `available_memory`       | `float64`        |
| `gc_pause_duration`      | `float64`        |
| `stack_memory_usage`     | `float64`        |

4. **Memory Profile**: Provides information on heap allocation by the service, heap allocation by the system, total allocation by the service, and total memory by the OS.

| Field Name               | Value (Datatype) |
| ------------------------ | ---------------- |
| `heap_alloc_by_service`  | `float64`        |
| `heap_alloc_by_system`   | `float64`        |
| `total_alloc_by_service` | `float64`        |
| `total_memory_by_os`     | `float64`        |

5. **Network IO**: Displays the number of bytes sent and received.

| Field Name       | Value (Datatype) |
| ---------------- | ---------------- |
| `bytes_sent`     | `float64`        |
| `bytes_received` | `float64`        |

6. **Overall Health**: Provides an overall health percentage for the service.

| Field Name               | Value (Datatype) |
| ------------------------ | ---------------- |
| `overall_health_percent` | `float64`        |

## Contributing

We welcome contributions! If you encounter any issues or have suggestions, please submit a pull request or open an issue.

For more information on how to contribute, please refer to the CONTRIBUTING.md file.

If you find MoniGo useful, consider giving it a star! ⭐

## Contact

For questions or feedback, please open an issue or contact me at iyashjayesh@gmail.com

<!-- ## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=iyashjayesh/monigo&type=Date)](https://star-history.com/#iyashjayesh/monigo&Date) -->

<!-- Next things to do -->

<!-- / StartDashboard  -->

<!-- 1. Register StoreInfo Only once when the server starts
   Now on every time interval, do below things:
1. Store Service Metrics
1. Store Runtime Metrics -->

<!-- List of Pages -->

<!-- 1. Dashboard
	1.1. Service Metrics
	1.1.1 Guage Charts for (Service Health, Memory Health, CPU Health etc)
	1.2 CPU Metrics (Detailed)
	1.3 Memory Metrics (Detailed)
	1.4. Go Routines Number Metrics
	1.5. Charts for all the above metrics
		1.5.1 Load Chart
		1.5.2 CPU Chart
<!-- 2. Go Routines Stats -->
<!-- 3. Function Metrics -->
<!-- 4. Performace Report -->
