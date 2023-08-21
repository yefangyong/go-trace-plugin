package tracer

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"

	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/zipkin"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

const (
	ProviderJaeger   = "jaeger"
	ProviderZipkin   = "zipkin"
	ProviderOtlpHttp = "otlp-http"
)

// Conf struct
type Conf struct {
	Id       string // service id
	Name     string // service name
	Env      string // environment
	Endpoint string // service address
}

// NewConf init conf
func NewConf(id, name, env, endpoint string) *Conf {
	return &Conf{
		Id:       id,
		Name:     name,
		Env:      env,
		Endpoint: endpoint,
	}
}

// Option struct
type Option struct {
	provider string  // provider type
	sampling float64 // sampling rate
}

type OptionFunc func(*Option)

// WithProvider set provider type
func (c *Conf) WithProvider(provider string) OptionFunc {
	return func(o *Option) {
		o.provider = provider
	}
}

// WithSampling set sampling rate
func (c *Conf) WithSampling(sampling float64) OptionFunc {
	return func(o *Option) {
		o.sampling = sampling
	}
}

// TraceProvider create provider
func (c *Conf) TraceProvider(options ...OptionFunc) *tracesdk.TracerProvider {
	option := &Option{
		provider: ProviderJaeger,
		sampling: 1,
	}

	for _, o := range options {
		o(option)
	}

	exp, err := c.CreateTracerProvider(option.provider, c.Endpoint)
	if err != nil {
		panic(fmt.Sprintf("create tracer provider error：%s", err.Error()))
	}

	tp := tracesdk.NewTracerProvider(
		// 将基于父span的采样率设置为100%
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(1.0))),
		// 始终确保再生成中批量处理
		tracesdk.WithBatcher(exp),
		// 在资源中记录有关此应用程序的信息
		tracesdk.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String(c.Name),
		)),
	)

	return tp
}

// CreateTracerProvider create provider
func (c *Conf) CreateTracerProvider(provider, url string) (tracesdk.SpanExporter, error) {
	switch provider {
	case ProviderJaeger:
		// jaeger
		return jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	case ProviderZipkin:
		return zipkin.New(url)
		// zipkin
	case ProviderOtlpHttp:
		ctx := context.Background()
		client := otlptracehttp.NewClient(otlptracehttp.WithEndpoint(url), otlptracehttp.WithInsecure(), otlptracehttp.WithCompression(1))
		return otlptrace.New(ctx, client)
		// otlp-http
	default:
		// default
		return nil, fmt.Errorf("not support provider: %s", provider)
	}
}
