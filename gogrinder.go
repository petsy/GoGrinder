package gogrinder

import (
	"fmt"
	"io"
	"os"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/finklabs/graceful"
	time "github.com/finklabs/ttime"
)

// Modify stdout during testing.
var stdout io.Writer = os.Stdout

// ISO8601 should be available in ttime package but we keep it here for now.
var ISO8601 = "2006-01-02T15:04:05.999Z"

type Scenario interface {
	Testscenario(name string, scenario interface{})
	Teststep(name string, step func(Meta)) func(Meta)
	Schedule(name string, testcase func(Meta)) error
	DoIterations(testcase func(Meta), iterations int, pacing float64, parallel bool)
	Run(testcase func(Meta), delay float64, runfor float64, rampup float64, users int, pacing float64)
	Exec() error
	Thinktime(tt int64)
}

// TestScenario datastructure that brings all the GoGrinder functionality together.
// TestScenario supports multiple interfaces (TestConfig, TestStatistics).
type TestScenario struct {
	TestConfig // needs to be anonymous to promote access to struct field and methods
	TestStatistics
	testscenarios map[string]interface{}            // testscenarios registry for testscenarios
	teststeps     map[string]func(Meta) interface{} // registry for teststeps
	wg            sync.WaitGroup                    // waitgroup for teststeps
	status        status                            // status (stopped, running, stopping) (used in Report())
}

// Constants of internal test status.
type status int

const (
	stopped = iota
	running
	stopping
)

// Constructor takes care of initializing the TestScenario datastructure.
func NewTest() *TestScenario {
	t := TestScenario{
		testscenarios: make(map[string]interface{}),
		teststeps:     make(map[string]func(Meta) interface{}),
		status:        stopped,

		TestConfig: TestConfig{
			config: make(map[string]interface{}),
		},

		TestStatistics: TestStatistics{
			stats:        make(map[string]stats_value),
			measurements: make(chan Metric),
		},
	}
	return &t
}

// This is the "standard" gogrinder behaviour. If you need a special configuration
// or setup then maybe you should start with this code.
func (test *TestScenario) GoGrinder() error {
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
	fe, err := os.OpenFile("event-log.txt", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Error("can not open event log file: %v", err)
		// we do not need to stop in this case...
	}
	defer fe.Close()
	lr := &EventReporter{
		&log.Logger{
			Out:       fe,
			Formatter: &log.JSONFormatter{},
			Hooks:     make(log.LevelHooks),
			Level:     log.InfoLevel,
		},
	}
	test.AddReportPlugins(lr)

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
