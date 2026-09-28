//go:build !init

package main

import (
	_ "github.com/RiemannChow/nantianmen-ai-gateway/resources/access"
	_ "github.com/RiemannChow/nantianmen-ai-gateway/resources/permit"
	_ "github.com/RiemannChow/nantianmen-ai-gateway/resources/plugin"
)

func doCheck() {
}
