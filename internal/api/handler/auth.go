package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"

	"user_service/api/kitex_gen/api"
	userApi "user_service/api/kitex_gen/user"
	"user_service/internal/auth/service"
	"user_service/internal/pkg/errno"
	userService "user_service/internal/user/service"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
	userService *userService.UserService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *service.AuthService, userService *userService.UserService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
	}
}

// OAuth2Login OAuth2登录
func (h *AuthHandler) OAuth2Login(ctx context.Context, c *app.RequestContext) {
	var req api.OAuth2LoginRequest
	if err := c.BindAndValidate(&req); err != nil {
		hlog.Errorf("绑定请求参数失败: %v", err)
		responseError(c, errno.ParamError)
		return
	}

	// 调用服务进行OAuth2登录
	token, refreshToken, expiresAt, userID, isNewUser, err := h.authService.OAuth2Login(ctx, req.Provider, req.Code, req.RedirectUri)
	if err != nil {
		hlog.Errorf("OAuth2登录失败: %v", err)
		responseError(c, err)
		return
	}

	// 获取用户信息
	user, err := h.userService.GetUserInfo(ctx, userID)
	if err != nil {
		hlog.Errorf("获取用户信息失败: %v", err)
		responseError(c, err)
		return
	}

	// 构建用户信息响应
	userInfo := &userApi.UserInfo{
		Id:        userID,
		Username:  user.Username,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Phone:     user.Phone,
		CreatedAt: user.CreatedAt.Unix(),
		UpdatedAt: user.UpdatedAt.Unix(),
	}

	// 返回登录结果
	responseSuccess(c, utils.H{
		"token":         token,
		"refresh_token": refreshToken,
		"expires_in":    expiresAt.Unix(), // 移除 - 0
		"user_info":     userInfo,
		"is_new_user":   isNewUser,
	})
}

// RefreshToken 刷新令牌
func (h *AuthHandler) RefreshToken(ctx context.Context, c *app.RequestContext) {
	var req api.RefreshTokenRequest
	if err := c.BindAndValidate(&req); err != nil {
		hlog.Errorf("绑定请求参数失败: %v", err)
		responseError(c, errno.ParamError)
		return
	}

	// 调用服务刷新令牌
	token, refreshToken, expiresAt, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		hlog.Errorf("刷新令牌失败: %v", err)
		responseError(c, err)
		return
	}

	// 返回结果
	responseSuccess(c, utils.H{
		"token":         token,
		"refresh_token": refreshToken,
		"expires_in":    expiresAt.Unix(), // 移除 - 0
	})
}
