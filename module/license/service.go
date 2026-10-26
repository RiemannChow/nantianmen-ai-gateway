package license

import (
	"context"
	"encoding/json"
	"reflect"
	"sync"

	"github.com/eolinker/go-common/autowire"
	"github.com/eolinker/go-common/utils"

	setting_service "github.com/RiemannChow/nantianmen-ai-gateway/service/setting"
)

const KeyLicensePayload = "system.license.payload"

// ILicenseService is the license entitlement service.
type ILicenseService interface {
	// Get returns the currently loaded license view.
	Get(ctx context.Context) *LicenseView
	// Activate validates and stores a signed license payload (raw JSON string).
	Activate(ctx context.Context, payload string) (*LicenseView, error)
	// Enabled reports whether the given module is currently entitled.
	Enabled(module string) bool
	// Modules returns the entitlement status of all enterprise modules.
	Modules(ctx context.Context) *ModulesView
	// Check returns a LicenseError if the module is not entitled; nil otherwise.
	Check(ctx context.Context, module string) error
}

func init() {
	autowire.Auto[ILicenseService](func() reflect.Value {
		return reflect.ValueOf(new(imlLicenseService))
	})
}

var (
	svc ILicenseService
)

// Check is a package-level convenience for controllers to gate a capability.
// It is safe to call before the service is wired (treated as community).
func Check(ctx context.Context, module string) error {
	if svc == nil {
		return gate(ctx, module, nil)
	}
	return svc.Check(ctx, module)
}

// gate applies the entitlement rules given an optional current license.
func gate(ctx context.Context, module string, l *License) error {
	// Community-only modules are always allowed.
	if _, ok := EnterpriseModules[module]; !ok {
		return nil
	}
	if l == nil || l.Edition != EditionEnterprise || verify(l) != nil {
		return NewLicenseError("")
	}
	if l.IsExpired() {
		return NewLicenseError("企业版 License 已过期，请续费或更新 License")
	}
	if !l.HasModule(module) {
		return NewLicenseError("当前 License 未包含该模块（" + module + "），请升级授权")
	}
	// Node-limit enforcement is a future hook: a production build would count
	// active gateway instances (data-plane nodes) against MaxNodes here.
	return nil
}

type imlLicenseService struct {
	settingService setting_service.ISettingService `autowired:""`

	mu      sync.RWMutex
	current *License
	valid   bool
	loaded  bool
}

func (i *imlLicenseService) OnInit() {
	svc = i
	i.reload(context.Background())
}

// operator returns the acting user id, falling back to "system" when unknown.
func operator(ctx context.Context) string {
	if id := utils.UserId(ctx); id != "" {
		return id
	}
	return "system"
}

func (i *imlLicenseService) reload(ctx context.Context) {
	raw, has := i.settingService.Get(ctx, KeyLicensePayload)
	i.mu.Lock()
	defer i.mu.Unlock()
	if !has || raw == "" {
		i.current = nil
		i.valid = false
		i.loaded = false
		return
	}
	var l License
	if err := json.Unmarshal([]byte(raw), &l); err != nil {
		i.current = nil
		i.valid = false
		i.loaded = true
		return
	}
	if err := verify(&l); err != nil {
		i.current = &l
		i.valid = false
		i.loaded = true
		return
	}
	i.current = &l
	i.valid = true
	i.loaded = true
}

func (i *imlLicenseService) Get(ctx context.Context) *LicenseView {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return toView(i.current, i.valid, i.loaded)
}

func (i *imlLicenseService) Activate(ctx context.Context, payload string) (*LicenseView, error) {
	var l License
	if err := json.Unmarshal([]byte(payload), &l); err != nil {
		return nil, errInvalid("malformed json")
	}
	if err := verify(&l); err != nil {
		return nil, errInvalid(err.Error())
	}
	// Persist the exact payload the user submitted (signature included).
	if err := i.settingService.Set(ctx, KeyLicensePayload, payload, operator(ctx)); err != nil {
		return nil, err
	}
	i.reload(ctx)
	return i.Get(ctx), nil
}

func (i *imlLicenseService) Enabled(module string) bool {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return gate(context.Background(), module, i.current) == nil
}

func (i *imlLicenseService) Modules(ctx context.Context) *ModulesView {
	i.mu.RLock()
	defer i.mu.RUnlock()
	out := make([]ModuleStatus, 0, len(EnterpriseModules))
	for name, label := range EnterpriseModules {
		status := ModuleStatus{Name: name, Label: label}
		if i.current != nil && i.current.Edition == EditionEnterprise && i.valid {
			if i.current.IsExpired() {
				status.Reason = "expired"
			} else if !i.current.HasModule(name) {
				status.Reason = "not_included"
			} else {
				status.Enabled = true
			}
		} else {
			status.Reason = "no_license"
		}
		out = append(out, status)
	}
	return &ModulesView{Modules: out}
}

func (i *imlLicenseService) Check(ctx context.Context, module string) error {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return gate(ctx, module, i.current)
}
