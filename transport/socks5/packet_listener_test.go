package socks5

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"

	"github.com/Jigsaw-Code/outline-sdk/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/things-go/go-socks5"
)

func TestSOCKS5Associate(t *testing.T) {
	// Create a local listener
	// This creates a UDP server that responded to "ping"
	// message with "pong" response
	locIP := net.ParseIP("127.0.0.1")
	// Create a local listener
	echoServerAddr := &net.UDPAddr{IP: locIP, Port: 0}
	echoServer := setupUDPEchoServer(t, echoServerAddr)
	defer echoServer.Close()

	// Create a socks server to proxy "ping" message
	cator := socks5.UserPassAuthenticator{Credentials: socks5.StaticCredentials{
		"testusername": "testpassword",
	}}
	proxySrv := socks5.NewServer(
		socks5.WithAuthMethods([]socks5.Authenticator{cator}),
	)

	// Create SOCKS5 proxy on localhost with a random port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	proxyServerAddress := listener.Addr().String()

	go func() {
		err := proxySrv.Serve(listener)
		defer listener.Close()
		require.NoError(t, err)
	}()

	// Wait until the server is ready.
	ready := make(chan bool, 1)
	go waitUntilServerReady(proxyServerAddress, ready)
	<-ready

	// Connect to local proxy, auth and start the PacketConn.
	client, err := NewClient(&transport.TCPEndpoint{Address: proxyServerAddress})
	require.NotNil(t, client)
	require.NoError(t, err)
	err = client.SetCredentials([]byte("testusername"), []byte("testpassword"))
	require.NoError(t, err)
	client.EnablePacket(&transport.UDPDialer{})
	conn, err := client.ListenPacket(context.Background())
	require.NoError(t, err)
	defer conn.Close()

	// Send "ping" message.
	_, err = conn.WriteTo([]byte("ping"), echoServer.LocalAddr())
	require.NoError(t, err)
	// Max wait time for response.
	conn.SetDeadline(time.Now().Add(time.Second))
	response := make([]byte, 1024)
	n, addr, err := conn.ReadFrom(response)
	require.Equal(t, echoServer.LocalAddr().String(), addr.String())
	require.NoError(t, err)
	require.Equal(t, []byte("pong"), response[:n])
}

func TestUDPLoopBack(t *testing.T) {
	// Create a local listener.
	locIP := net.ParseIP("127.0.0.1")
	echoServerAddr := &net.UDPAddr{IP: locIP, Port: 0}
	echoServer := setupUDPEchoServer(t, echoServerAddr)
	defer echoServer.Close()

	packDialer := transport.UDPDialer{}
	conn, err := packDialer.DialPacket(context.Background(), echoServer.LocalAddr().String())
	require.NoError(t, err)
	conn.Write([]byte("ping"))
	response := make([]byte, 1024)
	n, err := conn.Read(response)
	require.NoError(t, err)
	assert.Equal(t, []byte("pong"), response[:n])
}

func setupUDPEchoServer(t *testing.T, serverAddr *net.UDPAddr) *net.UDPConn {
	server, err := net.ListenUDP("udp", serverAddr)
	require.NoError(t, err)
	go func() {
		buf := make([]byte, 2048)
		for {
			n, remote, err := server.ReadFrom(buf)
			if err != nil {
				return
			}
			if bytes.Equal(buf[:n], []byte("ping")) {
				server.WriteTo([]byte("pong"), remote)
			}
		}
	}()

	t.Cleanup(func() {
		server.Close()
	})

	return server
}

func waitUntilServerReady(addr string, ready chan<- bool) {
	for {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			conn.Close()
			ready <- true
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}
