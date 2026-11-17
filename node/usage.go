package node

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sublink/models"
	"sublink/services/mihomo"
	"sublink/utils"
	"sync"
	"time"

	"github.com/metacubex/mihomo/constant"
)

// FetchAirportUsageInfo 独立获取单个机场的用量信息
// 仅请求订阅地址获取 subscription-userinfo header，不解析节点内容
func FetchAirportUsageInfo(airport *models.Airport) (*UsageInfo, error) {
	if airport == nil {
		return nil, fmt.Errorf("airport is nil")
	}

	if !airport.FetchUsageInfo {
		return nil, fmt.Errorf("usage fetching is not enabled for this airport")
	}

	client := &http.Client{
		Timeout: 10 * time.Second, // 用量获取使用较短超时
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: airport.SkipTLSVerify},
		},
	}

	// 配置代理（如果启用）
	if airport.DownloadWithProxy {
		var proxyNodeLink string

		if airport.ProxyLink != "" {
			proxyNodeLink = airport.ProxyLink
			utils.Info("Usage fetch uses the specified proxy")
		} else {
			// 自动选择最佳代理
			if bestNode, err := models.GetBestProxyNode(); err == nil && bestNode != nil {
				utils.Info("Usage fetch auto-selected proxy node: %s", bestNode.Name)
				proxyNodeLink = bestNode.Link
			}
		}

		if proxyNodeLink != "" {
			proxyAdapter, err := mihomo.GetMihomoAdapter(proxyNodeLink)
			if err != nil {
				utils.Error("Failed to create proxy adapter: %v; requesting directly", err)
			} else {
				client.Transport = &http.Transport{
					DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
						host, portStr, splitErr := net.SplitHostPort(addr)
						if splitErr != nil {
							return nil, fmt.Errorf("split host port error: %v", splitErr)
						}

						portUint, parseErr := strconv.ParseUint(portStr, 10, 16)
						if parseErr != nil {
							return nil, fmt.Errorf("invalid port: %v", parseErr)
						}

						metadata := &constant.Metadata{
							Host:    host,
							DstPort: uint16(portUint),
							Type:    constant.HTTP,
						}

						return proxyAdapter.DialContext(ctx, metadata)
					},
					TLSClientConfig: &tls.Config{InsecureSkipVerify: airport.SkipTLSVerify},
				}
			}
		}
	}

	// 设置通用 User-Agent（机场配置了设备伪装时使用设备档位）
	userAgent := airport.EffectiveUserAgent()
	if userAgent == "" {
		userAgent = "clash.meta"
	}
	requestHeaders := airport.EffectiveRequestHeaders()

	// 优先使用 HEAD 请求，减少数据传输
	subscriptionURL, err := resolveHappSubscriptionURL(airport.URL)
	if err != nil {
		return nil, err
	}

	var resp *http.Response
	headReq, err := http.NewRequestWithContext(context.Background(), http.MethodHead, subscriptionURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	applyRequestHeaders(headReq, userAgent, requestHeaders)

	resp, err = client.Do(headReq)
	if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// HEAD 请求失败或返回非 2xx，回退到 GET 请求
		if resp != nil {
			_ = resp.Body.Close()
		}
		if err != nil {
			utils.Debug("Airport [%s] HEAD request failed: %v; trying GET", airport.Name, err)
		} else {
			utils.Debug("Airport [%s] HEAD returned status %d; trying GET", airport.Name, resp.StatusCode)
		}

		getReq, err := http.NewRequestWithContext(context.Background(), http.MethodGet, subscriptionURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %v", err)
		}
		applyRequestHeaders(getReq, userAgent, requestHeaders)

		resp, err = client.Do(getReq)
		if err != nil {
			return nil, fmt.Errorf("airport request failed: %v", err)
		}
	}
	defer func() { _ = resp.Body.Close() }()

	// 解析 subscription-userinfo header
	subUserInfo := resp.Header.Get("subscription-userinfo")
	if subUserInfo == "" {
		utils.Warn("Airport [%s] returned no usage header", airport.Name)
		return FailedUsageInfo(), nil
	}

	usageInfo := ParseSubscriptionUserInfo(subUserInfo)
	if usageInfo == nil {
		utils.Warn("Airport [%s] failed to parse usage header", airport.Name)
		return FailedUsageInfo(), nil
	}

	utils.Info("Airport [%s] usage fetched: upload=%d, download=%d, total=%d, expire=%d",
		airport.Name, usageInfo.Upload, usageInfo.Download, usageInfo.Total, usageInfo.Expire)

	return usageInfo, nil
}

