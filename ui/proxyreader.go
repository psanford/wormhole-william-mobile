package ui

import (
	"io"
)

type countReader struct {
	count       int64
	countUpdate chan int64
	r           io.Reader
}

func newCountReader(r io.Reader) *countReader {
	return &countReader{
		countUpdate: make(chan int64, 1),
		r:           r,
	}
}

func (cr *countReader) Read(p []byte) (int, error) {
	n, err := cr.r.Read(p)
	cr.count += int64(n)
	select {
	case cr.countUpdate <- cr.count:
	default:
	}
	return n, err
}

func (cr *countReader) Close() error {
	close(cr.countUpdate)
	return nil
}
