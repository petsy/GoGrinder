package main

import (
	"fmt"
	"os"

	"github.com/finklabs/GoGrinder/gogrinder"
	time "github.com/finklabs/ttime"
)

// initialize the GoGrinder
var gg = gogrinder.NewTest()

var thinktime = gg.Thinktime

// define testcases using teststeps
func tc1(m *gogrinder.Meta, s gogrinder.Settings) {
	//fmt.Println(meta["Iteration"])
	b := gg.NewBracket("01_01_teststep")
	time.Sleep(20 * time.Millisecond)
	b.End(m)

	thinktime(0.050)

	b = gg.NewBracket("01_02_teststep")
	time.Sleep(30 * time.Millisecond)
	b.End(m)
}

func tc2(m *gogrinder.Meta, s gogrinder.Settings) {
	b := gg.NewBracket("02_01_teststep")
	time.Sleep(100 * time.Millisecond)
	b.End(m)
	thinktime(0.100)
}

func tc3(m *gogrinder.Meta, s gogrinder.Settings) {
	b := gg.NewBracket("03_01_teststep")
	time.Sleep(150 * time.Millisecond)
	b.End(m)
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
