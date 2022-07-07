package output

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/boons1215/auto-label/helper"
	"github.com/boons1215/auto-label/util"
	"github.com/boons1215/auto-label/ven"
	"github.com/fatih/color"
)

var (
	green = color.New(color.FgHiGreen)
)

// process the input and prepare data in csv format
func PrepareCsvData(newVen []ven.Ven, raw []helper.Workload) ([][]string, [][]string) {
	// ven find in the csv record
	recordExistData := [][]string{
		{"href", "hostname", "app", "env", "loc", "category"},
	}

	// ven not found in the csv record
	recordNotFound := [][]string{
		{"href", "hostname", "app", "env", "loc", "category"},
	}

	var inRecord string

	for i := 0; i < len(newVen); i++ {
		ven := &newVen[i]
		for _, r := range raw {
			if util.Normalise(ven.Hostname) == util.Normalise(r.Hostname) {
				ven.App = r.App
				ven.Env = strings.ToUpper(r.Env)

				if r.Env == "UAT" {
					ven.Loc = strings.ToUpper("Singapore")
				} else {
					ven.Loc = strings.ToUpper("SGP")
				}

				ven.Category = r.Category
				if ven.Env == "PROD" {
					ven.Env = strings.Replace(ven.Env, "PROD", "PRODUCTION", -1)
				}
				adata := [][]string{
					{ven.Href, ven.Hostname, ven.App, ven.Env, ven.Loc, ven.Category},
				}
				recordExistData = append(recordExistData, adata...)
				inRecord = ven.Hostname
			}
		}

		if newVen[i].Hostname != inRecord {
			ndata := [][]string{
				{ven.Href, ven.Hostname},
			}
			recordNotFound = append(recordNotFound, ndata...)
		}
	}

	return recordExistData, recordNotFound
}

// process the csv data and generate csv report
func ConsolidateCsv(recordExistData [][]string, reportName string) {
	outputFileName := fmt.Sprintf("%s_ven-report-%s.csv", reportName, time.Now().Format("20060102_150405"))
	csvFile, err := os.Create(outputFileName)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	csvWriter := csv.NewWriter(csvFile)

	for _, row := range recordExistData {
		_ = csvWriter.Write(row)
	}

	green.Println("* CSV report generated: ", outputFileName)
	csvWriter.Flush()
	csvFile.Close()
}
