package connections

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type PostgresConn struct {
	Conn *sql.DB
	Name string
}

type PostgresCred struct {
	User     string
	Password string
	DBName   string
	Host     string
	Port     string
	SSLMode  string
}

func (p *PostgresCred) GetSource() string {
	return fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s", p.User, p.Password, p.DBName, p.Host, p.Port, p.SSLMode)
}

func InitPostgresDB(ps *PostgresCred, dm *ConnectionMetrics) (*sql.DB, error) {
	db, err := sql.Open("postgres", ps.GetSource())
	if err != nil {
		log.Printf("Error opening database: %q\n", err)
		return nil, err
	}

	db.SetMaxOpenConns(dm.OpenConnections)
	db.SetMaxIdleConns(dm.IdleConnections)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(2 * time.Minute)

	err = db.Ping()
	if err != nil {
		log.Printf("Error connecting to the database: %q\n", err)
		return nil, err
	}

	log.Printf("Successfully connected to the Postgres database: %s\n", ps.DBName)
	return db, nil
}

func (p *PostgresConn) MonitorConnection(maxAttempts int, delay time.Duration) error {
	var err error
	if p == nil || p.Conn == nil {
		return fmt.Errorf("postgres connection or DB object is nil")
	}

	if err = p.Conn.Ping(); err != nil {
		err = p.RetryConnection(maxAttempts, delay)
	}
	return err
}

func (p *PostgresConn) RetryConnection(maxAttempts int, delay time.Duration) error {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := p.Conn.Ping()
		if err == nil {
			return nil
		}
		lastErr = err
		log.Printf("Connection attempt %d/%d failed: %v", attempt, maxAttempts, err)
		time.Sleep(delay)
	}
	return fmt.Errorf("failed to establish connection after %d attempts: %v", maxAttempts, lastErr)
}

func (p *PostgresConn) InitialiseData(table TableDefinition) error {
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = $1 AND table_name = $2
		)`
	var exists bool

	err := p.Conn.QueryRow(query, table.Schema, table.Name).Scan(&exists)
	if err != nil {
		return err
	}

	if !exists {
		createQuery := getCreateTableQuery(table)
		_, err = p.Conn.Exec(createQuery)
		if err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}
	return nil
}

func getCreateTableQuery(table TableDefinition) string {
	columns := make([]string, 0, len(table.Columns))
	for _, col := range table.Columns {
		columns = append(columns, fmt.Sprintf("%s %s", col.Name, col.Type))
	}

	return fmt.Sprintf("CREATE TABLE %s.%s (%s)", table.Schema, table.Name, strings.Join(columns, ", "))
}

func (p *PostgresConn) AddData(table TableDefinition, data []interface{}) error {
	var err error
	if p.Conn == nil {

		host := getEnv("DB_HOST", "localhost")
		user := getEnv("DB_USER", "postgres")
		password := getEnv("DB_PASSWORD", "Week7890")
		dbName := getEnv("DB_NAME", "summervilledb")
		port := getEnv("DB_PORT", "5432")
		sslMode := getEnv("DB_SSLMODE", "disable")

		// For Docker environment, use container name
		if os.Getenv("DOCKER_ENV") == "true" {
			host = "postgres" // Use the service name from docker-compose
			log.Printf("Running in Docker environment, connecting to PostgreSQL at %s\n", host)
		}

		log.Printf("Attempting to connect to PostgreSQL at %s:%s\n", host, port)

		p.Conn, err = InitPostgresDB(&PostgresCred{
			User:     user,
			Password: password,
			DBName:   dbName,
			Host:     host,
			Port:     port,
			SSLMode:  sslMode,
		}, &ConnectionMetrics{
			OpenConnections: 1,
			IdleConnections: 1,
			QueryCount:      0,
			LastQueryTime:   0,
		})

		if err != nil {
			return fmt.Errorf("error initializing database connection: %v", err)
		}

		if err := p.InitialiseData(table); err != nil {
			return err
		}

		return p.insertData(table, data)
	}

	if err := p.InitialiseData(table); err != nil {
		return err
	}

	return p.insertData(table, data)
}
func (p *PostgresConn) insertData(table TableDefinition, data []interface{}) error {
	if len(data) == 0 {
		return nil
	}

	columns := table.GetColumns()
	columnsStr := strings.Join(columns, ", ")

	if rowMap, ok := data[0].(map[string]interface{}); ok {
		placeholders := make([]string, len(columns))
		values := make([]interface{}, len(columns))

		for i, colName := range columns {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			if val, exists := rowMap[colName]; exists {
				values[i] = val
			} else {
				values[i] = nil
			}
		}

		query := fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s)",
			table.Schema, table.Name,
			columnsStr, strings.Join(placeholders, ", "))

		_, err := p.Conn.Exec(query, values...)
		return err
	}

	return fmt.Errorf("unsupported data format")
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}

func (p *PostgresConn) PurgeData(table TableDefinition) error {
	query := fmt.Sprintf("TRUNCATE TABLE %s.%s", table.Schema, table.Name)
	_, err := p.Conn.Exec(query)
	return err
}

func (p *PostgresConn) PurgeAllData() error {
	query := `
		SELECT table_schema, table_name 
		FROM information_schema.tables 
		WHERE table_schema NOT IN ('pg_catalog', 'information_schema')`

	rows, err := p.Conn.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var schema, name string
		if err := rows.Scan(&schema, &name); err != nil {
			return err
		}

		truncateQuery := fmt.Sprintf("TRUNCATE TABLE %s.%s CASCADE", schema, name)
		if _, err := p.Conn.Exec(truncateQuery); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (p *PostgresConn) CloseConnection() error {
	return p.Conn.Close()
}
