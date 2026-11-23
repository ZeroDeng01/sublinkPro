package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"sublink/database"
	"sublink/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const realKaringWindowsUA = "Karing/1.2.23.2606 platform/windows;mihomo/1.19.28;clash-verge;FLClash"

func deviceRequest(method, path, hwid, ua string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequestWithContext(context.Background(), method, path, nil)
	c.Request.Header.Set("User-Agent", ua)
	if hwid != "" {
		c.Request.Header.Set("X-HWID", hwid)
	}
	GetClient(c)
	return recorder
}

func TestDeviceProtectedSubscriptionHTTP(t *testing.T) {
	setupClientsAPITestDB(t)
	createClientSubscriptionFixture(t, writeTestClashTemplate(t), writeTestSurgeTemplate(t), "device-http", "device-http-token", "private-node")
	share, err := models.GetSubscriptionShareByToken("device-http-token")
	if err != nil {
		t.Fatal(err)
	}
	limit := 1
	if err := share.UpdateWithDevicePolicy(&limit); err != nil {
		t.Fatal(err)
	}
	path := "/c/?token=device-http-token&client=v2ray"
	for _, tt := range []struct{ hwid, ua, code string }{{"", "GenericClient/1.0", "hwid_required"}, {"", realKaringWindowsUA, "hwid_required"}} {
		w := deviceRequest("GET", path, tt.hwid, tt.ua)
		if w.Code != 403 || !strings.Contains(w.Body.String(), tt.code) {
			t.Fatalf("denial: %d %s", w.Code, w.Body)
		}
		if w.Header().Get("subscription-userinfo") != "" {
			t.Fatal("denial leaked subscription metadata")
		}
	}
	w := deviceRequest("HEAD", path, "a", realKaringWindowsUA)
	if w.Code != 200 || w.Body.Len() != 0 {
		t.Fatalf("HEAD: %d %s", w.Code, w.Body)
	}
	devices, err := models.ListShareDevices(share.ID)
	if err != nil || len(devices) != 0 {
		t.Fatal("HEAD allocated device")
	}
	for range 2 {
		w = deviceRequest("GET", path, "a", realKaringWindowsUA)
		if w.Code != 200 || w.Body.Len() == 0 {
			t.Fatalf("valid import: %d %s", w.Code, w.Body)
		}
	}
	for _, client := range []string{"v2ray", "clash", "surge", "mihomo", "uri", "sing-box"} {
		w = deviceRequest("GET", "/c/?token=device-http-token&client="+client, "b", realKaringWindowsUA)
		if w.Code != 403 || !strings.Contains(w.Body.String(), "device_limit_exceeded") {
			t.Fatalf("format bypass %s: %d %s", client, w.Code, w.Body)
		}
	}
	// URL parameters are not a substitute for the request header.
	w = deviceRequest("GET", path+"&hwid=a&X-HWID=a", "", realKaringWindowsUA)
	if w.Code != 403 {
		t.Fatal("query parameter bypass")
	}
	if w.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatal("device responses must not be cached")
	}
}

func TestDeviceGenerationFailureDoesNotBind(t *testing.T) {
	setupClientsAPITestDB(t)
	createClientSubscriptionFixture(t, "missing-device-test-template.yaml", "", "device-failure", "device-failure-token", "private-node")
	share, err := models.GetSubscriptionShareByToken("device-failure-token")
	if err != nil {
		t.Fatal(err)
	}
	limit := 1
	if err := share.UpdateWithDevicePolicy(&limit); err != nil {
		t.Fatal(err)
	}
	w := deviceRequest("GET", "/c/?token=device-failure-token&client=clash", "a", realKaringWindowsUA)
	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected generation failure, got %d %s", w.Code, w.Body)
	}
	devices, err := models.ListShareDevices(share.ID)
	if err != nil || len(devices) != 0 {
		t.Fatal("failed generation bound a device")
	}
}

func TestDevicePolicyRecheckedBeforePublishing(t *testing.T) {
	setupClientsAPITestDB(t)
	createClientSubscriptionFixture(t, "", "", "device-race", "device-race-token", "private-node")
	share, err := models.GetSubscriptionShareByToken("device-race-token")
	if err != nil {
		t.Fatal(err)
	}
	limit := 1
	if err := share.UpdateWithDevicePolicy(&limit); err != nil {
		t.Fatal(err)
	}
	testGetClientAfterResolveSubscriptionNameHook = func(_ *gin.Context) {
		identity := models.ShareDeviceIdentity{UserAgent: realKaringWindowsUA, HWID: "other-device"}
		if err := models.CheckShareDevice(share.ID, share.Token, identity, true); err != nil {
			t.Error(err)
		}
	}
	w := deviceRequest("GET", "/c/?token=device-race-token&client=v2ray", "a", realKaringWindowsUA)
	if w.Code != 403 || !strings.Contains(w.Body.String(), "device_limit_exceeded") {
		t.Fatalf("race leaked nodes: %d %s", w.Code, w.Body)
	}
	if w.Header().Get("Content-Disposition") != "" {
		t.Fatal("race leaked subscription headers")
	}
}

