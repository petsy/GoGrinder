package gogrinder

import (
	"fmt"
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
	elapsed *prometheus.GaugeVec
}

func NewMetricsReporter() *MetricsReporter {
	elapsed := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gogrinder_elapsed_ms",
		Help: "Current time elapsed of gogrinder teststep",
	}, []string{})
	prometheus.MustRegister(elapsed)
	return &MetricsReporter{elapsed}
}

// Implement the reporter interface
func (r *MetricsReporter) Register(teststep string) {}

// Update the GoGrinder node reporter.
func (r *MetricsReporter) Update(meta Meta) {
	r.elapsed.With(prometheus.Labels{
		"timestamp": meta["timestamp"].(time.Time).UTC().Format(ISO8601),
		"user":      fmt.Sprintf("%d", meta["user"].(int)),
		"iteration": fmt.Sprintf("%d", meta["iteration"].(int)),
		"teststep":  meta["teststep"].(string),
	}).Set(float64(meta["elapsed"].(time.Duration)) / float64(time.Millisecond))
}

// Assemble the Server for the Prometheus reporter
func NewPrometheusReporterServer() *graceful.Server {
	handler := prometheus.Handler()
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
