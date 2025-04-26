package tracer

import (
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
		provider.WithSampleRate(config.SampleRate),
	)

	// 添加Hertz中间件
	tracing.SetHertzServerTrace()

	// 设置Kitex添加链路追踪
	tracing.SetKitexClientTrace()

	hlog.Infof("Zipkin链路追踪初始化成功: %s", config.Endpoint)
	klog.Infof("Zipkin链路追踪初始化成功: %s", config.Endpoint)

	return p.Shutdown, nil
}
