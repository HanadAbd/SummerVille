package connections

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type CSVConn struct {
	Name       string
	FilePath   string
	Reader     *csv.Reader
	Writer     *csv.Writer
	File       *os.File
	FileInfo   os.FileInfo
	Metrics    ConnectionMetrics
	Connected  bool
	HasHeaders bool
	Headers    []string
}

type CSVCredential struct {
	FilePath  string
	Encoding  string
	HasHeader bool
}

func (c *CSVConn) InitCSV() error {
	dir := filepath.Dir(c.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for CSV file: %w", err)
	}

	_, err := os.Stat(c.FilePath)
	fileExists := !os.IsNotExist(err)

	file, err := os.OpenFile(c.FilePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to get file info: %w", err)
	}

	reader := csv.NewReader(file)
	writer := csv.NewWriter(file)

	c.File = file
	c.Reader = reader
	c.Writer = writer
	c.FileInfo = fileInfo
	c.Connected = true

	if fileExists && fileInfo.Size() > 0 {
		headers, err := reader.Read()
		if err != nil {
			return fmt.Errorf("failed to read CSV headers: %w", err)
		}
		c.Headers = headers
		c.HasHeaders = true
	}

	return nil
}

func (c *CSVConn) RetryConnection(maxAttempts int, delay time.Duration) error {
	if c.Connected {
		return nil
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := c.InitCSV(); err != nil {
			lastErr = err
			log.Printf("CSV connection attempt %d/%d failed: %v", attempt, maxAttempts, err)
			time.Sleep(delay)
			continue
		}
		return nil
	}

	return fmt.Errorf("failed to reconnect to CSV file after %d attempts: %v", maxAttempts, lastErr)
}

func (c *CSVConn) MonitorConnection() ConnectionMetrics {
	if c.File == nil {
		c.Connected = false
		c.Metrics.Status = "disconnected"
		c.Metrics.LastError = fmt.Errorf("file handle is nil")
		c.Metrics.LastErrorTime = time.Now()
		return c.Metrics
	}

	fileInfo, err := os.Stat(c.FilePath)
	if err != nil {
		c.Connected = false
		c.Metrics.Status = "disconnected"
		c.Metrics.LastError = err
		c.Metrics.LastErrorTime = time.Now()
		return c.Metrics
	}

	if fileInfo.ModTime() != c.FileInfo.ModTime() {
		c.Metrics.LastInfo = "File has been modified since connection was established"
	}

	c.Connected = true
	c.Metrics.Status = "connected"
	return c.Metrics
}

func (c *CSVConn) AddData(table TableDefinition, data []interface{}) error {
	if !c.Connected || c.Writer == nil {
		if err := c.InitCSV(); err != nil {
			return fmt.Errorf("failed to initialize CSV connection: %w", err)
		}
	}

	startTime := time.Now()

	var headers []string
	if !c.HasHeaders && len(table.Columns) > 0 {
		headers = make([]string, len(table.Columns))
		for i, col := range table.Columns {
			headers[i] = col.Name
		}

		if err := c.Writer.Write(headers); err != nil {
			c.Metrics.LastError = err
			c.Metrics.LastErrorTime = time.Now()
			return fmt.Errorf("failed to write headers to CSV: %w", err)
		}

		c.Headers = headers
		c.HasHeaders = true
	} else {
		headers = c.Headers
	}

	// Process each data item
	for _, item := range data {
		var strRow []string

		// Handle map[string]interface{} (column:value format)
		if rowMap, isMap := item.(map[string]interface{}); isMap {
			strRow = make([]string, len(headers))
			for i, colName := range headers {
				if val, exists := rowMap[colName]; exists {
					strRow[i] = fmt.Sprint(val)
				} else {
					strRow[i] = "" // Empty string for missing values
				}
			}
		} else if row, isSlice := item.([]interface{}); isSlice {
			// Handle []interface{} (ordered values format)
			strRow = make([]string, len(row))
			for i, val := range row {
				strRow[i] = fmt.Sprint(val)
			}
		} else {
			return fmt.Errorf("data item is neither a map nor a slice of interface{}")
		}

		if err := c.Writer.Write(strRow); err != nil {
			c.Metrics.LastError = err
			c.Metrics.LastErrorTime = time.Now()
			return fmt.Errorf("failed to write data row to CSV: %w", err)
		}
	}

	c.Writer.Flush()
	if err := c.Writer.Error(); err != nil {
		c.Metrics.LastError = err
		c.Metrics.LastErrorTime = time.Now()
		return fmt.Errorf("failed to flush CSV data: %w", err)
	}

	c.Metrics.LastQueryTime = time.Since(startTime)
	c.Metrics.QueryCount++

	return nil
}

func NewCSVConn(name string) *CSVConn {
	conn := &CSVConn{
		Name:       name,
		FilePath:   filepath.Join("test_data", fmt.Sprintf("%s.csv", name)),
		Reader:     nil,
		Writer:     nil,
		File:       nil,
		FileInfo:   nil,
		Metrics:    getDefaultMetrics(),
		Connected:  false,
		HasHeaders: false,
		Headers:    []string{},
	}
	return conn
}

func (c *CSVConn) IntialiseData(table TableDefinition) error {
	if !c.Connected || c.Writer == nil {
		if err := c.InitCSV(); err != nil {
			return fmt.Errorf("failed to initialize CSV connection: %w", err)
		}
	}

	if c.FileInfo.Size() == 0 && len(table.Columns) > 0 {
		headers := make([]string, len(table.Columns))
		for i, col := range table.Columns {
			headers[i] = col.Name
		}

		if err := c.Writer.Write(headers); err != nil {
			return fmt.Errorf("failed to write headers to CSV: %w", err)
		}

		c.Writer.Flush()
		c.Headers = headers
		c.HasHeaders = true
	}

	return nil
}

func (c *CSVConn) PurgeData(table TableDefinition) error {
	if c.File != nil {
		c.File.Close()
	}

	if err := os.Remove(c.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove CSV file: %w", err)
	}

	return c.InitCSV()
}

func (c *CSVConn) PurgeAllData() error {
	return c.PurgeData(TableDefinition{})
}

func (c *CSVConn) CloseConnection() error {
	if c.Writer != nil {
		c.Writer.Flush()
	}

	if c.File != nil {
		err := c.File.Close()
		c.File = nil
		c.Connected = false
		return err
	}

	return nil
}
