package util

import(
	"testing"
)

func TestRandString(t *testing.T) {

	r := RandString(2000)

	if len(r) != 2000 {
		t.Errorf("RandString length not as exptected: %d", len(r))
	}
}
