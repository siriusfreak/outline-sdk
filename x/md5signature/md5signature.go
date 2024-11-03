package md5signature

import (
	"fmt"
	"golang.org/x/sys/unix"
	"net"
	"unsafe"
)

const socketFlag = 14

type signature struct {
	Addr  [16]byte
	Len   uint16
	Flags uint16
	Key   [80]byte
}

func Add(conn *net.TCPConn, remoteAddr string, data string) error {
	ip := net.ParseIP(remoteAddr)
	if ip == nil {
		return fmt.Errorf("invalid remote IP address: %s", remoteAddr)
	}

	address, err := ip.To16().MarshalText()
	if err != nil {
		return fmt.Errorf("failed to marshal IP address: %w", err)
	}

	key := []byte(data)

	sig := signature{
		Addr: [16]byte(address),
		Len:  uint16(len(data)),
		Key:  [80]byte(key),
	}

	if err := setOption(conn, sig); err != nil {
		return fmt.Errorf("failed to set socket option: %w", err)
	}

	return nil
}

func setOption(conn *net.TCPConn, md5sig signature) error {
	file, err := conn.File()
	if err != nil {
		return fmt.Errorf("failed to get file descriptor: %w", err)
	}
	defer file.Close()

	size := unsafe.Sizeof(md5sig)
	buffer := (*[unsafe.Sizeof(md5sig)]byte)(unsafe.Pointer(&md5sig))[:size]
	fd := int(file.Fd())

	err = unix.SetsockoptString(fd, unix.IPPROTO_TCP, socketFlag, string(buffer))
	if err != nil {
		return fmt.Errorf("failed to set TCP_MD5SIG: %w", err)
	}

	return nil
}

func Remove(conn *net.TCPConn) error {
	file, err := conn.File()
	if err != nil {
		return fmt.Errorf("failed to get underlying file descriptor: %w", err)
	}
	defer file.Close()

	fd := int(file.Fd())

	err = unix.SetsockoptString(fd, unix.IPPROTO_TCP, socketFlag, "")
	if err != nil {
		return fmt.Errorf("failed to clear TCP_MD5SIG: %w", err)
	}

	return nil
}
