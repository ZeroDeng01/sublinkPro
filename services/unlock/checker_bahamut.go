package unlock

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"sublink/models"
)

const (
	bahamutProbeBodyLimit = 32 * 1024
	bahamutTraceBodyLimit = 8 * 1024
	// bahamutDefaultTimeout 与 runtime.go 全局默认对齐：解锁检测对慢节点不需要宽裕兜底，
	// 慢节点即使解锁成功用户也不会选用，超时即判为不解锁是正确的产品语义。
	bahamutDefaultTimeout = 5 * time.Second
	// bahamutRetryDeadline 控制整体上界：仅给 token.php CDN 偶发 5xx 留 1 次额外重试机会，
	// 不再做长时间兜底。最多 2 次请求 × 5s 超时 = 10s 上界。
	bahamutRetryDeadline = 10 * time.Second
	bahamutMaxAttempts   = 2
	// bahamutAdID 是巴哈姆特 token.php 接口固定的广告位参数
	bahamutAdID = "89422"
	// bahamutLenientSn 是入口探针动画 SN：港澳台 IP 均可获取 token，通过即表示在巴哈姆特服务区
	bahamutLenientSn = "37783"
	// bahamutStrictSn 是台湾确认探针动画 SN：仅台湾 IP 可获取 token，通过即表示出口在台湾
	bahamutStrictSn = "38832"
)

type bahamutAnimeUnlockChecker struct{}

func (bahamutAnimeUnlockChecker) Key() string { return models.UnlockProviderBahamut }

func (bahamutAnimeUnlockChecker) Aliases() []string {
	return []string{"bahamut", "bahamutanime", "ani_gamer", "animedance"}
}

func (bahamutAnimeUnlockChecker) Meta() models.UnlockProviderMeta {
	return models.UnlockProviderMeta{
		Value:       models.UnlockProviderBahamut,
		Label:       "Bahamut Anime",
		Description: "检测巴哈姆特动画疯（ani.gamer.com.tw）地区可访问性",
		Category:    "streaming",
	}
}

func (bahamutAnimeUnlockChecker) RenameVariableMeta() models.UnlockRenameVariableMeta {
	return models.UnlockRenameVariableMeta{Provider: models.UnlockProviderBahamut}
}

// Check 执行巴哈姆特动画疯解锁检测。
// Transport 复用 runtime（必须走节点代理），cookie jar 独立（维持 deviceid→token session）。
func (bahamutAnimeUnlockChecker) Check(runtime UnlockRuntime) models.UnlockProviderResult {
	client := bahamutClient(runtime)
	headers := bahamutHeaders()

	deviceIDBody, err := bahamutFetchBody(client, "https://ani.gamer.com.tw/ajax/getdeviceid.php", headers, bahamutProbeBodyLimit)
	if err != nil {
		return bahamutErrorResult(runtime, err.Error())
	}
	deviceIDResp := &unlockHTTPResponse{RawBody: string(deviceIDBody)}
	deviceID, unsupported := evaluateBahamutDeviceID(deviceIDResp.RawBody)
	// HTML 拦截页短路：地区完全不在服务范围，跳过后续 token 探针
	if unsupported {
		return evaluateBahamutUnlockProbe(runtime, deviceIDResp, nil, nil, nil)
	}

	// 入口探针先查：通过是 Available 的前提，不通过直接短路
	lenientBody, err := bahamutFetchBody(client, bahamutTokenURL(bahamutLenientSn, deviceID), headers, bahamutProbeBodyLimit)
	if err != nil {
		return bahamutErrorResult(runtime, err.Error())
	}
	lenientResp := &unlockHTTPResponse{RawBody: string(lenientBody)}
	if !evaluateBahamutToken(lenientResp.RawBody) {
		return evaluateBahamutUnlockProbe(runtime, deviceIDResp, lenientResp, nil, nil)
	}

	// 台湾确认探针：通过则直接返回 TW，无需 trace
	strictBody, err := bahamutFetchBody(client, bahamutTokenURL(bahamutStrictSn, deviceID), headers, bahamutProbeBodyLimit)
	if err != nil {
		return bahamutErrorResult(runtime, err.Error())
	}
	strictResp := &unlockHTTPResponse{RawBody: string(strictBody)}
	if evaluateBahamutToken(strictResp.RawBody) {
		return evaluateBahamutUnlockProbe(runtime, deviceIDResp, lenientResp, strictResp, nil)
	}

	// 入口通过但台湾探针未通过：查 trace 拿港澳台具体 region（可选，失败不影响整体可用性判定）
	traceBody, traceErr := bahamutFetchBody(client, "https://ani.gamer.com.tw/cdn-cgi/trace", headers, bahamutTraceBodyLimit)
	var traceResp *unlockHTTPResponse
	if traceErr == nil && len(traceBody) > 0 {
		traceResp = &unlockHTTPResponse{RawBody: string(traceBody)}
	}
	return evaluateBahamutUnlockProbe(runtime, deviceIDResp, lenientResp, strictResp, traceResp)
}

