package util

import (
	"fmt"
	"testing"
)

func TestXmlReader(t *testing.T) {
	xr := XmlReader("bang_0.xml", "record")

	for i := 0; ; i++ {
		if e, ok := <-xr; ok {
			exp := fmt.Sprintf("<sth>%d</sth>", i)
			if e != exp {
				t.Errorf("record expected: %s, but was: %s!", exp, e)
			}
		} else {
			break
		}
	}
}


func TestGzXmlReader(t *testing.T) {
	xr := GzXmlReader("bang_0.xml.gz", "record")

	for i := 0; ; i++ {
		if e, ok := <-xr; ok {
			exp := fmt.Sprintf("<sth>%d</sth>", i)
			if e != exp {
				t.Errorf("record expected: %s, but was: %s!", exp, e)
			}
		} else {
			break
		}
	}
}