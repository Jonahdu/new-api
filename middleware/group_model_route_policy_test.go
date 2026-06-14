package middleware

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func TestApplyGroupModelRoutePolicyOverridesPassThroughForMatchingGroupModel(t *testing.T) {
	setting := dto.ChannelSettings{PassThroughBodyEnabled: false}
	policy := map[string]map[string]groupModelRoutePolicy{
		"memrouter": {
			"deepseek-v4-flash": {PassThroughBodyEnabled: boolPtr(true)},
		},
	}

	result := applyGroupModelRoutePolicy(setting, policy, "memrouter", "deepseek-v4-flash")

	if !result.PassThroughBodyEnabled {
		t.Fatal("expected matching group/model policy to enable pass-through body")
	}
}

func TestApplyGroupModelRoutePolicyKeepsChannelDefaultWhenNoMatch(t *testing.T) {
	setting := dto.ChannelSettings{PassThroughBodyEnabled: true}
	policy := map[string]map[string]groupModelRoutePolicy{
		"memrouter": {
			"deepseek-v4-flash": {PassThroughBodyEnabled: boolPtr(false)},
		},
	}

	result := applyGroupModelRoutePolicy(setting, policy, "ds", "deepseek-v4-flash")

	if !result.PassThroughBodyEnabled {
		t.Fatal("expected unmatched policy to keep channel default pass-through body setting")
	}
}

func TestLoadGroupModelRoutePolicyReadsOptionMap(t *testing.T) {
	withGroupModelRoutePolicyOption(t, `{"memrouter":{"deepseek-v4-flash":{"pass_through_body_enabled":true}}}`)

	policy := loadGroupModelRoutePolicy()
	result := applyGroupModelRoutePolicy(dto.ChannelSettings{}, policy, "memrouter", "deepseek-v4-flash")

	if !result.PassThroughBodyEnabled {
		t.Fatal("expected policy loaded from OptionMap to enable pass-through body")
	}
}

func TestSetupContextForSelectedChannelAppliesGroupModelRoutePolicy(t *testing.T) {
	withGroupModelRoutePolicyOption(t, `{"memrouter":{"deepseek-v4-flash":{"pass_through_body_enabled":true}}}`)
	gin.SetMode(gin.TestMode)
	ctx := &gin.Context{}
	ctx.Set("group", "memrouter")
	setting := `{"pass_through_body_enabled":false}`
	channel := &model.Channel{Id: 1, Name: "opencode", Type: 1, Setting: &setting}

	if err := SetupContextForSelectedChannel(ctx, channel, "deepseek-v4-flash"); err != nil {
		t.Fatalf("SetupContextForSelectedChannel returned error: %v", err)
	}
	actual, ok := common.GetContextKeyType[dto.ChannelSettings](ctx, constant.ContextKeyChannelSetting)
	if !ok {
		t.Fatal("expected channel setting in context")
	}
	if !actual.PassThroughBodyEnabled {
		t.Fatal("expected group/model policy to override channel setting in context")
	}
}

func TestSetupContextForSelectedChannelKeepsChannelSettingWhenPolicyIsInvalid(t *testing.T) {
	withGroupModelRoutePolicyOption(t, `{bad json`)
	gin.SetMode(gin.TestMode)
	ctx := &gin.Context{}
	ctx.Set("group", "memrouter")
	setting := `{"pass_through_body_enabled":true}`
	channel := &model.Channel{Id: 1, Name: "opencode", Type: 1, Setting: &setting}

	if err := SetupContextForSelectedChannel(ctx, channel, "deepseek-v4-flash"); err != nil {
		t.Fatalf("SetupContextForSelectedChannel returned error: %v", err)
	}
	actual, ok := common.GetContextKeyType[dto.ChannelSettings](ctx, constant.ContextKeyChannelSetting)
	if !ok {
		t.Fatal("expected channel setting in context")
	}
	if !actual.PassThroughBodyEnabled {
		t.Fatal("expected invalid policy to keep channel pass-through body setting")
	}
}

func withGroupModelRoutePolicyOption(t *testing.T, raw string) {
	t.Helper()
	common.OptionMapRWMutex.Lock()
	if common.OptionMap == nil {
		common.OptionMap = map[string]string{}
	}
	previous, hadPrevious := common.OptionMap["GroupModelRoutePolicy"]
	common.OptionMap["GroupModelRoutePolicy"] = raw
	common.OptionMapRWMutex.Unlock()

	t.Cleanup(func() {
		common.OptionMapRWMutex.Lock()
		if hadPrevious {
			common.OptionMap["GroupModelRoutePolicy"] = previous
		} else {
			delete(common.OptionMap, "GroupModelRoutePolicy")
		}
		common.OptionMapRWMutex.Unlock()
	})
}

func boolPtr(value bool) *bool {
	return &value
}
