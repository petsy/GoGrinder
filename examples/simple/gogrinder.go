package main

import (
	"fmt"
	"os"
	"time"

	"github.com/finklabs/GoGrinder/gogrinder"
)

// initialize the GoGrinder
var gg = gogrinder.NewTest()

// sleep step factory
func myStep(duration time.Duration) func(m gogrinder.Meta, args ...interface{}) {
	return func(m gogrinder.Meta, args ...interface{}) {
		time.Sleep(duration * time.Millisecond)
	}
}

// instrument teststeps
var ts1 = gg.TeststepBasic("01_01_teststep", myStep(50))
var ts2 = gg.TeststepBasic("02_01_teststep", myStep(100))
var ts3 = gg.TeststepBasic("03_01_teststep", myStep(150))
var thinktime = gg.Thinktime

// define testcases using teststeps
func tc1(m gogrinder.Meta, s gogrinder.Settings) {
	//fmt.Println(meta["Iteration"])
	ts1(m)
	thinktime(0.050)
}
func tc2(m gogrinder.Meta, s gogrinder.Settings) {
	ts2(m)
	thinktime(0.100)
}
func tc3(m gogrinder.Meta, s gogrinder.Settings) {
	ts3(m)
	thinktime(0.150)
}

// this is my endurance test scenario
func endurance() {
	// use the tests with the loadmodel config (json file)
	gg.Schedule("01_testcase", tc1)
	gg.Schedule("02_testcase", tc2)
	gg.Schedule("03_testcase", tc3)
}

// this is my baseline test scenario
func baseline() {
	// use the tests with a explicit configuration
	gg.DoIterations(tc1, 5, 0, false)
	gg.DoIterations(tc2, 5, 0, false)
	gg.DoIterations(tc3, 5, 0, false)
}

func init() {
	// register the scenarios defined above
	gg.Testscenario("endurance", endurance)
	gg.Testscenario("baseline", baseline)
	// register the testcases as scearios to allow single execution mode
	gg.Testscenario("01_testcase", tc1)
	gg.Testscenario("02_testcase", tc2)
	gg.Testscenario("03_testcase", tc3)
}

func main() {
	err := gogrinder.GoGrinder(gg)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
