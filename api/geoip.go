package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sublink/config"
	"sublink/models"
	"sublink/services/geoip"
	"sublink/utils"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// 默认下载地址
const DefaultGeoIPDownloadURL = "https://git.io/GeoLite2-City.mmdb"

// 下载状态
var (
	downloadMu       sync.Mutex
	isDownloading    bool
	downloadProgress int
	downloadError    string
	stopDownload     chan struct{} // 停止下载信号
	downloadSource   string        // 下载来源: "auto" 或 "manual"
)

// GeoIPConfigResponse GeoIP 配置响应
type GeoIPConfigResponse struct {
	DownloadURL string `json:"downloadUrl"` // 下载地址
	UseProxy    bool   `json:"useProxy"`    // 是否使用代理
	ProxyLink   string `json:"proxyLink"`   // 代理节点链接
	LastUpdate  string `json:"lastUpdate"`  // 上次更新时间
}

// GeoIPStatusResponse GeoIP 状态响应
type GeoIPStatusResponse struct {
	Available     bool   `json:"available"`     // 数据库是否可用
	Path          string `json:"path"`          // 文件路径
	Size          int64  `json:"size"`          // 文件大小（字节）
	SizeFormatted string `json:"sizeFormatted"` // 格式化的文件大小
	ModTime       string `json:"modTime"`       // 最后修改时间
	Downloading   bool   `json:"downloading"`   // 是否正在下载
	Progress      int    `json:"progress"`      // 下载进度 (0-100)
	Error         string `json:"error"`         // 错误信息
	Source        string `json:"source"`        // 下载来源: "auto" 或 "manual"
}

// GetGeoIPConfig 获取 GeoIP 配置
func GetGeoIPConfig(c *gin.Context) {
	downloadURL, _ := models.GetSetting("geoip_download_url")
	if downloadURL == "" {
		downloadURL = DefaultGeoIPDownloadURL
	}

	useProxy, _ := models.GetSetting("geoip_use_proxy")
	proxyLink, _ := models.GetSetting("geoip_proxy_link")
	lastUpdate, _ := models.GetSetting("geoip_last_update")

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": GeoIPConfigResponse{
			DownloadURL: downloadURL,
			UseProxy:    useProxy == "true",
			ProxyLink:   proxyLink,
			LastUpdate:  lastUpdate,
		},
	})
}

