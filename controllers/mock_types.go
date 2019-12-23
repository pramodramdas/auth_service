package controllers

import (
	models "github.com/pramod/auth_service/models"
)

type MockUser struct {
	EmpID    string
	Email    string
	Name     string
	Password string
	Role     string
	Age      int
}

type MockAppMail struct {
	Subject          string
	To               string
	ToName           string
	From             string
	FromName         string
	PlainTextContent string
	HtmlContent      string
}

func (u MockUser) GetUser(match map[string]interface{}) ([]models.User, error) {
	// password := "textsecret"
	// bytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	// hash := string(bytes)
	if match["email"] == "nouser@gmail.com" || match["empId"] == "def" {
		return []models.User{}, nil
	} else if match["empId"] == "abc" {
		userArray := []models.User{}
		userArray = append(userArray, models.User{EmpID: "abc", Email: "abc@gmail.com"})
		return userArray, nil
	}

	users := []models.User{models.User{Password: "$2a$04$2FT/rphEJM/vgJp4MdXff.XhQgvpcmC0uG.LMvcisBeIKTT9hrG3i"}}
	return users, nil
}

func (u MockUser) CreateUserTable() (bool, error) {
	return true, nil
}

func (u MockUser) ExtractUserFromInterface(userInter map[string]interface{}) (models.User, error) {
	return models.User{}, nil
}

func (u MockUser) CreateUser() (bool, error) {
	return true, nil
}

func (u MockUser) UpdateUser() (bool, error) {
	return true, nil
}

func (u MockUser) ModifyPassword(empId, password string) (bool, error) {
	return true, nil
}

func (am *MockAppMail) AssignDefault() {

}

func (am *MockAppMail) SendMail() error {
	return nil
}

func (am *MockAppMail) GetToMail() string {
	return am.To
}
