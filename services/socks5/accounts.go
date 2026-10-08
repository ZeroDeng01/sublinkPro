package socks5

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	"sublink/models"
)

const (
	settingAccounts = "socks5_accounts"
	maxAccounts     = 128
)

var (
	accountIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,31}$`)
	accountsMu       sync.Mutex
)

// Account is an independently authenticated SOCKS5 identity bound to a routing profile.
type Account struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"-"`
	ProfileID string `json:"profileId"`
	Enabled   bool   `json:"enabled"`
}

type storedAccount struct {
	ID                string `json:"id"`
	Username          string `json:"username"`
	EncryptedPassword string `json:"password"`
	ProfileID         string `json:"profileId"`
	Enabled           bool   `json:"enabled"`
}

// PublicAccount is safe for API responses and never exposes the password.
type PublicAccount struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	ProfileID      string `json:"profileId"`
	Enabled        bool   `json:"enabled"`
	HasPassword    bool   `json:"hasPassword"`
	MaskedPassword string `json:"maskedPassword,omitempty"`
}

func accountToPublic(account Account) PublicAccount {
	masked := ""
	if account.Password != "" {
		masked = "••••••••"
	}
	return PublicAccount{
		ID: account.ID, Username: account.Username, ProfileID: account.ProfileID, Enabled: account.Enabled,
		HasPassword: account.Password != "", MaskedPassword: masked,
	}
}

func normalizeAccountFields(account Account) (Account, error) {
	account.ID = strings.ToLower(strings.TrimSpace(account.ID))
	account.Username = strings.TrimSpace(account.Username)
	account.ProfileID = strings.ToLower(strings.TrimSpace(account.ProfileID))
	if account.ProfileID == "" {
		account.ProfileID = defaultProfileID
	}
	if !accountIDPattern.MatchString(account.ID) {
		return account, errors.New("SOCKS5 account ID must start with a letter or digit and contain only lowercase letters, digits, underscores, or hyphens")
	}
	if account.Username == "" || len(account.Username) > 255 {
		return account, errors.New("SOCKS5 account username must contain 1-255 bytes")
	}
	if strings.ContainsAny(account.Username, "@\x00") {
		return account, errors.New("SOCKS5 account username cannot contain @ or NUL")
	}
	if account.Password == "" || len(account.Password) > 255 {
		return account, errors.New("SOCKS5 account password must contain 1-255 bytes")
	}
	return account, nil
}

func normalizeAccount(base Config, account Account) (Account, error) {
	account, err := normalizeAccountFields(account)
	if err != nil {
		return account, err
	}
	if strings.EqualFold(account.Username, strings.TrimSpace(base.Username)) {
		return account, errors.New("SOCKS5 account username conflicts with the legacy gateway username")
	}
	profiles, err := newRoutingProfileStore(base)
	if err != nil {
		return account, err
	}
	if _, _, ok := profiles.resolve(account.ProfileID); !ok {
		return account, errors.New("SOCKS5 account routing profile was not found or is disabled")
	}
	return account, nil
}

func loadAccounts(base Config) ([]Account, error) {
	value, err := models.GetSetting(settingAccounts)
	if err != nil || strings.TrimSpace(value) == "" {
		return []Account{}, nil
	}
	var stored []storedAccount
	if err := json.Unmarshal([]byte(value), &stored); err != nil {
		return nil, fmt.Errorf("invalid SOCKS5 accounts: %w", err)
	}
	if len(stored) > maxAccounts {
		return nil, fmt.Errorf("SOCKS5 accounts exceed the limit of %d", maxAccounts)
	}
	accounts := make([]Account, 0, len(stored))
	seenIDs := make(map[string]struct{}, len(stored))
	seenUsers := make(map[string]struct{}, len(stored))
	for _, item := range stored {
		password, decryptErr := models.DecryptUserAISecret(item.EncryptedPassword)
		if decryptErr != nil {
			return nil, fmt.Errorf("decrypt SOCKS5 account %q password: %w", item.ID, decryptErr)
		}
		account, normalizeErr := normalizeAccountFields(Account{
			ID: item.ID, Username: item.Username, Password: password, ProfileID: item.ProfileID, Enabled: item.Enabled,
		})
		if normalizeErr != nil {
			return nil, fmt.Errorf("invalid SOCKS5 account %q: %w", item.ID, normalizeErr)
		}
		usernameKey := strings.ToLower(account.Username)
		if _, exists := seenIDs[account.ID]; exists {
			return nil, fmt.Errorf("duplicate SOCKS5 account ID %q", account.ID)
		}
		if _, exists := seenUsers[usernameKey]; exists {
			return nil, fmt.Errorf("duplicate SOCKS5 account username %q", account.Username)
		}
		seenIDs[account.ID] = struct{}{}
		seenUsers[usernameKey] = struct{}{}
		accounts = append(accounts, account)
	}
	sort.Slice(accounts, func(i, j int) bool { return accounts[i].ID < accounts[j].ID })
	return accounts, nil
}

