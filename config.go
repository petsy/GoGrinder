package gogrinder

import (
	"encoding/json"
	"fmt"
	"github.com/xeipuuv/gojsonschema"
	"io/ioutil"
	"os"
)

// reader for the loadmodel.json file
func (test *Test) ReadLoadmodel() {
	var filename string
	if len(os.Args) == 2 {
		filename = os.Args[1]
	} else {
		fmt.Fprintf(os.Stderr, "Error: argument for loadmodel required!\n")
		os.Exit(1)
	}

	schema := `{
        "$schema": "http://json-schema.org/draft-04/schema#",
        "description": "schema for GoGrinder loadmodel.",
        "type":"object",
        "properties":{
            "Scenario":        { "type":"string"},
            "Pacing":          { "type":"boolean"},
            "ThinkTimeFactor": { "type": "number"},
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
        "required": ["Scenario", "Pacing", "ThinkTimeFactor"],
        "additionalProperties": false
    }`

	test.ReadLoadmodelSchema(filename, schema)
}

// reader for the loadmodel.json file - you can provide your own schema
func (test *Test) ReadLoadmodelSchema(filename string, schema string) {
	var document string
	var documentLoader gojsonschema.JSONLoader
	buf, err := ioutil.ReadFile(filename)
	document = string(buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
	documentLoader = gojsonschema.NewStringLoader(document)

	schemaLoader := gojsonschema.NewStringLoader(schema)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		panic(err.Error())
	}

	if !result.Valid() {
		fmt.Printf("Error: The loadmodel is not valid :\n")
		for _, desc := range result.Errors() {
			fmt.Printf("- %s\n", desc)
		}
	}

	json.Unmarshal([]byte(document), &test.loadmodel)
}

func (test *Test) GetScenarioConfig() (string, bool, float64) {
	scenario := test.loadmodel["Scenario"].(string)
	pacing := test.loadmodel["Pacing"].(bool)
	tt := test.loadmodel["ThinkTimeFactor"].(float64)
	return scenario, pacing, tt
}

// return iterations, pacing
func (test *Test) GetTestcaseConfig(testcase string) (int64, int64) {
	conf := test.loadmodel["Loadmodel"].([]interface{})
	for _, tc := range conf {
		entry := tc.(map[string]interface{})
		if entry["Testcase"] == testcase {
			iterations := int64(entry["Iterations"].(float64))
			pacing := int64(entry["Pacing"].(float64))
			return iterations, pacing
		}
	}
	fmt.Fprintf(os.Stderr, "Error: configuration for %s not found\n", testcase)
	os.Exit(1)
	return 0, 0
}
