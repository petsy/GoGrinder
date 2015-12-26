package gogrinder

import (
	"fmt"
	time "github.com/finklabs/ttime"
	"sort"
)

type stats_value struct {
	avg   time.Duration
	min   time.Duration
	max   time.Duration
	count int64
}

type stats map[string]stats_value

// update the statistics with a new measurement
func (test *Test) update(testcase string, mm time.Duration) {
	val, exists := test.stats[testcase]
	if exists {
		val.avg = (time.Duration(val.count)*val.avg +
			mm) / time.Duration(val.count+1)
		if mm > val.max {
			val.max = mm
		}
		if mm < val.min {
			val.min = mm
		}
		val.count++
		test.stats[testcase] = val
	} else {
		// create a new statistic for t
		test.stats[testcase] = stats_value{mm, mm, mm, 1}
	}
}


// convert time.Duration to ms in float64
func d2f(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}


// format the statistics to stdout
func (test *Test) Report() {
	s := test.stats
	// sort the results by testcase
    keys := make([]string, 0, len(s))
    for tc := range s {
        keys = append(keys, tc)
    }
    sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintf(stdout, "%s, %f, %f, %f, %d\n", k, d2f(s[k].avg), 
			d2f(s[k].min), d2f(s[k].max), s[k].count)
	}
}
