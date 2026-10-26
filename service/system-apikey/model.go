package system_apikey

import (
	"time"

	"github.com/RiemannChow/nantianmen-ai-gateway/common/crypto"
	"github.com/RiemannChow/nantianmen-ai-gateway/stores/system"
)

type APIKey struct {
	Id       string
	Name     string
	Value    string
	Creator  string
	Updater  string
	CreateAt time.Time
	UpdateAt time.Time
	Expired  int64
}

type Create struct {
	Id      string
	Name    string
	Value   string
	Expired int64
}

type Update struct {
	Name    *string
	Value   *string
	Expired *int64
}

func FromEntity(s *system.APIKey) *APIKey {
	// Transparently decrypt the secret value on read so callers (API response,
	// data-plane md5) always see plaintext; when encryption is disabled or the
	// row is legacy plaintext, Decrypt is a no-op.
	value := s.Value
	if dec, err := crypto.Decrypt(s.Value); err == nil {
		value = dec
	}
	return &APIKey{
		Id:       s.UUID,
		Name:     s.Name,
		Value:    value,
		Creator:  s.Creator,
		Updater:  s.Updater,
		CreateAt: s.CreateAt,
		UpdateAt: s.UpdateAt,
		Expired:  s.Expired,
	}
}
