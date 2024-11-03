//go:build linux || darwin

package fake

import (
	"fmt"
	"net"
	"syscall"
)

type SocketDescriptor int

func setsockoptInt(fd SocketDescriptor, level, opt int, value int) error {
	fmt.Printf("setsockoptInt: %x, %x, %d to value %d\n", fd, level, opt, value)
	return syscall.SetsockoptInt(int(fd), level, opt, value)
}

func setSocketLinger(fd SocketDescriptor, onoff int32, linger int32) error {
	fmt.Printf("setSocketLinger: %d, %d\n", onoff, linger)
	return syscall.SetsockoptLinger(int(fd), syscall.SOL_SOCKET, syscall.SO_LINGER, &syscall.Linger{
		Onoff:  onoff,
		Linger: linger,
	})
}

func clearSocketLinger(fd SocketDescriptor) error {
	fmt.Printf("clearSocketLinger\n")
	return syscall.SetsockoptLinger(int(fd), syscall.SOL_SOCKET, syscall.SO_LINGER, nil)
}

func sendTo(fd SocketDescriptor, data []byte, flags int) (err error) {
	fmt.Printf("sendTo: %d, %v, %d\n", fd, data, flags)
	return syscall.Sendto(int(fd), data, flags, nil)
}

func getSocketDescriptor(conn *net.TCPConn) (SocketDescriptor, error) {
	file, err := conn.File()
	if err != nil {
		return 0, err
	}
	return SocketDescriptor(file.Fd()), nil
}
