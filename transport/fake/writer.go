// Copyright 2023 The Outline Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fake

import (
	"io"
)

type fakeWriter struct {
	writer      io.Writer
	prefixBytes int64
	fakeData    []byte
	fakeOffset  int64
}

var _ io.Writer = (*fakeWriter)(nil)

type fakeWriterReaderFrom struct {
	*fakeWriter
	rf io.ReaderFrom
}

var _ io.ReaderFrom = (*fakeWriterReaderFrom)(nil)

// NewWriter creates a [io.Writer] that ensures the byte sequence is split at prefixBytes.
// A write will end right after byte index prefixBytes - 1, before a write starting at byte index prefixBytes.
// For example, if you have a write of [0123456789] and prefixBytes = 3, you will get writes [012] and [3456789].
// If the input writer is a [io.ReaderFrom], the output writer will be too.
func NewWriter(writer io.Writer, prefixBytes int64, fakeData []byte, fakeOffset int64) io.Writer {
	sw := &fakeWriter{writer, prefixBytes, fakeData, fakeOffset}
	if rf, ok := writer.(io.ReaderFrom); ok {
		return &fakeWriterReaderFrom{sw, rf}
	}
	return sw
}

func (w *fakeWriterReaderFrom) ReadFrom(source io.Reader) (int64, error) {
	reader := io.MultiReader(io.LimitReader(source, w.prefixBytes), source)
	written, err := w.rf.ReadFrom(reader)
	w.prefixBytes -= written
	return written, err
}

func (w *fakeWriter) Write(data []byte) (written int, err error) {
	if w.fakeOffset < int64(len(w.fakeData)) {
		n, err := w.writer.Write(w.fakeData[w.fakeOffset:])
		w.fakeOffset += int64(n)
		if err != nil {
			return n, err
		}
	}
	if 0 < w.prefixBytes && w.prefixBytes < int64(len(data)) {
		written, err = w.writer.Write(data[:w.prefixBytes])
		w.prefixBytes -= int64(written)
		if err != nil {
			return written, err
		}
		data = data[written:]
	}
	n, err := w.writer.Write(data)
	written += n
	w.prefixBytes -= int64(n)
	return written, err
}