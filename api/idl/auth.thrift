namespace go auth

// Token信息
struct TokenInfo {
    1: string access_token;
    2: string refresh_token;
    3: i64 expires_in;
    4: string token_type;
}

// 验证Token请求
struct ValidateTokenRequest {
    1: string token (api.query="token", api.vd="len($) > 0");
}

// 验证Token响应
struct ValidateTokenResponse {
    1: i32 code;
    2: string message;
    3: bool is_valid;
    4: i64 user_id;
}

// 刷新Token请求
struct RefreshTokenRequest {
    1: string refresh_token (api.query="refresh_token", api.vd="len($) > 0");
}

// 刷新Token响应
struct RefreshTokenResponse {
    1: i32 code;
    2: string message;
    3: TokenInfo token_info;
}

// OAuth2登录请求
struct OAuth2LoginRequest {
    1: string provider (api.query="provider", api.vd="len($) > 0"); // 'github', 'google', etc.
    2: string code (api.query="code", api.vd="len($) > 0");
    3: string redirect_uri (api.query="redirect_uri");
}

// OAuth2登录响应
struct OAuth2LoginResponse {
    1: i32 code;
    2: string message;
    3: TokenInfo token_info;
    4: i64 user_id;
    5: bool is_new_user;
}

// 授权服务
service AuthService {
    ValidateTokenResponse ValidateToken(1: ValidateTokenRequest req);
    RefreshTokenResponse RefreshToken(1: RefreshTokenRequest req);
    OAuth2LoginResponse OAuth2Login(1: OAuth2LoginRequest req);
} 