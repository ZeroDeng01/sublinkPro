package models

import (
	"reflect"
	"strconv"
	"testing"
	"time"

	"sublink/database"

	countryRegexp "github.com/dlclark/regexp2/v2/compat"
)

func TestSmartGroupDynamicMembersAcrossOriginalGroups(t *testing.T) {
	setupSubcriptionCopyTestDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	old := time.Now().Add(-96 * time.Hour).Format("2006-01-02 15:04:05")
	healthy := func(name, group, country, tested string) Node {
		return createSubcriptionTestNode(t, Node{Name: name, LinkName: name, Group: group, LinkCountry: country,
			DelayStatus: "success", SpeedStatus: "success", DelayTime: 120, Speed: 4,
			LatencyCheckAt: tested, SpeedCheckAt: tested})
	}
	gb := healthy("gb-a", "original-a", "GB", now)
	de := healthy("de-b", "original-b", "DE", now)
	healthy("fr-expired", "original-c", "FR", old)
	healthy("us-other", "original-d", "US", now)
	createSubcriptionTestNode(t, Node{Name: "gb-failed", Group: "original-e", LinkCountry: "GB", DelayStatus: "timeout"})
	group := SmartGroup{Name: "Europe", Countries: "gb, fr, DE,GB", MaxDelay: 500, MinSpeed: 4, MaxAgeHours: 72}
	if err := group.Validate(); err != nil {
		t.Fatal(err)
	}
	if group.Countries != "GB,FR,DE" {
		t.Fatalf("normalized countries = %q", group.Countries)
	}
	if err := database.DB.Create(&group).Error; err != nil {
		t.Fatal(err)
	}
	candidates, err := group.CandidateIDs()
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 4 {
		t.Fatalf("candidate IDs = %v, want four regardless of test status", candidates)
	}
	members, err := group.Members()
	if err != nil {
		t.Fatal(err)
	}
	if len(members) != 2 || members[0].ID != gb.ID || members[1].ID != de.ID {
		t.Fatalf("members = %v, want GB and DE nodes", nodeNames(members))
	}
	if members[0].Group != "original-a" || members[1].Group != "original-b" {
		t.Fatalf("source groups changed: %q, %q", members[0].Group, members[1].Group)
	}
	sub := &Subcription{Name: "smart-only", SmartGroupIDs: strconv.Itoa(group.ID), RefreshUsageOnRequest: true}
	if err := sub.Add(); err != nil {
		t.Fatal(err)
	}
	if err := sub.GetSub("clash"); err != nil {
		t.Fatal(err)
	}
	if got := nodeNames(sub.Nodes); !reflect.DeepEqual(got, []string{"gb-a", "de-b"}) {
		t.Fatalf("smart-only subscription nodes = %v", got)
	}
}

func TestNormalizeSmartGroupIDsRejectsPartialInvalidInput(t *testing.T) {
	setupSubcriptionCopyTestDB(t)
	group := SmartGroup{Name: "valid", Countries: "GB"}
	if err := database.DB.Create(&group).Error; err != nil {
		t.Fatal(err)
	}
	id := strconv.Itoa(group.ID)
	for _, input := range []string{id + ",abc", "," + id, id + ",-1", id + ","} {
		if _, err := NormalizeSmartGroupIDs(input); err == nil {
			t.Errorf("expected invalid %q to be rejected", input)
		}
	}
	if got, err := NormalizeSmartGroupIDs(id + "," + id); err != nil || got != id {
		t.Errorf("deduplication = %q, %v", got, err)
	}
}

