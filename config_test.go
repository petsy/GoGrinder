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
      "Runfor": 1.980,
      "Pacing": 120
    },
    {
      "Testcase": "02_testcase",
      "Users": 1,
      "Runfor": 1.980,
      "Pacing": 220
    },
    {
      "Testcase": "03_testcase",
      "Users": 1,
      "Runfor": 1.980,
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
	expDelay, expRampup, expUsers, expRunfor, expPacing := 10.0, 20.0, 30.0, 40.0, 50.0

	tc1 := make(map[string]interface{})
	tc1["Testcase"] = "testcase1"
	tc1["Delay"] = expDelay
	tc1["Rampup"] = expRampup
	tc1["Users"] = expUsers
	tc1["Runfor"] = expRunfor
	tc1["Pacing"] = expPacing

	fake := NewTest()
	l := make([]interface{}, 1)
	l[0] = tc1
	fake.loadmodel["Loadmodel"] = l

	delay, runfor, rampup, users, pacing, _ := fake.GetTestcaseConfig("testcase1")

	if delay != float64(expDelay) {
		t.Errorf("Delay %s not as expected!", delay)
	}
	if rampup != float64(expRampup) {
		t.Errorf("Rampup %s not as expected!", rampup)
	}
	if users != int(expUsers) {
		t.Errorf("Users %s not as expected!", users)
	}
	if runfor != float64(expRunfor) {
		t.Errorf("Runfor %s not as expected!", runfor)
	}
	if pacing != float64(expPacing) {
		t.Errorf("Pacing %s not as expected!", pacing)
	}
}

func TestGetTestcaseConfigUsingDefaults(t *testing.T) {
	// Note: the JSON format itself has no integers (unlike JSON Schema). In JSON all values are float64.
	expUsers, expRunfor, expPacing := 1.0, 20.0, 30.0

	tc1 := make(map[string]interface{})
	tc1["Testcase"] = "testcase1"
	tc1["Users"] = expUsers
	tc1["Runfor"] = expRunfor
	tc1["Pacing"] = expPacing

	fake := NewTest()
	l := make([]interface{}, 1)
	l[0] = tc1
	fake.loadmodel["Loadmodel"] = l

	delay, runfor, rampup, users, pacing, _ := fake.GetTestcaseConfig("testcase1")

	if users != int(expUsers) {
		t.Errorf("Users %s not as expected!", users)
	}
	if runfor != expRunfor {
		t.Errorf("Runfor %s not as expected!", runfor)
	}
	if pacing != expPacing {
		t.Errorf("Pacing %s not as expected!", pacing)
	}
	// defaults
	if delay != 0.0 {
		t.Errorf("Default Delay %s not as expected!", delay)
	}
	if rampup != 0.0 {
		t.Errorf("Default Rampup %s not as expected!", rampup)
	}
}

func TestGetTestcaseConfigMissingLoadmodel(t *testing.T) {
	fake := NewTest()

	_, _, _, _, _, err := fake.GetTestcaseConfig("testcase1")
	error := err.Error()
	if error != "config for testcase testcase1 not found" {
		t.Error("Error handling for missing testcase config not as expected: %s!", error)
	}
}

func TestGetTestcaseConfigMissingTestcase(t *testing.T) {
	tc1 := make(map[string]interface{})
	tc1["Testcase"] = "testcase1"
	// Note: the JSON format itself has no integers (unlike JSON Schema). In JSON all values are float64.
	tc1["Users"] = 1.0
	tc1["Iterations"] = 20.0
	tc1["Pacing"] = 30.0

	fake := NewTest()
	l := make([]interface{}, 1)
	l[0] = tc1
	fake.loadmodel["Loadmodel"] = l

	_, _, _, _, _, err := fake.GetTestcaseConfig("testcase2")
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
	defer os.Remove(file.Name())
	file.WriteString(loadmodel)
	t.Log(file.Name())

	fake := NewTest()
	fake.ReadLoadmodel(file.Name())
	scenario := fake.loadmodel["Scenario"]
	if scenario != "scenario1" {
		t.Errorf("Scenario %s not 'scenario1' as expected!", scenario)
	}

	count := len(fake.loadmodel["Loadmodel"].([]interface{}))
	if count != 3 {
		t.Errorf("Expected to find 3 testcase entries in the loadmodel but found %d!", count)
	}
}
