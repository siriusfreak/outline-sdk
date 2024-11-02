package fake

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

// collectWrites is an [io.Writer] that appends each write to the writes slice.
type collectWrites struct {
	writes [][]byte
}

var _ io.Writer = (*collectWrites)(nil)

// Write appends a copy of the data to the writes slice.
func (w *collectWrites) Write(data []byte) (int, error) {
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	w.writes = append(w.writes, dataCopy)
	return len(data), nil
}

// collectReader is an [io.Reader] that appends each Read from the Reader to the reads slice.
type collectReader struct {
	io.Reader
	reads [][]byte
}

func (r *collectReader) Read(buf []byte) (int, error) {
	n, err := r.Reader.Read(buf)
	if n > 0 {
		read := make([]byte, n)
		copy(read, buf[:n])
		r.reads = append(r.reads, read)
	}
	return n, err
}

func TestWrite_FullFake(t *testing.T) {
	var innerWriter collectWrites
	fakeData := []byte("Fake data")   // 9 bytes
	fakeBytes := int64(len(fakeData)) // 9
	fakeOffset := int64(0)
	fakeWriter := NewWriter(&innerWriter, fakeBytes, fakeData, fakeOffset)
	n, err := fakeWriter.Write([]byte("Request")) // 7 bytes
	require.NoError(t, err)
	require.Equal(t, 16, n) // 9 fake + 7 real
	require.Equal(t, [][]byte{[]byte("Fake data"), []byte("Request")}, innerWriter.writes)
}

func TestWrite_PartialFake(t *testing.T) {
	var innerWriter collectWrites
	fakeData := []byte("Fake data") // 9 bytes
	fakeBytes := int64(5)           // Inject first 5 bytes: "Fake "
	fakeOffset := int64(0)
	fakeWriter := NewWriter(&innerWriter, fakeBytes, fakeData, fakeOffset)
	n, err := fakeWriter.Write([]byte("Request")) // 7 bytes
	require.NoError(t, err)
	require.Equal(t, 12, n) // 5 fake + 7 real
	require.Equal(t, [][]byte{[]byte("Fake "), []byte("Request")}, innerWriter.writes)
}

func TestWrite_NoFake(t *testing.T) {
	var innerWriter collectWrites
	fakeData := []byte("Fake data") // 9 bytes
	fakeBytes := int64(0)           // No fake data
	fakeOffset := int64(0)
	fakeWriter := NewWriter(&innerWriter, fakeBytes, fakeData, fakeOffset)
	n, err := fakeWriter.Write([]byte("Request")) // 7 bytes
	require.NoError(t, err)
	require.Equal(t, 7, n) // 0 fake + 7 real
	require.Equal(t, [][]byte{[]byte("Request")}, innerWriter.writes)
}

func TestWrite_WithOffset(t *testing.T) {
	var innerWriter collectWrites
	fakeData := []byte("Fake data") // 9 bytes
	fakeBytes := int64(4)           // Inject 4 bytes starting from offset
	fakeOffset := int64(5)          // fakeData[5:] = "data"
	fakeWriter := NewWriter(&innerWriter, fakeBytes, fakeData, fakeOffset)
	n, err := fakeWriter.Write([]byte("Request")) // 7 bytes
	require.NoError(t, err)
	require.Equal(t, 11, n) // 4 fake + 7 real
	require.Equal(t, [][]byte{[]byte("data"), []byte("Request")}, innerWriter.writes)
}

func TestWrite_NeedsTwoWrites(t *testing.T) {
	var innerWriter collectWrites
	fakeData := []byte("Fake data") // 9 bytes
	fakeBytes := int64(6)           // Inject first 6 bytes: "Fake d"
	fakeOffset := int64(0)
	fakeWriter := NewWriter(&innerWriter, fakeBytes, fakeData, fakeOffset)
	n, err := fakeWriter.Write([]byte("Request")) // 7 bytes
	require.NoError(t, err)
	require.Equal(t, 13, n) // 6 fake + 7 real
	require.Equal(t, [][]byte{[]byte("Fake d"), []byte("Request")}, innerWriter.writes)
}

