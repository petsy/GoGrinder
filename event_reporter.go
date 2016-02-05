package gogrinder

import (
	log "github.com/Sirupsen/logrus"
)

type Reporter interface {
	Update(m Metric)
}

// EventReporter
type EventReporter struct {
	// event logger instance
	elog *log.Logger
}

// TODO FIXME
// Log teststep metrics to the event-log.
func (r *EventReporter) Update(m Metric) {

	r.elog.WithFields(log.Fields{
		"teststep": m.GetValues()["teststep"],
		// TODO: timestamp needs proper formatting
		/*"timestamp": meta["timestamp"].(time.Time),
		"user":      meta["user"].(int),
		"iteration": meta["iteration"].(int),
		"teststep":  meta["teststep"].(string),
		"elapsed":   meta["elapsed"].(time.Duration),*/
	}).Info()
}
