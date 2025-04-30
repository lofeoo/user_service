package handler

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"user_service/api/kitex_gen/api"
	userApi "user_service/api/kitex_gen/user"
	"user_service/internal/pkg/errno"
	"user_service/internal/user/service"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register 用户注册
func (h *UserHandler) Register(ctx context.Context, c *app.RequestContext) {
	var req api.RegisterRequest
	if err := c.BindAndValidate(&req); err != nil {
		hlog.Errorf("绑定请求参数失败: %v", err)
		responseError(c, errno.ParamError)
		return
	}

	// 调用服务进行注册
	userID, err := h.userService.Register(ctx, req.Username, req.Password, req.Email, req.Phone)
	if err != nil {
		hlog.Errorf("用户注册失败: %v", err)
		responseError(c, err)
		return
	}

	// 返回注册结果
	responseSuccess(c, utils.H{
		"user_id": userID,
	})
}

// Login 用户登录
func (h *UserHandler) Login(ctx context.Context, c *app.RequestContext) {
	var req api.LoginRequest
	if err := c.BindAndValidate(&req); err != nil {
		hlog.Errorf("绑定请求参数失败: %v", err)
		responseError(c, errno.ParamError)
		return
	}

	// 调用服务进行登录
	userID, token, expiresAt, err := h.userService.Login(ctx, req.Username, req.Password)
	if err != nil {
		hlog.Errorf("用户登录失败: %v", err)
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
		"token":      token,
		"expires_in": expiresAt.Unix() - 0,
		"user_info":  userInfo,
	})
}

// GetUserInfo 获取用户信息
func (h *UserHandler) GetUserInfo(ctx context.Context, c *app.RequestContext) {
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		hlog.Errorf("解析用户ID失败: %v", err)
		responseError(c, errno.ParamError)
		return
	}

	// 检查是否有权限访问该用户信息
	// 从上下文获取当前用户ID
	ctxUserID, exists := c.Get("user_id")
	if !exists {
		responseError(c, errno.UnauthorizedError)
		return
	}

	// 只有本人才能查看自己的信息
	if ctxUserID.(int64) != userID {
		responseError(c, errno.UnauthorizedError)
		return
	}

	// 获取用户信息
	user, err := h.userService.GetUserInfo(ctx, userID)
	if err != nil {
		hlog.Errorf("获取用户信息失败: %v", err)
		responseError(c, err)
		return
	}

	// 构建用户详细信息响应
	userInfo := &userApi.UserInfo{
		Id:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Avatar:    user.Avatar,
		Phone:     user.Phone,
		CreatedAt: user.CreatedAt.Unix(),
		UpdatedAt: user.UpdatedAt.Unix(),
	}

	userDetail := &userApi.UserDetail{
		BaseInfo: userInfo,
		RealName: user.RealName,
		Gender:   user.Gender,
		Birthday: user.Birthday,
		Address:  user.Address,
		Bio:      user.Bio,
	}

	// 返回用户信息
	responseSuccess(c, utils.H{
		"user_detail": userDetail,
	})
}

// UpdateUserInfo 更新用户信息
func (h *UserHandler) UpdateUserInfo(ctx context.Context, c *app.RequestContext) {
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		hlog.Errorf("解析用户ID失败: %v", err)
		responseError(c, errno.ParamError)
		return
	}

	// 检查是否有权限访问该用户信息
	// 从上下文获取当前用户ID
	ctxUserID, exists := c.Get("user_id")
	if !exists {
		responseError(c, errno.UnauthorizedError)
		return
	}

	// 只有本人才能更新自己的信息
	if ctxUserID.(int64) != userID {
		responseError(c, errno.UnauthorizedError)
		return
	}

	var req api.UpdateUserInfoRequest
	if err := c.BindAndValidate(&req); err != nil {
		hlog.Errorf("绑定请求参数失败: %v", err)
		responseError(c, errno.ParamError)
		return
	}

	// 获取原始用户信息
	user, err := h.userService.GetUserInfo(ctx, userID)
	if err != nil {
		hlog.Errorf("获取用户信息失败: %v", err)
		responseError(c, err)
		return
	}

	// 更新用户信息
	if req.Username != nil {
		user.Username = *req.Username
	}
	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}
	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}
	if req.RealName != nil {
		user.RealName = *req.RealName
	}
	if req.Gender != nil {
		user.Gender = *req.Gender
	}
	if req.Birthday != nil {
		user.Birthday = *req.Birthday
	}
	if req.Address != nil {
		user.Address = *req.Address
	}
	if req.Bio != nil {
		user.Bio = *req.Bio
	}

	// 调用服务更新用户信息
	if err := h.userService.UpdateUserInfo(ctx, user); err != nil {
		hlog.Errorf("更新用户信息失败: %v", err)
		responseError(c, err)
		return
	}

	// 返回成功
	responseSuccess(c, nil)
}

// ResetPassword 重置密码
func (h *UserHandler) ResetPassword(ctx context.Context, c *app.RequestContext) {
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		hlog.Errorf("解析用户ID失败: %v", err)
		responseError(c, errno.ParamError)
		return
	}

	// 检查是否有权限访问该用户信息
	// 从上下文获取当前用户ID
	ctxUserID, exists := c.Get("user_id")
	if !exists {
		responseError(c, errno.UnauthorizedError)
		return
	}

	// 只有本人才能重置自己的密码
	if ctxUserID.(int64) != userID {
		responseError(c, errno.UnauthorizedError)
		return
	}

	var req api.ResetPasswordRequest
	if err := c.BindAndValidate(&req); err != nil {
		hlog.Errorf("绑定请求参数失败: %v", err)
		responseError(c, errno.ParamError)
		return
	}

	// 调用服务重置密码
	if err := h.userService.ResetPassword(ctx, userID, req.OldPassword, req.NewPassword_); err != nil {
		hlog.Errorf("重置密码失败: %v", err)
		responseError(c, err)
		return
	}

	// 返回成功
	responseSuccess(c, nil)
}

// 响应错误
func responseError(c *app.RequestContext, err error) {
	var e *errno.Errno
	if t, ok := err.(*errno.Errno); ok {
		e = t
	} else if t, ok := err.(errno.Errno); ok {
		e = &t
	} else {
		e = &errno.InternalError
	}

	c.JSON(consts.StatusOK, &api.BaseResponse{
		Code:    int32(e.Code),
		Message: e.Message,
	})
}

// 响应成功
func responseSuccess(c *app.RequestContext, data interface{}) {
	resp := &api.BaseResponse{
		Code:    int32(errno.Success.Code),
		Message: errno.Success.Message,
	}

	if data != nil {
		// 将data转换为map
		if m, ok := data.(utils.H); ok {
			// 将 utils.H 类型转换为 map[string]string 类型
			strMap := make(map[string]string)
			for k, v := range m {
				strMap[k] = fmt.Sprint(v)
			}
			resp.Data = strMap
		}
	}

	c.JSON(consts.StatusOK, resp)
}
