package gogrinder

import (
	"github.com/GeertJohan/go.rice"
	"net/http"
)

// webserver is terminated once main exits
func Webserver() {
	go func() {
		http.Handle("/", http.FileServer(rice.MustFindBox("static").HTTPBox()))
		http.ListenAndServe(":3000", nil)
	}()
}

// TODO add the API
func Restserver() {

}
