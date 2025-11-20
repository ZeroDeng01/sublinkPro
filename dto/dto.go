package dto

import "time"

// 订阅节点排序请求体结构
type SubcriptionNodeSortUpdate struct {
	ID       int            `json:"ID" binding:"required"`
	NodeSort []NodeSortItem `json:"NodeSort" binding:"required"`
}

type NodeSortItem struct {
	ID      int    `json:"ID"`
	Name    string `json:"Name"`
	Sort    int    `json:"Sort"`
	IsGroup *bool  `json:"IsGroup"` // 标识是否为分组，使用指针以区分false和未设置
}

// UserAccessKey 用户访问密钥请求体结构
type UserAccessKey struct {
	UserName    string `binding:"required"`
	ExpiredAt   *time.Time
	Description string
}

// SubSchedulerAddRequest 订阅调度添加请求体结构
type SubSchedulerAddRequest struct {
	ID       int    `json:"id"`
	Name     string `json:"name" binding:"required"`
	URL      string `json:"url" binding:"required,url"`
	CronExpr string `json:"cron_expr" binding:"required"`
	Enabled  bool   `json:"enabled"`
	Group    string `json:"group"`
}
