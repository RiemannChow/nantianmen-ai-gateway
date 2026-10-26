package aisecurity

// KeyAISecurityRules is the setting key holding the AI-security rule config.
const KeyAISecurityRules = "ai_security.rules"

// Moderation action applied when a rule matches.
const (
	ActionFlag = "flag" // 仅标记/拦截（不修改原文）
	ActionMask = "mask" // 脱敏替换
)

// PIIPattern is a single PII-identification rule (regex based).
type PIIPattern struct {
	Name   string `json:"name"`   // 规则名，如「手机号」
	Regex  string `json:"regex"`  // Go 正则
	Action string `json:"action"` // flag | mask
}

// ModeratorConfig points at an external content-moderation endpoint
// (OpenAI moderation compatible or any HTTP service). Optional; when disabled
// detection is purely local/heuristic.
type ModeratorConfig struct {
	Enabled bool              `json:"enabled"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

// AISecurityRules is the persisted rule set for AI-security scanning.
type AISecurityRules struct {
	PIIPatterns       []PIIPattern     `json:"pii_patterns"`
	InjectionKeywords []string         `json:"injection_keywords"`
	ModerationWords   []string         `json:"moderation_words"`
	ExternalModerator *ModeratorConfig `json:"external_moderator,omitempty"`
}

// ScanInput is the body for a scan request.
type ScanInput struct {
	Text  string   `json:"text"`
	Types []string `json:"types"` // 可选：pii / injection / moderation；为空则全部
}

// Finding is one detection result.
type Finding struct {
	Type     string `json:"type"`     // pii / prompt_injection / moderation
	Rule     string `json:"rule"`     // 命中的规则名
	Match    string `json:"match"`    // 命中片段
	Action   string `json:"action"`   // flag | mask
	Position int    `json:"position"` // 在原文中的起始位置
}

// ScanResult aggregates findings for one scan.
type ScanResult struct {
	Safe      bool      `json:"safe"`
	Findings  []Finding `json:"findings"`
	ScannedAt int64     `json:"scanned_at"`
}

// defaultPIIPatterns ship so the scanner is useful with zero configuration.
var defaultPIIPatterns = []PIIPattern{
	{Name: "手机号", Regex: `1[3-9]\d{9}`, Action: ActionMask},
	{Name: "身份证", Regex: `[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]`, Action: ActionMask},
	{Name: "邮箱", Regex: `[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`, Action: ActionMask},
	{Name: "银行卡", Regex: `(?:\d[ ]?){15,19}`, Action: ActionMask},
	{Name: "IPv4", Regex: `\b(?:\d{1,3}\.){3}\d{1,3}\b`, Action: ActionMask},
}

// defaultInjectionKeywords are common prompt-injection fingerprints.
var defaultInjectionKeywords = []string{
	"ignore previous instructions",
	"忽略之前的指令",
	"ignore all previous",
	"忽略以上所有",
	"disregard the system prompt",
	"忽略系统提示",
	"system prompt override",
	"you are now",
	"你现在是一个",
	"reveal your instructions",
	"泄露你的指令",
	"act as an unrestricted",
	"请忘记你的规则",
}

// defaultModerationWords are local policy words flagged by content moderation.
var defaultModerationWords = []string{
	"违反",
}
