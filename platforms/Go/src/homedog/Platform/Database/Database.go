package common

import (
	"github.com/go-pg/pg"
	"log"
)

func Connect() *pg.DB {
	db := pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "p",
		Database: "postgres",
		Addr:     "db.homedog:5432",
	})
	log.Println("Connected to homedog.db")
	return db
}
