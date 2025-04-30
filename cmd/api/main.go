package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/kitex/pkg/utils"
	"github.com/hertz-contrib/registry/nacos"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"

	"user_service/internal/api/handler"
	"user_service/internal/api/router"
	"user_service/internal/auth/service"
	"user_service/internal/pkg/conf"
	"user_service/internal/pkg/db"
	"user_service/internal/pkg/redis"
	"user_service/internal/pkg/tracer"
	"user_service/internal/user/dao"
	userService "user_service/internal/user/service"
)

func main() {
	// 1. 初始化配置
	config, err := conf.Init("config\\config.yaml")
	if err != nil {
		panic(fmt.Sprintf("初始化配置失败: %v", err))
	}

	// 2. 初始化MySQL
	if _, err := db.InitMySQL(config.MySQL); err != nil {
		panic(fmt.Sprintf("初始化MySQL失败: %v", err))
	}

	// 3. 初始化Redis
	if _, err := redis.InitRedis(config.Redis); err != nil {
		panic(fmt.Sprintf("初始化Redis失败: %v", err))
	}

	// 4. 初始化链路追踪
	cleanup, err := tracer.InitZipkinTracer(config.Zipkin)
	if err != nil {
		hlog.Warnf("初始化Zipkin失败: %v", err)
	} else {
		defer cleanup()
	}

	// 5. 初始化Nacos服务注册
	// 服务注册中心配置
	sc := []constant.ServerConfig{
		{
			IpAddr: config.Nacos.Host,
			Port:   config.Nacos.Port,
		},
	}

	// 客户端配置
	cc := constant.ClientConfig{
		NamespaceId:         config.Nacos.Namespace,
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "log/nacos/log",
		CacheDir:            "log/nacos/cache",
		Username:            config.Nacos.User,
		Password:            config.Nacos.Password,
	}

	// 创建Nacos客户端
	nacosClient, err := clients.NewNamingClient(vo.NacosClientParam{
		ClientConfig:  &cc,
		ServerConfigs: sc,
	})
	if err != nil {
		panic(fmt.Sprintf("创建Nacos客户端失败: %v", err))
	}

	// 创建注册器
	r := nacos.NewNacosRegistry(nacosClient, nacos.WithRegistryGroup(config.Nacos.Group))
	// r, _ := nacos.NewDefaultNacosRegistry()
	// 创建Nacos解析器
	// resolver := nacosResolver.NewNacosResolver(
	// 	nacosClient,
	// 	nacosResolver.WithCluster("DEFAULT"),
	// )
	// _ = resolver // 使用resolver避免未使用警告

	// 6. 初始化Service
	// 用户DAO
	userDAO := dao.NewUserDAO()
	oauthDAO := dao.NewOAuthDAO()
	tokenDAO := dao.NewTokenDAO()

	// 用户服务
	userSvc := userService.NewUserService(userDAO, oauthDAO, tokenDAO, config.Auth)

	// 认证服务
	authSvc := service.NewAuthService(userDAO, oauthDAO, tokenDAO, config.Auth)

	// 7. 初始化Handler
	userHandler := handler.NewUserHandler(userSvc)
	authHandler := handler.NewAuthHandler(authSvc, userSvc)

	// 8. 创建Hertz服务器
	addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	h := server.Default(
		server.WithHostPorts(addr),
		server.WithRegistry(r, &registry.Info{
			ServiceName: config.Server.Name,
			Addr:        utils.NewNetAddr("tcp", addr),
			Weight:      10,
		}),
	)

	// h := server.Default(server.WithHostPorts(addr))

	// h := server.Default(
	// 	server.WithHostPorts(addr),
	// 	server.WithRegistry(r, &registry.Info{
	// 		ServiceName: config.Server.Name,
	// 		Addr:        utils.NewNetAddr("tcp", addr),
	// 		Weight:      10,
	// 		Tags: map[string]string{
	// 			rpcinfo.RPCRole: "api-gateway",
	// 		},
	// 	}),
	// )

	// 9. 注册路由
	router.Register(h, userHandler, authHandler, authSvc)

	// 10. 启动服务
	hlog.Infof("API网关服务启动，监听地址: %s", addr)

	// 优雅关闭
	go func() {
		h.Spin()
	}()

	// 等待中断信号，优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	hlog.Info("正在关闭服务器...")

	if err := h.Shutdown(context.Background()); err != nil {
		hlog.Errorf("关闭服务器出错: %v", err)
	}

	hlog.Info("服务器已安全关闭")
}
