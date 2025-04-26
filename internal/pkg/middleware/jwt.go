package middleware

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"user_service/internal/auth/service"
	"user_service/internal/pkg/errno"
)

// JWT中间件
func JWT(authService *service.AuthService) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 从请求头中获取Token
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			// 尝试从查询参数获取Token
			token = c.Query("token")
		}

		// 检查Token是否存在
		if token == "" {
			// Token不存在
			responseError(c, errno.UnauthorizedError)
			c.Abort()
			return
		}

		// 如果Token是Bearer类型，去掉Bearer前缀
		if strings.HasPrefix(token, "Bearer ") {
			token = token[7:]
		}

		// 验证Token
		valid, userID, err := authService.ValidateToken(ctx, token)
		if err != nil {
			hlog.Errorf("验证Token失败: %v", err)
			if err == errno.TokenExpiredError {
				responseError(c, errno.TokenExpiredError)
			} else {
				responseError(c, errno.TokenInvalidError)
			}
			c.Abort()
			return
		}

		if !valid {
			responseError(c, errno.TokenInvalidError)
			c.Abort()
			return
		}

		// 将用户ID存储到上下文中
		c.Set("user_id", userID)

		// 继续处理请求
		c.Next(ctx)
	}
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

	c.JSON(consts.StatusOK, map[string]interface{}{
		"code":    e.Code,
		"message": e.Message,
	})
}
