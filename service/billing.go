package service

import (
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/QuantumNous/new-api/types"
	"github.com/gin-gonic/gin"
)

// 计费来源类型（仅保留常量供其他模块引用）
const (
	BillingSourceWallet       = "wallet"
	BillingSourceSubscription = "subscription"
)

// PreConsumeBilling 已删除：免额度模式，直接返回 nil
func PreConsumeBilling(c *gin.Context, preConsumedQuota int, relayInfo *relaycommon.RelayInfo) *types.NewAPIError {
	return nil
}

// SettleBilling 已删除：免额度模式，直接返回 nil
func SettleBilling(ctx *gin.Context, relayInfo *relaycommon.RelayInfo, actualQuota int) error {
	return nil
}