// SaveGeoIPConfig 保存 GeoIP 配置
func SaveGeoIPConfig(c *gin.Context) {
	var req struct {
		DownloadURL string `json:"downloadUrl"`
		UseProxy    bool   `json:"useProxy"`
		ProxyLink   string `json:"proxyLink"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "参数错误"})
		return
	}

	// 保存配置
	if err := models.SetSetting("geoip_download_url", req.DownloadURL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "保存配置失败"})
		return
	}

	useProxyStr := "false"
	if req.UseProxy {
		useProxyStr = "true"
	}
	if err := models.SetSetting("geoip_use_proxy", useProxyStr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "保存配置失败"})
		return
	}

	if err := models.SetSetting("geoip_proxy_link", req.ProxyLink); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "保存配置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "保存成功"})
}

// GetGeoIPStatus 获取 GeoIP 状态
func GetGeoIPStatus(c *gin.Context) {
	info := geoip.GetDBInfo()

	downloadMu.Lock()
	downloading := isDownloading
	progress := downloadProgress
	errMsg := downloadError
	source := downloadSource
	downloadMu.Unlock()

	sizeFormatted := ""
	if info.Size > 0 {
		sizeFormatted = formatFileSize(info.Size)
	}

	modTime := ""
	if !info.ModTime.IsZero() {
		modTime = info.ModTime.Format("2006-01-02 15:04:05")
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": GeoIPStatusResponse{
			Available:     info.Available,
			Path:          info.Path,
			Size:          info.Size,
			SizeFormatted: sizeFormatted,
			ModTime:       modTime,
			Downloading:   downloading,
			Progress:      progress,
			Error:         errMsg,
			Source:        source,
		},
	})
}

// DownloadGeoIP 下载 GeoIP 数据库
func DownloadGeoIP(c *gin.Context) {
	downloadMu.Lock()
	if isDownloading {
		downloadMu.Unlock()
		c.JSON(http.StatusConflict, gin.H{"code": 409, "msg": "正在下载中，请稍候"})
		return
	}
	isDownloading = true
	downloadProgress = 0
	downloadError = ""
	downloadMu.Unlock()

	// 获取下载配置
	downloadURL, _ := models.GetSetting("geoip_download_url")
	if downloadURL == "" {
		downloadURL = DefaultGeoIPDownloadURL
	}

	useProxy, _ := models.GetSetting("geoip_use_proxy")
	proxyLink, _ := models.GetSetting("geoip_proxy_link")

	// 异步下载
	go func() {
		defer func() {
			downloadMu.Lock()
			isDownloading = false
			downloadMu.Unlock()
		}()

		err := downloadGeoIPFile(downloadURL, useProxy == "true", proxyLink)
		if err != nil {
			downloadMu.Lock()
			downloadError = err.Error()
			downloadMu.Unlock()
			utils.Error("下载 GeoIP 数据库失败: %v", err)
			return
		}

		// 更新最后更新时间
		models.SetSetting("geoip_last_update", time.Now().Format("2006-01-02 15:04:05"))

		// 重新加载数据库
		if err := geoip.Reload(); err != nil {
			downloadMu.Lock()
			downloadError = "数据库加载失败: " + err.Error()
			downloadMu.Unlock()
			return
		}

		downloadMu.Lock()
		downloadProgress = 100
		downloadMu.Unlock()

		utils.Info("GeoIP 数据库下载并加载成功")
	}()

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "开始下载"})
}

// StopGeoIPDownload 停止 GeoIP 数据库下载
func StopGeoIPDownload(c *gin.Context) {
	downloadMu.Lock()
	if !isDownloading {
		downloadMu.Unlock()
		c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "当前没有正在进行的下载任务"})
		return
	}
	// 发送停止信号
	if stopDownload != nil {
		close(stopDownload)
		stopDownload = nil
	}
	downloadMu.Unlock()

	c.JSON(http.StatusOK, gin.H{"code": 200, "msg": "已发送停止信号"})
}

// downloadGeoIPFile 下载 GeoIP 文件
func downloadGeoIPFile(url string, useProxy bool, proxyLink string) error {
	targetPath := config.GetGeoIPPath()

	// 确保目录存在
	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 创建 HTTP 客户端
	client, _, err := utils.CreateProxyHTTPClient(useProxy, proxyLink, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("创建 HTTP 客户端失败: %v", err)
	}

	// 发起请求
	utils.Info("开始下载 GeoIP 数据库: %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; SublinkPro/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("下载请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败: HTTP %d", resp.StatusCode)
	}

	// 创建临时文件
	tmpPath := targetPath + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer func() {
		file.Close()
		os.Remove(tmpPath) // 清理临时文件
	}()

	// 下载并跟踪进度
	totalSize := resp.ContentLength
	var downloaded int64 = 0
	buf := make([]byte, 32*1024) // 32KB buffer

	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := file.Write(buf[:n]); writeErr != nil {
				return fmt.Errorf("写入文件失败: %v", writeErr)
			}
			downloaded += int64(n)

			// 更新进度
			if totalSize > 0 {
				progress := int(float64(downloaded) / float64(totalSize) * 100)
				if progress > 99 {
					progress = 99 // 保留最后 1% 给加载步骤
				}
				downloadMu.Lock()
				downloadProgress = progress
				downloadMu.Unlock()
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取响应失败: %v", err)
		}
	}

	file.Close()

	// 验证文件大小
	fileInfo, err := os.Stat(tmpPath)
	if err != nil || fileInfo.Size() < 1024*1024 { // 文件至少 1MB
		return fmt.Errorf("下载的文件无效，文件过小")
	}

	// 移动临时文件到目标位置
	if err := os.Rename(tmpPath, targetPath); err != nil {
		// 如果 rename 失败（跨设备），尝试复制
		if copyErr := copyFile(tmpPath, targetPath); copyErr != nil {
			return fmt.Errorf("保存文件失败: %v", copyErr)
		}
	}

	utils.Info("GeoIP 数据库下载完成: %s (%.2f MB)", targetPath, float64(fileInfo.Size())/1024/1024)
	return nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// formatFileSize 格式化文件大小
func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// AutoDownloadGeoIP 启动时自动下载 GeoIP 数据库（异步）
// 如果数据库已存在则不下载
func AutoDownloadGeoIP() {
	// 检查是否已存在
	if geoip.IsAvailable() {
		return
	}

	downloadMu.Lock()
	if isDownloading {
		downloadMu.Unlock()
		return
	}
	isDownloading = true
	downloadProgress = 0
	downloadError = ""
	downloadSource = "auto"
	stopDownload = make(chan struct{})
	downloadMu.Unlock()

	utils.Info("[GeoIP] 数据库不存在，开始自动下载...")

	go func() {
		defer func() {
			downloadMu.Lock()
			isDownloading = false
			downloadSource = ""
			downloadMu.Unlock()
		}()

		// 获取下载配置
		downloadURL, _ := models.GetSetting("geoip_download_url")
		if downloadURL == "" {
			downloadURL = DefaultGeoIPDownloadURL
		}

		useProxy, _ := models.GetSetting("geoip_use_proxy")
		proxyLink, _ := models.GetSetting("geoip_proxy_link")

		err := downloadGeoIPFileWithProgress(downloadURL, useProxy == "true", proxyLink, true)
		if err != nil {
			downloadMu.Lock()
			downloadError = err.Error()
			downloadMu.Unlock()
			utils.Error("[GeoIP] 自动下载失败: %v", err)
			utils.Info("[GeoIP] 请在系统右上角菜单 -> GeoIP 数据库中手动配置下载")
			return
		}

		// 更新最后更新时间
		models.SetSetting("geoip_last_update", time.Now().Format("2006-01-02 15:04:05"))

		// 重新加载数据库
		if err := geoip.Reload(); err != nil {
			downloadMu.Lock()
			downloadError = "数据库加载失败: " + err.Error()
			downloadMu.Unlock()
			utils.Error("[GeoIP] 数据库加载失败: %v", err)
			return
		}

		downloadMu.Lock()
		downloadProgress = 100
		downloadMu.Unlock()

		utils.Info("[GeoIP] 数据库自动下载并加载成功")
	}()
}

// downloadGeoIPFileWithProgress 下载 GeoIP 文件（支持停止和控制台进度显示）
func downloadGeoIPFileWithProgress(url string, useProxy bool, proxyLink string, showConsoleProgress bool) error {
	targetPath := config.GetGeoIPPath()

	// 确保目录存在
	dir := filepath.Dir(targetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 创建 HTTP 客户端
	client, _, err := utils.CreateProxyHTTPClient(useProxy, proxyLink, 5*time.Minute)
	if err != nil {
		return fmt.Errorf("创建 HTTP 客户端失败: %v", err)
	}

	// 发起请求
	if showConsoleProgress {
		utils.Debug("[GeoIP] 下载地址: %s", url)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; SublinkPro/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("下载请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败: HTTP %d", resp.StatusCode)
	}

	// 创建临时文件
	tmpPath := targetPath + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer func() {
		file.Close()
		os.Remove(tmpPath) // 清理临时文件
	}()

	// 下载并跟踪进度
	totalSize := resp.ContentLength
	var downloaded int64 = 0
	buf := make([]byte, 32*1024) // 32KB buffer
	lastPrintProgress := -1

	for {
		// 检查停止信号
		downloadMu.Lock()
		stop := stopDownload
		downloadMu.Unlock()

		select {
		case <-stop:
			return fmt.Errorf("下载已被用户停止")
		default:
		}

		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := file.Write(buf[:n]); writeErr != nil {
				return fmt.Errorf("写入文件失败: %v", writeErr)
			}
			downloaded += int64(n)

			// 更新进度
			if totalSize > 0 {
				progress := int(float64(downloaded) / float64(totalSize) * 100)
				if progress > 99 {
					progress = 99 // 保留最后 1% 给加载步骤
				}
				downloadMu.Lock()
				downloadProgress = progress
				downloadMu.Unlock()

				// 控制台进度显示（每 10% 显示一次）
				if showConsoleProgress && progress/10 > lastPrintProgress/10 {
					utils.Debug("[GeoIP] 下载进度: %d%% (%.1f MB / %.1f MB)",
						progress,
						float64(downloaded)/1024/1024,
						float64(totalSize)/1024/1024)
					lastPrintProgress = progress
				}
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return fmt.Errorf("读取响应失败: %v", readErr)
		}
	}

	file.Close()

	// 验证文件大小
	fileInfo, err := os.Stat(tmpPath)
	if err != nil || fileInfo.Size() < 1024*1024 { // 文件至少 1MB
		return fmt.Errorf("下载的文件无效，文件过小")
	}

	// 移动临时文件到目标位置
	if err := os.Rename(tmpPath, targetPath); err != nil {
		// 如果 rename 失败（跨设备），尝试复制
		if copyErr := copyFile(tmpPath, targetPath); copyErr != nil {
			return fmt.Errorf("保存文件失败: %v", copyErr)
		}
	}

	if showConsoleProgress {
		utils.Info("[GeoIP] 下载完成: %s (%.2f MB)", targetPath, float64(fileInfo.Size())/1024/1024)
	}
	return nil
}

// IsGeoIPDownloading 检查是否正在下载
func IsGeoIPDownloading() bool {
	downloadMu.Lock()
	defer downloadMu.Unlock()
	return isDownloading
}

// GetGeoIPDownloadStatus 获取下载状态（供内部使用）
func GetGeoIPDownloadStatus() (downloading bool, progress int, errMsg string, source string) {
	downloadMu.Lock()
	defer downloadMu.Unlock()
	return isDownloading, downloadProgress, downloadError, downloadSource
}
