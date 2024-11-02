package fake

import (
	"bytes"
	"io"
)

type fakeWriter struct {
	writer     io.Writer
	fakeBytes  int64
	fakeData   []byte
	fakeOffset int64
	ttl        int64
}

var _ io.Writer = (*fakeWriter)(nil)

type fakeWriterReaderFrom struct {
	*fakeWriter
	rf io.ReaderFrom
}

var _ io.ReaderFrom = (*fakeWriterReaderFrom)(nil)

// NewWriter creates a [io.Writer] that ensures the fake data is written before the real data.
// A write will end right after byte index fakeBytes - 1, before a write starting at byte index fakeBytes.
// For example, if you have a write of [0123456789], fakeData = [abc], fakeOffset = 1, and fakeBytes = 3,
// you will get writes [bc] and [0123456789]. If the input writer is a [io.ReaderFrom], the output writer will be too.
func NewWriter(writer io.Writer, fakeBytes int64, fakeData []byte, fakeOffset int64, fakeTtl int64) io.Writer {
	sw := &fakeWriter{writer, fakeBytes, fakeData, fakeOffset, fakeTtl}
	if rf, ok := writer.(io.ReaderFrom); ok {
		return &fakeWriterReaderFrom{sw, rf}
	}
	return sw
}

func (w *fakeWriterReaderFrom) ReadFrom(source io.Reader) (written int64, err error) {
	fakeData := w.getFakeData()
	if fakeData != nil {
		fakeN, err := w.rf.ReadFrom(bytes.NewReader(fakeData))
		written += fakeN
		if err != nil {
			return written, err
		}
	}
	reader := io.MultiReader(io.LimitReader(source, w.fakeBytes), source)
	n, err := w.rf.ReadFrom(reader)
	written += n
	return written, err
}

func (w *fakeWriter) Write(data []byte) (written int, err error) {
	fakeData := w.getFakeData()
	if fakeData != nil {
		fakeN, err := w.writer.Write(fakeData)
		written += fakeN
		if err != nil {
			return written, err
		}
	}
	n, err := w.writer.Write(data)
	written += n
	return written, err
}

func (w *fakeWriter) getFakeData() []byte {
	if w.fakeOffset >= int64(len(w.fakeData)) {
		return nil
	}
	data := w.fakeData[w.fakeOffset:]
	if w.fakeBytes < int64(len(data)) {
		data = data[:w.fakeBytes]
	}
	if len(data) == 0 {
		return nil
	}
	return data
}
