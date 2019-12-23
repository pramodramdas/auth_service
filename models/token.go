package models

import (
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/pramod/auth_service/config"
	util "github.com/pramod/auth_service/utils"
)

type utilAll util.UtilMembers

var utilMem *utilAll

func init() {
	utilMem = &utilAll{UtilInterface: util.Util{}}
}

type Token struct {
}

func (t Token) CreateTokenTable() (bool, error) {
	query := `Create table if not exists tokens (
		used_token varchar(300) not null,
		cleanup_time float not null,
		PRIMARY KEY (used_token)
	)`

	_, err := config.DB.Exec(query)
	if err != nil {
		fmt.Printf("An error occurred in CreateUserTable. %v", err)
		return false, err
	}

	return true, nil
}

func (t Token) InsertUsedToken(tokenString string) (bool, error) {

	obj, err := utilMem.UtilInterface.IsValidToken(tokenString, "FORGOT_PASSWORD_SECRET")

	if err != nil {
		return false, err
	}

	if util.IsZeroValue(obj["exp"]) == true {
		return false, errors.New("exp missing")
	}

	queryObj := sq.Insert("tokens").Columns("used_token", "cleanup_time").Values(tokenString, obj["exp"].(float64))

	query, args, err := queryObj.PlaceholderFormat(sq.Dollar).ToSql()
	// fmt.Println(query, args)
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	_, err = config.DB.Exec(query, args...)

	if err != nil {
		fmt.Println(err)
		return false, err
	}
	return true, nil
}

func (t Token) IsTokenUsed(tokenString string) (bool, error) {
	var count int
	err := config.DB.QueryRow("SELECT COUNT(*) FROM tokens where used_token= '" + tokenString + "'").Scan(&count)

	if err != nil {
		fmt.Println(err)
		return false, err
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}
