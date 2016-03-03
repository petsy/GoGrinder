package req

import (
	"net/http"
	"encoding/json"
	"bytes"
)

// some experimental features
// ! careful this might be changed or go away completely
// if you are looking for a more stable API please use Do, DoRaw, DoJson
// Due to the experimental character of brackets it is lacking some testing
// and documentation.

func NewPostJsonRequest(url string, msg map[string]interface{}) (*http.Request, error) {
	b, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	r, err := http.NewRequest("POST", url, bytes.NewReader(b))
	r.Header.Set("Content-Type", "application/json")
	return r, err
}

func NewPutJsonRequest(url string, msg map[string]interface{}) (*http.Request, error) {
	b, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	r, err := http.NewRequest("PUT", url, bytes.NewReader(b))
	r.Header.Set("Content-Type", "application/json")
	return r, err
}
