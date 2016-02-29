package http

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/finklabs/GoGrinder/gogrinder"
	"github.com/finklabs/GoGrinder/util"
	"github.com/finklabs/airbiscuit"
	"github.com/finklabs/graceful"
	time "github.com/finklabs/ttime"
)

// this testsuite's aim is to cover the scope of the samples in
// github.com/finklabs/GoGrinder-samples/benchmark

// initialize the GoGrinder
var gg = gogrinder.NewTest()

// instrument teststeps
var ts1 = gg.Teststep("01_01_teststep", GetRaw)
var ts2 = gg.Teststep("02_01_teststep", PostRaw)

// define testcases using teststeps
func tc1(m gogrinder.Meta, s gogrinder.Settings) {
	ts1(m, "http://localhost:3001/get_stuff")
}

func tc2(m gogrinder.Meta, s gogrinder.Settings) {
	ts2(m, "http://localhost:3001/post_stuff", util.NewRandReader(2000))
}

// endurance test scenario
func endurance() {
	gg.Schedule("01_testcase", tc1)
	gg.Schedule("02_testcase", tc2)
}

var airbiscuitLoadmodel string = `{
	  "Scenario": "scenario1",
	  "ThinkTimeFactor": 0,
	  "ThinkTimeVariance": 0,
	  "PacingVariance": 0,
	  "Loadmodel": [
		{"Pacing":0,"Runfor":0.95,"Testcase":"01_testcase","Users":10},
		{"Pacing":0,"Runfor":0.95,"Testcase":"02_testcase","Users":10}
	  ]
	}`

func TestIntegrationOfHttpPackage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	// init
	gg.Testscenario("scenario1", endurance)

	// main part
	err := gg.ReadConfigValidate(airbiscuitLoadmodel, gogrinder.LoadmodelSchema)
	if err != nil {
		t.Fatalf("Error while reading loadmodel config: %s!", err.Error())
	}

	// start the airbiscuit server
	s := &airbiscuit.Stats{Sleep: 50 * time.Millisecond}
	r := airbiscuit.Router(s)

	srv := graceful.Server{
		Timeout: 50 * time.Millisecond,
		Server: &http.Server{
			Handler: r,
			Addr:    ":3001",
		},
	}

	// stop server after wait time
	go func() {
		time.Sleep(time.Duration(1050 * int(time.Millisecond)))
		srv.Stop(80 * time.Millisecond)
		fmt.Printf("Get count: %d\n", s.G)
		fmt.Printf("Post count: %d\n", s.P)
	}()

	go srv.ListenAndServe()

	// run the test
	start := time.Now()
	gg.Exec() // exec the scenario that has been selected in the config file
	execution := time.Now().Sub(start)

	// verify total run time of the endurance scenario
	if execution > 1100*time.Millisecond {
		t.Errorf("Error: execution time of scenario1 not as expected: %v\n", execution)
	}

	results := gg.Results("")
	// check 01_01_teststep (get requests)
	if results[0].Teststep != "01_01_teststep" {
		t.Errorf("Teststep name not as expected: %s!", results[0].Teststep)
	}
	if results[0].Count < 170 {
		t.Errorf("Less than 170 get requests: %v!", results[0].Count)
	}
	if results[0].Avg < 50.0 && results[0].Avg > 62.0 {
		t.Errorf("Average not as expected: %f!", results[0].Avg)
	}
	if results[0].Min < 50.0 && results[0].Min > 62.0 {
		t.Errorf("Minimum not as expected: %f!", results[0].Min)
	}
	if results[0].Max < 50.0 && results[0].Max > 62.0 {
		t.Errorf("Maximum not as expected: %f!", results[0].Max)
	}

	// check 02_01_teststep (post requests)
	if results[1].Teststep != "02_01_teststep" {
		t.Errorf("Teststep name not as expected: %s!", results[1].Teststep)
	}
	if results[1].Count < 170 {
		t.Errorf("Less than 170 get requests: %v!", results[1].Count)
	}
	if results[1].Avg < 50.0 && results[1].Avg > 62.0 {
		t.Errorf("Average not as expected: %f!", results[1].Avg)
	}
	if results[1].Min < 50.0 && results[1].Min > 62.0 {
		t.Errorf("Minimum not as expected: %f!", results[1].Min)
	}
	if results[1].Max < 50.0 && results[1].Max > 62.0 {
		t.Errorf("Maximum not as expected: %f!", results[1].Max)
	}
}
