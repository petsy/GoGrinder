package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/finklabs/GoGrinder/gogrinder"
	"github.com/finklabs/GoGrinder/req"
	"github.com/finklabs/GoGrinder/util"
)

// initialize the GoGrinder
var gg = gogrinder.NewTest()

// define testcases using teststeps
func tc1(m *gogrinder.Meta, s gogrinder.Settings) {
	var mm *req.HttpMetric
	c := req.NewDefaultClient()
	form := url.Values{}
	form.Add("username", "gogrinder")

	b := gg.NewBracket("01_01_login")
	{
		r, err := http.NewRequest("POST", "http://localhost:3001/login",
			strings.NewReader(form.Encode()))
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		if err != nil {
			m.Error += err.Error()
			mm = &req.HttpMetric{*m, 0, 0, 400}
		}
		_, _, mm = req.DoRaw(c, r, m)
	}
	b.End(mm)

	b = gg.NewBracket("01_02_get")
	{
		r, err := http.NewRequest("GET", "http://localhost:3001/get_private", nil)
		if err != nil {
			m.Error += err.Error()
			mm = &req.HttpMetric{*m, 0, 0, 400}
		}
		_, _, mm = req.DoRaw(c, r, m)
	}
	b.End(mm)
}

func tc2(m *gogrinder.Meta, s gogrinder.Settings) {
	var mm *req.HttpMetric
	c := req.NewDefaultClient()

	b := gg.NewBracket("02_01_post")
	{
		r, err := http.NewRequest("POST", "http://localhost:3001/post_stuff",
			util.NewRandReader(2000))
		if err != nil {
			m.Error += err.Error()
			mm = &req.HttpMetric{*m, 0, 0, 400}
		}
		_, _, mm = req.DoRaw(c, r, m)
	}
	b.End(mm)
}

// this is my endurance test scenario
func endurance() {
	// use the tests with the loadmodel config (json file)
	gg.Schedule("01_get", tc1)
	gg.Schedule("02_post", tc2)
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
	gg.Testscenario("01_get", tc1)
	gg.Testscenario("02_post", tc2)
}

func main() {
	err := gogrinder.GoGrinder(gg)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
