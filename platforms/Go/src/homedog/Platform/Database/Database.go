package common

import (
	"github.com/go-pg/pg"
	"log"
	"os"
)

func Connect() *pg.DB {
	db := pg.Connect(&pg.Options{
		User:     "postgres",
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Database: "postgres",
		Addr:     "db.homedog:5432",
	})
	log.Println("Connected to homedog.db")
	return db
}
