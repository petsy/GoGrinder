package gogrinder


import (
	"fmt"
	"sync"
)


type stats_key struct {
	sc int8
	tc int8
	ts int8
}

type stats_value struct {
	avg float64
	count int
	min float64
	max float64
}

type stats map[stats_key]stats_value

type Scenario struct {
	stats stats
	wg sync.WaitGroup
}

// Constructor takes care of initializing the stats map
//func NewScenario() *Scenario {
//	return &Scenario{stats: make(stats)}
//}
func NewScenario(name string) *Scenario {
	return &Scenario{stats: make(stats)}
}


func (scenario *Scenario) update(t stats_key, mm float64) {
	// update the statistics with the new measurement
	val, exists := scenario.stats[t]
	if exists {
		val.avg = (float64(val.count) * val.avg + 
			mm) / (float64(val.count) + 1.0)
		if mm > val.max { val.max = mm }
		if mm < val.min { val.min = mm }
		val.count++
		scenario.stats[t] = val
	} else {
		// create a new statistic for t
		scenario.stats[t] = stats_value{mm, 1, mm, mm}
	}
}


// format the statistics to stdout
func (scenario *Scenario) Report() {
	for k, v := range scenario.stats {
		fmt.Println("tc: ", k, " stats: ", v)
	}
}

// wait until everything in the waitgroup is done
func (scenario *Scenario) Wait() {
	scenario.wg.Wait()
}


// func main() {
// 	scenario := NewScenario()

// 	k := stats_key{0,0,0}
// 	scenario.update(k, 0.99)
// 	scenario.update(k, 0.99)
// 	scenario.update(k, 0.99)

// 	l := stats_key{0,0,1}
// 	scenario.update(l, 0.11)

// 	scenario.Report()
// }
