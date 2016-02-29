package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/finklabs/GoGrinder/gogrinder"
	"github.com/finklabs/GoGrinder/http"
	"github.com/finklabs/GoGrinder/util"
)

func main() {
	// note: this test script is structured differently so the testcase can have
	// access to the main scope
	rec := []<-chan string{
		util.XmlReader("bang_0.xml", "record"), util.XmlReader("bang_1.xml", "record")}

	// initialize the GoGrinder
	gg := gogrinder.NewTest()

	// instrument teststeps
	post := gg.Teststep("01_01_xmlcowboys_post", http.PostRaw)

	// define testcases using teststeps
	xmlcowboys_01_post := func(m gogrinder.Meta, s gogrinder.Settings) {
		c := http.NewDefaultClient()
		r := rec[m.User]
		str := <-r
		//fmt.Println(str)
		base := s["server_url"].(string)
		post(m, c, base+"/post_stuff", strings.NewReader(str)) //.(http.ResponseRaw).Raw
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

	gg.AddReportPlugin(http.NewHttpMetricReporter())
	err := gogrinder.GoGrinder(gg)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
