package controllers

import (
	"fmt"
	"io"
	"os"
	"net/http"

	"github.com/gorilla/context"
	"golang.org/x/crypto/bcrypt"

	"encoding/json"

	"github.com/lib/pq"
	models "github.com/pramod/auth_service/models"
	"strings"
	util "github.com/pramod/auth_service/utils"
)

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
	addUserSpan, ctx := zepkinMem.ZepkinInterface.CustomStartSpanFromContext(req.Context(), "AddUser")
	var u models.User
	var data util.JsonResponse

	res.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&u)

	if err != nil {
		fmt.Println(err.Error())
		utilMem.UtilInterface.SendErrorResponse(res, "error")
		return
	}

	if os.Getenv("GO_ENV") == "testing" {
		userMem = &userAll{UserInterface: MockUser(u)}
	} else {
		userMem = &userAll{UserInterface: u}
	}

	// validate except age
	if u.EmpID == "" || u.Name == "" || u.Email == "" || u.Password == "" || u.Role == "" {
		utilMem.UtilInterface.SendErrorResponse(res, "missing parameters")
		return
	}

	createUserSpan, ctx := zepkinMem.ZepkinInterface.CustomStartSpanFromContext(ctx, "CreateUser")
	// call CreateUser function
	// _, err = u.CreateUser()
	_, err = userMem.UserInterface.CreateUser()
	zepkinMem.ZepkinInterface.CustomSpanFinish(createUserSpan)

	if err != nil {
		fmt.Println(err)
		pqErr := err.(*pq.Error)
		if pqErr.Code == "23505" {
			utilMem.UtilInterface.SendErrorResponse(res, "user already exists")
		} else {
			utilMem.UtilInterface.SendErrorResponse(res, "error")
		}
		return
	}

	// jsonData, err := json.Marshal(util.JsonResponse{true, "user created"})
	// res.WriteHeader(200)
	// res.Write(jsonData)
	data = util.JsonResponse{true, "user created"}
	json.NewEncoder(res).Encode(data)
	zepkinMem.ZepkinInterface.CustomSpanFinish(addUserSpan)
}

func EditUser(res http.ResponseWriter, req *http.Request) {
	editUserSpan, ctx := zepkinMem.ZepkinInterface.CustomStartSpanFromContext(req.Context(), "EditUser")
	var u models.User
	var data util.JsonResponse

	res.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&u)

	if err != nil {
		fmt.Println(err.Error())
		utilMem.UtilInterface.SendErrorResponse(res, "error")
		return
	}

	if os.Getenv("GO_ENV") == "testing" {
		userMem = &userAll{UserInterface: MockUser(u)}
	} else {
		userMem = &userAll{UserInterface: u}
	}

	if u == (models.User{}) || util.IsZeroValue(u.EmpID) == true {
		fmt.Println("empty body or empId missing")
		utilMem.UtilInterface.SendErrorResponse(res, "empty body or empId missing")
		return
	}

	if util.IsZeroValue(u.Password) != true || util.IsZeroValue(u.Role) != true {
		fmt.Println("some of the fields cannot be edited")
		utilMem.UtilInterface.SendErrorResponse(res, "some of the fields cannot be edited")
		return
	}

	updateUserSpan, ctx := zepkinMem.ZepkinInterface.CustomStartSpanFromContext(ctx, "UpdateUser")
	//call updateUser
	// _, err = u.UpdateUser()
	_, err = userMem.UserInterface.UpdateUser()
	zepkinMem.ZepkinInterface.CustomSpanFinish(updateUserSpan)

	if err != nil {
		fmt.Println(err)
		utilMem.UtilInterface.SendErrorResponse(res, "error")
		return
	}
	data = util.JsonResponse{true, "edit successful"}
	json.NewEncoder(res).Encode(data)
	zepkinMem.ZepkinInterface.CustomSpanFinish(editUserSpan)
}

