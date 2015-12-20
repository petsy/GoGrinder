package gogrinder

import (
	"fmt"
	"time"
)


type stats_value struct {
	avg time.Duration
	min time.Duration
	max time.Duration
	count int64
}


type stats map[string]stats_value


// update the statistics with a new measurement
func (test *Test) update(testcase string, mm time.Duration) {
	val, exists := test.stats[testcase]
	if exists {
		val.avg = (time.Duration(val.count) * val.avg + 
			mm) / time.Duration(val.count + 1)
		if mm > val.max { val.max = mm }
		if mm < val.min { val.min = mm }
		val.count++
		test.stats[testcase] = val
	} else {
		// create a new statistic for t
		test.stats[testcase] = stats_value{mm, mm, mm, 1}
	}
}


// format the statistics to stdout
func (test *Test) Report() {
	for k, v := range test.stats {
		fmt.Println(k, ",", v.avg, ",", v.min, ",", v.max, ",", v.count)
	}
}
