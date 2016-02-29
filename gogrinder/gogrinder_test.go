package gogrinder

import ()

// Common test utils

// someMetric is used in multiple tests e.g. TestEventReporterUpdateWithSomeMetric
type someMetric struct {
	Meta     // std. GoGrinder metric info
	Code int `json:"status"` // http status code
}

/*
func (m someMetric) MarshalJSON() ([]byte, error) {
	// explicit marshaling of ts and elapsed!
	// from here: http://choly.ca/post/go-json-marshalling/
	type Alias someMetric
	return json.Marshal(&struct {
		Elapsed []byte `json:"elapsed"`
		Alias
	}{
		Elapsed: strconv.AppendFloat(nil, float64(m.Elapsed) /
			float64(time.Millisecond), 'f', 6, 64),
		//Elapsed: "moin",
		Alias:    (Alias)(m),
	})
}
*/

// GoGrinder tests
