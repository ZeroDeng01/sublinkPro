package models

import (
	"strconv"
	"sublink/cache"
	"sublink/database"
	"sublink/utils"
)

type SubLogs struct {
	ID            int
	IP            string
	Date          string
	Addr          string
	Count         int
	SubcriptionID int
	ShareID       int // 关联的分享ID，用于区分不同分享入口
}

// subLogsCache 使用新的泛型缓存
var subLogsCache *cache.MapCache[int, SubLogs]

func init() {
	subLogsCache = cache.NewMapCache(func(sl SubLogs) int { return sl.ID })
	subLogsCache.AddIndex("subcriptionID", func(sl SubLogs) string { return strconv.Itoa(sl.SubcriptionID) })
	subLogsCache.AddIndex("shareID", func(sl SubLogs) string { return strconv.Itoa(sl.ShareID) })
}

// InitSubLogsCache 初始化订阅日志缓存
func InitSubLogsCache() error {
	utils.Info("开始加载订阅日志到缓存")
	var sublogs []SubLogs
	if err := database.DB.Find(&sublogs).Error; err != nil {
		return err
	}

	subLogsCache.LoadAll(sublogs)
	utils.Info("订阅日志缓存初始化完成，共加载 %d 条记录", subLogsCache.Count())

	cache.Manager.Register("sublogs", subLogsCache)
	return nil
}

// Add 添加IP (Write-Through)
func (iplog *SubLogs) Add() error {
	err := database.DB.Create(iplog).Error
	if err != nil {
		return err
	}
	subLogsCache.Set(iplog.ID, *iplog)
	return nil
}

// Find 查找IP
func (iplog *SubLogs) Find(id int) error {
	// 先从缓存查找
	logs := subLogsCache.GetByIndex("subcriptionID", strconv.Itoa(id))
	for _, l := range logs {
		if l.IP == iplog.IP {
			*iplog = l
			return nil
		}
	}
	return database.DB.Where("ip = ? and subcription_id  = ?", iplog.IP, id).First(iplog).Error
}

// FindByShare 根据IP、订阅ID和分享ID精确查找
func (iplog *SubLogs) FindByShare(subcriptionID, shareID int) error {
	// 先从缓存查找
	if shareID > 0 {
		logs := subLogsCache.GetByIndex("shareID", strconv.Itoa(shareID))
		for _, l := range logs {
			if l.IP == iplog.IP && l.SubcriptionID == subcriptionID {
				*iplog = l
				return nil
			}
		}
		return database.DB.Where("ip = ? and subcription_id = ? and share_id = ?", iplog.IP, subcriptionID, shareID).First(iplog).Error
	}
	// 如果没有shareID，回退到订阅级别
	return iplog.Find(subcriptionID)
}

// Update 更新IP (Write-Through)
func (iplog *SubLogs) Update() error {
	err := database.DB.Where("id = ? or ip = ?", iplog.ID, iplog.IP).Updates(iplog).Error
	if err != nil {
		return err
	}
	// 从DB读取更新后的数据
	var updated SubLogs
	if err := database.DB.First(&updated, iplog.ID).Error; err == nil {
		subLogsCache.Set(updated.ID, updated)
	}
	return nil
}

// List 获取IP列表
func (iplog *SubLogs) List() ([]SubLogs, error) {
	return subLogsCache.GetAll(), nil
}

// GetBySubcriptionID 根据订阅ID获取日志列表
func GetSubLogsBySubcriptionID(subcriptionID int) []SubLogs {
	return subLogsCache.GetByIndex("subcriptionID", strconv.Itoa(subcriptionID))
}

// GetSubLogsByShareID 根据分享ID获取日志列表
func GetSubLogsByShareID(shareID int) []SubLogs {
	return subLogsCache.GetByIndex("shareID", strconv.Itoa(shareID))
}

// Del 删除IP (Write-Through)
func (iplog *SubLogs) Del() error {
	err := database.DB.Delete(iplog).Error
	if err != nil {
		return err
	}
	subLogsCache.Delete(iplog.ID)
	return nil
}
