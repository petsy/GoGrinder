package gogrinder

import (
	"bytes"
	"testing"

	time "github.com/finklabs/ttime"
)

func TestCheckTestStatisticsImplementsStatisticsInterface(t *testing.T) {
	s := &TestStatistics{}
	if _, ok := interface{}(s).(Statistics); !ok {
		t.Errorf("TestStatistics does not implement the Statistics interface!")
	}
}
func TestUpdateOneMeasurement(t *testing.T) {
	fake := NewTest()
	// first measurement
	done := fake.Collect() // this needs a collector to unblock update

	fake.Update(&Meta{Teststep: "sth", Elapsed: Elapsed(8 * time.Millisecond), Timestamp: Timestamp(time.Now())})
	close(fake.measurements)
	<-done
	if v, ok := fake.stats["sth"]; ok {
		if v.avg != 8*time.Millisecond {
			t.Errorf("Statistics update avg %d not as expected 8ms!\n", v.avg)
		}
		if v.min != 8*time.Millisecond {
			t.Errorf("Statistics update min %d not as expected 8ms!\n", v.min)
		}
		if v.max != 8*time.Millisecond {
			t.Errorf("Statistics update max %d not as expected 8ms!\n", v.max)
		}
	} else {
		t.Errorf("Update failed to insert a value for 'sth'!")
	}
}

func TestUpdateMultipleMeasurements(t *testing.T) {
	fake := NewTest()
	done := fake.Collect() // this needs a collector to unblock update
	fake.Update(&Meta{Teststep: "sth", Elapsed: Elapsed(8 * time.Millisecond), Timestamp: Timestamp(time.Now())})
	fake.Update(&Meta{Teststep: "sth", Elapsed: Elapsed(10 * time.Millisecond), Timestamp: Timestamp(time.Now())})
	fake.Update(&Meta{Teststep: "sth", Elapsed: Elapsed(2 * time.Millisecond), Timestamp: Timestamp(time.Now())})
	close(fake.measurements)
	<-done
	if v, ok := fake.stats["sth"]; ok {
		if v.avg != 6666666*time.Nanosecond {
			t.Errorf("Statistics update avg %d not as expected 6.66ms!\n", v.avg)
		}
		if v.min != 2*time.Millisecond {
			t.Errorf("Statistics update min %d not as expected 2ms!\n", v.min)
		}
		if v.max != 10*time.Millisecond {
			t.Errorf("Statistics update max %d not as expected 10ms!\n", v.max)
		}
	} else {
		t.Errorf("Update failed to insert values for 'sth'!")
	}
}

func TestReset(t *testing.T) {
	fake := NewTest()
	done := fake.Collect() // this needs a collector to unblock update
	// first measurement
	fake.Update(&Meta{Teststep: "sth", Elapsed: Elapsed(8 * time.Millisecond), Timestamp: Timestamp(time.Now())})
	close(fake.measurements)
	<-done
	if _, ok := fake.stats["sth"]; ok {
		fake.Reset()
		// now the measurement should be gone
		if _, ok := fake.stats["sth"]; ok {
			t.Error("Reset failed to clear the statistics!\n")
		}
	} else {
		t.Errorf("Update failed to insert values for 'sth'!")
	}
}

func TestReport(t *testing.T) {
	bak := stdout
	stdout = new(bytes.Buffer)
	defer func() { stdout = bak }()

	fake := NewTest()
	done := fake.Collect() // this needs a collector to unblock update
	insert := func(name string) {
		fake.Update(&Meta{Teststep: name, Elapsed: Elapsed(8 * time.Millisecond), Timestamp: Timestamp(time.Now())})
		fake.Update(&Meta{Teststep: name, Elapsed: Elapsed(10 * time.Millisecond), Timestamp: Timestamp(time.Now())})
		fake.Update(&Meta{Teststep: name, Elapsed: Elapsed(2 * time.Millisecond), Timestamp: Timestamp(time.Now())})
	}
	insert("tc2")
	insert("tc1")
	insert("tc3")

	close(fake.measurements)
	<-done
	fake.Report(stdout) // run the report
	report := stdout.(*bytes.Buffer).String()
	if report != ("tc1, 6.666666, 2.000000, 10.000000, 3, 0\n" +
		"tc2, 6.666666, 2.000000, 10.000000, 3, 0\n" +
		"tc3, 6.666666, 2.000000, 10.000000, 3, 0\n") {
		t.Fatalf("Report output not as expected: %s", report)
	}
}

