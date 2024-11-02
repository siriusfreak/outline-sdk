// Copyright 2024 The Outline Authors
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

package configurl

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Jigsaw-Code/outline-sdk/transport"
	"github.com/Jigsaw-Code/outline-sdk/transport/fake"
)

func registerFakeStreamDialer(r TypeRegistry[transport.StreamDialer], typeID string, newSD BuildFunc[transport.StreamDialer]) {
	r.RegisterType(typeID, func(ctx context.Context, config *Config) (transport.StreamDialer, error) {
		sd, err := newSD(ctx, config.BaseConfig)
		if err != nil {
			return nil, err
		}
		prefixBytesStr := config.URL.Opaque
		prefixBytes, err := strconv.Atoi(prefixBytesStr)
		if err != nil {
			return nil, fmt.Errorf("prefixBytes is not a number: %v. Fake config should be in fake:<number> format", prefixBytesStr)
		}
		var fakeData []byte  // TODO: Read fake data from the CLI or use a default value (depending on the protocol).
		var fakeOffset int64 // TODO: Read fake offset from the CLI or use a default value (0).
		return fake.NewStreamDialer(sd, int64(prefixBytes), fakeData, fakeOffset)
	})
}
