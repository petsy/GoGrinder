package http

import (
	"net/http"
	//"golang.org/x/net/html"
	"bufio"

	"github.com/finklabs/GoGrinder"
	time "github.com/finklabs/ttime"
	"io"
	"gopkg.in/xmlpath.v2"
)

// Assemble Reader from bufio that measures time until first byte
type metricReader struct {
	bytes int
	start time.Time
	firstByteAfter time.Duration
	readFrom *bufio.Reader
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

// Get returns a HTML Tokenizer.
func Get(url string) func(gogrinder.Meta) (interface{}, gogrinder.Metric) {
	return func(m gogrinder.Meta) (interface{}, gogrinder.Metric) {
		start := time.Now()
		resp, err := http.Get(url)
		defer resp.Body.Close()
		mr := newMetricReader(resp.Body)

		// read the response body and parse into document
		//t := html.NewTokenizer(mr)
		t, err := xmlpath.Parse(mr)

		m.Elapsed = time.Now().Sub(start)
		//return make(map[string]string), HttpMetric{m, mr.firstByteAfter, mr.bytes,
		//	resp.StatusCode, err.Error()}
		return t, HttpMetric{m, mr.firstByteAfter, mr.bytes,
			resp.StatusCode, err.Error()}
	}
}



// TODO
// * do this testdriven
// * start with tockenizer!!!
// * create a special reader to report first byte
// * create versions for json and raw, too (GetJson, GetRaw)
// * complete this with POST, PUT, DELETE








