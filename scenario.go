// Package gogrinder provides functionality for implementing and executing load & performance tests.
//
// TODO: writeup purpose of gogrinder
package gogrinder

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"reflect"
	"sync"

	"github.com/finklabs/graceful"
	time "github.com/finklabs/ttime"
)

// modify these during testing
var stdout io.Writer = os.Stdout

// this should be available in time but we keep it here for now
var ISO8601 = "2006-01-02T15:04:05.999Z"

// internal status of Test
type status int
const (
	stopped = iota
	running
	stopping
)


// Test datastructure that brings all the functionality together
//
// Test supports the following interfaces: TODO
//  statistics
//  restserver
//  configuration
type Test struct {
	loadmodel     map[string]interface{}  // datastructure to load the json loadmodel from file
	testscenarios map[string]interface{}  // testscenarios registry for testscenarios
	teststeps     map[string]func()       // registry for teststeps
	lock          sync.RWMutex            // lock that is used on stats
	stats         stats                   // collect and aggregate results
	wg            sync.WaitGroup          // waitgroup for teststeps
	measurements  chan measurement        // channel used to collect measurements from teststeps
	server        graceful.Server         // stoppable http server
	status        status                  // status (stopped, running, stopping) (used in Report())
}

// Constructor takes care of initializing the Test datastructure
func NewTest() *Test {
	t := Test{
		loadmodel:     make(map[string]interface{}),
		testscenarios: make(map[string]interface{}),
		teststeps:     make(map[string]func()),
		stats:         make(stats),
		measurements:  make(chan measurement),
	}
	return &t
}

// internally used pacemaker in nanoseconds
func paceMaker(pace time.Duration) {
	if pace < 0 {
		return
	}
	time.Sleep(pace)
}

// add a testscenario to testscenarios registry
func (test *Test) Testscenario(name string, scenario interface{}) {
	// TODO: make sure it is a function with none or single parameter!
	test.testscenarios[name] = scenario
}

// instrument a teststep and add it to the teststeps registry
func (test *Test) Teststep(name string, step func()) func() {
	// TODO this should provide meta info for the report, too
	its := func() {
		start := time.Now()
		step()
		test.update(name, time.Now().Sub(start), start)
	}
	test.teststeps[name] = its
	return its
}

// schedule a testcase according to its loadmodel config
func (test *Test) Schedule(name string, testcase func(map[string]interface{})) error {
	iterations, pacing, err := test.GetTestcaseConfig(name)
	if err != nil {
		return err
	}
	test.Run(testcase, iterations, pacing, true)
	return nil
}

// run a testcase
func (test *Test) Run(testcase func(map[string]interface{}),
	// TODO: this needs to be stoppable from the outside!
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
			paceMaker(time.Duration(pacing)*time.Millisecond - time.Now().Sub(start))
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

// execute the scenario set in the config file
func (test *Test) Exec() error {
	sel, _, _ := test.GetScenarioConfig()
	// check that the scenario exists
	if scenario, ok := test.testscenarios[sel]; ok {
		test.reset()           // clear stats from previous run
		done := test.collect() // start the collector
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

// this takes ThinkTimeFactor and ThinkTimeVariance into account
// thinktime is given in number of milliseconds. So for example 3000 equates to 3 seconds.
func (test *Test) Thinktime(tt int64) {
	_, ttf, ttv := test.GetScenarioConfig()
	r := (rand.Float64() * 2.0) - 1.0 // r in [-1.0 - 1.0)
	v := float64(tt) * ttf * ((r * ttv) + 1.0) * float64(time.Millisecond)
	time.Sleep(time.Duration(v))
}
