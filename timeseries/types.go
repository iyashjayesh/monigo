package timeseries

import "github.com/nakabonne/tstorage"

// Label represents a metric label (key-value pair).
type Label struct {
	Name  string
	Value string
}

// DataPoint represents a single time-series data point.
type DataPoint struct {
	Timestamp int64
	Value     float64
}

// Row represents a single metric row to be inserted into storage.
type Row struct {
	Metric    string
	Labels    []Label
	DataPoint DataPoint
}

// toTStorageLabels converts monigo Labels to tstorage Labels.
func toTStorageLabels(labels []Label) []tstorage.Label {
	out := make([]tstorage.Label, len(labels))
	for i, l := range labels {
		out[i] = tstorage.Label{Name: l.Name, Value: l.Value}
	}
	return out
}

// fromTStorageLabels converts tstorage Labels to monigo Labels.
func fromTStorageLabels(labels []tstorage.Label) []Label {
	out := make([]Label, len(labels))
	for i, l := range labels {
		out[i] = Label{Name: l.Name, Value: l.Value}
	}
	return out
}

// toTStorageRows converts monigo Rows to tstorage Rows.
func toTStorageRows(rows []Row) []tstorage.Row {
	out := make([]tstorage.Row, len(rows))
	for i, r := range rows {
		out[i] = tstorage.Row{
			Metric:    r.Metric,
			Labels:    toTStorageLabels(r.Labels),
			DataPoint: tstorage.DataPoint{Timestamp: r.DataPoint.Timestamp, Value: r.DataPoint.Value},
		}
	}
	return out
}

// fromTStorageDataPoints converts tstorage DataPoints to monigo DataPoints.
func fromTStorageDataPoints(points []*tstorage.DataPoint) []DataPoint {
	out := make([]DataPoint, len(points))
	for i, p := range points {
		out[i] = DataPoint{Timestamp: p.Timestamp, Value: p.Value}
	}
	return out
}
