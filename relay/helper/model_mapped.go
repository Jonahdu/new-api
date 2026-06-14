package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/relay/common"
	relayconstant "github.com/QuantumNous/new-api/relay/constant"
	"github.com/QuantumNous/new-api/setting/ratio_setting"
	"github.com/gin-gonic/gin"
)

// modelMappingEntry represents an object-format model mapping value.
// Example: "claude-opus-4-8[1M]": {"model": "qwen3.7-plus", "url": "https://..."}
type modelMappingEntry struct {
	Model string `json:"model"`
	URL   string `json:"url,omitempty"`
}

func ModelMappedHelper(c *gin.Context, info *common.RelayInfo, request dto.Request) error {
	if info.ChannelMeta == nil {
		info.ChannelMeta = &common.ChannelMeta{}
	}

	isResponsesCompact := info.RelayMode == relayconstant.RelayModeResponsesCompact
	originModelName := info.OriginModelName
	mappingModelName := originModelName
	if isResponsesCompact && strings.HasSuffix(originModelName, ratio_setting.CompactModelSuffix) {
		mappingModelName = strings.TrimSuffix(originModelName, ratio_setting.CompactModelSuffix)
	}

	// map model name
	modelMapping := c.GetString("model_mapping")
	if modelMapping != "" && modelMapping != "{}" {
		// Parse as raw message map first to support both string and object formats
		var rawMap map[string]json.RawMessage
		err := json.Unmarshal([]byte(modelMapping), &rawMap)
		if err != nil {
			return fmt.Errorf("unmarshal_model_mapping_failed")
		}

		// Build model name map and track URL override for the matching model
		modelNameMap := make(map[string]string, len(rawMap))
		var urlOverride string

		for key, rawVal := range rawMap {
			// Try string format first: "model1": "alias1"
			var strVal string
			if err := json.Unmarshal(rawVal, &strVal); err == nil {
				modelNameMap[key] = strVal
				continue
			}

			// Object format: "model1": {"model": "alias1", "url": "..."}
			var entry modelMappingEntry
			if err := json.Unmarshal(rawVal, &entry); err != nil || entry.Model == "" {
				continue
			}
			modelNameMap[key] = entry.Model
			if strings.EqualFold(key, mappingModelName) && entry.URL != "" {
				urlOverride = entry.URL
			}
		}

		// 支持链式模型重定向，最终使用链尾的模型
		currentModel := mappingModelName
		visitedModels := map[string]bool{
			currentModel: true,
		}
		for {
			if mappedModel, exists := modelNameMap[currentModel]; exists && mappedModel != "" {
				// 模型重定向循环检测，避免无限循环
				if visitedModels[mappedModel] {
					if mappedModel == currentModel {
						if currentModel == info.OriginModelName {
							info.IsModelMapped = false
							return nil
						} else {
							info.IsModelMapped = true
							break
						}
					}
					return errors.New("model_mapping_contains_cycle")
				}
				visitedModels[mappedModel] = true
				currentModel = mappedModel
				info.IsModelMapped = true
			} else {
				break
			}
		}
		if info.IsModelMapped {
			info.UpstreamModelName = currentModel
		}

		// Apply URL override — use the full URL from model_mapping as the target base URL.
		// This supports routing different models to different upstream endpoints/paths.
		if urlOverride != "" {
			info.ChannelBaseUrl = urlOverride
			// Auto-switch adapter based on URL path:
			//   /v1/messages          → Claude adapter (passthrough, no translation)
			//   /v1/chat/completions  → keep default adapter (OpenAI translation)
			if strings.Contains(urlOverride, "/v1/messages") {
				info.ApiType = constant.APITypeAnthropic
			}
		}
	}

	if isResponsesCompact {
		finalUpstreamModelName := mappingModelName
		if info.IsModelMapped && info.UpstreamModelName != "" {
			finalUpstreamModelName = info.UpstreamModelName
		}
		info.UpstreamModelName = finalUpstreamModelName
		info.OriginModelName = ratio_setting.WithCompactModelSuffix(finalUpstreamModelName)
	}
	if request != nil {
		request.SetModelName(info.UpstreamModelName)
	}
	return nil
}
