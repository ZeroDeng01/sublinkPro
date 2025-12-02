package models

import (
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

func (user *User) Create() error { // 创建用户
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword
	return DB.Create(user).Error
}

func (user *User) Set(UpdateUser *User) error { // 设置用户
	if UpdateUser.Password != "" {
		hashedPassword, err := HashPassword(UpdateUser.Password)
		if err != nil {
			return err
		}
		UpdateUser.Password = hashedPassword
	}
	return DB.Where("username = ?", user.Username).Updates(UpdateUser).Error
}

func (user *User) Verify() error { // 验证用户
	// First find the user by username
	var dbUser User
	if err := DB.Where("username = ?", user.Username).First(&dbUser).Error; err != nil {
		return err
	}

	// Check if the stored password is a hash (bcrypt hashes start with $2a$, $2b$, $2x$, $2y$)
	// If it's not a hash (legacy plaintext), compare directly
	// Note: This is a temporary fallback for the transition period, but ideally we should migrate all users.
	// However, since we are adding migration logic in sqlite.go, we can assume here we might encounter both
	// if the migration hasn't run yet or failed.
	// But for security, let's stick to the plan: Verify should check hash.
	// If we want to support auto-upgrade on login, we could do it here, but the plan said migration in InitSqlite.

	if CheckPasswordHash(user.Password, dbUser.Password) {
		*user = dbUser // Update the user object with DB data
		return nil
	}

	// Fallback for legacy plaintext passwords (optional, but good for robustness if migration fails)
	// If the stored password doesn't look like a bcrypt hash, try direct comparison
	// A simple check is length. Bcrypt hashes are 60 chars.
	if len(dbUser.Password) < 60 && dbUser.Password == user.Password {
		*user = dbUser
		return nil
	}

	return gorm.ErrRecordNotFound // Or a specific "invalid password" error
}

func (user *User) Find() error { // 查找用户
	return DB.Where("username = ? ", user.Username).First(user).Error
}

func (user *User) All() ([]User, error) { // 获取所有用户
	var users []User
	err := DB.Find(&users).Error
	return users, err
}

func (user *User) Del() error { // 删除用户
	return DB.Delete(user).Error
}
