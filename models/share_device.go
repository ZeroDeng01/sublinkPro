package models

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"sublink/database"
	"sublink/utils"

	"gorm.io/gorm"
)

// ShareDevice identifies a client installation, not a trusted physical device.
type ShareDevice struct {
	ID           int       `gorm:"primaryKey" json:"id"`
	ShareID      int       `gorm:"uniqueIndex:idx_share_hwid;index" json:"share_id"`
	HWIDHash     string    `gorm:"uniqueIndex:idx_share_hwid;size:64" json:"-"`
	Name         string    `gorm:"size:100" json:"name"`
	OS           string    `gorm:"size:100" json:"os"`
	Model        string    `gorm:"size:200" json:"model"`
	Revoked      bool      `gorm:"default:false" json:"revoked"`
	CreatedAt    time.Time `json:"created_at"`
	LastAccessAt time.Time `json:"last_access_at"`
}

type ShareDeviceIdentity struct{ UserAgent, HWID, OS, Model string }

type ShareAccessError struct{ Code, Message string }

func (e *ShareAccessError) Error() string { return e.Message }

func deviceDenied(code, message string) error { return &ShareAccessError{Code: code, Message: message} }

func IsShareNotFoundError(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }

// lockShare uses a write before any read. SQLite obtains its writer lock while
// MySQL/PostgreSQL lock the share row. All device/policy mutations use this lock.
func lockShare(tx *gorm.DB, id int) (*SubscriptionShare, error) {
	result := tx.Model(&SubscriptionShare{}).Where("id = ?", id).UpdateColumn("device_revision", gorm.Expr("device_revision + 1"))
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected != 1 {
		return nil, gorm.ErrRecordNotFound
	}
	var share SubscriptionShare
	if err := tx.First(&share, id).Error; err != nil {
		return nil, err
	}
	return &share, nil
}

func validateDeviceLimit(limit int) error {
	if limit < 0 || limit > 10000 {
		return fmt.Errorf("设备上限必须是 0 到 10000 之间的整数")
	}
	return nil
}

func activeDeviceCount(tx *gorm.DB, shareID int) (int64, error) {
	var count int64
	err := tx.Model(&ShareDevice{}).Where("share_id = ? AND revoked = ?", shareID, false).Count(&count).Error
	return count, err
}

func checkDeviceLimit(tx *gorm.DB, shareID, limit int) error {
	if err := validateDeviceLimit(limit); err != nil {
		return err
	}
	if limit == 0 {
		return nil
	}
	count, err := activeDeviceCount(tx, shareID)
	if err != nil {
		return err
	}
	if count > int64(limit) {
		return deviceDenied("device_limit_below_bound", "请先撤销设备，再降低设备上限")
	}
	return nil
}

// CheckShareDevice runs both before rendering and immediately before publishing
// the finished response. Only the latter is allowed to allocate a device slot.
func CheckShareDevice(id int, token string, identity ShareDeviceIdentity, bind bool) error {
	var share SubscriptionShare
	if err := database.DB.First(&share, id).Error; err != nil {
		return err
	}
	if err := validateShareDeviceAccess(&share, token, identity); err != nil {
		return err
	}
	// Unlimited downloads never need a device transaction or a revision write.
	// A limited download rechecks the current policy under the admission lock.
	if share.MaxDevices == 0 {
		return nil
	}
	if !bind {
		return checkShareDeviceAdmission(database.DB, &share, identity, false)
	}
	return database.DB.Transaction(func(tx *gorm.DB) error {
		share, err := lockShare(tx, id)
		if err != nil {
			return err
		}
		if err := validateShareDeviceAccess(share, token, identity); err != nil {
			return err
		}
		if share.MaxDevices == 0 {
			return nil
		}
		return checkShareDeviceAdmission(tx, share, identity, true)
	})
}

func validateShareDeviceAccess(share *SubscriptionShare, token string, identity ShareDeviceIdentity) error {
	if share.Token != token || share.IsExpired() {
		return deviceDenied("share_unavailable", "分享已失效或已禁用")
	}
	if share.MaxDevices == 0 {
		return nil
	}
	if len(identity.HWID) == 0 || len(identity.HWID) > 512 || strings.TrimSpace(identity.HWID) != identity.HWID {
		return deviceDenied("hwid_required", "缺少有效 X-HWID，请使用支持此请求头的客户端并开启发送")
	}
	for _, c := range identity.HWID {
		// Proxies may combine repeated header fields into a comma-separated value.
		if c < 33 || c > 126 || c == ',' {
			return deviceDenied("hwid_invalid", "X-HWID 格式无效，必须提供单个设备标识")
		}
	}
	return nil
}

