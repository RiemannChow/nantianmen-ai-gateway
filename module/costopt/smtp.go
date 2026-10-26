package costopt

import (
	"github.com/eolinker/go-common/cftool"
)

// SMTPConfig is the SMTP server used to deliver email alerts. It is configured
// under the `smtp:` section of config.yml and injected via autowire.
type SMTPConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	From     string `json:"from" yaml:"from"`
}

func init() {
	cftool.Register[SMTPConfig]("smtp")
}
