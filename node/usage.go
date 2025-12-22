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
		return nil, fmt.Errorf("机场对象为空")
	}

	if !airport.FetchUsageInfo {
		return nil, fmt.Errorf("机场未开启用量信息获取")
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
			utils.Info("用量获取使用指定代理")
		} else {
			// 自动选择最佳代理
			if bestNode, err := models.GetBestProxyNode(); err == nil && bestNode != nil {
				utils.Info("用量获取自动选择代理节点: %s", bestNode.Name)
				proxyNodeLink = bestNode.Link
			}
		}

		if proxyNodeLink != "" {
			proxyAdapter, err := mihomo.GetMihomoAdapter(proxyNodeLink)
			if err != nil {
				utils.Error("创建代理适配器失败: %v，将直接请求", err)
			} else {
				client.Transport = &http.Transport{
					DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
						host, portStr, splitErr := net.SplitHostPort(addr)
						if splitErr != nil {
							return nil, fmt.Errorf("split host port error: %v", splitErr)
						}

						portInt, atoiErr := strconv.Atoi(portStr)
						if atoiErr != nil {
							return nil, fmt.Errorf("invalid port: %v", atoiErr)
						}

						if portInt < 0 || portInt > 65535 {
							return nil, fmt.Errorf("port out of range: %d", portInt)
						}

						metadata := &constant.Metadata{
							Host:    host,
							DstPort: uint16(portInt),
							Type:    constant.HTTP,
						}

						return proxyAdapter.DialContext(ctx, metadata)
					},
					TLSClientConfig: &tls.Config{InsecureSkipVerify: airport.SkipTLSVerify},
				}
			}
		}
	}

	// 设置通用 User-Agent
	userAgent := "clash.meta"
	if airport.UserAgent != "" {
		userAgent = airport.UserAgent
	}

	// 优先使用 HEAD 请求，减少数据传输
	var resp *http.Response
	headReq, err := http.NewRequest("HEAD", airport.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}
	headReq.Header.Set("User-Agent", userAgent)

	resp, err = client.Do(headReq)
	if err != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// HEAD 请求失败或返回非 2xx，回退到 GET 请求
		if resp != nil {
			resp.Body.Close()
		}
		if err != nil {
			utils.Debug("机场【%s】HEAD 请求失败: %v，尝试 GET 请求", airport.Name, err)
		} else {
			utils.Debug("机场【%s】HEAD 请求返回状态码 %d，尝试 GET 请求", airport.Name, resp.StatusCode)
		}

		getReq, err := http.NewRequest("GET", airport.URL, nil)
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %v", err)
		}
		getReq.Header.Set("User-Agent", userAgent)

		resp, err = client.Do(getReq)
		if err != nil {
			return nil, fmt.Errorf("请求机场失败: %v", err)
		}
	}
	defer resp.Body.Close()

	// 解析 subscription-userinfo header
	subUserInfo := resp.Header.Get("subscription-userinfo")
	if subUserInfo == "" {
		utils.Warn("机场【%s】未返回用量信息 header", airport.Name)
		return FailedUsageInfo(), nil
	}

	usageInfo := ParseSubscriptionUserInfo(subUserInfo)
	if usageInfo == nil {
		utils.Warn("机场【%s】用量信息 header 解析失败", airport.Name)
		return FailedUsageInfo(), nil
	}

	utils.Info("机场【%s】用量获取成功: 上传=%d, 下载=%d, 总量=%d, 过期=%d",
		airport.Name, usageInfo.Upload, usageInfo.Download, usageInfo.Total, usageInfo.Expire)

	return usageInfo, nil
}

// UpdateAirportUsageInfo 获取并保存机场用量到数据库
// 返回最新的 UsageInfo 或错误
func UpdateAirportUsageInfo(airportID int) (*UsageInfo, error) {
	airport, err := models.GetAirportByID(airportID)
	if err != nil {
		return nil, fmt.Errorf("获取机场失败: %v", err)
	}

	if airport == nil {
		return nil, fmt.Errorf("机场不存在: %d", airportID)
	}

	if !airport.FetchUsageInfo {
		return nil, fmt.Errorf("机场【%s】未开启用量信息获取", airport.Name)
	}

	usageInfo, err := FetchAirportUsageInfo(airport)
	if err != nil {
		return nil, err
	}

	// 保存到数据库
	if usageInfo != nil {
		if err := airport.UpdateUsageInfo(usageInfo.Upload, usageInfo.Download, usageInfo.Total, usageInfo.Expire); err != nil {
			utils.Error("保存机场【%s】用量信息失败: %v", airport.Name, err)
			return usageInfo, fmt.Errorf("保存用量信息失败: %v", err)
		}
		utils.Info("机场【%s】用量信息已保存", airport.Name)
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
					Error:     fmt.Errorf("获取机场失败: %v", err),
				})
				return
			}

			if airport == nil {
				resultsMap.Store(airportID, &UsageResult{
					AirportID: airportID,
					Error:     fmt.Errorf("机场不存在"),
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
	resultsMap.Range(func(key, value interface{}) bool {
		results[key.(int)] = value.(*UsageResult)
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
		utils.Debug("没有需要刷新用量的机场")
		return
	}

	// 转换为切片
	ids := make([]int, 0, len(airportIDs))
	for id := range airportIDs {
		ids = append(ids, id)
	}

	utils.Info("开始刷新 %d 个机场的用量信息", len(ids))

	// 批量更新用量
	results := BatchUpdateAirportUsage(ids)

	// 统计结果
	successCount := 0
	failCount := 0
	skipCount := 0
	for _, result := range results {
		if result.Error != nil {
			failCount++
			utils.Error("机场【%s】用量刷新失败: %v", result.AirportName, result.Error)
		} else if result.UsageInfo != nil {
			successCount++
		} else {
			skipCount++
		}
	}

	utils.Info("用量刷新完成: 成功=%d, 失败=%d, 跳过=%d", successCount, failCount, skipCount)
}
