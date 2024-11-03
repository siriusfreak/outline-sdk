package fake

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/Jigsaw-Code/outline-sdk/x/packet"
	"io"
	"net"
	"syscall"
	"time"
)

type fakeWriter struct {
	conn       *net.TCPConn
	fd         SocketDescriptor
	writer     io.Writer
	fakeData   []byte
	fakeOffset int64
	fakeBytes  int64
	ttl        int
}

var _ io.Writer = (*fakeWriter)(nil)

type fakeWriterReaderFrom struct {
	*fakeWriter
	rf io.ReaderFrom
}

var _ io.ReaderFrom = (*fakeWriterReaderFrom)(nil)

// NewWriter creates a [io.Writer] that ensures the fake data is written before the real data.
// A write will end right after byte index FakeBytes - 1, before a write starting at byte index FakeBytes.
// For example, if you have a write of [0123456789], FakeData = [abc], FakeOffset = 1, and FakeBytes = 3,
// you will get writes [bc] and [0123456789]. If the input writer is a [io.ReaderFrom], the output writer will be too.
func NewWriter(conn *net.TCPConn, fd SocketDescriptor, writer io.Writer, fakeData []byte, fakeOffset int64, fakeBytes int64, ttl int) io.Writer {
	sw := &fakeWriter{conn, fd, writer, fakeData, fakeOffset, fakeBytes, ttl}
	if rf, ok := writer.(io.ReaderFrom); ok {
		return &fakeWriterReaderFrom{sw, rf}
	}
	return sw
}

func (w *fakeWriterReaderFrom) ReadFrom(source io.Reader) (written int64, err error) {
	panic("implement me")
}

func (w *fakeWriter) Write(data []byte) (written int, err error) {
	fakeData := w.getFakeData(bufio.NewReader(bytes.NewReader(data)))
	if fakeData != nil {
		if err := setsockoptInt(w.fd, syscall.IPPROTO_IP, syscall.IP_TTL, w.ttl); err != nil {
			return written, fmt.Errorf("failed to set TTL before writing fake data: %w", err)
		}
		if err := setSocketLinger(w.fd, 1, 0); err != nil {
			return written, fmt.Errorf("failed to set SO_LINGER before writing fake data: %w", err)
		}
		fmt.Printf("Writing fake data with TTL %d:\n---\n%s\n---\n", w.ttl, fakeData)
		fakeData = append(fakeData, make([]byte, len(data)-len(fakeData))...)
		err := w.send(fakeData, 0)
		written += len(fakeData)
		if err != nil {
			return written, err
		}
		time.Sleep(200 * time.Millisecond)
		//if err := setsockoptInt(w.fd, syscall.IPPROTO_IP, syscall.IP_TTL, 68); err != nil {
		//	err = fmt.Errorf("failed to restore TTL after writing fake data: %w", err)
		//}
		if err := clearSocketLinger(w.fd); err != nil {
			err = fmt.Errorf("failed to restore SO_LINGER after writing fake data: %w", err)
		}
	}
	fmt.Printf("Writing real data:\n---\n%s\n---\n", data)
	if err := w.send(data, 0); err != nil {
		return written, fmt.Errorf("failed to send real data: %w", err)
	}
	written += len(data)
	return written, err
}

func (w *fakeWriter) getFakeData(dataReader *bufio.Reader) []byte {
	fakeData := w.fakeData
	if fakeData == nil {
		isHttp := packet.IsHTTP(dataReader)
		fakeData = getDefaultFakeData(isHttp)
	}
	if w.fakeOffset >= int64(len(fakeData)) {
		return nil
	}
	fakeData = fakeData[w.fakeOffset:]
	if w.fakeBytes < int64(len(fakeData)) {
		fakeData = fakeData[:w.fakeBytes]
	}
	if len(fakeData) == 0 {
		return nil
	}
	return fakeData
}

func (w *fakeWriter) send(data []byte, flags int) error {
	// Use SyscallConn to access the underlying file descriptor safely
	rawConn, err := w.conn.SyscallConn()
	if err != nil {
		return fmt.Errorf("oob strategy was unable to get raw conn: %w", err)
	}

	// Use Control to execute Sendto on the file descriptor
	var sendErr error
	err = rawConn.Control(func(fd uintptr) {
		sendErr = sendTo(SocketDescriptor(fd), data, flags)
	})
	if err != nil {
		return fmt.Errorf("oob strategy was unable to control socket: %w", err)
	}
	if sendErr != nil {
		return fmt.Errorf("oob strategy was unable to send data: %w", sendErr)
	}
	return nil
}
