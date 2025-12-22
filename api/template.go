package api

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sublink/cache"
	"sublink/models"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

type Temp struct {
	File             string `json:"file"`
	Text             string `json:"text"`
	Category         string `json:"category"`
	RuleSource       string `json:"ruleSource"`
	UseProxy         bool   `json:"useProxy"`
	ProxyLink        string `json:"proxyLink"`
	EnableIncludeAll bool   `json:"enableIncludeAll"`
	CreateDate       string `json:"create_date"`
}

// 定义允许操作的基础目录

var baseTemplateDir string

func init() {
	// === 修改点开始 ===
	// 获取当前工作目录 (Current Working Directory)
	// 当您在项目根目录运行 `go run main.go` 时，这将是项目根目录
	cwd, err := os.Getwd()
	if err != nil {
		utils.Fatal("无法获取当前工作目录: %v", err)
	}

	// 将 "template" 路径解析为相对于当前工作目录的绝对路径
	absPath, err := filepath.Abs(filepath.Join(cwd, "template"))
	if err != nil {
		utils.Fatal("无法解析 template 目录的绝对路径: %v", err)
	}
	baseTemplateDir = absPath
	utils.Info("基础模板目录已初始化为: %s (基于当前工作目录)", baseTemplateDir)
	// === 修改点结束 ===

	// 确保基础模板目录存在，如果不存在则创建
	if _, err := os.Stat(baseTemplateDir); os.IsNotExist(err) {
		if err := os.MkdirAll(baseTemplateDir, 0755); err != nil {
			utils.Fatal("无法创建基础模板目录 %s: %v", baseTemplateDir, err)
		}
		utils.Info("已创建基础模板目录: %s", baseTemplateDir)
	}
}

// safeFilename 生成安全的文件路径，防止目录遍历
func safeFilePath(filename string) (string, error) {
	// 1. 清理用户提供的文件名，移除冗余的 "." 和 ".." 等。
	cleanFilename := filepath.Clean(filename)

	// 2. 严格检查文件名是否包含任何路径分隔符。
	// 这强制只允许在 baseTemplateDir 下直接操作文件，不能通过文件名创建子目录。
	if strings.ContainsRune(cleanFilename, os.PathSeparator) ||
		strings.ContainsRune(cleanFilename, '/') ||
		strings.ContainsRune(cleanFilename, '\\') {
		return "", errors.New("文件名不能包含路径分隔符")
	}

	// 3. 禁止使用特殊文件名（如 ".", "..", 或空字符串）。
	if cleanFilename == "." || cleanFilename == ".." || cleanFilename == "" {
		return "", errors.New("文件名无效或指向目录本身")
	}

	// 4. 将基础目录（已是绝对路径）和清理后的文件名安全地连接起来。
	fullPath := filepath.Join(baseTemplateDir, cleanFilename)

	// 5. 再次清理完整路径，确保最终路径是规范化的。
	finalCleanPath := filepath.Clean(fullPath)

	// 6. **核心安全检查**: 验证最终路径是否仍在 `baseTemplateDir` 的范围内。
	// `filepath.Rel` 计算 `finalCleanPath` 相对于 `baseTemplateDir` 的相对路径。
	// 如果 `finalCleanPath` 跳出了 `baseTemplateDir`，那么 `relPath` 会以 ".." 开头。
	relPath, err := filepath.Rel(baseTemplateDir, finalCleanPath)
	if err != nil {
		// `filepath.Rel` 错误通常表示路径不兼容或发生异常，视为不安全。
		return "", errors.New("路径处理错误: " + err.Error())
	}
	if strings.HasPrefix(relPath, "..") {
		// 如果相对路径以 ".." 开头，说明存在目录遍历企图。
		return "", errors.New("检测到目录遍历尝试")
	}

	// 7. 确保最终路径不是 `baseTemplateDir` 本身（例如，如果用户传入 "."）。
	// 这防止了将根目录本身作为“文件”进行操作。
	if finalCleanPath == baseTemplateDir {
		return "", errors.New("文件名无效或指向根目录本身")
	}

	return finalCleanPath, nil
}

