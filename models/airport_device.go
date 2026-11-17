package models

import (
	"fmt"
	"strings"
	"sublink/utils"

	"gorm.io/gorm"
)

// AirportDeviceProfileAndroid 等常量定义可选的设备伪装档位。
// 值会持久化到 airports.device_profile，新增档位必须保持向后兼容。
const (
	AirportDeviceProfileNone    = ""
	AirportDeviceProfileAndroid = "android"
	AirportDeviceProfileIOS     = "ios"
	AirportDeviceProfileWindows = "windows"
	AirportDeviceProfileMacOS   = "macos"
	AirportDeviceProfileLinux   = "linux"
	// AirportDeviceProfileCustom 允许用户手填系统/型号/应用名称。
	AirportDeviceProfileCustom = "custom"
)

// airportDevicePreset 描述一个设备档位下客户端会发送的标识信息。
// 机场面板（例如依赖 x-device-os 的 Remnawave 系面板）据此判断请求来源设备。
type airportDevicePreset struct {
	UserAgent   string
	DeviceOS    string
	OSVersion   string
	DeviceModel string
	AppName     string
}

var airportDevicePresets = map[string]airportDevicePreset{
	AirportDeviceProfileAndroid: {
		UserAgent:   "v2rayTun/5.4.1 (Android 14; Pixel 7)",
		DeviceOS:    "Android",
		OSVersion:   "14",
		DeviceModel: "Pixel 7",
		AppName:     "v2rayTun",
	},
	AirportDeviceProfileIOS: {
		UserAgent:   "v2rayTun/5.4.1 (iOS 17.4; iPhone15,2)",
		DeviceOS:    "iOS",
		OSVersion:   "17.4",
		DeviceModel: "iPhone15,2",
		AppName:     "v2rayTun",
	},
	AirportDeviceProfileWindows: {
		UserAgent:   "v2rayN/6.31 (Windows 11)",
		DeviceOS:    "Windows",
		OSVersion:   "11",
		DeviceModel: "PC",
		AppName:     "v2rayN",
	},
	AirportDeviceProfileMacOS: {
		UserAgent:   "v2rayTun/5.4.1 (macOS 14.5)",
		DeviceOS:    "macOS",
		OSVersion:   "14.5",
		DeviceModel: "MacBookPro18,3",
		AppName:     "v2rayTun",
	},
	AirportDeviceProfileLinux: {
		UserAgent:   "v2rayN/6.31 (Linux)",
		DeviceOS:    "Linux",
		OSVersion:   "6.5",
		DeviceModel: "PC",
		AppName:     "v2rayN",
	},
}

// NormalizeAirportDeviceProfile 统一设备档位取值，未知档位按“不伪装”处理。
func NormalizeAirportDeviceProfile(profile string) string {
	normalized := strings.ToLower(strings.TrimSpace(profile))
	if normalized == AirportDeviceProfileCustom {
		return AirportDeviceProfileCustom
	}
	if _, ok := airportDevicePresets[normalized]; ok {
		return normalized
	}
	return AirportDeviceProfileNone
}

// EffectiveUserAgent 返回拉取机场订阅时实际使用的 User-Agent。
// 显式配置的 User-Agent 优先，其次才使用设备档位预置值。
func (a *Airport) EffectiveUserAgent() string {
	if a == nil {
		return ""
	}
	if userAgent := strings.TrimSpace(a.UserAgent); userAgent != "" {
		return userAgent
	}
	if preset, ok := airportDevicePresets[NormalizeAirportDeviceProfile(a.DeviceProfile)]; ok {
		return preset.UserAgent
	}
	return ""
}

// airportDevicePresetValues 返回预置档位对应的设备字段值。
// 仅用于新建/切换档位时填表，以及历史数据的一次性回填；运行时不再隐式使用。
func airportDevicePresetValues(profile string) (airportDevicePreset, bool) {
	preset, ok := airportDevicePresets[NormalizeAirportDeviceProfile(profile)]
	return preset, ok
}

// backfillAirportDevicePresetFields 为历史机场把预置档位值写入设备字段。
// 旧版本在运行时隐式使用预置值，改为显式字段后需要一次性落库，
// 否则这些机场会突然停止发送设备伪装头。已填写的字段保持不变。
func backfillAirportDevicePresetFields(db *gorm.DB) error {
	if db == nil {
		return nil
	}

	var airports []Airport
	if err := db.Select("ID", "DeviceProfile", "DeviceOS", "OSVersion", "DeviceModel", "AppName").Find(&airports).Error; err != nil {
		return fmt.Errorf("failed to query airport device fields: %w", err)
	}

	updated := 0
	for i := range airports {
		preset, ok := airportDevicePresetValues(airports[i].DeviceProfile)
		if !ok {
			continue
		}
		updates := make(map[string]any, 4)
		if strings.TrimSpace(airports[i].DeviceOS) == "" {
			updates["device_os"] = preset.DeviceOS
		}
		if strings.TrimSpace(airports[i].OSVersion) == "" {
			updates["os_version"] = preset.OSVersion
		}
		if strings.TrimSpace(airports[i].DeviceModel) == "" {
			updates["device_model"] = preset.DeviceModel
		}
		if strings.TrimSpace(airports[i].AppName) == "" {
			updates["app_name"] = preset.AppName
		}
		if len(updates) == 0 {
			continue
		}
		if err := db.Model(&Airport{}).Where("id = ?", airports[i].ID).Updates(updates).Error; err != nil {
			return fmt.Errorf("failed to backfill airport device fields (id=%d): %w", airports[i].ID, err)
		}
		updated++
	}
	if updated > 0 {
		utils.Info("Backfilled device masquerade fields for %d airports", updated)
	}
	return nil
}

// EffectiveRequestHeaders 在用户自定义请求头之上补充设备伪装与 X-HWID。
// 用户自己配置的同名请求头优先，避免覆盖手工调好的值。
func (a *Airport) EffectiveRequestHeaders() AirportRequestHeaders {
	if a == nil {
		return nil
	}
	headers := make(AirportRequestHeaders, 0, len(a.RequestHeaders)+5)
	seen := make(map[string]bool, len(a.RequestHeaders)+5)
	for _, header := range a.RequestHeaders {
		key := strings.TrimSpace(header.Key)
		if key == "" {
			continue
		}
		seen[strings.ToLower(key)] = true
		headers = append(headers, header)
	}
	appendHeader := func(key, value string) {
		if value == "" || seen[strings.ToLower(key)] {
			return
		}
		headers = append(headers, AirportRequestHeader{Key: key, Value: value})
		seen[strings.ToLower(key)] = true
	}

	profile := NormalizeAirportDeviceProfile(a.DeviceProfile)
	if profile != AirportDeviceProfileNone {
		// 设备字段是运行时唯一的取值来源：档位预置只用于新建/切换档位时填表，
		// 留空表示不发送该头，不会再用预置值回退。
		appendHeader("X-Device-OS", strings.TrimSpace(a.DeviceOS))
		appendHeader("X-Ver-OS", strings.TrimSpace(a.OSVersion))
		appendHeader("X-Device-Model", strings.TrimSpace(a.DeviceModel))
		appendHeader("X-App-Name", strings.TrimSpace(a.AppName))
	}

	if !a.DisableHWID {
		hwid := strings.TrimSpace(a.HWID)
		if hwid == "" {
			// 历史数据没有保存 HWID 时按机场 ID 派生，保证同一机场始终是同一台“设备”。
			hwid = utils.DeriveHWID(a.ID)
		}
		appendHeader("X-HWID", hwid)
	}

	return headers
}
