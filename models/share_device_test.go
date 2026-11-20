package models

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"sublink/database"
)

const testKaringUA = "Karing/1.2.23.2606 platform/windows;mihomo/1.19.28"

func newDeviceShare(t *testing.T, limit int) *SubscriptionShare {
	t.Helper()
	share := &SubscriptionShare{Token: fmt.Sprintf("device-test-%d", time.Now().UnixNano()), Enabled: true, KaringOnly: true, MaxDevices: limit}
	if err := share.Add(); err != nil {
		t.Fatal(err)
	}
	return share
}

func deviceIdentity(id string) ShareDeviceIdentity {
	return ShareDeviceIdentity{UserAgent: testKaringUA, HWID: id, OS: "windows", Model: "test"}
}

func requireDeviceError(t *testing.T, err error, code string) {
	t.Helper()
	var denied *ShareAccessError
	if !errors.As(err, &denied) || denied.Code != code {
		t.Fatalf("want %s, got %v", code, err)
	}
}

func TestShareDeviceAdmissionLifecycle(t *testing.T) {
	setupSubscriptionShareTestDB(t)
	share := newDeviceShare(t, 1)
	check := func(id string, bind bool) error {
		return CheckShareDevice(share.ID, share.Token, deviceIdentity(id), bind)
	}
	if err := check("device-a", false); err != nil {
		t.Fatal(err)
	}
	devices, err := ListShareDevices(share.ID)
	if err != nil || len(devices) != 0 {
		t.Fatalf("preflight allocated a slot: %v %v", devices, err)
	}
	for range 3 {
		if err := check("device-a", true); err != nil {
			t.Fatal(err)
		}
	}
	requireDeviceError(t, check("device-b", true), "device_limit_exceeded")
	devices, err = ListShareDevices(share.ID)
	if err != nil || len(devices) != 1 {
		t.Fatalf("expected one device: %v", err)
	}
	if devices[0].HWIDHash == "device-a" || len(devices[0].HWIDHash) != 64 {
		t.Fatal("HWID must be hashed")
	}
	revoked := true
	if err := UpdateShareDevice(share.ID, devices[0].ID, &revoked, nil); err != nil {
		t.Fatal(err)
	}
	requireDeviceError(t, check("device-a", true), "device_revoked")
	if err := check("device-b", true); err != nil {
		t.Fatal(err)
	}
	revoked = false
	requireDeviceError(t, UpdateShareDevice(share.ID, devices[0].ID, &revoked, nil), "device_limit_exceeded")
	two := 2
	if err := share.UpdateWithDevicePolicy(nil, &two); err != nil {
		t.Fatal(err)
	}
	if err := UpdateShareDevice(share.ID, devices[0].ID, &revoked, nil); err != nil {
		t.Fatal(err)
	}
	one := 1
	requireDeviceError(t, share.UpdateWithDevicePolicy(nil, &one), "device_limit_below_bound")
	oldToken := share.Token
	newToken, err := ResetShareDevices(share.ID)
	if err != nil {
		t.Fatal(err)
	}
	listed := GetSharesBySubscriptionID(share.SubscriptionID)
	if len(listed) != 1 || listed[0].Token != newToken {
		t.Fatal("reset must keep the share visible with its new token")
	}
	requireDeviceError(t, CheckShareDevice(share.ID, oldToken, deviceIdentity("device-c"), true), "share_unavailable")
	requireDeviceError(t, CheckShareDevice(share.ID, newToken, deviceIdentity("device-a"), true), "device_revoked")
	if err := CheckShareDevice(share.ID, newToken, deviceIdentity("device-c"), true); err != nil {
		t.Fatal(err)
	}
	// A fresh cache cannot forget device records.
	resetSubscriptionShareCacheForTest()
	if err := CheckShareDevice(share.ID, newToken, deviceIdentity("device-c"), true); err != nil {
		t.Fatal(err)
	}
	if err := share.Delete(); err != nil {
		t.Fatal(err)
	}
	devices, err = ListShareDevices(share.ID)
	if err != nil || len(devices) != 0 {
		t.Fatal("delete left device records")
	}
}

