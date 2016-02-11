package gogrinder

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"sort"
	"sync"

	time "github.com/finklabs/ttime"
)

type Statistics interface {
	Update(Metric)
	Collect() <-chan bool
	Reset()
	Results(since string) []Result
	Report(io.Writer)
	SetReportPlugins(reporters ...Reporter)
	AddReportPlugin(reporter Reporter)
	Csv() (string, error)
}

// Every type implements the Metric type since it is so simple.
// Only important thing is that every Metric type embeds Meta.
type Metric interface{}

type TestStatistics struct {
	lock         sync.RWMutex           // lock that is used on stats
	stats        map[string]stats_value // collect and aggregate results
	measurements chan Metric
	reporters    []Reporter
}

// Internal datastructure to collect and aggregate measurements.
type stats_value struct {
	avg   time.Duration
	min   time.Duration
	max   time.Duration
	count int64
	last  time.Time
}

// []Result is what is what you get from test.Results().
// Not sure if it is necessary to export this???
type Result struct {
	Teststep string  `json:"teststep"`
	Avg      float64 `json:"avg_ms"`
	Min      float64 `json:"min_ms"`
	Max      float64 `json:"max_ms"`
	Count    int64   `json:"count"`
	Last     string  `json:"last"`
}

// Simple approach to sorting of the results.
// byTestcase implements sort.Interface for []Results based on the Testcase field.
type byTeststep []Result

func (a byTeststep) Len() int           { return len(a) }
func (a byTeststep) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byTeststep) Less(i, j int) bool { return a[i].Teststep < a[j].Teststep }

// Update and Collect work closely together via the measurements channel.
func (test *TestStatistics) Update(m Metric) {
	test.measurements <- m
}

func (test *TestStatistics) SetReportPlugins(reporters ...Reporter) {
	test.reporters = reporters
}

func (test *TestStatistics) AddReportPlugin(reporter Reporter) {
	test.reporters = append(test.reporters, reporter)
}

// Collect all measurements. It blocks until measurements channel is closed.
func (test *TestStatistics) Collect() <-chan bool {
	done := make(chan bool)
	go func(test *TestStatistics) {
		for metric := range test.measurements {
			// call the default reporter
			test.default_reporter(metric)
			// call the plugged in reporters
			for _, reporter := range test.reporters {
				reporter.Update(metric)
			}
		}
		done <- true
	}(test)
	return done
}

// function to process the incoming measurements and update the stats
// this is also the default-reporter. All other reporters are in reporter.go
func (test *TestStatistics) default_reporter(m Metric) {
	v := reflect.ValueOf(m)
	// use of m.GetMeta() is kind of lame...
	// but reading the fields through reflection appears very wrong, too
	teststep := v.FieldByName("Teststep").Interface().(string)
	elapsed := v.FieldByName("Elapsed").Interface().(time.Duration)
	timestamp := v.FieldByName("Timestamp").Interface().(time.Time)
	test.lock.RLock()
	val, exists := test.stats[teststep]
	test.lock.RUnlock()
	if exists {
		val.avg = (time.Duration(val.count)*val.avg +
			elapsed) / time.Duration(val.count+1)
		if elapsed > val.max {
			val.max = elapsed
		}
		if elapsed < val.min {
			val.min = elapsed
		}
		val.last = timestamp
		val.count++
		test.lock.Lock()
		test.stats[teststep] = val
		test.lock.Unlock()
	} else {
		// create a new statistic for t
		test.lock.Lock()
		test.stats[teststep] = stats_value{elapsed, elapsed, elapsed, 1, timestamp}
		test.lock.Unlock()
	}
}

// Reset the statistics (measurements from previous run are deleted).
func (test *TestStatistics) Reset() {
	test.lock.Lock()
	test.stats = make(map[string]stats_value)
	test.lock.Unlock()
	test.measurements = make(chan Metric)
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
			copy = append(copy, Result{k, d2f(v.avg), d2f(v.min), d2f(v.max),
				v.count, v.last.UTC().Format(ISO8601)})
		}
	}
	sort.Sort(byTeststep(copy))
	return copy
}

// Format the statistics to stdout.
func (test *TestStatistics) Report(w io.Writer) {
	res := test.Results("") // get all results
	for _, s := range res {
		fmt.Fprintf(w, "%s, %f, %f, %f, %d\n", s.Teststep, s.Avg,
			s.Min, s.Max, s.Count)
	}
}

// helper to convert the field name into json-tag
func f2j(field string) string {
	f, ok := reflect.TypeOf((*Result)(nil)).Elem().FieldByName(field)
	if !ok {
		panic("Field '%s' not found in Result struct!")
	}
	return string(f.Tag.Get("json"))
}

func (test *TestStatistics) Csv() (string, error) {
	var b bytes.Buffer

	// write the header (using json tags)
	_, err := fmt.Fprintf(&b, "%s, %s, %s, %s, %s\n", f2j("Teststep"), f2j("Avg"),
		f2j("Min"), f2j("Max"), f2j("Count"))
	if err != nil {
		return b.String(), err
	}

	// write the lines
	test.Report(&b)
	return b.String(), nil
}
