package gogrinder

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/xeipuuv/gojsonschema"
)


type Config interface {
	ReadLoadmodel() error
	ReadLoadmodelSchema(document string, schema string) error
	GetScenarioConfig() (string, float64, float64)
	GetTestcaseConfig(testcase string) (int64, int64, error)
}

type TestConfig struct {
	loadmodel     map[string]interface{}  // datastructure to hold the json loadmodel loaded from file
}

// Default schema to validate loadmodel.json files.
var LoadmodelSchema string = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "description": "schema for GoGrinder loadmodel.",
    "type":"object",
    "properties":{
        "Scenario":        { "type":"string"},
        "ThinkTimeFactor": { "type": "number"},
        "ThinkTimeVariance": { "type": "number"},
        "Loadmodel": {
            "type":"array",
            "items": {
                "type":"object",
                "properties":{
                    "Testcase":   { "type": "string"},
                    "Users":      { "type": "integer"},
                    "Iterations": { "type": "integer"},
                    "Pacing":     { "type": "integer"}
                },
                "required": ["Testcase", "Users", "Iterations", "Pacing"],
                "additionalProperties": true
            }
        }
    },
    "required": ["Scenario", "ThinkTimeFactor", "ThinkTimeVariance"],
    "additionalProperties": false
}`

// Reader for the loadmodel.json file. Use the GoGrinder schema for loadmodel validation.
func (test *TestConfig) ReadLoadmodel() error {
	var filename string
	if len(os.Args) == 2 {
		filename = os.Args[1]
	} else {
		return fmt.Errorf("argument for loadmodel required")
	}
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

// Return iterations, pacing from the loadmodel configuration.
func (test *TestConfig) GetTestcaseConfig(testcase string) (int64, int64, error) {
	if conf, ok := test.loadmodel["Loadmodel"]; ok {
		if len(conf.([]interface{})) > 0 {
			for _, tc := range conf.([]interface{}) {
				entry := tc.(map[string]interface{})
				if entry["Testcase"] == testcase {
					iterations := int64(entry["Iterations"].(float64))
					pacing := int64(entry["Pacing"].(float64))
					return iterations, pacing, nil
				}
			}
		}
	}
	// configuration not found!
	return 0, 0, fmt.Errorf("config for testcase %s not found", testcase)
}
