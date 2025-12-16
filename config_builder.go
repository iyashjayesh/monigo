package monigo

import "net/http"

// MonigoBuilder is the builder for the Monigo struct
type MonigoBuilder struct {
	config *Monigo
}

// NewBuilder creates a new instance of the MonigoBuilder
func NewBuilder() *MonigoBuilder {
	return &MonigoBuilder{
		config: &Monigo{},
	}
}

// WithServiceName sets the service name
func (b *MonigoBuilder) WithServiceName(serviceName string) *MonigoBuilder {
	b.config.ServiceName = serviceName
	return b
}

// WithPort sets the dashboard port
func (b *MonigoBuilder) WithPort(port int) *MonigoBuilder {
	b.config.DashboardPort = port
	return b
}

// WithRetentionPeriod sets the data retention period
func (b *MonigoBuilder) WithRetentionPeriod(period string) *MonigoBuilder {
	b.config.DataRetentionPeriod = period
	return b
}

// WithDataPointsSyncFrequency sets the data points sync frequency
func (b *MonigoBuilder) WithDataPointsSyncFrequency(frequency string) *MonigoBuilder {
	b.config.DataPointsSyncFrequency = frequency
	return b
}

// WithTimeZone sets the time zone
func (b *MonigoBuilder) WithTimeZone(timeZone string) *MonigoBuilder {
	b.config.TimeZone = timeZone
	return b
}

// WithCustomBaseAPIPath sets the custom base API path
func (b *MonigoBuilder) WithCustomBaseAPIPath(path string) *MonigoBuilder {
	b.config.CustomBaseAPIPath = path
	return b
}

// WithMaxCPUUsage sets the max CPU usage
func (b *MonigoBuilder) WithMaxCPUUsage(usage float64) *MonigoBuilder {
	b.config.MaxCPUUsage = usage
	return b
}

// WithMaxMemoryUsage sets the max memory usage
func (b *MonigoBuilder) WithMaxMemoryUsage(usage float64) *MonigoBuilder {
	b.config.MaxMemoryUsage = usage
	return b
}

// WithMaxGoRoutines sets the max Go routines
func (b *MonigoBuilder) WithMaxGoRoutines(routines int) *MonigoBuilder {
	b.config.MaxGoRoutines = routines
	return b
}

// WithDashboardMiddleware sets the dashboard middleware
func (b *MonigoBuilder) WithDashboardMiddleware(middleware ...func(http.Handler) http.Handler) *MonigoBuilder {
	b.config.DashboardMiddleware = middleware
	return b
}

// WithAPIMiddleware sets the API middleware
func (b *MonigoBuilder) WithAPIMiddleware(middleware ...func(http.Handler) http.Handler) *MonigoBuilder {
	b.config.APIMiddleware = middleware
	return b
}

// WithAuthFunction sets the custom authentication function
func (b *MonigoBuilder) WithAuthFunction(authFunc func(*http.Request) bool) *MonigoBuilder {
	b.config.AuthFunction = authFunc
	return b
}

// Build builds the Monigo struct
func (b *MonigoBuilder) Build() *Monigo {
	return b.config
}
