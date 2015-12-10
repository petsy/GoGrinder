package gogrinder

import(
    "time"
    "reflect"
    "runtime"
    "sync"
    "net/http"
    "github.com/GeertJohan/go.rice"
)

// pacemaker in nanoseconds	
func paceMaker(pace time.Duration) {
    if pace < 0 { return }
    time.Sleep(pace)
}

// assemble the test runner
func TestFactory(testcase func(map[string]interface{}), meta map[string]interface{},
        wg *sync.WaitGroup) {
    wg.Add(1)
    iterations, pacing := GetTestcaseConfig(runtime.FuncForPC(reflect.ValueOf(testcase).Pointer()).Name())
    go func() {
        defer wg.Done()

        for i := int64(0); i < iterations; i++ {
            start := time.Now()
            meta["Iteration"] = i
            meta["User"] = 0
            testcase(meta)
            paceMaker(time.Duration(pacing) * time.Millisecond - time.Now().Sub(start))
        }
    }()
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

