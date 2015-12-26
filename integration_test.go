package gogrinder

import (
	"bytes"
	time "github.com/finklabs/ttime"
	"testing"
)

// TODO
// - test integration of various parts
// - test error handling for command line

// this testsuite's aim is to cover the scope of the samples in
// github.com/finklabs/GoGrinder-samples/simple

// initialize the GoGrinder
var gg = NewTest()

// sleep step factory
func myStep(duration time.Duration) func() {
	return func() {
		time.Sleep(duration * time.Millisecond)
	}
}

// instrument teststeps
var ts1 = gg.Teststep("01_01_teststep", myStep(100))
var ts2 = gg.Teststep("02_01_teststep", myStep(200))
var ts3 = gg.Teststep("03_01_teststep", myStep(300))
var thinktime = gg.Thinktime

// define testcases using teststeps
func tc1(meta map[string]interface{}) {
	ts1()
	thinktime(50)
}
func tc2(meta map[string]interface{}) {
	ts2()
	thinktime(50)
}
func tc3(meta map[string]interface{}) {
	ts3()
	thinktime(50)
}

func baseline() {
	// use the tests with a explicit configuration
	// baseline scenario has no concurrency so everything runs sequentially
	gg.Run(tc1, 5, 0, false)
	gg.Run(tc2, 5, 0, false)
	gg.Run(tc3, 5, 0, false)
}

// here the integration testcases

func TestBaseline(t *testing.T) {
	time.Freeze(time.Now())
	defer time.Unfreeze()
	bak := stdout
	stdout = new(bytes.Buffer)
	defer func() { stdout = bak }()

	// we do not need a full loadmodel for this
	loadmodel := `{
	  "Scenario": "baseline",
	  "ThinkTimeFactor": 2.0,
	  "ThinkTimeVariance": 0.0
	}`
	//  no Loadmodel required! ,"Loadmodel": []

	// init
	gg.Testscenario("baseline", baseline)

	// main part
	gg.ReadLoadmodelSchema(loadmodel, LoadmodelSchema)
	//gogrinder.Webserver()  // not necessary for the integration test

	start := time.Now()

	gg.Exec() // exec the scenario that has been selected in the config file
	execution := time.Now().Sub(start)

	// verify total run time of the baseline senario
	// 15 * 50 * 2 + 500 + 1000 + 1500 = 4500
	if execution != 4500*time.Millisecond {
		t.Errorf("Error: execution time of baseline test not as expected: %v\n", execution)
	}

	// verify Report!
	report := stdout.(*bytes.Buffer).String()
	if report != ("01_01_teststep, 100.000000, 100.000000, 100.000000, 5\n" +
		"02_01_teststep, 200.000000, 200.000000, 200.000000, 5\n" +
		"03_01_teststep, 300.000000, 300.000000, 300.000000, 5\n") {
		t.Fatalf("Report output of baseline scenario not as expected: %s", report)
	}
}
