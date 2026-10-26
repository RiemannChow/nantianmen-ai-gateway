package setting

import (
	"context"

	"github.com/RiemannChow/nantianmen-ai-gateway/common/crypto"
	"github.com/RiemannChow/nantianmen-ai-gateway/stores/setting"
	"github.com/eolinker/eosc/log"
	"github.com/eolinker/go-common/utils"
)

var (
	_ ISettingService = (*imlSettingService)(nil)
)

const (
	KeyInvokeAddress = "system.node.invoke_address"
	KeySitePrefix    = "system.setting.site_prefix"
)

// sensitiveKeys are setting names whose values hold secrets (API keys, SMTP
// passwords, webhook tokens). They are transparently encrypted at rest and
// decrypted on read. Non-sensitive keys are stored as-is.
var sensitiveKeys = map[string]struct{}{
	"costopt.judge":  {}, // OpenAI-compatible judge endpoint incl. api_key
	"costopt.notify": {}, // alert channels incl. SMTP password / webhook secret
}

type imlSettingService struct {
	store     setting.ISettingStore `autowired:""`
	secretCfg crypto.SecretConfig   `autowired:""`
}

// OnInit installs the configured AES master key so secrets are encrypted at rest.
// A missing/invalid key is logged and leaves encryption disabled (transparent).
func (i *imlSettingService) OnInit() {
	if err := crypto.Init(i.secretCfg.SecretKey); err != nil {
		log.Warn("crypto: master key unavailable, secrets will be stored in plaintext: ", err)
	}
}

func (i *imlSettingService) All(ctx context.Context) (map[string]string, error) {
	list, err := i.store.All(ctx)
	if err != nil {
		return nil, err
	}
	return utils.SliceToMapO(list, func(v *setting.Setting) (string, string) {
		val := v.Value
		if _, ok := sensitiveKeys[v.Name]; ok {
			if dec, derr := crypto.Decrypt(val); derr == nil {
				val = dec
			}
		}
		return v.Name, val
	}), nil
}

func (i *imlSettingService) Get(ctx context.Context, name string) (string, bool) {
	ev, err := i.store.Get(ctx, name)
	if err != nil {
		return "", false
	}
	val := ev.Value
	if _, ok := sensitiveKeys[name]; ok {
		if dec, derr := crypto.Decrypt(val); derr == nil {
			val = dec
		}
	}
	return val, true
}

func (i *imlSettingService) Set(ctx context.Context, name string, value string, operator string) error {
	storeVal := value
	if _, ok := sensitiveKeys[name]; ok {
		if enc, eerr := crypto.Encrypt(value); eerr == nil {
			storeVal = enc
		}
	}
	return i.store.Set(ctx, name, storeVal, operator)
}
