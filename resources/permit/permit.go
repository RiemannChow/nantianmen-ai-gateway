package permit

import (
	_ "embed"

	"github.com/eolinker/go-common/permit"
	"gopkg.in/yaml.v3"
)

var (

	//go:embed permit.yml
	// data 与 reset 为预留的权限规则引导路径（当前 init 中禁用，保留供后续启用）。
	//nolint:unused
	data []byte
)

func init() {

	//reset()

}
//nolint:unused
func reset() {
	pConfig := make(map[string]map[string][]string)
	err := yaml.Unmarshal(data, &pConfig)
	if err != nil {
		panic(err)
	}
	for group, rules := range pConfig {
		for access, paths := range rules {
			av := permit.FormatAccess(group, access)
			permit.AddPermitRule(av, paths...)
		}
	}

}
