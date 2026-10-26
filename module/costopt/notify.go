package costopt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"text/template"
	"time"

	"github.com/eolinker/eosc/log"
	"github.com/eolinker/go-common/register"
	"github.com/eolinker/go-common/server"
	"github.com/eolinker/go-common/utils"
)

const notifyCooldown = time.Hour

// OnInit is invoked by the go-common autowire lifecycle once the server is up.
// It launches the background alert-evaluation scheduler.
func (i *imlCostOptService) OnInit() {
	register.Handle(func(v server.Server) {
		ctx, cancel := context.WithCancel(context.Background())
		ctx = utils.SetUserId(ctx, "admin")
		go i.startNotifyScheduler(ctx)
		_ = cancel // scheduler is torn down when ctx is done
	})
}

func (i *imlCostOptService) startNotifyScheduler(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	// evaluate once shortly after startup so fresh installs alert promptly
	time.Sleep(20 * time.Second)
	i.evaluateAndNotify(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			i.evaluateAndNotify(ctx)
		}
	}
}

// evaluateAndNotify builds the active alert signals and dispatches each one whose
// level has worsened since the last dispatch (with a cooldown floor).
func (i *imlCostOptService) evaluateAndNotify(ctx context.Context) {
	cfg := i.GetNotifyConfig(ctx)
	if !cfg.Webhook.Enabled && !cfg.Email.Enabled {
		return
	}
	states := i.getNotifyStates(ctx)

	signals := i.collectSignals(ctx)

	now := time.Now().Unix()
	changed := false
	for _, sig := range signals {
		st := states[sig.key]
		if sig.rank <= 0 {
			continue
		}
		// only fire on worsening or fresh appearance (rank 0 == none/ok)
		if sig.rank > st.LastRank {
			if now < st.CooldownUntil {
				continue
			}
			if err := i.dispatch(ctx, cfg, sig.payload); err != nil {
				log.Errorf("[costopt] dispatch alert %s failed: %v", sig.key, err)
				continue
			}
			st.LastRank = sig.rank
			st.LastSeverity = sig.payload.Severity
			st.LastFiredAt = now
			st.CooldownUntil = now + int64(notifyCooldown.Seconds())
			states[sig.key] = st
			changed = true
			log.Infof("[costopt] fired alert %s (rank %d)", sig.key, sig.rank)
		}
	}
	if changed {
		i.saveNotifyStates(ctx, states)
	}
}

type alertSignal struct {
	key     string
	rank    int
	payload AlertPayload
}

func (i *imlCostOptService) collectSignals(ctx context.Context) []alertSignal {
	ms, now := monthStartNow()
	out := make([]alertSignal, 0, 8)

	// global budget
	if bs, err := i.BudgetStatus(ctx); err == nil && bs.Configured {
		rank := levelRank(bs.Level)
		if rank > 0 {
			out = append(out, alertSignal{
				key:  "budget",
				rank: rank,
				payload: AlertPayload{
					Title:    fmt.Sprintf("预算告警：月度支出%s", budgetLevelLabel(bs.Level)),
					Body:     fmt.Sprintf("本月已花费 %.2f / 预算 %.2f（占比 %.0f%%）。%s", bs.MonthSpend, bs.MonthlyBudget, bs.Ratio*100, bs.Message),
					Severity: bs.Level,
					Source:   "budget",
					Time:     now,
					Data: map[string]interface{}{
						"month_spend": bs.MonthSpend, "monthly_budget": bs.MonthlyBudget,
						"ratio": bs.Ratio, "level": bs.Level,
					},
				},
			})
		}
	}

	// per-consumer budgets
	if cb, err := i.GetConsumerBudgets(ctx); err == nil && cb != nil {
		for id, item := range cb.Items {
			if !item.Enabled {
				continue
			}
			st, err := i.ConsumerBudgetStatus(ctx, id)
			if err != nil || !st.Configured {
				continue
			}
			rank := levelRank(st.Level)
			if rank > 0 {
				out = append(out, alertSignal{
					key:  "consumer:" + id,
					rank: rank,
					payload: AlertPayload{
						Title:    fmt.Sprintf("消费者预算风险：%s %s", id, budgetLevelLabel(st.Level)),
						Body:     fmt.Sprintf("消费者「%s」本月已花费 %.2f / 预算 %.2f（占比 %.0f%%）。", id, st.MonthSpend, st.MonthlyBudget, st.Ratio*100),
						Severity: st.Level,
						Source:   "consumer:" + id,
						Time:     now,
						Data: map[string]interface{}{
							"consumer": id, "month_spend": st.MonthSpend,
							"monthly_budget": st.MonthlyBudget, "ratio": st.Ratio, "level": st.Level,
						},
					},
				})
			}
		}
	}

	// anomalies
	if an, err := i.Anomalies(ctx, ms, now); err == nil {
		for _, a := range an.Anomalies {
			if a.Severity == "info" {
				continue
			}
			rank := severityRank(a.Severity)
			out = append(out, alertSignal{
				key:  "anomaly:" + a.Type,
				rank: rank,
				payload: AlertPayload{
					Title:    a.Title,
					Body:     a.Detail,
					Severity: a.Severity,
					Source:   "anomaly:" + a.Type,
					Time:     now,
					Data:     map[string]interface{}{"type": a.Type, "severity": a.Severity},
				},
			})
		}
	}
	return out
}

