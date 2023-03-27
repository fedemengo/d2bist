package io

import "io"

type ReaderWithSize interface {
	io.Reader

	// Size return the size in bytes of the reader content
	Size() int
}

type readerWithSize struct {
	io.Reader
	size int
}

func (rs readerWithSize) Size() int {
	return rs.size
}

func NewReaderWithSize(r io.Reader, size int) ReaderWithSize {
	return readerWithSize{
		Reader: r,
		size:   size,
	}
}
