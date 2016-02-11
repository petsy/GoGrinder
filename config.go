package gogrinder

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/xeipuuv/gojsonschema"
	"os"
	"time"
)

type Config interface {
	ReadConfig(filename string) error
	ReadConfigValidate(document string, schema string) error
	WriteConfig() error
	GetSettings() Settings
	GetScenarioConfig() (string, float64, float64, float64)
	GetTestcaseConfig(testcase string) (float64, float64, float64, int, float64, error)
	GetConfigMap() map[string]interface{}
	GetConfigMTime() time.Time
}

type Settings map[string]interface{}

type TestConfig struct {
	config   map[string]interface{} // datastructure to hold the json config loaded from file
	filename string
	mtime    time.Time
}

// Default schema to validate loadmodel.json files.
var LoadmodelSchema string = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "description": "schema for GoGrinder loadmodel.",
    "type":"object",
    "properties": {
        "Scenario":          { "type":"string" },
        "ThinkTimeFactor":   { "type": "number" },
        "ThinkTimeVariance": { "type": "number" },
        "PacingVariance":    { "type": "number" },
        "Loadmodel": {
            "type":"array",
            "items": {
                "type":"object",
                "properties": {
                    "Testcase":   { "type": "string" },
                    "Delay":      { "type": "number" },
                    "Runfor":     { "type": "number" },
                    "Rampup":     { "type": "number" },
                    "Users":      { "type": "integer" },
                    "Pacing":     { "type": "number" }
                },
                "required": ["Testcase", "Runfor", "Users", "Pacing"],
                "additionalProperties": false
            }
        }
    },
    "required": ["Scenario"],
    "additionalProperties": true
}`

// Reader for the loadmodel.json file. Use the GoGrinder schema for loadmodel validation.
func (test *TestConfig) ReadConfig(filename string) error {
	test.filename = filename
	fi, err := os.Stat(filename)
	if err != nil {
		return err
	}
	test.mtime = fi.ModTime()

	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return test.ReadConfigValidate(string(buf), LoadmodelSchema)
}

// Read loadmodel from document - you can provide your own schema to validate the loadmodel.
func (test *TestConfig) ReadConfigValidate(document string, schema string) error {
	documentLoader := gojsonschema.NewStringLoader(document)
	schemaLoader := gojsonschema.NewStringLoader(schema)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		msg := "the loadmodel is not valid:"
		for _, desc := range result.Errors() {
			msg += fmt.Sprintf("\n- %s", desc)
		}
		return fmt.Errorf(msg)
	}

	return json.Unmarshal([]byte(document), &test.config)
}

// Write the loadmodel to file with the given filename.
func (test *TestConfig) WriteConfig() error {
	out, err := json.Marshal(test.config)
	if err != nil {
		return err
	}

	file, err := os.Create(test.filename)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(out)
	if err != nil {
		return err
	}

	fi, err := os.Stat(test.filename)
	if err != nil {
		return err
	}
	test.mtime = fi.ModTime()

	return err
}

// Return ThinkTimeFactor, ThinkTimeVariance from the loadmodel configuration.
func (test *TestConfig) GetScenarioConfig() (string, float64, float64, float64) {
	// defaults for optional properties
	ttf := 1.0
	ttv := 0.0
	pv := 0.0
	if f, ok := test.config["ThinkTimeFactor"].(float64); ok {
		ttf = f
	}
	if v, ok := test.config["ThinkTimeVariance"].(float64); ok {
		ttv = v
	}
	if p, ok := test.config["PacingVariance"].(float64); ok {
		pv = p
	}
	// required properties
	scenario := test.config["Scenario"].(string)
	return scenario, ttf, ttv, pv
}

// Return delay, runfor, rampup, users, pacing from the loadmodel configuration.
func (test *TestConfig) GetTestcaseConfig(testcase string) (float64, float64, float64, int, float64, error) {
	if conf, ok := test.config["Loadmodel"]; ok {
		if len(conf.([]interface{})) > 0 {
			for _, tc := range conf.([]interface{}) {
				entry := tc.(map[string]interface{})
				if entry["Testcase"] == testcase {
					// defaults for optional properties
					delay := 0.0
					rampup := 0.0
					if d, ok := entry["Delay"].(float64); ok {
						delay = d
					}
					if r, ok := entry["Rampup"].(float64); ok {
						rampup = r
					}
					// required properties
					runfor := entry["Runfor"].(float64)
					// Note: the JSON format itself has no integers (unlike JSON Schema). In JSON all values are float64.
					users := int(entry["Users"].(float64))
					pacing := entry["Pacing"].(float64)
					return delay, runfor, rampup, users, pacing, nil
				}
			}
		}
	}
	// configuration not found!
	return 0.0, 0.0, 0.0, 0, 0.0, fmt.Errorf("config for testcase %s not found", testcase)
}

// Return map containing additional properties from the json configuration file.
func (test *TestConfig) GetSettings() Settings {
	// defaults for optional properties
	opts := make(map[string]interface{})
	// little helper
	stdProperty := func(key string) bool {
		if key == "Scenario" {
			return true
		}
		if key == "ThinkTimeFactor" {
			return true
		}
		if key == "ThinkTimeVariance" {
			return true
		}
		if key == "PacingVariance" {
			return true
		}
		if key == "Loadmodel" {
			return true
		}
		return false
	}

	for k, v := range test.config {
		if !stdProperty(k) {
			opts[k] = v
		}
	}
	return opts
}

// Get the Json config data.
func (test *TestConfig) GetConfigMap() map[string]interface{} {
	return test.config
}

// Query the timestamp (mtime) of the config file.
func (test *TestConfig) GetConfigMTime() time.Time {
	return test.mtime
}
