package main

import (
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/finklabs/GoGrinder/gogrinder"
	"github.com/finklabs/GoGrinder/http"
	"github.com/finklabs/GoGrinder/util"
)

// initialize the GoGrinder
var gg = gogrinder.NewTest()

// instrument teststeps
var ts1 = gg.Teststep("01_01_teststep", http.GetRaw)
var ts2 = gg.Teststep("02_01_teststep", http.PostRaw)

// no thinktime here
//var thinktime = gg.Thinktime

// define testcases using teststeps
func tc1(m gogrinder.Meta, s gogrinder.Settings) {
	ts1(m, "http://localhost:3001/get_stuff") // <= look here
}

func tc2(m gogrinder.Meta, s gogrinder.Settings) {
	ts2(m, "http://localhost:3001/post_stuff", util.NewRandReader(2000))
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
	gg.Testscenario("scenario1", endurance)
	gg.Testscenario("baseline", baseline)
	// register the testcases as scearios to allow single execution mode
	gg.Testscenario("02_testcase", tc2)
}

func main() {
	{
		// creating a CPU profile
		f, err := os.Create("cpu.pprof")
		if err != nil {
			fmt.Println("Error: ", err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	// ... normal main program

	//defer profile.Start(profile.CPUProfile).Stop()  // profiling
	err := gogrinder.GoGrinder(gg)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
