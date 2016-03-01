// Package util is part of the gogrinder load & performance test tool.
// It provides a collection of utilities with the intend to
// ease the implementation of test-secenarios.
//
package util

import (
	"io"
	"math/rand"
	"time"
)

// Inspired from here:
// http://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// randomReader gets number of bytes. It ends with EOF error.
type randReader struct {
	rand.Source
	bytes int // how many bytes are left to read
}

// NewRandReader creates a new random reader with a time source.
func NewRandReader(bytes int) io.Reader {
	return NewRandReaderFrom(rand.NewSource(time.Now().UnixNano()), bytes)
}

// NewRandReaderFrom creates a new reader from your own rand.Source
func NewRandReaderFrom(src rand.Source, bytes int) io.Reader {
	return &randReader{src, bytes}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Read implements io.Reader
func (r *randReader) Read(buf []byte) (n int, err error) {
	if r.bytes <= 0 {
		return 0, io.EOF
	}

	toRead := min(len(buf), r.bytes)
	r.bytes -= toRead
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := toRead-1, r.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = r.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			buf[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return toRead, nil
}
