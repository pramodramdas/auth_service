package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	config "github.com/pramod/auth_service/config"
	allInt "github.com/pramod/auth_service/inter"
	models "github.com/pramod/auth_service/models"
	util "github.com/pramod/auth_service/utils"
	"golang.org/x/crypto/bcrypt"
)

type userObj struct {
	Name, Email, Role, Token string
}

type userAllInt allInt.AllMembers

type userAll models.UserMembers
type utilAll util.UtilMembers
type zepkinAll config.ZepkinMembers
type tokenAll models.TokenMembers
type mailAll util.MailMembers

var allMem *userAllInt
var utilMem *utilAll
var userMem *userAll
var zepkinMem *zepkinAll
var tokenMem *tokenAll
var mailMem *mailAll

func init() {
	utilMem = &utilAll{UtilInterface: util.Util{}}
	allMem = &userAllInt{UserInterface: models.User{}, ConfInterface: config.ZepConf{}}
	userMem = &userAll{UserInterface: models.User{}}
	zepkinMem = &zepkinAll{ZepkinInterface: config.ZepConf{}}
	tokenMem = &tokenAll{TokenInterface: models.Token{}}
	mailMem = &mailAll{MailInterface: &util.App_Mail{}}
}

func authenticateAndGetToken(ctx context.Context, u models.User) (userObj, util.CustomError) {
	// hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.MinCost)

	query := make(map[string]interface{})
	query["email"] = u.Email
	// query["password"] = hashedPassword

	dbSpan, _ := zepkinMem.ZepkinInterface.CustomStartSpanFromContext(ctx, "GetUser")
	userArray, err := userMem.UserInterface.GetUser(query)
	zepkinMem.ZepkinInterface.CustomSpanFinish(dbSpan)

	if len(userArray) == 0 {
		return userObj{}, util.CustomError{Err: errors.New("User not found"), StatusCode: 401}
	}

	hashSpan, _ := zepkinMem.ZepkinInterface.CustomStartSpanFromContext(ctx, "CompareHashAndPassword")
	err = bcrypt.CompareHashAndPassword([]byte(userArray[0].Password), []byte(u.Password))
	zepkinMem.ZepkinInterface.CustomSpanFinish(hashSpan)

	if err != nil {
		fmt.Println(err)
		return userObj{}, util.CustomError{Err: errors.New("username or password not matching"), StatusCode: 401}
	}
	userArray[0].Password = ""
	
	jwtSpan, _ := zepkinMem.ZepkinInterface.CustomStartSpanFromContext(ctx, "GetJWT")
	token, err := utilMem.UtilInterface.GetJWT(userArray[0], "JWT_SECRET")
	zepkinMem.ZepkinInterface.CustomSpanFinish(jwtSpan)

	if err != nil {
		fmt.Println(err)
		return userObj{}, util.CustomError{Err: errors.New("internal server error"), StatusCode: 500}
	}

	usrObj := userObj{
		Name:  userArray[0].Name,
		Email: userArray[0].Email,
		Role:  userArray[0].Role,
		Token: token,
	}

	return usrObj, util.CustomError{}
}

func Login(res http.ResponseWriter, req *http.Request) {
	var u models.User
	//var data util.JsonResponse
	loginSpan, ctx := zepkinMem.ZepkinInterface.CustomStartSpanFromContext(req.Context(), "Login")

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&u)

	if err != nil {
		fmt.Println(err.Error())
		utilMem.UtilInterface.SendErrorResponse(res, "error")
		return
	}

	if util.IsZeroValue(u.Email) == true || util.IsZeroValue(u.Password) == true {
		utilMem.UtilInterface.SendErrorResponse(res, "missing missing Email or Password")
		return
	}

	usrObj, errObj := authenticateAndGetToken(ctx, u)

	if errObj.Err != nil {
		if errObj.StatusCode > 0 {
			if errObj.StatusCode == 401 {
				res.WriteHeader(401)
			} else {
				res.WriteHeader(500)
			}
		}
		utilMem.UtilInterface.SendErrorResponse(res, errObj.Err.Error())
		return
	}

	json.NewEncoder(res).Encode(usrObj)
	zepkinMem.ZepkinInterface.CustomSpanFinish(loginSpan)
}

