package service

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"

	"user_service/internal/pkg/conf"
	"user_service/internal/pkg/errno"
	"user_service/internal/pkg/jwt"
	"user_service/internal/user/dao"
	"user_service/internal/user/model"
)

// UserService 用户服务
type UserService struct {
	userDAO  *dao.UserDAO
	oauthDAO *dao.OAuthDAO
	tokenDAO *dao.TokenDAO
	conf     *conf.AuthConfig
}

// NewUserService 创建用户服务
func NewUserService(userDAO *dao.UserDAO, oauthDAO *dao.OAuthDAO, tokenDAO *dao.TokenDAO, conf *conf.AuthConfig) *UserService {
	return &UserService{
		userDAO:  userDAO,
		oauthDAO: oauthDAO,
		tokenDAO: tokenDAO,
		conf:     conf,
	}
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, username, password, email, phone string) (int64, error) {
	// 检查用户名是否存在
	existUser, err := s.userDAO.GetByUsername(ctx, username)
	if err != nil {
		hlog.Errorf("通过用户名获取用户失败: %v", err)
		return 0, errno.InternalError
	}
	if existUser != nil {
		return 0, errno.UserExistError
	}

	// 检查邮箱是否存在
	existUser, err = s.userDAO.GetByEmail(ctx, email)
	if err != nil {
		hlog.Errorf("通过邮箱获取用户失败: %v", err)
		return 0, errno.InternalError
	}
	if existUser != nil {
		return 0, errno.UserExistError
	}

	// 创建用户
	user := &model.User{
		Username: username,
		Password: password,
		Email:    email,
		Phone:    phone,
		Status:   1, // 启用
	}

	userID, err := s.userDAO.Create(ctx, user)
	if err != nil {
		hlog.Errorf("创建用户失败: %v", err)
		return 0, errno.InternalError
	}

	return userID, nil
}

// Login 用户登录
func (s *UserService) Login(ctx context.Context, username, password string) (int64, string, time.Time, error) {
	// 获取用户
	user, err := s.userDAO.GetByUsername(ctx, username)
	if err != nil {
		hlog.Errorf("通过用户名获取用户失败: %v", err)
		return 0, "", time.Time{}, errno.InternalError
	}
	if user == nil {
		return 0, "", time.Time{}, errno.UserNotExistError
	}

	// 验证密码
	valid, err := s.userDAO.VerifyPassword(ctx, user.ID, password)
	if err != nil {
		hlog.Errorf("验证密码失败: %v", err)
		return 0, "", time.Time{}, errno.InternalError
	}
	if !valid {
		return 0, "", time.Time{}, errno.PasswordError
	}

	// 生成Token
	token, expiresAt, err := s.generateToken(ctx, user.ID)
	if err != nil {
		hlog.Errorf("生成令牌失败: %v", err)
		return 0, "", time.Time{}, errno.InternalError
	}

	return user.ID, token, expiresAt, nil
}

// GetUserInfo 获取用户信息
func (s *UserService) GetUserInfo(ctx context.Context, userID int64) (*model.User, error) {
	user, err := s.userDAO.GetByID(ctx, userID)
	if err != nil {
		hlog.Errorf("通过ID获取用户失败: %v", err)
		return nil, errno.InternalError
	}
	if user == nil {
		return nil, errno.UserNotExistError
	}

	// 不返回密码
	user.Password = ""

	return user, nil
}

// UpdateUserInfo 更新用户信息
func (s *UserService) UpdateUserInfo(ctx context.Context, user *model.User) error {
	// 获取原用户信息
	existUser, err := s.userDAO.GetByID(ctx, user.ID)
	if err != nil {
		hlog.Errorf("通过ID获取用户失败: %v", err)
		return errno.InternalError
	}
	if existUser == nil {
		return errno.UserNotExistError
	}

	// 保留不修改的字段
	user.Password = existUser.Password
	user.Status = existUser.Status
	user.CreatedAt = existUser.CreatedAt

	// 更新用户
	if err := s.userDAO.Update(ctx, user); err != nil {
		hlog.Errorf("更新用户失败: %v", err)
		return errno.InternalError
	}

	return nil
}

// ResetPassword 重置密码
func (s *UserService) ResetPassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	// 验证旧密码
	valid, err := s.userDAO.VerifyPassword(ctx, userID, oldPassword)
	if err != nil {
		hlog.Errorf("验证密码失败: %v", err)
		return errno.InternalError
	}
	if !valid {
		return errno.PasswordError
	}

	// 更新密码
	if err := s.userDAO.UpdatePassword(ctx, userID, newPassword); err != nil {
		hlog.Errorf("更新密码失败: %v", err)
		return errno.InternalError
	}

	// 删除用户的所有令牌，强制重新登录
	if err := s.tokenDAO.DeleteByUserID(ctx, userID); err != nil {
		hlog.Errorf("删除用户令牌失败: %v", err)
		// 不返回错误，密码已经更新成功
	}

	return nil
}

// 生成令牌
func (s *UserService) generateToken(ctx context.Context, userID int64) (string, time.Time, error) {
	// 生成JWT访问令牌
	expiresIn := s.conf.JWTExpiration
	expiresAt := time.Now().Add(expiresIn)
	accessToken, err := jwt.GenerateToken(s.conf.JWTSecret, userID, expiresIn)
	if err != nil {
		return "", time.Time{}, err
	}

	// 生成刷新令牌
	refreshExpiresIn := s.conf.RefreshExpiration
	refreshExpiresAt := time.Now().Add(refreshExpiresIn)
	refreshToken, err := jwt.GenerateRefreshToken(s.conf.JWTSecret, userID, refreshExpiresIn)
	if err != nil {
		return "", time.Time{}, err
	}

	// 存储刷新令牌
	token := &model.Token{
		UserID:                userID,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshExpiresAt,
	}

	// 删除旧的刷新令牌
	if err := s.tokenDAO.DeleteByUserID(ctx, userID); err != nil {
		hlog.Warnf("删除用户旧令牌失败: %v", err)
	}

	// 创建新的刷新令牌
	if _, err := s.tokenDAO.Create(ctx, token); err != nil {
		return "", time.Time{}, err
	}

	return accessToken, expiresAt, nil
}
