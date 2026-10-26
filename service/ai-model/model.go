package ai_model

import (
	"encoding/base64"
	"github.com/RiemannChow/nantianmen-ai-gateway/stores/ai"
	"time"
)

type ProviderModel struct {
	Id                  string // provider model:uuid
	Name                string
	Type                string
	AccessConfiguration string
	ModelParameters     string
	Provider            string
	Creator             string
	Updater             string
	CreateAt            time.Time
	UpdateAt            time.Time
}

func FromEntity(e *ai.ProviderModel) *ProviderModel {
	// 两个字段分别解码，各自失败时回退为原始值（兼容历史非 base64 数据）
	accessConfiguration, err := base64.RawStdEncoding.DecodeString(e.AccessConfiguration)
	if err != nil {
		accessConfiguration = []byte(e.AccessConfiguration)
	}
	modelParameters, err := base64.RawStdEncoding.DecodeString(e.ModelParameters)
	if err != nil {
		modelParameters = []byte(e.ModelParameters)
	}
	return &ProviderModel{
		Id:                  e.UUID,
		Name:                e.Name,
		Type:                e.Type,
		AccessConfiguration: string(accessConfiguration),
		ModelParameters:     string(modelParameters),
		Provider:            e.Provider,
		Creator:             e.Creator,
		Updater:             e.Updater,
		CreateAt:            e.CreateAt,
		UpdateAt:            e.UpdateAt,
	}
}

type Model struct {
	Name                *string
	Provider            *string
	Type                *string
	AccessConfiguration *string
	ModelParameters     *string
}
