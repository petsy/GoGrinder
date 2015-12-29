package gogrinder

import (
	"bytes"
	time "github.com/finklabs/ttime"
	"testing"
)

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
	// use the testcases with an explicit configuration
	// baseline scenario has no concurrency so everything runs sequentially
	gg.Run(tc1, 500, 0, false)
	gg.Run(tc2, 500, 0, false)
	gg.Run(tc3, 500, 0, false)
}

func scenario1() {
	// use the testcases with the loadmodel config of this scenario
	gg.Schedule("01_testcase", tc1)
	gg.Schedule("02_testcase", tc2)
	gg.Schedule("03_testcase", tc3)
}

// integration testcases for three modes:
// Run, Schedule and Debug testcase

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
	// 1500 * 50 * 2 + 500 * (100+200+300) = 450000
	if execution != 450000*time.Millisecond {
		t.Errorf("Error: execution time of baseline test not as expected: %v\n", execution)
	}

	// verify Report!
	report := stdout.(*bytes.Buffer).String()
	if report != ("01_01_teststep, 100.000000, 100.000000, 100.000000, 500\n" +
		"02_01_teststep, 200.000000, 200.000000, 200.000000, 500\n" +
		"03_01_teststep, 300.000000, 300.000000, 300.000000, 500\n") {
		t.Fatalf("Report output of baseline scenario not as expected: %s", report)
	}
}

func TestDebug(t *testing.T) {
	// just run a single testcase once
	time.Freeze(time.Now())
	defer time.Unfreeze()
	bak := stdout
	stdout = new(bytes.Buffer)
	defer func() { stdout = bak }()

	// we do not need a full loadmodel for this
	loadmodel := `{
	  "Scenario": "01_testcase",
	  "ThinkTimeFactor": 2.0,
	  "ThinkTimeVariance": 0.0
	}`
	//  no Loadmodel required! ,"Loadmodel": []

	// init
	gg.Testscenario("baseline", baseline)
	gg.Testscenario("01_testcase", tc1)

	// main part
	gg.ReadLoadmodelSchema(loadmodel, LoadmodelSchema)
	//gogrinder.Webserver()  // not necessary for the integration test

	start := time.Now()

	gg.Exec() // exec the scenario that has been selected in the config file
	execution := time.Now().Sub(start)

	// verify total run time of the baseline senario
	// 15 * 50 * 2 + 500 + 1000 + 1500 = 4500
	if execution != 200*time.Millisecond {
		t.Errorf("Error: execution time of debug test not as expected: %f ms.\n", d2f(execution))
	}

	// verify Report!
	report := stdout.(*bytes.Buffer).String()
	if report != "01_01_teststep, 100.000000, 100.000000, 100.000000, 1\n" {
		t.Fatalf("Report output of debug test not as expected: %s", report)
	}
}

/*
func TestScenario(t *testing.T) {
	// TODO add multiple users!
	time.Freeze(time.Now())
	defer time.Unfreeze()
	bak := stdout
	stdout = new(bytes.Buffer)
	defer func() { stdout = bak }()

	loadmodel := `{
	  "Scenario": "scenario1",
	  "ThinkTimeFactor": 2.00,
	  "ThinkTimeVariance": 0.1,
	  "Loadmodel": [
	    {
	      "Testcase": "01_testcase",
	      "Users": 1,
	      "Iterations": 1800,
	      "Pacing": 100
	    },
	    {
	      "Testcase": "02_testcase",
	      "Users": 1,
	      "Iterations": 900,
	      "Pacing": 100
	    },
	    {
	      "Testcase": "03_testcase",
	      "Users": 1,
	      "Iterations": 600,
	      "Pacing": 100
	    }
	  ]
	}`

	// init
	gg.Testscenario("scenario1", scenario1)

	// main part
	gg.ReadLoadmodelSchema(loadmodel, LoadmodelSchema)
	//gogrinder.Webserver()  // not necessary for the integration test

	start := time.Now()

	gg.Exec() // exec the scenario that has been selected in the config file
	execution := time.Now().Sub(start)

	// verify total run time of the baseline senario
	// 18 * (100+100) + 90 = 3690
	if execution <= 369000*time.Millisecond {
		t.Errorf("Error: execution time of scenario1 not as expected: %v\n", execution)
	}

	// verify Report!
	report := stdout.(*bytes.Buffer).String()
	if report != ("01_01_teststep, 100.000000, 100.000000, 100.000000, 1800\n" +
		"02_01_teststep, 200.000000, 200.000000, 200.000000, 900\n" +
		"03_01_teststep, 300.000000, 300.000000, 300.000000, 600\n") {
		t.Fatalf("Report output of baseline scenario not as expected: %s", report)
	}
}
*/
