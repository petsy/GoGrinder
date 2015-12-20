package gogrinder

import(
    "os"
    "fmt"
    "time"
    "sync"
    "net/http"
    "github.com/GeertJohan/go.rice"
)


type Test struct {
    testscenarios map[string]func()
    //testcases map[string]func(map[string]interface{})
    teststeps map[string]func()
    stats stats
    wg sync.WaitGroup
}


// Constructor takes care of initializing
func NewTest() *Test {
    return &Test{
        testscenarios: make(map[string]func()),
        //testcases: make(map[string]func(map[string]interface{})),
        teststeps: make(map[string]func()),
        stats: make(stats),
    }
}


// type Scenario struct {
//     stats stats
//     wg sync.WaitGroup
// }

// func NewScenario(name string) *Scenario {
//     return &Scenario{stats: make(stats)}
// }


// pacemaker in nanoseconds	
func paceMaker(pace time.Duration) {
    if pace < 0 { return }
    time.Sleep(pace)
}


// add a testscenario to testscenarios
func (test *Test) Testscenario(name string, scenario func()) {
    test.testscenarios[name]=scenario
}

// add a testcase to testcases
//func (test *Test) Testcase(name string, tc func(map[string]interface{})) func(map[string]interface{}) {
//    test.testcases[name]=tc
//    return tc
//}

// instrument a teststep and add it to teststeps
func (test *Test) Teststep(name string, step func()) func() {
    // TODO this should contain meta info in the report, too
    its := func() {
        start := time.Now()
        step()
        test.update(name, time.Now().Sub(start))
    }
    test.teststeps[name]=its
    return its
}



// schedule a testcase according to its loadmodel config
func (test *Test) Schedule(name string, testcase func(map[string]interface{})) {
    iterations, pacing := GetTestcaseConfig(name)
    test.Run(testcase, iterations, pacing, true)
}


// run a testcase
func (test *Test) Run(testcase func(map[string]interface{}), 
        iterations int64, pacing int64, parallel bool) {
    meta := make(map[string]interface{})
    f := func() {
        test.wg.Add(1)
        defer test.wg.Done()

        for i := int64(0); i < iterations; i++ {
            start := time.Now()
            meta["Iteration"] = i
            meta["User"] = 0
            testcase(meta)
            paceMaker(time.Duration(pacing) * time.Millisecond - time.Now().Sub(start))
        }
    }
    if parallel {
        go f() 
    } else {
        test.wg.Wait()  // wait for running goroutines to finish
        f() 
    }
}


// execute the scenario set in the config file
func (test *Test) Exec() {
    sel, _, _ := GetScenarioConfig()
    // check that the scenario exists
    if scenario, ok := test.testscenarios[sel]; ok {
        scenario() // execute the selected scenario
        test.wg.Wait()  // wait till end
        test.Report()
    } else {
        fmt.Fprintf(os.Stderr, "Error: scenario %s does not exist.\n", sel)
        os.Exit(1)
    }
}


// wait until everything in the waitgroup is done
//func (test *Test) Wait() {
//    test.wg.Wait()
//}


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
