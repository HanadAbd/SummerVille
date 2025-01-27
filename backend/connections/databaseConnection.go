package connections

/*
This script will get the environment
*/

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
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

func InitConnector() (*Connector, error) {
	mssqlDSN := getMssqlDSN("localhost\\SQLEXPRESS", "master", "trusted_connection=yes")
	postgresDSN := getPostgresDSN("postgres", "Week7890", "prototype-1", "localhost", "5432", "disable")
	excelUrl := "test_data/shift_handover.xlsx"

	maxOpenConns, maxIdleConns := 2, 3
	connMaxLifetime := time.Minute

	GetProdDatabase()

	postgresDB, err := InitPostgresDB(postgresDSN, maxOpenConns, maxIdleConns, connMaxLifetime)
	if err != nil {
		return nil, err
	}
	postgresConn := []*postgresConn{{Conn: postgresDB, Name: "production_data", Refreshtime: time.Now()}}

	mssqlDB, err := InitMssqlDB(mssqlDSN, maxOpenConns, maxIdleConns, connMaxLifetime)
	if err != nil {
		return nil, err
	}
	mssqlConn := []*mssqlConn{{Conn: mssqlDB, Name: "sensor_data", Refreshtime: time.Now()}}

	excelFile, err := InitExcel(excelUrl)
	if err != nil {
		return nil, err
	}
	excelConn := []*excelConn{{File: excelFile, Name: "shift_handover", Refreshtime: time.Now()}}

	SourcesConn = &Connector{
		PostgresDB: postgresConn,
		MssqlDB:    mssqlConn,
		ExcelFile:  excelConn,
	}
	return SourcesConn, nil
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

	fmt.Println("Successfully connected to the Production Database!")
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
