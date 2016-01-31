package gogrinder

import (
	"bytes"
	"testing"

	time "github.com/finklabs/ttime"
)

func TestUpdateOneMeasurement(t *testing.T) {
	fake := NewTest()
	// first measurement
	done := fake.Collect() // this needs a collector to unblock update
	fake.Update(Meta{"teststep": "sth", "elapsed": 8 * time.Millisecond, "timestamp": time.Now()})
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
	fake.Update(Meta{"teststep": "sth", "elapsed": 8 * time.Millisecond, "timestamp": time.Now()})
	fake.Update(Meta{"teststep": "sth", "elapsed": 10 * time.Millisecond, "timestamp": time.Now()})
	fake.Update(Meta{"teststep": "sth", "elapsed": 2 * time.Millisecond, "timestamp": time.Now()})
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
	fake.Update(Meta{"teststep": "sth", "elapsed": 8 * time.Millisecond, "timestamp": time.Now()})
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
		fake.Update(Meta{"teststep": name, "elapsed": 8 * time.Millisecond, "timestamp": time.Now()})
		fake.Update(Meta{"teststep": name, "elapsed": 10 * time.Millisecond, "timestamp": time.Now()})
		fake.Update(Meta{"teststep": name, "elapsed": 2 * time.Millisecond, "timestamp": time.Now()})
	}
	insert("tc2")
	insert("tc1")
	insert("tc3")

	close(fake.measurements)
	<-done
	fake.Report() // run the report
	report := stdout.(*bytes.Buffer).String()
	if report != ("tc1, 6.666666, 2.000000, 10.000000, 3\n" +
		"tc2, 6.666666, 2.000000, 10.000000, 3\n" +
		"tc3, 6.666666, 2.000000, 10.000000, 3\n") {
		t.Fatalf("Report output not as expected: %s", report)
	}
}

func TestDuration2Float(t *testing.T) {
	f := d2f(20 * time.Microsecond)
	if f != 0.020 {
		t.Fatalf("Duration to ms float64 conversion %f not as expected", f)
	}
}

func TestReportFeatureToggle(t *testing.T) {
	fake := NewTest()
	// check before
	if fake.reportFeature != true {
		t.Fatalf("ReportFeature should be activated by default.")
	}
	fake.ReportFeature(false)
	// check after
	if fake.reportFeature != false {
		t.Fatalf("ReportFeature should be deactivated after setting it to false!")
	}
}

func TestField2JsonTag(t *testing.T) {
	j := f2j("Teststep")
	if j != "teststep" {
		t.Fatalf("Tag expected: %s but was: %s", "teststep", j)
	}
}

func TestRead(t *testing.T) {
	bak := stdout
	stdout = new(bytes.Buffer)
	defer func() { stdout = bak }()

	fake := NewTest()
	done := fake.Collect() // this needs a collector to unblock update
	insert := func(name string) {
		fake.Update(Meta{"teststep": name, "elapsed": 8 * time.Millisecond, "timestamp": time.Now()})
		fake.Update(Meta{"teststep": name, "elapsed": 10 * time.Millisecond, "timestamp": time.Now()})
		fake.Update(Meta{"teststep": name, "elapsed": 2 * time.Millisecond, "timestamp": time.Now()})
	}
	insert("tc2")
	insert("tc1")
	insert("tc3")

	close(fake.measurements)
	<-done

	report, _ := fake.Csv()
	if report != ("teststep, avg, min, max, count\n" +
		"tc1, 6.666666, 2.000000, 10.000000, 3\n" +
		"tc2, 6.666666, 2.000000, 10.000000, 3\n" +
		"tc3, 6.666666, 2.000000, 10.000000, 3\n") {
		t.Fatalf("Read output not as expected: %s", report)
	}
}
