package gogrinder

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	time "github.com/finklabs/ttime"
)

func TestCheckEventReporterImplementsReporterInterface(t *testing.T) {
	s := &EventReporter{}
	if _, ok := interface{}(s).(Reporter); !ok {
		t.Errorf("EventReporter does not implement the Reporter interface!")
	}
}

func TestEventReporterUpdateWithSomeMetric(t *testing.T) {
	fake := NewTest()
	tmp, _ := ioutil.TempFile(os.TempDir(), "gogrinder_test")
	defer os.Remove(tmp.Name())
	fake.AddReportPlugin(&EventReporter{tmp})

	done := fake.Collect() // this needs a collector to unblock update

	now := time.Now()
	fake.Update(Metric(someMetric{Meta{Teststep: "sth", Elapsed: Elapsed(8 *
		time.Millisecond), Timestamp: Timestamp(now)}, 100}))
	exp := fmt.Sprintf(`{"testcase":"","teststep":"sth","user":0,"iteration"`+
		`:0,"ts":"%s","elapsed":8.000000,"status":100}`, now.Format(time.RFC3339Nano))

	buf, _ := ioutil.ReadFile(tmp.Name())
	last := strings.TrimSpace(string(buf))
	if last != exp {
		t.Errorf("Entry for SomeMetric expected: %s, but got: %s", exp, last)
	}

	close(fake.measurements)
	<-done
}

func TestEventReporterUpdateWithSomeMetricError(t *testing.T) {
	fake := NewTest()
	tmp, _ := ioutil.TempFile(os.TempDir(), "gogrinder_test")
	defer os.Remove(tmp.Name())
	fake.AddReportPlugin(&EventReporter{tmp})

	done := fake.Collect() // this needs a collector to unblock update

	now := time.Now()
	fake.Update(Metric(someMetric{Meta{Teststep: "sth", Elapsed: Elapsed(8 *
	time.Millisecond), Timestamp: Timestamp(now), Error: "something went wrong!"}, 100}))
	exp := fmt.Sprintf(`{"testcase":"","teststep":"sth","user":0,"iteration"` +
		`:0,"ts":"%s","elapsed":8.000000,"error":"something went wrong!",` +
		`"status":100}`, now.Format(time.RFC3339Nano))

	buf, _ := ioutil.ReadFile(tmp.Name())
	last := strings.TrimSpace(string(buf))
	if last != exp {
		t.Errorf("Entry for SomeMetric expected: %s, but got: %s", exp, last)
	}

	close(fake.measurements)
	<-done
}
