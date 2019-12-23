package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/pramod/auth_service/models"

	"github.com/gorilla/context"
	util "github.com/pramod/auth_service/utils"
)

type MockToken struct {
}

func (t MockToken) CreateTokenTable() (bool, error) {
	return true, nil
}

func (t MockToken) InsertUsedToken(tokenString string) (bool, error) {
	return true, nil
}

func (t MockToken) IsTokenUsed(tokenString string) (bool, error) {
	return false, nil
}
func TestAddUser(t *testing.T) {
	t.Run("AddUser missing parameters", func(t *testing.T) {
		payload, err := json.Marshal(MockUser{EmpID: "123", Name: "abc", Email: "abc@gmail.com", Password: "secret"})
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest("POST", "/api/AddUser", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatal(err)
		}

		res := httptest.NewRecorder()
		AddUser(res, req)
		if status := res.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		expected := util.JsonResponse{false, "missing parameters"}
		util.CheckExpectedActualJSON(res.Body, expected, t)
	})

	t.Run("AddUser with all parameters", func(t *testing.T) {
		payload, err := json.Marshal(MockUser{EmpID: "123", Name: "abc", Email: "abc@gmail.com", Password: "secret", Role: "admin"})
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest("POST", "/api/AddUser", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatal(err)
		}

		res := httptest.NewRecorder()
		AddUser(res, req)
		if status := res.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		expected := util.JsonResponse{true, "user created"}
		util.CheckExpectedActualJSON(res.Body, expected, t)
	})
}

func TestEditUser(t *testing.T) {
	t.Run("EditUser missing body", func(t *testing.T) {
		payload, err := json.Marshal(struct{}{})
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest("PUT", "/api/editUser", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatal(err)
		}

		res := httptest.NewRecorder()
		EditUser(res, req)
		if status := res.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		expected := util.JsonResponse{false, "empty body or empId missing"}
		util.CheckExpectedActualJSON(res.Body, expected, t)
	})

	t.Run("EditUser change password", func(t *testing.T) {
		payload, err := json.Marshal(MockUser{EmpID: "123", Password: "sdsds"})
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest("PUT", "/api/editUser", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatal(err)
		}

		res := httptest.NewRecorder()
		EditUser(res, req)
		if status := res.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		expected := util.JsonResponse{false, "some of the fields cannot be edited"}
		util.CheckExpectedActualJSON(res.Body, expected, t)
	})

	t.Run("EditUser change name", func(t *testing.T) {
		payload, err := json.Marshal(MockUser{EmpID: "123", Name: "sdsds"})
		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest("PUT", "/api/editUser", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatal(err)
		}

		res := httptest.NewRecorder()
		EditUser(res, req)
		if status := res.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		expected := util.JsonResponse{true, "edit successful"}

		util.CheckExpectedActualJSON(res.Body, expected, t)
	})
}

