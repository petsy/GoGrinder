package gogrinder

import(
    "time"
    "net/http"
    "github.com/GeertJohan/go.rice"
)

// pacemaker in nanoseconds	
func paceMaker(pace time.Duration) {
    if pace < 0 { return }
    time.Sleep(pace)
}

// add a testcase to the scenario
func (scenario *Scenario) Test(testcase string, tc func(map[string]interface{})) {
    meta := make(map[string]interface{})
    scenario.wg.Add(1)
    //iterations, pacing := GetTestcaseConfig(runtime.FuncForPC(reflect.ValueOf(testcase).Pointer()).Name())
    iterations, pacing := GetTestcaseConfig(testcase)
    go func() {
        defer scenario.wg.Done()

        for i := int64(0); i < iterations; i++ {
            start := time.Now()
            meta["Iteration"] = i
            meta["User"] = 0
            tc(meta)
            paceMaker(time.Duration(pacing) * time.Millisecond - time.Now().Sub(start))
        }
    }()
}

// instrumentation of a teststep
func (scenario *Scenario) Step(teststep string, step func()) func() {
    // TODO actual instrumentation
    return func() {
        step()
    }
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

