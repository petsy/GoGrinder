package gogrinder

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/finklabs/graceful"
	time "github.com/finklabs/ttime"
	"github.com/prometheus/client_golang/prometheus"
)

type Reporter interface {
	Register(teststep string)
	Update(m Meta)
}

// LogReporter
type LogReporter struct {
	// event logger instance
	elog *log.Logger
}

// Nothing to do but we want to implement the Reporter interface
func (r *LogReporter) Register(teststep string) {}

// Log teststep metrics to the event-log.
func (r *LogReporter) Update(meta Meta) {
	r.elog.WithFields(log.Fields{
		// TODO: timestamp needs proper formatting
		"timestamp": meta["timestamp"].(time.Time),
		"user":      meta["user"].(int),
		"iteration": meta["iteration"].(int),
		"teststep":  meta["teststep"].(string),
		"elapsed":   meta["elapsed"].(time.Duration),
	}).Info()
}

// MetricsReporter
type MetricsReporter struct {
	//elapsed *prometheus.HistogramVec
	elapsed *prometheus.SummaryVec
}

func NewMetricsReporter() *MetricsReporter {
	// I still think that a histogram is the way to go!
	// because computation is taken away from gogrinder
	// but I find Summary is much nicer in Grafana
	//elapsed := prometheus.NewHistogramVec(prometheus.HistogramOpts{
	//	Name: "gogrinder_elapsed_ms",
	//	Help: "Current time elapsed of gogrinder teststep",
	//}, []string{"teststep"})
	//regElapsed := prometheus.MustRegisterOrGet(elapsed).(*prometheus.HistogramVec)
	elapsed := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "gogrinder_elapsed_ms",
		Help:       "Current time elapsed of gogrinder teststep in ms.",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.005, 0.99: 0.001},
	}, []string{"teststep"})
	regElapsed := prometheus.MustRegisterOrGet(elapsed).(*prometheus.SummaryVec)
	return &MetricsReporter{regElapsed}
}

// Implement the reporter interface
func (r *MetricsReporter) Register(teststep string) {}

// Update the GoGrinder node reporter.
func (r *MetricsReporter) Update(meta Meta) {
	r.elapsed.WithLabelValues(
		meta["teststep"].(string),
		//fmt.Sprintf("%d", meta["user"].(int)),
		//fmt.Sprintf("%d", meta["iteration"].(int)),
		//meta["timestamp"].(time.Time).UTC().Format(ISO8601),
	).Observe(float64(meta["elapsed"].(time.Duration)) / float64(time.Millisecond))
}

// Assemble the Server for the Prometheus reporter
func NewPrometheusReporterServer() *graceful.Server {
	//handler := prometheus.Handler()
	// how to disable default go_collector metrics?
	handler := prometheus.UninstrumentedHandler()
	// register the /metrics route
	http.Handle("/metrics", handler)

	srv := &graceful.Server{
		Timeout: 5 * time.Second,
		Server: &http.Server{
			Handler: handler,
		},
	}
	return srv
}
