package inter

import (
	"context"

	"github.com/openzipkin/zipkin-go"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHttpMiddleware "github.com/openzipkin/zipkin-go/middleware/http"
	"github.com/openzipkin/zipkin-go/reporter"
	models "github.com/pramod/auth_service/models"
)

type ConfInterface interface {
	RegisterZipkinTacer() (*zipkin.Tracer, reporter.Reporter, error)
	RegisterZipkinClient() (*zipkinHttpMiddleware.Client, error)
	GetTracer() *zipkin.Tracer
	GetZipKinClient() *zipkinHttpMiddleware.Client
	CustomStartSpanFromContext(parentCtx context.Context, spanName string) (openzipkin.Span, context.Context)
	CustomSpanFinish(span openzipkin.Span)
}

type UserInterface interface {
	CreateUserTable() (bool, error)
	ExtractUserFromInterface(userInter map[string]interface{}) (models.User, error)
	CreateUser() (bool, error)
	UpdateUser() (bool, error)
	GetUser(match map[string]interface{}) ([]models.User, error)
	ModifyPassword(empId, password string) (bool, error)
}

type MailInterface interface {
}

type AllMembers struct {
	ConfInterface
	UserInterface
	MailInterface
}
