# 用户管理微服务

基于 Hertz + Kitex + JWT + OAuth2 的完善用户管理微服务系统。

## 功能特性

- 用户管理
  - 用户注册
  - 用户登录
  - 用户信息获取
  - 用户信息更新
  - 密码重置
- 认证授权
  - JWT 令牌认证
  - OAuth2 第三方登录
  - 令牌刷新
- 服务特性
  - 服务注册与发现 (Nacos)
  - 服务监控 (Prometheus)
  - 分布式链路追踪 (Zipkin)
  - 日志集中管理 (ELK)

## 技术栈

- API 网关: [Hertz](https://github.com/cloudwego/hertz)
- RPC 框架: [Kitex](https://github.com/cloudwego/kitex)
- 数据库: MySQL
- 缓存: Redis
- 认证: JWT + OAuth2
- 服务注册与发现: Nacos
- 链路追踪: Zipkin
- 日志系统: ELK (Elasticsearch, Logstash, Kibana)

## 项目结构

```
项目目录
├── api/                    # API 定义目录
│   ├── idl/                # IDL 接口定义文件
│   └── go_gen/             # 生成的 Go 代码
├── cmd/                    # 应用程序入口
│   ├── api/                # API 网关服务入口
│   ├── user/               # 用户服务入口
│   └── auth/               # 认证服务入口
├── config/                 # 配置文件
├── internal/               # 内部包
│   ├── api/                # API 网关实现
│   ├── user/               # 用户服务实现
│   ├── auth/               # 认证服务实现
│   └── pkg/                # 内部公共包
├── pkg/                    # 外部公共包
├── deploy/                 # 部署相关文件
├── scripts/                # 脚本文件
├── docs/                   # 文档
└── migrations/             # 数据库迁移文件
```

## 快速开始

### 环境要求

- Go 1.24+
- MySQL 8.0+
- Redis 6.0+
- Nacos 2.0+
- Zipkin

### 安装依赖

```bash
go mod tidy
```

### 配置

修改 `config/config.yaml` 文件：

```yaml
server:
  name: "user-service"
  host: "0.0.0.0"
  port: 8080
  mode: "debug"  # 开发环境设为 debug，生产环境设为 release

mysql:
  host: "localhost"
  port: 3306
  user: "root"
  password: "password"
  database: "user_service"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

nacos:
  host: "localhost"
  port: 8848
  namespace: "public"
  user: "nacos"
  password: "nacos"
  data_id: "user-service"
  group: "DEFAULT_GROUP"

zipkin:
  endpoint: "http://localhost:9411/api/v2/spans"
  service_name: "user-service"
  sample_rate: 1.0

auth:
  jwt_secret: "your-jwt-secret-key"
  jwt_expiration: "24h"
  refresh_expiration: "720h"
  oauth2_redirect_url: "http://localhost:8080/callback"
  github_client_id: "your-github-client-id"
  github_client_secret: "your-github-client-secret"
  google_client_id: "your-google-client-id"
  google_client_secret: "your-google-client-secret"
```

### 数据库迁移

执行数据库迁移：

```bash
migrate -database "mysql://root:password@tcp(localhost:3306)/user_service" -path migrations up
```

### 运行

启动 API 网关服务：

```bash
go run cmd/api/main.go
```

## API 接口

### 用户认证

| 方法 | 路径 | 描述 |
|------|------|------|
| POST | /api/v1/users/register | 用户注册 |
| POST | /api/v1/users/login | 用户登录 |
| POST | /api/v1/users/oauth2/login | OAuth2 登录 |
| POST | /api/v1/auth/refresh | 刷新令牌 |

### 用户管理

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/v1/users/:user_id | 获取用户信息 |
| PUT | /api/v1/users/:user_id | 更新用户信息 |
| POST | /api/v1/users/:user_id/reset-password | 重置密码 |

## 注意事项

- 所有需要认证的API需要在请求头中添加 `Authorization: Bearer <token>` 
- OAuth2 登录需要先在各平台注册应用并获取相应的 Client ID 和 Client Secret 