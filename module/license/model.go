package license

import (
	"fmt"
	"strings"
	"time"
)

// Edition values.
const (
	EditionCommunity  = "community"
	EditionEnterprise = "enterprise"
)

// EnterpriseModules lists the capabilities gated behind a commercial license.
// Community edition has none of these enabled.
var EnterpriseModules = map[string]string{
	"advanced_policy":   "高级策略（精细限流/熔断/灰度）",
	"sso":               "SSO / LDAP 单点登录",
	"audit":             "审计与合规报表",
	"multi_cluster":     "多集群容备",
	"sla":               "SLA 支持",
	"cost_optimization": "AI 成本优化与增值分析",
	"ai_security":       "AI 安全与内容防护",
	"tracing":           "链路追踪与 LLM 可观测性",
	"multi_tenant":      "多租户隔离与计量分离",
}

// License is the signed commercial entitlement document.
type License struct {
	Edition   string   `json:"edition"`
	Customer  string   `json:"customer"`
	Modules   []string `json:"modules"`
	MaxNodes  int      `json:"max_nodes"` // 0 = unlimited
	IssuedAt  int64    `json:"issued_at"`
	Expiry    int64    `json:"expiry"` // unix seconds, 0 = never expire
	SLA       string   `json:"sla,omitempty"`
	Signature string   `json:"signature"`
}

// signable is the canonical, signature-covered subset of a License.
type signable struct {
	Edition  string   `json:"edition"`
	Customer string   `json:"customer"`
	Modules  []string `json:"modules"`
	MaxNodes int      `json:"max_nodes"`
	IssuedAt int64    `json:"issued_at"`
	Expiry   int64    `json:"expiry"`
	SLA      string   `json:"sla,omitempty"`
}

func (l *License) signableCopy() *signable {
	return &signable{
		Edition:  l.Edition,
		Customer: l.Customer,
		Modules:  l.Modules,
		MaxNodes: l.MaxNodes,
		IssuedAt: l.IssuedAt,
		Expiry:   l.Expiry,
		SLA:      l.SLA,
	}
}

// IsExpired reports whether the license is past its expiry (0 expiry = never).
func (l *License) IsExpired() bool {
	if l.Expiry <= 0 {
		return false
	}
	return time.Now().Unix() > l.Expiry
}

// HasModule reports whether the license grants the given module.
func (l *License) HasModule(module string) bool {
	for _, m := range l.Modules {
		if m == module {
			return true
		}
	}
	return false
}

// LicenseError is returned by Check when a gated capability is not entitled.
// It implements pm3.IApiError so the framework serializes it as a domain error.
type LicenseError struct {
	code int
	msg  string
}

func (e *LicenseError) Error() string   { return e.msg }
func (e *LicenseError) Code() int       { return e.code }
func (e *LicenseError) Message() string { return e.msg }
func (e *LicenseError) Success() any    { return "fail" }

// NewLicenseError builds a license-required error with the given human message.
func NewLicenseError(msg string) *LicenseError {
	if msg == "" {
		msg = "该功能需要商业版授权，请激活南天门 AI 网关企业版 License"
	}
	return &LicenseError{code: 403, msg: msg}
}

// LicenseView is the API-facing projection of a License (plus validity flags).
type LicenseView struct {
	Edition  string   `json:"edition"`
	Customer string   `json:"customer"`
	Modules  []string `json:"modules"`
	MaxNodes int      `json:"max_nodes"`
	IssuedAt int64    `json:"issued_at"`
	Expiry   int64    `json:"expiry"`
	SLA      string   `json:"sla,omitempty"`
	Valid    bool     `json:"valid"`
	Loaded   bool     `json:"loaded"`
}

// ModuleStatus describes one enterprise module's entitlement state.
type ModuleStatus struct {
	Name    string `json:"name"`
	Label   string `json:"label"`
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason,omitempty"`
}

