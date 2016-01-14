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

	myhandler := handler(func(r *http.Request) (interface{}, *handlerError) {
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

	myhandler := handler(func(r *http.Request) (interface{}, *handlerError) {
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

	myhandler := handler(func(r *http.Request) (interface{}, *handlerError) {
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
	done := fake.Collect() // this needs a collector to unblock update
	now := time.Now().UTC()
	fake.Update("sth", 8*time.Millisecond, now)
	fake.Update("sth", 10*time.Millisecond, now)
	fake.Update("sth", 2*time.Millisecond, now)
	close(fake.measurements)
	<-done

	// invoke REST service
	request, _ := http.NewRequest("GET", "/statistics", nil)
	srv := TestServer{}
	srv.test = fake
	response, err := srv.getStatistics(request)

	if err != nil {
		t.Fatalf("Error while processing: %s", err)
	}
	if response.(map[string]interface{})["results"].([]Result)[0] != (Result{"sth", 6666666, 2000000, 10000000, 3, now.Format(ISO8601)}) {
		t.Fatalf("Response nsot as expected: %v", response.([]Result)[0])
	}
}

func TestHandlerStatisticsWithQuery(t *testing.T) {
	// test with 3 measurements (two stats)
	var fake = NewTest()
	done := fake.Collect() // this needs a collector to unblock update
	t1 := time.Now().UTC()
	fake.Update("sth", 8*time.Millisecond, t1)
	time.Sleep(5 * time.Millisecond)
	t2 := t1.Add(2 * time.Millisecond)
	fake.Update("else", 10*time.Millisecond, t1)
	fake.Update("else", 2*time.Millisecond, t2)
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
		t.Fatalf("Error while processing: %s", err)
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
		t.Fatalf("Error while processing: %s", err)
	}
	if len(response.(map[string]interface{})["results"].([]Result)) != 2 {
		t.Fatalf("Response should contain exactly 2 rows.")
	}
	// "else" is [0]
	if response.(map[string]interface{})["results"].([]Result)[0] !=
			(Result{"else", 6000000, 2000000, 10000000, 2, t2.Format(ISO8601)}) {
		t.Log(t2.Format(ISO8601))
		t.Log("Response 0: %v", response.(map[string]interface{})["results"].([]Result)[0])
		t.Log("Response 1: %v", response.(map[string]interface{})["results"].([]Result)[1])
		t.Fatalf("Response not as expected: %v", response.(map[string]interface{})["results"].([]Result)[0])
	}
	// "sth" is [1]
	if response.(map[string]interface{})["results"].([]Result)[1] !=
			(Result{"sth", 8000000, 8000000, 8000000, 1, t1.Format(ISO8601)}) {
		t.Log(t1)
		t.Fatalf("Response not as expected: %v", response.(map[string]interface{})["results"].([]Result)[1])
	}
}
