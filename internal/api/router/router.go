package router

import (
	"user_service/internal/api/handler"
	authService "user_service/internal/auth/service"
	"user_service/internal/pkg/middleware"

	"github.com/cloudwego/hertz/pkg/app/middlewares/server/recovery"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// Register 注册路由
func Register(h *server.Hertz, userHandler *handler.UserHandler, authHandler *handler.AuthHandler, authSvc *authService.AuthService) {
	// 注册全局中间件
	// h.Use(middlewares.Recovery()) // 使用新的 recover 中间件
	h.Use(recovery.Recovery()) // 错误恢复中间件

	// API组
	api := h.Group("/api/v1")

	// 公开路由组
	{
		// 用户注册
		api.POST("/users/register", userHandler.Register)

		// 用户登录
		api.POST("/users/login", userHandler.Login)

		// OAuth2登录
		api.POST("/users/oauth2/login", authHandler.OAuth2Login)

		// 刷新令牌
		api.POST("/auth/refresh", authHandler.RefreshToken)
	}

	// 需要认证的路由组
	auth := api.Group("/")
	auth.Use(middleware.JWT(authSvc))
	{
		// 获取用户信息
		auth.GET("/users/:user_id", userHandler.GetUserInfo)

		// 更新用户信息
		auth.PUT("/users/:user_id", userHandler.UpdateUserInfo)

		// 重置密码
		auth.POST("/users/:user_id/reset-password", userHandler.ResetPassword)
	}
}
