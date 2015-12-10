package gogrinder

import (
    "fmt"
    "os"
    "io/ioutil"
    "github.com/xeipuuv/gojsonschema"
    "encoding/json"
)

var loadmodel map[string]interface{}

var GetTestcaseConfig func(testcase string) (int64, int64)
var GetScenarioConfig func() (string, bool, float64)


// plugable reader for the loadmodel.json file
func ReadLoadmodel(filename string) (
		func() (string, bool, float64),
		func(testcase string) (int64, int64)) {
	// this returns two get functions, so the whole ReadLoadmodel is plugable
	// using ConfigInit
	var document string
	var documentLoader gojsonschema.JSONLoader
	if len(os.Args) == 2 {
    	buf, err := ioutil.ReadFile(filename)
    	document = string(buf)
	    if err!=nil{
		    fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		    os.Exit(1)
	    }
	    documentLoader = gojsonschema.NewStringLoader(document)
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

	json.Unmarshal([]byte(document), &loadmodel)

	// fmt.Println("%v", loadmodel)
	// fmt.Printf("%v\n", y["ThinkTimeFactor"].(float64))
	// fmt.Printf("%v\n", y["Loadmodel"].([]interface{})[0].(map[string]interface{})["Users"].(float64))

	// two get functions
	gsc := func() (string, bool, float64) {
		scenario := loadmodel["Scenario"].(string)
		pacing := loadmodel["Pacing"].(bool)
		tt := loadmodel["ThinkTimeFactor"].(float64)
		return scenario, pacing, tt
	}

	gtc := func(testcase string) (int64, int64) {
	// return iterations, pacing
		conf := loadmodel["Loadmodel"].([]interface{})
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

	return gsc, gtc
}


// plugin ReadLoadmodel
func ConfigPlugin(rlm func(filename string) (
		func() (string, bool, float64), 
		func(testcase string) (int64, int64))) {
	GetScenarioConfig, GetTestcaseConfig = rlm(os.Args[1])
}
