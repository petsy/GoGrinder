package gogrinder

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/GeertJohan/go.rice"
	log "github.com/Sirupsen/logrus"
	"github.com/finklabs/graceful"
	time "github.com/finklabs/ttime"
	"github.com/gorilla/mux"
)

type Server interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Router() *mux.Router
	NewTestServer(test *TestScenario) *TestServer
}

type TestServer struct {
	test            *TestScenario
	graceful.Server // stoppable http server
}

// Assemble the Webserver for the GoGrinder frontend. It takes a testscenario as argument.
func NewTestServer(test *TestScenario) *TestServer {
	var srv TestServer
	srv = TestServer{
		test: test,
		Server: graceful.Server{
			Timeout: 5 * time.Second,
			Server: &http.Server{
				Handler: srv.Router(),
			},
		},
	}
	return &srv
}

// Error response compliant with http.Error.
type handlerError struct {
	Error   error
	Message string
	Code    int
}

// A custom handler with common error and response formatting.
type handler func(r *http.Request) (interface{}, *handlerError)

// Attach the standard ServeHTTP method to our handler so the http library can call it.
func (fn handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// call the service function
	response, err := fn(r)
	if err != nil {
		log.Error(err.Error)
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Message), err.Code)
		return
	}
	if response == nil {
		log.Error("response from method is nil")
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, "no response from service method."),
			http.StatusInternalServerError)
		return
	}

	// turn the response into JSON
	bytes, e := json.Marshal(response)
	if e != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, "marshalling Json data failed."),
			http.StatusInternalServerError)
		return
	}

	// assemble header
	w.Header().Set("Content-Type", "application/json")

	// send the response and log
	w.Write(bytes)
	log.Debugf("%s %s %s %v", r.RemoteAddr, r.Method, r.URL, http.StatusOK)
}

/////////////////////////////////////
// actual service methods
/////////////////////////////////////
func (srv *TestServer) getStatistics(r *http.Request) (interface{}, *handlerError) {
	since := ""
	since = r.URL.Query().Get("since")
	res := make(map[string]interface{})
	res["results"] = srv.test.Results(since)
	res["running"] = srv.test.status != stopped // could be stopping or running
	return res, nil
}

func (srv *TestServer) getCsv(r *http.Request) (interface{}, *handlerError) {
	var e *handlerError
	csv, err := srv.test.Csv()
	if err != nil {
		e = &handlerError{err, "error encoding csv:", 500}
	}
	return csv, e
}

func (srv *TestServer) startTest(r *http.Request) (interface{}, *handlerError) {
	if srv.test.status == stopped {
		srv.test.Exec()
	}
	return make(map[string]string), nil
}

func (srv *TestServer) stopTest(r *http.Request) (interface{}, *handlerError) {
	if srv.test.status != stopped {
		srv.test.status = stopping
	}
	return make(map[string]string), nil
}

// update the configuration and write it to file.
func (srv *TestServer) updateConfig(r *http.Request) (interface{}, *handlerError) {
	// parse config
	config, err := ioutil.ReadAll(r.Body)
	if err != nil {
		e := handlerError{err, "write error while parsing the configuration", 500} // TODO corr error code
		return make(map[string]string), &e
	}
	srv.test.ReadConfigValidate(string(config), LoadmodelSchema)

	// write config to file
	err = srv.test.WriteConfig()
	if err != nil {
		e := handlerError{err, "write error while updating the configuration", 500} // TODO corr error code
		return make(map[string]string), &e
	}

	return make(map[string]string), nil
}

// write the config to file.
func (srv *TestServer) getConfig(r *http.Request) (interface{}, *handlerError) {
	res := make(map[string]interface{})
	res["config"] = srv.test.config
	res["mtime"] = srv.test.mtime
	return res, nil
}

// Stop the web server.
func (srv *TestServer) stopWebserver(r *http.Request) (interface{}, *handlerError) {
	// e.g. curl -X "DELETE" http://localhost:3030/stop
	srv.Stop(5 * time.Second)
	return make(map[string]string), nil
}

// To simplify testing the routes I extracted the Router() following this idea:
// https://groups.google.com/d/msg/golang-nuts/Xs-Ho1feGyg/xg5amXHsM_oJ
func (srv *TestServer) Router() *mux.Router {
	router := mux.NewRouter()

	// frontend
	box := rice.MustFindBox("web")
	//_ = box
	appFileServer := http.FileServer(box.HTTPBox())
	// dev mode fallback:
	//appFileServer := http.FileServer(
	//	http.Dir("/home/mark/devel/gocode/src/github.com/finklabs/GoGrinder/web/"))
	// app route:
	router.PathPrefix("/app/").Handler(http.StripPrefix("/app/", appFileServer))

	// REST routes
	router.Handle("/statistics", handler(srv.getStatistics)).Methods("GET")
	router.Handle("/csv", handler(srv.getCsv)).Methods("GET")
	router.Handle("/config", handler(srv.getConfig)).Methods("GET")
	router.Handle("/config", handler(srv.updateConfig)).Methods("PUT")
	router.Handle("/test", handler(srv.startTest)).Methods("POST")
	router.Handle("/test", handler(srv.stopTest)).Methods("DELETE")
	router.Handle("/stop", handler(srv.stopWebserver)).Methods("DELETE")

	return router
}
