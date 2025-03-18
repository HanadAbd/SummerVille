package simData

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

func HandleCSV(c *DataSources, row [][]interface{}) error {
	writer := c.ConnectionPoint.(*csv.Writer)
	for _, r := range row {
		strRow := make([]string, len(r))
		for i, val := range r {
			strRow[i] = fmt.Sprint(val)
		}
		if err := writer.Write(strRow); err != nil {
			return fmt.Errorf("writing row: %w", err)
		}
	}
	writer.Flush()
	return nil
}

func HandlePostgres(c *DataSources, row [][]interface{}) error {
	db := c.ConnectionPoint.(*sql.DB)
	tableName := strings.ToLower(c.Name)

	for _, r := range row {
		placeholders := make([]string, len(r))
		for i := range r {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
		}

		insertSQL := fmt.Sprintf("INSERT INTO %s VALUES (%s)", tableName, strings.Join(placeholders, ", "))

		if _, err := db.Exec(insertSQL, r...); err != nil {
			return fmt.Errorf("inserting row: %w", err)
		}
	}
	return nil
}

func HandleKafka(c *DataSources, row [][]interface{}) error {
	producer := c.ConnectionPoint.(sarama.SyncProducer)
	for _, r := range row {
		msg := &sarama.ProducerMessage{
			Topic: c.Name,
			Key:   sarama.StringEncoder(fmt.Sprint(r[0])),
			Value: sarama.StringEncoder(fmt.Sprint(r[1:])),
		}

		if _, _, err := producer.SendMessage(msg); err != nil {
			return fmt.Errorf("sending message: %w", err)
		}
	}
	return nil
}

func intialiseSource(c *DataSources, header []string) {
	var initializer DataSourceInitializer

	switch c.DataType {
	case CSV:
		if filepath, ok := c.ConnectionDetails.(string); ok {
			initializer = &CSVInitializer{filepath: filepath}
		}
	case Postgres:
		if connStr, ok := c.ConnectionDetails.(string); ok {
			initializer = &PostgresInitializer{connStr: connStr, name: c.Name}
		}
	case Kafka:
		if producer, ok := c.ConnectionDetails.(*KafkaInitializer); ok {
			initializer = &KafkaInitializer{broker: producer.broker, topic: producer.topic}
		}
	}

	if initializer != nil {
		if err := initializer.Initialize(header, c); err != nil {
			logging("Error initializing %s: %v\n", c.DataType, err)
			return
		}
	}

	c.Data = &Data{Header: header}
}

type DataSourceInitializer interface {
	Initialize(header []string, connection *DataSources) error
}

type CSVInitializer struct {
	filepath string
}

func (c *CSVInitializer) Initialize(header []string, connection *DataSources) error {
	file, err := os.OpenFile(c.filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("creating CSV file: %w", err)
	}

	writer := csv.NewWriter(file)
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("writing header: %w", err)
	}
	writer.Flush()

	connection.ConnectionPoint = writer
	return nil
}

type PostgresInitializer struct {
	connStr string
	name    string
}

func (p *PostgresInitializer) Initialize(header []string, connection *DataSources) error {
	db, err := sql.Open("postgres", p.connStr)
	if err != nil {
		return fmt.Errorf("connecting to Postgres: %w", err)
	}

	columns := make([]string, len(header))
	for i, h := range header {
		columns[i] = fmt.Sprintf("%s TEXT", strings.ToLower(h))
	}

	tableName := strings.ToLower(p.name)
	if _, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)); err != nil {
		return fmt.Errorf("dropping table: %w", err)
	}

	createSQL := fmt.Sprintf("CREATE TABLE %s (%s)", tableName, strings.Join(columns, ", "))
	if _, err := db.Exec(createSQL); err != nil {
		return fmt.Errorf("creating table: %w", err)
	}
	connection.ConnectionPoint = db

	return nil
}

type KafkaInitializer struct {
	broker string
	topic  string
}

func (k *KafkaInitializer) Initialize(header []string, connection *DataSources) error {
	if k.broker == "" {
		return fmt.Errorf("empty Kafka broker address")
	}

	log.Printf("Attempting to connect to Kafka broker at %s", k.broker)

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	var producer sarama.SyncProducer
	var err error

	maxRetries := 5
	retryDelay := time.Second * 2

	for attempt := 1; attempt <= maxRetries; attempt++ {
		producer, err = sarama.NewSyncProducer([]string{k.broker}, config)
		if err == nil {
			break
		}

		if attempt == maxRetries {
			return fmt.Errorf("failed to connect to Kafka after %d attempts: %w", maxRetries, err)
		}

		log.Printf("Failed to connect to Kafka broker %s (attempt %d/%d): %v. Retrying in %v...",
			k.broker, attempt, maxRetries, err, retryDelay)
		time.Sleep(retryDelay)
	}

	connection.ConnectionPoint = producer
	return nil
}
