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
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/Jigsaw-Code/outline-sdk/transport"
)

type fakeDialer struct {
	dialer     transport.StreamDialer
	splitPoint int64
	fakeData   []byte
	fakeOffset int64
	md5Sig     bool
}

var _ transport.StreamDialer = (*fakeDialer)(nil)

// NewStreamDialer creates a [transport.StreamDialer] that writes "fakeData" in the beginning of the stream and
// then splits the outgoing stream after writing "fakeBytes" bytes using [FakeWriter].
func NewStreamDialer(
	dialer transport.StreamDialer,
	prefixBytes int64,
	fakeData []byte,
	fakeOffset int64,
	md5Sig bool,
) (transport.StreamDialer, error) {
	if dialer == nil {
		return nil, errors.New("argument dialer must not be nil")
	}
	return &fakeDialer{
		dialer:     dialer,
		splitPoint: prefixBytes,
		fakeData:   fakeData,
		fakeOffset: fakeOffset,
		md5Sig:     md5Sig,
	}, nil
}

// DialStream implements [transport.StreamDialer].DialStream.
func (d *fakeDialer) DialStream(ctx context.Context, remoteAddr string) (transport.StreamConn, error) {
	innerConn, err := d.dialer.DialStream(ctx, remoteAddr)
	if err != nil {
		return nil, err
	}
	if d.md5Sig {
		conn := innerConn.(*net.TCPConn)
		err := setMd5Sig(conn, remoteAddr, conn.RemoteAddr().String())
		if err != nil {
			return nil, fmt.Errorf("failed to set MD5 signature: %w", err)
		}
	}
	return transport.WrapConn(innerConn, innerConn, NewWriter(innerConn, d.splitPoint, d.fakeData, d.fakeOffset)), nil
}