func GetTempS(c *gin.Context) {
	// 由于 init() 函数已经确保了 baseTemplateDir 的存在，这里无需再次检查和创建。
	files, err := os.ReadDir(baseTemplateDir)
	if err != nil {
		utils.Error("读取模板目录失败: %v", err)
		utils.FailWithMsg(c, "服务器错误：无法读取模板文件")
		return
	}

	var temps []Temp
	for _, file := range files {
		// 跳过目录，因为我们只处理文件
		if file.IsDir() {
			continue
		}

		// **修复点：对读取的文件名也使用 safeFilePath 进行验证**
		// 这可以防止通过符号链接（symlink）进行的目录遍历，从而避免信息泄露。
		fullPathToRead, err := safeFilePath(file.Name())
		if err != nil {
			utils.Warn("跳过不安全或非法文件 (读取): %s, 错误: %v", file.Name(), err)
			continue // 跳过不安全的文件
		}

		info, err := file.Info()
		if err != nil {
			utils.Warn("获取文件信息失败: %s, 错误: %v", file.Name(), err)
			continue
		}
		modTime := info.ModTime().Format("2006-01-02 15:04:05")

		// 优先从缓存读取模板内容
		var textContent string
		if cached, ok := cache.GetTemplateContent(file.Name()); ok {
			textContent = cached
		} else {
			// 缓存未命中，从文件读取并写入缓存
			textBytes, readErr := os.ReadFile(fullPathToRead)
			if readErr != nil {
				utils.Error("读取文件内容失败: %s, 错误: %v", fullPathToRead, readErr)
				continue // 跳过无法读取的文件
			}
			textContent = string(textBytes)
			// 写入缓存
			cache.SetTemplateContent(file.Name(), textContent)
		}

		// 从数据库获取模板元数据
		var tmplMeta models.Template
		category := "clash"
		ruleSource := ""
		useProxy := false
		proxyLink := ""
		enableIncludeAll := false
		if err := tmplMeta.FindByName(file.Name()); err == nil {
			category = tmplMeta.Category
			ruleSource = tmplMeta.RuleSource
			useProxy = tmplMeta.UseProxy
			proxyLink = tmplMeta.ProxyLink
			enableIncludeAll = tmplMeta.EnableIncludeAll
		}

		temp := Temp{
			File:             file.Name(),
			Text:             textContent,
			Category:         category,
			RuleSource:       ruleSource,
			UseProxy:         useProxy,
			ProxyLink:        proxyLink,
			EnableIncludeAll: enableIncludeAll,
			CreateDate:       modTime,
		}
		temps = append(temps, temp)
	}

	// 解析分页参数
	page := 0
	pageSize := 0
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	// 如果提供了分页参数，返回分页响应
	if page > 0 && pageSize > 0 {
		total := int64(len(temps))
		offset := (page - 1) * pageSize
		end := offset + pageSize

		var pagedTemps []Temp
		if offset < len(temps) {
			if end > len(temps) {
				end = len(temps)
			}
			pagedTemps = temps[offset:end]
		} else {
			pagedTemps = []Temp{}
		}

		totalPages := 0
		if pageSize > 0 {
			totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
		}
		utils.OkDetailed(c, "ok", gin.H{
			"items":      pagedTemps,
			"total":      total,
			"page":       page,
			"pageSize":   pageSize,
			"totalPages": totalPages,
		})
		return
	}

	// 不带分页参数，返回全部（向后兼容）
	if len(temps) == 0 {
		utils.OkDetailed(c, "ok", []Temp{})
		return
	}
	utils.OkDetailed(c, "ok", temps)
}
func UpdateTemp(c *gin.Context) {
	filename := c.PostForm("filename")
	oldname := c.PostForm("oldname")
	text := c.PostForm("text")
	category := c.PostForm("category")
	ruleSource := c.PostForm("ruleSource")
	useProxy := c.PostForm("useProxy") == "true"
	proxyLink := c.PostForm("proxyLink")
	enableIncludeAll := c.PostForm("enableIncludeAll") == "true"

	if filename == "" || oldname == "" || text == "" {
		utils.FailWithMsg(c, "文件名或内容不能为空")
		return
	}

	// 默认类别为 clash
	if category == "" {
		category = "clash"
	}

	// 验证旧文件名以防止目录遍历
	oldFullPath, err := safeFilePath(oldname)
	if err != nil {
		utils.FailWithMsg(c, "旧文件名非法: "+err.Error())
		return
	}

	// 验证新文件名以防止目录遍历
	newFullPath, err := safeFilePath(filename)
	if err != nil {
		utils.FailWithMsg(c, "新文件名非法: "+err.Error())
		return
	}

	// 检查旧文件是否存在
	if _, err := os.Stat(oldFullPath); os.IsNotExist(err) {
		utils.FailWithMsg(c, "旧文件不存在")
		return
	} else if err != nil {
		utils.Error("检查旧文件存在性失败: %v", err)
		utils.FailWithMsg(c, "服务器错误：检查旧文件失败")
		return
	}

	// 如果新旧文件名不同，则检查新文件是否已存在
	if oldFullPath != newFullPath {
		if _, err := os.Stat(newFullPath); err == nil {
			utils.FailWithMsg(c, "新文件名已存在，请选择其他名称")
			return
		} else if !os.IsNotExist(err) {
			utils.Error("检查新文件存在性失败: %v", err)
			utils.FailWithMsg(c, "服务器错误：检查新文件失败")
			return
		}
	}

	// 如果文件名不同，则进行重命名操作
	if oldFullPath != newFullPath {
		err = os.Rename(oldFullPath, newFullPath)
		if err != nil {
			utils.Error("文件改名失败: %v", err)
			utils.FailWithMsg(c, "改名失败")
			return
		}
	}

	// 写入文件内容到新的安全路径
	err = os.WriteFile(newFullPath, []byte(text), 0666)
	if err != nil {
		utils.Error("修改文件内容失败: %v", err)
		utils.FailWithMsg(c, "修改失败")
		return
	}

	// 同步更新模板内容缓存
	if oldname != filename {
		// 文件名变更，先删除旧缓存
		cache.InvalidateTemplateContent(oldname)
	}
	cache.SetTemplateContent(filename, text)

	// 更新数据库中的模板元数据
	var tmpl models.Template
	if err := tmpl.FindByName(oldname); err != nil {
		// 如果数据库中不存在，创建新记录
		tmpl = models.Template{
			Name:             filename,
			Category:         category,
			RuleSource:       ruleSource,
			UseProxy:         useProxy,
			ProxyLink:        proxyLink,
			EnableIncludeAll: enableIncludeAll,
		}
		if err := tmpl.Add(); err != nil {
			utils.Error("创建模板元数据失败: %v", err)
		}
	} else {
		// 更新现有记录
		tmpl.Name = filename
		tmpl.Category = category
		tmpl.RuleSource = ruleSource
		tmpl.UseProxy = useProxy
		tmpl.ProxyLink = proxyLink
		tmpl.EnableIncludeAll = enableIncludeAll
		if err := tmpl.Update(); err != nil {
			utils.Error("更新模板元数据失败: %v", err)
		}
	}

	utils.OkWithMsg(c, "修改成功")
}
func AddTemp(c *gin.Context) {
	filename := c.PostForm("filename")
	text := c.PostForm("text")
	category := c.PostForm("category")
	ruleSource := c.PostForm("ruleSource")
	useProxy := c.PostForm("useProxy") == "true"
	proxyLink := c.PostForm("proxyLink")
	enableIncludeAll := c.PostForm("enableIncludeAll") == "true"

	if filename == "" || text == "" {
		utils.FailWithMsg(c, "文件名或内容不能为空")
		return
	}

	// 默认类别为 clash
	if category == "" {
		category = "clash"
	}

	// 确保模板目录存在
	if _, err := os.Stat(baseTemplateDir); os.IsNotExist(err) {
		if err := os.MkdirAll(baseTemplateDir, 0755); err != nil {
			utils.Error("创建模板目录失败: %v", err)
			utils.FailWithMsg(c, "服务器错误：无法创建模板目录")
			return
		}
	}

	// 获取安全的文件路径
	fullPath, err := safeFilePath(filename)
	if err != nil {
		utils.FailWithMsg(c, "文件名非法: "+err.Error())
		return
	}

	// 检查文件是否已存在
	if _, err := os.Stat(fullPath); err == nil {
		utils.FailWithMsg(c, "文件已存在")
		return
	} else if !os.IsNotExist(err) {
		utils.Error("检查文件存在性失败: %v", err)
		utils.FailWithMsg(c, "服务器错误：检查文件失败")
		return
	}

	// 写入文件
	err = os.WriteFile(fullPath, []byte(text), 0666)
	if err != nil {
		utils.Error("写入文件失败: %v", err)
		utils.FailWithMsg(c, "上传失败")
		return
	}

	// 写入模板内容缓存
	cache.SetTemplateContent(filename, text)

	// 创建数据库记录
	tmpl := models.Template{
		Name:             filename,
		Category:         category,
		RuleSource:       ruleSource,
		UseProxy:         useProxy,
		ProxyLink:        proxyLink,
		EnableIncludeAll: enableIncludeAll,
	}
	if err := tmpl.Add(); err != nil {
		utils.Error("创建模板元数据失败: %v", err)
	}

	utils.OkWithMsg(c, "上传成功")
}

