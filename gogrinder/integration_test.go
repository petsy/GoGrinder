package gogrinder

import (
	"bytes"
	"testing"

	time "github.com/finklabs/ttime"
)

// this testsuite's aim is to cover the scope of the samples in
// github.com/finklabs/GoGrinder-samples/simple

// initialize the GoGrinder
var gg = NewTest()

var thinktime = gg.Thinktime

// define testcases using teststeps
func tc1(m *Meta, s Settings) {
	//fmt.Println(meta["Iteration"])
	b := gg.NewBracket("01_01_teststep")
	time.Sleep(50 * time.Millisecond)
	b.End(m)
	thinktime(0.050)
}

func tc2(m *Meta, s Settings) {
	b := gg.NewBracket("02_01_teststep")
	time.Sleep(100 * time.Millisecond)
	b.End(m)
	thinktime(0.100)
}

func tc3(m *Meta, s Settings) {
	b := gg.NewBracket("03_01_teststep")
	time.Sleep(150 * time.Millisecond)
	b.End(m)
	thinktime(0.150)
}

func baseline1() {
	// use the testcases with an explicit configuration
	// this baseline scenario has no concurrency so everything runs sequentially
	gg.DoIterations(tc1, 500, 0, false)
	gg.DoIterations(tc2, 500, 0, false)
	gg.DoIterations(tc3, 500, 0, false)
}

func baseline2() {
	// use the testcases with an explicit configuration
	// this "mimics" the first gogrinder scenarios
	gg.DoIterations(tc1, 18, 0.1, true)
	gg.DoIterations(tc2, 9, 0.1, true)
	gg.DoIterations(tc3, 6, 0.1, true)
}

func scenario1() {
	// use the testcases with the loadmodel config of this scenario
	gg.Schedule("01_testcase", tc1)
	gg.Schedule("02_testcase", tc2)
	gg.Schedule("03_testcase", tc3)
}

var noConcurrencyLoadmodel string = `{
	  "Scenario": "scenario1",
	  "ThinkTimeFactor": 2.00,
	  "ThinkTimeVariance": 0.0,
	  "PacingVariance": 0.0,
	  "Loadmodel": [
	    {
		  "Testcase": "01_testcase",
		  "Delay": 60.0,
		  "Runfor": 300.0,
		  "Rampup": 1.0,
		  "Users": 1,
		  "Pacing": 0.110
		}
	  ]
	}`

// Careful this bad boy messes up the fake clock big time!
// test helper to poll test status
/*
func (test *TestScenario) waitForStatus(s status) {
	for {
		if gg.status == s { break }
		time.Sleep(5 * time.Second)
	}
}
*/

// integration testcases for three modes:
// Run, Schedule and Debug testcase
func TestBaseline1(t *testing.T) {
	time.Freeze(time.Now())
	defer time.Unfreeze()
	bak := stdout
	stdout = new(bytes.Buffer)
	defer func() { stdout = bak }()

	// we do not need a full loadmodel to run the baseline scenario
	loadmodel := `{
	  "Scenario": "baseline",
	  "ThinkTimeFactor": 2.0,
	  "ThinkTimeVariance": 0.0,
	  "PacingVariance": 0.0
	}`
	//  no Loadmodel required! ,"Loadmodel": []

	// init
	gg.Testscenario("baseline", baseline1)

	// main part
	err := gg.ReadConfigValidate(loadmodel, LoadmodelSchema)
	if err != nil {
		t.Fatalf("Error while reading loadmodel config: %s!", err.Error())
	}

	start := time.Now()

	gg.Exec() // exec the scenario that has been selected in the config file
	execution := time.Now().Sub(start)

	// verify total run time of the baseline senario
	// 1500 * 50 * 2 + 500 * (100+200+300) = 450000
	if execution != 450000*time.Millisecond {
		t.Errorf("Error: execution time of baseline test not as expected: %v\n", execution)
	}

	gg.Report(stdout)
	// verify Report!
	report := stdout.(*bytes.Buffer).String()
	if report != ("01_01_teststep, 50.000000, 50.000000, 50.000000, 500, 0\n" +
		"02_01_teststep, 100.000000, 100.000000, 100.000000, 500, 0\n" +
		"03_01_teststep, 150.000000, 150.000000, 150.000000, 500, 0\n") {
		t.Fatalf("Report output of baseline scenario not as expected: %s", report)
	}
}

