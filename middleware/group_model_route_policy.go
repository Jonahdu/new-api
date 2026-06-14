package middleware

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
)

type groupModelRoutePolicy struct {
	PassThroughBodyEnabled *bool `json:"pass_through_body_enabled,omitempty"`
}

func applyGroupModelRoutePolicy(setting dto.ChannelSettings, policy map[string]map[string]groupModelRoutePolicy, group string, model string) dto.ChannelSettings {
	groupPolicy, ok := policy[group]
	if !ok {
		return setting
	}
	modelPolicy, ok := groupPolicy[model]
	if !ok || modelPolicy.PassThroughBodyEnabled == nil {
		return setting
	}
	setting.PassThroughBodyEnabled = *modelPolicy.PassThroughBodyEnabled
	return setting
}

func loadGroupModelRoutePolicy() map[string]map[string]groupModelRoutePolicy {
	common.OptionMapRWMutex.RLock()
	raw := common.OptionMap["GroupModelRoutePolicy"]
	common.OptionMapRWMutex.RUnlock()
	if raw == "" {
		return nil
	}
	policy := map[string]map[string]groupModelRoutePolicy{}
	if err := common.Unmarshal([]byte(raw), &policy); err != nil {
		return nil
	}
	return policy
}