func TestChangePassword(t *testing.T) {
	t.Run("ChangePassword wrong password", func(t *testing.T) {
		payload, err := json.Marshal(struct {
			OldPassword string
			NewPassword string
		}{"abc", "def"})

		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest("PUT", "/api/changePassword", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatal(err)
		}
		context.Set(req, "user", models.User{Email: "abc@gmail.com"})
		res := httptest.NewRecorder()
		ChangePassword(res, req)
		if status := res.Code; status != http.StatusUnauthorized {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		expected := util.JsonResponse{false, "username or password not matching"}
		util.CheckExpectedActualJSON(res.Body, expected, t)
	})

	t.Run("ChangePassword correct password", func(t *testing.T) {
		payload, err := json.Marshal(struct {
			OldPassword string
			NewPassword string
		}{"textsecret", "def"})

		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest("PUT", "/api/changePassword", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatal(err)
		}
		context.Set(req, "user", models.User{Email: "abc@gmail.com"})
		res := httptest.NewRecorder()
		ChangePassword(res, req)
		if status := res.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		expected := util.JsonResponse{true, "password reset successful"}
		util.CheckExpectedActualJSON(res.Body, expected, t)
	})

	t.Run("ChangePassword unknown user", func(t *testing.T) {
		payload, err := json.Marshal(struct {
			OldPassword string
			NewPassword string
		}{"textsecret", "def"})

		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest("PUT", "/api/changePassword", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatal(err)
		}
		context.Set(req, "user", models.User{Email: "nouser@gmail.com"})
		res := httptest.NewRecorder()
		ChangePassword(res, req)
		if status := res.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		expected := util.JsonResponse{false, "User not found"}
		util.CheckExpectedActualJSON(res.Body, expected, t)
	})

	t.Run("ChangePassword empty password", func(t *testing.T) {
		payload, err := json.Marshal(struct {
			OldPassword string
			NewPassword string
		}{"", "def"})

		if err != nil {
			t.Fatal(err)
		}
		req, err := http.NewRequest("PUT", "/api/changePassword", bytes.NewBuffer(payload))
		if err != nil {
			t.Fatal(err)
		}
		context.Set(req, "user", models.User{Email: "abc@gmail.com"})
		res := httptest.NewRecorder()
		ChangePassword(res, req)
		if status := res.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		expected := util.JsonResponse{false, "password cannot be empty"}
		util.CheckExpectedActualJSON(res.Body, expected, t)
	})
}

func TestResetPassword(t *testing.T) {
	t.Run("ResetPassword mismatch password", func(t *testing.T) {
		req, err := http.NewRequest("PUT", "/forgotPassword/abc", nil)
		if err != nil {
			t.Fatal(err)
		}

		form := url.Values{}
		form.Add("NewPassword", "new")
		form.Add("ConfirmPassword", "confirm")
		req.PostForm = form
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		res := httptest.NewRecorder()
		ResetPassword(res, req)
		if status := res.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		expected := util.JsonResponse{false, "Missing password or password not equal"}
		util.CheckExpectedActualJSON(res.Body, expected, t)
	})

	t.Run("ResetPassword valid password", func(t *testing.T) {
		tokenMem = &tokenAll{TokenInterface: MockToken{}}
		req, err := http.NewRequest("PUT", "/forgotPassword/abc", nil)
		if err != nil {
			t.Fatal(err)
		}
		token, err := utilMem.UtilInterface.GetJWT(
			struct {
				EmpID string
				Email string
			}{
				EmpID: "abc",
				Email: "abc@gmail.com",
			}, "FORGOT_PASSWORD_SECRET")

		if err != nil {
			t.Fatal(err)
		}
		form := url.Values{}
		form.Add("NewPassword", "new")
		form.Add("ConfirmPassword", "new")
		form.Add("resetToken", token)
		req.PostForm = form
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		res := httptest.NewRecorder()
		ResetPassword(res, req)
		if status := res.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		expected := util.JsonResponse{true, "password reset successful"}
		util.CheckExpectedActualJSON(res.Body, expected, t)
	})

	t.Run("ResetPassword empty resetToken", func(t *testing.T) {
		tokenMem = &tokenAll{TokenInterface: MockToken{}}
		req, err := http.NewRequest("PUT", "/forgotPassword/abc", nil)
		if err != nil {
			t.Fatal(err)
		}
		form := url.Values{}
		form.Add("NewPassword", "new")
		form.Add("ConfirmPassword", "new")
		form.Add("resetToken", "")
		req.PostForm = form
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		res := httptest.NewRecorder()
		ResetPassword(res, req)
		if status := res.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		expected := util.JsonResponse{false, "token missing"}
		util.CheckExpectedActualJSON(res.Body, expected, t)
	})

	t.Run("ResetPassword resetToken of different user", func(t *testing.T) {
		tokenMem = &tokenAll{TokenInterface: MockToken{}}
		req, err := http.NewRequest("PUT", "/forgotPassword/abc", nil)
		if err != nil {
			t.Fatal(err)
		}
		form := url.Values{}
		form.Add("NewPassword", "new")
		form.Add("ConfirmPassword", "new")
		form.Add("resetToken", "")
		req.PostForm = form
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		res := httptest.NewRecorder()
		ResetPassword(res, req)
		if status := res.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		expected := util.JsonResponse{false, "token missing"}
		util.CheckExpectedActualJSON(res.Body, expected, t)
	})
}
