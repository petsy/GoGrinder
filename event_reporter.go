package gogrinder

import (
	"encoding/json"
	"fmt"
	"os"
)

type Reporter interface {
	Update(m Metric)
}

// EventReporter
type EventReporter struct {
	logfile *os.File
}

// Log metrics to the event-log.
func (r *EventReporter) Update(m Metric) {
	// often the simple solution is the best!
	s, _ := json.Marshal(m)
	r.logfile.Write(s)
	fmt.Fprintln(r.logfile)
}
