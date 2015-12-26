package gogrinder

import (
	"bytes"
	"testing"
)

var loadmodel string = `{
  "Scenario": "scenario1",
  "ThinkTimeFactor": 0.99,
  "ThinkTimeVariance": 0.1,
  "Loadmodel": [
    {
      "Testcase": "01_testcase",
      "Users": 1,
      "Iterations": 18,
      "Pacing": 120
    },
    {
      "Testcase": "02_testcase",
      "Users": 1,
      "Iterations": 9,
      "Pacing": 220
    },
    {
      "Testcase": "03_testcase",
      "Users": 1,
      "Iterations": 6,
      "Pacing": 320
    }
  ]
}`

func TestGetScenarioConfig(t *testing.T) {
	expScenario, expTtf, expTtv := "Scenario1", 0.5, 0.1
	fake := NewTest()
	fake.loadmodel["Scenario"] = expScenario
	fake.loadmodel["ThinkTimeFactor"] = expTtf
	fake.loadmodel["ThinkTimeVariance"] = expTtv

	scenario, ttf, ttv := fake.GetScenarioConfig()
	if scenario != expScenario {
		t.Errorf("Scenario %s not as expected!", scenario)
	}
	if ttf != expTtf {
		t.Errorf("ThinkTimeFactor %s not as expected!", ttf)
	}
	if ttv != expTtv {
		t.Errorf("ThinkTimeVariance %s not as expected!", ttv)
	}
}

func TestGetTestcaseConfig(t *testing.T) {
	expIterations, expPacing := 20.0, 30.0

	tc1 := make(map[string]interface{})
	tc1["Testcase"] = "testcase1"
	tc1["Users"] = 1
	tc1["Iterations"] = expIterations
	tc1["Pacing"] = expPacing

	fake := NewTest()
	l := make([]interface{}, 1)
	l[0] = tc1
	fake.loadmodel["Loadmodel"] = l

	iterations, pacing := fake.GetTestcaseConfig("testcase1")
	if iterations != int64(expIterations) {
		t.Errorf("Iterations %s not as expected!", iterations)
	}
	if pacing != int64(expPacing) {
		t.Errorf("Pacing %s not as expected!", pacing)
	}
}

func TestGetTestcaseConfigMissingLoadmodel(t *testing.T) {
	bak := stderr
	stderr = new(bytes.Buffer)
	defer func() { stderr = bak }()
	code := 0
	osexit := exit
	exit = func(c int) { code = c }
	defer func() { exit = osexit }()

	fake := NewTest()

	fake.GetTestcaseConfig("testcase1")
	if stderr.(*bytes.Buffer).String() != "Error: configuration for testcase1 not found\n" {
		t.Error("Error handling for missing testcase configuration not as expected!")
	}
	if code != 1 {
		t.Error("Exit code for missing testcase configuration not 1!")
	}
}

func TestGetTestcaseConfigMissingTestcase(t *testing.T) {
	bak := stderr
	stderr = new(bytes.Buffer)
	defer func() { stderr = bak }()
	code := 0
	osexit := exit
	exit = func(c int) { code = c }
	defer func() { exit = osexit }()

	tc1 := make(map[string]interface{})
	tc1["Testcase"] = "testcase1"
	tc1["Users"] = 1
	tc1["Iterations"] = 20.0
	tc1["Pacing"] = 30.0

	fake := NewTest()
	l := make([]interface{}, 1)
	l[0] = tc1
	fake.loadmodel["Loadmodel"] = l

	fake.GetTestcaseConfig("testcase2")
	if stderr.(*bytes.Buffer).String() != "Error: configuration for testcase2 not found\n" {
		t.Error("Error handling for missing testcase configuration not as expected!")
	}
	if code != 1 {
		t.Error("Exit code for missing testcase configuration not 1!")
	}
}

func TestReadLoadmodelSchema(t *testing.T) {
	fake := NewTest()

	fake.ReadLoadmodelSchema(loadmodel, LoadmodelSchema)

	count := len(fake.loadmodel["Loadmodel"].([]interface{}))
	if count != 3 {
		t.Errorf("Expected to find 3 testcase entries in the loadmodel but found %d!", count)
	}
}

func TestReadLoadmodelSchemaInvalid(t *testing.T) {
	bak := stderr
	stderr = new(bytes.Buffer)
	defer func() { stderr = bak }()
	code := 0
	osexit := exit
	exit = func(c int) { code = c }
	defer func() { exit = osexit }()

	fake := NewTest()
	invalid := `{"this": "is NOT a loadmodel"}`

	fake.ReadLoadmodelSchema(invalid, LoadmodelSchema)

	error := stderr.(*bytes.Buffer).String()
	expected := "Error: The loadmodel is not valid:\n" +
		"- Scenario: Scenario is required\n" +
		"- ThinkTimeFactor: ThinkTimeFactor is required\n" +
		"- ThinkTimeVariance: ThinkTimeVariance is required\n" +
		"- this: Additional property this is not allowed\n"
	if error != expected {
		t.Errorf("Error msg not as expected: %s", error)
	}
	if code != 1 {
		t.Error("Exit code for invalid loadmodel not 1!")
	}
}

func TestReadLoadmodel(t *testing.T) {
	// this is part of the integration test
}