// UpdateAirportUsageInfo 获取并保存机场用量到数据库
// 返回最新的 UsageInfo 或错误
func UpdateAirportUsageInfo(airportID int) (*UsageInfo, error) {
	airport, err := models.GetAirportByID(airportID)
	if err != nil {
		return nil, fmt.Errorf("failed to load airport: %v", err)
	}

	if airport == nil {
		return nil, fmt.Errorf("airport does not exist: %d", airportID)
	}

	if !airport.FetchUsageInfo {
		return nil, fmt.Errorf("airport [%s] has usage fetching disabled", airport.Name)
	}

	usageInfo, err := FetchAirportUsageInfo(airport)
	if err != nil {
		// 记录失败点，供用量历史图表标记红色失败
		airport.RecordUsageFailure(err.Error())
		return nil, err
	}

	// 保存到数据库
	if usageInfo != nil {
		if usageInfo.Total < 0 {
			if saveErr := airport.UpdateUsageInfoWithStatus(usageInfo.Upload, usageInfo.Download, usageInfo.Total, usageInfo.Expire, true, "Usage info unavailable"); saveErr != nil {
				utils.Error("Failed to save usage failure state for airport [%s]: %v", airport.Name, saveErr)
				return usageInfo, fmt.Errorf("failed to save usage info: %v", saveErr)
			}
			utils.Warn("Airport [%s] returned no usage info; failure point recorded", airport.Name)
			return usageInfo, nil
		}
		if err := airport.UpdateUsageInfo(usageInfo.Upload, usageInfo.Download, usageInfo.Total, usageInfo.Expire); err != nil {
			utils.Error("Failed to save usage info for airport [%s]: %v", airport.Name, err)
			return usageInfo, fmt.Errorf("failed to save usage info: %v", err)
		}
		utils.Info("Usage info saved for airport [%s]", airport.Name)
	}

	return usageInfo, nil
}

// UsageResult 单个机场用量获取结果
type UsageResult struct {
	AirportID   int
	AirportName string
	UsageInfo   *UsageInfo
	Error       error
}

// BatchUpdateAirportUsage 批量更新多个机场的用量信息
// 并发获取每个机场的用量信息并更新到数据库
// 返回各机场的用量结果映射
func BatchUpdateAirportUsage(airportIDs []int) map[int]*UsageResult {
	var wg sync.WaitGroup
	var resultsMap sync.Map

	for _, id := range airportIDs {
		wg.Add(1)
		go func(airportID int) {
			defer wg.Done()

			airport, err := models.GetAirportByID(airportID)
			if err != nil {
				resultsMap.Store(airportID, &UsageResult{
					AirportID: airportID,
					Error:     fmt.Errorf("failed to load airport: %v", err),
				})
				return
			}

			if airport == nil {
				resultsMap.Store(airportID, &UsageResult{
					AirportID: airportID,
					Error:     fmt.Errorf("airport does not exist"),
				})
				return
			}

			// 未开启用量获取的跳过
			if !airport.FetchUsageInfo {
				resultsMap.Store(airportID, &UsageResult{
					AirportID:   airportID,
					AirportName: airport.Name,
					Error:       nil, // 不算错误，只是跳过
				})
				return
			}

			usageInfo, err := UpdateAirportUsageInfo(airportID)
			resultsMap.Store(airportID, &UsageResult{
				AirportID:   airportID,
				AirportName: airport.Name,
				UsageInfo:   usageInfo,
				Error:       err,
			})
		}(id)
	}

	// 等待所有并发任务完成
	wg.Wait()

	// 将 sync.Map 转换为普通 map 返回
	results := make(map[int]*UsageResult)
	resultsMap.Range(func(key, value any) bool {
		airportID, ok := key.(int)
		if !ok {
			utils.Error("Unexpected airport usage result key type: %T", key)
			return true
		}
		result, ok := value.(*UsageResult)
		if !ok {
			utils.Error("Unexpected airport usage result value type: %T", value)
			return true
		}
		results[airportID] = result
		return true
	})

	return results
}

// RefreshUsageForSubscriptionNodes 为订阅的节点刷新关联机场的用量信息
// 收集节点所属的所有机场ID，批量获取用量信息
func RefreshUsageForSubscriptionNodes(nodes []models.Node) {
	// 收集所有开启 FetchUsageInfo 的机场ID
	airportIDs := make(map[int]bool)
	for _, node := range nodes {
		if node.Source != "manual" && node.SourceID > 0 {
			airportIDs[node.SourceID] = true
		}
	}

	if len(airportIDs) == 0 {
		utils.Debug("No airports need a usage refresh")
		return
	}

	// 转换为切片
	ids := make([]int, 0, len(airportIDs))
	for id := range airportIDs {
		ids = append(ids, id)
	}

	utils.Info("Refreshing usage for %d airports", len(ids))

	// 批量更新用量
	results := BatchUpdateAirportUsage(ids)

	// 统计结果
	successCount := 0
	failCount := 0
	skipCount := 0
	for _, result := range results {
		if result.Error != nil {
			failCount++
			utils.Error("Airport [%s] usage refresh failed: %v", result.AirportName, result.Error)
		} else if result.UsageInfo != nil {
			successCount++
		} else {
			skipCount++
		}
	}

	utils.Info("Usage refresh finished: succeeded=%d, failed=%d, skipped=%d", successCount, failCount, skipCount)
}
