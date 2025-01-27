package connections

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func InitPostgresDB(source_cred string, maxOpenConns, maxIdleConns int, connMaxLifetime time.Duration) (*sql.DB, error) {
	db, err := sql.Open("postgres", source_cred)
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %q", err)
		return nil, err
	}

	fmt.Println("Successfully connected to the Postgres database!")
	return db, nil
}
