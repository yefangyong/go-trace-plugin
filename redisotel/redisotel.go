package redisotel

import (
	"context"

	"go.opentelemetry.io/otel/codes"

	"github.com/go-redis/redis/extra/rediscmd"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("go-redis/redis")
var _ redis.Hook = &TracingHook{}

type TracingHook struct{}

func (t TracingHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	if !trace.SpanFromContext(ctx).IsRecording() {
		return ctx, nil
	}
	ctx, span := tracer.Start(ctx, "REDIS "+cmd.FullName())
	span.SetAttributes(
		attribute.String("db.system", "redis"),
		attribute.String("db.statement", rediscmd.CmdString(cmd)),
	)
	return ctx, nil
}

func (t TracingHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	span := trace.SpanFromContext(ctx)
	if cmd.Err() != nil {
		recordError(ctx, span, cmd.Err())
	}
	span.End()
	return nil
}

func (t TracingHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	if !trace.SpanFromContext(ctx).IsRecording() {
		return ctx, nil
	}
	summary, cmdsString := rediscmd.CmdsString(cmds)

	ctx, span := tracer.Start(ctx, "REDIS Pipeline"+summary)
	span.SetAttributes(attribute.String("db.system", "redis"), attribute.Int("db.redis.cmd.num", len(cmds)), attribute.String("db.statement", cmdsString))
	return ctx, nil
}

func (t TracingHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	span := trace.SpanFromContext(ctx)
	if cmds[0].Err() != nil {
		recordError(ctx, span, cmds[0].Err())
	}
	span.End()
	return nil
}

func recordError(ctx context.Context, span trace.Span, err error) {
	if err != redis.Nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}