func TestSmartGroupKeywordSourceGroupsAndLatencyOnly(t *testing.T) {
	setupSubcriptionCopyTestDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	matched := createSubcriptionTestNode(t, Node{Name: "PH-Manila-A", LinkName: "PH-Manila-A", Group: "airport-a", LinkCountry: "PH",
		DelayStatus: "success", SpeedStatus: "untested", DelayTime: 240, Speed: 0, LatencyCheckAt: now})
	createSubcriptionTestNode(t, Node{Name: "PH-Manila-B", Group: "airport-b", LinkCountry: "PH", DelayStatus: "success", DelayTime: 120, LatencyCheckAt: now})
	cebu := createSubcriptionTestNode(t, Node{Name: "PH-Cebu", Group: "airport-a", LinkCountry: "PH", DelayStatus: "success", DelayTime: 90, LatencyCheckAt: now})
	createSubcriptionTestNode(t, Node{Name: "PH-Manila-failed", Group: "airport-a", LinkCountry: "PH", DelayStatus: "timeout"})
	gb := createSubcriptionTestNode(t, Node{Name: "GB-Manila", Group: "airport-a", LinkCountry: "GB", DelayStatus: "success", DelayTime: 110, LatencyCheckAt: now})

	group := SmartGroup{Name: "Philippines", Countries: "ph", Keyword: "  manila  ", SourceGroups: []string{" airport-a ", "AIRPORT-A"}, MaxAgeHours: 72}
	if err := group.Validate(); err != nil {
		t.Fatal(err)
	}
	if group.Keyword != "manila" || !reflect.DeepEqual(group.SourceGroups, []string{"airport-a"}) {
		t.Fatalf("normalized conditions: %+v", group)
	}
	ids, err := group.CandidateIDs()
	if err != nil || len(ids) != 4 {
		t.Fatalf("country OR keyword candidates = %v, %v", ids, err)
	}
	if group.MatchSource(gb) != "keyword" || group.MatchSource(cebu) != "landingCountry" {
		t.Fatalf("match sources: GB=%q Cebu=%q", group.MatchSource(gb), group.MatchSource(cebu))
	}
	members, err := group.Members()
	if err != nil || len(members) != 3 || members[0].ID != matched.ID || members[1].ID != cebu.ID || members[2].ID != gb.ID {
		t.Fatalf("latency-only members = %v, %v", nodeNames(members), err)
	}
	if err := database.DB.Create(&group).Error; err != nil {
		t.Fatal(err)
	}
	sub := &Subcription{Name: "ph-only", SmartGroupIDs: strconv.Itoa(group.ID), RefreshUsageOnRequest: true}
	if err := sub.Add(); err != nil {
		t.Fatal(err)
	}
	if err := sub.GetSub("clash"); err != nil || len(sub.Nodes) != 3 || sub.Nodes[0].ID != matched.ID || sub.Nodes[1].ID != cebu.ID || sub.Nodes[2].ID != gb.ID {
		t.Fatalf("filtered subscription nodes = %v, %v", nodeNames(sub.Nodes), err)
	}
	group.MinSpeed = 1
	members, err = group.Members()
	if err != nil || len(members) != 0 {
		t.Fatalf("speed-required members = %v, %v", nodeNames(members), err)
	}
	group.MinSpeed = 0
	group.MaxDelay = 200
	members, err = group.Members()
	if err != nil || len(members) != 2 || members[0].ID != cebu.ID || members[1].ID != gb.ID {
		t.Fatalf("latency-limit members = %v, %v", nodeNames(members), err)
	}
}

func TestSmartGroupLatencyFreshnessWithoutSpeed(t *testing.T) {
	setupSubcriptionCopyTestDB(t)
	old := time.Now().Add(-96 * time.Hour).Format("2006-01-02 15:04:05")
	createSubcriptionTestNode(t, Node{Name: "old-tcp", LinkCountry: "PH", DelayStatus: "success", DelayTime: 100, LatencyCheckAt: old})
	group := SmartGroup{Name: "recent", Countries: "PH", MaxAgeHours: 72}
	members, err := group.Members()
	if err != nil || len(members) != 0 {
		t.Fatalf("stale latency members = %v, %v", nodeNames(members), err)
	}
	group.MaxAgeHours = 0
	members, err = group.Members()
	if err != nil || len(members) != 1 {
		t.Fatalf("unlimited freshness members = %v, %v", nodeNames(members), err)
	}
}

