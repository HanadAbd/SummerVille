package etl

import (
	_ "github.com/denisenkom/go-mssqldb"
)

// func fullRefreshMSSQL(mssqlConn *sql.DB, prodConn *sql.DB, srcTableName string, destTableName string) {
// 	rows, err := mssqlConn.Query(fmt.Sprintf("SELECT * FROM %s", srcTableName))
// 	if err != nil {
// 		fmt.Printf("Error querying source database: %v\n", err)
// 		return
// 	}
// 	defer rows.Close()

// 	columns, err := rows.Columns()
// 	if err != nil {
// 		fmt.Printf("Error getting columns: %v\n", err)
// 		return
// 	}

// 	tx, err := prodConn.Begin()
// 	if err != nil {
// 		fmt.Printf("Error beginning transaction: %v\n", err)
// 		return
// 	}
// 	defer tx.Rollback()

// 	insertStmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", destTableName,
// 		strings.Join(columns, ", "), strings.Repeat("?, ", len(columns)-1)+"?")
// 	for rows.Next() {
// 		values := make([]interface{}, len(columns))
// 		valuePtrs := make([]interface{}, len(columns))
// 		for i := range values {
// 			valuePtrs[i] = &values[i]
// 		}

// 		if err := rows.Scan(valuePtrs...); err != nil {
// 			fmt.Printf("Error scanning row: %v\n", err)
// 			return
// 		}

// 		if _, err := tx.Exec(insertStmt, values...); err != nil {
// 			fmt.Printf("Error inserting data: %v\n", err)
// 			return
// 		}
// 	}

// 	if err := tx.Commit(); err != nil {
// 		fmt.Printf("Error committing transaction: %v\n", err)
// 		return
// 	}
// 	fmt.Println("MSSQL data refreshed")
// }
