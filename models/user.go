package models

import (
	"crypto/sha256"
	"fmt"
	"sublink/cache"
	"sublink/database"
	"sublink/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID       int
	Username string
	Password string
	Role     string
	Nickname string
}

// userCache 使用新的泛型缓存
var userCache *cache.MapCache[int, User]

func init() {
	userCache = cache.NewMapCache(func(u User) int { return u.ID })
	userCache.AddIndex("username", func(u User) string { return u.Username })
}

// InitUserCache 初始化用户缓存
func InitUserCache() error {
	utils.Info("开始加载用户到缓存")
	var users []User
	if err := database.DB.Find(&users).Error; err != nil {
		return err
	}

	userCache.LoadAll(users)
	utils.Info("用户缓存初始化完成，共加载 %d 个用户", userCache.Count())

	cache.Manager.Register("user", userCache)
	return nil
}

// HashPassword hashes the password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash checks if the provided password matches the hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Create 创建用户 (Write-Through)
func (user *User) Create() error {
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword
	err = database.DB.Create(user).Error
	if err != nil {
		return err
	}
	userCache.Set(user.ID, *user)
	return nil
}

// Set 设置用户 (Write-Through)
func (user *User) Set(UpdateUser *User) error {
	if UpdateUser.Password != "" {
		hashedPassword, err := HashPassword(UpdateUser.Password)
		if err != nil {
			return err
		}
		UpdateUser.Password = hashedPassword
	}
	err := database.DB.Where("username = ?", user.Username).Updates(UpdateUser).Error
	if err != nil {
		return err
	}
	// 更新缓存
	var updated User
	if err := database.DB.Where("username = ?", user.Username).First(&updated).Error; err == nil {
		userCache.Set(updated.ID, updated)
	}
	return nil
}

// Verify 验证用户
func (user *User) Verify() error {
	// 先从缓存查找用户
	users := userCache.GetByIndex("username", user.Username)
	var dbUser User
	if len(users) > 0 {
		dbUser = users[0]
	} else {
		// 缓存未命中，从数据库查询
		if err := database.DB.Where("username = ?", user.Username).First(&dbUser).Error; err != nil {
			return err
		}
		userCache.Set(dbUser.ID, dbUser)
	}

	if CheckPasswordHash(user.Password, dbUser.Password) {
		*user = dbUser
		return nil
	}

	// Fallback for legacy plaintext passwords
	if len(dbUser.Password) < 60 && dbUser.Password == user.Password {
		*user = dbUser
		return nil
	}

	return gorm.ErrRecordNotFound
}

// Find 查找用户
func (user *User) Find() error {
	// 先从缓存查找
	users := userCache.GetByIndex("username", user.Username)
	if len(users) > 0 {
		*user = users[0]
		return nil
	}
	return database.DB.Where("username = ? ", user.Username).First(user).Error
}

// All 获取所有用户
func (user *User) All() ([]User, error) {
	return userCache.GetAll(), nil
}

// Del 删除用户 (Write-Through)
func (user *User) Del() error {
	err := database.DB.Delete(user).Error
	if err != nil {
		return err
	}
	userCache.Delete(user.ID)
	return nil
}

// GenerateCredentialSign 生成凭证签名（用户名+密码hash 的 SHA256 前16位）
// 用于 JWT 失效机制：密码或用户名变化后，签名自动失效
func GenerateCredentialSign(username, passwordHash string) string {
	h := sha256.Sum256([]byte(username + passwordHash))
	return fmt.Sprintf("%x", h)[:16]
}

// VerifyCredentialSign 验证凭证签名
// 通过用户名查找用户并验证签名是否匹配
func VerifyCredentialSign(username, sign string) bool {
	users := userCache.GetByIndex("username", username)
	if len(users) == 0 {
		// 缓存未命中时从数据库查询
		var user User
		if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
			return false
		}
		userCache.Set(user.ID, user)
		return GenerateCredentialSign(username, user.Password) == sign
	}
	return GenerateCredentialSign(username, users[0].Password) == sign
}
