package models

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"os"
	sq "github.com/Masterminds/squirrel"
	"github.com/mitchellh/mapstructure"
	"github.com/pramod/auth_service/config"
	util "github.com/pramod/auth_service/utils"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	EmpID    string
	Email    string
	Name     string
	Password string
	Role     string
	Age      int
}

func (u User) CreateUserTable() (bool, error) {
	query := `Create table if not exists users (
		empid varchar(20) not null,
		email varchar(100) not null UNIQUE,
		name varchar(100) not null,
		password varchar(100) not null,
		role varchar(20) not null,
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

func (u User) InsertSeedUser() () {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.MinCost)

	if err != nil {
		fmt.Printf("An error occurred in InsertSeedUser. %v", err)
	}

	query := fmt.Sprintf(`insert into users values ('1','%s', 'abc', '%s', 'admin', 20) on conflict do nothing`, adminEmail, hashedPassword)
	_, err = config.DB.Exec(query)
	if err != nil {
		fmt.Printf("An error occurred in InsertSeedUser. %v", err)
	}
}


func extractUsers(rows *sql.Rows) ([]User, error) {
	var result []User
	for rows.Next() {
		var email, name, empId, password, role string
		var age int

		if err := rows.Scan(&empId, &email, &name, &password, &role, &age); err != nil {
			fmt.Println(err)
			return []User{}, err
		}
		result = append(result, User{empId, email, name, password, role, age})
	}
	if err := rows.Err(); err != nil {
		fmt.Println(err)
		return []User{}, err
	}
	return result, nil
}

func (u User) ExtractUserFromInterface(userInter map[string]interface{}) (User, error) {
	userObj := User{}
	err := mapstructure.Decode(userInter, &userObj)
	return userObj, err
}

func (u User) CreateUser() (bool, error) {
	query := `Insert into users (empid, email, name, password, role, age) values ($1, $2, $3, $4, $5, $6)`
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.MinCost)

	if err != nil {
		fmt.Printf("An error occurred in CreateUser. %v", err)
		return false, err
	}

	_, err = config.DB.Exec(query, u.EmpID, u.Email, u.Name, hashedPassword, strings.ToLower(u.Role), u.Age)

	if err != nil {
		fmt.Printf("An error occurred in CreateUser. %v", err)
		return false, err
	}
	return true, nil
}

func (u User) UpdateUser() (bool, error) {
	queryObj := sq.Update("users")
	if util.IsZeroValue(u.EmpID) != true {
		queryObj = queryObj.Where(sq.Eq{"empid": u.EmpID})
	} else {
		return false, errors.New("missing EmpId")
	}
	if util.IsZeroValue(u.Email) != true {
		queryObj.Set("email", u.Email)
	}
	if util.IsZeroValue(u.Name) != true {
		queryObj = queryObj.Set("name", u.Name)
	}
	if util.IsZeroValue(u.Age) != true {
		queryObj = queryObj.Set("age", u.Age)
	}
	//query, args, err
	query, args, err := queryObj.PlaceholderFormat(sq.Dollar).ToSql()

	// fmt.Println(query, args)

	if err != nil {
		fmt.Println(err)
		return false, err
	}
	// db := queryObj.RunWith(config.DB)
	// result, err := db.Exec()

	//or
	// fmt.Println(args)
	_, err = config.DB.Exec(query, args...)

	if err != nil {
		fmt.Println(err)
		return false, err
	}
	return true, nil
}

func (u User) GetUser(match map[string]interface{}) ([]User, error) {
	queryObj := sq.Select("*").From("users")
	var result []User

	for k := range match {
		v := reflect.ValueOf(match[k])

		switch v.Kind() {
		case reflect.String:
			queryObj = queryObj.Where(sq.Eq{k: v.String()})
			// case reflect.Slice:
			// 	queryObj = queryObj.Where(sq.Eq{k: string(v.Interface().([]byte))})
		}
	}

	query, args, err := queryObj.PlaceholderFormat(sq.Dollar).ToSql()
	// fmt.Println(query, args, err)

	rows, err := config.DB.Query(query, args...)

	if err != nil {
		fmt.Println(err)
		return []User{}, err
	}
	defer rows.Close()

	result, err = extractUsers(rows)

	return result, nil
}

func (u User) ModifyPassword(empId, password string) (bool, error) {
	queryObj := sq.Update("users")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)

	if util.IsZeroValue(empId) != true {
		queryObj = queryObj.Where(sq.Eq{"empid": empId})
	} else {
		return false, errors.New("missing EmpId")
	}

	if err != nil {
		fmt.Printf("An error occurred in ModifyPassword. %v", err)
		return false, err
	}

	queryObj = queryObj.Set("password", hashedPassword)

	query, args, err := queryObj.PlaceholderFormat(sq.Dollar).ToSql()
	//fmt.Println(args)
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
