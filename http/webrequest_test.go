package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/finklabs/GoGrinder"
	time "github.com/finklabs/ttime"
)

type testReader struct {
}

func (fb testReader) Read(p []byte) (n int, err error) {
	time.Sleep(55 * time.Millisecond)
	sr := strings.NewReader("markfink")
	return sr.Read(p)
}

func TestFirstByteAfterReader(t *testing.T) {
	time.Freeze(time.Now())
	defer time.Unfreeze()
	tr := testReader{}
	fbr := newMetricReader(tr)

	b1 := make([]byte, 4)
	fbr.Read(b1)

	body := string(b1)
	if !(body == "mark") {
		t.Fatalf("Read buffer was expected '%s', but was: '%v'", "mark", body)
	}
	if fbr.firstByteAfter != 55*time.Millisecond {
		t.Fatalf("First byte was expected after 55 ms but was: %v", fbr.firstByteAfter)
	}

	// read a second time
	b2 := make([]byte, 4)
	fbr.Read(b2)
	body = string(b2)
	if body != "fink" {
		t.Fatalf("Read buffer was expected '%s', but was: '%v'", "fink", body)
	}
}

func TestGoquery(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "<!DOCTYPE html><html><body><h1>My First Heading</h1>"+
			"<p>My first paragraph.</p></body></html>")
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	doc.Find("h1").Each(func(i int, s *goquery.Selection) {
		if s.Text() != "My First Heading" {
			t.Fatalf("Heading was expected '%s', but was: '%s'", "My First Heading", s.Text())
		}
	})
}

func TestGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<!DOCTYPE html><html><body><h1>My First Heading</h1>" +
			"<p>My first paragraph.</p></body></html>"))
	}))
	defer ts.Close()

	m := gogrinder.Meta{Testcase: "sth", Teststep: "else", User: 0, Iteration: 0}
	val, metric := Get(ts.URL)(m)
	if len(metric.(HttpMetric).err) > 0 {
		t.Fatal(metric.(HttpMetric).err)
	}
	resp := val.(Response)

	resp.Doc.Find("html body h1").Each(func(i int, s *goquery.Selection) {
		if s.Text() != "My First Heading" {
			t.Fatalf("Heading was expected '%s', but was: '%s'", "My First Heading", s.Text())
		}
	})
}

func TestGetRaw(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<!DOCTYPE html><html><body><h1>My First Heading</h1>" +
			"<p>My first paragraph.</p></body></html>"))
	}))
	defer ts.Close()

	m := gogrinder.Meta{Testcase: "sth", Teststep: "else", User: 0, Iteration: 0}
	val, metric := GetRaw(ts.URL)(m)
	if len(metric.(HttpMetric).err) > 0 {
		t.Fatal(metric.(HttpMetric).err)
	}
	resp := val.(ResponseRaw)

	if string(resp.Raw) != "<!DOCTYPE html><html><body><h1>My First Heading</h1>"+
		"<p>My first paragraph.</p></body></html>" {
		t.Fatalf("GetRaw response not as expected: '%s'", resp.Raw)
	}
}

func TestGetJson(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id": 1,"name": "A green door","price": 12.50,"tags":` +
		`["home", "green"]}`))
	}))
	defer ts.Close()

	m := gogrinder.Meta{Testcase: "sth", Teststep: "else", User: 0, Iteration: 0}
	val, metric := GetJson(ts.URL)(m)
	if len(metric.(HttpMetric).err) > 0 {
		t.Fatal(metric.(HttpMetric).err)
	}
	resp := val.(ResponseJson)

	id := resp.Json["id"].(float64)
	name := resp.Json["name"].(string)

	if id != 1.0 {
		t.Fatalf("Id was expected '%f', but was: '%f'", 1.0, id)
	}
	if name != "A green door" {
		t.Fatalf("Id was expected '%s', but was: '%s'", "A green door", name)
	}
}
