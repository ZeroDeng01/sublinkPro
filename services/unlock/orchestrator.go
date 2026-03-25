package unlock

import (
	"sublink/models"
	"sublink/services/mihomo"
	"sync"
	"time"

	"github.com/metacubex/mihomo/constant"
)

func CheckUnlock(nodeLink string, timeout time.Duration, landingCountry string, providers []string) models.UnlockSummary {
	proxyAdapter, err := mihomo.GetMihomoAdapter(nodeLink)
	if err != nil {
		return models.UnlockSummary{Providers: buildUnlockErrorResults(providers, err.Error())}
	}
	return CheckUnlockWithAdapter(proxyAdapter, timeout, landingCountry, providers)
}

func CheckUnlockWithAdapter(proxyAdapter constant.Proxy, timeout time.Duration, landingCountry string, providers []string) models.UnlockSummary {
	requestedProviders := providers
	providerList := ResolveUnlockProviders(requestedProviders)
	if len(requestedProviders) == 0 {
		providerList = ListRegisteredUnlockProviders()
	}
	if len(requestedProviders) > 0 && len(providerList) == 0 {
		return models.UnlockSummary{Providers: buildUnlockErrorResults(requestedProviders, "no_registered_unlock_provider")}
	}
	runtime := newUnlockRuntime(proxyAdapter, timeout, landingCountry)
	results := make([]models.UnlockProviderResult, len(providerList))
	workerLimit := 3
	if workerLimit > len(providerList) {
		workerLimit = len(providerList)
	}
	sem := make(chan struct{}, workerLimit)
	var wg sync.WaitGroup
	for idx, provider := range providerList {
		wg.Add(1)
		go func(index int, providerName string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			checker, exists := GetUnlockChecker(providerName)
			if !exists {
				status := models.NormalizeUnlockStatus(models.UnlockStatusUnknown)
				if status == "" {
					status = models.UnlockStatusUnknown
				}
				results[index] = models.UnlockProviderResult{Provider: providerName, Status: status, Reason: "unsupported_provider"}
				return
			}
			result := checker.Check(runtime)
			if result.Provider == "" {
				result.Provider = providerName
			}
			results[index] = result
		}(idx, provider)
	}
	wg.Wait()
	return models.UnlockSummary{Providers: results, UpdatedAt: time.Now().Format("2006-01-02 15:04:05")}
}

func buildUnlockErrorResults(providers []string, reason string) []models.UnlockProviderResult {
	providerList := ResolveUnlockProviders(providers)
	if len(providers) == 0 && len(providerList) == 0 {
		providerList = ListRegisteredUnlockProviders()
	}
	if len(providers) > 0 && len(providerList) == 0 {
		providerList = models.NormalizeUnlockProviders(providers)
	}
	results := make([]models.UnlockProviderResult, 0, len(providerList))
	for _, provider := range providerList {
		status := models.NormalizeUnlockStatus(models.UnlockStatusError)
		if status == "" {
			status = models.UnlockStatusError
		}
		results = append(results, models.UnlockProviderResult{Provider: provider, Status: status, Reason: reason})
	}
	return results
}
