package database

import (
	"database/sql"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	for {
		var err error
		connStr := "postgres://postgres:pass@postgres:5432/go-pg?sslmode=disable"
		DB, err = sql.Open("postgres", connStr)
		if err != nil {
			log.Printf("An error occured while interacting with the db, retrying in 30 seconds: %s", err)
			time.Sleep(30 * time.Second)
			continue
		}
		if err = DB.Ping(); err != nil {
			log.Printf("An error occured while interacting with the db, retrying in 30 seconds: %s", err)
			time.Sleep(30 * time.Second)
			continue
		}
		driver, err := postgres.WithInstance(DB, &postgres.Config{})
		if err != nil {
			log.Printf("An error occured while interacting with the db, retrying in 30 seconds: %s", err)
			time.Sleep(30 * time.Second)
			continue
		}
		m, err := migrate.NewWithDatabaseInstance(
			"file://../../internal/database/migrations",
			"postgres", driver)
		if err != nil {
			log.Printf("An error occured while interacting with the db, retrying in 30 seconds: %s", err)
			time.Sleep(30 * time.Second)
			continue
		}
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Printf("An error occured while interacting with the db, retrying in 30 seconds: %s", err)
			time.Sleep(30 * time.Second)
			continue
		}
		log.Println("Connected to DB")
		break
	}
}
