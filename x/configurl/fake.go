package configurl

import (
	"context"
	"fmt"
	"github.com/Jigsaw-Code/outline-sdk/transport"
	"github.com/Jigsaw-Code/outline-sdk/x/fake"
)

func registerFakeStreamDialer(r TypeRegistry[transport.StreamDialer], typeID string, newSD BuildFunc[transport.StreamDialer]) {
	r.RegisterType(typeID, func(ctx context.Context, config *Config) (transport.StreamDialer, error) {
		sd, err := newSD(ctx, config.BaseConfig)
		if err != nil {
			return nil, err
		}
		settings, err := fake.ParseSettings(config.URL.Opaque)
		if err != nil {
			return nil, fmt.Errorf("failed to parse settings: %w", err)
		}
		return fake.NewStreamDialer(sd, settings.FakeData, settings.FakeOffset, settings.FakeBytes, settings.FakeTtl, settings.Md5Sig)
	})
}
