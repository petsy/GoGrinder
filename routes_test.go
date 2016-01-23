package gogrinder

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	time "github.com/finklabs/ttime"
)

// TODO func TestHandlerServeHTTPInvalidJson(t *testing.T) ?

//func TestGetLoadmodel(t *testing.T) {
//	// use since query with ISO8601 datetime
//	//request, _ := http.NewRequest("GET", "/loadmodel", nil)
//	//response := httptest.NewRecorder()
//
//	var fake = NewTest()
//	response, err := fake.getLoadmodel(nil, nil)
//
//	if err != nil {
//		t.Fatalf("Error while processing: %s", err)
//	}
//
//	//body := response.Body
//	b := book{"Ender's Game", "Orson Scott Card", 1}
//	if response.(book) != b {
//		t.Fatalf("Response not as expected: %v", response)
//	}
//}

////////////////////////////////
// test routes
////////////////////////////////
func TestRouteGetCsv(t *testing.T) {
	// test with 3 measurements
	fake := NewTest()
	srv := TestServer{}
	srv.test = fake
	// put 3 measurements into the fake server
	done := fake.Collect() // this needs a collector to unblock update
	now := time.Now().UTC()
	fake.Update(meta{"testcase":"sth", "elapsed":8*time.Millisecond, "last":now})
	fake.Update(meta{"testcase":"sth", "elapsed":10*time.Millisecond, "last":now})
	fake.Update(meta{"testcase":"sth", "elapsed":2*time.Millisecond, "last":now})
	close(fake.measurements)
	<-done

	// invoke REST service
	req, _ := http.NewRequest("GET", "/csv", nil)
	rsp := httptest.NewRecorder()
	// I separated the Router() from the actual Webserver()
	// In this way I can test routes without running a server
	srv.Router().ServeHTTP(rsp, req)
	if rsp.Code != http.StatusOK {
		t.Fatalf("Status code expected: %v but was: %v", http.StatusOK, rsp.Code)
	}

	body := rsp.Body.String()
	if body != `"testcase, avg, min, max, count\nsth, 6.666666, 2.000000, 10.000000, 3\n"` {
		t.Fatalf("Response not as expected: %s", body)
	}
}

func TestRouteGetStatistics(t *testing.T) {
	// test with 3 measurements
	fake := NewTest()
	srv := TestServer{}
	srv.test = fake
	// put 3 measurements into the fake server
	done := fake.Collect() // this needs a collector to unblock update
	now := time.Now().UTC()
	fake.Update(meta{"testcase":"sth", "elapsed":8*time.Millisecond, "last":now})
	fake.Update(meta{"testcase":"sth", "elapsed":10*time.Millisecond, "last":now})
	fake.Update(meta{"testcase":"sth", "elapsed":2*time.Millisecond, "last":now})
	close(fake.measurements)
	<-done

	// invoke REST service
	req, _ := http.NewRequest("GET", "/statistics", nil)
	rsp := httptest.NewRecorder()
	// I separated the Router() from the actual Webserver()
	// In this way I can test routes without running a server
	srv.Router().ServeHTTP(rsp, req)
	if rsp.Code != http.StatusOK {
		t.Fatalf("Status code expected: %s but was: %v", "200", rsp.Code)
	}

	body := rsp.Body.String()
	if body != fmt.Sprintf(`{"results":[{"testcase":"sth","avg":6666666,"min":2000000,`+
		`"max":10000000,"count":3,"last":"%s"}],"running":false}`, now.Format(ISO8601)) {
		t.Fatalf("Response not as expected: %s", body)
	}
}

// TODO test the routes as well
func TestHandlerStatisticsWithQuery(t *testing.T) {
	// test with 3 measurements (two stats)
	var fake = NewTest()
	done := fake.Collect() // this needs a collector to unblock update
	t1 := time.Now().UTC()
	fake.Update(meta{"testcase":"sth", "elapsed":8*time.Millisecond, "last":t1})
	time.Sleep(5 * time.Millisecond)
	t2 := t1.Add(2 * time.Millisecond)
	fake.Update(meta{"testcase":"else", "elapsed":10*time.Millisecond, "last":t1})
	fake.Update(meta{"testcase":"else", "elapsed":2*time.Millisecond, "last":t2})
	t3 := t2.Add(2 * time.Millisecond)
	close(fake.measurements)
	<-done

	// invoke REST service for stats update
	//iso8601 := "2006-01-02T15:04:05.999Z"
	ts2 := t2.Format(ISO8601)
	request, _ := http.NewRequest("GET", "/statistics?since="+ts2, nil)
	srv := TestServer{}
	srv.test = fake
	response, err := srv.getStatistics(request)
	if err != nil {
		t.Fatalf("Error while processing: %s", err.Message)
	}
	if len(response.(map[string]interface{})["results"].([]Result)) != 1 {
		t.Fatalf("Response should contain exactly 1 row.")
	}
	if response.(map[string]interface{})["results"].([]Result)[0] !=
		(Result{"else", 6000000, 2000000, 10000000, 2, t2.Format(ISO8601)}) {
		t.Fatalf("Response not as expected: %v", response.([]Result)[0])
	}

	// update but no new data
	ts3 := t3.Format(ISO8601)
	request, _ = http.NewRequest("GET", "/statistics?since="+ts3, nil)
	response, err = srv.getStatistics(request)

	if len(response.(map[string]interface{})["results"].([]Result)) != 0 {
		t.Fatalf("Response should contain 0 rows.")
	}

	// get all rows
	request, _ = http.NewRequest("GET", "/statistics", nil)
	response, err = srv.getStatistics(request)

	if err != nil {
		t.Fatalf("Error while processing: %s", err.Message)
	}
	if len(response.(map[string]interface{})["results"].([]Result)) != 2 {
		t.Fatalf("Response should contain exactly 2 rows.")
	}
	// "else" is [0]
	if response.(map[string]interface{})["results"].([]Result)[0] !=
		(Result{"else", 6000000, 2000000, 10000000, 2, t2.Format(ISO8601)}) {
		t.Log(t2.Format(ISO8601))
		t.Logf("Response 0: %v", response.(map[string]interface{})["results"].([]Result)[0])
		t.Logf("Response 1: %v", response.(map[string]interface{})["results"].([]Result)[1])
		t.Fatalf("Response not as expected: %v", response.(map[string]interface{})["results"].([]Result)[0])
	}
	// "sth" is [1]
	if response.(map[string]interface{})["results"].([]Result)[1] !=
		(Result{"sth", 8000000, 8000000, 8000000, 1, t1.Format(ISO8601)}) {
		t.Log(t1)
		t.Fatalf("Response not as expected: %v", response.(map[string]interface{})["results"].([]Result)[1])
	}
}

