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
	now := Timestamp(time.Now().UTC())
	fake.Update(&Meta{Teststep: "sth", Elapsed: Elapsed(8 * time.Millisecond), Timestamp: now})
	fake.Update(&Meta{Teststep: "sth", Elapsed: Elapsed(10 * time.Millisecond), Timestamp: now})
	fake.Update(&Meta{Teststep: "sth", Elapsed: Elapsed(2 * time.Millisecond), Timestamp: now})
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
	if body != "teststep, avg_ms, min_ms, max_ms, count, error\n" +
		"sth, 6.666666, 2.000000, 10.000000, 3, 0\n" {
		t.Fatalf("Response not as expected: %s!", body)
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
	fake.Update(&Meta{Teststep: "sth", Elapsed: Elapsed(8 * time.Millisecond), Timestamp: Timestamp(now)})
	fake.Update(&Meta{Teststep: "sth", Elapsed: Elapsed(10 * time.Millisecond), Timestamp: Timestamp(now)})
	fake.Update(&Meta{Teststep: "sth", Elapsed: Elapsed(2 * time.Millisecond), Timestamp: Timestamp(now)})
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
	if body != fmt.Sprintf(`{"results":[{"teststep":"sth","avg_ms":6.666666,"min_ms":2,`+
		`"max_ms":10,"count":3,"error":0,"last":"%s"}],"running":false}`,
		now.Format(ISO8601)) {
		t.Fatalf("Response not as expected: %s", body)
	}
}

func TestRouteHandlerStatisticsWithQuery(t *testing.T) {
	// test with 3 measurements (two stats)
	srv := TestServer{}
	fake := NewTest()
	srv.test = fake
	done := srv.test.Collect() // this needs a collector to unblock update
	t1 := time.Now().UTC()
	srv.test.Update(&Meta{Teststep: "sth", Elapsed: Elapsed(8 * time.Millisecond),
		Timestamp: Timestamp(t1)})
	time.Sleep(5 * time.Millisecond)
	t2 := t1.Add(2 * time.Millisecond)
	srv.test.Update(&Meta{Teststep: "else", Elapsed: Elapsed(10 *
		time.Millisecond), Timestamp: Timestamp(t1)})
	srv.test.Update(&Meta{Teststep: "else", Elapsed: Elapsed(2 *
		time.Millisecond), Timestamp: Timestamp(t2)})
	t3 := t2.Add(2 * time.Millisecond)
	close(fake.measurements)
	<-done

	{
		// startTest
		//iso8601 := "2006-01-02T15:04:05.999Z"
		ts2 := t2.Format(ISO8601)
		req, _ := http.NewRequest("GET", "/statistics?since="+ts2, nil)
		rsp := httptest.NewRecorder()
		srv.Router().ServeHTTP(rsp, req)
		if rsp.Code != http.StatusOK {
			t.Fatalf("Status code expected: %s but was: %v", "200", rsp.Code)
		}
		results := rsp.Body.String()
		if results != fmt.Sprintf(`{"results":[{"teststep":"else","avg_ms":6,"min_ms":2,`+
			`"max_ms":10,"count":2,"error":0,"last":"%s"}],"running":false}`,
			t2.Format(ISO8601)) {
			t.Errorf("Results not as expected: %s!", results)
		}
	}

	{
		// update but no new data
		ts3 := t3.Format(ISO8601)
		req, _ := http.NewRequest("GET", "/statistics?since="+ts3, nil)
		rsp := httptest.NewRecorder()
		srv.Router().ServeHTTP(rsp, req)
		if rsp.Code != http.StatusOK {
			t.Fatalf("Status code expected: %s but was: %v", "200", rsp.Code)
		}
		results := rsp.Body.String()
		if results != `{"results":[],"running":false}` {
			t.Errorf("Results not as expected: %s!", results)
		}
	}

	{
		// get all rows
		req, _ := http.NewRequest("GET", "/statistics", nil)
		rsp := httptest.NewRecorder()
		srv.Router().ServeHTTP(rsp, req)
		if rsp.Code != http.StatusOK {
			t.Fatalf("Status code expected: %s but was: %v", "200", rsp.Code)
		}
		results := rsp.Body.String()
		if results != fmt.Sprintf(`{"results":[{"teststep":"else","avg_ms":6,"min_ms":2,"max_ms":10,`+
			`"count":2,"error":0,"last":"%s"},{"teststep":"sth","avg_ms":8,"min_ms":8,"max_ms":8,`+
			`"count":1,"error":0,"last":"%s"}],"running":false}`,
			t2.Format(ISO8601), t1.Format(ISO8601)) {
			t.Errorf("Results not as expected: %s!", results)
		}
	}
}

func TestRouteStartStop(t *testing.T) {
	// prepare
	time.Freeze(time.Now())
	defer time.Unfreeze()
	srv := TestServer{}
	srv.test = NewTest()
	tc1 := func(meta *Meta, s Settings) { srv.test.Thinktime(0.050) }
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

	if srv.test.Status() == Running {
		t.Fatalf("Status code expected not running but was: %v", srv.test.Status())
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
	loadmodel := `{
	  "Scenario": "scenario1",
	  "ThinkTimeFactor": 2.0,
	  "ThinkTimeVariance": 0.1
	}`
	srv.test.ReadConfigValidate(loadmodel, LoadmodelSchema)

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
	scenario := NewTest()
	scenario.filename = file.Name()
	srv.test = scenario

	{
		config := `{"Scenario":"scenario1","ThinkTimeFactor":2,"ThinkTimeVariance":0.1}`
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

func TestRouteWebserverStop(t *testing.T) {
	srv := TestServer{}
	srv.test = NewTest()
	req, _ := http.NewRequest("DELETE", "/stop", nil)
	rsp := httptest.NewRecorder()
	srv.Router().ServeHTTP(rsp, req)
	if rsp.Code != http.StatusOK {
		t.Fatalf("Status code expected: %s but was: %v", "200", rsp.Code)
	}
}
