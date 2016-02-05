package gogrinder

import (
	time "github.com/finklabs/ttime"
)

// Datatype to collect reference information about the execution of a teststep
type Meta struct {
	Testcase  string
	Teststep  string
	User      int
	Iteration int
	Timestamp time.Time
	Elapsed   time.Duration // elapsed time [ns]
}

type Metric interface {
	GetValues() map[string]string
	GetMeta() Meta
}

// implement the Metric interface for Meta so it can be used for "simple" case
func (m Meta) GetValues() map[string]string {
	return nil
}

func (m Meta) GetMeta() Meta {
	return m
}
