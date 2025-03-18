package connections

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

func InitMssqlDB(dataSourceName string, maxOpenConns, maxIdleConns int, connMaxLifetime time.Duration) (*sql.DB, error) {
	db, err := sql.Open("sqlserver", dataSourceName)
	if err != nil {
		log.Println("Error opening database: %q", err)
		return nil, err
	}

	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	err = db.Ping()
	if err != nil {
		log.Println("Error connecting to the database: %q", err)
		return nil, err
	}

	// fmt.Println("Successfully connected to the MSSQL database!")
	return db, nil
}
