package filestore

import (
	"bytes"
	"io"
)

type Reader struct {
	r   io.ReaderAt
	pos int
	err error
	buf []byte
}

func NewReader(r io.ReaderAt, pos int) *Reader {
	return &Reader{r: r, pos: pos}
}

func (s *Reader) read() {
	if s.pos == 0 {
		s.err = io.EOF
		return
	}
	size := 1024
	if size > s.pos {
		size = s.pos
	}
	s.pos -= size
	buf2 := make([]byte, size, size+len(s.buf))

	// ReadAt attempts to read full buff!
	_, s.err = s.r.ReadAt(buf2, int64(s.pos))
	if s.err == nil {
		s.buf = append(buf2, s.buf...)
	}
}

func (s *Reader) LastLine() (line string, err error) {
	for {
		line, err = s.LineReversed()
		if err != nil {
			if err == io.EOF {
				return "", nil
			}
			return "", err
		}
		if len(line) == 0 {
			continue
		}
		return line, err
	}
}

func (s *Reader) LineReversed() (line string, err error) {
	if s.err != nil {
		return "", s.err
	}
	for {
		lineStart := bytes.LastIndexByte(s.buf, '\n')
		if lineStart >= 0 {
			// We have a complete line:
			var line string
			line, s.buf = string(dropCR(s.buf[lineStart+1:])), s.buf[:lineStart]
			return line, nil
		}
		// Need more data:
		s.read()
		if s.err != nil {
			if s.err == io.EOF {
				if len(s.buf) > 0 {
					return string(dropCR(s.buf)), nil
				}
			}
			return "", s.err
		}
	}
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}
