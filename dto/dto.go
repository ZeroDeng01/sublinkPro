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

// SubSchedulerAddRequest 订阅调度添加请求体结构
type SubSchedulerAddRequest struct {
	ID                int    `json:"ID"`
	Name              string `json:"Name" binding:"required"`
	URL               string `json:"URL" binding:"required,url"`
	CronExpr          string `json:"CronExpr" binding:"required"`
	Enabled           bool   `json:"Enabled"`
	Group             string `json:"Group"`
	DownloadWithProxy bool   `json:"DownloadWithProxy"`
	ProxyLink         string `json:"ProxyLink"`
	UserAgent         string `json:"UserAgent"`
}
