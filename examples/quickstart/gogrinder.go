package main

import (
	"fmt"
	"os"

	"github.com/finklabs/GoGrinder/gogrinder"
	"github.com/finklabs/GoGrinder/http"
	"github.com/finklabs/GoGrinder/util"
)

// initialize the GoGrinder
var gg = gogrinder.NewTest()

// instrument teststeps
var ts1 = gg.Teststep("01_01_teststep", http.GetRaw)
var ts2 = gg.Teststep("02_01_teststep", http.PostRaw)

// define testcases using teststeps
func tc1(m gogrinder.Meta, s gogrinder.Settings) {
	c := http.NewDefaultClient()
	ts1(m, c, "http://localhost:3001/get_stuff")
}

func tc2(m gogrinder.Meta, s gogrinder.Settings) {
	c := http.NewDefaultClient()
	ts2(m, c, "http://localhost:3001/post_stuff", util.NewRandReader(2000))
}

// this is my endurance test scenario
func endurance() {
	// use the tests with the loadmodel config (json file)
	gg.Schedule("01_testcase", tc1)
	gg.Schedule("02_testcase", tc2)
}

// this is my baseline test scenario
func baseline() {
	// use the tests with a explicit configuration
	gg.DoIterations(tc2, 5, 0, false)
}

func init() {
	// register the scenarios defined above
	gg.Testscenario("quickstart", endurance)
	gg.Testscenario("baseline", baseline)
	// register the testcases as scenarios to allow single debug execution
	gg.Testscenario("02_testcase", tc2)
}

func main() {
	err := gogrinder.GoGrinder(gg)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
