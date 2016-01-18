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
	DoIterations(testcase func(map[string]interface{}), iterations int, pacing float64, parallel bool)
	Run(testcase func(map[string]interface{}), delay float64, runfor float64, rampup float64, users int, pacing float64)
	Exec() error
	Thinktime(tt int64)
}

// TestScenario datastructure that brings all the GoGrinder functionality together.
// TestScenario supports multiple interfaces (TestConfig, TestStatistics).
type TestScenario struct {
	TestConfig // needs to be anonymous to promote access to struct field and methods
	TestStatistics
	testscenarios map[string]interface{} // testscenarios registry for testscenarios
	teststeps     map[string]func()      // registry for teststeps
	wg            sync.WaitGroup         // waitgroup for teststeps
	status        status                 // status (stopped, running, stopping) (used in Report())
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
		status:        stopped,

		TestConfig: TestConfig{
			loadmodel: make(map[string]interface{}),
		},

		TestStatistics: TestStatistics{
			stats:         make(stats),
			measurements:  make(chan measurement),
			reportFeature: true,
		},
	}
	return &t
}

// paceMaker is used internally. It is not an internal function for testability.
// Parameter <pace> is given in nanoseconds.
func (test *TestScenario) paceMaker(pacing time.Duration, elapsed time.Duration) {
	_, _, _, pv := test.GetScenarioConfig()
	const small = 2 * time.Second

	// calculate the variable pacing
	r := (rand.Float64() * 2.0) - 1.0 // r in [-1.0 - 1.0)
	v := float64(pacing) * ((r * pv) + 1.0)
	p := time.Duration(v - float64(elapsed))
	if p < 0 {
		return
	}

	// split up in small intervals so we can stop out of this
	for ; p > small; p = p - small {
		if test.status != running {
			break
		}
		time.Sleep(small)
	}
	// remaining sleep time
	if test.status == running {
		time.Sleep(p)
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
	delay, runfor, rampup, users, pacing, err := test.GetTestcaseConfig(name)
	if err != nil {
		return err
	}
	test.Run(testcase, delay, runfor, rampup, users, pacing)
	return nil
}

func (test *TestScenario) DoIterations(testcase func(map[string]interface{}),
	iterations int, pacing float64, parallel bool) {
	meta := make(map[string]interface{})
	f := func() {
		defer test.wg.Done()

		for i := 0; i < iterations; i++ {
			start := time.Now()
			meta["Iteration"] = i
			meta["User"] = 0
			if test.status == stopping {
				break
			}
			testcase(meta)
			if test.status == stopping {
				break
			}
			test.paceMaker(time.Duration(pacing*float64(time.Second)), time.Now().Sub(start))
		}
	}
	if parallel {
		test.wg.Add(1)
		go f()
	} else {
		test.wg.Wait() // sequential processing: wait for running goroutines to finish
		test.wg.Add(1)
		f()
	}
}

// Run a testcase. Settings are specified in Seconds!
func (test *TestScenario) Run(testcase func(map[string]interface{}), delay float64, runfor float64, rampup float64,
	users int, pacing float64) {
	test.wg.Add(1) // the "Scheduler" itself is a goroutine!
	go func() {
		// ramp up the users
		defer test.wg.Done()
		time.Sleep(time.Duration(delay * float64(time.Second)))
		userStart := time.Now()

		test.wg.Add(int(users))
		for i := 0; i < users; i++ {
			// start user
			go func() {
				defer test.wg.Done()
				time.Sleep(time.Duration(rampup * float64(time.Second)))

				for j := 0; time.Now().Sub(userStart) < time.Duration((runfor)*float64(time.Second)); j++ {
					// next iteration
					start := time.Now()
					meta := make(map[string]interface{})
					meta["User"] = i
					meta["Iteration"] = j
					if test.status == stopping {
						break
					}
					testcase(meta)
					if test.status == stopping {
						break
					}
					test.paceMaker(time.Duration(pacing*float64(time.Second)), time.Now().Sub(start))
				}
			}()
		}
	}()
}

// Execute the scenario set in the loadmodel.json file.
func (test *TestScenario) Exec() error {
	sel, _, _, _ := test.GetScenarioConfig()
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
		test.Report()
		test.status = stopped
	} else {
		return fmt.Errorf("scenario %s does not exist", sel)
	}
	return nil
}

// Thinktime takes ThinkTimeFactor and ThinkTimeVariance into account.
// tt is given in Seconds. So for example 3.0 equates to 3 seconds; 0.3 to 300ms.
func (test *TestScenario) Thinktime(tt float64) {
	if test.status == running {
		_, ttf, ttv, _ := test.GetScenarioConfig()
		r := (rand.Float64() * 2.0) - 1.0 // r in [-1.0 - 1.0)
		v := float64(tt) * ttf * ((r * ttv) + 1.0) * float64(time.Second)
		time.Sleep(time.Duration(v))
	}
}

// This is the "standard" behaviour. If you need a special configuration maybe you can start with this code.
func (test *TestScenario) GoGrinder() {
	filename, noExec, noReport, noFrontend, port, err := GetCLI()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	test.ReportFeature(!noReport)
	test.ReadConfig(filename)

	exec := func() {
		err := test.Exec() // exec the scenario that has been selected in the config file
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	frontend := func() {
		srv := NewTestServer(test)
		srv.Addr = fmt.Sprintf(":%d", port)
		srv.ListenAndServe()
	}

	// handle the different run modes
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
}