func TestWrite_Compound(t *testing.T) {
	var innerWriter collectWrites
	// First fakeWriter: fakeBytes=1, fakeData="F"
	fakeData1 := []byte("F")
	fakeBytes1 := int64(1)
	fakeOffset1 := int64(0)
	writer1 := NewWriter(&innerWriter, fakeBytes1, fakeData1, fakeOffset1)

	// Second fakeWriter: fakeBytes=5, fakeData="ake d", fakeOffset=0
	fakeData2 := []byte("ake d") // Total fakeData now: "Fake d"
	fakeBytes2 := int64(5)
	fakeOffset2 := int64(0)
	fakeWriter := NewWriter(writer1, fakeBytes2, fakeData2, fakeOffset2)

	// Write "Request"
	n, err := fakeWriter.Write([]byte("Request")) // 7 bytes
	require.NoError(t, err)
	require.Equal(t, 6, n) // 1 fake + 5 real (Note: total fake data is 6, real data is 7; however, only 5 fakeBytes)
	// Since fakeBytes2=5, fakeData2="ake d", which is 5 bytes.
	require.Equal(t, [][]byte{[]byte("F"), []byte("ake d")}, innerWriter.writes)
}

func TestReadFrom_FullFake(t *testing.T) {
	fakeData := []byte("Fake data") // 9 bytes
	fakeBytes := int64(9)           // Inject all fake data
	fakeOffset := int64(0)
	var buffer bytes.Buffer
	fakeWriter := NewWriter(&buffer, fakeBytes, fakeData, fakeOffset)
	rf, ok := fakeWriter.(io.ReaderFrom)
	require.True(t, ok)

	cr := &collectReader{Reader: bytes.NewReader([]byte("Request"))} // 7 bytes
	n, err := rf.ReadFrom(cr)
	require.NoError(t, err)
	require.Equal(t, int64(16), n) // 9 fake + 7 real
	require.Equal(t, [][]byte{[]byte("Fake data"), []byte("Request")}, cr.reads)
}

func TestReadFrom_PartialFake(t *testing.T) {
	fakeData := []byte("Fake data") // 9 bytes
	fakeBytes := int64(5)           // Inject first 5 bytes: "Fake "
	fakeOffset := int64(0)
	var buffer bytes.Buffer
	fakeWriter := NewWriter(&buffer, fakeBytes, fakeData, fakeOffset)
	rf, ok := fakeWriter.(io.ReaderFrom)
	require.True(t, ok)

	cr := &collectReader{Reader: bytes.NewReader([]byte("Request"))} // 7 bytes
	n, err := rf.ReadFrom(cr)
	require.NoError(t, err)
	require.Equal(t, int64(12), n) // 5 fake + 7 real
	require.Equal(t, [][]byte{[]byte("Fake "), []byte("Request")}, cr.reads)
}

func TestReadFrom_NoFake(t *testing.T) {
	fakeData := []byte("Fake data") // 9 bytes
	fakeBytes := int64(0)           // No fake data
	fakeOffset := int64(0)
	var buffer bytes.Buffer
	fakeWriter := NewWriter(&buffer, fakeBytes, fakeData, fakeOffset)
	rf, ok := fakeWriter.(io.ReaderFrom)
	require.True(t, ok)

	cr := &collectReader{Reader: bytes.NewReader([]byte("Request"))} // 7 bytes
	n, err := rf.ReadFrom(cr)
	require.NoError(t, err)
	require.Equal(t, int64(7), n) // 0 fake + 7 real
	require.Equal(t, [][]byte{[]byte("Request")}, cr.reads)
}

func TestReadFrom_WithOffset(t *testing.T) {
	fakeData := []byte("Fake data") // 9 bytes
	fakeBytes := int64(4)           // Inject 4 bytes starting from offset
	fakeOffset := int64(5)          // fakeData[5:] = "data"
	var buffer bytes.Buffer
	fakeWriter := NewWriter(&buffer, fakeBytes, fakeData, fakeOffset)
	rf, ok := fakeWriter.(io.ReaderFrom)
	require.True(t, ok)

	cr := &collectReader{Reader: bytes.NewReader([]byte("Request"))} // 7 bytes
	n, err := rf.ReadFrom(cr)
	require.NoError(t, err)
	require.Equal(t, int64(11), n) // 4 fake + 7 real
	require.Equal(t, [][]byte{[]byte("data"), []byte("Request")}, cr.reads)
}

