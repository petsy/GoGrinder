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
  "PacingVariance": 0.0,
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
	expScenario, expTtf, expTtv, expPv := "Scenario1", 0.5, 0.1, 0.2
	fake := NewTest()
	fake.config["Scenario"] = expScenario
	fake.config["ThinkTimeFactor"] = expTtf
	fake.config["ThinkTimeVariance"] = expTtv
	fake.config["PacingVariance"] = expPv

	scenario, ttf, ttv, pv := fake.GetScenarioConfig()
	if scenario != expScenario {
		t.Errorf("Scenario %s not as expected!", scenario)
	}
	if ttf != expTtf {
		t.Errorf("ThinkTimeFactor %f not as expected!", ttf)
	}
	if ttv != expTtv {
		t.Errorf("ThinkTimeVariance %f not as expected!", ttv)
	}
	if pv != expPv {
		t.Errorf("PacingVariance %f not as expected!", ttv)
	}
}

func TestGetScenarioConfigUsingDefaults(t *testing.T) {
	expScenario, expTtf, expTtv, expPv := "Scenario1", 1.0, 0.0, 0.0
	fake := NewTest()
	fake.config["Scenario"] = expScenario

	scenario, ttf, ttv, pv := fake.GetScenarioConfig()
	if scenario != expScenario {
		t.Errorf("Scenario %s not as expected!", scenario)
	}
	// defaults
	if ttf != expTtf {
		t.Errorf("Default ThinkTimeFactor %f not as expected!", ttf)
	}
	if ttv != expTtv {
		t.Errorf("Default ThinkTimeVariance %f not as expected!", ttv)
	}
	if pv != expPv {
		t.Errorf("Default PacingVariance %f not as expected!", ttv)
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
	fake.config["Loadmodel"] = l

	delay, runfor, rampup, users, pacing, _ := fake.GetTestcaseConfig("testcase1")

	if delay != float64(expDelay) {
		t.Errorf("Delay %f not as expected!", delay)
	}
	if rampup != float64(expRampup) {
		t.Errorf("Rampup %f not as expected!", rampup)
	}
	if users != int(expUsers) {
		t.Errorf("Users %d not as expected!", users)
	}
	if runfor != float64(expRunfor) {
		t.Errorf("Runfor %f not as expected!", runfor)
	}
	if pacing != float64(expPacing) {
		t.Errorf("Pacing %f not as expected!", pacing)
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
	fake.config["Loadmodel"] = l

	delay, runfor, rampup, users, pacing, _ := fake.GetTestcaseConfig("testcase1")

	if users != int(expUsers) {
		t.Errorf("Users %d not as expected!", users)
	}
	if runfor != expRunfor {
		t.Errorf("Runfor %f not as expected!", runfor)
	}
	if pacing != expPacing {
		t.Errorf("Pacing %f not as expected!", pacing)
	}
	// defaults
	if delay != 0.0 {
		t.Errorf("Default Delay %f not as expected!", delay)
	}
	if rampup != 0.0 {
		t.Errorf("Default Rampup %f not as expected!", rampup)
	}
}

func TestGetTestcaseConfigMissingLoadmodel(t *testing.T) {
	fake := NewTest()

	_, _, _, _, _, err := fake.GetTestcaseConfig("testcase1")
	e := err.Error()
	if e != "config for testcase testcase1 not found" {
		t.Errorf("Error handling for missing testcase config not as expected: %s!", e)
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
	fake.config["Loadmodel"] = l

	_, _, _, _, _, err := fake.GetTestcaseConfig("testcase2")
	e := err.Error()
	if e != "config for testcase testcase2 not found" {
		t.Errorf("Error handling for missing testcase configuration not as expected: %s!", e)
	}
}

func TestReadLoadmodelSchema(t *testing.T) {
	fake := NewTest()

	fake.ReadConfigValidate(loadmodel, LoadmodelSchema)

	count := len(fake.config["Loadmodel"].([]interface{}))
	if count != 3 {
		t.Errorf("Expected to find 3 testcase entries in the loadmodel but found %d!", count)
	}
}

func TestReadLoadmodelSchemaInvalid(t *testing.T) {
	fake := NewTest()
	invalid := `{"this": "is NOT a loadmodel"}`

	err := fake.ReadConfigValidate(invalid, LoadmodelSchema)

	expected := "the loadmodel is not valid:\n" +
		"- Scenario: Scenario is required"
	e := err.Error()
	if e != expected {
		t.Errorf("Error msg not as expected: %s", e)
	}
}

func TestReadLoadmodel(t *testing.T) {
	file, _ := ioutil.TempFile(os.TempDir(), "gogrinder_test")
	defer os.Remove(file.Name())
	file.WriteString(loadmodel)
	t.Log(file.Name())

	fake := NewTest()
	fake.ReadConfig(file.Name())
	scenario := fake.config["Scenario"]
	if scenario != "scenario1" {
		t.Errorf("Scenario %s not 'scenario1' as expected!", scenario)
	}

	count := len(fake.config["Loadmodel"].([]interface{}))
	if count != 3 {
		t.Errorf("Expected to find 3 testcase entries in the loadmodel but found %d!", count)
	}
}

func TestWriteLoadmodel(t *testing.T) {
	file, _ := ioutil.TempFile(os.TempDir(), "gogrinder_test")
	defer os.Remove(file.Name())

	fake := NewTest()
	fake.config["Scenario"] = "scenario1"
	fake.config["ThinkTimeFactor"] = 2.0
	fake.config["ThinkTimeVariance"] = 0.1
	fake.filename = file.Name()

	fake.WriteConfig()

	buf, err := ioutil.ReadFile(file.Name())
	if err != nil {
		t.Errorf("Unexpected problem while reading from the file %s!", file.Name())
	}

	loadmodel := string(buf)
	if loadmodel != `{"Scenario":"scenario1","ThinkTimeFactor":2,"ThinkTimeVariance":0.1}` {
		t.Errorf("Loadmodel not as expected: %s!", loadmodel)
	}
}

func TestGetSettings(t *testing.T) {
	fake := NewTest()
	loadmodel := `{
	  "Scenario": "baseline",
	  "ThinkTimeFactor": 2.0,
	  "ThinkTimeVariance": 0.0,
	  "PacingVariance": 0.0,
	  "AdditionalProperty": 123
	}`
	fake.ReadConfigValidate(loadmodel, LoadmodelSchema)

	opts := fake.GetSettings()

	if v, ok := opts["AdditionalProperty"]; !ok || v != 123.0 {
		t.Errorf("Error: AdditionalProperty not found, or value expected %f, but was %f!", 123.0, v)
	}
	// make sure std. fields are not contained!
	if _, ok := opts["Scenario"]; ok {
		t.Errorf("Error: additional properties must not contain 'Scenario'!")
	}
	if _, ok := opts["ThinkTimeFactor"]; ok {
		t.Errorf("Error: additional properties must not contain 'ThinkTimeFactor'!")
	}
	if _, ok := opts["ThinkTimeVariance"]; ok {
		t.Errorf("Error: additional properties must not contain 'ThinkTimeVariance'!")
	}
	if _, ok := opts["PacingVariance"]; ok {
		t.Errorf("Error: additional properties must not contain 'PacingVariance'!")
	}
	if _, ok := opts["Loadmodel"]; ok {
		t.Errorf("Error: additional properties must not contain 'Loadmodel'!")
	}
}

func TestCheckTestConfigImplementsConfigInterface(t *testing.T) {
	s := &TestConfig{}
	if _, ok := interface{}(s).(Config); !ok {
		t.Errorf("TestConfig does not implement the Config interface!")
	}
}
