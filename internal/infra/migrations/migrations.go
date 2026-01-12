package migrations

import (
	"database/sql"
	"log"

	"github.com/pressly/goose/v3"
)

func Run(db *sql.DB, dir string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db, dir); err != nil {
		return err
	}

	return nil
}

func RunWithLog(db *sql.DB, dir string) {
	if err := Run(db, dir); err != nil {
		log.Fatalf("migrations error: %v", err)
	}
	log.Println("migrations applied successfully")
}
