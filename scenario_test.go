package gogrinder

import(
	"testing"
    time "github.com/finklabs/ttime"
)


func TestThinktimeNoVariance(t *testing.T) {
	// create a fake loadmodel for testing
	var fake = NewTest()
	fake.loadmodel["Scenario"]="scenario1"
	fake.loadmodel["ThinkTimeFactor"]=1.0
	fake.loadmodel["ThinkTimeVariance"]=0.0

	time.Freeze(time.Now())  // prepare for testing
	defer time.Unfreeze()

	start := time.Now()
	fake.Thinktime(20)
	sleep := time.Now().Sub(start)
	if sleep != 20*time.Millisecond {
		t.Error("Expected to sleep for 20ms but something went wrong!")
	}
}

func TestThinktimeVariance(t *testing.T) {
	// create a fake loadmodel for testing
	var fake = NewTest()
	fake.loadmodel["Scenario"]="scenario1"
	fake.loadmodel["ThinkTimeFactor"]=2.0
	fake.loadmodel["ThinkTimeVariance"]=0.1

	min, max, avg := 20.0, 20.0, 0.0
	time.Freeze(time.Now())  // prepare for testing
	defer time.Unfreeze()

	for i := 0; i < 1000; i++ {
		start := time.Now()
		fake.Thinktime(10)
		sleep := float64(time.Now().Sub(start)) / float64(time.Millisecond)
		if sleep < min { min = sleep }
		if max < sleep { max = sleep }
		avg += sleep
	}
	avg = avg/1000
	if min < 18.0 { t.Errorf("Minimum sleep time %f out of defined range!\n", min) }
	if max >= 22.0 { t.Errorf("Maximum sleep time %f out of defined range!", max) }
	t.Logf("Minimum sleep time %f\n", min)
	t.Logf("Maximum sleep time %f\n", max)
	t.Logf("Average sleep time %f\n", avg)
	if avg < 19.9 || avg > 20.1 { t.Fatalf("Average sleep time %f out of defined range!", avg) }
}
