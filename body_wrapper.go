package main

import (
	"io"

	"github.com/http-wasm/http-wasm-guest-tinygo/handler/api"
)

// BodyReader wraps a WASI Body to implement io.Reader
type BodyReader struct {
	body api.Body
	eof  bool
}

// NewBodyReader creates a new BodyReader from a WASI Body
func NewBodyReader(body api.Body) *BodyReader {
	return &BodyReader{body: body}
}

// Read implements io.Reader interface
func (br *BodyReader) Read(p []byte) (n int, err error) {
	if br.eof {
		return 0, io.EOF
	}

	size, eof := br.body.Read(p)
	br.eof = eof

	if eof && size == 0 {
		return 0, io.EOF
	}

	return int(size), nil
}

// BodyWriter wraps a WASI Body to implement io.Writer
type BodyWriter struct {
	body api.Body
}

// NewBodyWriter creates a new BodyWriter from a WASI Body
func NewBodyWriter(body api.Body) *BodyWriter {
	return &BodyWriter{body: body}
}

// Write implements io.Writer interface
func (bw *BodyWriter) Write(p []byte) (n int, err error) {
	bw.body.Write(p)
	return len(p), nil
}
