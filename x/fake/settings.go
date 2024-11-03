package fake

import (
	"fmt"
	"strconv"
	"strings"
)

const defaultTtl = 8

type Settings struct {
	FakeData   []byte
	FakeOffset int64
	FakeBytes  int64
	FakeTtl    int
	Md5Sig     bool
}

// ParseSettings expects the settings to be in the format "[data(str)]:<offset(int)>:<bytes(int)>:<ttl(int)>:<Md5Sig(int=0/1)>"
func ParseSettings(raw string) (*Settings, error) {
	result := Settings{
		FakeTtl: defaultTtl,
	}
	parts := strings.Split(raw, ":")
	if len(parts) > 0 && parts[0] != "" {
		result.FakeData = []byte(parts[0])
	}
	if len(parts) > 1 && parts[1] != "" {
		offset, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse offset: %w", err)
		}
		result.FakeOffset = offset
	}
	if len(parts) > 2 && parts[2] != "" {
		bytes, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse bytes: %w", err)
		}
		result.FakeBytes = bytes
	}
	if len(parts) > 3 && parts[3] != "" {
		ttl, err := strconv.Atoi(parts[3])
		if err != nil {
			return nil, fmt.Errorf("failed to parse TTL: %w", err)
		}
		result.FakeTtl = ttl
	}
	if len(parts) > 4 && parts[4] != "" {
		md5Sig, err := strconv.ParseBool(parts[4])
		if err != nil {
			return nil, fmt.Errorf("failed to parse MD5 signature flag: %w", err)
		}
		result.Md5Sig = md5Sig
	}
	return &result, nil
}
