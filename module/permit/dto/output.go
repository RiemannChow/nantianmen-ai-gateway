package permit_dto

import (
	"github.com/RiemannChow/nantianmen-ai-gateway/service/permit-type"
	"strings"
)

type Permission struct {
	Access      string   `json:"access"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Grant       []*Grant `json:"grant"`
}
type Grant = permit_type.Target

type Option = Grant

func SearchOptions(ops []*Option, keyword string) []*Option {
	if keyword == "" {
		return ops
	}
	rs := make([]*Option, 0, len(ops))

	for _, op := range ops {
		if op.Name == keyword || strings.Contains(op.Name, keyword) || op.Label == keyword || strings.Contains(op.Label, keyword) {
			rs = append(rs, op)
		}
	}
	return rs
}
