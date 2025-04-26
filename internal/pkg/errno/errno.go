package errno

import "fmt"

// Errno 定义错误类型
type Errno struct {
	Code    int
	Message string
}

func (e Errno) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

// 定义错误码常量
var (
	// 通用错误
	Success           = Errno{Code: 0, Message: "成功"}
	InternalError     = Errno{Code: 10001, Message: "内部服务错误"}
	ParamError        = Errno{Code: 10002, Message: "参数错误"}
	UnauthorizedError = Errno{Code: 10003, Message: "未授权"}
	NotFoundError     = Errno{Code: 10004, Message: "资源不存在"}

	// 用户相关错误
	UserExistError    = Errno{Code: 20001, Message: "用户已存在"}
	UserNotExistError = Errno{Code: 20002, Message: "用户不存在"}
	PasswordError     = Errno{Code: 20003, Message: "密码错误"}

	// 认证相关错误
	TokenInvalidError = Errno{Code: 30001, Message: "无效的令牌"}
	TokenExpiredError = Errno{Code: 30002, Message: "令牌已过期"}
	RefreshTokenError = Errno{Code: 30003, Message: "刷新令牌失败"}
	OAuth2LoginError  = Errno{Code: 30004, Message: "OAuth2登录失败"}
)

// New 创建自定义错误
func New(errno *Errno, msg string) *Errno {
	return &Errno{Code: errno.Code, Message: msg}
}