func TestShareDeviceValidationAndIndependentPolicies(t *testing.T) {
	setupSubscriptionShareTestDB(t)
	share := newDeviceShare(t, 1)
	for _, ua := range []string{"", "curl Karing/1.2.23.2606 platform/windows", "Karing", "notKaring/1.2.3 platform/windows"} {
		identity := deviceIdentity("a")
		identity.UserAgent = ua
		requireDeviceError(t, CheckShareDevice(share.ID, share.Token, identity, true), "karing_required")
	}
	for _, id := range []string{"", " a", "a\n", string(make([]byte, 513))} {
		err := CheckShareDevice(share.ID, share.Token, deviceIdentity(id), true)
		if err == nil {
			t.Fatal("accepted invalid HWID")
		}
	}
	if err := CheckShareDevice(share.ID, share.Token, deviceIdentity("a"), true); err != nil {
		t.Fatal(err)
	}
	other := newDeviceShare(t, 1)
	if err := CheckShareDevice(other.ID, other.Token, deviceIdentity("b"), true); err != nil {
		t.Fatal(err)
	}
	unlimited := 0
	if err := share.UpdateWithDevicePolicy(nil, &unlimited); err != nil {
		t.Fatal(err)
	}
	if err := CheckShareDevice(share.ID, share.Token, deviceIdentity(""), true); err != nil {
		t.Fatal(err)
	}
	noKaring := false
	one := 1
	if err := other.UpdateWithDevicePolicy(&noKaring, &one); err != nil {
		t.Fatal(err)
	}
	identity := deviceIdentity("b")
	identity.UserAgent = "OtherClient"
	if err := CheckShareDevice(other.ID, other.Token, identity, true); err != nil {
		t.Fatal(err)
	}
	share.Enabled = false
	if err := share.Update(); err != nil {
		t.Fatal(err)
	}
	requireDeviceError(t, CheckShareDevice(share.ID, share.Token, deviceIdentity("a"), true), "share_unavailable")
}

func TestShareDeviceConcurrentAdmission(t *testing.T) {
	setupSubscriptionShareTestDB(t)
	share := newDeviceShare(t, 1)
	var successes atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := range 24 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			if CheckShareDevice(share.ID, share.Token, deviceIdentity(fmt.Sprintf("device-%d", i)), true) == nil {
				successes.Add(1)
			}
		}(i)
	}
	close(start)
	wg.Wait()
	if successes.Load() != 1 {
		t.Fatalf("expected one admitted device, got %d", successes.Load())
	}
	devices, err := ListShareDevices(share.ID)
	if err != nil || len(devices) != 1 {
		t.Fatalf("invalid device count: %d %v", len(devices), err)
	}
}

func TestShareDeviceTokenRefreshPreservesBindingAndPolicies(t *testing.T) {
	setupSubscriptionShareTestDB(t)
	share := newDeviceShare(t, 1)
	if err := CheckShareDevice(share.ID, share.Token, deviceIdentity("a"), true); err != nil {
		t.Fatal(err)
	}
	oldToken := share.Token
	share.Token = "rotated-test-token"
	// Legacy update callers omit the new fields and must not erase them.
	share.MaxDevices = 0
	share.KaringOnly = false
	if err := share.Update(); err != nil {
		t.Fatal(err)
	}
	fresh, err := GetSubscriptionShareByToken(share.Token)
	if err != nil {
		t.Fatal(err)
	}
	if fresh.MaxDevices != 1 || !fresh.KaringOnly {
		t.Fatal("ordinary update erased policy")
	}
	requireDeviceError(t, CheckShareDevice(share.ID, oldToken, deviceIdentity("a"), true), "share_unavailable")
	if err := CheckShareDevice(share.ID, share.Token, deviceIdentity("a"), true); err != nil {
		t.Fatal(err)
	}
	requireDeviceError(t, CheckShareDevice(share.ID, share.Token, deviceIdentity("b"), true), "device_limit_exceeded")
	shares := []SubscriptionShare{*fresh}
	if err := PopulateShareDeviceCounts(shares); err != nil {
		t.Fatal(err)
	}
	if shares[0].DeviceCount != 1 {
		t.Fatal("incorrect aggregate count")
	}
	if err := database.DB.Model(fresh).Updates(map[string]any{"expire_type": ExpireTypeDateTime, "expire_at": time.Now().Add(-time.Second)}).Error; err != nil {
		t.Fatal(err)
	}
	requireDeviceError(t, CheckShareDevice(share.ID, share.Token, deviceIdentity("a"), true), "share_unavailable")
}
