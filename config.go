package gogrinder

import (
	"encoding/json"
	"fmt"
	"github.com/xeipuuv/gojsonschema"
	"io/ioutil"
	"os"
)

// Default schema to validate loadmodel.json files
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

// reader for the loadmodel.json file
func (test *Test) ReadLoadmodel() error {
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

// read loadmodel from document - you can provide your own schema
func (test *Test) ReadLoadmodelSchema(document string, schema string) error {
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

func (test *Test) GetScenarioConfig() (string, float64, float64) {
	scenario := test.loadmodel["Scenario"].(string)
	ttf := test.loadmodel["ThinkTimeFactor"].(float64)
	ttv := test.loadmodel["ThinkTimeVariance"].(float64)
	return scenario, ttf, ttv
}

// return iterations, pacing
func (test *Test) GetTestcaseConfig(testcase string) (int64, int64, error) {
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
