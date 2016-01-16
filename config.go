package gogrinder

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/xeipuuv/gojsonschema"
)

type Config interface {
	ReadLoadmodel() error
	ReadLoadmodelSchema(document string, schema string) error
	GetScenarioConfig() (string, float64, float64)
	GetTestcaseConfig(testcase string) (float64, float64, float64, int, float64, error)
}

type TestConfig struct {
	loadmodel map[string]interface{} // datastructure to hold the json loadmodel loaded from file
}

// Default schema to validate loadmodel.json files.
var LoadmodelSchema string = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "description": "schema for GoGrinder loadmodel.",
    "type":"object",
    "properties": {
        "Scenario":        { "type":"string" },
        "ThinkTimeFactor": { "type": "number" },
        "ThinkTimeVariance": { "type": "number" },
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
                "additionalProperties": true
            }
        }
    },
    "required": ["Scenario", "ThinkTimeFactor", "ThinkTimeVariance"],
    "additionalProperties": false
}`

// Reader for the loadmodel.json file. Use the GoGrinder schema for loadmodel validation.
func (test *TestConfig) ReadLoadmodel(filename string) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return test.ReadLoadmodelSchema(string(buf), LoadmodelSchema)
}

// Read loadmodel from document - you can provide your own schema to validate the loadmodel.
func (test *TestConfig) ReadLoadmodelSchema(document string, schema string) error {
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

	return json.Unmarshal([]byte(document), &test.loadmodel)
}

// Return ThinkTimeFactor, ThinkTimeVariance from the loadmodel configuration.
func (test *TestConfig) GetScenarioConfig() (string, float64, float64) {
	scenario := test.loadmodel["Scenario"].(string)
	ttf := test.loadmodel["ThinkTimeFactor"].(float64)
	ttv := test.loadmodel["ThinkTimeVariance"].(float64)
	return scenario, ttf, ttv
}

// Return delay, runfor, rampup, users, pacing from the loadmodel configuration.
func (test *TestConfig) GetTestcaseConfig(testcase string) (float64, float64, float64, int, float64, error) {
	if conf, ok := test.loadmodel["Loadmodel"]; ok {
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