func TestDuration2Float(t *testing.T) {
	f := d2f(20 * time.Microsecond)
	if f != 0.020 {
		t.Fatalf("Duration to ms float64 conversion %f not as expected", f)
	}
}

func TestField2JsonTag(t *testing.T) {
	j := f2j("Teststep")
	if j != "teststep" {
		t.Fatalf("Tag expected: %s but was: %s", "teststep", j)
	}
}

func TestCsv(t *testing.T) {
	bak := stdout
	stdout = new(bytes.Buffer)
	defer func() { stdout = bak }()

	fake := NewTest()
	done := fake.Collect() // this needs a collector to unblock update
	insert := func(name string) {
		fake.Update(&Meta{Teststep: name, Elapsed: Elapsed(8 * time.Millisecond), Timestamp: Timestamp(time.Now())})
		fake.Update(&Meta{Teststep: name, Elapsed: Elapsed(10 * time.Millisecond), Timestamp: Timestamp(time.Now())})
		fake.Update(&Meta{Teststep: name, Elapsed: Elapsed(2 * time.Millisecond), Timestamp: Timestamp(time.Now())})
	}
	insert("tc2")
	insert("tc1")
	insert("tc3")

	close(fake.measurements)
	<-done

	report, _ := fake.Csv()
	if report != ("teststep, avg_ms, min_ms, max_ms, count, error\n" +
		"tc1, 6.666666, 2.000000, 10.000000, 3, 0\n" +
		"tc2, 6.666666, 2.000000, 10.000000, 3, 0\n" +
		"tc3, 6.666666, 2.000000, 10.000000, 3, 0\n") {
		t.Fatalf("Read output not as expected: %s", report)
	}
}

func TestSetReportPlugins(t *testing.T) {
	mr := NewMetricReporter()
	ts := &TestStatistics{}
	ts.SetReportPlugins(mr)

	if ts.reporters[0] != mr {
		t.Errorf("SetReportPlugins did not set the MetricReporter")
	}
}

func TestAddReportPlugin(t *testing.T) {
	ts := &TestStatistics{}
	mr := NewMetricReporter()
	er := &EventReporter{}
	ts.SetReportPlugins(mr)
	ts.AddReportPlugin(er)

	if ts.reporters[0] != mr {
		t.Errorf("AddReportPlugin changed already set reporters!")
	}

	if ts.reporters[1] != er {
		t.Errorf("AddReportPlugin did not set the EventReporter!")
	}
}

func TestReportWithSomeMetric(t *testing.T) {
	bak := stdout
	stdout = new(bytes.Buffer)
	defer func() { stdout = bak }()

	fake := NewTest()
	done := fake.Collect() // this needs a collector to unblock update
	insert := func(name string) {
		fake.Update(Metric(&someMetric{Meta{Teststep: name, Elapsed: Elapsed(8 * time.Millisecond), Timestamp: Timestamp(time.Now())}, 100}))
		fake.Update(Metric(&someMetric{Meta{Teststep: name, Elapsed: Elapsed(10 * time.Millisecond), Timestamp: Timestamp(time.Now())}, 200}))
		fake.Update(Metric(&someMetric{Meta{Teststep: name, Elapsed: Elapsed(2 * time.Millisecond), Timestamp: Timestamp(time.Now())}, 300}))
	}
	insert("tc2")
	insert("tc1")
	insert("tc3")

	close(fake.measurements)
	<-done
	fake.Report(stdout) // run the report
	report := stdout.(*bytes.Buffer).String()
	if report != ("tc1, 6.666666, 2.000000, 10.000000, 3, 0\n" +
		"tc2, 6.666666, 2.000000, 10.000000, 3, 0\n" +
		"tc3, 6.666666, 2.000000, 10.000000, 3, 0\n") {
		t.Fatalf("Report output not as expected: %s", report)
	}
}

type someReporter struct{}

func (s someReporter) Update(m Metric) {
	//v := reflect.ValueOf(m)
	//teststep := v.FieldByName("Teststep").Interface().(string)
	teststep := m.GetTeststep()
	if teststep != "sth" {
		panic("Error: invalid teststep name during test execution.")
	}
}

func TestCollectCallsReporterUpdate(t *testing.T) {
	fake := NewTest()
	fake.AddReportPlugin(someReporter{})

	// first measurement
	done := fake.Collect() // this needs a collector to unblock update

	fake.Update(&Meta{Teststep: "sth", Elapsed: Elapsed(8 * time.Millisecond), Timestamp: Timestamp(time.Now())})
	close(fake.measurements)
	<-done
}
