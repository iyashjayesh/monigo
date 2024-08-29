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
