package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/lib/pq"
	"github.com/xuri/excelize/v2"
)

/*
MSSQL:

INSERT INTO sensor_data (sensor_id, product_id, class_rate, belimish_rate)
VALUES (1, 101, 95.75, 0.25);

postgres:

	INSERT INTO public.production_data (
		production_id, machine_id, product_id, time_start, time_elapsed, next_machine_id, status, station_id
	) VALUES (
		nextval('machine_data_id_seq'), 'machine_1', 101, '2023-10-01', '2023-10-02', 'machine_2', 1, 'station_1'
	);
*/

type shift_handover struct {
	name        string
	shift_start string
	shift_end   string
	shift_name  string
	shift_type  string
}

type sensorData struct {
	sensorID     int
	productID    int
	classRate    float64
	belimishRate float64
}

type productionData struct {
	productionID  int
	machineID     string
	productID     int
	timeStart     string
	timeElapsed   string
	nextMachineID string
	status        int
	stationID     string
}

var entries = 10000

func populateExcel(data []shift_handover) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	f.SetCellValue("Sheet1", "A1", "Name")
	f.SetCellValue("Sheet1", "B1", "Shift Start")
	f.SetCellValue("Sheet1", "C1", "Shift End")
	f.SetCellValue("Sheet1", "D1", "Shift Name")
	f.SetCellValue("Sheet1", "E1", "Shift Type")

	for i, entry := range data {
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), entry.name)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), entry.shift_start)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), entry.shift_end)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+2), entry.shift_name)
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", i+2), entry.shift_type)
	}

	if err := f.SaveAs("test_data/shift_handover.xlsx"); err != nil {
		fmt.Println(err)
	}
}

func getData() []shift_handover {
	names := []string{"John Doe", "Jane Doe", "John Smith", "Jane Smith", "Bob Doe", "Bob Smith"}
	shift_types := []string{"Morning", "Afternoon", "Night"}
	shift_names := []string{"Blue", "Green", "Yellow"}
	shift_start := map[string]string{
		"Blue":   "06:00",
		"Green":  "14:00",
		"Yellow": "22:00",
	}
	shift_end := map[string]string{
		"Blue":   "14:00",
		"Green":  "22:00",
		"Yellow": "06:00",
	}

	var data []shift_handover
	now := time.Now()
	for i := 0; i < 14; i++ { // past 2 weeks
		date := now.AddDate(0, 0, -i)
		for _, shift_name := range shift_names {
			name := names[rand.Intn(len(names))]
			start_time := shift_start[shift_name]
			end_time := shift_end[shift_name]
			start_datetime := date.Format("2006-01-02") + " " + start_time
			end_datetime := date.Format("2006-01-02") + " " + end_time
			if shift_name == "Yellow" {
				end_datetime = date.AddDate(0, 0, 1).Format("2006-01-02") + " " + end_time
			}
			data = append(data, shift_handover{
				name:        name,
				shift_start: start_datetime,
				shift_end:   end_datetime,
				shift_name:  shift_name,
				shift_type:  shift_types[rand.Intn(len(shift_types))],
			})
		}
	}
	return data
}

