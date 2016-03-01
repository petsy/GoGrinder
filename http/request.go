// Package http is part of the GoGrinder load & performance test tool.
// It adds instrumentation to the net/http package and a prometheus reporter.
//
package http

import (
	"bufio"
	"bytes"
	"encoding/json"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/finklabs/GoGrinder/gogrinder"
	time "github.com/finklabs/ttime"
)

// Default client which implements cookiejar
func NewDefaultClient() *http.Client {
	cookieJar, _ := cookiejar.New(nil)

	client := &http.Client{
		Jar: cookieJar,
	}

	return client
}

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

func doJson(m gogrinder.Meta, c *http.Client, r *http.Request) (interface{}, gogrinder.Metric) {
	start := time.Now()
	hm := HttpMetric{m, gogrinder.Elapsed(0), 0, 421} // http status Misdirected Request
	hm.Timestamp = gogrinder.Timestamp(start)
	rr := ResponseJson{}

	resp, err := c.Do(r)
	if err != nil {
		hm.Error += err.Error()
	}
	if resp != nil {
		defer resp.Body.Close()
		mr := newMetricReader(start, resp.Body)

		// read the response body and parse as json
		raw, err := ioutil.ReadAll(mr)
		if err != nil {
			m.Error += err.Error()
		}
		doc := make(map[string]interface{})
		if len(raw) > 0 {
			if raw[0] == '[' {
				// a REST service response to be an array does not seem to be a good idea:
				// http://stackoverflow.com/questions/12293979/how-do-i-return-a-json-array-with-bottle
				// many applications do this anyway...
				// so for now we need a workaround:
				var array []interface{}
				err = json.Unmarshal(raw, &array)
				doc["data"] = array
			} else {
				err = json.Unmarshal(raw, &doc)
			}
			if err != nil {
				m.Error += err.Error()
			}
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
	if len(args) != 2 {
		m.Error += "GetJson requires http.Client and a string url argument.\n"
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	c := args[0].(*http.Client)
	url := args[1].(string)
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		m.Error += err.Error()
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	return doJson(m, c, r)
}

func PostJson(m gogrinder.Meta, args ...interface{}) (interface{}, gogrinder.Metric) {
	if len(args) != 3 {
		m.Error += "PostJson requires http.Client, string url and map[string]interface{} arguments.\n"
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	c := args[0].(*http.Client)
	url := args[1].(string)
	msg := args[2].(map[string]interface{})
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
	return doJson(m, c, r)
}

func PutJson(m gogrinder.Meta, args ...interface{}) (interface{}, gogrinder.Metric) {
	if len(args) != 3 {
		m.Error += "PutJson requires http.Client, string url and map[string]interface{} arguments.\n"
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	c := args[0].(*http.Client)
	url := args[1].(string)
	msg := args[2].(map[string]interface{})
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
	return doJson(m, c, r)
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

func doRaw(m gogrinder.Meta, c *http.Client, r *http.Request) (interface{}, gogrinder.Metric) {
	start := time.Now()
	hm := HttpMetric{m, gogrinder.Elapsed(0), 0, 421} // http status Misdirected Request
	hm.Timestamp = gogrinder.Timestamp(start)
	rr := ResponseRaw{}

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
	if len(args) != 2 {
		m.Error += "GetRaw requires http.Client and string url argument.\n"
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	c := args[0].(*http.Client)
	url := args[1].(string)
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		m.Error += err.Error()
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	return doRaw(m, c, r)
}

func PostRaw(m gogrinder.Meta, args ...interface{}) (interface{}, gogrinder.Metric) {
	if len(args) != 3 {
		m.Error += "PostRaw requires http.Client, string url and io.Reader arguments.\n"
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	c := args[0].(*http.Client)
	url := args[1].(string)
	r := args[2].(io.Reader)
	req, err := http.NewRequest("POST", url, r)
	if err != nil {
		m.Error = err.Error()
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	return doRaw(m, c, req)
}

func FormRaw(m gogrinder.Meta, args ...interface{}) (interface{}, gogrinder.Metric) {
	if len(args) != 3 {
		m.Error += "FormRaw requires http.Client, string url and url.Values arguments.\n"
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	c := args[0].(*http.Client)
	u := args[1].(string)
	f := args[2].(url.Values)
	req, err := http.NewRequest("POST", u, strings.NewReader(f.Encode()))
	if err != nil {
		m.Error = err.Error()
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	return doRaw(m, c, req)
}

func PutRaw(m gogrinder.Meta, args ...interface{}) (interface{}, gogrinder.Metric) {
	if len(args) != 3 {
		m.Error += "PutRaw requires http.Client, string url and io.Reader arguments.\n"
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	c := args[0].(*http.Client)
	url := args[1].(string)
	r := args[2].(io.Reader)
	req, err := http.NewRequest("PUT", url, r)
	if err != nil {
		m.Error += err.Error()
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	return doRaw(m, c, req)
}

func DeleteRaw(m gogrinder.Meta, args ...interface{}) (interface{}, gogrinder.Metric) {
	if len(args) != 2 {
		m.Error += "DeleteRaw requires http.Client and a string url argument.\n"
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	c := args[0].(*http.Client)
	url := args[1].(string)
	r, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		m.Error += err.Error()
		return ResponseJson{}, HttpMetric{m, 0, 0, 400}
	}
	return doRaw(m, c, r)
}

// DOC
// Response from Get consists of goquery Doc and http Header.
type Response struct {
	Doc    *goquery.Document
	Header http.Header
}

func do(m gogrinder.Meta, c *http.Client, r *http.Request) (interface{}, gogrinder.Metric) {
	start := time.Now()
	hm := HttpMetric{m, gogrinder.Elapsed(0), 0, 421} // http status Misdirected Request
	hm.Timestamp = gogrinder.Timestamp(start)
	rr := Response{}

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
	if len(args) != 2 {
		m.Error += "Get requires http.Client and a string url argument.\n"
		return Response{}, HttpMetric{m, 0, 0, 400}
	}
	c := args[0].(*http.Client)
	url := args[1].(string)
	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		m.Error += err.Error()
		return Response{}, HttpMetric{m, 0, 0, 400}
	}
	return do(m, c, r)
}

func Post(m gogrinder.Meta, args ...interface{}) (interface{}, gogrinder.Metric) {
	if len(args) != 3 {
		m.Error += "Post requires http.Client, string url and html.Node arguments.\n"
		return Response{}, HttpMetric{m, 0, 0, 400}
	}
	c := args[0].(*http.Client)
	url := args[1].(string)
	msg := args[2].(*html.Node)
	var buf bytes.Buffer // alternatively use io.Pipe()
	err := html.Render(&buf, msg)

	if err != nil {
		m.Error += err.Error()
		return Response{}, HttpMetric{m, 0, 0, 400}
	}
	r, err := http.NewRequest("POST", url, &buf)
	r.Header.Set("Content-Type", "application/xml")
	if err != nil {
		m.Error += err.Error()
		return Response{}, HttpMetric{m, 0, 0, 400}
	}
	return do(m, c, r)
}

func Put(m gogrinder.Meta, args ...interface{}) (interface{}, gogrinder.Metric) {
	if len(args) != 3 {
		m.Error += "Put requires http.Client, string url and html.Node arguments.\n"
		return Response{}, HttpMetric{m, 0, 0, 400}
	}
	c := args[0].(*http.Client)
	url := args[1].(string)
	msg := args[2].(*html.Node)
	var buf bytes.Buffer // alternatively use io.Pipe()
	err := html.Render(&buf, msg)

	if err != nil {
		m.Error += err.Error()
		return Response{}, HttpMetric{m, 0, 0, 400}
	}
	r, err := http.NewRequest("PUT", url, &buf)
	r.Header.Set("Content-Type", "application/xml")
	if err != nil {
		m.Error += err.Error()
		return Response{}, HttpMetric{m, 0, 0, 400}
	}
	return do(m, c, r)
}
