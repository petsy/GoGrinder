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
func (test *Test) ReadLoadmodel() {
	var filename string
	if len(os.Args) == 2 {
		filename = os.Args[1]
	} else {
		fmt.Fprintf(stderr, "Error: argument for loadmodel required!\n")
		exit(1)
	}
	buf, err := ioutil.ReadFile(filename)
	document := string(buf)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %s\n", err)
		exit(1)
	}

	test.ReadLoadmodelSchema(document, LoadmodelSchema)
}

// read loadmodel from document - you can provide your own schema
func (test *Test) ReadLoadmodelSchema(document string, schema string) {
	//var documentLoader gojsonschema.JSONLoader
	documentLoader := gojsonschema.NewStringLoader(document)
	schemaLoader := gojsonschema.NewStringLoader(schema)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %s\n", err)
		exit(1)
	}

	if !result.Valid() {
		fmt.Fprintf(stderr, "Error: The loadmodel is not valid:\n")
		for _, desc := range result.Errors() {
			fmt.Fprintf(stderr, "- %s\n", desc)
		}
		exit(1)
	}

	json.Unmarshal([]byte(document), &test.loadmodel)
}

func (test *Test) GetScenarioConfig() (string, float64, float64) {
	scenario := test.loadmodel["Scenario"].(string)
	ttf := test.loadmodel["ThinkTimeFactor"].(float64)
	ttv := test.loadmodel["ThinkTimeVariance"].(float64)
	return scenario, ttf, ttv
}

// return iterations, pacing
func (test *Test) GetTestcaseConfig(testcase string) (int64, int64) {
	if conf, ok := test.loadmodel["Loadmodel"]; ok {
		if len(conf.([]interface{})) > 0 {
			for _, tc := range conf.([]interface{}) {
				entry := tc.(map[string]interface{})
				if entry["Testcase"] == testcase {
					iterations := int64(entry["Iterations"].(float64))
					pacing := int64(entry["Pacing"].(float64))
					return iterations, pacing
				}
			}
		}
	}
	// configuration not found!
	fmt.Fprintf(stderr, "Error: configuration for %s not found\n", testcase)
	exit(1)
	return 0, 0
}
