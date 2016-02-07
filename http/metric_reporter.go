package http

import (
	"github.com/finklabs/GoGrinder"
	time "github.com/finklabs/ttime"
	"github.com/prometheus/client_golang/prometheus"
)

// helper
func newSummaryVec(name string, help string) *prometheus.SummaryVec {
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

type HttpMetric struct {
	gogrinder.Meta               // std. GoGrinder metric info
	firstByte      time.Duration // first byte after [ns]
	bytes          int           // response size [kb]
	code           int           // http status code
	err            string        // error message
}

// implement the Metric interface
func (m HttpMetric) GetValues() map[string]string {
	return nil
}

func (m HttpMetric) GetMeta() gogrinder.Meta {
	return m.Meta
}

// Specific prometheus reporter for HttpMetric.
// All metrics are represents as vectors of teststeps
type HttpMetricReporter struct {
	elapsed   *prometheus.SummaryVec
	firstByte *prometheus.SummaryVec
	bytes     *prometheus.SummaryVec
	code      *prometheus.CounterVec
	error     *prometheus.CounterVec
}

func NewHttpMetricReporter() *HttpMetricReporter {
	return &HttpMetricReporter{
		newSummaryVec(
			"gogrinder_elapsed_ms",
			"Current time elapsed of gogrinder teststep in ms."),
		newSummaryVec(
			"gogrinder_first_byte_ms",
			"Current time of gogrinder teststep until first byte received in ms."),
		newSummaryVec(
			"gogrinder_response_kb",
			"Current response of gogrinder teststep in kb."),
		prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gogrinder_response_code_count",
			Help: "Current response code of gogrinder teststep.",
		}, []string{"teststep", "code"}),
		prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "gogrinder_error_count",
			Help: "Current error of gogrinder teststep.",
		}, []string{"teststep"}),
	}
}

// We did not find out a way to implement a generic prometheus reporter.
// So this is a specific prometheus reporter that deals with HttpMetric values.
func (r *HttpMetricReporter) Update(m gogrinder.Metric) {
	// find out if we deal with a HttpMetric
	if h, ok := m.(HttpMetric); ok {
		r.bytes.WithLabelValues(h.Teststep).Observe(float64(h.bytes) / float64(1024))
		r.firstByte.WithLabelValues(h.Teststep).Observe(float64(h.firstByte))
		r.elapsed.WithLabelValues(h.Teststep).Observe(float64(h.Elapsed))
		r.code.WithLabelValues(h.Teststep, string(h.code)).Inc()
		r.error.WithLabelValues(h.Teststep).Inc()
	}
}
