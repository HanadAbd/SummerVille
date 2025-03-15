package connections

/*
This script will get the environment
*/

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

var (
	ProdConn    *sql.DB
	SourcesConn *Connector
)

func getMssqlDSN(server, database, trustedConnection string) string {
	return fmt.Sprintf("server=%s;database=%s;%s", server, database, trustedConnection)
}
func getPostgresDSN(user, password, dbname, host, port, sslmode string) string {
	return fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s", user, password, dbname, host, port, sslmode)
}

func GetProdDatabase() (*sql.DB, error) {
	var err error
	if ProdConn != nil {
		return ProdConn, nil
	}
	prodEnv := getProdCred()
	ProdConn, err = sql.Open("postgres", prodEnv)
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
		return nil, err
	}

	ProdConn.SetMaxOpenConns(1)
	ProdConn.SetMaxIdleConns(1)
	ProdConn.SetConnMaxLifetime(0)

	err = ProdConn.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %q", err)
		return nil, err
	}

	// fmt.Println("Successfully connected to the Production Database!")
	return ProdConn, nil
}
func getProdCred() string {
	return fmt.Sprintf(
		"user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
	)
}

func InitConnector() (*Connector, error) {
	return nil, nil
}

func CloseConnector() error {
	ProdConn.Close()
	for _, conn := range SourcesConn.PostgresDB {
		conn.Conn.Close()
	}
	for _, conn := range SourcesConn.MssqlDB {
		conn.Conn.Close()
	}
	return nil
}