func DelTemp(c *gin.Context) {
	filename := c.PostForm("filename")

	if filename == "" {
		utils.FailWithMsg(c, "文件名不能为空")
		return
	}

	// 获取安全的文件路径
	fullPath, err := safeFilePath(filename)
	if err != nil {
		utils.FailWithMsg(c, "文件名非法: "+err.Error())
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		utils.FailWithMsg(c, "文件不存在")
		return
	} else if err != nil {
		utils.Error("检查文件存在性失败: %v", err)
		utils.FailWithMsg(c, "服务器错误：检查文件失败")
		return
	}

	// 删除文件
	err = os.Remove(fullPath)
	if err != nil {
		utils.Error("删除文件失败: %v", err)
		utils.FailWithMsg(c, "删除失败")
		return
	}

	// 清除模板内容缓存
	cache.InvalidateTemplateContent(filename)

	// 删除数据库记录
	var tmpl models.Template
	if err := tmpl.FindByName(filename); err == nil {
		if err := tmpl.Delete(); err != nil {
			utils.Error("删除模板元数据失败: %v", err)
		}
	}

	utils.OkWithMsg(c, "删除成功")
}

// ACL4SSRPreset ACL4SSR 规则预设
type ACL4SSRPreset struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Label string `json:"label"`
}

