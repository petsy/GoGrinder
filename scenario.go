package gogrinder

import (
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"sync"

	time "github.com/finklabs/ttime"
)

// Scenario builds on interfaces Config and Statistics.
type Scenario interface {
	Config
	Statistics
	Testscenario(name string, scenario interface{})
	TeststepBasic(name string, step func(Meta, ...interface{})) func(Meta, ...interface{}) interface{}
	Teststep(name string, step func(Meta, ...interface{}) (interface{}, Metric)) func(Meta, ...interface{}) interface{}
	Schedule(name string, testcase func(Meta, Settings)) error
	DoIterations(testcase func(Meta, Settings),
		iterations int, pacing float64, parallel bool)
	Run(name string, testcase func(Meta, Settings),
		delay float64, runfor float64, rampup float64, users int, pacing float64,
		settings Settings)
	Exec() error
	Thinktime(tt float64)
	Status() Status
	Stop()
	Wait()
}

type Timestamp time.Time

func (ts Timestamp) MarshalJSON() ([]byte, error) {
	t := time.Time(ts)
	if y := t.Year(); y < 0 || y >= 10000 {
		// RFC 3339 is clear that years are 4 digits exactly.
		// See golang.org/issue/4556#c15 for more discussion.
		return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
	}
	return []byte(t.Format(`"` + time.RFC3339Nano + `"`)), nil
}

type Elapsed time.Duration

func (e Elapsed) MarshalJSON() ([]byte, error) {
	// explicit marshaling of ts and elapsed!
	// from here: http://choly.ca/post/go-json-marshalling/
	return strconv.AppendFloat(nil, float64(e)/
		float64(time.Millisecond), 'f', 6, 64), nil
}

// Datatype to collect reference information about the execution of a teststep
type Meta struct {
	Testcase  string    `json:"testcase"`
	Teststep  string    `json:"teststep"`
	User      int       `json:"user"`
	Iteration int       `json:"iteration"`
	Timestamp Timestamp `json:"ts"`
	Elapsed   Elapsed   `json:"elapsed"` // elapsed time [ns]
	Error     string    `json:"error,omitempty"`
}

// TestScenario datastructure that brings all the GoGrinder functionality together.
// TestScenario supports multiple interfaces (TestConfig, TestStatistics).
type TestScenario struct {
	TestConfig // needs to be anonymous to promote access to struct field and methods
	TestStatistics
	testscenarios map[string]interface{}
	teststeps     map[string]func(Meta, ...interface{}) interface{}
	wg            sync.WaitGroup // waitgroup for teststeps
	status        Status         // status (stopped, running, stopping)
}

// Constants of internal test status.
type Status int

const (
	Stopped = iota
	Running
	Stopping
)

// Constructor takes care of initializing the TestScenario datastructure.
func NewTest() *TestScenario {
	t := TestScenario{
		testscenarios: make(map[string]interface{}),
		teststeps:     make(map[string]func(Meta, ...interface{}) interface{}),
		status:        Stopped,

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

// paceMaker is used internally. For testability it is not implemented as an internal function.
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
		if test.status != Running {
			break
		}
		time.Sleep(small)
	}
	// remaining sleep time
	if test.status == Running {
		time.Sleep(p)
	}
}

// Add a testscenario to testscenarios registry.
func (test *TestScenario) Testscenario(name string, scenario interface{}) {
	// TODO: make sure it is a function with none or single parameter!
	test.testscenarios[name] = scenario
}

// Instrument a teststep and add it to the teststeps registry.
// This implements Elapsed. For more detailed metrics please implement Teststep.
func (test *TestScenario) TeststepBasic(name string, step func(Meta, ...interface{})) func(Meta, ...interface{}) interface{} {
	its := func(meta Meta, args ...interface{}) interface{} {
		start := time.Now()
		meta.Teststep = name
		meta.Timestamp = Timestamp(start)
		step(meta, args...)

		meta.Elapsed = Elapsed(time.Now().Sub(start))
		test.Update(meta)
		return nil
	}
	test.teststeps[name] = its
	return its
}

// Instrument a teststep and add it to the teststeps registry.
// Teststeps need to return payload and Metric.
func (test *TestScenario) Teststep(name string, step func(Meta, ...interface{}) (interface{}, Metric)) func(Meta, ...interface{}) interface{} {
	its := func(meta Meta, args ...interface{}) interface{} {
		meta.Teststep = name
		result, metric := step(meta, args...)
		test.Update(metric)
		return result
	}

	test.teststeps[name] = its
	return its
}

// Schedule a testcase according to its config in the loadmodel.json config file.
func (test *TestScenario) Schedule(name string, testcase func(Meta, Settings)) error {
	delay, runfor, rampup, users, pacing, err := test.GetTestcaseConfig(name)
	settings := test.GetSettings()
	if err != nil {
		return err
	}
	test.Run(name, testcase, delay, runfor, rampup, users, pacing, settings)
	return nil
}