func TestRouteStartStop(t *testing.T) {
	// prepare
	time.Freeze(time.Now())
	defer time.Unfreeze()
	srv := TestServer{}
	srv.test = NewTest()
	tc1 := func(meta meta) { srv.test.Thinktime(0.050) }
	srv.test.Testscenario("fake", func() { gg.DoIterations(tc1, 500, 0, false) })
	loadmodel := `{"Scenario": "fake", "ThinkTimeFactor": 2.0, "ThinkTimeVariance": 0.0	}`
	srv.test.ReadConfigValidate(loadmodel, LoadmodelSchema)

	{
		// startTest
		req, _ := http.NewRequest("POST", "/test", nil)
		rsp := httptest.NewRecorder()
		srv.Router().ServeHTTP(rsp, req)
		if rsp.Code != http.StatusOK {
			t.Fatalf("Status code expected: %s but was: %v", "200", rsp.Code)
		}
	}
	// another fake clock problem here!
	//	if srv.test.status != running {
	//		t.Fatalf("Status code expected: %v but was: %v", running, srv.test.status)
	//	}

	{
		// stopTest
		req, _ := http.NewRequest("DELETE", "/test", nil)
		rsp := httptest.NewRecorder()
		srv.Router().ServeHTTP(rsp, req)
		if rsp.Code != http.StatusOK {
			t.Fatalf("Status code expected: %s but was: %v", "200", rsp.Code)
		}
	}

	if srv.test.status == running {
		t.Fatalf("Status code expected not running but was: %v", srv.test.status)
	}
}

func TestRouteGetConfig(t *testing.T) {
	// prepare
	time.Freeze(time.Now())
	defer time.Unfreeze()
	file, _ := ioutil.TempFile(os.TempDir(), "gogrinder_test")
	defer os.Remove(file.Name())

	srv := TestServer{}
	srv.test = NewTest()
	srv.test.config["Scenario"] = "scenario1"
	srv.test.config["ThinkTimeFactor"] = 2.0
	srv.test.config["ThinkTimeVariance"] = 0.1
	srv.test.filename = file.Name()

	req, _ := http.NewRequest("GET", "/config", nil)
	rsp := httptest.NewRecorder()
	srv.Router().ServeHTTP(rsp, req)
	if rsp.Code != http.StatusOK {
		t.Fatalf("Status code expected: %s but was: %v", "200", rsp.Code)
	}

	config := rsp.Body.String()
	if config != `{"config":{"Scenario":"scenario1","ThinkTimeFactor":2,`+
		`"ThinkTimeVariance":0.1},"mtime":"0001-01-01T00:00:00Z"}` {
		t.Errorf("Config not as expected: %s!", config)
	}
}

func TestRouteSaveConfig(t *testing.T) {
	// prepare
	time.Freeze(time.Now())
	defer time.Unfreeze()
	file, _ := ioutil.TempFile(os.TempDir(), "gogrinder_test")
	defer os.Remove(file.Name())

	srv := TestServer{}
	srv.test = NewTest()
	srv.test.filename = file.Name()

	config := `{"Scenario":"scenario1","ThinkTimeFactor":2,"ThinkTimeVariance":0.1}`

	{
		req, _ := http.NewRequest("PUT", "/config", strings.NewReader(config))
		rsp := httptest.NewRecorder()
		srv.Router().ServeHTTP(rsp, req)
		if rsp.Code != http.StatusOK {
			t.Fatalf("Status code expected: %s but was: %v", "200", rsp.Code)
		}
	}

	buf, err := ioutil.ReadFile(file.Name())
	if err != nil {
		t.Errorf("Unexpected problem while reading from the config file %s!", file.Name())
	}
	loadmodel := string(buf)
	if loadmodel != `{"Scenario":"scenario1","ThinkTimeFactor":2,"ThinkTimeVariance":0.1}` {
		t.Errorf("Config not as expected: %s!", loadmodel)
	}
}

func TestRouteApp(t *testing.T) {
	srv := TestServer{}
	srv.test = NewTest()
	req, _ := http.NewRequest("GET", "/app/", nil)
	rsp := httptest.NewRecorder()
	srv.Router().ServeHTTP(rsp, req)

	if rsp.Code != http.StatusOK {
		t.Fatalf("Status code expected: %v but was: %v", http.StatusOK, rsp.Code)
	}
}