// evaluateBahamutUnlockProbe 根据各探针响应判定巴哈姆特解锁状态（纯函数，便于测试）。
// 判定顺序：deviceID 受限 → 入口探针不通过 → 台湾探针通过 → 港澳台 trace。
func evaluateBahamutUnlockProbe(runtime UnlockRuntime, deviceIDResp, lenientResp, strictResp, traceResp *unlockHTTPResponse) models.UnlockProviderResult {
	if deviceIDResp == nil || deviceIDResp.RawBody == "" {
		return bahamutErrorResult(runtime, "network_connection")
	}

	_, unsupported := evaluateBahamutDeviceID(deviceIDResp.RawBody)
	if unsupported {
		return models.UnlockProviderResult{Provider: models.UnlockProviderBahamut, Status: models.UnlockStatusUnsupported, Region: runtime.LandingCountry, Reason: "not_available"}
	}

	// 入口探针不通过：完全不在巴哈姆特服务区
	if lenientResp == nil || !evaluateBahamutToken(lenientResp.RawBody) {
		return models.UnlockProviderResult{Provider: models.UnlockProviderBahamut, Status: models.UnlockStatusUnsupported, Region: runtime.LandingCountry, Reason: "region_unavailable"}
	}

	// 台湾确认探针通过：完整台湾解锁
	if strictResp != nil && evaluateBahamutToken(strictResp.RawBody) {
		return models.UnlockProviderResult{Provider: models.UnlockProviderBahamut, Status: models.UnlockStatusAvailable, Region: "TW"}
	}

	// 入口通过但台湾探针未通过：港澳台可用，从 trace 拿精确 region
	region := ""
	if traceResp != nil {
		region = evaluateBahamutRegion(traceResp.RawBody)
	}
	if region == "" {
		// trace 失败或返回无效地区码时降级为 partial，避免误报精确地区
		return models.UnlockProviderResult{Provider: models.UnlockProviderBahamut, Status: models.UnlockStatusPartial, Region: runtime.LandingCountry, Detail: "primary_token_only"}
	}
	return models.UnlockProviderResult{Provider: models.UnlockProviderBahamut, Status: models.UnlockStatusAvailable, Region: region}
}

// evaluateBahamutDeviceID 解析 getdeviceid.php 响应，返回 deviceID 和是否为不支持地区。
// 非 JSON 且以 < 开头的响应视为地区受限的 HTML 拦截页。
func evaluateBahamutDeviceID(rawBody string) (deviceID string, unsupported bool) {
	var result struct {
		Deviceid string `json:"deviceid"`
	}
	if err := json.Unmarshal([]byte(rawBody), &result); err != nil {
		if strings.HasPrefix(strings.TrimSpace(rawBody), "<") {
			return "", true
		}
		return "", false
	}
	return result.Deviceid, false
}

// evaluateBahamutToken 解析 token.php 响应，判断该 SN 是否可获取 token。
// animeSn 非零表示当前 IP 在该 SN 对应的可观看地区内。
func evaluateBahamutToken(rawBody string) bool {
	var result struct {
		AnimeSn int `json:"animeSn"`
	}
	if err := json.Unmarshal([]byte(rawBody), &result); err != nil {
		return false
	}
	return result.AnimeSn != 0
}

