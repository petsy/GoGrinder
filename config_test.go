package gogrinder

import (
	"io/ioutil"
	"os"
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

	iterations, pacing, _ := fake.GetTestcaseConfig("testcase1")
	if iterations != int64(expIterations) {
		t.Errorf("Iterations %s not as expected!", iterations)
	}
	if pacing != int64(expPacing) {
		t.Errorf("Pacing %s not as expected!", pacing)
	}
}

func TestGetTestcaseConfigMissingLoadmodel(t *testing.T) {
	fake := NewTest()

	_, _, err := fake.GetTestcaseConfig("testcase1")
	error := err.Error()
	if error != "config for testcase testcase1 not found" {
		t.Error("Error handling for missing testcase config not as expected: %s!", error)
	}
}

func TestGetTestcaseConfigMissingTestcase(t *testing.T) {
	tc1 := make(map[string]interface{})
	tc1["Testcase"] = "testcase1"
	tc1["Users"] = 1
	tc1["Iterations"] = 20.0
	tc1["Pacing"] = 30.0

	fake := NewTest()
	l := make([]interface{}, 1)
	l[0] = tc1
	fake.loadmodel["Loadmodel"] = l

	_, _, err := fake.GetTestcaseConfig("testcase2")
	error := err.Error()
	if error != "config for testcase testcase2 not found" {
		t.Error("Error handling for missing testcase configuration not as expected: %s!", error)
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
	fake := NewTest()
	invalid := `{"this": "is NOT a loadmodel"}`

	err := fake.ReadLoadmodelSchema(invalid, LoadmodelSchema)

	expected := "the loadmodel is not valid:\n" +
		"- Scenario: Scenario is required\n" +
		"- ThinkTimeFactor: ThinkTimeFactor is required\n" +
		"- ThinkTimeVariance: ThinkTimeVariance is required\n" +
		"- this: Additional property this is not allowed"
	error := err.Error()
	if error != expected {
		t.Errorf("Error msg not as expected: %s", error)
	}
}

func TestReadLoadmodel(t *testing.T) {
	file, _ := ioutil.TempFile(os.TempDir(), "gogrinder_test")
	//defer os.Remove(file.Name())
	file.WriteString(loadmodel)
	t.Log(file.Name())
	// prepare empty argument see TestChangingArgs
	// in https://golang.org/src/flag/flag_test.go
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"gogrinder", file.Name()}
	t.Log(len(os.Args))
	t.Log(os.Args)

	fake := NewTest()
	fake.ReadLoadmodel()
	scenario := fake.loadmodel["Scenario"]
	if scenario != "scenario1" {
		t.Errorf("Scenario %s not 'scenario1' as expected!", scenario)
	}

	count := len(fake.loadmodel["Loadmodel"].([]interface{}))
	if count != 3 {
		t.Errorf("Expected to find 3 testcase entries in the loadmodel but found %d!", count)
	}
}

func TestReadLoadmodelErrorMissingArg(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{}
	t.Log(len(os.Args))

	fake := NewTest()
	err := fake.ReadLoadmodel()

	error := err.Error()
	if error != "argument for loadmodel required" {
		t.Errorf("Error msg for missing loadmodel argument not as expected: %s", error)
	}
}