func callShareJSON(t *testing.T, handler gin.HandlerFunc, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequestWithContext(t.Context(), "POST", path, strings.NewReader(string(encoded)))
	c.Request.Header.Set("Content-Type", "application/json")
	handler(c)
	return w
}

func TestShareDeviceAPIDefaultsAndManagement(t *testing.T) {
	setupClientsAPITestDB(t)
	sub := models.Subcription{Name: "device-defaults", DefaultMaxDevices: 2}
	if err := sub.Add(); err != nil {
		t.Fatal(err)
	}
	if err := models.CreateDefaultShareForSubscription(sub.ID); err != nil {
		t.Fatal(err)
	}
	w := callShareJSON(t, ShareAdd, "/api/v1/shares/add", gin.H{"subscription_id": sub.ID, "name": "customer"})
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	shares := models.GetSharesBySubscriptionID(sub.ID)
	for _, share := range shares {
		if share.MaxDevices != 2 {
			t.Fatal("share did not inherit defaults")
		}
	}
	w = callShareJSON(t, ShareBatchAdd, "/api/v1/shares/batch-add", gin.H{"subscription_id": sub.ID, "base_name": "batch", "count": 2, "enabled": true, "max_devices": 0})
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	shares = models.GetSharesBySubscriptionID(sub.ID)
	var share models.SubscriptionShare
	for _, s := range shares {
		if s.Name == "customer" {
			share = s
		}
		if strings.HasPrefix(s.Name, "batch") && s.MaxDevices != 0 {
			t.Fatal("explicit zero was ignored")
		}
	}
	identity := models.ShareDeviceIdentity{UserAgent: realKaringWindowsUA, HWID: "a"}
	if err := models.CheckShareDevice(share.ID, share.Token, identity, true); err != nil {
		t.Fatal(err)
	}
	devices, err := models.ListShareDevices(share.ID)
	if err != nil || len(devices) != 1 {
		t.Fatal("device missing")
	}
	path := "/api/v1/shares/device-update?shareId=" + url.QueryEscape(jsonNumber(share.ID))
	w = callShareJSON(t, ShareDeviceUpdate, path, gin.H{"id": devices[0].ID, "revoked": true, "name": "customer phone"})
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	w = deviceRequest("GET", "/c/?token="+share.Token, "a", realKaringWindowsUA)
	if w.Code != 403 {
		t.Fatal("revoked device passed")
	}
	// Limits remain independent of later subscription default changes.
	if err := database.DB.Model(&sub).Update("default_max_devices", 3).Error; err != nil {
		t.Fatal(err)
	}
	fresh, err := models.GetSubscriptionShareByToken(share.Token)
	if err != nil || fresh.MaxDevices != 2 {
		t.Fatal("default change modified existing share")
	}
}

func jsonNumber(value int) string { encoded, _ := json.Marshal(value); return string(encoded) }

