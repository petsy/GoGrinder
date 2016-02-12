package gogrinder

import (
	"testing"
	"io/ioutil"
	"os"
	"fmt"
	"strings"

	time "github.com/finklabs/ttime"
)

func TestCheckEventReporterImplementsReporterInterface(t *testing.T) {
	s := &EventReporter{}
	if _, ok := interface{}(s).(Reporter); !ok {
		t.Errorf("EventReporter does not implement the Reporter interface!")
	}
}

// someMetric is used in TestEventReporterUpdateWithSomeMetric
type someMetric struct {
	Meta     // std. GoGrinder metric info
	Code int `json:"status"` // http status code
}

func TestEventReporterUpdateWithSomeMetric(t *testing.T) {
	fake := NewTest()
	tmp, _ := ioutil.TempFile(os.TempDir(), "gogrinder_test")
	defer os.Remove(tmp.Name())
	fake.AddReportPlugin(&EventReporter{tmp})

	done := fake.Collect() // this needs a collector to unblock update

	now := time.Now()
	fake.Update(Metric(someMetric{Meta{Teststep: "sth", Elapsed: 8 *
	time.Millisecond, Timestamp: now}, 100}))
	exp := fmt.Sprintf(`{"testcase":"","teststep":"sth","user":0,"iteration"` +
	`:0,"ts":"%s","elapsed":8.000000,"status":100}`, now.Format(time.RFC3339Nano))

	buf, _ := ioutil.ReadFile(tmp.Name())
	last := strings.TrimSpace(string(buf))
	if last != exp {
		t.Errorf("Entry for SomeMetric expected: %s, but got: %s", exp, last)
	}

	close(fake.measurements)
	<-done
}
