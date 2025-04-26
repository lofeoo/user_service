namespace go api

include "user.thrift"
include "auth.thrift"

// API响应基础结构
struct BaseResponse {
    1: i32 code;
    2: string message;
    3: optional map<string, string> data;
}

// 用户注册请求
struct RegisterRequest {
    1: string username (api.form="username", api.vd="len($) > 3 && len($) < 24");
    2: string password (api.form="password", api.vd="len($) > 6 && len($) < 32");
    3: string email (api.form="email", api.vd="email($)");
    4: string phone (api.form="phone", api.vd="len($) > 0");
}

// 用户登录请求
struct LoginRequest {
    1: string username (api.form="username", api.vd="len($) > 0");
    2: string password (api.form="password", api.vd="len($) > 0");
}

// OAuth2登录请求
struct OAuth2LoginRequest {
    1: string provider (api.form="provider", api.vd="len($) > 0");
    2: string code (api.form="code", api.vd="len($) > 0");
    3: string redirect_uri (api.form="redirect_uri");
}

// 获取用户信息请求
struct GetUserInfoRequest {
    1: i64 user_id (api.path="user_id", api.vd="$ > 0");
}

// 更新用户信息请求
struct UpdateUserInfoRequest {
    1: i64 user_id (api.path="user_id", api.vd="$ > 0");
    2: optional string username (api.form="username");
    3: optional string email (api.form="email");
    4: optional string phone (api.form="phone");
    5: optional string avatar (api.form="avatar");
    6: optional string real_name (api.form="real_name");
    7: optional string gender (api.form="gender");
    8: optional string birthday (api.form="birthday");
    9: optional string address (api.form="address");
    10: optional string bio (api.form="bio");
}

// 重置密码请求
struct ResetPasswordRequest {
    1: i64 user_id (api.path="user_id", api.vd="$ > 0");
    2: string old_password (api.form="old_password", api.vd="len($) > 0");
    3: string new_password (api.form="new_password", api.vd="len($) > 6 && len($) < 32");
}

// 刷新Token请求
struct RefreshTokenRequest {
    1: string refresh_token (api.form="refresh_token", api.vd="len($) > 0");
}

// API服务
service ApiService {
    // 用户相关API
    BaseResponse Register(1: RegisterRequest req) (api.post="/api/v1/users/register");
    BaseResponse Login(1: LoginRequest req) (api.post="/api/v1/users/login");
    BaseResponse OAuth2Login(1: OAuth2LoginRequest req) (api.post="/api/v1/users/oauth2/login");
    BaseResponse GetUserInfo(1: GetUserInfoRequest req) (api.get="/api/v1/users/:user_id");
    BaseResponse UpdateUserInfo(1: UpdateUserInfoRequest req) (api.put="/api/v1/users/:user_id");
    BaseResponse ResetPassword(1: ResetPasswordRequest req) (api.post="/api/v1/users/:user_id/reset-password");
    
    // 认证相关API
    BaseResponse RefreshToken(1: RefreshTokenRequest req) (api.post="/api/v1/auth/refresh");
} 