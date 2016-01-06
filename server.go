package gogrinder

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/GeertJohan/go.rice"
	"github.com/finklabs/graceful"
	"github.com/gorilla/mux"
	time "github.com/finklabs/ttime"
)

// error response compliant with http.Error
type handlerError struct {
	Error   error
	Message string
	Code    int
}

// a custom handler with common error and response formatting
type handler func(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError)

// attach the standard ServeHTTP method to our handler so the http library can call it
func (fn handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: make the logger plugable
	// call the service function
	response, err := fn(w, r)

	// check for errors
	if err != nil {
		log.Printf("ERROR: %v\n", err.Error)
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Message), err.Code)
		return
	}
	if response == nil {
		log.Printf("ERROR: response from method is nil\n")
		http.Error(w, "Internal server error.", http.StatusInternalServerError)
		return
	}

	// turn the response into JSON
	bytes, e := json.Marshal(response)
	if e != nil {
		http.Error(w, "Error marshalling Json data.", http.StatusInternalServerError)
		return
	}

	// send the response and log
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
	log.Printf("%s %s %s %d", r.RemoteAddr, r.Method, r.URL, 200)
}

// actual REST handlers
func (test *Test) getStatistics(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError) {
	if since, ok := r.URL.Query()["since"]; ok {
		//RFC3339Nano := "2006-01-02T15:04:05.999999999Z07:00"
		iso8601 := "2006-01-02T15:04:05.999Z"
		t, err := time.Parse(iso8601, since[0])
		if err != nil {
			return nil, &handlerError{err, "since should be ISO8601", http.StatusBadRequest}
		}
		s := test.StatsUpdate(t)
		return s, nil
	} else {
		s := test.Stats()
		return s, nil
	}
}

// simple get op
//func getLoadmodel(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError) {
//}

// stop the server
func (test *Test) StopWebserver(w http.ResponseWriter, r *http.Request) (interface{}, *handlerError) {
	// e.g. curl -X "DELETE" http://localhost:3000/stop
	test.server.Stop(5 * time.Second)
	return make(map[string]string), nil
}

func (test *Test) Webserver() {
	router := mux.NewRouter()

	// frontend
	router.Handle("/app", http.FileServer(rice.MustFindBox("web").HTTPBox()))

	// REST routes
	router.Handle("/statistics", handler(test.getStatistics)).Methods("GET")
	//router.Handle("/loadmodel", handler(getLoadmodel)).Methods("GET")
	//router.Handle("/loadmodel", handler(updateLoadmodel)).Methods("PUT")
	//router.Handle("/test", handler(startTest)).Methods("POST")
	//router.Handle("/test", handler(stopTest)).Methods("DELETE")
	router.Handle("/stop", handler(test.StopWebserver)).Methods("DELETE")

	test.server = graceful.Server{
		Timeout: 5 * time.Second,

		Server: &http.Server{
			Addr:    ":3000",
			Handler: router,
		},
	}

	// start the stoppable server (this uses graceful, a stoppable server)
	test.server.ListenAndServe()
}
