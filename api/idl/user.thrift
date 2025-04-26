namespace go user

// 用户基本信息
struct UserInfo {
    1: i64 id;
    2: string username;
    3: string email;
    4: string avatar;
    5: string phone;
    6: i64 created_at;
    7: i64 updated_at;
}

// 用户详细信息
struct UserDetail {
    1: UserInfo base_info;
    2: string real_name;
    3: string gender;
    4: string birthday;
    5: string address;
    6: string bio;
}

// 注册请求
struct RegisterRequest {
    1: string username (api.query="username", api.vd="len($) > 3 && len($) < 24");
    2: string password (api.query="password", api.vd="len($) > 6 && len($) < 32");
    3: string email (api.query="email", api.vd="email($)");
    4: string phone (api.query="phone", api.vd="len($) > 0");
}

// 注册响应
struct RegisterResponse {
    1: i32 code;
    2: string message;
    3: i64 user_id;
}

// 登录请求
struct LoginRequest {
    1: string username (api.query="username", api.vd="len($) > 0");
    2: string password (api.query="password", api.vd="len($) > 0");
}

// 登录响应
struct LoginResponse {
    1: i32 code;
    2: string message;
    3: string token;
    4: i64 expires_in;
    5: UserInfo user_info;
}

// 获取用户信息请求
struct GetUserInfoRequest {
    1: i64 user_id (api.query="user_id", api.vd="$ > 0");
}

// 获取用户信息响应
struct GetUserInfoResponse {
    1: i32 code;
    2: string message;
    3: UserDetail user_detail;
}

// 更新用户信息请求
struct UpdateUserInfoRequest {
    1: i64 user_id (api.query="user_id", api.vd="$ > 0");
    2: optional string username;
    3: optional string email;
    4: optional string phone;
    5: optional string avatar;
    6: optional string real_name;
    7: optional string gender;
    8: optional string birthday;
    9: optional string address;
    10: optional string bio;
}

// 更新用户信息响应
struct UpdateUserInfoResponse {
    1: i32 code;
    2: string message;
}

// 重置密码请求
struct ResetPasswordRequest {
    1: i64 user_id (api.query="user_id", api.vd="$ > 0");
    2: string old_password (api.query="old_password", api.vd="len($) > 0");
    3: string new_password (api.query="new_password", api.vd="len($) > 6 && len($) < 32");
}

// 重置密码响应
struct ResetPasswordResponse {
    1: i32 code;
    2: string message;
}

// 用户服务
service UserService {
    RegisterResponse Register(1: RegisterRequest req);
    LoginResponse Login(1: LoginRequest req);
    GetUserInfoResponse GetUserInfo(1: GetUserInfoRequest req);
    UpdateUserInfoResponse UpdateUserInfo(1: UpdateUserInfoRequest req);
    ResetPasswordResponse ResetPassword(1: ResetPasswordRequest req);
} 