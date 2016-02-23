package util

import (
	"io/ioutil"
	"testing"
)

func TestRandReader(t *testing.T) {

	r := NewRandReader(2000)
	buf, _ := ioutil.ReadAll(r)
	str := string(buf)

	if len(str) != 2000 {
		t.Errorf("RandReader result length not as exptected: %d", len(str))
	}
}
