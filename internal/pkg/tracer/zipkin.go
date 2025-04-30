package tracer

import (
	"context"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/kitex-contrib/obs-opentelemetry/provider"
	"github.com/kitex-contrib/obs-opentelemetry/tracing"

	"user_service/internal/pkg/conf"
)

// InitZipkinTracer 初始化Zipkin追踪器
func InitZipkinTracer(config *conf.ZipkinConfig) (func(), error) {
	// 创建一个provider.NewOpenTelemetryProvider用于初始化OpenTelemetry
	p := provider.NewOpenTelemetryProvider(
		provider.WithServiceName(config.ServiceName),
		provider.WithExportEndpoint(config.Endpoint),
		provider.WithInsecure(), // 使用不安全的连接模式，因为WithSampleRate已不再支持
	)

	// 添加Hertz中间件
	tracing.NewServerSuite() // 创建新的Hertz服务端链路追踪Provider

	// 设置Kitex添加链路追踪
	tracing.NewClientSuite() // 创建新的Kitex客户端链路追踪Provider

	hlog.Infof("Zipkin链路追踪初始化成功: %s", config.Endpoint)
	klog.Infof("Zipkin链路追踪初始化成功: %s", config.Endpoint)

	return func() {
		// 使用空的context调用Shutdown
		if err := p.Shutdown(context.Background()); err != nil {
			hlog.Errorf("关闭Zipkin追踪器失败: %v", err)
		}
	}, nil
}
