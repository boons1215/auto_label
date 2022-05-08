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

var client *http.Client

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
		Timeout: 10 * time.Second,
	}

	var wg sync.WaitGroup
	wg.Add(2)
	// retrieve data from raw csv report which provided by user
	raw := helper.GetCsv(report, &wg)

	// retrieve new VEN without labels from PCE
	newVen := ven.GetNewVen(pce, id, user, key, client, &wg, async)
	wg.Wait()

	// retrieve the existing defined labels from PCE
	pceLabelInfo := label.GetAllLabels(pce, id, user, key, client, async)

	// prepare a csv draft report for user
	data := output.PrepareCsvData(newVen, raw, fixedLoc)
	output.ConsolidateCsv(data)

	// yes or no for next steps
	fmt.Println()
	confirm := util.ShallProceed("Proceed for new label check?")
	if confirm {
		fmt.Println()
		// create two new slices for collecting labels in the raw csv report
		appRawLabel := []string{}
		envRawLabel := []string{}

		for _, v := range data[1:] {
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

		// ready for labelling new VEN
		if ready {
			fmt.Println()
			confirm = util.ShallProceed("Ready for labelling new VENs?")

			if confirm {
				collector := make(map[string]string)

				label.MapLabelHref(collector, appLabelAsPerReport, pceLabelInfo, "app")
				label.MapLabelHref(collector, envLabelAsPerReport, pceLabelInfo, "env")

				// location label is fixed as "SGP"
				for _, v := range pceLabelInfo {
					if v.Value == "SGP" && v.Key == "loc" {
						collector[v.Value] = v.Href
					}
				}

				// map the href to the workloads
				for i := range newVen {
					n := &newVen[i]
					n.App = collector[n.App]
					n.Env = collector[n.Env]
					n.Loc = collector[n.Loc]
				}

				ven.UpdateVenLabel(pce, id, user, key, client, newVen)
			}
		}

		return
	}
}