func TestSmartGroupNameCountryFallbackAndExclusionCounts(t *testing.T) {
	setupSubcriptionCopyTestDB(t)
	oldRules := countryRuleCache
	t.Cleanup(func() { countryRuleCache = oldRules })
	if err := database.DB.AutoMigrate(&CountryRule{}); err != nil {
		t.Fatal(err)
	}
	if err := database.DB.Create(&CountryRule{ID: 910001, CountryCode: "PH", CountryName: "菲律宾", Pattern: "(?i)菲律宾|Philippines|🇵🇭", Enabled: true}).Error; err != nil {
		t.Fatal(err)
	}
	resetCountryRuleCacheForTest()
	if err := InitCountryRuleCache(); err != nil {
		t.Fatal(err)
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	good := createSubcriptionTestNode(t, Node{Name: "Manila 1", LinkName: "🇵🇭 Manila", Group: "source-a", DelayStatus: "success", DelayTime: 100, LatencyCheckAt: now})
	createSubcriptionTestNode(t, Node{Name: "Manila 2", LinkName: "Philippines slow", Group: "source-b", DelayStatus: "timeout"})
	mislabeled := createSubcriptionTestNode(t, Node{Name: "Manila 3", LinkName: "🇵🇭 mislabeled", LinkCountry: "GB", DelayStatus: "success", DelayTime: 100, LatencyCheckAt: now})
	group := SmartGroup{Name: "Philippines", Countries: "PH", MaxAgeHours: 72}
	ids, err := group.CandidateIDs()
	if err != nil || len(ids) != 3 {
		t.Fatalf("country-or-name candidate IDs = %v, %v", ids, err)
	}
	if group.MatchSource(mislabeled) != "countryName" {
		t.Fatalf("stored GB / name PH should match by name, got %q", group.MatchSource(mislabeled))
	}
	members, stats, err := group.MembersWithStats()
	if err != nil || len(members) != 2 || members[0].ID != good.ID || members[1].ID != mislabeled.ID || stats.CandidateCount != 3 || stats.DelayUnusable != 1 {
		t.Fatalf("members = %v, stats = %+v, err = %v", nodeNames(members), stats, err)
	}
	if members[0].LinkCountry != "" {
		t.Fatalf("name inference must not overwrite landing country: %+v", members[0])
	}
	if country, source := SmartGroupCountryForDisplay(members[0]); country != "PH" || source != "name" {
		t.Fatalf("inferred display country = %q from %q", country, source)
	}
	group.SourceGroups = []string{"source-b"}
	ids, err = group.CandidateIDs()
	if err != nil || len(ids) != 1 || ids[0] == good.ID {
		t.Fatalf("filtered inferred candidates = %v, %v", ids, err)
	}
}

func TestSmartGroupCountryCodeTokenMatchesNameWithoutRule(t *testing.T) {
	setupSubcriptionCopyTestDB(t)
	matched := createSubcriptionTestNode(t, Node{Name: "PH / Manila", LinkName: "PH / Manila", LinkCountry: "US"})
	createSubcriptionTestNode(t, Node{Name: "ALPHABET node", LinkName: "ALPHABET node", LinkCountry: "US"})
	group := SmartGroup{Name: "PH", Countries: "PH"}
	ids, err := group.CandidateIDs()
	if err != nil || len(ids) != 1 || ids[0] != matched.ID {
		t.Fatalf("country code token candidates = %v, %v", ids, err)
	}
	if got := group.MatchSource(matched); got != "countryName" {
		t.Fatalf("match source = %q, want countryName", got)
	}
}

func TestSmartGroupAutomaticCheckValidation(t *testing.T) {
	for _, tt := range []struct {
		name, cron             string
		profile                int
		schedule, after, valid bool
	}{
		{"legacy", "", 0, false, false, true},
		{"named-weekday", "0/15 * * * MON", 1, true, false, true},
		{"scheduled", " 0  3 * * * ", 1, true, false, true},
		{"update-only", "", 1, false, true, true},
		{"missing-profile", "0 3 * * *", 0, true, false, false},
		{"update-missing-profile", "", 0, false, true, false},
		{"negative-profile", "", -1, false, false, false},
		{"missing-cron", "", 1, true, false, false},
		{"zero-step", "*/0 * * * *", 1, true, false, false},
		{"six-fields", "0 0 3 * * *", 1, true, false, false},
		{"timezone-prefix", "CRON_TZ=UTC 0 3 * * *", 1, true, false, false},
		{"invalid-hour", "0 24 * * *", 1, true, false, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			group := SmartGroup{Name: tt.name, Countries: "GB", CheckProfileID: tt.profile, CronExpr: tt.cron, ScheduleEnabled: tt.schedule, CheckAfterAirportUpdate: tt.after}
			err := group.Validate()
			if (err == nil) != tt.valid {
				t.Fatalf("Validate() = %v, want valid %v", err, tt.valid)
			}
			if tt.name == "scheduled" && group.CronExpr != "0 3 * * *" {
				t.Fatalf("Cron not normalized: %q", group.CronExpr)
			}
		})
	}
}

