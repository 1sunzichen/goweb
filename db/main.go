package db

import (
	"database/sql"
	"log"
)

func init() {
	//127.0.0.1
	db, err := sql.Open("mysql", "root:123456@tcp(localhost:3306)/notesql?parseTime=true")
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	db.SetMaxIdleConns(20)
	db.SetMaxOpenConns(20)

	if err := db.Ping(); err != nil {
		log.Fatalln(err)
	}
}