func (test *TestScenario) DoIterations(testcase func(Meta, Settings),
	iterations int, pacing float64, parallel bool) {
	f := func(test *TestScenario) {
		settings := test.GetSettings()
		defer test.wg.Done()

		for i := 0; i < iterations; i++ {
			start := time.Now()
			meta := Meta{Iteration: i, User: 0}
			if test.status == Stopping {
				break
			}
			testcase(meta, settings)
			if test.status == Stopping {
				break
			}
			test.paceMaker(time.Duration(pacing*float64(time.Second)), time.Now().Sub(start))
		}
	}
	if parallel {
		test.wg.Add(1)
		go f(test)
	} else {
		test.wg.Wait() // sequential processing: wait for running goroutines to finish
		test.wg.Add(1)
		f(test)
	}
}

// Run a testcase. Settings are specified in Seconds!
func (test *TestScenario) Run(name string, testcase func(Meta, Settings),
	delay float64, runfor float64, rampup float64, users int, pacing float64,
	settings Settings) {
	test.wg.Add(1) // the "Scheduler" itself is a goroutine!
	go func(test *TestScenario) {
		// ramp up the users
		defer test.wg.Done()
		time.Sleep(time.Duration(delay * float64(time.Second)))
		userStart := time.Now()

		test.wg.Add(int(users))
		for i := 0; i < users; i++ {
			// start user
			go func(nbr int) {
				defer test.wg.Done()
				time.Sleep(time.Duration(rampup * float64(time.Second)))

				for j := 0; time.Now().Sub(userStart) <
					time.Duration((runfor)*float64(time.Second)); j++ {
					// next iteration
					start := time.Now()
					meta := Meta{Testcase: name, Iteration: j, User: nbr}
					if test.status == Stopping {
						break
					}
					testcase(meta, settings)
					if test.status == Stopping {
						break
					}
					test.paceMaker(time.Duration(pacing*float64(time.Second)), time.Now().Sub(start))
				}
			}(i)
		}
	}(test)
}

// Execute the scenario set in the loadmodel.json file.
func (test *TestScenario) Exec() error {
	sel, _, _, _ := test.GetScenarioConfig()
	// check that the scenario exists
	if scenario, ok := test.testscenarios[sel]; ok {
		test.Reset()           // clear stats from previous run
		done := test.Collect() // start the collector
		test.status = Running

		fn := reflect.ValueOf(scenario)
		fnType := fn.Type()
		// some magic so we can call scenarios OR single testcases
		if fnType.Kind() == reflect.Func && fnType.NumOut() == 0 {
			if fnType.NumIn() == 0 {
				// execute the selected scenario
				fn.Call([]reflect.Value{})
			}
			if fnType.NumIn() == 2 {
				// debugging of single testcase executions
				meta := Meta{}
				settings := Settings{}
				fn.Call([]reflect.Value{reflect.ValueOf(meta),
					reflect.ValueOf(settings)},
				)
			}
			if fnType.NumIn() != 0 && fnType.NumIn() != 2 {
				return fmt.Errorf("expected a function with zero or two parameters to implement %s", sel)
			}
		} else {
			return fmt.Errorf("expected a function without return value to implement %s", sel)
		}
		// wait for testcases to finish
		// note: keep this in the foreground - do not put any of this into a goroutine!
		test.Wait()
		//test.wg.Wait()           // wait till end
		//close(test.measurements) // need to close the channel so that collect can exit, too
		<-done // wait for collector to finish
		//test.status = Stopped
	} else {
		return fmt.Errorf("scenario %s does not exist", sel)
	}
	return nil
}

// Thinktime takes ThinkTimeFactor and ThinkTimeVariance into account.
// tt is given in Seconds. So for example 3.0 equates to 3 seconds; 0.3 to 300ms.
func (test *TestScenario) Thinktime(tt float64) {
	if test.status == Running {
		_, ttf, ttv, _ := test.GetScenarioConfig()
		r := (rand.Float64() * 2.0) - 1.0 // r in [-1.0 - 1.0)
		v := float64(tt) * ttf * ((r * ttv) + 1.0) * float64(time.Second)
		time.Sleep(time.Duration(v))
	}
}

// Read the Status of the test: Running, Stopping, Stopped
func (test *TestScenario) Status() Status {
	return test.status
}

// Initiate scenario stopping.
func (test *TestScenario) Stop() {
	if test.Status() != Stopped {
		test.status = Stopping
	}
}

// Careful this is an internal exposed to ease testing.
// you need to also pull from the Collectors done channel!
func (test *TestScenario) Wait() {
	test.wg.Wait()           // wait till end
	close(test.measurements) // need to close the channel so that collect can exit, too
	test.status = Stopped
}
