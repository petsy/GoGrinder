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
	fake.update("sth", 8*time.Millisecond)
	fake.update("sth", 10*time.Millisecond)
	fake.update("sth", 2*time.Millisecond)
	close(fake.measurements)
	<-done

	// invoke REST service
	response, err := fake.getStatistics(nil, nil)

	if err != nil {
		t.Fatalf("Error while processing: %s", err)
	}
	if response.(stats)["sth"] != (stats_value{6666666, 2000000, 10000000, 3}) {
		t.Fatalf("Response not as expected: %v", response.(stats)["sth"])
	}
}

//func TestHandlerStatisticsWithQuery(t *testing.T) {
//	// use since query with ISO8601 datetime
//	request, _ := http.NewRequest("GET", "/statistics?since=2015-12-31T22:00:00.000Z", nil)
//	//response := httptest.NewRecorder()
//
//	var fake = NewTest()
//	response, err := fake.getStatistics(nil, request)
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
