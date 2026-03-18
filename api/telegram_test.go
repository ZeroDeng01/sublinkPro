package api

import (
	"encoding/json"
	"net/http"
	"sublink/services/notifications"
	"sublink/services/telegram"
	"testing"
)

func TestGetTelegramConfigReturnsDefaultsAndEventOptions(t *testing.T) {
	setupSettingAPITestDB(t)

	recorder := performJSONRequest(t, GetTelegramConfig, http.MethodGet, nil)
	response := decodeAPIResponse(t, recorder)

	if response.Code != 200 {
		t.Fatalf("expected response code 200, got %d", response.Code)
	}

	var data struct {
		Enabled      bool                            `json:"enabled"`
		EventKeys    []string                        `json:"eventKeys"`
		EventOptions []notifications.EventDefinition `json:"eventOptions"`
		Connected    bool                            `json:"connected"`
	}
	if err := json.Unmarshal(response.Data, &data); err != nil {
		t.Fatalf("unmarshal telegram config data: %v", err)
	}

	if data.Enabled {
		t.Fatalf("expected telegram to be disabled by default")
	}
	if data.Connected {
		t.Fatalf("expected telegram to be disconnected by default")
	}
	if len(data.EventKeys) == 0 {
		t.Fatalf("expected default telegram event keys")
	}
	if len(data.EventOptions) == 0 {
		t.Fatalf("expected telegram event options")
	}
}

func TestUpdateTelegramConfigPersistsSettingsAndEvents(t *testing.T) {
	setupSettingAPITestDB(t)

	recorder := performJSONRequest(t, UpdateTelegramConfig, http.MethodPost, map[string]interface{}{
		"enabled":   false,
		"botToken":  "test-bot-token",
		"chatId":    123456789,
		"useProxy":  true,
		"proxyLink": "socks5://127.0.0.1:1080",
		"eventKeys": []string{
			"security.user_login",
			"subscription.sync_failed",
		},
	})
	response := decodeAPIResponse(t, recorder)

	if response.Code != 200 {
		t.Fatalf("expected response code 200, got %d", response.Code)
	}

	config, err := telegram.LoadConfig()
	if err != nil {
		t.Fatalf("load telegram config: %v", err)
	}

	if config.Enabled {
		t.Fatalf("expected telegram to remain disabled")
	}
	if config.BotToken != "test-bot-token" {
		t.Fatalf("unexpected bot token: %s", config.BotToken)
	}
	if config.ChatID != 123456789 {
		t.Fatalf("unexpected chat id: %d", config.ChatID)
	}
	if !config.UseProxy {
		t.Fatalf("expected proxy to be enabled")
	}
	if config.ProxyLink != "socks5://127.0.0.1:1080" {
		t.Fatalf("unexpected proxy link: %s", config.ProxyLink)
	}
	if len(config.EventKeys) != 2 {
		t.Fatalf("expected 2 selected telegram events, got %d", len(config.EventKeys))
	}
}
