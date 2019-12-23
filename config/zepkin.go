package config

import (
	"context"
	"log"
	"os"

	"github.com/openzipkin/zipkin-go"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHttpMiddleware "github.com/openzipkin/zipkin-go/middleware/http"
	"github.com/openzipkin/zipkin-go/reporter"
	zipkinHttpReporter "github.com/openzipkin/zipkin-go/reporter/http"
	util "github.com/pramod/auth_service/utils"
	//logreporter "github.com/openzipkin/zipkin-go/reporter/log"
)

var tracer *zipkin.Tracer

var zipkinClient *zipkinHttpMiddleware.Client

type ZepConf struct {
}

func (c ZepConf) RegisterZipkinTacer() (*zipkin.Tracer, reporter.Reporter, error) {
		//reporterTimeoutOption := zipkinHttpReporter.Timeout(time.Duration(5 * time.Second))
	//reporterRetryOption := zipkinHttpReporter.MaxRetry(10)
		// reporterBatchOption := zipkinHttpReporter.BatchInterval(time.Duration(5 * time.Second))
	//reporter := zipkinHttpReporter.NewReporter(os.Getenv("ZIPKIN_REPORTER"), reporterRetryOption)
	reporter := zipkinHttpReporter.NewReporter(os.Getenv("ZIPKIN_REPORTER"))
		//reporter := logreporter.NewReporter(log.New(os.Stderr, "", log.LstdFlags))
		// defer reporter.Close()

	// If you are hitting any other service from this service then specify service host
	zipkinEndpoint, err := openzipkin.NewEndpoint(os.Getenv("ZIPKIN_SERVICE_NAME"), os.Getenv("ZIPKIN_ENDPOINT"))
	if err != nil {
		log.Fatalf("Failed to create local endpoint: %+v\n", err)
		return nil, nil, err
	}

	//tracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(zipkinEndpoint))
	sampler, err := zipkin.NewCountingSampler(1)
	if err != nil {
		return nil, nil, err
	}

	tracer, err = zipkin.NewTracer(
		reporter,
		zipkin.WithSampler(sampler),
		zipkin.WithLocalEndpoint(zipkinEndpoint),
	)
	if err != nil {
		log.Fatalf("unable to create tracer: %+v\n", err)
		return nil, nil, err
	}

	return tracer, reporter, nil
}

func (c ZepConf) RegisterZipkinClient() (*zipkinHttpMiddleware.Client, error) {
	var err error
	zipkinClient, err = zipkinHttpMiddleware.NewClient(tracer, zipkinHttpMiddleware.ClientTrace(true))
	if err != nil {
		log.Fatalf("unable to create client: %+v\n", err)
	}
	return zipkinClient, err
}

func (c ZepConf) GetTracer() *zipkin.Tracer {
	return tracer
}

func (c ZepConf) GetZipKinClient() *zipkinHttpMiddleware.Client {
	return zipkinClient
}

//"http://"+os.Getenv("ZIPKIN_HOST")+":"+os.Getenv("ZIPKIN_PORT")+"/api/v2/spans"

func (c ZepConf) CustomStartSpanFromContext(parentCtx context.Context, spanName string) (openzipkin.Span, context.Context) {
	var span openzipkin.Span
	var ctx context.Context
	if util.IsZeroValue(tracer) != true {
		span, ctx = tracer.StartSpanFromContext(parentCtx, spanName)
	}
	return span, ctx
}

func (c ZepConf) CustomSpanFinish(span openzipkin.Span) {
	if util.IsZeroValue(tracer) != true {
		if util.IsZeroValue(span) != true {
			span.Finish()
		}
	}
}

func CustomStartSpan(spanName string) openzipkin.Span {
	var span openzipkin.Span
	if util.IsZeroValue(tracer) != true {
		span = tracer.StartSpan(spanName)
	}
	return span
}