func checkShareDeviceAdmission(tx *gorm.DB, share *SubscriptionShare, identity ShareDeviceIdentity, bind bool) error {
	id := share.ID
	sum := sha256.Sum256([]byte(identity.HWID))
	hash := hex.EncodeToString(sum[:])
	var device ShareDevice
	err := tx.Where("share_id = ? AND hw_id_hash = ?", id, hash).First(&device).Error
	if err == nil {
		if device.Revoked {
			return deviceDenied("device_revoked", "此设备已被撤销，请联系管理员")
		}
		if bind {
			return tx.Model(&device).Update("last_access_at", time.Now()).Error
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	count, err := activeDeviceCount(tx, id)
	if err != nil {
		return err
	}
	if count >= int64(share.MaxDevices) {
		return deviceDenied("device_limit_exceeded", "设备数量已达上限，请联系管理员换机")
	}
	if !bind {
		return nil
	}
	device = ShareDevice{ShareID: id, HWIDHash: hash, OS: truncateDeviceText(identity.OS, 100), Model: truncateDeviceText(identity.Model, 200), LastAccessAt: time.Now()}
	return tx.Create(&device).Error
}

func truncateDeviceText(value string, maxLen int) string {
	runes := []rune(strings.TrimSpace(value))
	if len(runes) > maxLen {
		runes = runes[:maxLen]
	}
	return string(runes)
}

func ListShareDevices(shareID int) ([]ShareDevice, error) {
	devices := []ShareDevice{}
	err := database.DB.Where("share_id = ?", shareID).Order("id ASC").Find(&devices).Error
	return devices, err
}

func UpdateShareDevice(shareID, deviceID int, revoked *bool, name *string) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		share, err := lockShare(tx, shareID)
		if err != nil {
			return err
		}
		var device ShareDevice
		if err := tx.Where("share_id = ? AND id = ?", shareID, deviceID).First(&device).Error; err != nil {
			return err
		}
		updates := map[string]any{}
		if name != nil {
			if len([]rune(*name)) > 100 {
				return fmt.Errorf("设备备注最多 100 字")
			}
			updates["name"] = *name
		}
		if revoked != nil {
			if !*revoked && device.Revoked && share.MaxDevices > 0 {
				count, err := activeDeviceCount(tx, shareID)
				if err != nil {
					return err
				}
				if count >= int64(share.MaxDevices) {
					return deviceDenied("device_limit_exceeded", "设备数量已达上限")
				}
			}
			updates["revoked"] = *revoked
		}
		return tx.Model(&device).Updates(updates).Error
	})
}

func ResetShareDevices(shareID int) (string, error) {
	token, err := GenerateToken()
	if err != nil {
		return "", err
	}
	err = database.DB.Transaction(func(tx *gorm.DB) error {
		if _, err := lockShare(tx, shareID); err != nil {
			return err
		}
		if err := tx.Model(&ShareDevice{}).Where("share_id = ?", shareID).Update("revoked", true).Error; err != nil {
			return err
		}
		return tx.Model(&SubscriptionShare{}).Where("id = ?", shareID).Update("token", token).Error
	})
	if err == nil {
		if _, cacheErr := refreshSubscriptionShareCache(shareID); cacheErr != nil {
			invalidateSubscriptionShareCache(shareID)
			utils.Warn("分享设备已重置，但刷新分享缓存失败: %v", cacheErr)
		}
	}
	return token, err
}

func PopulateShareDeviceCounts(shares []SubscriptionShare) error {
	if len(shares) == 0 {
		return nil
	}
	ids := make([]int, len(shares))
	for i := range shares {
		ids[i] = shares[i].ID
	}
	var counts []struct {
		ShareID int
		Count   int
	}
	if err := database.DB.Model(&ShareDevice{}).Select("share_id, count(*) AS count").Where("share_id IN ? AND revoked = ?", ids, false).Group("share_id").Scan(&counts).Error; err != nil {
		return err
	}
	byID := map[int]int{}
	for _, c := range counts {
		byID[c.ShareID] = c.Count
	}
	for i := range shares {
		shares[i].DeviceCount = byID[shares[i].ID]
	}
	return nil
}
