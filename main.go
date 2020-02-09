package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/context"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	zipkinHttpMiddleware "github.com/openzipkin/zipkin-go/middleware/http"
	config "github.com/pramod/auth_service/config"
	controllers "github.com/pramod/auth_service/controllers"
	allInt "github.com/pramod/auth_service/inter"
	models "github.com/pramod/auth_service/models"
	util "github.com/pramod/auth_service/utils"
)

type utilAll util.UtilMembers
type userAll models.UserMembers
type tokenAll models.TokenMembers
type zepkinAll config.ZepkinMembers

var utilMem *utilAll
var userMem *userAll
var tokenMem *tokenAll
var zepkinMem *zepkinAll

var ai allInt.AllMembers

func init() {
	ai = allInt.AllMembers{ConfInterface: config.ZepConf{}, UserInterface: models.User{}}
	utilMem = &utilAll{UtilInterface: util.Util{}}
	userMem = &userAll{UserInterface: models.User{}}
	tokenMem = &tokenAll{TokenInterface: models.Token{}}
	zepkinMem = &zepkinAll{ZepkinInterface: config.ZepConf{}}
	godotenv.Load()
	config.DbInit()
	userMem.UserInterface.CreateUserTable()
	(models.User{}).InsertSeedUser()
	tokenMem.TokenInterface.CreateTokenTable()
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func authorizeToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		obj, err := utilMem.UtilInterface.IsValidToken(r.Header.Get("token"), "JWT_SECRET")

		if err != nil {
			w.WriteHeader(403)
			utilMem.UtilInterface.SendErrorResponse(w, err.Error())
			return
		}
		//convert from map[string]interface{} to struct
		userObj, err := userMem.UserInterface.ExtractUserFromInterface(obj)

		context.Set(r, "user", userObj)
		next.ServeHTTP(w, r)
	})
}

func main() {
	allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "token"})
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"OPTIONS", "GET", "HEAD", "POST", "PUT", "DELETE"})
	
	router := mux.NewRouter().StrictSlash(true)
	tracer, reporter, err := zepkinMem.ZepkinInterface.RegisterZipkinTacer()
	if err != nil {
		log.Fatal(err)
	}
	zepkinMem.ZepkinInterface.RegisterZipkinClient()
	if err != nil {
		log.Fatal(err)
	}
	defer reporter.Close()

	router.Use(logRequests)

	// http.DefaultClient.Transport, err = zipkinhttpMiddleware.NewTransport(
	// 	tracer,
	// 	zipkinhttpMiddleware.TransportTrace(true),
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }

	serverMiddleware := zipkinHttpMiddleware.NewServerMiddleware(
		tracer,
		zipkinHttpMiddleware.TagResponseSize(true),
	)
	router.Use(serverMiddleware)

	router.HandleFunc("/", controllers.Home).Methods("GET")
	router.HandleFunc("/login", controllers.Login).Methods("POST")
	router.HandleFunc("/isValidToken", controllers.ValidateToken).Methods("GET")
	// // router.HandleFunc("/forgotPassword/{email}", controllers.ForgotPassword).Methods("POST")
	router.HandleFunc("/forgotPassword/{empId}", controllers.ForgotPassword).Methods("POST")
	router.HandleFunc("/resetPasswordPage/{resetToken}", controllers.ResetPasswordPage).Methods("GET")
	router.HandleFunc("/resetPassword", controllers.ResetPassword).Methods("POST")

	appRouter := router.PathPrefix("/api/").Subrouter()
	appRouter.Use(authorizeToken)
	appRouter.HandleFunc("/addUser", controllers.AddUser).Methods("POST")
	appRouter.HandleFunc("/editUser", controllers.EditUser).Methods("PUT")
	appRouter.HandleFunc("/changePassword", controllers.ChangePassword).Methods("POST")
	// appRouter.HandleFunc("/test", controllers.Home).Methods("GET")
	http.ListenAndServe(os.Getenv("HTTP_HOST")+":"+os.Getenv("HTTP_PORT"), handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)(router))
}