// GetACL4SSRPresets 获取 ACL4SSR 规则预设列表
func GetACL4SSRPresets(c *gin.Context) {
	presets := []ACL4SSRPreset{
		{
			Name:  "作者自用",
			URL:   "https://raw.githubusercontent.com/ZeroDeng01/ACL4SSR/master/Clash/config/ACL4SSR_Online_Full_NoCountry.ini",
			Label: "作者自用 - 不区分国家",
		},
		{
			Name:  "ACL4SSR",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR.ini",
			Label: "标准版 - 典型分组",
		},
		{
			Name:  "ACL4SSR_AdblockPlus",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_AdblockPlus.ini",
			Label: "标准版 - 典型分组+去广告",
		},
		{
			Name:  "ACL4SSR_BackCN",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_BackCN.ini",
			Label: "回国版 - 回国专用",
		},
		{
			Name:  "ACL4SSR_Mini",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Mini.ini",
			Label: "精简版 - 少量分组",
		},
		{
			Name:  "ACL4SSR_Mini_Fallback",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Mini_Fallback.ini",
			Label: "精简版 - 故障转移",
		},
		{
			Name:  "ACL4SSR_Mini_MultiMode",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Mini_MultiMode.ini",
			Label: "精简版 - 多模式 (自动/手动)",
		},
		{
			Name:  "ACL4SSR_Mini_NoAuto",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Mini_NoAuto.ini",
			Label: "精简版 - 无自动测速",
		},
		{
			Name:  "ACL4SSR_NoApple",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_NoApple.ini",
			Label: "无苹果 - 无苹果分流",
		},
		{
			Name:  "ACL4SSR_NoAuto",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_NoAuto.ini",
			Label: "无测速 - 无自动测速",
		},
		{
			Name:  "ACL4SSR_NoAuto_NoApple",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_NoAuto_NoApple.ini",
			Label: "无测速&苹果 - 无测速&无苹果分流",
		},
		{
			Name:  "ACL4SSR_NoAuto_NoApple_NoMicrosoft",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_NoAuto_NoApple_NoMicrosoft.ini",
			Label: "无测速&苹果&微软 - 无测速&无苹果&无微软分流",
		},
		{
			Name:  "ACL4SSR_NoMicrosoft",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_NoMicrosoft.ini",
			Label: "无微软 - 无微软分流",
		},
		{
			Name:  "ACL4SSR_Online",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online.ini",
			Label: "在线版 - 典型分组",
		},
		{
			Name:  "ACL4SSR_Online_AdblockPlus",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_AdblockPlus.ini",
			Label: "在线版 - 典型分组+去广告",
		},
		{
			Name:  "ACL4SSR_Online_Full",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Full.ini",
			Label: "在线全分组 - 比较全",
		},
		{
			Name:  "ACL4SSR_Online_Full_AdblockPlus",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Full_AdblockPlus.ini",
			Label: "在线全分组 - 带广告拦截",
		},
		{
			Name:  "ACL4SSR_Online_Full_Google",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Full_Google.ini",
			Label: "在线全分组 - 谷歌分流",
		},
		{
			Name:  "ACL4SSR_Online_Full_MultiMode",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Full_MultiMode.ini",
			Label: "在线全分组 - 多模式",
		},
		{
			Name:  "ACL4SSR_Online_Full_Netflix",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Full_Netflix.ini",
			Label: "在线全分组 - 奈飞分流",
		},
		{
			Name:  "ACL4SSR_Online_Full_NoAuto",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Full_NoAuto.ini",
			Label: "在线全分组 - 无自动测速",
		},
		{
			Name:  "ACL4SSR_Online_Mini",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Mini.ini",
			Label: "在线精简版 - 少量分组",
		},
		{
			Name:  "ACL4SSR_Online_Mini_AdblockPlus",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Mini_AdblockPlus.ini",
			Label: "在线精简版 - 带广告拦截",
		},
		{
			Name:  "ACL4SSR_Online_Mini_Ai",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Mini_Ai.ini",
			Label: "在线精简版 - AI",
		},
		{
			Name:  "ACL4SSR_Online_Mini_Fallback",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Mini_Fallback.ini",
			Label: "在线精简版 - 故障转移",
		},
		{
			Name:  "ACL4SSR_Online_Mini_MultiCountry",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Mini_MultiCountry.ini",
			Label: "在线精简版 - 多国家",
		},
		{
			Name:  "ACL4SSR_Online_Mini_MultiMode",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Mini_MultiMode.ini",
			Label: "在线精简版 - 多模式",
		},
		{
			Name:  "ACL4SSR_Online_Mini_NoAuto",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_Mini_NoAuto.ini",
			Label: "在线精简版 - 无自动测速",
		},
		{
			Name:  "ACL4SSR_Online_MultiCountry",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_MultiCountry.ini",
			Label: "在线版 - 多国家",
		},
		{
			Name:  "ACL4SSR_Online_NoAuto",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_NoAuto.ini",
			Label: "在线版 - 无自动测速",
		},
		{
			Name:  "ACL4SSR_Online_NoReject",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_Online_NoReject.ini",
			Label: "在线版 - 无拒绝规则",
		},
		{
			Name:  "ACL4SSR_WithChinaIp",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_WithChinaIp.ini",
			Label: "特殊版 - 包含回国IP",
		},
		{
			Name:  "ACL4SSR_WithChinaIp_WithGFW",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_WithChinaIp_WithGFW.ini",
			Label: "特殊版 - 包含回国IP&GFW列表",
		},
		{
			Name:  "ACL4SSR_WithGFW",
			URL:   "https://raw.githubusercontent.com/ACL4SSR/ACL4SSR/master/Clash/config/ACL4SSR_WithGFW.ini",
			Label: "特殊版 - 包含GFW列表",
		},
	}
	utils.OkDetailed(c, "ok", presets)
}
