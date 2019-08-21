package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/context"
	"github.com/lib/pq"

	// _ "github.com/lib/pq"
	"github.com/pramod/auth_service/config"
	models "github.com/pramod/auth_service/models"
	util "github.com/pramod/auth_service/utils"
)

func CreateUserTable() (bool, error) {
	query := `Create table if not exists users (
		empid varchar(20) not null,
		email varchar(20) not null UNIQUE,
		name varchar(20) not null,
		password varchar(60) not null,
		role varchar(10) not null,
		age int,
		PRIMARY KEY (empID)
	)`
	_, err := config.DB.Exec(query)
	if err != nil {
		fmt.Printf("An error occurred in CreateUserTable. %v", err)
		return false, err
	}
	return true, nil
}

func Home(res http.ResponseWriter, req *http.Request) {
	sessionData := context.Get(req, "user")
	name := ""
	if sessionData != nil {
		//from interface{} to struct
		name = sessionData.(models.User).Name
	}

	io.WriteString(res, fmt.Sprintf("welcome user %s", name))
}

func AddUser(res http.ResponseWriter, req *http.Request) {
	var u models.User
	var data util.JsonResponse

	res.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&u)
	if err != nil {
		fmt.Println(err.Error())
		util.SendErrorResponse(res, "error")
		return
	}

	// validate except age
	if u.EmpID == "" || u.Name == "" || u.Email == "" || u.Password == "" || u.Role == "" {
		util.SendErrorResponse(res, "missing parameters")
		return
	}
	// call CreateUser function
	_, err = u.CreateUser()
	if err != nil {
		fmt.Println(err)
		pqErr := err.(*pq.Error)
		if pqErr.Code == "23505" {
			util.SendErrorResponse(res, "user already exists")
		} else {
			util.SendErrorResponse(res, "error")
		}
		return
	}

	// jsonData, err := json.Marshal(util.JsonResponse{true, "user created"})
	// res.WriteHeader(200)
	// res.Write(jsonData)
	data = util.JsonResponse{true, "user created"}
	json.NewEncoder(res).Encode(data)
}

func EditUser(res http.ResponseWriter, req *http.Request) {
	var u models.User
	var data util.JsonResponse

	res.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&u)

	if err != nil {
		fmt.Println(err.Error())
		util.SendErrorResponse(res, "error")
		return
	}

	if u == (models.User{}) || util.IsZeroValue(u.EmpID) == true {
		fmt.Println("empty body or empId missing")
		util.SendErrorResponse(res, "empty body or empId missing")
		return
	}

	if util.IsZeroValue(u.Password) != true || util.IsZeroValue(u.Role) != true {
		fmt.Println("some of the fields cannot be edited")
		util.SendErrorResponse(res, "some of the fields cannot be edited")
		return
	}

	//call updateUser
	_, err = u.UpdateUser()

	if err != nil {
		fmt.Println(err)
		util.SendErrorResponse(res, "error")
		return
	}
	data = util.JsonResponse{true, "edit successful"}
	json.NewEncoder(res).Encode(data)
}
