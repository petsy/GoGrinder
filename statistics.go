package gogrinder

import (
	"fmt"
	"sort"

	time "github.com/finklabs/ttime"
)

// internal datastructure used on the test.measurements channel
type measurement struct {
	testcase string
	value    time.Duration
	last     time.Time
}

// internal datastructure to collect and aggregate measurements
type stats_value struct {
	avg   time.Duration
	min   time.Duration
	max   time.Duration
	count int64
	last  time.Time
}
type stats map[string]stats_value

// this is what is what you get from Results()
type result struct {
	Testcase string        `json:"testcase"`
	Avg      time.Duration `json:"avg"`
	Min      time.Duration `json:"min"`
	Max      time.Duration `json:"max"`
	Count    int64         `json:"count"`
	Last     string        `json:"last"`
}

// simple approach to sorting
// ByTestcase implements sort.Interface for []result based on
// the Testcase field.
type ByTestcase []result
func (a ByTestcase) Len() int           { return len(a) }
func (a ByTestcase) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTestcase) Less(i, j int) bool { return a[i].Testcase < a[j].Testcase }

// update and collect work closely together
func (test *Test) update(testcase string, mm time.Duration, last time.Time) {
	test.measurements <- measurement{testcase, mm, last}
}

// collect all measurements. it blocks until test.measurements channel is closed
func (test *Test) collect() <-chan bool {
	done := make(chan bool)
	go func(test *Test) {
		for mm := range test.measurements {
			//fmt.Println(mm)
			val, exists := test.stats[mm.testcase]
			if exists {
				val.avg = (time.Duration(val.count)*val.avg +
					mm.value) / time.Duration(val.count+1)
				if mm.value > val.max {
					val.max = mm.value
				}
				if mm.value < val.min {
					val.min = mm.value
				}
				val.last = mm.last
				val.count++
				test.lock.Lock()
				test.stats[mm.testcase] = val
				test.lock.Unlock()
			} else {
				// create a new statistic for t
				test.lock.Lock()
				test.stats[mm.testcase] = stats_value{mm.value, mm.value, mm.value, 1, mm.last}
				test.lock.Unlock()
			}
		}
		done <- true
	}(test)
	return done
}

// reset the statistics (measurements from previous run are deleted)
func (test *Test) reset() {
	test.lock.Lock()
	test.stats = make(stats)
	test.lock.Unlock()
	test.measurements = make(chan measurement)
}

// helper to convert time.Duration to ms in float64
func d2f(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}

// give mt the stats that have been updated since <since> in ISO8601
// if since can not be parsed it returns all stats!
func (test *Test) Results(since string) []result {
	test.lock.RLock()
	copy := []result{}
	defer test.lock.RUnlock()

	s, err := time.Parse(ISO8601, since)
	all := (err != nil)
	for k, v := range test.stats {
		if all || (v.last.After(s)) {
			copy = append(copy, result{k, v.avg, v.min, v.max, v.count, v.last.UTC().Format(ISO8601)})
		}
	}
	sort.Sort(ByTestcase(copy))
	return copy
}

// format the statistics to stdout
func (test *Test) Report() {
	res := test.Results("") // get all results
	for _, s := range res {
		fmt.Fprintf(stdout, "%s, %f, %f, %f, %d\n", s.Testcase, d2f(s.Avg),
			d2f(s.Min), d2f(s.Max), s.Count)
	}
}
