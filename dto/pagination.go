package dto

// PaginationRequest 通用分页请求参数
type PaginationRequest struct {
	Page     int `form:"page" json:"page"`         // 页码，1-indexed
	PageSize int `form:"pageSize" json:"pageSize"` // 每页条数
}

// PaginationResponse 通用分页响应
type PaginationResponse struct {
	Items      interface{} `json:"items"`      // 数据列表
	Total      int64       `json:"total"`      // 总条数
	Page       int         `json:"page"`       // 当前页码
	PageSize   int         `json:"pageSize"`   // 每页条数
	TotalPages int         `json:"totalPages"` // 总页数
}

// NewPaginationResponse 创建分页响应
func NewPaginationResponse(items interface{}, total int64, page, pageSize int) PaginationResponse {
	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	return PaginationResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}

// GetOffset 获取分页偏移量
func (p *PaginationRequest) GetOffset() int {
	if p.Page < 1 {
		p.Page = 1
	}
	return (p.Page - 1) * p.PageSize
}

// Normalize 规范化分页参数
func (p *PaginationRequest) Normalize(defaultPageSize, maxPageSize int) {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = defaultPageSize
	}
	if p.PageSize > maxPageSize {
		p.PageSize = maxPageSize
	}
}

// IsPaginated 检查是否需要分页
func (p *PaginationRequest) IsPaginated() bool {
	return p.Page > 0 && p.PageSize > 0
}
