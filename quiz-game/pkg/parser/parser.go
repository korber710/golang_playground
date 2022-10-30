package parser

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
)

func ParseProblemFile(filename string) (datas [][]string, err error) {
	// Open file
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	// Close the file after
	defer f.Close()

	// Read csv values using csv.Reader
	csvReader := csv.NewReader(f)
	datas, err = csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("data:", datas)

	return datas, err
}