func TestReadFrom_NeedsTwoReads(t *testing.T) {
	fakeData := []byte("Fake data") // 9 bytes
	fakeBytes := int64(6)           // Inject first 6 bytes: "Fake d"
	fakeOffset := int64(0)
	var buffer bytes.Buffer
	fakeWriter := NewWriter(&buffer, fakeBytes, fakeData, fakeOffset)
	rf, ok := fakeWriter.(io.ReaderFrom)
	require.True(t, ok)

	// First ReadFrom with "Request1" (8 bytes)
	cr1 := &collectReader{Reader: bytes.NewReader([]byte("Request1"))} // 8 bytes
	n1, err1 := rf.ReadFrom(cr1)
	require.NoError(t, err1)
	require.Equal(t, int64(6+8), n1) // 6 fake + 8 real
	require.Equal(t, [][]byte{[]byte("Fake d"), []byte("Request1")}, cr1.reads)

	// Second ReadFrom with "Request2" (8 bytes)
	cr2 := &collectReader{Reader: bytes.NewReader([]byte("Request2"))} // 8 bytes
	n2, err2 := rf.ReadFrom(cr2)
	require.NoError(t, err2)
	require.Equal(t, int64(6+8), n2) // 6 fake + 8 real
	require.Equal(t, [][]byte{[]byte("Fake d"), []byte("Request2")}, cr2.reads)
}

func TestReadFrom_Compound(t *testing.T) {
	var innerWriter collectWrites
	// First fakeWriter: fakeBytes=3, fakeData="Fake "
	fakeData1 := []byte("Fake ")
	fakeBytes1 := int64(3)
	fakeOffset1 := int64(0)
	writer1 := NewWriter(&innerWriter, fakeBytes1, fakeData1, fakeOffset1)

	// Second fakeWriter: fakeBytes=5, fakeData="data", fakeOffset=0
	fakeData2 := []byte("data")
	fakeBytes2 := int64(5)
	fakeOffset2 := int64(0)
	fakeWriter := NewWriter(writer1, fakeBytes2, fakeData2, fakeOffset2)

	// Wrap the collectWrites with ReadFrom capability
	var buffer bytes.Buffer
	fakeWriterFinal := NewWriter(&buffer, 0, nil, 0) // No additional fake data
	fakeWriterFinal = fakeWriter                     // Overwrite to use the second fakeWriter

	rf, ok := fakeWriterFinal.(io.ReaderFrom)
	require.True(t, ok)

	cr := &collectReader{Reader: bytes.NewReader([]byte("Request"))} // 7 bytes
	n, err := rf.ReadFrom(cr)
	require.NoError(t, err)
	require.Equal(t, int64(3+4+7), n) // 3 fake1 + 4 fake2 +7 real =14
	require.Equal(t, [][]byte{[]byte("Fake "), []byte("data"), []byte("Request")}, innerWriter.writes)
}

func TestWrite_WithOffsetBeyondFakeData(t *testing.T) {
	var innerWriter collectWrites
	fakeData := []byte("Fake data") // 9 bytes
	fakeBytes := int64(4)           // Attempt to inject 4 bytes
	fakeOffset := int64(10)         // Offset beyond fakeData length
	fakeWriter := NewWriter(&innerWriter, fakeBytes, fakeData, fakeOffset)
	n, err := fakeWriter.Write([]byte("Request")) // 7 bytes
	require.NoError(t, err)
	require.Equal(t, 7, n) // 0 fake + 7 real
	require.Equal(t, [][]byte{[]byte("Request")}, innerWriter.writes)
}

func TestReadFrom_WithOffsetBeyondFakeData(t *testing.T) {
	fakeData := []byte("Fake data") // 9 bytes
	fakeBytes := int64(5)           // Attempt to inject 5 bytes
	fakeOffset := int64(10)         // Offset beyond fakeData length
	var buffer bytes.Buffer
	fakeWriter := NewWriter(&buffer, fakeBytes, fakeData, fakeOffset)
	rf, ok := fakeWriter.(io.ReaderFrom)
	require.True(t, ok)

	cr := &collectReader{Reader: bytes.NewReader([]byte("Request"))} // 7 bytes
	n, err := rf.ReadFrom(cr)
	require.NoError(t, err)
	require.Equal(t, int64(7), n) // 0 fake + 7 real
	require.Equal(t, [][]byte{[]byte("Request")}, cr.reads)
}

func BenchmarkReadFrom(b *testing.B) {
	fakeData := []byte("Fake data") // 9 bytes
	fakeBytes := int64(5)           // Inject first 5 bytes: "Fake "
	fakeOffset := int64(0)
	for n := 0; n < b.N; n++ {
		reader := bytes.NewReader([]byte("Request"))
		var buffer bytes.Buffer
		fakeWriter := NewWriter(&buffer, fakeBytes, fakeData, fakeOffset)
		rf, ok := fakeWriter.(io.ReaderFrom)
		if !ok {
			b.Fatalf("Writer does not implement io.ReaderFrom")
		}
		_, err := rf.ReadFrom(reader)
		if err != nil && err != io.EOF {
			b.Fatalf("ReadFrom failed: %v", err)
		}
	}
}
