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
	firstByteAfter gogrinder.Elapsed
	readFrom       *bufio.Reader
}

func newMetricReader(start time.Time, readFrom io.Reader) *metricReader {
	// wrap into buffered reader
	return &metricReader{0, start, gogrinder.Elapsed(0), bufio.NewReader(readFrom)}
}

func (fb *metricReader) Read(p []byte) (n int, err error) {
	if fb.firstByteAfter == gogrinder.Elapsed(0) {
		fb.readFrom.ReadByte()
		fb.firstByteAfter = gogrinder.Elapsed(time.Now().Sub(fb.start))
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

	c := &http.Client{} // Defaultclient
	resp, err := c.Do(r)
	if err != nil {
		m.Error += err.Error()
	}
	defer resp.Body.Close()

	mr := newMetricReader(start, resp.Body)

	// read the response body and parse as json
	doc := make(map[string]interface{})
	raw, err := ioutil.ReadAll(mr)
	if err != nil {
		m.Error += err.Error()
	}
	err = json.Unmarshal(raw, &doc)
	if err != nil {
		m.Error += err.Error()
	}

	m.Elapsed = gogrinder.Elapsed(time.Now().Sub(start))

	return ResponseJson{doc, resp.Header}, HttpMetric{m, mr.firstByteAfter, mr.bytes,
		resp.StatusCode}
}

func GetJson(url string) func(gogrinder.Meta) (interface{}, gogrinder.Metric) {
	return func(m gogrinder.Meta) (interface{}, gogrinder.Metric) {
		r, err := http.NewRequest("Get", url, nil)
		if err != nil {
			m.Error += err.Error()
			return ResponseJson{}, HttpMetric{m, 0, 0, 400}
		}
		return doJson(r, m)
	}
}

func PostJson(url string, msg map[string]interface{}) func(gogrinder.Meta) (interface{}, gogrinder.Metric) {
	return func(m gogrinder.Meta) (interface{}, gogrinder.Metric) {
		b, err := json.Marshal(msg)
		if err != nil {
			m.Error += err.Error()
			return ResponseJson{}, HttpMetric{m, 0, 0, 400}
		}
		r, err := http.NewRequest("Post", url, bytes.NewReader(b))
		r.Header.Set("Content-Type", "application/json")
		if err != nil {
			m.Error += err.Error()
			return ResponseJson{}, HttpMetric{m, 0, 0, 400}
		}
		return doJson(r, m)
	}
}

func PutJson(url string, msg map[string]interface{}) func(gogrinder.Meta) (interface{}, gogrinder.Metric) {
	return func(m gogrinder.Meta) (interface{}, gogrinder.Metric) {
		b, err := json.Marshal(msg)
		if err != nil {
			m.Error += err.Error()
			return ResponseJson{}, HttpMetric{m, 0, 0, 400}
		}
		r, err := http.NewRequest("Put", url, bytes.NewReader(b))
		r.Header.Set("Content-Type", "application/json")
		if err != nil {
			m.Error += err.Error()
			return ResponseJson{}, HttpMetric{m, 0, 0, 400}
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

	c := &http.Client{} // Defaultclient
	resp, err := c.Do(r)
	if err != nil {
		m.Error += err.Error()
	}
	defer resp.Body.Close()
	mr := newMetricReader(start, resp.Body)

	// read the response body
	raw, err := ioutil.ReadAll(mr)
	if err != nil {
		m.Error += err.Error()
	}

	m.Elapsed = gogrinder.Elapsed(time.Now().Sub(start))
	return ResponseRaw{raw, resp.Header}, HttpMetric{m, mr.firstByteAfter, mr.bytes,
		resp.StatusCode}
}

func GetRaw(url string) func(gogrinder.Meta) (interface{}, gogrinder.Metric) {
	return func(m gogrinder.Meta) (interface{}, gogrinder.Metric) {
		r, err := http.NewRequest("Get", url, nil)
		if err != nil {
			m.Error += err.Error()
			return ResponseJson{}, HttpMetric{m, 0, 0, 400}
		}
		return doRaw(r, m)
	}
}

func PostRaw(url string, r io.Reader) func(gogrinder.Meta) (interface{}, gogrinder.Metric) {
	return func(m gogrinder.Meta) (interface{}, gogrinder.Metric) {
		req, err := http.NewRequest("Post", url, r)
		if err != nil {
			m.Error = err.Error()
			return ResponseJson{}, HttpMetric{m, 0, 0, 400}
		}
		return doRaw(req, m)
	}
}

func PutRaw(url string, r io.Reader) func(gogrinder.Meta) (interface{}, gogrinder.Metric) {
	return func(m gogrinder.Meta) (interface{}, gogrinder.Metric) {
		req, err := http.NewRequest("Put", url, r)
		if err != nil {
			m.Error += err.Error()
			return ResponseJson{}, HttpMetric{m, 0, 0, 400}
		}
		return doRaw(req, m)
	}
}

func DeleteRaw(url string) func(gogrinder.Meta) (interface{}, gogrinder.Metric) {
	return func(m gogrinder.Meta) (interface{}, gogrinder.Metric) {
		r, err := http.NewRequest("Delete", url, nil)
		if err != nil {
			m.Error += err.Error()
			return ResponseJson{}, HttpMetric{m, 0, 0, 400}
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

	c := &http.Client{} // Defaultclient
	resp, err := c.Do(r)
	if err != nil {
		m.Error += err.Error()
	}
	defer resp.Body.Close()
	mr := newMetricReader(start, resp.Body)

	// read the response body and parse into document
	doc, err := goquery.NewDocumentFromReader(mr)
	if err != nil {
		m.Error += err.Error()
	}

	m.Elapsed = gogrinder.Elapsed(time.Now().Sub(start))
	return Response{doc, resp.Header}, HttpMetric{m, mr.firstByteAfter, mr.bytes,
		resp.StatusCode}
}

// Get returns a goquery document.
// I used https://github.com/puerkitobio/goquery
// because it provides JQuery features and is based on Go's net/http.
func Get(url string) func(gogrinder.Meta) (interface{}, gogrinder.Metric) {
	return func(m gogrinder.Meta) (interface{}, gogrinder.Metric) {
		r, err := http.NewRequest("Get", url, nil)
		if err != nil {
			m.Error += err.Error()
			return ResponseJson{}, HttpMetric{m, 0, 0, 400}
		}
		return do(r, m)
	}
}

func Post(url string, msg *html.Node) func(gogrinder.Meta) (interface{}, gogrinder.Metric) {
	return func(m gogrinder.Meta) (interface{}, gogrinder.Metric) {
		var buf bytes.Buffer // alternatively use io.Pipe()
		err := html.Render(&buf, msg)

		if err != nil {
			m.Error += err.Error()
			return ResponseJson{}, HttpMetric{m, 0, 0, 400}
		}
		r, err := http.NewRequest("Post", url, &buf)
		r.Header.Set("Content-Type", "application/xml")
		if err != nil {
			m.Error += err.Error()
			return ResponseJson{}, HttpMetric{m, 0, 0, 400}
		}
		return do(r, m)
	}
}

func Put(url string, msg *html.Node) func(gogrinder.Meta) (interface{}, gogrinder.Metric) {
	return func(m gogrinder.Meta) (interface{}, gogrinder.Metric) {
		var buf bytes.Buffer // alternatively use io.Pipe()
		err := html.Render(&buf, msg)

		if err != nil {
			m.Error += err.Error()
			return ResponseJson{}, HttpMetric{m, 0, 0, 400}
		}
		r, err := http.NewRequest("Put", url, &buf)
		r.Header.Set("Content-Type", "application/xml")
		if err != nil {
			m.Error += err.Error()
			return ResponseJson{}, HttpMetric{m, 0, 0, 400}
		}
		return do(r, m)
	}
}