// ModulesView is returned by the modules status endpoint.
type ModulesView struct {
	Modules []ModuleStatus `json:"modules"`
}

func toView(l *License, valid, loaded bool) *LicenseView {
	if l == nil {
		return &LicenseView{Edition: EditionCommunity, Valid: valid, Loaded: loaded}
	}
	return &LicenseView{
		Edition:  l.Edition,
		Customer: l.Customer,
		Modules:  l.Modules,
		MaxNodes: l.MaxNodes,
		IssuedAt: l.IssuedAt,
		Expiry:   l.Expiry,
		SLA:      l.SLA,
		Valid:    valid,
		Loaded:   loaded,
	}
}

// errInvalid is a generic error for malformed/invalid license payloads.
func errInvalid(detail string) error {
	return fmt.Errorf("invalid license: %s", detail)
}

// SLADetail is the presentational projection of an SLA tier sold with the
// enterprise license. The license only carries a level code/label string; the
// concrete commitments (availability target, response time, support channel)
// are derived from a fixed tier table so they never need to be re-signed.
type SLADetail struct {
	Level        string `json:"level"`         // normalized code: standard | gold | platinum | custom
	LevelLabel   string `json:"level_label"`   // 支持等级
	Availability string `json:"availability"`  // 可用性目标
	ResponseTime string `json:"response_time"` // 响应时效
	Support      string `json:"support"`       // 支持方式
	Description  string `json:"description"`
}

// slaTiers is the fixed table mapping a normalized level code to its commitments.
var slaTiers = map[string]SLADetail{
	"standard": {
		Level:        "standard",
		LevelLabel:   "标准支持",
		Availability: "99.5%",
		ResponseTime: "4 小时（工作时间）",
		Support:      "工单 / 社区",
		Description:  "适用于非关键业务，提供标准工单与社区支持。",
	},
	"gold": {
		Level:        "gold",
		LevelLabel:   "金牌支持",
		Availability: "99.9%",
		ResponseTime: "1 小时",
		Support:      "7×24 工单 + 即时通讯",
		Description:  "适用于生产环境，提供 7×24 响应与更高可用性保障。",
	},
	"platinum": {
		Level:        "platinum",
		LevelLabel:   "白金支持",
		Availability: "99.99%",
		ResponseTime: "15 分钟",
		Support:      "7×24 + 专属客户成功经理",
		Description:  "适用于核心关键业务，提供最高可用性与专属保障。",
	},
}

// slaDefault is returned when the license's SLA string does not match a known
// tier (e.g. a bespoke contract label). It keeps the projection non-nil.
var slaDefault = SLADetail{
	Level:        "custom",
	LevelLabel:   "定制支持",
	Availability: "按合同约定",
	ResponseTime: "按合同约定",
	Support:      "联系厂商",
	Description:  "该 License 包含定制 SLA，具体承诺以商务合同为准。",
}

// normalizeSLALevel maps a free-form SLA string (which may be a code like "gold"
// or a human label like "Gold (99.9%)") to a known tier code via case-insensitive
// keyword matching, so legacy licenses keep resolving after the field semantics
// tightened.
func normalizeSLALevel(sla string) string {
	lower := strings.ToLower(strings.TrimSpace(sla))
	switch {
	case strings.Contains(lower, "platinum"):
		return "platinum"
	case strings.Contains(lower, "gold"):
		return "gold"
	case strings.Contains(lower, "standard"):
		return "standard"
	default:
		return ""
	}
}

// ResolveSLA derives a presentational SLA detail from the license's SLA string.
// It never returns nil; unknown labels fall back to slaDefault.
func ResolveSLA(sla string) *SLADetail {
	if code := normalizeSLALevel(sla); code != "" {
		d := slaTiers[code]
		// carry the raw label forward if it is more descriptive than the tier name
		return &d
	}
	d := slaDefault
	if strings.TrimSpace(sla) != "" {
		d.LevelLabel = strings.TrimSpace(sla)
	}
	return &d
}
