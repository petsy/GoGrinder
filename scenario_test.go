package gogrinder

import (
	"reflect"
	"testing"

	time "github.com/finklabs/ttime"
)

func TestThinktimeNoVariance(t *testing.T) {
	// create a fake loadmodel for testing
	var fake = NewTest()
	fake.status = running
	fake.loadmodel["Scenario"] = "scenario1"
	fake.loadmodel["ThinkTimeFactor"] = 1.0
	fake.loadmodel["ThinkTimeVariance"] = 0.0

	time.Freeze(time.Now())
	defer time.Unfreeze()

	start := time.Now()
	fake.Thinktime(20)
	sleep := time.Now().Sub(start)
	if sleep != 20*time.Millisecond {
		t.Error("Expected to sleep for 20ms but something went wrong!")
	}
}

func TestThinktimeVariance(t *testing.T) {
	// create a fake loadmodel for testing
	var fake = NewTest()
	fake.status = running
	fake.loadmodel["Scenario"] = "scenario1"
	fake.loadmodel["ThinkTimeFactor"] = 2.0
	fake.loadmodel["ThinkTimeVariance"] = 0.1

	min, max, avg := 20.0, 20.0, 0.0
	time.Freeze(time.Now())
	defer time.Unfreeze()

	for i := 0; i < 1000; i++ {
		start := time.Now()
		fake.Thinktime(10)
		sleep := float64(time.Now().Sub(start)) / float64(time.Millisecond)
		if sleep < min {
			min = sleep
		}
		if max < sleep {
			max = sleep
		}
		avg += sleep
	}
	avg = avg / 1000
	if min < 18.0 {
		t.Errorf("Minimum sleep time %f out of defined range!\n", min)
	}
	if max >= 22.0 {
		t.Errorf("Maximum sleep time %f out of defined range!", max)
	}
	t.Logf("Minimum sleep time %f\n", min)
	t.Logf("Maximum sleep time %f\n", max)
	t.Logf("Average sleep time %f\n", avg)
	if avg < 19.9 || avg > 20.1 {
		t.Fatalf("Average sleep time %f out of defined range!", avg)
	}
}

func TestPaceMaker(t *testing.T) {
	time.Freeze(time.Now())
	defer time.Unfreeze()

	var fake = NewTest()
	fake.status = running
	start := time.Now()
	fake.paceMaker(10)
	if time.Now().Sub(start) != 10 {
		t.Fatal("Function paceMaker sleep out of range!")
	}
}

func TestPaceMakerNegativeValue(t *testing.T) {
	time.Freeze(time.Now())
	defer time.Unfreeze()

	var fake = NewTest()
	fake.status = running
	start := time.Now()
	fake.paceMaker(-10)
	if time.Now().Sub(start) != 0 {
		t.Fatal("Function paceMaker sleep out of range!")
	}
}

func TestTestscenario(t *testing.T) {
	var fake = NewTest()
	dummy := func() {}

	fake.Testscenario("sth", dummy)

	if v, ok := fake.testscenarios["sth"]; ok {
		sf1 := reflect.ValueOf(v)
		sf2 := reflect.ValueOf(dummy)
		if sf1.Pointer() != sf2.Pointer() {
			t.Fatal("Testscenario 'sth' does not contain dummy function!")
		}
	} else {
		t.Fatal("Testscenario 'sth' missing!")
	}
}

func TestTeststep(t *testing.T) {
	time.Freeze(time.Now())
	defer time.Unfreeze()

	var fake = NewTest()
	step := func() { time.Sleep(20) }

	its := fake.Teststep("sth", step)

	if v, ok := fake.teststeps["sth"]; ok {
		sf1 := reflect.ValueOf(v)
		sf2 := reflect.ValueOf(its)
		if sf1.Pointer() != sf2.Pointer() {
			t.Fatal("Teststep 'sth' does not contain step function!")
		}
	} else {
		t.Fatal("Teststep 'sth' missing!")
	}

	// run the teststep (note: a different angle would be to mock out update)
	done := fake.collect() // this needs a collector to unblock update
	its()
	fake.wg.Wait()
	close(fake.measurements)
	<-done

	if v, ok := fake.stats["sth"]; ok {

		if v.avg != 20.0 {
			t.Fatalf("Teststep 'sth' measurement %f not 20ns!\n", v.avg)
		}
	} else {
		t.Fatal("Teststep 'sth' missing in stats!")
	}
}

func TestRunSequential(t *testing.T) {
	time.Freeze(time.Now())
	defer time.Unfreeze()

	fake := NewTest()
	var counter int64 = 0
	// assemble testcase
	tc1 := func(meta map[string]interface{}) {
		// check meta
		if meta["Iteration"] != counter {
			t.Errorf("Iteration %d but expected %d!", meta["Iteration"], counter)
		}
		if meta["User"] != 0 {
			t.Error("User meta not as expected!")
		}

		time.Sleep(20)
		counter++
	}

	// run the testcase
	start := time.Now()
	fake.Run(tc1, 20, 0, false)
	if time.Now().Sub(start) != 400 {
		t.Error("Testcase execution time not as expected!")
	}
	if counter != 20 {
		t.Error("Testcase iteration counter not as expected!")
	}

	// TODO run multiple users!
}

func TestScheduleErrorUnknownTestcase(t *testing.T) {
	fake := NewTest()
	err := fake.Schedule("unknown_testcase", func(map[string]interface{}) {})

	error := err.Error()
	if error != "config for testcase unknown_testcase not found" {
		t.Errorf("Error msg for unknown testcase not as expected: %s", error)
	}
}

func TestExecErrorUnknownScenario(t *testing.T) {
	fake := NewTest()
	fake.loadmodel["Scenario"] = "scenario1"
	fake.loadmodel["ThinkTimeFactor"] = 2.0
	fake.loadmodel["ThinkTimeVariance"] = 0.1
	err := fake.Exec()

	error := err.Error()
	if error != "scenario scenario1 does not exist" {
		t.Errorf("Error msg for missing scenario not as expected: %s", error)
	}
}

func TestExecErrorFunctionWithReturnValue(t *testing.T) {
	fake := NewTest()
	fake.loadmodel["Scenario"] = "01_testcase"
	fake.loadmodel["ThinkTimeFactor"] = 2.0
	fake.loadmodel["ThinkTimeVariance"] = 0.1
	fake.Testscenario("01_testcase", func() int64 { return 42 })

	err := fake.Exec()

	error := err.Error()
	if error != "expected a function without return value to implement 01_testcase" {
		t.Errorf("Error msg for function with return value not as expected: %s", error)
	}
}

func TestExecErrorFunctionWithTwoParams(t *testing.T) {
	fake := NewTest()
	fake.loadmodel["Scenario"] = "01_testcase"
	fake.loadmodel["ThinkTimeFactor"] = 2.0
	fake.loadmodel["ThinkTimeVariance"] = 0.1
	fake.Testscenario("01_testcase", func(a, b int64) {})

	err := fake.Exec()

	error := err.Error()
	if error != "expected a function with zero or one parameter to implement 01_testcase" {
		t.Errorf("Error msg for function two or more params not as expected: %s", error)
	}
}
