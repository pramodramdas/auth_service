package models

type UserInterface interface {
	CreateUserTable() (bool, error)
	ExtractUserFromInterface(userInter map[string]interface{}) (User, error)
	CreateUser() (bool, error)
	UpdateUser() (bool, error)
	GetUser(match map[string]interface{}) ([]User, error)
	ModifyPassword(empId, password string) (bool, error)
}

type UserMembers struct {
	UserInterface
}
