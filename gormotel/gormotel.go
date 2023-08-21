package gormotel

import (
	"context"

	"go.opentelemetry.io/otel/codes"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type TracePlugin struct {
}

const (
	gormSpanKey        = "gorm_span"
	callBackBeforeName = "trace:before"
	callBackAfterName  = "trace:after"
)

var _ gorm.Plugin = TracePlugin{}

var tracer = otel.Tracer(gormSpanKey)

func (g TracePlugin) Name() string {
	return "TracePlugin"
}

func (g TracePlugin) Initialize(db *gorm.DB) error {
	// 开始前
	_ = db.Callback().Create().Before("gorm:before_create").Register(callBackBeforeName, before)
	_ = db.Callback().Query().Before("gorm:query").Register(callBackBeforeName, before)
	_ = db.Callback().Delete().Before("gorm:before_delete").Register(callBackBeforeName, before)
	_ = db.Callback().Update().Before("gorm:setup_reflect_value").Register(callBackBeforeName, before)
	_ = db.Callback().Row().Before("gorm:row").Register(callBackBeforeName, before)
	_ = db.Callback().Raw().Before("gorm:raw").Register(callBackBeforeName, before)

	// 结束后
	_ = db.Callback().Create().After("gorm:after_create").Register(callBackAfterName, after)
	_ = db.Callback().Query().After("gorm:after_query").Register(callBackAfterName, after)
	_ = db.Callback().Delete().After("gorm:after_delete").Register(callBackAfterName, after)
	_ = db.Callback().Update().After("gorm:after_update").Register(callBackAfterName, after)
	_ = db.Callback().Row().After("gorm:row").Register(callBackAfterName, after)
	_ = db.Callback().Raw().After("gorm:raw").Register(callBackAfterName, after)
	return nil

	// 开始后
}

func before(db *gorm.DB) {
	ctx := db.Statement.Context
	if !trace.SpanFromContext(ctx).IsRecording() {
		return
	}
	ctx, span := tracer.Start(ctx, "Gorm Sql")
	span.SetAttributes(attribute.String("db.system", "gorm"))
	db.InstanceSet(gormSpanKey, ctx)
}

func after(db *gorm.DB) {
	ctx, ok := db.InstanceGet(gormSpanKey)
	if !ok {
		return
	}

	span := trace.SpanFromContext(ctx.(context.Context))
	defer span.End()

	// sql
	span.SetAttributes(
		attribute.Int64("db.rows", db.RowsAffected),
		attribute.String("db.sql", db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars)),
	)

	// error
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		span.RecordError(db.Error)
		span.SetStatus(codes.Error, db.Error.Error())
	}

	return
}
