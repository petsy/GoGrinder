package gogrinder

import(
    "time"
    "os"
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

// yoah-man assemble the test runner
func TestFactory(filename string, testcase func(*os.File), iterations int, 
        pacing int64, wg *sync.WaitGroup) {
    wg.Add(1)
    go func() {
        defer wg.Done()
        f, err := os.Create(filename)
        if err != nil {
            panic(err)
        }
        defer f.Close()

        for i:=0; i<iterations; i++ {
            start := time.Now()
            testcase(f)
            paceMaker(time.Duration(pacing) * time.Millisecond - time.Now().Sub(start))
        }
    }()
}


// read load model configuration from json file
func GetScriptParams(testcase func(*os.File), wg *sync.WaitGroup) (
        string, func(*os.File), int, int64, *sync.WaitGroup) {
    name := runtime.FuncForPC(reflect.ValueOf(testcase).Pointer()).Name()
    if name == "main.testcase1" { return "./dat1.txt", testcase, 18, 120, wg }
    if name == "main.testcase2" { return "./dat2.txt", testcase, 9, 220, wg }
    return "./dat3.txt", testcase, 6, 320, wg
}


// webserver is terminated once main exits
func Webserver() {
    go func() {
        //http.Handle("/", http.FileServer(http.Dir("./static")))
        http.Handle("/", http.FileServer(rice.MustFindBox("static").HTTPBox()))
        http.ListenAndServe(":3000", nil)
    }()
}


// TODO add the API
func Restserver() {

}


// from aogaeru:
// {
//     "Configuration": {
//         "url": "http://localhost:8000",
//         "webdriver": "PhantomJS"
//     },
//     "Results": {
//         "Filename": "results.csv"
//     },
//     "Loadmodel": [
//     {
//         "Name":   "testcase01",
//         "Script":   {
//             "File": "scenarios/supercars/supercars_details.py",
//             "Class":   "SupercarsDetailsTest"
//         },
//         "Pacing":   {"Runfor": 120, "Min": 11, "Max": 13, "ThinkTimeFactor": 1.0},
//         "Schedule": {"Delay": 0, "Users": 100, "Rampup": 12}
//     }
//     ]

// }
