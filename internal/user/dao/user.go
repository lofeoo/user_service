package dao

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"user_service/internal/pkg/db"
	"user_service/internal/user/model"
)

// UserDAO 用户数据访问对象
type UserDAO struct {
	db *gorm.DB
}

// NewUserDAO 创建用户DAO
func NewUserDAO() *UserDAO {
	return &UserDAO{db: db.DB}
}

// WithContext 获取事务
func (dao *UserDAO) WithContext(ctx context.Context) *gorm.DB {
	return dao.db.WithContext(ctx)
}

// Create 创建用户
func (dao *UserDAO) Create(ctx context.Context, user *model.User) (int64, error) {
	// 对密码进行加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	user.Password = string(hashedPassword)

	if err := dao.WithContext(ctx).Create(user).Error; err != nil {
		return 0, err
	}
	return user.ID, nil
}

// GetByID 通过ID获取用户
func (dao *UserDAO) GetByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	if err := dao.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByUsername 通过用户名获取用户
func (dao *UserDAO) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	if err := dao.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail 通过邮箱获取用户
func (dao *UserDAO) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := dao.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (dao *UserDAO) Update(ctx context.Context, user *model.User) error {
	return dao.WithContext(ctx).Save(user).Error
}

// UpdatePassword 更新密码
func (dao *UserDAO) UpdatePassword(ctx context.Context, id int64, password string) error {
	// 对密码进行加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return dao.WithContext(ctx).Model(&model.User{}).Where("id = ?", id).
		Update("password", string(hashedPassword)).Error
}

// VerifyPassword 验证密码
func (dao *UserDAO) VerifyPassword(ctx context.Context, id int64, password string) (bool, error) {
	var user model.User
	if err := dao.WithContext(ctx).Select("password").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil, nil
}

// OAuthDAO OAuth账号数据访问对象
type OAuthDAO struct {
	db *gorm.DB
}

// NewOAuthDAO 创建OAuth DAO
func NewOAuthDAO() *OAuthDAO {
	return &OAuthDAO{db: db.DB}
}

// WithContext 获取事务
func (dao *OAuthDAO) WithContext(ctx context.Context) *gorm.DB {
	return dao.db.WithContext(ctx)
}

// Create 创建OAuth账号
func (dao *OAuthDAO) Create(ctx context.Context, oauth *model.OAuthAccount) (int64, error) {
	if err := dao.WithContext(ctx).Create(oauth).Error; err != nil {
		return 0, err
	}
	return oauth.ID, nil
}

// GetByProviderAndID 通过提供商和提供商用户ID获取OAuth账号
func (dao *OAuthDAO) GetByProviderAndID(ctx context.Context, provider, providerUserID string) (*model.OAuthAccount, error) {
	var oauth model.OAuthAccount
	if err := dao.WithContext(ctx).Where("provider = ? AND provider_user_id = ?", provider, providerUserID).
		First(&oauth).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &oauth, nil
}

// UpdateToken 更新OAuth Token
func (dao *OAuthDAO) UpdateToken(ctx context.Context, id int64, accessToken, refreshToken string, expiresAt time.Time) error {
	return dao.WithContext(ctx).Model(&model.OAuthAccount{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"expires_at":    expiresAt,
		}).Error
}

// TokenDAO 令牌数据访问对象
type TokenDAO struct {
	db *gorm.DB
}

// NewTokenDAO 创建Token DAO
func NewTokenDAO() *TokenDAO {
	return &TokenDAO{db: db.DB}
}

// WithContext 获取事务
func (dao *TokenDAO) WithContext(ctx context.Context) *gorm.DB {
	return dao.db.WithContext(ctx)
}

// Create 创建令牌
func (dao *TokenDAO) Create(ctx context.Context, token *model.Token) (int64, error) {
	if err := dao.WithContext(ctx).Create(token).Error; err != nil {
		return 0, err
	}
	return token.ID, nil
}

// GetByRefreshToken 通过刷新令牌获取令牌
func (dao *TokenDAO) GetByRefreshToken(ctx context.Context, refreshToken string) (*model.Token, error) {
	var token model.Token
	if err := dao.WithContext(ctx).Where("refresh_token = ?", refreshToken).First(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

// DeleteByRefreshToken 通过刷新令牌删除令牌
func (dao *TokenDAO) DeleteByRefreshToken(ctx context.Context, refreshToken string) error {
	return dao.WithContext(ctx).Where("refresh_token = ?", refreshToken).Delete(&model.Token{}).Error
}

// DeleteByUserID 通过用户ID删除令牌
func (dao *TokenDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	return dao.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.Token{}).Error
}
