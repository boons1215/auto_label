package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/boons1215/auto-label/helper"
	"github.com/boons1215/auto-label/label"
	"github.com/boons1215/auto-label/output"
	"github.com/boons1215/auto-label/util"
	"github.com/boons1215/auto-label/ven"
)

func main() {
	// flags
	var file string
	flag.StringVar(&file, "f", "", "import the csv file")
	flag.Parse()

	if len(file) == 0 {
		fmt.Println("Usage: auto-label -f data.csv")
		flag.PrintDefaults()
		os.Exit(1)
	}

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}

	var (
		client   *http.Client
		pce      = config.PCE
		id       = config.OrgId
		user     = config.ApiUser
		key      = config.ApiKey
		report   = os.Args[2]
		fixedLoc = config.FixedLocLabel
		ready    = false
		async    = false
	)

	client = &http.Client{
		Timeout: 600 * time.Second,
	}

	var wg sync.WaitGroup
	wg.Add(2)
	// retrieve data from raw csv report which provided by user
	raw := helper.GetCsv(report, &wg)

	// retrieve new VEN without labels from PCE
	newVen := ven.GetNewVen(pce, id, user, key, client, &wg, async)
	wg.Wait()

	fmt.Println("Total VEN discovered: ", len(newVen))

	// retrieve the existing defined labels from PCE
	pceLabelInfo := label.GetAllLabels(pce, id, user, key, client, async)

	// prepare a csv draft report for user
	recordExistData, recordNotFound := output.PrepareCsvData(newVen, raw, fixedLoc)

	output.ConsolidateCsv(recordExistData, "recordExist")
	output.ConsolidateCsv(recordNotFound, "recordNotFound")

	// yes or no for next steps
	fmt.Println()
	confirm := util.ShallProceed("Proceed for new label check?")
	if confirm {
		fmt.Println()
		// create two new slices for collecting labels in the raw csv report
		appRawLabel := []string{}
		envRawLabel := []string{}

		for _, v := range recordExistData[1:] {
			appRawLabel = append(appRawLabel, v[2])
			envRawLabel = append(envRawLabel, v[3])
		}

		// compare the label set, output the label name which is not found in the pce as slice
		appLabelAsPerReport := label.UniqueSlice(appRawLabel)
		envLabelAsPerReport := label.UniqueSlice(envRawLabel)

		newAppLabel := label.UniqueLabel(appLabelAsPerReport, pceLabelInfo, "app")
		newEnvLabel := label.UniqueLabel(envLabelAsPerReport, pceLabelInfo, "env")

		fmt.Printf("%d - New App Labels to be created: %s \n", len(newAppLabel), newAppLabel)
		fmt.Printf("%d - New Env Labels to be created: %s \n", len(newEnvLabel), newEnvLabel)
		fmt.Println()

		// create new labels steps
		if len(newAppLabel) == 0 && len(newEnvLabel) == 0 {
			fmt.Println("No new labels need to be created.")
			fmt.Println()
			ready = true
		} else {
			fmt.Println()
			confirm := util.ShallProceed("Create these new labels on PCE?")
			fmt.Println()

			if confirm {
				label.CreateNewLabels(pce, id, user, key, "app", newAppLabel, client)
				label.CreateNewLabels(pce, id, user, key, "env", newEnvLabel, client)
			} else {
				os.Exit(0)
			}

			// retrieve the existing defined labels from PCE
			pceLabelInfo = label.GetAllLabels(pce, id, user, key, client, async)
			ready = true
		}

		// filter VENs that not listed in the csv record as a new 2D slice
		updatedVENList := [][]string{}

		for j := range newVen {
			ven := &newVen[j]
			for i := range recordExistData {
				if util.Normalise(ven.Hostname) == util.Normalise(recordExistData[i][1]) {
					row := []string{ven.Href, ven.Hostname, ven.App, ven.Env, ven.Loc}
					updatedVENList = append(updatedVENList, row)
				}
			}
		}

		// labelling stage
		if ready {
			fmt.Println()
			confirm = util.ShallProceed("Ready for labelling new VENs?")

			if confirm {
				// collect label href
				collector := make(map[string]string)

				label.MapLabelHref(collector, appLabelAsPerReport, pceLabelInfo, "app")
				label.MapLabelHref(collector, envLabelAsPerReport, pceLabelInfo, "env")

				// location label is fixed as "SGP"
				for _, v := range pceLabelInfo {
					if v.Value == fixedLoc && v.Key == "loc" {
						collector[v.Value] = v.Href
					}
				}

				// map the app/env/loc label href to the workloads
				for i := range updatedVENList {
					updatedVENList[i][2] = collector[updatedVENList[i][2]]
					updatedVENList[i][3] = collector[updatedVENList[i][3]]
					updatedVENList[i][4] = collector[updatedVENList[i][4]]
				}

				ven.UpdateVenLabel(pce, id, user, key, client, updatedVENList)
			}
		}

		// enforce the ven based on condition:
		// UAT env - direct enforce
		// PROD env - check application_category, only enforce cat 3/4
		fmt.Println()
		confirm = util.ShallProceed("Ready for enforcing the VENs?")

		if confirm {
			envLabelHref := make(map[string]string)

			for i := range pceLabelInfo {
				if pceLabelInfo[i].Key == "env" && pceLabelInfo[i].Value == "UAT" {
					envLabelHref[pceLabelInfo[i].Value] = pceLabelInfo[i].Href
				}

				if pceLabelInfo[i].Key == "env" && pceLabelInfo[i].Value == "PRODUCTION" {
					envLabelHref[pceLabelInfo[i].Value] = pceLabelInfo[i].Href
				}
			}

			ven.EnforceVen(pce, id, user, key, client, updatedVENList, recordExistData, envLabelHref)
		}

		return
	}
}
