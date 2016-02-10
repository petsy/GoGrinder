package gogrinder

import (
	time "github.com/finklabs/ttime"
)

// Datatype to collect reference information about the execution of a teststep
type Meta struct {
	Testcase  string        `json:"testcase"`
	Teststep  string        `json:"teststep"`
	User      int           `json:"user"`
	Iteration int           `json:"iteration"`
	Timestamp time.Time     `json:"ts"`
	Elapsed   time.Duration `json:"elapsed"` // elapsed time [ns]
	Error     string        `json:"error,omitempty"`
}

// Every type implements the Metric type since it is so simple.
// Only important thing is that every Metric type embeds Meta.
type Metric interface {}