func saveAccounts(accounts []Account) error {
	stored := make([]storedAccount, 0, len(accounts))
	for _, account := range accounts {
		encrypted, err := models.EncryptUserAISecret(account.Password)
		if err != nil {
			return fmt.Errorf("encrypt SOCKS5 account %q password: %w", account.ID, err)
		}
		stored = append(stored, storedAccount{
			ID: account.ID, Username: account.Username, EncryptedPassword: encrypted,
			ProfileID: account.ProfileID, Enabled: account.Enabled,
		})
	}
	encoded, err := json.Marshal(stored)
	if err != nil {
		return err
	}
	return models.SetSetting(settingAccounts, string(encoded))
}

func ListAccounts(base Config) ([]PublicAccount, error) {
	accounts, err := loadAccounts(base)
	if err != nil {
		return nil, err
	}
	result := make([]PublicAccount, 0, len(accounts))
	for _, account := range accounts {
		result = append(result, accountToPublic(account))
	}
	return result, nil
}

func CreateAccount(base Config, input Account) (PublicAccount, error) {
	accountsMu.Lock()
	defer accountsMu.Unlock()
	accounts, err := loadAccounts(base)
	if err != nil {
		return PublicAccount{}, err
	}
	if len(accounts) >= maxAccounts {
		return PublicAccount{}, fmt.Errorf("SOCKS5 accounts are limited to %d", maxAccounts)
	}
	normalized, err := normalizeAccount(base, input)
	if err != nil {
		return PublicAccount{}, err
	}
	for _, account := range accounts {
		if account.ID == normalized.ID {
			return PublicAccount{}, errors.New("SOCKS5 account ID already exists")
		}
		if strings.EqualFold(account.Username, normalized.Username) {
			return PublicAccount{}, errors.New("SOCKS5 account username already exists")
		}
	}
	accounts = append(accounts, normalized)
	sort.Slice(accounts, func(i, j int) bool { return accounts[i].ID < accounts[j].ID })
	if err := saveAccounts(accounts); err != nil {
		return PublicAccount{}, err
	}
	return accountToPublic(normalized), nil
}

func UpdateAccount(base Config, id string, input Account, clearPassword bool) (PublicAccount, error) {
	accountsMu.Lock()
	defer accountsMu.Unlock()
	id = strings.ToLower(strings.TrimSpace(id))
	accounts, err := loadAccounts(base)
	if err != nil {
		return PublicAccount{}, err
	}
	index := -1
	for i := range accounts {
		if accounts[i].ID == id {
			index = i
			break
		}
	}
	if index < 0 {
		return PublicAccount{}, errors.New("SOCKS5 account was not found")
	}
	input.ID = id
	if clearPassword {
		return PublicAccount{}, errors.New("SOCKS5 account password cannot be cleared")
	}
	if input.Password == "" {
		input.Password = accounts[index].Password
	}
	normalized, err := normalizeAccount(base, input)
	if err != nil {
		return PublicAccount{}, err
	}
	for i, account := range accounts {
		if i != index && strings.EqualFold(account.Username, normalized.Username) {
			return PublicAccount{}, errors.New("SOCKS5 account username already exists")
		}
	}
	accounts[index] = normalized
	if err := saveAccounts(accounts); err != nil {
		return PublicAccount{}, err
	}
	return accountToPublic(normalized), nil
}

func DeleteAccount(base Config, id string) error {
	accountsMu.Lock()
	defer accountsMu.Unlock()
	id = strings.ToLower(strings.TrimSpace(id))
	accounts, err := loadAccounts(base)
	if err != nil {
		return err
	}
	result := make([]Account, 0, len(accounts))
	found := false
	for _, account := range accounts {
		if account.ID == id {
			found = true
			continue
		}
		result = append(result, account)
	}
	if !found {
		return errors.New("SOCKS5 account was not found")
	}
	return saveAccounts(result)
}

func RoutingProfileReferencedByAccount(base Config, profileID string) (bool, error) {
	accounts, err := loadAccounts(base)
	if err != nil {
		return false, err
	}
	profileID = strings.ToLower(strings.TrimSpace(profileID))
	for _, account := range accounts {
		if account.ProfileID == profileID {
			return true, nil
		}
	}
	return false, nil
}

type accountStore struct {
	mu       sync.RWMutex
	accounts map[string]Account
}

func newAccountStore(base Config) (*accountStore, error) {
	accounts, err := loadAccounts(base)
	if err != nil {
		return nil, err
	}
	store := &accountStore{}
	store.replace(accounts)
	return store, nil
}

func (s *accountStore) replace(accounts []Account) {
	mapped := make(map[string]Account, len(accounts))
	for _, account := range accounts {
		mapped[account.Username] = account
	}
	s.mu.Lock()
	s.accounts = mapped
	s.mu.Unlock()
}

func (s *accountStore) authenticate(username string, password []byte) (account Account, found bool, authenticated bool) {
	s.mu.RLock()
	account, found = s.accounts[username]
	s.mu.RUnlock()
	if !found || !account.Enabled {
		return account, found, false
	}
	if subtle.ConstantTimeCompare(password, []byte(account.Password)) != 1 {
		return account, true, false
	}
	return account, true, true
}
