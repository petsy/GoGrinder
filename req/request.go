// Package req is part of the GoGrinder load & performance test tool.
// It provides instrumentation for the net/http package and a prometheus reporter.
//
package req

import (
	"bufio"
	//"bytes"
	"encoding/json"
	//"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	//"net/url"
	//"strings"

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

func newMetricReader(readFrom io.Reader) *metricReader {
	// wrap into buffered reader
	return &metricReader{0, time.Now(), gogrinder.Elapsed(0), bufio.NewReader(readFrom)}
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
func DoJson(c *http.Client, r *http.Request, m *gogrinder.Meta) (map[string]interface{}, http.Header, *HttpMetric) {
	hm := &HttpMetric{*m, gogrinder.Elapsed(0), 0, 421} // http status Misdirected Request
	resp, err := c.Do(r)
	if err != nil {
		hm.Error += err.Error()
	}
	if resp != nil {
		defer resp.Body.Close()
		mr := newMetricReader(resp.Body)

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

		hm.FirstByte = mr.firstByteAfter
		hm.Bytes = mr.bytes
		hm.Code = resp.StatusCode
		return doc, resp.Header, hm
	}

	return nil, nil, hm
}

// RAW
func DoRaw(c *http.Client, r *http.Request, m *gogrinder.Meta) ([]byte, http.Header, *HttpMetric) {
	hm := &HttpMetric{*m, gogrinder.Elapsed(0), 0, 421} // http status Misdirected Request

	resp, err := c.Do(r)
	if err != nil {
		hm.Error += err.Error()
	}
	if resp != nil {
		defer resp.Body.Close()
		mr := newMetricReader(resp.Body)

		// read the response body
		raw, err := ioutil.ReadAll(mr)
		if err != nil {
			hm.Error += err.Error()
		}

		hm.FirstByte = mr.firstByteAfter
		hm.Bytes = mr.bytes
		hm.Code = resp.StatusCode
		return raw, resp.Header, hm
	}

	return nil, nil, hm
}

// DOC
func Do(c *http.Client, r *http.Request, m *gogrinder.Meta) (*goquery.Document, http.Header, *HttpMetric) {
	hm := &HttpMetric{*m, gogrinder.Elapsed(0), 0, 421} // http status Misdirected Request
	resp, err := c.Do(r)
	if err != nil {
		hm.Error += err.Error()
	}
	if resp != nil {
		defer resp.Body.Close()
		mr := newMetricReader(resp.Body)

		// read the response body and parse into document
		doc, err := goquery.NewDocumentFromReader(mr)
		if err != nil {
			m.Error += err.Error()
		}

		hm.FirstByte = mr.firstByteAfter
		hm.Bytes = mr.bytes
		hm.Code = resp.StatusCode
		return doc, resp.Header, hm
	}

	return nil, nil, hm
}
