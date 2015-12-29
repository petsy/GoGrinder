package gogrinder

import (
	"bytes"
	time "github.com/finklabs/ttime"
	"testing"
)

func TestUpdateOneMeasurement(t *testing.T) {
	fake := NewTest()
	// first measurement
	done := fake.collect()  // this needs a collector to unblock update
	fake.update("sth", 8*time.Millisecond)
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
	done := fake.collect()  // this needs a collector to unblock update
	fake.update("sth", 8*time.Millisecond)
	fake.update("sth", 10*time.Millisecond)
	fake.update("sth", 2*time.Millisecond)
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
	done := fake.collect()  // this needs a collector to unblock update
	// first measurement
	fake.update("sth", 8*time.Millisecond)
	close(fake.measurements)
	<-done
	if _, ok := fake.stats["sth"]; ok {
		fake.reset()
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
	done := fake.collect()  // this needs a collector to unblock update
	insert := func(name string) {
		fake.update(name, 8*time.Millisecond)
		fake.update(name, 10*time.Millisecond)
		fake.update(name, 2*time.Millisecond)
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
