package output

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/boons1215/auto-label/helper"
	"github.com/boons1215/auto-label/ven"
)

// process the input and prepare data in csv format
func PrepareCsvData(newVen []ven.Ven, raw []helper.Workload, fixedLoc string) [][]string {
	data := [][]string{
		{"href", "hostname", "app", "env", "loc"},
	}

	for i := 0; i < len(newVen); i++ {
		ven := &newVen[i]
		for _, r := range raw {
			if ven.Hostname == r.Hostname {
				ven.App = r.App
				ven.Env = r.Env
				ven.Loc = fixedLoc
				if ven.Env == "PROD" {
					ven.Env = strings.Replace(ven.Env, "PROD", "PRODUCTION", -1)
				}
			}
		}
		adata := [][]string{
			{ven.Href, ven.Hostname, ven.App, ven.Env, ven.Loc},
		}
		data = append(data, adata...)
	}

	return data
}

// process the csv data and generate csv report
func ConsolidateCsv(data [][]string) {
	outputFileName := fmt.Sprintf("report-%s.csv", time.Now().Format("20060102_150405"))
	csvFile, err := os.Create(outputFileName)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	csvWriter := csv.NewWriter(csvFile)

	for _, row := range data {
		_ = csvWriter.Write(row)
	}

	fmt.Printf("* CSV report generated: %s", outputFileName)
	fmt.Println()
	fmt.Println()
	csvWriter.Flush()
	csvFile.Close()
}
