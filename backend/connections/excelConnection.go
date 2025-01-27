package connections

import (
	"fmt"
	"os"

	"github.com/xuri/excelize/v2"
)

func InitExcel(filePath string) (*excelize.File, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open the file: %v", err)
	}
	defer file.Close()

	excelFile, err := excelize.OpenReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open the excel file: %v", err)
	}
	fmt.Println("Successfully connected to the Excel File!")
	return excelFile, nil
}
