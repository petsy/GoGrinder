package gogrinder

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
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

func TestWebserverStartStop(t *testing.T) {
	// Webserver pretty much consists of boilerplate code from the graceful server docu
	// we assume for now that this works
	fake := NewTest()
	srv := NewTestServer(fake)

	l, _ := net.Listen("tcp", ":0") // 0 results in free port assignment by OS
	defer l.Close()
	fmt.Println(l.Addr().String())

	go srv.Serve(l)
	fmt.Println(srv.Addr)

	// get index.html
	http.Get(l.Addr().String() + "/app/index.html")

	// and stop
	srv.Stop(0)
}
