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

// TODO: clean up the other POST and PUT methods!

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
	hm := HttpMetric{m, gogrinder.Elapsed(0), 0, 421} // http status Misdirected Request
	hm.Timestamp = gogrinder.Timestamp(start)
	rr := ResponseJson{}

	c := &http.Client{} // Defaultclient
	resp, err := c.Do(r)
	if err != nil {
		hm.Error += err.Error()
	}
	if resp != nil {
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
		rr = ResponseJson{doc, resp.Header}
		//...

		hm.FirstByte = mr.firstByteAfter
		hm.Bytes = mr.bytes
		hm.Code = resp.StatusCode
	}

	hm.Elapsed = gogrinder.Elapsed(time.Now().Sub(start))
	return rr, hm
}

func GetJson(m gogrinder.Meta, args ...interface{}) (interface{}, gogrinder.Metric) {
	url := args[0].(string)
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		m.Error += err.Error()
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	return doJson(r, m)
}

func PostJson(m gogrinder.Meta, args ...interface{}) (interface{}, gogrinder.Metric) {
	url := args[0].(string)
	msg := args[1].(map[string]interface{})
	b, err := json.Marshal(msg)
	if err != nil {
		m.Error += err.Error()
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	r, err := http.NewRequest("POST", url, bytes.NewReader(b))
	r.Header.Set("Content-Type", "application/json")
	if err != nil {
		m.Error += err.Error()
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	return doJson(r, m)
}

func PutJson(m gogrinder.Meta, args ...interface{}) (interface{}, gogrinder.Metric) {
	url := args[0].(string)
	msg := args[1].(map[string]interface{})
	b, err := json.Marshal(msg)
	if err != nil {
		m.Error += err.Error()
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	r, err := http.NewRequest("PUT", url, bytes.NewReader(b))
	r.Header.Set("Content-Type", "application/json")
	if err != nil {
		m.Error += err.Error()
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	return doJson(r, m)
}

// RAW

// Response from GetRaw consists of []byte and http Header.
type ResponseRaw struct {
	Raw    []byte
	Header http.Header
}

/* extracting / reusing the non-boilerplate-part from doRaw could be difficult
func responseRaw(mr metricReader, resp ResponseRaw) *ResponseRaw {
	// read the response body
	raw, err := ioutil.ReadAll(mr)
	if err != nil {
		hm.Error += err.Error()
	}
	return ResponseRaw{raw, resp.Header}
}
*/

func doRaw(r *http.Request, m gogrinder.Meta) (interface{}, gogrinder.Metric) {
	start := time.Now()
	hm := HttpMetric{m, gogrinder.Elapsed(0), 0, 421} // http status Misdirected Request
	hm.Timestamp = gogrinder.Timestamp(start)
	rr := ResponseRaw{}

	c := &http.Client{} // Defaultclient
	resp, err := c.Do(r)
	if err != nil {
		hm.Error += err.Error()
	}
	if resp != nil {
		defer resp.Body.Close()
		mr := newMetricReader(start, resp.Body)

		// read the response body
		raw, err := ioutil.ReadAll(mr)
		if err != nil {
			hm.Error += err.Error()
		}
		rr = ResponseRaw{raw, resp.Header}
		//...
		hm.FirstByte = mr.firstByteAfter
		hm.Bytes = mr.bytes
		hm.Code = resp.StatusCode
	}

	hm.Elapsed = gogrinder.Elapsed(time.Now().Sub(start))
	return rr, hm
}

func GetRaw(m gogrinder.Meta, args ...interface{}) (interface{}, gogrinder.Metric) {
	url := args[0].(string)
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		m.Error += err.Error()
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	return doRaw(r, m)
}

func PostRaw(m gogrinder.Meta, args ...interface{}) (interface{}, gogrinder.Metric) {
	url := args[0].(string)
	r := args[1].(io.Reader)
	req, err := http.NewRequest("POST", url, r)
	if err != nil {
		m.Error = err.Error()
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	return doRaw(req, m)
}

func PutRaw(m gogrinder.Meta, args ...interface{}) (interface{}, gogrinder.Metric) {
	url := args[0].(string)
	r := args[1].(io.Reader)
	req, err := http.NewRequest("PUT", url, r)
	if err != nil {
		m.Error += err.Error()
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	return doRaw(req, m)
}

func DeleteRaw(m gogrinder.Meta, args ...interface{}) (interface{}, gogrinder.Metric) {
	url := args[0].(string)
	r, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		m.Error += err.Error()
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	return doRaw(r, m)
}

// DOC
// Response from Get consists of goquery Doc and http Header.
type Response struct {
	Doc    *goquery.Document
	Header http.Header
}

func do(r *http.Request, m gogrinder.Meta) (interface{}, gogrinder.Metric) {
	start := time.Now()
	hm := HttpMetric{m, gogrinder.Elapsed(0), 0, 421} // http status Misdirected Request
	hm.Timestamp = gogrinder.Timestamp(start)
	rr := Response{}

	c := &http.Client{} // Defaultclient
	resp, err := c.Do(r)
	if err != nil {
		hm.Error += err.Error()
	}
	if resp != nil {
		defer resp.Body.Close()
		mr := newMetricReader(start, resp.Body)

		// read the response body and parse into document
		doc, err := goquery.NewDocumentFromReader(mr)
		if err != nil {
			m.Error += err.Error()
		}
		rr = Response{doc, resp.Header}
		// ...
		hm.FirstByte = mr.firstByteAfter
		hm.Bytes = mr.bytes
		hm.Code = resp.StatusCode
	}

	hm.Elapsed = gogrinder.Elapsed(time.Now().Sub(start))
	return rr, hm
}

// Get returns a goquery document.
// I used https://github.com/puerkitobio/goquery
// because it provides JQuery features and is based on Go's net/http.
func Get(m gogrinder.Meta, args ...interface{}) (interface{}, gogrinder.Metric) {
	url := args[0].(string)
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		m.Error += err.Error()
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	return do(r, m)
}

func Post(m gogrinder.Meta, args ...interface{}) (interface{}, gogrinder.Metric) {
	url := args[0].(string)
	msg := args[1].(*html.Node)
	var buf bytes.Buffer // alternatively use io.Pipe()
	err := html.Render(&buf, msg)

	if err != nil {
		m.Error += err.Error()
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	r, err := http.NewRequest("POST", url, &buf)
	r.Header.Set("Content-Type", "application/xml")
	if err != nil {
		m.Error += err.Error()
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	return do(r, m)
}

func Put(m gogrinder.Meta, args ...interface{}) (interface{}, gogrinder.Metric) {
	url := args[0].(string)
	msg := args[1].(*html.Node)
	var buf bytes.Buffer // alternatively use io.Pipe()
	err := html.Render(&buf, msg)

	if err != nil {
		m.Error += err.Error()
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	r, err := http.NewRequest("PUT", url, &buf)
	r.Header.Set("Content-Type", "application/xml")
	if err != nil {
		m.Error += err.Error()
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	return do(r, m)
}
