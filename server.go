package gogrinder

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/GeertJohan/go.rice"
	"github.com/finklabs/graceful"
	time "github.com/finklabs/ttime"
	"github.com/gorilla/mux"
)

// error response compliant with http.Error
type handlerError struct {
	Error   error
	Message string
	Code    int
}

// a custom handler with common error and response formatting
type handler func(r *http.Request) (interface{}, *handlerError)

// attach the standard ServeHTTP method to our handler so the http library can call it
func (fn handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: make the logger plugable
	// call the service function
	response, err := fn(r)

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

	// assemble header
	w.Header().Set("Content-Type", "application/json")
	// not sure if we still need the CORS issue fix
//	if origin := r.Header.Get("Origin"); origin != "" {
//		fmt.Println(origin)
//		w.Header().Set("Access-Control-Allow-Origin", origin)
//		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
//		w.Header().Set("Access-Control-Allow-Headers",
//			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
//	}

	// send the response and log
	w.Write(bytes)
	log.Printf("%s %s %s %d", r.RemoteAddr, r.Method, r.URL, 200)
}

// actual REST handlers
func (test *Test) getStatistics(r *http.Request) (interface{}, *handlerError) {
	since := ""
	since = r.URL.Query().Get("since")
	res := make(map[string]interface{})
	res["results"] = test.Results(since)
	res["running"] = test.status != stopped  // could be stopping or running
	return res, nil
}

// TODO: start stop of server processes needs testing!
func (test *Test) startTest(r *http.Request) (interface{}, *handlerError) {
	if (test.status == stopped) {
		test.Exec()
	}
	return make(map[string]string), nil
}

func (test *Test) stopTest(r *http.Request) (interface{}, *handlerError) {
	if (test.status != stopped) {
		test.status = stopping
	}
	return make(map[string]string), nil
}

// simple get op
//func getLoadmodel(r *http.Request) (interface{}, *handlerError) {
//}

// stop the server
func (test *Test) stopWebserver(r *http.Request) (interface{}, *handlerError) {
	// e.g. curl -X "DELETE" http://localhost:3000/stop
	test.server.Stop(5 * time.Second)
	return make(map[string]string), nil
}

// TODO: we need some kind of integration test to make sure routes work as expected
func (test *Test) Webserver() {
	router := mux.NewRouter()

	// frontend
	box := rice.MustFindBox("web")
	_ = box
	// prod mode:
	//appFileServer := http.FileServer(box.HTTPBox())
	// dev mode:
	appFileServer := http.FileServer(http.Dir("/home/mark/devel/gocode/src/github.com/finklabs/GoGrinder/web/"))
	// app route:
	router.PathPrefix("/app/").Handler(http.StripPrefix("/app/", appFileServer))

	// REST routes
	router.Handle("/statistics", handler(test.getStatistics)).Methods("GET")
	//router.Handle("/loadmodel", handler(getLoadmodel)).Methods("GET")
	//router.Handle("/loadmodel", handler(updateLoadmodel)).Methods("PUT")
	router.Handle("/test", handler(test.startTest)).Methods("POST")
	router.Handle("/test", handler(test.stopTest)).Methods("DELETE")
	router.Handle("/stop", handler(test.stopWebserver)).Methods("DELETE")

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