func levelRank(l string) int {
	switch l {
	case "warning":
		return 1
	case "exceeded":
		return 2
	default:
		return 0
	}
}

func severityRank(s string) int {
	switch s {
	case "warning":
		return 1
	case "critical":
		return 3
	default:
		return 0
	}
}

func budgetLevelLabel(l string) string {
	switch l {
	case "warning":
		return "已达预警线"
	case "exceeded":
		return "已超预算"
	default:
		return "正常"
	}
}

// dispatch sends the payload to every enabled channel and aggregates errors.
func (i *imlCostOptService) dispatch(ctx context.Context, cfg *NotifyConfig, payload AlertPayload) error {
	var errs []string
	if cfg.Webhook.Enabled {
		if err := i.sendWebhook(cfg.Webhook, payload); err != nil {
			errs = append(errs, "webhook: "+err.Error())
		}
	}
	if cfg.Email.Enabled {
		if err := i.sendEmail(cfg.Email, payload, i.smtpConfig); err != nil {
			errs = append(errs, "email: "+err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return nil
}

// renderTemplate renders tmpl against payload; falls back to a default JSON body.
func renderTemplate(tmpl string, payload AlertPayload) (string, error) {
	if strings.TrimSpace(tmpl) == "" {
		b, err := json.Marshal(payload)
		if err != nil {
			return payload.Body, nil
		}
		return string(b), nil
	}
	t, err := template.New("alert").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, payload); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (i *imlCostOptService) sendWebhook(cfg WebhookChannelConfig, payload AlertPayload) error {
	if cfg.URL == "" {
		return fmt.Errorf("webhook 未配置 URL")
	}
	body, err := renderTemplate(cfg.Template, payload)
	if err != nil {
		return err
	}
	method := strings.ToUpper(cfg.Method)
	if method == "" {
		method = http.MethodPost
	}
	req, err := http.NewRequest(method, cfg.URL, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook 返回 %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

func (i *imlCostOptService) sendEmail(cfg EmailChannelConfig, payload AlertPayload, smtpCfg *SMTPConfig) error {
	if len(cfg.To) == 0 {
		return fmt.Errorf("邮件未配置收件人")
	}
	if smtpCfg == nil || smtpCfg.Host == "" {
		return fmt.Errorf("邮件未配置 SMTP（config.yml 的 smtp: 段）")
	}
	subject, err := renderTemplate(cfg.SubjectTemplate, payload)
	if err != nil {
		return err
	}
	body, err := renderTemplate(cfg.BodyTemplate, payload)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", smtpCfg.From))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(cfg.To, ", ")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(body)

	addr := net.JoinHostPort(smtpCfg.Host, fmt.Sprintf("%d", smtpCfg.Port))
	auth := smtp.PlainAuth("", smtpCfg.Username, smtpCfg.Password, smtpCfg.Host)
	return smtp.SendMail(addr, auth, smtpCfg.From, cfg.To, buf.Bytes())
}

// SendTestNotification fires a test payload through the requested channel so the
// user can verify connectivity from the UI without waiting for a real alert.
func (i *imlCostOptService) SendTestNotification(ctx context.Context, channel string) error {
	cfg := i.GetNotifyConfig(ctx)
	payload := AlertPayload{
		Title:    "南天门 AI 网关 — 告警触达测试",
		Body:     "这是一条来自「AI 成本优化 → 通知配置」的测试告警，说明告警通道配置正常。",
		Severity: "info",
		Source:   "test",
		Time:     time.Now().Unix(),
		Data:     map[string]interface{}{"test": true},
	}
	switch channel {
	case "email":
		if !cfg.Email.Enabled {
			return fmt.Errorf("邮件通道未启用")
		}
		return i.sendEmail(cfg.Email, payload, i.smtpConfig)
	case "webhook", "":
		if !cfg.Webhook.Enabled {
			return fmt.Errorf("Webhook 通道未启用")
		}
		return i.sendWebhook(cfg.Webhook, payload)
	default:
		return fmt.Errorf("未知通道 %s", channel)
	}
}

// ---- notify state persistence ----

func (i *imlCostOptService) getNotifyStates(ctx context.Context) map[string]AlertState {
	raw, has := i.settingService.Get(ctx, KeyNotifyState)
	out := map[string]AlertState{}
	if !has || raw == "" {
		return out
	}
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func (i *imlCostOptService) saveNotifyStates(ctx context.Context, states map[string]AlertState) {
	raw, err := json.Marshal(states)
	if err != nil {
		return
	}
	if err := i.settingService.Set(ctx, KeyNotifyState, string(raw), operator(ctx)); err != nil {
		log.Errorf("[costopt] save notify state failed: %v", err)
	}
}
