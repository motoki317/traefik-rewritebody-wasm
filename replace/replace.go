// Package replace contains modified code from github.com/icholy/replace@v0.6.0.
//
// github.com/icholy/replace is licensed under MIT license, and you may view their license at the repository.
package replace

import (
	"bytes"

	"golang.org/x/text/transform"
)

// Transformer replaces text in a stream
// See: http://golang.org/x/text/transform
type Transformer struct {
	transform.NopResetter

	old, new []byte
	oldlen   int

	replaceCount int
	maxCount     int
}

var _ transform.Transformer = (*Transformer)(nil)

// Bytes returns a transformer that replaces all instances of old with new.
// Unlike bytes.Replace, empty old values don't match anything.
func Bytes(old, new []byte, maxCount int) *Transformer {
	return &Transformer{
		old:      old,
		new:      new,
		oldlen:   len(old),
		maxCount: maxCount,
	}
}

// String returns a transformer that replaces all instances of old with new.
// Unlike strings.Replace, empty old values don't match anything.
func String(old, new string, maxCount int) *Transformer {
	return Bytes([]byte(old), []byte(new), maxCount)
}

func (t *Transformer) allowReplacement() bool {
	return t.maxCount <= 0 || t.replaceCount < t.maxCount
}

// Transform implements golang.org/x/text/transform#Transformer
func (t *Transformer) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	var n int
	// don't do anything for empty old string. We're forced to do this because an optimization in
	// transform.String prevents us from generating any output when the src is empty.
	// see: https://github.com/golang/text/blob/master/transform/transform.go#L570-L576
	if t.oldlen == 0 {
		n, err = fullcopy(dst, src)
		return n, n, err
	}
	// replace all instances of old with new
	for t.allowReplacement() {
		i := bytes.Index(src[nSrc:], t.old)
		if i == -1 {
			break
		}
		// copy everything up to the match
		n, err = fullcopy(dst[nDst:], src[nSrc:nSrc+i])
		nSrc += n
		nDst += n
		if err != nil {
			return
		}
		// copy the new value
		n, err = fullcopy(dst[nDst:], t.new)
		if err != nil {
			return
		}
		nDst += n
		nSrc += t.oldlen
		t.replaceCount++
	}
	// if we're at the end, tack on any remaining bytes
	if atEOF {
		n, err = fullcopy(dst[nDst:], src[nSrc:])
		nDst += n
		nSrc += n
		return
	}

	if t.allowReplacement() {
		// skip everything except the trailing len(r.old) - 1
		// we do this becasue there could be a match straddling
		// the boundary
		if skip := len(src[nSrc:]) - t.oldlen + 1; skip > 0 {
			n, err = fullcopy(dst[nDst:], src[nSrc:nSrc+skip])
			nSrc += n
			nDst += n
			if err != nil {
				return
			}
		}
	} else {
		n, err = fullcopy(dst[nDst:], src[nSrc:])
		nSrc += n
		nDst += n
		if err != nil {
			return
		}
	}
	err = transform.ErrShortSrc
	return
}

func fullcopy(dst, src []byte) (n int, err error) {
	n = copy(dst, src)
	if n < len(src) {
		err = transform.ErrShortDst
	}
	return
}
