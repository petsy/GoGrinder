package gogrinder

import (
	time "github.com/finklabs/ttime"
	"github.com/prometheus/client_golang/prometheus"
)

// helper
func NewSummaryVec(name string, help string) *prometheus.SummaryVec {
	// I still think that a histogram is the way to go!
	// because computation is taken away from gogrinder
	// but I find Summary is much nicer in Grafana
	//elapsed := prometheus.NewHistogramVec(prometheus.HistogramOpts{
	//	Name: "gogrinder_elapsed_ms",
	//	Help: "Current time elapsed of gogrinder teststep",
	//}, []string{"teststep"})
	//regElapsed := prometheus.MustRegisterOrGet(elapsed).(*prometheus.HistogramVec)
	return prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       name,
		Help:       help,
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.005, 0.99: 0.001},
	}, []string{"teststep"})
}


// Specific prometheus reporter for Meta metric.
// All metrics are represents as vectors of teststeps
type MetricReporter struct {
	elapsed   *prometheus.SummaryVec
	error     *prometheus.CounterVec
}

func NewMetricReporter() *MetricReporter {
	return &MetricReporter{
		NewSummaryVec(
			"gogrinder_elapsed_ms",
			"Current time elapsed of gogrinder teststep in ms."),
		prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gogrinder_error_count",
			Help: "Current error of gogrinder teststep.",
		}, []string{"teststep"}),
	}
}

// We did not find out a way to implement a generic prometheus reporter.
// So this is a specific prometheus reporter that deals with HttpMetric values.
func (r *MetricReporter) Update(m Metric) {
	// find out if we deal with a Meta
	if h, ok := m.(Meta); ok {
		r.elapsed.WithLabelValues(h.Teststep).Observe(float64(h.Elapsed) / float64(time.Millisecond))
		if len(h.Error) > 0 {	r.error.WithLabelValues(h.Teststep).Inc() }
	}
}
