package utils

import "unsafe"

type Arena struct {
	b []byte
}

func NewArena(bytes int) *Arena {
	return &Arena{
		b: make([]byte, 0, bytes),
	}
}

func (a *Arena) PutString(s string) string {
	if s == "" {
		return ""
	}
	start := len(a.b)
	if cap(a.b) < len(a.b)+len(s) {
		a.b = append(a.b, s...)
	} else {
		a.b = a.b[:start+len(s)]
	}
	buf := a.b[start : start+len(s)]
	copy(buf, s)
	return unsafe.String(unsafe.SliceData(buf), len(s))
}

func (a *Arena) PutBytes(bytes []byte) []byte {
	if len(bytes) == 0 {
		return bytes
	}
	start := len(a.b)
	if cap(a.b) < len(a.b)+len(bytes) {
		a.b = append(a.b, bytes...)
	} else {
		a.b = a.b[:start+len(bytes)]
	}
	buf := a.b[start : start+len(bytes)]
	copy(buf, bytes)
	return buf
}
