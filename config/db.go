package config

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func DbInit() {
	var err error
	DB, err = sql.Open("postgres", os.Getenv("DB_URL"))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	if err = DB.Ping(); err != nil {
		fmt.Println(err)
		panic(err)
	}
	fmt.Println("you connected to your db")
}
