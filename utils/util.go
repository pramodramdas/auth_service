package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type JsonResponse struct {
	Success bool
	Msg     string
}

func SendErrorResponse(res http.ResponseWriter, msg string) {
	json.NewEncoder(res).Encode(JsonResponse{false, msg})
}

//return true if zero value
func IsZeroValue(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

func GetJWT(data interface{}) (string, error) {
	var dataMap map[string]interface{}
	userMap, _ := json.Marshal(data)
	json.Unmarshal(userMap, &dataMap)

	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)

	for k := range dataMap {
		v := reflect.ValueOf(dataMap[k])

		switch v.Kind() {
		case reflect.String:
			claims[k] = v.String()
		case reflect.Int:
			claims[k] = v.Int()
		}
	}
	expTime, err := strconv.ParseInt(os.Getenv("JWT_EXPIRE_TIME"), 10, 32)

	if err != nil {
		fmt.Println(err)
		return "", err
	}
	// fmt.Println(time.Duration(expTime))
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(expTime)).Unix()
	claims["iat"] = time.Now().Unix()
	token.Claims = claims

	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func IsValidToken(tokenString string) (map[string]interface{}, error) {

	if IsZeroValue(tokenString) == true {
		fmt.Println("token missing")
		return make(map[string]interface{}), errors.New("token missing")
	}

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		fmt.Println(err)
		return make(map[string]interface{}), err
	}
	return claims, nil
}