func TestBaseline2(t *testing.T) {
	time.Freeze(time.Now())
	defer time.Unfreeze()
	bak := stdout
	stdout = new(bytes.Buffer)
	defer func() { stdout = bak }()

	// we do not need a full loadmodel to run the baseline scenario
	loadmodel := `{
	  "Scenario": "baseline",
	  "ThinkTimeFactor": 2.0,
	  "ThinkTimeVariance": 0.1,
	  "PacingVariance": 0.0
	}`
	//  no Loadmodel required! ,"Loadmodel": []

	// init
	gg.Testscenario("baseline", baseline2)

	// main part
	err := gg.ReadConfigValidate(loadmodel, LoadmodelSchema)
	if err != nil {
		t.Fatalf("Error while reading loadmodel config: %s!", err.Error())
	}

	start := time.Now()

	gg.Exec() // exec the scenario that has been selected in the config file
	execution := time.Now().Sub(start)

	// verify total run time of the baseline senario
	// 18 * (100+100) + 90 = 3690
	//if execution <= 369000*time.Millisecond {
	if execution <= 3690*time.Millisecond {
		t.Errorf("Error: execution time of scenario1 not as expected: %v\n", execution)
	}

	gg.Report(stdout)
	// verify Report!
	report := stdout.(*bytes.Buffer).String()
	if report != ("01_01_teststep, 50.000000, 50.000000, 50.000000, 18, 0\n" +
		"02_01_teststep, 100.000000, 100.000000, 100.000000, 9, 0\n" +
		"03_01_teststep, 150.000000, 150.000000, 150.000000, 6, 0\n") {
		t.Fatalf("Report output of baseline2 scenario not as expected: %s", report)
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
	  "ThinkTimeVariance": 0.0,
	  "PacingVariance": 0.0
	}`

	// init
	gg.Reset()
	gg.Testscenario("baseline", baseline1)
	gg.Testscenario("01_testcase", tc1)

	// main part
	err := gg.ReadConfigValidate(loadmodel, LoadmodelSchema)
	if err != nil {
		t.Fatalf("Error while reading loadmodel config: %s!", err.Error())
	}

	start := time.Now()

	gg.Exec() // exec the scenario that has been selected in the config file
	execution := time.Now().Sub(start)

	// verify total run time of the baseline senario
	// 50+2*50 =150ms
	if execution != 150*time.Millisecond {
		t.Errorf("Error: execution time of debug test not as expected: %f ms.\n", d2f(execution))
	}

	gg.Report(stdout)
	// verify Report!
	report := stdout.(*bytes.Buffer).String()
	if report != "01_01_teststep, 50.000000, 50.000000, 50.000000, 1, 0\n" {
		t.Fatalf("Report output of debug test not as expected: %s", report)
	}
}

func TestAScenarioAvoidingConcurrency(t *testing.T) {
	time.Freeze(time.Now())
	defer time.Unfreeze()
	bak := stdout
	stdout = new(bytes.Buffer)
	defer func() { stdout = bak }()

	// init
	gg.Testscenario("scenario1", scenario1)

	// main part
	err := gg.ReadConfigValidate(noConcurrencyLoadmodel, LoadmodelSchema)
	if err != nil {
		t.Fatalf("Error while reading loadmodel config: %s!", err.Error())
	}

	start := time.Now()

	gg.Exec() // exec the scenario that has been selected in the config file
	execution := time.Now().Sub(start)

	// verify total run time of the baseline senario
	if execution != 360*time.Second {
		t.Errorf("Error: execution time of scenario1 not as expected: %v\n", execution)
	}

	gg.Report(stdout)
	// verify Report!
	report := stdout.(*bytes.Buffer).String()
	if report != ("01_01_teststep, 50.000000, 50.000000, 50.000000, 2000, 0\n") {
		t.Fatalf("Report output of scenario1 not as expected: %s", report)
	}
}

/*
// TODO this test is flaky - the current approach to faketime (ttime) has concurrency issues
// The concurrent executions mess up the fake clock. There is no evidence that there is a problem with GoGrinder itself.
// In real time the test runs fine (see https://github.com/finklabs/GoGrinder-samples/tree/master/simple)
// The most promising approach to fixing this problem: https://github.com/golang/go/issues/13788
func TestAScenario(t *testing.T) {
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
		  "Runfor": 1.980,
		  "Pacing": 0.110
		},
		{
		  "Testcase": "02_testcase",
		  "Users": 2,
		  "Runfor": 1.980,
		  "Pacing": 0.220
		},
		{
		  "Testcase": "03_testcase",
		  "Users": 3,
		  "Runfor": 1.980,
		  "Pacing": 0.330
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
	//if execution <= 369000*time.Millisecond {
	if execution <= 2000*time.Millisecond {
		t.Errorf("Error: execution time of scenario1 not as expected: %v\n", execution)
	}

	// verify Report!
	report := stdout.(*bytes.Buffer).String()
	if report != ("01_01_teststep, 50.000000, 50.000000, 50.000000, 18\n" +
		"02_01_teststep, 100.000000, 100.000000, 100.000000, 18\n" +
		"03_01_teststep, 150.000000, 150.000000, 150.000000, 18\n") {
		t.Fatalf("Report output of scenario1 not as expected: %s", report)
	}
}
*/

/*
// this test also has concurrency which messes up the fake clock
// two more tests like that are needed:
// * -no-frontend
// * -no-exec
func TestGoGrinder(t *testing.T) {
	//time.Freeze(time.Now())
	//defer time.Unfreeze()
	bak := stdout
	stdout = new(bytes.Buffer)
	defer func() { stdout = bak }()

	file, _ := ioutil.TempFile(os.TempDir(), "gogrinder_test")
	defer os.Remove(file.Name())
	file.WriteString(noConcurrencyLoadmodel)
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"gogrinder", file.Name()}

	// reset and run GoGrinder
	gg.Testscenario("scenario1", scenario1)
	gg.Reset()
	gg.status = stopped
	go gg.GoGrinder()
	// wait for test stopped
	gg.waitForStatus(running)
	gg.waitForStatus(stopped)
	// stop Webserver
	req, _ := http.NewRequest("DELETE", "http://localhost:3030/stop", nil)
	http.DefaultClient.Do(req)

	// verify Report!
	report := stdout.(*bytes.Buffer).String()
	if report != ("01_01_teststep, 50.000000, 50.000000, 50.000000, 2000\n") {
		t.Fatalf("Report output of scenario1 not as expected: %s", report)
	}
}
*/

// TODO
/*
func TestSettings(t *testing.T) {
	time.Freeze(time.Now())
	defer time.Unfreeze()
	bak := stdout
	stdout = new(bytes.Buffer)
	defer func() { stdout = bak }()

	test := NewTest()
	//tsts1 := test.TeststepBasic("tsts1", func(meta Meta, args ...interface{}) { time.Sleep(50 * time.Millisecond) })

	tstc1 := func(meta *Meta, settings Settings) {
		// we want to make sure, that the settings work E2E so we verify them
		// within the teststep itself!
		_, ok := settings["Awesome"]
		if !ok {
			t.Errorf("Error: expected 'settings' to contain 'Awesome'!")
		}
		if settings["Awesome"].(string) != "yeah!" {
			t.Errorf("Error: expected 'settings' to contain 'Awesome'!")
		}
		b := gg.NewBracket("tsts1")
		//tsts1(meta)
		time.Sleep(50 * time.Millisecond)
		b.End(meta)
		test.Thinktime(0.050)
	}
	test.Testscenario("fake", func() { test.DoIterations(tstc1, 500, 0, false) })

	// we do not need a full loadmodel to run the fake scenario
	loadmodel := `{
	  "Scenario": "fake",
	  "ThinkTimeFactor": 2.0,
	  "ThinkTimeVariance": 0.1,
	  "PacingVariance": 0.0,
	  "Awesome": "yeah!"
	}`

	err := test.ReadConfigValidate(loadmodel, LoadmodelSchema)
	if err != nil {
		t.Fatalf("Error while reading loadmodel config: %s!", err.Error())
	}
	test.Exec() // exec the scenario that has been selected in the config file
	// verify Report to make sure the teststep was executed

	test.Report(stdout)
	report := stdout.(*bytes.Buffer).String()
	if report != ("tsts1, 50.000000, 50.000000, 50.000000, 500, 0\n") {
		t.Fatalf("Report output of 'fake' scenario not as expected: %s", report)
	}
}
*/
