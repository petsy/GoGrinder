package http

import (
	"net/http"
	//"golang.org/x/net/html"

	"github.com/finklabs/GoGrinder"
)

func Get(url string) func(gogrinder.Meta) (interface{}, gogrinder.Metric) {
	return func(m gogrinder.Meta) (interface{}, gogrinder.Metric) {
		_, err := http.Get(url)

		// read the response body and parse into document
		//t := html.NewTokenizer(resp.Body)

		return make(map[string]string), HttpMetric{m, 0, 0, 0, err.Error()}
	}
}
