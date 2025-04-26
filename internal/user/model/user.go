package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID        int64     `gorm:"primaryKey;column:id"`
	Username  string    `gorm:"column:username;type:varchar(24);not null;uniqueIndex:idx_username"`
	Password  string    `gorm:"column:password;type:varchar(128);not null"`
	Email     string    `gorm:"column:email;type:varchar(64);not null;uniqueIndex:idx_email"`
	Phone     string    `gorm:"column:phone;type:varchar(20);default:'';index:idx_phone"`
	Avatar    string    `gorm:"column:avatar;type:varchar(255);default:''"`
	RealName  string    `gorm:"column:real_name;type:varchar(32);default:''"`
	Gender    string    `gorm:"column:gender;type:varchar(10);default:''"`
	Birthday  string    `gorm:"column:birthday;type:varchar(10);default:''"`
	Address   string    `gorm:"column:address;type:varchar(255);default:''"`
	Bio       string    `gorm:"column:bio;type:text"`
	Status    int8      `gorm:"column:status;type:tinyint(1);default:1"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 表名
func (User) TableName() string {
	return "users"
}

// OAuthAccount OAuth账号模型
type OAuthAccount struct {
	ID             int64     `gorm:"primaryKey;column:id"`
	UserID         int64     `gorm:"column:user_id;type:bigint(20);not null;index:idx_user_id"`
	Provider       string    `gorm:"column:provider;type:varchar(32);not null"`
	ProviderUserID string    `gorm:"column:provider_user_id;type:varchar(128);not null;uniqueIndex:idx_provider_uid,priority:1"`
	AccessToken    string    `gorm:"column:access_token;type:varchar(255);default:''"`
	RefreshToken   string    `gorm:"column:refresh_token;type:varchar(255);default:''"`
	ExpiresAt      time.Time `gorm:"column:expires_at"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 表名
func (OAuthAccount) TableName() string {
	return "oauth_accounts"
}

// Token 令牌模型
type Token struct {
	ID                    int64     `gorm:"primaryKey;column:id"`
	UserID                int64     `gorm:"column:user_id;type:bigint(20);not null;index:idx_user_id"`
	RefreshToken          string    `gorm:"column:refresh_token;type:varchar(255);not null;uniqueIndex:idx_refresh_token"`
	RefreshTokenExpiresAt time.Time `gorm:"column:refresh_token_expires_at;not null"`
	CreatedAt             time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt             time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 表名
func (Token) TableName() string {
	return "tokens"
}