// evaluateBahamutRegion 从 cdn-cgi/trace 响应中提取两字母地区代码（大写）。
func evaluateBahamutRegion(rawBody string) string {
	idx := strings.Index(rawBody, "loc=")
	if idx == -1 {
		return ""
	}
	rest := rawBody[idx+4:]
	idx = strings.Index(rest, "\n")
	if idx == -1 {
		return ""
	}
	loc := strings.TrimSpace(rest[:idx])
	if len(loc) == 2 {
		return strings.ToUpper(loc)
	}
	return ""
}

// bahamutClient 构造独立 client 用于 session 管理。
// Transport 复用 runtime（必须走节点代理），cookie jar 独立（维持 deviceid→token session）。
// Timeout 优先采纳 runtime.Timeout（已对齐全局解锁检测语义），仅在缺失时回退到默认值。
func bahamutClient(runtime UnlockRuntime) *http.Client {
	jar, _ := cookiejar.New(nil)
	timeout := runtime.Timeout
	if timeout <= 0 {
		timeout = bahamutDefaultTimeout
	}
	return &http.Client{
		Transport:     runtime.Client.Transport,
		Timeout:       timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		Jar:           jar,
	}
}

// bahamutFetchBody 执行 GET 请求并返回受限的 body。
// 最多重试 1 次（仅为容忍 token.php CDN 偶发 5xx），整体受 bahamutRetryDeadline 约束。
// 遇到 DNS/超时/连接拒绝/TLS 握手等明确不可恢复错误立即停止重试。
func bahamutFetchBody(client *http.Client, target string, headers map[string]string, bodyLimit int64) ([]byte, error) {
	if bodyLimit <= 0 {
		bodyLimit = bahamutProbeBodyLimit
	}
	timeout := client.Timeout
	if timeout <= 0 {
		timeout = bahamutDefaultTimeout
	}

	var lastErr error
	deadline := time.Now().Add(bahamutRetryDeadline)
	for i := 0; i < bahamutMaxAttempts; i++ {
		if time.Now().After(deadline) {
			break
		}
		body, err := bahamutFetchBodyOnce(client, target, headers, bodyLimit, timeout)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if bahamutShouldStopRetry(err) {
			break
		}
	}
	return nil, lastErr
}

func bahamutFetchBodyOnce(client *http.Client, target string, headers map[string]string, bodyLimit int64, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		if strings.TrimSpace(key) != "" && strings.TrimSpace(value) != "" {
			req.Header.Set(key, value)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	return io.ReadAll(io.LimitReader(resp.Body, bodyLimit))
}

// bahamutShouldStopRetry 判断错误是否属于"重试也不会更好"的不可恢复类。
// 仅 token.php 的 CDN 5xx 这类瞬时错误才值得重试，其余一律短路返回。
func bahamutShouldStopRetry(err error) bool {
	if err == nil {
		return false
	}
	errText := strings.ToLower(err.Error())
	stopKeywords := []string{
		"no such host",
		"timeout",
		"deadline exceeded",
		"connection refused",
		"connection reset",
		"tls",
		"x509",
		"certificate",
		"proxy",
	}
	for _, kw := range stopKeywords {
		if strings.Contains(errText, kw) {
			return true
		}
	}
	return false
}

func bahamutTokenURL(sn, deviceID string) string {
	return fmt.Sprintf("https://ani.gamer.com.tw/ajax/token.php?adID=%s&sn=%s&device=%s", bahamutAdID, sn, deviceID)
}

func bahamutHeaders() map[string]string {
	return map[string]string{
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
		"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Edg/124.0.0.0",
	}
}

func bahamutErrorResult(runtime UnlockRuntime, reason string) models.UnlockProviderResult {
	return models.UnlockProviderResult{Provider: models.UnlockProviderBahamut, Status: models.UnlockStatusError, Region: runtime.LandingCountry, Reason: reason}
}

func init() {
	RegisterUnlockChecker(bahamutAnimeUnlockChecker{})
}
