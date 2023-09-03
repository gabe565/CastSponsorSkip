package device

import (
	"context"
)

type Option func(device *Device)

func WithContext(ctx context.Context) Option {
	return func(device *Device) {
		device.ctx, device.cancel = context.WithCancel(ctx)
	}
}
