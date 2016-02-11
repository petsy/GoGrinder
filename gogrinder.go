// Package gogrinder provides tools for implementing and executing
// load & performance tests.
//
package gogrinder

import (
	"fmt"
	"io"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/finklabs/graceful"
	time "github.com/finklabs/ttime"
)

// Modify stdout during testing.
var stdout io.Writer = os.Stdout

// ISO8601 should be available in ttime package but we keep it here for now.
var ISO8601 = "2006-01-02T15:04:05.999Z"

// This is the "standard" gogrinder behaviour. If you need a special configuration
// or setup then maybe you should start with this code.
func GoGrinder(test Scenario) error {
	var err error
	filename, noExec, noReport, noFrontend, noPrometheus, port, logLevel, err := GetCLI()
	if err != nil {
		return err
	}
	ll, _ := log.ParseLevel(logLevel)
	log.SetLevel(ll)
	err = test.ReadConfig(filename)
	if err != nil {
		return err
	}

	// prepare reporter plugins
	// initialize the event logger
	fe, err := os.OpenFile("event-log.txt", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		log.Error("can not open event log file: %v", err)
		// we do not need to stop in this case...
	}
	defer fe.Close()
	test.AddReportPlugin(&EventReporter{fe})

	exec := func() {
		err = test.Exec()
		if !noReport {
			test.Report(stdout)
		}
	}

	frontend := func() {
		srv := NewTestServer(test)
		srv.Addr = fmt.Sprintf(":%d", port)
		err = srv.ListenAndServe()
	}

	// prometheus reporter needs to "wrap" all test executions
	var srv *graceful.Server
	if !noPrometheus {
		srv = NewPrometheusReporterServer()
		srv.Addr = fmt.Sprintf(":%d", 9110)
		go srv.ListenAndServe()
		// if for example the port is in use we continue...
	}

	// handle the different run modes
	// invalid mode of noExec && noFrontend is handled in cli.go
	if noExec {
		frontend()
	}
	if noFrontend {
		exec()
	}
	if !noExec && !noFrontend {
		// this is the "normal" case - webserver is blocking
		go exec()
		frontend()
	}

	// run for another +2 * scrape_interval so we read all metrics in
	if !noPrometheus {
		time.Sleep(11 * time.Second)
		srv.Stop(1 * time.Second)
	}

	return err
}
