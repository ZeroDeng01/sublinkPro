package unlock

import (
	"fmt"
	"sublink/models"
)

type disneyUnlockChecker struct{}

func (disneyUnlockChecker) Key() string { return models.UnlockProviderDisney }

func (disneyUnlockChecker) Aliases() []string {
	return []string{"disney", "disney+", "disneyplus", "disney_plus"}
}

func (disneyUnlockChecker) Meta() models.UnlockProviderMeta {
	return models.UnlockProviderMeta{Value: models.UnlockProviderDisney, Label: "Disney+", Description: "检测 Disney+ 服务入口是否可访问及是否明显受限", Category: "streaming"}
}

func (disneyUnlockChecker) RenameVariableMeta() models.UnlockRenameVariableMeta {
	return models.UnlockRenameVariableMeta{Provider: models.UnlockProviderDisney}
}

func (disneyUnlockChecker) Check(runtime UnlockRuntime) models.UnlockProviderResult {
	resp, err := fetchUnlockProbe(runtime, "https://www.disneyplus.com/", nil)
	if err != nil {
		return models.UnlockProviderResult{Provider: models.UnlockProviderDisney, Status: models.UnlockStatusError, Reason: err.Error()}
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		if containsAny(resp.Body, []string{"not available in your region", "not available in your country"}) {
			return models.UnlockProviderResult{Provider: models.UnlockProviderDisney, Status: models.UnlockStatusRestricted, Reason: "region_blocked"}
		}
		return models.UnlockProviderResult{Provider: models.UnlockProviderDisney, Status: models.UnlockStatusReachable}
	}
	return models.UnlockProviderResult{Provider: models.UnlockProviderDisney, Status: models.UnlockStatusRestricted, Reason: fmt.Sprintf("status_%d", resp.StatusCode)}
}

func init() {
	RegisterUnlockChecker(disneyUnlockChecker{})
}
