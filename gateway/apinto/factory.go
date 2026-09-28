package apinto

import (
	"github.com/RiemannChow/nantianmen-ai-gateway/gateway"
)

func init() {
	gateway.Register("apinto", &Factory{})
}

var _ gateway.IFactory = &Factory{}

type Factory struct {
}

func (f *Factory) Create(config *gateway.ClientConfig) (gateway.IClientDriver, error) {
	return NewClientDriver(config)
}
