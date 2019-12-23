package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
	config "github.com/pramod/auth_service/config"
	models "github.com/pramod/auth_service/models"
	util "github.com/pramod/auth_service/utils"
)

func init() {
	os.Setenv("GO_ENV", "testing")
	os.Setenv("JWT_SECRET", "dsdshbj")
	os.Setenv("JWT_EXPIRE_TIME", "1")
	os.Setenv("FORGOT_PASSWORD_SECRET", "yygygjhbj")
}

func TestLogin(t *testing.T) {
	t.Run("test authenticateAndGetToken ", func(t *testing.T) {
		utilMem = &utilAll{UtilInterface: util.Util{}}
		allMem = &userAllInt{UserInterface: MockUser{}, ConfInterface: config.ZepConf{}}
		userMem = &userAll{UserInterface: MockUser{}}

		ctx := context.Background()

		usrObj, errObj := authenticateAndGetToken(ctx, models.User{Email: "test@gmail.com", Password: "textsecret"})
		fmt.Println(usrObj)
		fmt.Println(errObj)
	})
}

func TestValidateToken(t *testing.T) {
	t.Run("test ValidateToken with no token", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/isValidToken", nil)
		if err != nil {
			t.Fatal(err)
		}
		res := httptest.NewRecorder()
		ValidateToken(res, req)

		if status := res.Code; status != http.StatusForbidden {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusForbidden)
		}
	})

	t.Run("test ValidateToken with valid token", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/isValidToken", nil)
		if err != nil {
			t.Fatal(err)
		}
		token, err := util.Util{}.GetJWT(struct{}{}, "JWT_SECRET")
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("token", token)

		res := httptest.NewRecorder()
		ValidateToken(res, req)

		if status := res.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		expected := util.JsonResponse{true, "authorized"}

		result := util.JsonResponse{}
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			log.Fatalln(err)
		}

		if reflect.DeepEqual(expected, result) != true {
			t.Errorf("handler returned unexpected body: got %v want %v",
				expected, fmt.Sprintf("%+v", result))
		}
	})
}

func TestForgotPassword(t *testing.T) {
	t.Run("test ForgotPassword with non existing user", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/forgotPassword/abc", nil)
		if err != nil {
			t.Fatal(err)
		}

		vars := map[string]string{
			"empId": "def",
		}

		req = mux.SetURLVars(req, vars)

		res := httptest.NewRecorder()
		// handler := http.HandlerFunc(ForgotPassword)
		// handler.ServeHTTP(res, req)
		ForgotPassword(res, req)
		if status := res.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := util.JsonResponse{false, "User not found"}

		result := util.JsonResponse{}
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			log.Fatalln(err)
		}

		if reflect.DeepEqual(expected, result) != true {
			t.Errorf("handler returned unexpected body: got %v want %v",
				expected, fmt.Sprintf("%+v", result))
		}
	})

	t.Run("test ForgotPassword with correct user", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/forgotPassword/abc", nil)
		if err != nil {
			t.Fatal(err)
		}

		vars := map[string]string{
			"empId": "abc",
		}

		req = mux.SetURLVars(req, vars)

		res := httptest.NewRecorder()
		ForgotPassword(res, req)

		if status := res.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := util.JsonResponse{true, "reset link has been sent to your email id"}

		result := util.JsonResponse{}
		if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
			log.Fatalln(err)
		}

		if reflect.DeepEqual(expected, result) != true {
			t.Errorf("handler returned unexpected body: got %v want %v",
				expected, fmt.Sprintf("%+v", result))
		}

		expectedToMail := "abc@gmail.com"

		if reflect.DeepEqual(expectedToMail, mailMem.MailInterface.GetToMail()) != true {
			t.Errorf("handler returned unexpected body: got %v want %v",
				mailMem.MailInterface.GetToMail(), fmt.Sprintf("%+v", expectedToMail))
		}
	})
}
