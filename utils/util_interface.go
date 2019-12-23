package utils

import "net/http"

type UtilInterface interface {
	SendErrorResponse(res http.ResponseWriter, msg string)
	GetJWT(data interface{}, secret_type string) (string, error)
	IsValidToken(tokenString, secretType string) (map[string]interface{}, error)
}

type UtilMembers struct {
	UtilInterface
}

type MailInterface interface {
	AssignDefault()
	SendMail() error
	GetToMail() string
}

type MailMembers struct {
	MailInterface
}
