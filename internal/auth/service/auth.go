package service

import (
	"context"
	"errors"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"

	"user_service/internal/pkg/conf"
	"user_service/internal/pkg/errno"
	"user_service/internal/pkg/jwt"
	"user_service/internal/pkg/oauth"
	"user_service/internal/user/dao"
	"user_service/internal/user/model"
)

// AuthService 认证服务
type AuthService struct {
	userDAO  *dao.UserDAO
	oauthDAO *dao.OAuthDAO
	tokenDAO *dao.TokenDAO
	conf     *conf.AuthConfig
}

// NewAuthService 创建认证服务
func NewAuthService(userDAO *dao.UserDAO, oauthDAO *dao.OAuthDAO, tokenDAO *dao.TokenDAO, conf *conf.AuthConfig) *AuthService {
	return &AuthService{
		userDAO:  userDAO,
		oauthDAO: oauthDAO,
		tokenDAO: tokenDAO,
		conf:     conf,
	}
}

// ValidateToken 验证令牌
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (bool, int64, error) {
	// 解析令牌
	claims, err := jwt.ParseToken(tokenString, s.conf.JWTSecret)
	if err != nil {
		hlog.Errorf("解析令牌失败: %v", err)
		return false, 0, errno.TokenInvalidError
	}

	// 检查令牌是否过期
	if claims.ExpiresAt.Time.Before(time.Now()) {
		return false, 0, errno.TokenExpiredError
	}

	return true, claims.UserID, nil
}

// RefreshToken 刷新令牌
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, string, time.Time, error) {
	// 解析刷新令牌
	claims, err := jwt.ParseToken(refreshToken, s.conf.JWTSecret)
	if err != nil {
		hlog.Errorf("解析刷新令牌失败: %v", err)
		return "", "", time.Time{}, errno.TokenInvalidError
	}

	// 检查令牌是否过期
	if claims.ExpiresAt.Time.Before(time.Now()) {
		return "", "", time.Time{}, errno.TokenExpiredError
	}

	// 验证刷新令牌是否存在
	token, err := s.tokenDAO.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		hlog.Errorf("获取刷新令牌失败: %v", err)
		return "", "", time.Time{}, errno.InternalError
	}
	if token == nil {
		return "", "", time.Time{}, errno.TokenInvalidError
	}

	// 检查刷新令牌是否过期
	if token.RefreshTokenExpiresAt.Before(time.Now()) {
		return "", "", time.Time{}, errno.TokenExpiredError
	}

	// 删除旧的刷新令牌
	if err := s.tokenDAO.DeleteByRefreshToken(ctx, refreshToken); err != nil {
		hlog.Errorf("删除旧的刷新令牌失败: %v", err)
		// 继续生成新令牌
	}

	// 生成新的访问令牌
	expiresIn := s.conf.JWTExpiration
	expiresAt := time.Now().Add(expiresIn)
	accessToken, err := jwt.GenerateToken(s.conf.JWTSecret, claims.UserID, expiresIn)
	if err != nil {
		hlog.Errorf("生成访问令牌失败: %v", err)
		return "", "", time.Time{}, errno.InternalError
	}

	// 生成新的刷新令牌
	refreshExpiresIn := s.conf.RefreshExpiration
	refreshExpiresAt := time.Now().Add(refreshExpiresIn)
	newRefreshToken, err := jwt.GenerateRefreshToken(s.conf.JWTSecret, claims.UserID, refreshExpiresIn)
	if err != nil {
		hlog.Errorf("生成刷新令牌失败: %v", err)
		return "", "", time.Time{}, errno.InternalError
	}

	// 存储新的刷新令牌
	newToken := &model.Token{
		UserID:                claims.UserID,
		RefreshToken:          newRefreshToken,
		RefreshTokenExpiresAt: refreshExpiresAt,
	}

	if _, err := s.tokenDAO.Create(ctx, newToken); err != nil {
		hlog.Errorf("存储刷新令牌失败: %v", err)
		return "", "", time.Time{}, errno.InternalError
	}

	return accessToken, newRefreshToken, expiresAt, nil
}