func TestShareBatchDevicePolicyPartialFailure(t *testing.T) {
	setupClientsAPITestDB(t)
	shares := []*models.SubscriptionShare{
		{Token: "batch-full", Enabled: true, MaxDevices: 2},
		{Token: "batch-empty", Enabled: true, MaxDevices: 2},
	}
	for _, share := range shares {
		if err := share.Add(); err != nil {
			t.Fatal(err)
		}
	}
	for _, hwid := range []string{"device-a", "device-b"} {
		identity := models.ShareDeviceIdentity{UserAgent: realKaringWindowsUA, HWID: hwid}
		if err := models.CheckShareDevice(shares[0].ID, shares[0].Token, identity, true); err != nil {
			t.Fatal(err)
		}
	}
	w := callShareJSON(t, ShareBatchUpdate, "/api/v1/shares/batch-update", gin.H{
		"ids": []int{shares[0].ID, shares[1].ID}, "max_devices": 1,
	})
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "1/2") {
		t.Fatalf("expected partial failure, got %d %s", w.Code, w.Body)
	}
	var response struct {
		Code         int                    `json:"code"`
		Data         []int                  `json:"data"`
		Results      []ShareBatchItemResult `json:"results"`
		SuccessCount int                    `json:"success_count"`
		FailedCount  int                    `json:"failed_count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Code != 400 || response.SuccessCount != 1 || response.FailedCount != 1 || len(response.Results) != 2 {
		t.Fatalf("missing structured partial results: %s", w.Body)
	}
	if response.Results[0].ID != shares[0].ID || response.Results[0].Success || response.Results[0].Error == "" || response.Results[1].ID != shares[1].ID || !response.Results[1].Success {
		t.Fatalf("incorrect per-share results: %s", w.Body)
	}
	if len(response.Data) != 1 || response.Data[0] != shares[1].ID {
		t.Fatalf("incorrect successful IDs: %s", w.Body)
	}
	for i, share := range shares {
		if err := share.Find(); err != nil {
			t.Fatal(err)
		}
		if i == 0 && share.MaxDevices != 2 {
			t.Fatal("failed share update must preserve its entire device policy")
		}
		if i == 1 && share.MaxDevices != 1 {
			t.Fatal("successful share update must persist despite the batch error")
		}
	}
}

func TestDeviceProtectedSubscriptionRejectsDuplicateHWID(t *testing.T) {
	setupClientsAPITestDB(t)
	createClientSubscriptionFixture(t, "", "", "duplicate-hwid", "duplicate-hwid-token", "private-node")
	share, err := models.GetSubscriptionShareByToken("duplicate-hwid-token")
	if err != nil {
		t.Fatal(err)
	}
	limit := 1
	if err := share.UpdateWithDevicePolicy(&limit); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/c/?token="+share.Token, nil)
	c.Request.Header.Set("User-Agent", realKaringWindowsUA)
	c.Request.Header.Add("X-HWID", "a")
	c.Request.Header.Add("X-HWID", "b")
	GetClient(c)
	if w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), "hwid_invalid") {
		t.Fatalf("ambiguous HWID was not rejected: %d %s", w.Code, w.Body)
	}
	devices, err := models.ListShareDevices(share.ID)
	if err != nil || len(devices) != 0 {
		t.Fatal("ambiguous HWID allocated a device slot")
	}
}

func TestSubscriptionCreateDeviceLimitsAreOptIn(t *testing.T) {
	setupClientsAPITestDB(t)
	for _, tc := range []struct {
		name      string
		values    url.Values
		wantLimit int
	}{
		{"default-unrestricted", url.Values{}, 0},
		{"explicit-device-limit", url.Values{"DefaultMaxDevices": {"1"}}, 1},
		{"explicit-unlimited", url.Values{"DefaultMaxDevices": {"0"}}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.values.Set("name", tc.name)
			tc.values.Set("groups", "device-defaults")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/subcription/add", strings.NewReader(tc.values.Encode()))
			c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			SubAdd(c)
			var sub models.Subcription
			if err := database.DB.Where("name = ?", tc.name).First(&sub).Error; err != nil {
				t.Fatalf("create failed: %v response=%s", err, w.Body)
			}
			if sub.DefaultMaxDevices != tc.wantLimit {
				t.Fatalf("got default limit %d, want %d", sub.DefaultMaxDevices, tc.wantLimit)
			}
			shares := models.GetSharesBySubscriptionID(sub.ID)
			if len(shares) != 1 || shares[0].MaxDevices != tc.wantLimit {
				t.Fatal("default share must inherit explicitly chosen policy")
			}
		})
	}
}

func TestDeviceLimitAcceptsGenericHWIDClient(t *testing.T) {
	setupClientsAPITestDB(t)
	createClientSubscriptionFixture(t, "", "", "generic-hwid", "generic-hwid-token", "private-node")
	share, err := models.GetSubscriptionShareByToken("generic-hwid-token")
	if err != nil {
		t.Fatal(err)
	}
	limit := 1
	if err := share.UpdateWithDevicePolicy(&limit); err != nil {
		t.Fatal(err)
	}
	w := deviceRequest("GET", "/c/?token="+share.Token+"&client=v2ray", "generic-device", "GenericClient/1.0")
	if w.Code != http.StatusOK {
		t.Fatalf("device limit denied generic client: %d %s", w.Code, w.Body)
	}
}

func TestShareBatchUpdateInterleavesWithDeviceReset(t *testing.T) {
	setupClientsAPITestDB(t)
	expireAt := time.Now().Add(24 * time.Hour).Round(time.Second)
	shares := []*models.SubscriptionShare{
		{Token: "batch-reset-a", Name: "original-a", Enabled: true, MaxDevices: 2, ExpireType: models.ExpireTypeDateTime, ExpireAt: &expireAt},
		{Token: "batch-reset-b", Name: "original-b", Enabled: true, MaxDevices: 2, ExpireType: models.ExpireTypeDateTime, ExpireAt: &expireAt},
	}
	for _, share := range shares {
		if err := share.Add(); err != nil {
			t.Fatal(err)
		}
		if err := models.CheckShareDevice(share.ID, share.Token, models.ShareDeviceIdentity{HWID: "old-device"}, true); err != nil {
			t.Fatal(err)
		}
	}
	// Rotate tokens after batch validation has loaded its original snapshots,
	// before any requested field is applied. No timing-dependent sleeps needed.
	resetTokens := make(map[int]string)
	triggered := false
	callback := "test:reset_after_batch_load"
	if err := database.DB.Callback().Query().After("gorm:query").Register(callback, func(tx *gorm.DB) {
		loaded, ok := tx.Statement.Dest.(*models.SubscriptionShare)
		if triggered || !ok || loaded.ID != shares[1].ID {
			return
		}
		triggered = true
		for _, share := range shares {
			token, err := models.ResetShareDevices(share.ID)
			if err != nil {
				t.Fatal(err)
			}
			resetTokens[share.ID] = token
			current := &models.SubscriptionShare{ID: share.ID}
			if err := current.UpdateFields(map[string]any{"name": "edited-after-load", "enabled": false}); err != nil {
				t.Fatal(err)
			}
		}
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := database.DB.Callback().Query().Remove(callback); err != nil {
			t.Error(err)
		}
	})
	w := callShareJSON(t, ShareBatchUpdate, "/api/v1/shares/batch-update", gin.H{
		"ids": []int{shares[0].ID, shares[1].ID}, "max_devices": 1,
	})
	if !triggered || w.Code != http.StatusOK {
		t.Fatalf("interleaved batch failed: %d %s", w.Code, w.Body)
	}
	for _, share := range shares {
		if err := share.Find(); err != nil {
			t.Fatal(err)
		}
		if share.Token != resetTokens[share.ID] || share.Name != "edited-after-load" || share.Enabled || share.MaxDevices != 1 || share.ExpireAt == nil || !share.ExpireAt.Equal(expireAt) {
			t.Fatalf("batch overwrote fields absent from the request: %+v", share)
		}
		devices, err := models.ListShareDevices(share.ID)
		if err != nil || len(devices) != 1 || !devices[0].Revoked {
			t.Fatalf("batch undid device reset: %+v, %v", devices, err)
		}
	}
}

func TestShareUpdatePreservesOmittedFieldsAndExplicitZero(t *testing.T) {
	setupClientsAPITestDB(t)
	expireAt := time.Now().Add(24 * time.Hour).Round(time.Second)
	share := &models.SubscriptionShare{Token: "single-patch-token", Name: "keep-name", Enabled: true, MaxDevices: 2, ExpireType: models.ExpireTypeDateTime, ExpireAt: &expireAt}
	if err := share.Add(); err != nil {
		t.Fatal(err)
	}
	w := callShareJSON(t, ShareUpdate, "/api/v1/shares/update", gin.H{"id": share.ID, "max_devices": 0, "enabled": false})
	if w.Code != http.StatusOK {
		t.Fatalf("patch failed: %d %s", w.Code, w.Body)
	}
	if err := share.Find(); err != nil {
		t.Fatal(err)
	}
	if share.MaxDevices != 0 || share.Enabled || share.Token != "single-patch-token" || share.Name != "keep-name" || share.ExpireAt == nil || !share.ExpireAt.Equal(expireAt) {
		t.Fatalf("explicit zero/false or omitted fields were not preserved: %+v", share)
	}
	w = callShareJSON(t, ShareBatchUpdate, "/api/v1/shares/batch-update", gin.H{"ids": []int{share.ID}, "expire_type": models.ExpireTypeNever})
	if w.Code != http.StatusOK {
		t.Fatalf("expiry update failed: %d %s", w.Code, w.Body)
	}
	if err := share.Find(); err != nil {
		t.Fatal(err)
	}
	if share.ExpireType != models.ExpireTypeNever || share.ExpireAt != nil || share.MaxDevices != 0 || share.Enabled {
		t.Fatalf("expiry update failed to clear date or changed other fields: %+v", share)
	}
	revision := share.DeviceRevision
	w = callShareJSON(t, ShareBatchUpdate, "/api/v1/shares/batch-update", gin.H{"ids": []int{share.ID}})
	if w.Code != http.StatusOK {
		t.Fatalf("keep-original batch failed: %d %s", w.Code, w.Body)
	}
	if err := share.Find(); err != nil {
		t.Fatal(err)
	}
	if share.DeviceRevision != revision || share.MaxDevices != 0 || share.Enabled {
		t.Fatalf("batch without selected changes wrote the share: %+v", share)
	}
}
