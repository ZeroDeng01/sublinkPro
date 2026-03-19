package models

import (
	"fmt"
	"sort"
	"strings"
	"sublink/database"
	"time"
)

type Webhook struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Name        string     `json:"name" gorm:"size:255;not null;default:''"`
	URL         string     `json:"url" gorm:"size:2048;not null;default:''"`
	Method      string     `json:"method" gorm:"size:16;not null;default:'POST'"`
	ContentType string     `json:"contentType" gorm:"size:128;not null;default:'application/json'"`
	Headers     string     `json:"headers" gorm:"type:text"`
	Body        string     `json:"body" gorm:"type:text"`
	Enabled     bool       `json:"enabled" gorm:"not null;default:false;index"`
	EventKeys   string     `json:"eventKeys" gorm:"type:text"`
	LastTestAt  *time.Time `json:"lastTestAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

func (w *Webhook) Add() error {
	if database.DB == nil {
		return fmt.Errorf("数据库未初始化")
	}
	return database.DB.Create(w).Error
}

func (w *Webhook) Update() error {
	if database.DB == nil {
		return fmt.Errorf("数据库未初始化")
	}
	if w.ID == 0 {
		return fmt.Errorf("Webhook ID 不能为空")
	}

	updates := map[string]interface{}{
		"name":         w.Name,
		"url":          w.URL,
		"method":       w.Method,
		"content_type": w.ContentType,
		"headers":      w.Headers,
		"body":         w.Body,
		"enabled":      w.Enabled,
		"event_keys":   w.EventKeys,
		"last_test_at": w.LastTestAt,
		"updated_at":   time.Now(),
	}

	return database.DB.Model(&Webhook{}).Where("id = ?", w.ID).Updates(updates).Error
}

func (w *Webhook) Delete() error {
	if database.DB == nil {
		return fmt.Errorf("数据库未初始化")
	}
	if w.ID == 0 {
		return fmt.Errorf("Webhook ID 不能为空")
	}
	return database.DB.Delete(&Webhook{}, w.ID).Error
}

func GetWebhookByID(id uint) (*Webhook, error) {
	if database.DB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	var webhook Webhook
	if err := database.DB.First(&webhook, id).Error; err != nil {
		return nil, err
	}
	return &webhook, nil
}

func ListWebhooks() ([]Webhook, error) {
	if database.DB == nil {
		return nil, fmt.Errorf("数据库未初始化")
	}

	var webhooks []Webhook
	if err := database.DB.Order("id ASC").Find(&webhooks).Error; err != nil {
		return nil, err
	}

	sort.SliceStable(webhooks, func(i, j int) bool {
		leftNamed := strings.TrimSpace(webhooks[i].Name) != ""
		rightNamed := strings.TrimSpace(webhooks[j].Name) != ""
		if leftNamed != rightNamed {
			return leftNamed
		}
		return webhooks[i].ID < webhooks[j].ID
	})

	return webhooks, nil
}
