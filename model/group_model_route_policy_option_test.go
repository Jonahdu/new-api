package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
)

func TestInitOptionMapIncludesGroupModelRoutePolicy(t *testing.T) {
	InitOptionMap()

	common.OptionMapRWMutex.RLock()
	value, ok := common.OptionMap["GroupModelRoutePolicy"]
	common.OptionMapRWMutex.RUnlock()

	if !ok {
		t.Fatal("expected GroupModelRoutePolicy option to be initialized")
	}
	if value != "{}" {
		t.Fatalf("expected empty GroupModelRoutePolicy default, got %q", value)
	}
}

func TestUpdateOptionStoresGroupModelRoutePolicy(t *testing.T) {
	InitOptionMap()
	raw := `{"memrouter":{"deepseek-v4-flash":{"pass_through_body_enabled":true}}}`

	if err := updateOptionMap("GroupModelRoutePolicy", raw); err != nil {
		t.Fatalf("updateOptionMap returned error: %v", err)
	}

	common.OptionMapRWMutex.RLock()
	actual := common.OptionMap["GroupModelRoutePolicy"]
	common.OptionMapRWMutex.RUnlock()

	if actual != raw {
		t.Fatalf("expected GroupModelRoutePolicy to be stored, got %q", actual)
	}
}
