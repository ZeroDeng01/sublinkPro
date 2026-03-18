package telegram

import (
	"reflect"
	"sublink/database"
	"sublink/models"
	"sublink/services/notifications"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTelegramTestDB(t *testing.T) {
	t.Helper()

	oldDB := database.DB
	oldDialect := database.Dialect
	oldInitialized := database.IsInitialized

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.SystemSetting{}); err != nil {
		t.Fatalf("auto migrate system_settings: %v", err)
	}

	database.DB = db
	database.Dialect = database.DialectSQLite
	database.IsInitialized = false
	if err := models.InitSettingCache(); err != nil {
		t.Fatalf("init setting cache: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Exec("DELETE FROM system_settings").Error
		database.DB = db
		database.Dialect = database.DialectSQLite
		database.IsInitialized = false
		_ = models.InitSettingCache()

		database.DB = oldDB
		database.Dialect = oldDialect
		database.IsInitialized = oldInitialized
		if oldDB != nil {
			_ = models.InitSettingCache()
		}
	})
}

func TestLoadConfigReturnsDefaults(t *testing.T) {
	setupTelegramTestDB(t)

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("load telegram config: %v", err)
	}

	if config.Enabled {
		t.Fatalf("expected telegram to be disabled by default")
	}
	if config.BotToken != "" {
		t.Fatalf("expected empty bot token by default, got %s", config.BotToken)
	}
	if config.ChatID != 0 {
		t.Fatalf("expected empty chat id by default, got %d", config.ChatID)
	}
	if config.UseProxy {
		t.Fatalf("expected proxy to be disabled by default")
	}

	want := notifications.DefaultEventKeys(notifications.ChannelTelegram)
	if !reflect.DeepEqual(config.EventKeys, want) {
		t.Fatalf("telegram default event keys = %#v, want %#v", config.EventKeys, want)
	}
}

func TestSaveConfigRoundTripPersistsFieldsAndEvents(t *testing.T) {
	setupTelegramTestDB(t)

	err := SaveConfig(&Config{
		Enabled:   true,
		BotToken:  "bot-token",
		ChatID:    123456,
		UseProxy:  true,
		ProxyLink: "socks5://127.0.0.1:1080",
		EventKeys: []string{
			"security.user_login",
			"subscription.sync_failed",
		},
	})
	if err != nil {
		t.Fatalf("save telegram config: %v", err)
	}

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("reload telegram config: %v", err)
	}

	if !config.Enabled {
		t.Fatalf("expected telegram to be enabled")
	}
	if config.BotToken != "bot-token" {
		t.Fatalf("unexpected bot token: %s", config.BotToken)
	}
	if config.ChatID != 123456 {
		t.Fatalf("unexpected chat id: %d", config.ChatID)
	}
	if !config.UseProxy {
		t.Fatalf("expected proxy to be enabled")
	}
	if config.ProxyLink != "socks5://127.0.0.1:1080" {
		t.Fatalf("unexpected proxy link: %s", config.ProxyLink)
	}

	wantEventKeys := []string{
		"subscription.sync_failed",
		"security.user_login",
	}
	if !reflect.DeepEqual(config.EventKeys, wantEventKeys) {
		t.Fatalf("telegram event keys = %#v, want %#v", config.EventKeys, wantEventKeys)
	}
}

func TestGetStatusReturnsDefaultWhenBotNotStarted(t *testing.T) {
	botMutex.Lock()
	oldBot := globalBot
	globalBot = nil
	botMutex.Unlock()
	defer func() {
		botMutex.Lock()
		globalBot = oldBot
		botMutex.Unlock()
	}()

	status := GetStatus()

	if status["enabled"] != false {
		t.Fatalf("expected enabled=false, got %#v", status["enabled"])
	}
	if status["connected"] != false {
		t.Fatalf("expected connected=false, got %#v", status["connected"])
	}
	if status["botUsername"] != "" {
		t.Fatalf("expected empty bot username, got %#v", status["botUsername"])
	}
}

func TestGetStatusReturnsBotSnapshot(t *testing.T) {
	bot := &TelegramBot{
		connected:   true,
		lastError:   "",
		botUsername: "sublink_bot",
		botID:       42,
	}

	botMutex.Lock()
	oldBot := globalBot
	globalBot = bot
	botMutex.Unlock()
	defer func() {
		botMutex.Lock()
		globalBot = oldBot
		botMutex.Unlock()
	}()

	status := GetStatus()

	if status["enabled"] != true {
		t.Fatalf("expected enabled=true, got %#v", status["enabled"])
	}
	if status["connected"] != true {
		t.Fatalf("expected connected=true, got %#v", status["connected"])
	}
	if status["botUsername"] != "sublink_bot" {
		t.Fatalf("expected bot username sublink_bot, got %#v", status["botUsername"])
	}
	if status["botId"] != int64(42) {
		t.Fatalf("expected bot id 42, got %#v", status["botId"])
	}
}
