package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/finklabs/GoGrinder/gogrinder"
	"github.com/finklabs/GoGrinder/req"
	"github.com/finklabs/GoGrinder/util"
)

func main() {
	// note: this test script is structured differently so the testcase can have
	// access to the main scope
	rec := []<-chan string{
		util.XmlReader("bang_0.xml", "record"), util.XmlReader("bang_1.xml", "record")}

	// initialize the GoGrinder
	gg := gogrinder.NewTest()

	// define testcases using teststeps
	xmlcowboys_01_post := func(m *gogrinder.Meta, s gogrinder.Settings) {
		var mm *req.HttpMetric
		c := req.NewDefaultClient()
		r := rec[m.User]
		str := <-r
		//fmt.Println(str)
		base := s["server_url"].(string)
		b := gg.NewBracket("01_01_xmlcowboys_post")
		{
			r, err := http.NewRequest("POST", base+"/post_stuff",
				strings.NewReader(str))
			if err != nil {
				m.Error += err.Error()
				mm = &req.HttpMetric{*m, 0, 0, 400}
			}
			_, _, mm = req.DoRaw(c, r, m)
		}
		b.End(mm)

	}

	// this is my endurance test scenario
	endurance := func() {
		// use the tests with the loadmodel config (json file)
		gg.Schedule("xmlcowboys_01_post", xmlcowboys_01_post)
	}

	// this is my baseline test scenario
	baseline := func() {
		// use the tests with a explicit configuration
		gg.DoIterations(xmlcowboys_01_post, 5, 0, false)
	}

	// register the scenarios defined above
	gg.Testscenario("endurance", endurance)
	gg.Testscenario("baseline", baseline)
	// register the testcases as scenarios to allow single execution mode
	gg.Testscenario("xmlcowboys_01_post", xmlcowboys_01_post)

	gg.AddReportPlugin(req.NewHttpMetricReporter())
	err := gogrinder.GoGrinder(gg)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
