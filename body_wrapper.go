package main

import (
	"io"
)

// Body interface represents the WASI Body interface
type Body interface {
	WriteTo(io.Writer) (size uint64, err error)
	Read([]byte) (size uint32, eof bool)
	Write([]byte)
	WriteString(string)
}

// BodyReader wraps a WASI Body to implement io.Reader
type BodyReader struct {
	body Body
	eof  bool
}

// NewBodyReader creates a new BodyReader from a WASI Body
func NewBodyReader(body Body) *BodyReader {
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
	body Body
}

// NewBodyWriter creates a new BodyWriter from a WASI Body
func NewBodyWriter(body Body) *BodyWriter {
	return &BodyWriter{body: body}
}

// Write implements io.Writer interface
func (bw *BodyWriter) Write(p []byte) (n int, err error) {
	bw.body.Write(p)
	return len(p), nil
}
