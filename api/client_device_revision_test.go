package api

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sublink/database"
	"sublink/models"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

func TestUnlimitedSubscriptionDoesNotWriteDeviceRevision(t *testing.T) {
	setupClientsAPITestDB(t)
	createClientSubscriptionFixture(t, writeTestClashTemplate(t), "", "unlimited", "unlimited-token", "private-node")
	share, err := models.GetSubscriptionShareByToken("unlimited-token")
	if err != nil {
		t.Fatal(err)
	}
	for _, client := range []string{"v2ray", "clash"} {
		w := deviceRequest(http.MethodGet, "/c/?token=unlimited-token&client="+client, "", "Happ")
		if w.Code != http.StatusOK || w.Body.Len() == 0 {
			t.Fatalf("ordinary download: %d %s", w.Code, w.Body)
		}
	}
	models.WaitForPendingAccessRecords()
	var fresh models.SubscriptionShare
	if err := database.DB.First(&fresh, share.ID).Error; err != nil {
		t.Fatal(err)
	}
	if fresh.DeviceRevision != 0 {
		t.Fatalf("ordinary downloads wrote revision %d", fresh.DeviceRevision)
	}
	devices, err := models.ListShareDevices(share.ID)
	if err != nil || len(devices) != 0 {
		t.Fatalf("ordinary downloads bound devices: %v %v", devices, err)
	}
}

func TestUnlimitedShareWriterStreamsBeforeCompletion(t *testing.T) {
	setupClientsAPITestDB(t)
	share := models.SubscriptionShare{Token: "stream-token", Enabled: true}
	if err := share.Add(); err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/c/", nil)
	c.Set("subscriptionContentReady", true)
	writer := &unlimitedShareWriter{ResponseWriter: c.Writer, context: c, shareID: share.ID, token: share.Token, header: make(http.Header), status: http.StatusOK}
	c.Writer = writer
	if _, err := writer.WriteString("first"); err != nil {
		t.Fatal(err)
	}
	if w.Body.String() != "first" {
		t.Fatal("ordinary response was buffered until completion")
	}
	if _, err := writer.WriteString("second"); err != nil || w.Body.String() != "firstsecond" {
		t.Fatalf("stream continuation: %v %s", err, w.Body)
	}
}

func TestUnlimitedShareRechecksStateBeforePublishing(t *testing.T) {
	for _, change := range []string{"token", "disabled", "limited"} {
		t.Run(change, func(t *testing.T) {
			setupClientsAPITestDB(t)
			createClientSubscriptionFixture(t, "", "", "state-check", "state-check-token", "private-node")
			share, err := models.GetSubscriptionShareByToken("state-check-token")
			if err != nil {
				t.Fatal(err)
			}
			testGetClientAfterResolveSubscriptionNameHook = func(_ *gin.Context) {
				updates := map[string]any{"enabled": false}
				switch change {
				case "token":
					updates = map[string]any{"token": "new-token"}
				case "limited":
					updates = map[string]any{"max_devices": 1}
				}
				if err := share.UpdateFields(updates); err != nil {
					t.Error(err)
				}
			}
			w := deviceRequest(http.MethodGet, "/c/?token=state-check-token&client=v2ray", "", "Happ")
			if w.Code != http.StatusForbidden || w.Header().Get("Content-Disposition") != "" || w.Header().Get("subscription-userinfo") != "" {
				t.Fatalf("state change leaked content or metadata: %d %v %s", w.Code, w.Header(), w.Body)
			}
		})
	}
}

func TestRepeatedAndProxyMergedDeviceHeadersAreRejected(t *testing.T) {
	for _, values := range [][]string{{"a", "b"}, {"a,b"}, {"a, b"}, {"a,a"}, {"a,"}} {
		t.Run(strings.Join(values, "/"), func(t *testing.T) {
			setupClientsAPITestDB(t)
			createClientSubscriptionFixture(t, "", "", "duplicate", "duplicate-token", "private-node")
			share, err := models.GetSubscriptionShareByToken("duplicate-token")
			if err != nil {
				t.Fatal(err)
			}
			limit := 1
			if err := share.UpdateWithDevicePolicy(&limit); err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/c/?token=duplicate-token&client=v2ray", nil)
			for _, value := range values {
				c.Request.Header.Add("X-HWID", value)
			}
			GetClient(c)
			if w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), "hwid_invalid") {
				t.Fatalf("ambiguous HWID accepted: %d %s", w.Code, w.Body)
			}
			devices, err := models.ListShareDevices(share.ID)
			if err != nil || len(devices) != 0 {
				t.Fatalf("ambiguous HWID consumed slot: %v %v", devices, err)
			}
		})
	}
}

// These are HTTP contract regressions, not real-client platform validation.
func TestGenericDeviceClientOutputAndRefreshLifecycle(t *testing.T) {
	for _, ua := range []string{realKaringWindowsUA, "Happ"} {
		for _, format := range []string{"clash", "v2ray"} {
			t.Run(fmt.Sprintf("%s/%s", ua, format), func(t *testing.T) {
				setupClientsAPITestDB(t)
				createClientSubscriptionFixture(t, writeTestClashTemplate(t), "", "lifecycle", "lifecycle-token", "private-node")
				share, err := models.GetSubscriptionShareByToken("lifecycle-token")
				if err != nil {
					t.Fatal(err)
				}
				limit := 1
				if err := share.UpdateWithDevicePolicy(&limit); err != nil {
					t.Fatal(err)
				}
				for range 2 {
					w := deviceRequest(http.MethodGet, "/c/?token=lifecycle-token&client="+format, "single-device", ua)
					if w.Code != http.StatusOK {
						t.Fatalf("import/refresh: %d %s", w.Code, w.Body)
					}
					if format == "clash" {
						var config map[string]any
						if err := yaml.Unmarshal(w.Body.Bytes(), &config); err != nil || config["proxies"] == nil {
							t.Fatalf("invalid Clash YAML: %v", err)
						}
					} else {
						decoded, err := base64.StdEncoding.DecodeString(w.Body.String())
						if err != nil || !strings.Contains(string(decoded), "ss://") {
							t.Fatalf("invalid Base64 subscription: %v", err)
						}
					}
				}
				devices, err := models.ListShareDevices(share.ID)
				if err != nil || len(devices) != 1 {
					t.Fatalf("refresh consumed another slot: %v %v", devices, err)
				}
				revoked := true
				if err := models.UpdateShareDevice(share.ID, devices[0].ID, &revoked, nil); err != nil {
					t.Fatal(err)
				}
				w := deviceRequest(http.MethodGet, "/c/?token=lifecycle-token&client="+format, "single-device", ua)
				if w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), "device_revoked") {
					t.Fatalf("revoked refresh: %d %s", w.Code, w.Body)
				}
			})
		}
	}
}
