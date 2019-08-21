package main

import (
	"fmt"
	"net/http"
	"os"

	models "github.com/pramod/auth_service/models"
	util "github.com/pramod/auth_service/utils"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/pramod/auth_service/config"
	"github.com/pramod/auth_service/controllers"
)

func init() {
	godotenv.Load()
	config.DbInit()
	controllers.CreateUserTable()
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func authorizeToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		obj, err := util.IsValidToken(r.Header.Get("token"))

		if err != nil {
			w.WriteHeader(403)
			util.SendErrorResponse(w, err.Error())
			return
		}
		//convert from map[string]interface{} to struct
		userObj, err := models.ExtractUserFromInterface(obj)

		context.Set(r, "user", userObj)
		next.ServeHTTP(w, r)
	})
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.Use(logRequests)
	router.HandleFunc("/", controllers.Home).Methods("GET")
	router.HandleFunc("/login", controllers.Login).Methods("POST")
	router.HandleFunc("/isValidToken", controllers.ValidateToken).Methods("GET")

	appRouter := router.PathPrefix("/api/").Subrouter()
	appRouter.Use(authorizeToken)
	appRouter.HandleFunc("/addUser", controllers.AddUser).Methods("POST")
	appRouter.HandleFunc("/editUser", controllers.EditUser).Methods("PUT")
	appRouter.HandleFunc("/test", controllers.Home).Methods("GET")
	http.ListenAndServe(os.Getenv("HTTP_HOST")+":"+os.Getenv("HTTP_PORT"), router)
}
