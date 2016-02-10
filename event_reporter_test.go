package gogrinder

import (
	"testing"
)

func TestCheckEventReporterImplementsReporterInterface(t *testing.T) {
	s := &EventReporter{}
	if _, ok := interface{}(s).(Reporter); !ok {
		t.Errorf("EventReporter does not implement the Reporter interface!")
	}
}
