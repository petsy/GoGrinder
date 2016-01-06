package gogrinder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	time "github.com/finklabs/ttime"
)

func TestHandlerServeHTTP(t *testing.T) {
	// make sure handler's ServeHTTP works
	request, _ := http.NewRequest("GET", "/something", nil)
	response := httptest.NewRecorder()

	myhandler := handler(func(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError) {
		return "moinmoin", nil
	})

	myhandler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("Status code not %s as expected: %v", "200", response.Code)
	}
	if value, ok := response.Header()["Content-Type"]; !ok || value[0] != "application/json" {
		t.Fatalf("Content-Type not %s as expected: %s", "application/json", value)
	}
	body := response.Body
	x, _ := json.Marshal("moinmoin")
	if body.String() != string(x) {
		t.Fatalf("Response not as expected: %v", body)
	}
}

func TestHandlerServeHTTPWrapError(t *testing.T) {
	// make sure handler's ServeHTTP works
	request, _ := http.NewRequest("GET", "/something", nil)
	response := httptest.NewRecorder()

	myhandler := handler(func(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError) {
		return nil, &handlerError{fmt.Errorf("sorry"), "error 500", http.StatusInternalServerError}
	})

	myhandler.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("Status code not %s  as expected: %v", "500", response.Code)
	}
}

func TestHandlerServeHTTPEmptyResponse(t *testing.T) {
	// make sure handler's ServeHTTP works
	request, _ := http.NewRequest("GET", "/something", nil)
	response := httptest.NewRecorder()

	myhandler := handler(func(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError) {
		return nil, nil
	})

	myhandler.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("Status code not %s  as expected: %v", "500", response.Code)
	}
}

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

func TestGetStatistics(t *testing.T) {
	// test with 3 measurements
	var fake = NewTest()
	done := fake.collect() // this needs a collector to unblock update
	now := time.Now()
	fake.update("sth", 8*time.Millisecond, now)
	fake.update("sth", 10*time.Millisecond, now)
	fake.update("sth", 2*time.Millisecond, now)
	close(fake.measurements)
	<-done

	// invoke REST service
	request, _ := http.NewRequest("GET", "/statistics", nil)
	response, err := fake.getStatistics(nil, request)

	if err != nil {
		t.Fatalf("Error while processing: %s", err)
	}
	if response.(stats)["sth"] != (stats_value{6666666, 2000000, 10000000, 3, now}) {
		t.Fatalf("Response not as expected: %v", response.(stats)["sth"])
	}
}

func TestHandlerStatisticsWithQuery(t *testing.T) {
	// test with 3 measurements (two stats)
	var fake = NewTest()
	done := fake.collect() // this needs a collector to unblock update
	t1 := time.Now().UTC()
	fake.update("sth", 8*time.Millisecond, t1)
	time.Sleep(5 * time.Millisecond)
	t2 := time.Now().UTC()
	fake.update("else", 10*time.Millisecond, t2)
	fake.update("else", 2*time.Millisecond, t2)
	close(fake.measurements)
	<-done

	// invoke REST service for stats update
	iso8601 := "2006-01-02T15:04:05.999Z"
	ts := t2.Format(iso8601)
	request, _ := http.NewRequest("GET", "/statistics?since=" + ts, nil)
	response, err := fake.getStatistics(nil, request)

	if err != nil {
		t.Fatalf("Error while processing: %s", err)
	}
	if len(response.(stats)) != 1 {
		t.Fatalf("Response should contain exactly 1 row.")
	}
	if response.(stats)["else"] != (stats_value{6000000, 2000000, 10000000, 2, t2}) {
		t.Log(t2)
		t.Fatalf("Response not as expected: %v", response.(stats)["else"])
	}

	// get all rows
	request, _ = http.NewRequest("GET", "/statistics", nil)
	response, err = fake.getStatistics(nil, request)

	if err != nil {
		t.Fatalf("Error while processing: %s", err)
	}
	if len(response.(stats)) != 2 {
		t.Fatalf("Response should contain exactly 2 rows.")
	}
	if response.(stats)["sth"] != (stats_value{8000000, 8000000, 8000000, 1, t1}) {
		t.Fatalf("Response not as expected: %v", response.(stats)["sth"])
	}
	if response.(stats)["else"] != (stats_value{6000000, 2000000, 10000000, 2, t2}) {
		t.Fatalf("Response not as expected: %v", response.(stats)["else"])
	}
}
