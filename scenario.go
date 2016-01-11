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

//var stderr io.Writer = os.Stderr

var ISO8601 = "2006-01-02T15:04:05.999Z"

type Test struct {
	loadmodel     map[string]interface{}
	testscenarios map[string]interface{}
	teststeps     map[string]func()
	lock          sync.RWMutex
	stats         stats
	wg            sync.WaitGroup
	measurements  chan measurement
	server        graceful.Server
	running       bool
}

// Constructor takes care of initializing
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

// internaly used pacemaker in nanoseconds
func paceMaker(pace time.Duration) {
	if pace < 0 {
		return
	}
	time.Sleep(pace)
}

// add a testscenario to testscenarios
func (test *Test) Testscenario(name string, scenario interface{}) {
	// TODO: make sure it is a function with none or single parameter!
	test.testscenarios[name] = scenario
}

// instrument a teststep and add it to teststeps
func (test *Test) Teststep(name string, step func()) func() {
	// TODO this should contain meta info in the report, too
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
	iterations int64, pacing int64, parallel bool) {
	meta := make(map[string]interface{})
	f := func() {
		defer test.wg.Done()

		for i := int64(0); i < iterations; i++ {
			start := time.Now()
			meta["Iteration"] = i
			meta["User"] = 0
			testcase(meta)
			paceMaker(time.Duration(pacing)*time.Millisecond - time.Now().Sub(start))
		}
	}
	// TODO: this is incomplete. !multiple! users must run in parallel.
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
		test.running = true

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
		test.running = false
		test.Report()
	} else {
		return fmt.Errorf("scenario %s does not exist", sel)
	}
	return nil
}

// this takes ThinkTimeFactor and ThinkTimeVariance into account
// thinktime is given in ms
func (test *Test) Thinktime(tt int64) {
	_, ttf, ttv := test.GetScenarioConfig()
	r := (rand.Float64() * 2.0) - 1.0 // r in [-1.0 - 1.0)
	v := float64(tt) * ttf * ((r * ttv) + 1.0) * float64(time.Millisecond)
	time.Sleep(time.Duration(v))
}
