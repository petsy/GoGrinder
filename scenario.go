// Package gogrinder provides functionality for implementing and executing load & performance tests.
//
package gogrinder

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"reflect"
	"sync"

	time "github.com/finklabs/ttime"
)

// Modify stdout during testing.
var stdout io.Writer = os.Stdout

// ISO8601 should be available in ttime package but we keep it here for now.
var ISO8601 = "2006-01-02T15:04:05.999Z"

type Scenario interface {
	Testscenario(name string, scenario interface{})
	Teststep(name string, step func()) func()
	Schedule(name string, testcase func(map[string]interface{})) error
	Run(testcase func(map[string]interface{}), iterations int64, pacing int64, parallel bool)
	Exec() error
	Thinktime(tt int64)
}

// TestScenario datastructure that brings all the GoGrinder functionality together.
// TestScenario supports multiple interfaces (TestConfig, TestStatistics).
type TestScenario struct {
	TestConfig  // needs to be anonymous to promote access to struct field and methods
	TestStatistics
	testscenarios map[string]interface{}  // testscenarios registry for testscenarios
	teststeps     map[string]func()       // registry for teststeps
	wg            sync.WaitGroup          // waitgroup for teststeps
	status        status                  // status (stopped, running, stopping) (used in Report())
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
		teststeps:     make(map[string]func()),

		TestConfig:	TestConfig{
			loadmodel:     make(map[string]interface{}),
		},

		TestStatistics: TestStatistics{
			stats:         make(stats),
			measurements:  make(chan measurement),
		},
	}
	return &t
}

// paceMaker is used internally. It is not an internal function for testability.
// Parameter <pace> is given in nanoseconds.
func (test *TestScenario) paceMaker(pace time.Duration) {
	const small = 2 * time.Second
	if pace < 0 {
		return
	}
	// split up in small intervals so we can stop out of this
	for ;pace>small;pace=pace-small {
		if test.status != running { break }
		time.Sleep(small)
	}
	// remaining sleep time
	if test.status == running {
		time.Sleep(pace)
	}
}

// Add a testscenario to testscenarios registry.
func (test *TestScenario) Testscenario(name string, scenario interface{}) {
	// TODO: make sure it is a function with none or single parameter!
	test.testscenarios[name] = scenario
}

// Instrument a teststep and add it to the teststeps registry.
func (test *TestScenario) Teststep(name string, step func()) func() {
	// TODO this should provide meta info for the report, too
	its := func() {
		start := time.Now()
		step()
		test.Update(name, time.Now().Sub(start), start)
	}
	test.teststeps[name] = its
	return its
}

// Schedule a testcase according to its config in the loadmodel.json config file.
func (test *TestScenario) Schedule(name string, testcase func(map[string]interface{})) error {
	iterations, pacing, err := test.GetTestcaseConfig(name)
	if err != nil {
		return err
	}
	test.Run(testcase, iterations, pacing, true)
	return nil
}

// Run a testcase.
func (test *TestScenario) Run(testcase func(map[string]interface{}),
	iterations int64, pacing int64, parallel bool) {
	meta := make(map[string]interface{})
	f := func() {
		defer test.wg.Done()

		for i := int64(0); i < iterations; i++ {
			start := time.Now()
			meta["Iteration"] = i
			meta["User"] = 0
			if test.status == stopping { break }
			testcase(meta)
			if test.status == stopping { break }
			test.paceMaker(time.Duration(pacing)*time.Millisecond - time.Now().Sub(start))
		}
	}
	// TODO: this is incomplete. !multiple! users must run in parallel!
	if parallel {
		test.wg.Add(1)
		go f()
	} else {
		test.wg.Wait() // sequential processing: wait for running goroutines to finish
		test.wg.Add(1)
		f()
	}
}

// Execute the scenario set in the loadmodel.json file.
func (test *TestScenario) Exec() error {
	sel, _, _ := test.GetScenarioConfig()
	// check that the scenario exists
	if scenario, ok := test.testscenarios[sel]; ok {
		test.Reset()           // clear stats from previous run
		done := test.Collect() // start the collector
		test.status = running

		fn := reflect.ValueOf(scenario)
		fnType := fn.Type()
		// some magic so we can call scenarios OR single testcases
		if fnType.Kind() == reflect.Func && fnType.NumOut() == 0 {
			if fnType.NumIn() == 0 {
				// execute the selected scenario
				fn.Call([]reflect.Value{})
			}
			if fnType.NumIn() == 1 {
				// debugging of single testcase executions
				meta := make(map[string]interface{})
				meta["Iteration"] = 0
				meta["User"] = 0
				fn.Call([]reflect.Value{reflect.ValueOf(meta)})
			}
			if fnType.NumIn() > 1 {
				return fmt.Errorf("expected a function with zero or one parameter to implement %s", sel)
			}
		} else {
			return fmt.Errorf("expected a function without return value to implement %s", sel)
		}
		// wait for testcases to finish
		// note: keep this in the foreground - do not put any of this into a goroutine!
		test.wg.Wait()           // wait till end
		close(test.measurements) // need to close the channel so that collect can exit, too
		<-done                   // wait for collector to finish
		test.status = stopped
		test.Report()
	} else {
		return fmt.Errorf("scenario %s does not exist", sel)
	}
	return nil
}

// Thinktime takes ThinkTimeFactor and ThinkTimeVariance into account.
// tt is given in number of milliseconds. So for example 3000 equates to 3 seconds.
func (test *TestScenario) Thinktime(tt int64) {
	if test.status == running {
		_, ttf, ttv := test.GetScenarioConfig()
		r := (rand.Float64() * 2.0) - 1.0 // r in [-1.0 - 1.0)
		v := float64(tt) * ttf * ((r * ttv) + 1.0) * float64(time.Millisecond)
		time.Sleep(time.Duration(v))
	}
}
