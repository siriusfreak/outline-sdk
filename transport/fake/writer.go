package fake

import (
	"io"
)

type fakeWriter struct {
	writer     io.Writer
	fakeBytes  int64
	fakeData   []byte
	fakeOffset int64
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
func NewWriter(writer io.Writer, fakeBytes int64, fakeData []byte, fakeOffset int64) io.Writer {
	sw := &fakeWriter{writer, fakeBytes, fakeData, fakeOffset}
	if rf, ok := writer.(io.ReaderFrom); ok {
		return &fakeWriterReaderFrom{sw, rf}
	}
	return sw
}

func (w *fakeWriterReaderFrom) ReadFrom(source io.Reader) (int64, error) {
	reader := io.MultiReader(io.LimitReader(source, w.fakeBytes), source)
	written, err := w.rf.ReadFrom(reader)
	w.fakeBytes -= written
	return written, err
}

func (w *fakeWriter) Write(data []byte) (written int, err error) {
	fakeN, err := w.writeFakeData()
	written += fakeN
	if err != nil {
		return fakeN, err
	}
	n, err := w.writer.Write(data)
	written += n
	return written, err
}

func (w *fakeWriter) writeFakeData() (int, error) {
	if w.fakeOffset >= int64(len(w.fakeData)) {
		return 0, nil
	}
	data := w.fakeData[w.fakeOffset:]
	if w.fakeBytes < int64(len(data)) {
		data = data[:w.fakeBytes]
	}
	if len(data) == 0 {
		return 0, nil
	}
	n, err := w.writer.Write(data)
	return n, err
}
