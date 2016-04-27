package gogrinder
import (
	"testing"
	"io/ioutil"
	"os"
	"time"
	"fmt"
	"strings"
)

func TestCheckJtlReporterImplementsReporterInterface(t *testing.T) {
	s := &JtlReporter{}
	if _, ok := interface{}(s).(Reporter); !ok {
		t.Errorf("JtlReporter does not implement the Reporter interface!")
	}
}

func TestJtlReporterUpdateWithSomeMetric(t *testing.T) {
	fake := NewTest()
	tmp, _ := ioutil.TempFile(os.TempDir(), "gogrinder_test")
	defer os.Remove(tmp.Name())
	fake.AddReportPlugin(&JtlReporter{tmp})

	done := fake.Collect() // this needs a collector to unblock update

	now := time.Now()
	fake.Update(Metric(&someMetric{Meta{Teststep: "sth", Elapsed: Elapsed(8 *
		time.Millisecond), Timestamp: Timestamp(now)}, 100}))
	exp := fmt.Sprintf(`%d,8,sth,,,,text,true,,,,`, now.UnixNano()/1000000)

	buf, _ := ioutil.ReadFile(tmp.Name())
	last := strings.TrimSpace(string(buf))
	if last != exp {
		t.Errorf("Entry for SomeMetric expected: %s, but got: %s", exp, last)
	}

	close(fake.measurements)
	<-done
}

func TestJtlReporterUpdateWithSomeMetricError(t *testing.T) {
	fake := NewTest()
	tmp, _ := ioutil.TempFile(os.TempDir(), "gogrinder_test")
	defer os.Remove(tmp.Name())
	fake.AddReportPlugin(&JtlReporter{tmp})

	done := fake.Collect() // this needs a collector to unblock update

	now := time.Now()
	fake.Update(Metric(&someMetric{Meta{Teststep: "sth", Elapsed: Elapsed(8 *
		time.Millisecond), Timestamp: Timestamp(now), Error: "something went wrong!"}, 100}))
    exp := fmt.Sprintf(`%d,8,sth,,,,text,false,,,,`, now.UnixNano()/1000000)

	buf, _ := ioutil.ReadFile(tmp.Name())
	last := strings.TrimSpace(string(buf))
	if last != exp {
		t.Errorf("Entry for SomeMetric expected: %s, but got: %s", exp, last)
	}

	close(fake.measurements)
	<-done
}
