package ttl

import (
	"fmt"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"net"
	"net/netip"
)

func Set(conn net.Conn, ttl int) (old int, err error) {
	addr, err := netip.ParseAddrPort(conn.RemoteAddr().String())
	if err != nil {
		return 0, err
	}

	switch {
	case addr.Addr().Is4():
		conn := ipv4.NewConn(conn)
		old, _ = conn.TTL()
		if err := conn.SetTTL(ttl); err != nil {
			return 0, fmt.Errorf("failed to set TTL: %w", err)
		}
	case addr.Addr().Is6():
		conn := ipv6.NewConn(conn)
		old, _ = conn.HopLimit()
		if err := conn.SetHopLimit(ttl); err != nil {
			return 0, fmt.Errorf("failed to set hop limit: %w", err)
		}
	}

	return
}
