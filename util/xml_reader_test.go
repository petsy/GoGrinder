package util

import (
	"fmt"
	"testing"
)

func TestXmlReader(t *testing.T) {
	xr := XmlReader("bang_0.xml", "record")

	for i := 0; i < 10; i++ {
		exp := fmt.Sprintf("<sth>%d</sth>", i)
		e := <-xr
		if e != exp {
			t.Errorf("record expected: %s, but was: %s!", exp, e)
		}
	}
}
