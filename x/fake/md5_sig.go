package fake

import (
	"fmt"
	"golang.org/x/sys/unix"
	"net"
	"unsafe"
)

const tcpMd5sigFlag = 14

type md5Signature struct {
	Addr  [16]byte
	Len   uint16
	Flags uint16
	Key   [80]byte
}

func setMd5Sig(conn *net.TCPConn, remoteAddr string, data string) error {
	ip := net.ParseIP(remoteAddr)
	if ip == nil {
		return fmt.Errorf("invalid remote IP address: %s", remoteAddr)
	}

	address, err := ip.To16().MarshalText()
	if err != nil {
		return fmt.Errorf("failed to marshal IP address: %v", err)
	}

	key := []byte(data)

	md5sig := md5Signature{
		Addr: [16]byte(address),
		Len:  uint16(len(data)),
		Key:  [80]byte(key),
	}

	if err := setSocketOption(conn, md5sig); err != nil {
		return fmt.Errorf("failed to set socket option: %v", err)
	}

	return nil
}

func setSocketOption(conn *net.TCPConn, md5sig md5Signature) error {
	file, err := conn.File()
	if err != nil {
		return fmt.Errorf("failed to get underlying file descriptor: %v", err)
	}
	defer file.Close()

	size := unsafe.Sizeof(md5sig)

	buffer := (*[unsafe.Sizeof(md5sig)]byte)(unsafe.Pointer(&md5sig))[:size]
	fd := int(file.Fd())

	err = unix.SetsockoptString(fd, unix.IPPROTO_TCP, tcpMd5sigFlag, string(buffer))
	if err != nil {
		return fmt.Errorf("failed to set TCP_MD5SIG: %v", err)
	}

	return nil
}
