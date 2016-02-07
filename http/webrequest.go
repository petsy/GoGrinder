package http

import (
	"bufio"
	"bytes"
	"encoding/json"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/finklabs/GoGrinder"
	time "github.com/finklabs/ttime"
)

// Assemble Reader from bufio that measures time until first byte
type metricReader struct {
	bytes          int
	start          time.Time
	firstByteAfter time.Duration
	readFrom       *bufio.Reader
}

func newMetricReader(start time.Time, readFrom io.Reader) *metricReader {
	// wrap into buffered reader
	return &metricReader{0, start, time.Duration(0), bufio.NewReader(readFrom)}
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

// JSON

// Response from GetJson consists of json and http Header.
type ResponseJson struct {
	Json   map[string]interface{}
	Header http.Header
}

func doJson(r *http.Request, m gogrinder.Meta) (interface{}, gogrinder.Metric) {
	start := time.Now()
	error := ""

	c := &http.Client{} // Defaultclient
	resp, err := c.Do(r)
	if err != nil {
		error += err.Error()
	}
	defer resp.Body.Close()

	mr := newMetricReader(start, resp.Body)

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

func GetJson(url string) func(gogrinder.Meta) (interface{}, gogrinder.Metric) {
	return func(m gogrinder.Meta) (interface{}, gogrinder.Metric) {
		r, err := http.NewRequest("Get", url, nil)
		if err != nil {
			return ResponseJson{}, HttpMetric{m, 0, 0, 400, err.Error()}
		}
		return doJson(r, m)
	}
}

func PostJson(url string, msg map[string]interface{}) func(gogrinder.Meta) (interface{}, gogrinder.Metric) {
	return func(m gogrinder.Meta) (interface{}, gogrinder.Metric) {
		b, err := json.Marshal(msg)
		if err != nil {
			return ResponseJson{}, HttpMetric{m, 0, 0, 400, err.Error()}
		}
		r, err := http.NewRequest("Post", url, bytes.NewReader(b))
		if err != nil {
			return ResponseJson{}, HttpMetric{m, 0, 0, 400, err.Error()}
		}
		return doJson(r, m)
	}
}

// RAW

// Response from GetRaw consists of []byte and http Header.
type ResponseRaw struct {
	Raw    []byte
	Header http.Header
}

func doRaw(r *http.Request, m gogrinder.Meta) (interface{}, gogrinder.Metric) {
	start := time.Now()
	error := ""

	c := &http.Client{} // Defaultclient
	resp, err := c.Do(r)
	if err != nil {
		error += err.Error()
	}
	defer resp.Body.Close()
	mr := newMetricReader(start, resp.Body)

	// read the response body
	raw, err := ioutil.ReadAll(mr)
	if err != nil {
		error += err.Error()
	}

	m.Elapsed = time.Now().Sub(start)
	return ResponseRaw{raw, resp.Header}, HttpMetric{m, mr.firstByteAfter, mr.bytes,
		resp.StatusCode, error}
}

func GetRaw(url string) func(gogrinder.Meta) (interface{}, gogrinder.Metric) {
	return func(m gogrinder.Meta) (interface{}, gogrinder.Metric) {
		r, err := http.NewRequest("Get", url, nil)
		if err != nil {
			return ResponseJson{}, HttpMetric{m, 0, 0, 400, err.Error()}
		}
		return doRaw(r, m)
	}
}

// DOC
// Response from Get consists of goquery Doc and http Header.
type Response struct {
	Doc    *goquery.Document
	Header http.Header
}

func do(r *http.Request, m gogrinder.Meta) (interface{}, gogrinder.Metric) {
	start := time.Now()
	error := ""

	c := &http.Client{} // Defaultclient
	resp, err := c.Do(r)
	if err != nil {
		error += err.Error()
	}
	defer resp.Body.Close()
	mr := newMetricReader(start, resp.Body)

	// read the response body and parse into document
	doc, err := goquery.NewDocumentFromReader(mr)
	if err != nil {
		error += err.Error()
	}

	m.Elapsed = time.Now().Sub(start)
	return Response{doc, resp.Header}, HttpMetric{m, mr.firstByteAfter, mr.bytes,
		resp.StatusCode, error}
}

// Get returns a goquery document.
// I used https://github.com/puerkitobio/goquery
// because it provides JQuery features and is based on Go's net/http.
func Get(url string) func(gogrinder.Meta) (interface{}, gogrinder.Metric) {
	return func(m gogrinder.Meta) (interface{}, gogrinder.Metric) {
		r, err := http.NewRequest("Get", url, nil)
		if err != nil {
			return ResponseJson{}, HttpMetric{m, 0, 0, 400, err.Error()}
		}
		return do(r, m)
	}
}

//func Post(url string, msg *goquery.Document) func(gogrinder.Meta) (interface{}, gogrinder.Metric) {
func Post(url string, msg *html.Node) func(gogrinder.Meta) (interface{}, gogrinder.Metric) {
	return func(m gogrinder.Meta) (interface{}, gogrinder.Metric) {
		var buf bytes.Buffer  // alternatively use io.Pipe()
		//err := html.Render(&buf, msg.Nodes[0])
		err := html.Render(&buf, msg)

		if err != nil {
			return ResponseJson{}, HttpMetric{m, 0, 0, 400, err.Error()}
		}
		r, err := http.NewRequest("Post", url, &buf)
		if err != nil {
			return ResponseJson{}, HttpMetric{m, 0, 0, 400, err.Error()}
		}
		return do(r, m)
	}
}