func generateProductionData() []productionData {
	var data []productionData

	now := time.Now()
	twoWeeksAgo := now.AddDate(0, 0, -14)

	product_id := 4

	station_map := map[string][]string{
		"station_1": {"machine_1", "machine_2", "machine_3"},
		"station_2": {"machine_4", "machine_5", "machine_6"},
		"station_3": {"machine_7", "machine_8", "machine_9", "machine_10"},
	}
	stations := []string{}
	for key := range station_map {
		stations = append(stations, key)
	}

	for i := range entries {

		timeStart := twoWeeksAgo.Add(time.Duration(rand.Int63n(int64(now.Sub(twoWeeksAgo)))))
		timeElapsed := timeStart.Add(time.Duration(rand.Intn(300)+1) * time.Second) // 1 to 300 seconds (5 minutes max)
		product_id += rand.Intn(5) + 1

		station := stations[rand.Intn(len(stations))]
		machine_list := station_map[station]
		machine := machine_list[rand.Intn(len(machine_list))]
		data = append(data, productionData{
			productionID:  i + 1,
			machineID:     fmt.Sprintf("%s", machine),
			productID:     product_id,
			timeStart:     timeStart.Format("2006-01-02 15:04:05"),
			timeElapsed:   timeElapsed.Format("2006-01-02 15:04:05"),
			nextMachineID: fmt.Sprintf("%d", rand.Intn(10)+1),
			status:        rand.Intn(2),
			stationID:     fmt.Sprintf("%s", station),
		})
	}
	return data
}

func insertProductionData(db *sql.DB, data []productionData) {
	_, err := db.Exec("DELETE FROM production_data")
	if err != nil {
		fmt.Printf("Error deleting table: %v\n", err)
	}
	for _, entry := range data {
		_, err := db.Exec("INSERT INTO public.production_data (production_id, machine_id, product_id, time_start, time_elapsed, next_machine_id, status, station_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
			entry.productionID, entry.machineID, entry.productID, entry.timeStart, entry.timeElapsed, entry.nextMachineID, entry.status, entry.stationID)
		if err != nil {
			log.Fatal("Error during Query: ", err)
		}
	}
}

func insertSensorData(db *sql.DB, data []sensorData) {
	_, err := db.Exec("DELETE FROM sensor_data")
	if err != nil {
		log.Fatal("Error during deletion query: ", err)
	}
	for _, entry := range data {
		query := `INSERT INTO sensor_data (sensor_id, product_id, class_rate, belimish_rate) VALUES (@p1, @p2, @p3, @p4);`

		_, err := db.Exec(query, entry.sensorID, entry.productID, entry.classRate, entry.belimishRate)
		if err != nil {
			log.Fatalf("Failed to execute query: %v", err)
		}
		if err != nil {
			log.Fatal("Error during Query: ", err, entry)
		}
	}
}
func generateSensorData(p_data []productionData) []sensorData {
	all_p_id := make([]int, len(p_data))
	for i, val := range p_data {
		all_p_id[i] = val.productID
	}
	var data []sensorData
	for i := 0; i < entries*10; i++ {
		product_id := all_p_id[rand.Intn(len(all_p_id))]
		data = append(data, sensorData{
			sensorID:     i + 1,
			productID:    product_id,
			classRate:    rand.Float64() * 100,
			belimishRate: rand.Float64(),
		})
	}

	return data
}

func populatePostgres() []productionData {

	pgUser := "postgres"
	pgPassword := "Week7890"
	pgDBName := "prototype-1"
	pgHost := "localhost"
	pgPort := "5432"
	pgSSLMode := "disable"

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=%s",
		pgUser, pgPassword, pgDBName, pgHost, pgPort, pgSSLMode)
	dbPostgres, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer dbPostgres.Close()
	productionData := generateProductionData()
	insertProductionData(dbPostgres, productionData)

	return productionData

}

func populateMssql(p_data []productionData) {
	conn_str := "server=localhost\\SQLEXPRESS;database=master;trusted_connection=yes"
	dbMSSQL, err := sql.Open("sqlserver", conn_str)
	fmt.Println("Connected to MSSQL database.")
	if err != nil {
		log.Fatal("Error Connecting: ", err)
	}
	defer dbMSSQL.Close()
	sensorData := generateSensorData(p_data)
	insertSensorData(dbMSSQL, sensorData)

}
func doTestData() {
	rand.Seed(time.Now().UnixNano())
	p_data := populatePostgres()

	populateMssql(p_data)

	shiftData := getData()
	populateExcel(shiftData)

	fmt.Println("Data inserted into databases and Excel file populated with shift handover data.")
}
