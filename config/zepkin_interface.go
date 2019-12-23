package config

import (
	"context"

	"github.com/openzipkin/zipkin-go"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHttpMiddleware "github.com/openzipkin/zipkin-go/middleware/http"
	"github.com/openzipkin/zipkin-go/reporter"
)

type ZepkinInterface interface {
	RegisterZipkinTacer() (*zipkin.Tracer, reporter.Reporter, error)
	RegisterZipkinClient() (*zipkinHttpMiddleware.Client, error)
	GetTracer() *zipkin.Tracer
	GetZipKinClient() *zipkinHttpMiddleware.Client
	CustomStartSpanFromContext(parentCtx context.Context, spanName string) (openzipkin.Span, context.Context)
	CustomSpanFinish(span openzipkin.Span)
}

type ZepkinMembers struct {
	ZepkinInterface
}
