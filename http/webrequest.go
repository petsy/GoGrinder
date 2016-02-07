package http

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/finklabs/GoGrinder"
	time "github.com/finklabs/ttime"
	"io/ioutil"
)

// Assemble Reader from bufio that measures time until first byte
type metricReader struct {
	bytes          int
	start          time.Time
	firstByteAfter time.Duration
	readFrom       *bufio.Reader
}

func newMetricReader(readFrom io.Reader) *metricReader {
	// wrap into buffered reader
	return &metricReader{0, time.Now(), time.Duration(0), bufio.NewReader(readFrom)}
}

func (fb *metricReader) Read(p []byte) (n int, err error) {
	if fb.firstByteAfter == time.Duration(0) {
		fb.readFrom.ReadByte()
		fb.firstByteAfter = time.Now().Sub(fb.start)
		fb.readFrom.UnreadByte()
	}
	n, err = fb.readFrom.Read(p)
	fb.bytes += n
	return
}

// Response from Get consists of goquery Doc and http Header.
type Response struct {
	Doc    *goquery.Document
	Header http.Header
}

// Response from GetRaw consists of []byte and http Header.
type ResponseRaw struct {
	Raw    []byte
	Header http.Header
}

// Response from GetRaw consists of gjson and http Header.
type ResponseJson struct {
	Json   map[string]interface{}
	Header http.Header
}

// Get returns a goquery document.
// I used https://github.com/puerkitobio/goquery
// because it provides JQuery features and is based on Go's net/http.
func Get(url string) func(gogrinder.Meta) (interface{}, gogrinder.Metric) {
	return func(m gogrinder.Meta) (interface{}, gogrinder.Metric) {
		error := ""
		start := time.Now()
		resp, err := http.Get(url)
		if err != nil {
			error += err.Error()
		}
		defer resp.Body.Close()
		mr := newMetricReader(resp.Body)

		// read the response body and parse into document
		doc, err := goquery.NewDocumentFromReader(mr)
		if err != nil {
			error += err.Error()
		}

		m.Elapsed = time.Now().Sub(start)
		return Response{doc, resp.Header}, HttpMetric{m, mr.firstByteAfter, mr.bytes,
			resp.StatusCode, error}
	}
}

func GetRaw(url string) func(gogrinder.Meta) (interface{}, gogrinder.Metric) {
	return func(m gogrinder.Meta) (interface{}, gogrinder.Metric) {
		error := ""
		start := time.Now()
		resp, err := http.Get(url)
		if err != nil {
			error += err.Error()
		}
		defer resp.Body.Close()
		mr := newMetricReader(resp.Body)

		// read the response body
		raw, err := ioutil.ReadAll(mr)
		if err != nil {
			error += err.Error()
		}

		m.Elapsed = time.Now().Sub(start)
		return ResponseRaw{raw, resp.Header}, HttpMetric{m, mr.firstByteAfter, mr.bytes,
			resp.StatusCode, error}
	}
}

func GetJson(url string) func(gogrinder.Meta) (interface{}, gogrinder.Metric) {
	return func(m gogrinder.Meta) (interface{}, gogrinder.Metric) {
		error := ""
		start := time.Now()
		resp, err := http.Get(url)
		if err != nil {
			error += err.Error()
		}
		defer resp.Body.Close()
		mr := newMetricReader(resp.Body)

		// read the response body and parse as json
		doc := make(map[string]interface{})
		raw, err := ioutil.ReadAll(mr)
		if err != nil {
			error = err.Error()
		}
		err = json.Unmarshal(raw, &doc)
		if err != nil {
			error += err.Error()
		}

		m.Elapsed = time.Now().Sub(start)
		return ResponseJson{doc, resp.Header}, HttpMetric{m, mr.firstByteAfter, mr.bytes,
			resp.StatusCode, error}
	}
}

// TODO
// * create versions for json and raw, too (GetJson, GetRaw)
// * complete this with POST, PUT, DELETE
// * expose the Header
