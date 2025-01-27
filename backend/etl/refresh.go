package etl

import (
	"database/sql"
	"fmt"
	"foo/backend/connections"
	"strconv"
	"time"
)

var (
	FullRefreshTime        time.Duration = time.Minute * 60
	IncrementalRefreshTime time.Duration = time.Minute * 5
)

func Refresh() {
	ProdConn := connections.ProdConn
	if ProdConn == nil {
		fmt.Printf("Error getting production database\n")
		return
	}

	fullRefresh(ProdConn)
	intialiseTimer(ProdConn)
}

func intialiseTimer(ProdConn *sql.DB) {
	fullRefreshTicker := time.NewTicker(FullRefreshTime)
	incrementalRefreshTicker := time.NewTicker(IncrementalRefreshTime)
	defer fullRefreshTicker.Stop()
	defer incrementalRefreshTicker.Stop()

	fullRefreshed := false
	for {
		select {
		case <-fullRefreshTicker.C:
			fullRefreshed = true
			fullRefresh(ProdConn)
		case <-incrementalRefreshTicker.C:
			if fullRefreshed {
				fullRefreshed = false
				incrementalRefreshTicker.Reset(IncrementalRefreshTime)
				continue
			}
			incrementalRefresh(ProdConn)
		}
	}
}

func fullRefresh(ProdConn *sql.DB) {

	postgresSources := connections.SourcesConn.PostgresDB
	sources := []string{}
	for _, source := range postgresSources {
		fullRefreshPostgres(source.Conn, ProdConn, "public."+source.Name, source.Name)
		sources = append(sources, source.Name)
	}

	refreshStagingTable(ProdConn, sources)
}

// To do: Implement the logic to refresh the staging table
func refreshStagingTable(ProdConn *sql.DB, sources []string) {

}

func incrementalRefresh(ProdConn *sql.DB) {
	postgresSources := connections.SourcesConn.PostgresDB
	for _, source := range postgresSources {
		incrementalRefreshPostgres(source.Conn, ProdConn, "public."+source.Name, source.Name)
	}

}
func incrementalRefreshPostgres(postgresConn *sql.DB, prodConn *sql.DB, srcTableName string, destTableName string) {
	getIncrementalRefreshCondition(postgresConn, prodConn, srcTableName, destTableName)
	partialRefreshPostgres(postgresConn, prodConn, srcTableName, destTableName)

}

// TODO: Implement the logic to get the incremental refresh condition
func getIncrementalRefreshCondition(postgresConn *sql.DB, prodConn *sql.DB, srcTableName string, destTableName string) {

}

// TODO: Implement the logic to partially refresh the data
func partialRefreshPostgres(postgresConn *sql.DB, prodConn *sql.DB, srcTableName string, destTableName string) {
}
func fullRefreshPostgres(postgresConn *sql.DB, prodConn *sql.DB, srcTableName string, destTableName string) {
	refreshRawData(postgresConn, prodConn, srcTableName, destTableName)
	/*At this point data would be inserted into the staging table to present to the user to make changes,
	and then the data would be moved to the raw schema table*/
}

func refreshRawData(postgresConn *sql.DB, prodConn *sql.DB, srcTableName string, destTableName string) {
	offset := 0
	limit := 1000
	for {
		strLimit := strconv.Itoa(limit)
		strOffset := strconv.Itoa(offset)
		rows, err := postgresConn.Query(fmt.Sprintf("SELECT * FROM %s LIMIT %s OFFSET %s;", srcTableName, strLimit, strOffset))
		if err != nil {
			fmt.Printf("Error querying source database: %v\n", err)
			return
		}
		if rows == nil {
			fmt.Printf("No rows returned\n")
			break
		}
		defer rows.Close()

		columns, err := rows.ColumnTypes()
		if err != nil {
			fmt.Printf("Error getting column types: %v\n", err)
			return
		}

		if tableExists(prodConn, destTableName, "staging") {
			fmt.Printf("Table %s already exists\n", destTableName)
		} else {
			createTable(prodConn, columns, destTableName, "staging")
		}

		tx, err := prodConn.Begin()
		if err != nil {
			fmt.Printf("Error beginning transaction: %v\n", err)
			return
		}
		defer tx.Rollback()

		counter := 1
		values := ""
		for rows.Next() {
			values += "("
			for i := range columns {
				if i > 0 {
					values += fmt.Sprintf("$%d", counter)
					counter++
				}
				if i < len(columns)-1 {
					values += ","
				}
			}
			values += "),"
		}
		query := fmt.Sprintf("INSERT INTO staging.%s (%s) VALUES %s", destTableName, columns, values)
		args := make([]interface{}, len(columns))
		for rows.Next() {
			err := rows.Scan(args...)
			if err != nil {
				fmt.Printf("Error scanning row: %v\n", err)
				return
			}
		}
		_, err = tx.Exec(query, args...)
		if err != nil {
			fmt.Printf("Error inserting data: %v\n", err)
			return
		}

		err = tx.Commit()
		if err != nil {
			fmt.Printf("Error committing transaction: %v\n", err)
			return
		}
		fmt.Printf("Data inserted successfully\n")
	}
}

func createTable(prodDB *sql.DB, columnstype []*sql.ColumnType, tableName, schema string) {
	tableColumns := ""
	for i, col := range columnstype {
		if i > 0 {
			tableColumns += ", "
		}
		tableColumns += fmt.Sprintf("%s %s", col, columnstype[i])
	}

	createTableQuery := fmt.Sprintf("CREATE TABLE %s.%s (%s)", schema, tableName, tableColumns)
	if _, err := prodDB.Exec(createTableQuery); err != nil {
		fmt.Printf("Error creating table: %v\n%v\n", err, createTableQuery)
	} else {
		fmt.Printf("Table %s created successfully\n", tableName)
	}
}

func tableExists(prodDB *sql.DB, tableName, schema string) bool {
	query := fmt.Sprintf("SELECT table_name FROM information_schema.tables WHERE table_schema = '%s' AND table_name = '%s'", schema, tableName)
	rows, err := prodDB.Query(query)
	if err != nil {
		fmt.Printf("Error querying database: %v\n", err)
		return false
	}
	defer rows.Close()
	return rows.Next()
}
