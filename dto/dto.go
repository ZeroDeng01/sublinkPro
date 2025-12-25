package dto

import "time"

// 订阅节点排序请求体结构
type SubcriptionNodeSortUpdate struct {
	ID       int            `json:"ID" binding:"required"`
	NodeSort []NodeSortItem `json:"NodeSort" binding:"required"`
}

type NodeSortItem struct {
	ID      int    `json:"ID"` // 节点ID（非分组时必需）
	Name    string `json:"Name"`
	Sort    int    `json:"Sort"`
	IsGroup *bool  `json:"IsGroup"` // 标识是否为分组，使用指针以区分false和未设置
}

// UserAccessKey 用户访问密钥请求体结构
type UserAccessKey struct {
	UserName    string     `json:"username" binding:"required"`
	ExpiredAt   *time.Time `json:"expiredAt"`
	Description string     `json:"description"`
}

// AirportRequest 机场添加/更新请求体结构
type AirportRequest struct {
	ID                int    `json:"id"`
	Name              string `json:"name" binding:"required"`
	URL               string `json:"url" binding:"required,url"`
	CronExpr          string `json:"cronExpr" binding:"required"`
	Enabled           bool   `json:"enabled"`
	Group             string `json:"group"`
	DownloadWithProxy bool   `json:"downloadWithProxy"`
	ProxyLink         string `json:"proxyLink"`
	UserAgent         string `json:"userAgent"`
	FetchUsageInfo    bool   `json:"fetchUsageInfo"` // 是否获取用量信息
	SkipTLSVerify     bool   `json:"skipTLSVerify"`  // 是否跳过TLS证书验证
	Remark            string `json:"remark"`         // 备注信息
	Logo              string `json:"logo"`           // Logo配置
}

// BatchSortRequest 批量排序请求
type BatchSortRequest struct {
	ID        int    `json:"ID" binding:"required"`        // 订阅ID
	SortBy    string `json:"sortBy" binding:"required"`    // 排序字段: source, name, protocol, delay, speed, country
	SortOrder string `json:"sortOrder" binding:"required"` // 排序方向: asc, desc
}
