// Support PNM format output
// based on https://metacpan.org/source/RATCLIFFE/Sane-0.05/lib/Sane.pm

package sane

import (
	"fmt"
	"io"
)

func (m Image) PNM() []byte {
	f := m.fs[0]
	header := f.PnmHeader()
	data := make([]byte, len(header)+len(f.data))
	copy(data, header)
	copy(data[len(header):], f.data)
	return data
}

func (f *Frame) PnmHeader() []byte {
	const header = "P%d\n# SANE data follows\n%d %d\n%d\n"
	pform := 4
	if f.Depth == 1 {
		p4 := []byte(fmt.Sprintf(header, pform, f.Width, f.Height, 0))
		return p4[:len(p4)-2]
	}
	switch f.Format {
	case FrameRgb, FrameRed, FrameGreen, FrameBlue:
		pform = 6
	case FrameGray:
		pform = 5
	default:
		return nil
	}
	maxval := 255
	if f.Depth > 8 {
		maxval = 65535
	}
	return []byte(fmt.Sprintf(header, pform, f.Width, f.Height, maxval))
}

type Reader struct {
	Scanner *Conn
	buffer  []byte
	i       int64
}

func (c *Conn) NewReader() *Reader {
	p := Reader{Scanner: c}
	return &p
}

func (r *Reader) Next() (err error) {
	r.i = 0
	r.buffer = nil
	f, err := r.Scanner.ReadFrame()
	if err != nil {
		return err
	}
	header := f.PnmHeader()
	if header == nil {
		return fmt.Errorf("unknown frame type %d", f.Format)
	}
	r.buffer = append(r.buffer, header...)
	r.buffer = append(r.buffer, f.data...)
	return
}

func (r *Reader) Read(b []byte) (n int, err error) {
	if r.buffer == nil || r.i >= int64(len(r.buffer)) {
		return 0, io.EOF
	}
	n = copy(b, r.buffer[r.i:])
	r.i += int64(n)
	return
}
