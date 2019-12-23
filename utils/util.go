package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type JsonResponse struct {
	Success bool
	Msg     string
}

type CustomError struct {
	Err        error
	StatusCode int
}

type Util struct {
}

func (u Util) SendErrorResponse(res http.ResponseWriter, msg string) {
	json.NewEncoder(res).Encode(JsonResponse{false, msg})
}

//return true if zero value
func IsZeroValue(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

func GetProtocol(req *http.Request) string {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	return scheme
}

func (u Util) GetJWT(data interface{}, secret_type string) (string, error) {

	if os.Getenv(secret_type) == "" {
		return "", errors.New("secret is empty or not defined")
	}

	var dataMap map[string]interface{}
	userMap, _ := json.Marshal(data)
	json.Unmarshal(userMap, &dataMap)

	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)

	for k := range dataMap {
		v := reflect.ValueOf(dataMap[k])

		switch v.Kind() {
		case reflect.String:
			claims[k] = v.String()
		case reflect.Int:
			claims[k] = v.Int()
		}
	}
	expTime, err := strconv.ParseInt(os.Getenv("JWT_EXPIRE_TIME"), 10, 32)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	// fmt.Println(time.Duration(expTime))
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(expTime)).Unix()
	claims["iat"] = time.Now().Unix()
	token.Claims = claims

	return token.SignedString([]byte(os.Getenv(secret_type)))
}

func (u Util) IsValidToken(tokenString, secretType string) (map[string]interface{}, error) {

	if IsZeroValue(tokenString) == true {
		fmt.Println("token missing")
		return make(map[string]interface{}), errors.New("token missing")
	}

	secret := os.Getenv(secretType)

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		fmt.Println(err)
		return make(map[string]interface{}), err
	}
	return claims, nil
}

func CheckExpectedActualJSON(actual *bytes.Buffer, expected JsonResponse, t *testing.T) {
	result := JsonResponse{}
	if err := json.NewDecoder(actual).Decode(&result); err != nil {
		log.Fatalln(err)
	}

	if reflect.DeepEqual(expected, result) != true {
		t.Errorf("handler returned unexpected body: got %v want %v",
			expected, fmt.Sprintf("%+v", result))
	}
}

type App_Mail struct {
	Subject          string
	To               string
	ToName           string
	From             string
	FromName         string
	PlainTextContent string
	HtmlContent      string
}

func (am *App_Mail) AssignDefault() {
	//func (am *App_Mail) AssignDefault() {
	am.From = os.Getenv("APP_SENDER_MAIL_ID")
	am.FromName = os.Getenv("APP_SENDER_MAIL_ID")
}

func (am *App_Mail) SendMail() error {
	// func (am *App_Mail) SendMail() error {
	fmt.Println(am)
	from := mail.NewEmail(am.FromName, am.From)
	subject := am.Subject
	to := mail.NewEmail(am.ToName, am.To)
	plainTextContent := am.PlainTextContent
	htmlContent := am.HtmlContent
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))

	response, err := client.Send(message)
	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Println(response.Body)
	fmt.Println(response.Headers)
	fmt.Println(response.StatusCode)
	return nil
}

func (am *App_Mail) GetToMail() string {
	return am.To
}
