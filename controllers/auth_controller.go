package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	models "github.com/pramod/auth_service/models"
	util "github.com/pramod/auth_service/utils"
	"golang.org/x/crypto/bcrypt"
)

func Login(res http.ResponseWriter, req *http.Request) {
	var u models.User
	//var data util.JsonResponse

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&u)

	if err != nil {
		fmt.Println(err.Error())
		util.SendErrorResponse(res, "error")
		return
	}

	if util.IsZeroValue(u.Email) == true || util.IsZeroValue(u.Password) == true {
		util.SendErrorResponse(res, "missing missing Email or Password")
		return
	}

	// hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.MinCost)

	query := make(map[string]interface{})
	query["email"] = u.Email
	// query["password"] = hashedPassword

	userArray, err := models.GetUser(query)

	if len(userArray) == 0 {
		util.SendErrorResponse(res, "User not found")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(userArray[0].Password), []byte(u.Password))
	if err != nil {
		res.WriteHeader(401)
		util.SendErrorResponse(res, "username or password not matching")
		return
	}
	userArray[0].Password = ""

	token, err := util.GetJWT(userArray[0])

	if err != nil {
		res.WriteHeader(500)
		util.SendErrorResponse(res, "Internal server error")
		return
	}

	userObj := struct {
		Name, Email, Role, Token string
	}{
		Name:  userArray[0].Name,
		Email: userArray[0].Email,
		Role:  userArray[0].Role,
		Token: token,
	}
	json.NewEncoder(res).Encode(userObj)
}

func ValidateToken(res http.ResponseWriter, req *http.Request) {
	_, err := util.IsValidToken(req.Header.Get("token"))

	if err != nil {
		res.WriteHeader(403)
		util.SendErrorResponse(res, err.Error())
		return
	}
	json.NewEncoder(res).Encode(util.JsonResponse{true, "authorized"})
}