// OAuth2Login OAuth2登录
func (s *AuthService) OAuth2Login(ctx context.Context, provider, code, redirectURI string) (string, string, time.Time, int64, bool, error) {
	// 获取OAuth2提供商
	var clientID, clientSecret string
	switch provider {
	case "github":
		clientID = s.conf.GithubClientID
		clientSecret = s.conf.GithubClientSecret
	case "google":
		clientID = s.conf.GoogleClientID
		clientSecret = s.conf.GoogleClientSecret
	default:
		return "", "", time.Time{}, 0, false, errors.New("不支持的OAuth2提供商")
	}

	// 创建OAuth2提供商
	p, err := oauth.GetProvider(provider, clientID, clientSecret)
	if err != nil {
		hlog.Errorf("获取OAuth2提供商失败: %v", err)
		return "", "", time.Time{}, 0, false, errno.OAuth2LoginError
	}

	// 交换授权码获取访问令牌
	accessToken, err := p.ExchangeToken(ctx, code, redirectURI)
	if err != nil {
		hlog.Errorf("交换OAuth2令牌失败: %v", err)
		return "", "", time.Time{}, 0, false, errno.OAuth2LoginError
	}

	// 获取用户信息
	userInfo, err := p.GetUserInfo(ctx, accessToken)
	if err != nil {
		hlog.Errorf("获取OAuth2用户信息失败: %v", err)
		return "", "", time.Time{}, 0, false, errno.OAuth2LoginError
	}

	// 查找OAuth账号
	oauthAccount, err := s.oauthDAO.GetByProviderAndID(ctx, provider, userInfo.ID)
	if err != nil {
		hlog.Errorf("获取OAuth账号失败: %v", err)
		return "", "", time.Time{}, 0, false, errno.InternalError
	}

	var userID int64
	var isNewUser bool

	if oauthAccount == nil {
		// 新用户，创建用户
		// 检查邮箱是否已存在
		existUser, err := s.userDAO.GetByEmail(ctx, userInfo.Email)
		if err != nil {
			hlog.Errorf("通过邮箱获取用户失败: %v", err)
			return "", "", time.Time{}, 0, false, errno.InternalError
		}

		if existUser != nil {
			// 邮箱已存在，关联账号
			userID = existUser.ID
		} else {
			// 创建新用户
			user := &model.User{
				Username: userInfo.Name,
				Password: "", // OAuth2用户没有密码
				Email:    userInfo.Email,
				Avatar:   userInfo.AvatarURL,
				Status:   1, // 启用
			}

			userID, err = s.userDAO.Create(ctx, user)
			if err != nil {
				hlog.Errorf("创建用户失败: %v", err)
				return "", "", time.Time{}, 0, false, errno.InternalError
			}
			isNewUser = true
		}

		// 创建OAuth账号
		expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30天
		oauthAccountModel := &model.OAuthAccount{
			UserID:         userID,
			Provider:       provider,
			ProviderUserID: userInfo.ID,
			AccessToken:    accessToken,
			RefreshToken:   "", // 部分提供商没有刷新令牌
			ExpiresAt:      expiresAt,
		}

		if _, err := s.oauthDAO.Create(ctx, oauthAccountModel); err != nil {
			hlog.Errorf("创建OAuth账号失败: %v", err)
			return "", "", time.Time{}, 0, false, errno.InternalError
		}
	} else {
		// 更新OAuth账号
		userID = oauthAccount.UserID
		expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30天
		if err := s.oauthDAO.UpdateToken(ctx, oauthAccount.ID, accessToken, "", expiresAt); err != nil {
			hlog.Errorf("更新OAuth账号失败: %v", err)
			// 不返回错误，继续登录
		}
	}

	// 生成JWT令牌
	expiresIn := s.conf.JWTExpiration
	expiresAt := time.Now().Add(expiresIn)
	token, err := jwt.GenerateToken(s.conf.JWTSecret, userID, expiresIn)
	if err != nil {
		hlog.Errorf("生成令牌失败: %v", err)
		return "", "", time.Time{}, 0, false, errno.InternalError
	}

	// 生成刷新令牌
	refreshExpiresIn := s.conf.RefreshExpiration
	refreshExpiresAt := time.Now().Add(refreshExpiresIn)
	refreshToken, err := jwt.GenerateRefreshToken(s.conf.JWTSecret, userID, refreshExpiresIn)
	if err != nil {
		hlog.Errorf("生成刷新令牌失败: %v", err)
		return "", "", time.Time{}, 0, false, errno.InternalError
	}

	// 存储刷新令牌
	tokenModel := &model.Token{
		UserID:                userID,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshExpiresAt,
	}

	// 删除旧的刷新令牌
	if err := s.tokenDAO.DeleteByUserID(ctx, userID); err != nil {
		hlog.Warnf("删除用户旧令牌失败: %v", err)
	}

	// 创建新的刷新令牌
	if _, err := s.tokenDAO.Create(ctx, tokenModel); err != nil {
		hlog.Errorf("存储刷新令牌失败: %v", err)
		return "", "", time.Time{}, 0, false, errno.InternalError
	}

	return token, refreshToken, expiresAt, userID, isNewUser, nil
}