func ChangePassword(res http.ResponseWriter, req *http.Request) {
	changePassSpan, ctx := zepkinMem.ZepkinInterface.CustomStartSpanFromContext(req.Context(), "ChangePassword")
	var data util.JsonResponse
	resetPassword := struct {
		OldPassword string
		NewPassword string
	}{}

	sessionData := context.Get(req, "user")
	if util.IsZeroValue(sessionData.(models.User).Email) == true {
		res.WriteHeader(403)
		utilMem.UtilInterface.SendErrorResponse(res, "emailId missing")
		return
	}

	res.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&resetPassword)

	if err != nil {
		fmt.Println(err.Error())
		utilMem.UtilInterface.SendErrorResponse(res, "error")
		return
	}

	if util.IsZeroValue(resetPassword.OldPassword) == true || util.IsZeroValue(resetPassword.NewPassword) == true {
		utilMem.UtilInterface.SendErrorResponse(res, "password cannot be empty")
		return
	}

	query := make(map[string]interface{})
	query["email"] = sessionData.(models.User).Email

	getUserSpan, ctx := zepkinMem.ZepkinInterface.CustomStartSpanFromContext(ctx, "GetUser")
	userArray, err := userMem.UserInterface.GetUser(query)
	zepkinMem.ZepkinInterface.CustomSpanFinish(getUserSpan)

	if len(userArray) == 0 {
		utilMem.UtilInterface.SendErrorResponse(res, "User not found")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(userArray[0].Password), []byte(resetPassword.OldPassword))
	if err != nil {
		res.WriteHeader(401)
		utilMem.UtilInterface.SendErrorResponse(res, "username or password not matching")
		return
	}

	modifyPasswordSpan, ctx := zepkinMem.ZepkinInterface.CustomStartSpanFromContext(ctx, "ModifyPassword")
	_, err = userMem.UserInterface.ModifyPassword(sessionData.(models.User).EmpID, resetPassword.NewPassword)
	zepkinMem.ZepkinInterface.CustomSpanFinish(modifyPasswordSpan)

	if err != nil {
		fmt.Println(err)
		utilMem.UtilInterface.SendErrorResponse(res, "error")
		return
	}
	data = util.JsonResponse{true, "password reset successful"}
	json.NewEncoder(res).Encode(data)
	zepkinMem.ZepkinInterface.CustomSpanFinish(changePassSpan)
}

func ResetPassword(res http.ResponseWriter, req *http.Request) {
	resetPassSpan, ctx := zepkinMem.ZepkinInterface.CustomStartSpanFromContext(req.Context(), "ResetPassword")
	req.ParseForm()
	newPassword := strings.Join(req.Form["NewPassword"], "")
	confirmPassword := strings.Join(req.Form["ConfirmPassword"], "")
	resetToken := strings.Join(req.Form["resetToken"], "")

	if newPassword == "" || confirmPassword == "" || newPassword != confirmPassword {
		utilMem.UtilInterface.SendErrorResponse(res, "Missing password or password not equal")
		return
	}
	obj, err := utilMem.UtilInterface.IsValidToken(resetToken, "FORGOT_PASSWORD_SECRET")

	if err != nil {
		utilMem.UtilInterface.SendErrorResponse(res, err.Error())
		return
	}

	isTokenUsedSpan, ctx := zepkinMem.ZepkinInterface.CustomStartSpanFromContext(ctx, "IsTokenUsed")
	used, err := tokenMem.TokenInterface.IsTokenUsed(resetToken)
	zepkinMem.ZepkinInterface.CustomSpanFinish(isTokenUsedSpan)

	if err != nil {
		utilMem.UtilInterface.SendErrorResponse(res, err.Error())
		return
	} else if used == true {
		utilMem.UtilInterface.SendErrorResponse(res, "token already used")
		return
	}

	extractUserSpan, ctx := zepkinMem.ZepkinInterface.CustomStartSpanFromContext(ctx, "ExtractUserFromInterface")
	userObj, err := userMem.UserInterface.ExtractUserFromInterface(obj)
	zepkinMem.ZepkinInterface.CustomSpanFinish(extractUserSpan)

	if err != nil {
		utilMem.UtilInterface.SendErrorResponse(res, err.Error())
		return
	}

	modifyPassSpan, ctx := zepkinMem.ZepkinInterface.CustomStartSpanFromContext(ctx, "ModifyPassword")
	_, err = userMem.UserInterface.ModifyPassword(userObj.EmpID, newPassword)
	zepkinMem.ZepkinInterface.CustomSpanFinish(modifyPassSpan)

	if err != nil {
		utilMem.UtilInterface.SendErrorResponse(res, err.Error())
		return
	}

	insertUsedTokenSpan, ctx := zepkinMem.ZepkinInterface.CustomStartSpanFromContext(ctx, "InsertUsedToken")
	_, err = tokenMem.TokenInterface.InsertUsedToken(resetToken)
	zepkinMem.ZepkinInterface.CustomSpanFinish(insertUsedTokenSpan)

	if err != nil {
		utilMem.UtilInterface.SendErrorResponse(res, err.Error())
		return
	}

	json.NewEncoder(res).Encode(util.JsonResponse{true, "password reset successful"})
	zepkinMem.ZepkinInterface.CustomSpanFinish(resetPassSpan)
}