func TestSmartGroupCountryRulesDoNotMatchProtocolNames(t *testing.T) {
	setupSubcriptionCopyTestDB(t)
	oldRules := countryRuleCache
	t.Cleanup(func() { countryRuleCache = oldRules })
	resetCountryRuleCacheForTest()
	if err := database.DB.AutoMigrate(&CountryRule{}); err != nil {
		t.Fatal(err)
	}
	for _, rule := range []CountryRule{
		{ID: 920001, CountryCode: "TR", CountryName: "土耳其", Pattern: "(?i)土耳其|TR|Turkey|🇹🇷", Enabled: true},
		{ID: 920002, CountryCode: "ES", CountryName: "西班牙", Pattern: "(?i)西班牙|ES|Spain|🇪🇸", Enabled: true},
		{ID: 920003, CountryCode: "RO", CountryName: "罗马尼亚", Pattern: "(?i)罗马尼亚|RO|Romania|🇷🇴", Enabled: true},
		{ID: 920004, CountryCode: "IN", CountryName: "印度", Pattern: "(?i)印度|IN|India|🇮🇳", Enabled: true},
	} {
		if err := database.DB.Create(&rule).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := InitCountryRuleCache(); err != nil {
		t.Fatal(err)
	}
	now := time.Now().Format("2006-01-02 15:04:05")
	add := func(name, protocol, country string) Node {
		return createSubcriptionTestNode(t, Node{Name: name, LinkName: name, Protocol: protocol, LinkCountry: country, DelayStatus: "success", DelayTime: 100, LatencyCheckAt: now})
	}
	unrelated := []Node{
		add("trojan Hong Kong", "trojan", "HK"),
		add("vmess Singapore", "vmess", "SG"),
		add("vless Japan", "vless", "JP"),
		add("🇭🇰 Trojan 01", "trojan", ""),
	}
	trueTR := []Node{add("🇹🇷 Istanbul trojan", "trojan", ""), add("TR-01", "ss", ""), add("土耳其 02", "vless", ""), add("Turkey 03", "vmess", ""), add("Ankara", "trojan", "TR")}
	for _, code := range []string{"TR", "ES", "RO", "IN"} {
		group := SmartGroup{Name: code, Countries: code, MaxAgeHours: 72}
		for _, n := range unrelated {
			if got := group.MatchSource(n); got != "" {
				t.Errorf("%s matched protocol/non-country name %q via %s", code, n.LinkName, got)
			}
		}
	}
	group := SmartGroup{Name: "Turkey", Countries: "TR", MaxAgeHours: 72}
	ids, err := group.CandidateIDs()
	want := make([]int, 0, len(trueTR))
	for _, n := range trueTR {
		want = append(want, n.ID)
	}
	if err != nil || !reflect.DeepEqual(ids, want) {
		t.Fatalf("TR candidates=%v want=%v err=%v", ids, want, err)
	}
	members, err := group.Members()
	if err != nil || len(members) != len(trueTR) {
		t.Fatalf("TR members=%v err=%v", nodeNames(members), err)
	}
	for _, n := range unrelated {
		n.LinkCountry = ""
		if code, source := SmartGroupCountryForDisplay(n); code != "" || source != "unknown" {
			t.Errorf("protocol name %q displayed as %s (%s)", n.Name, code, source)
		}
	}
	if code, source := SmartGroupCountryForDisplay(trueTR[0]); code != "TR" || source != "name" {
		t.Fatalf("flag country=%s source=%s", code, source)
	}
	// Explicit node-name keywords remain an intentional OR; they are not an
	// implicit protocol filter and country selection alone cannot activate them.
	group.Keyword = "trojan"
	if group.MatchSource(unrelated[0]) != "keyword" {
		t.Fatal("explicit keyword semantics changed")
	}
	if err := database.DB.Create(&group).Error; err != nil {
		t.Fatal(err)
	}
	group.Keyword = ""
	if err := database.DB.Model(&group).Update("keyword", "").Error; err != nil {
		t.Fatal(err)
	}
	sub := &Subcription{Name: "Turkey-only", SmartGroupIDs: strconv.Itoa(group.ID), RefreshUsageOnRequest: true}
	if err := sub.Add(); err != nil {
		t.Fatal(err)
	}
	if err := sub.GetSub("clash"); err != nil || len(sub.Nodes) != len(trueTR) {
		t.Fatalf("subscription country leaked: %v err=%v", nodeNames(sub.Nodes), err)
	}
}

func TestSmartGroupBoundCountryAbbreviationsKeepsExplicitRegex(t *testing.T) {
	for _, tt := range []struct{ pattern, want string }{
		{"(?i)TR|Turkey|🇹🇷", "(?i)(?<![A-Za-z])TR(?![A-Za-z])|Turkey|🇹🇷"},
		{"(?i)英国|GB|UK|United Kingdom|🇬🇧", "(?i)英国|(?<![A-Za-z])GB(?![A-Za-z])|(?<![A-Za-z])UK(?![A-Za-z])|United Kingdom|🇬🇧"},
		{"(?i)(?<![A-Za-z])TR(?![A-Za-z])|Turkey", "(?i)(?<![A-Za-z])TR(?![A-Za-z])|Turkey"},
		{"(?i)custom-TR.*|[A-Z]{2}", "(?i)custom-TR.*|[A-Z]{2}"},
		{"(?i)(TR|RO)|Turkey", "(?i)(TR|RO)|Turkey"},
		{"[A|TR|B]|Turkey", "[A|TR|B]|Turkey"},
		{"[]|TR|B]|Turkey", "[]|TR|B]|Turkey"},
		{"[^]|TR|B]|Turkey", "[^]|TR|B]|Turkey"},
		{`\Qfoo|TR|bar\E`, `\Qfoo|TR|bar\E`},
		{"(?#ignore|TR|B)Turkey|TR", "(?#ignore|TR|B)Turkey|(?<![A-Za-z])TR(?![A-Za-z])"},
		{"(?ix)Turkey # ignore|TR|B\n|TR", "(?ix)Turkey # ignore|TR|B\n|TR"},
		{`foo\|TR|Turkey`, `foo\|TR|Turkey`},
		{`[A\]|TR|B]|TR`, `[A\]|TR|B]|(?<![A-Za-z])TR(?![A-Za-z])`},
	} {
		got := smartGroupBoundCountryAbbreviations(tt.pattern)
		if got != tt.want {
			t.Errorf("%q => %q want %q", tt.pattern, got, tt.want)
		}
		if _, err := countryRegexp.Compile(got); err != nil {
			t.Errorf("transformed regex %q no longer compiles: %v", got, err)
		}
	}
}

func TestSmartGroupExactFlagAndCountryCodeWithoutRules(t *testing.T) {
	oldRules := countryRuleCache
	t.Cleanup(func() { countryRuleCache = oldRules })
	resetCountryRuleCacheForTest()
	group := SmartGroup{Name: "Turkey", Countries: "TR"}
	for _, tt := range []struct {
		name, protocol string
		want           bool
	}{
		{"🇹🇷 node", "trojan", true},
		{"🇹🇷 node", "vmess", true},
		{"TR / 01", "vless", true},
		{"tr-02", "ss", true},
		{"trojan 01", "trojan", false},
		{"🇭🇰 trojan 01", "trojan", false},
		{"transit node", "ss", false},
		{"countryless", "trojan", false},
	} {
		node := Node{Name: tt.name, LinkName: tt.name, Protocol: tt.protocol}
		if got := group.MatchSource(node); (got != "") != tt.want {
			t.Errorf("name=%q protocol=%s match=%s want=%v", tt.name, tt.protocol, got, tt.want)
		}
	}
}

func TestSmartGroupCountryPatternCacheTracksRuleUpdates(t *testing.T) {
	rule := CountryRule{ID: 930001, CountryCode: "TR", Pattern: "(?i)TR|Turkey", Enabled: true}
	first := smartGroupCachedCountryRegex(rule)
	if first == nil || first.MatchString("trojan") || !first.MatchString("tr-01") {
		t.Fatal("guarded country pattern did not enforce boundaries")
	}
	if smartGroupCachedCountryRegex(rule) != first {
		t.Fatal("unchanged country pattern was recompiled")
	}
	rule.Pattern = "Istanbul"
	updated := smartGroupCachedCountryRegex(rule)
	if updated == nil || updated == first || !updated.MatchString("Istanbul") || updated.MatchString("Turkey") {
		t.Fatal("country pattern update retained stale matching")
	}
	rule.Enabled = false
	if len(smartGroupCountryRulePatterns([]CountryRule{rule})) != 0 {
		t.Fatal("disabled country rule used cached regex")
	}
}

func TestSmartGroupCountryAliasBoundaries(t *testing.T) {
	for _, alias := range []string{"TR", "ES", "IN", "UK", "USA", "UAE", "LA"} {
		pattern, err := countryRegexp.Compile(smartGroupBoundCountryAbbreviations("(?i)" + alias))
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{alias, alias + "-01", "01 " + alias, "🇹🇷" + alias + "节点"} {
			if !pattern.MatchString(name) {
				t.Errorf("alias %s missed standalone name %q", alias, name)
			}
		}
		for _, name := range []string{"prefix" + alias, alias + "suffix", "pre" + alias + "post"} {
			if pattern.MatchString(name) {
				t.Errorf("alias %s matched inside word %q", alias, name)
			}
		}
	}
}

func TestSmartGroupSubscriptionsDeduplicateSourcesAndPreserveCopies(t *testing.T) {
	setupSubcriptionCopyTestDB(t)
	now := time.Now().Format("2006-01-02 15:04:05")
	a := createSubcriptionTestNode(t, Node{Name: "one", LinkName: "one", Group: "source", LinkCountry: "GB", DelayStatus: "success", DelayTime: 100, LatencyCheckAt: now})
	b := createSubcriptionTestNode(t, Node{Name: "two", LinkName: "two", LinkCountry: "DE", DelayStatus: "success", DelayTime: 120, LatencyCheckAt: now})
	first := SmartGroup{Name: "Europe", Countries: "GB,DE", MaxAgeHours: 72}
	second := SmartGroup{Name: "Germany", Countries: "DE", MaxAgeHours: 72}
	for _, g := range []*SmartGroup{&first, &second} {
		if err := database.DB.Create(g).Error; err != nil {
			t.Fatal(err)
		}
	}
	sub := &Subcription{Name: "mixed-smart", SmartGroupIDs: strconv.Itoa(first.ID) + "," + strconv.Itoa(second.ID), RefreshUsageOnRequest: true}
	if err := sub.Add(); err != nil {
		t.Fatal(err)
	}
	if err := database.DB.Create(&SubcriptionNode{SubcriptionID: sub.ID, NodeID: a.ID, Sort: 0}).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.DB.Create(&SubcriptionGroup{SubcriptionID: sub.ID, GroupName: "source", Sort: 1}).Error; err != nil {
		t.Fatal(err)
	}
	if err := sub.GetSub("clash"); err != nil {
		t.Fatal(err)
	}
	if len(sub.Nodes) != 2 || sub.Nodes[0].ID != a.ID || sub.Nodes[1].ID != b.ID {
		t.Fatalf("duplicate source contributions: %+v", sub.Nodes)
	}
	copySub, err := sub.Copy()
	if err != nil || copySub.SmartGroupIDs != sub.SmartGroupIDs {
		t.Fatalf("copy lost group IDs: %+v, %v", copySub, err)
	}
	// A fresh result can become unusable without modifying the smart group.
	b.DelayStatus = "timeout"
	if err := b.UpdateSpeed(); err != nil {
		t.Fatal(err)
	}
	if err := copySub.GetSub("clash"); err != nil || len(copySub.Nodes) != 1 || copySub.Nodes[0].ID != a.ID {
		t.Fatalf("membership did not update: %+v, %v", copySub.Nodes, err)
	}
}
