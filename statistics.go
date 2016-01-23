package gogrinder

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
	"sync"

	time "github.com/finklabs/ttime"
)

type Statistics interface {
	Update(testcase string, mm time.Duration, last time.Time)
	Collect() <-chan bool
	Reset()
	Results(since string) []Result
	Report()
}

type TestStatistics struct {
	lock          sync.RWMutex     // lock that is used on stats
	stats         map[string]stats_value            // collect and aggregate results
	//measurements  chan measurement // channel used to collect measurements from teststeps
	measurements chan meta
	reportFeature bool             // specify to print a console report
}

// internal datatype to collect information about the execution of a teststep
type meta map[string]interface{}

// Internal datastructure used on the test.measurements channel.
/*type measurement struct {
	testcase  string
	user      int
	iteration int
	last      time.Time
	value     time.Duration
	reference string
	//meta       map([string]interface{})
}*/

// Internal datastructure to collect and aggregate measurements.
type stats_value struct {
	avg   time.Duration
	min   time.Duration
	max   time.Duration
	count int64
	last  time.Time
}

// []Results is what is what you get from test.Results().
// Not sure if it is necessary to export this???
type Result struct {
	Testcase string        `json:"testcase"`
	Avg      time.Duration `json:"avg"`
	Min      time.Duration `json:"min"`
	Max      time.Duration `json:"max"`
	Count    int64         `json:"count"`
	Last     string        `json:"last"`
}

// Simple approach to sorting of the results.
// byTestcase implements sort.Interface for []Results based on the Testcase field.
type byTestcase []Result

func (a byTestcase) Len() int           { return len(a) }
func (a byTestcase) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byTestcase) Less(i, j int) bool { return a[i].Testcase < a[j].Testcase }

// Update and Collect work closely together via the measurements channel.
func (test *TestStatistics) Update(meta meta) {
	test.measurements <- meta  //measurement{testcase, user, iteration, last, mm}
}

// Collect all measurements. It blocks until measurements channel is closed.
func (test *TestStatistics) Collect() <-chan bool {
	done := make(chan bool)
	go func(test *TestStatistics) {
		for meta := range test.measurements {
			// make sure meta contains essential keys
			testcase, ok := meta["testcase"].(string); if !ok {
				panic("meta needs to contain 'testcase' key!")
			}
			elapsed, ok := meta["elapsed"].(time.Duration); if !ok {
				panic("meta needs to contain 'elapsed' key!")
			}
			last, ok := meta["last"].(time.Time); if !ok {
				panic("meta needs to contain 'last' key!")
			}
			val, exists := test.stats[testcase]
			if exists {
				val.avg = (time.Duration(val.count)*val.avg +
				elapsed) / time.Duration(val.count+1)
				if elapsed > val.max {
					val.max = elapsed
				}
				if elapsed < val.min {
					val.min = elapsed
				}
				val.last = last
				val.count++
				test.lock.Lock()
				test.stats[testcase] = val
				test.lock.Unlock()
			} else {
				// create a new statistic for t
				test.lock.Lock()
				test.stats[testcase] = stats_value{elapsed, elapsed, elapsed, 1, last}
				test.lock.Unlock()
			}
		}
		done <- true
	}(test)
	return done
}

// Reset the statistics (measurements from previous run are deleted).
func (test *TestStatistics) Reset() {
	test.lock.Lock()
	test.stats = make(map[string]stats_value)
	test.lock.Unlock()
	test.measurements = make(chan meta)
}

// Helper to convert time.Duration to ms in float64.
func d2f(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}

// Give me the stats that have been updated since <since> in ISO8601.
// In case since can not be parsed it returns all available results!
func (test *TestStatistics) Results(since string) []Result {
	test.lock.RLock()
	copy := []Result{}
	defer test.lock.RUnlock()

	s, err := time.Parse(ISO8601, since)
	all := (err != nil)
	for k, v := range test.stats {
		if all || (v.last.After(s)) {
			copy = append(copy, Result{k, v.avg, v.min, v.max, v.count, v.last.UTC().Format(ISO8601)})
		}
	}
	sort.Sort(byTestcase(copy))
	return copy
}

// Format the statistics to stdout.
func (test *TestStatistics) Report() {
	if test.reportFeature {
		res := test.Results("") // get all results
		for _, s := range res {
			fmt.Fprintf(stdout, "%s, %f, %f, %f, %d\n", s.Testcase, d2f(s.Avg),
				d2f(s.Min), d2f(s.Max), s.Count)
		}
	}
}

// Feature Toggle
func (test *TestStatistics) ReportFeature(set bool) {
	test.reportFeature = set
}

// helper to convert the field name into json-tag
func f2j(field string) string {
	f, ok := reflect.TypeOf((*Result)(nil)).Elem().FieldByName(field)
	if !ok {
		panic("Field '%s' not found in Result struct!")
	}
	return string(f.Tag.Get("json"))
}

// not completely sure implementing the io.Reader interface is the right strategy???
// https://medium.com/@mschuett/golangs-reader-interface-bd2917d5ce83#.8xfskt8ib
// implementing the Reader increments appears like overkill for this
func (test *TestStatistics) Csv() (string, error) {
	var b bytes.Buffer

	res := test.Results("") // get all results
	// write the header (using json tags)
	_, err := fmt.Fprintf(&b, "%s, %s, %s, %s, %s\n", f2j("Testcase"), f2j("Avg"),
		f2j("Min"), f2j("Max"), f2j("Count"))
	if err != nil {
		return b.String(), err
	}

	// write the lines
	for _, s := range res {
		_, err := fmt.Fprintf(&b, "%s, %f, %f, %f, %d\n", s.Testcase, d2f(s.Avg),
			d2f(s.Min), d2f(s.Max), s.Count)
		if err != nil {
			return b.String(), err
		}
	}
	return b.String(), nil
}