func ValidateToken(res http.ResponseWriter, req *http.Request) {
	_, err := utilMem.UtilInterface.IsValidToken(req.Header.Get("token"), "JWT_SECRET")

	if err != nil {
		res.WriteHeader(403)
		utilMem.UtilInterface.SendErrorResponse(res, err.Error())
		return
	}
	json.NewEncoder(res).Encode(util.JsonResponse{true, "authorized"})
}

func ForgotPassword(res http.ResponseWriter, req *http.Request) {
	fmt.Println("ForgotPassword")
	forgotSpan, ctx := zepkinMem.ZepkinInterface.CustomStartSpanFromContext(req.Context(), "ForgotPassword")
	empId := mux.Vars(req)["empId"]

	query := make(map[string]interface{})
	query["empId"] = empId
	
	dbSpan, _ := zepkinMem.ZepkinInterface.CustomStartSpanFromContext(ctx, "GetUser")
	userArray, err := userMem.UserInterface.GetUser(query)
	zepkinMem.ZepkinInterface.CustomSpanFinish(dbSpan)

	if len(userArray) == 0 {
		utilMem.UtilInterface.SendErrorResponse(res, "User not found")
		return
	}

	jwtSpan, _ := zepkinMem.ZepkinInterface.CustomStartSpanFromContext(ctx, "GetJWT")
	token, err := utilMem.UtilInterface.GetJWT(
		struct {
			EmpID string
			Email string
		}{
			EmpID: userArray[0].EmpID,
			Email: userArray[0].Email,
		}, "FORGOT_PASSWORD_SECRET")
	zepkinMem.ZepkinInterface.CustomSpanFinish(jwtSpan)

	if err != nil {
		res.WriteHeader(500)
		utilMem.UtilInterface.SendErrorResponse(res, "Internal server error")
		return
	}

	mailObj := util.App_Mail{}
	mailObj.AssignDefault()
	mailObj.Subject = fmt.Sprintf("Forgot password link for %s://%s", util.GetProtocol(req), req.Host)
	mailObj.To = userArray[0].Email
	mailObj.ToName = userArray[0].Email
	mailObj.PlainTextContent = fmt.Sprintf("Reset password link %s://%s/resetPasswordPage/%s", util.GetProtocol(req), req.Host, token)
	mailObj.HtmlContent = fmt.Sprintf("<strong><a href=\"%s://%s/resetPasswordPage/%s\">Reset password link </a></strong>", util.GetProtocol(req), req.Host, token)

	if os.Getenv("GO_ENV") == "testing" {
		mockMail := MockAppMail(mailObj)
		mailMem = &mailAll{MailInterface: &mockMail}
	} else {
		mailMem = &mailAll{MailInterface: &mailObj}
	}

	mailSpan, _ := zepkinMem.ZepkinInterface.CustomStartSpanFromContext(ctx, "SendMail")
	err = mailMem.MailInterface.SendMail()
	zepkinMem.ZepkinInterface.CustomSpanFinish(mailSpan)

	if err != nil {
		utilMem.UtilInterface.SendErrorResponse(res, err.Error())
		return
	}
	json.NewEncoder(res).Encode(util.JsonResponse{true, "reset link has been sent to your email id"})
	zepkinMem.ZepkinInterface.CustomSpanFinish(forgotSpan)
}

func ResetPasswordPage(res http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	html := `<html>
		<body>
			<form method="post" action="/resetPassword">
				<input type="hidden" name="resetToken" value="` + params["resetToken"] + `">
				New Password: <input type="password" name="NewPassword" /><br/>
				Confirm Password: <input type="password" name="ConfirmPassword" /><br/>
				<input type="submit" name="submit"/>
			</form>
		</body>
	</html>`
	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(res, html)
}
