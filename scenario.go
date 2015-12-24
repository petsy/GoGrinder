package gogrinder

import (
	"fmt"
	"github.com/GeertJohan/go.rice"
	"net/http"
	"os"
	"reflect"
	"sync"
    "math/rand"
    time "github.com/finklabs/ttime"
)

type Test struct {
	loadmodel     map[string]interface{}
	testscenarios map[string]interface{}
	teststeps     map[string]func()
	stats         stats
	wg            sync.WaitGroup
}

// Constructor takes care of initializing
func NewTest() *Test {
	return &Test{
		loadmodel: make(map[string]interface{}),
		testscenarios: make(map[string]interface{}),
		teststeps:     make(map[string]func()),
		stats:         make(stats),
	}
}

// pacemaker in nanoseconds
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
		test.update(name, time.Now().Sub(start))
	}
	test.teststeps[name] = its
	return its
}

// schedule a testcase according to its loadmodel config
func (test *Test) Schedule(name string, testcase func(map[string]interface{})) {
	iterations, pacing := test.GetTestcaseConfig(name)
	test.Run(testcase, iterations, pacing, true)
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
	if parallel {
		test.wg.Add(1)
		go f()
	} else {
		test.wg.Wait() // wait for running goroutines to finish
		test.wg.Add(1)
		f()
	}
}

// execute the scenario set in the config file
func (test *Test) Exec() {
	sel, _, _ := test.GetScenarioConfig()
	// check that the scenario exists
	if scenario, ok := test.testscenarios[sel]; ok {
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
				panic("Error: Expected a function with zero or one parameter to implement " + sel)
			}
		} else {
			panic("Error: Expected a function without return value to implement " + sel)
		}
		test.wg.Wait() // wait till end
		test.Report()
	} else {
		fmt.Fprintf(os.Stderr, "Error: scenario %s does not exist.\n", sel)
		os.Exit(1)
	}
}


// wrapper around the time.Sleep function that can be replaced by a fake
//func (test *Test) Sleep(d time.Duration) {
//    time.Sleep(d)
//}


// 
func (test *Test) Thinktime(tt int64) {
    _, ttf, ttv := test.GetScenarioConfig()
    r := (rand.Float64() * 2.0) - 1.0  // r in [-1.0 - 1.0)
    v := float64(tt) * ttf * ((r * ttv) + 1.0) * float64(time.Millisecond)
    time.Sleep(time.Duration(v))
}


// webserver is terminated once main exits
func Webserver() {
	go func() {
		http.Handle("/", http.FileServer(rice.MustFindBox("static").HTTPBox()))
		http.ListenAndServe(":3000", nil)
	}()
}

// TODO add the API
func Restserver() {

}
