package ven

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/boons1215/auto-label/helper"
	"github.com/fatih/color"
)

type Ven struct {
	Href     string `json:"href"`
	Hostname string `json:"hostname"`
	App      string
	Env      string
	Loc      string
}

var (
	red   = color.New(color.FgHiRed)
	green = color.New(color.FgHiGreen)
)

// retrieve all new ven without app/env/loc labels
func GetNewVen(pce, orgId, apiUser, apiKey string, client *http.Client, wg *sync.WaitGroup, async bool) []Ven {
	defer wg.Done()

	url := pce + "/api/v2/orgs/" + orgId
	path := "/workloads?"
	path += "representation=workload_labels&managed=true&"
	//path += "representation=workload_labels&"
	path += "labels=[[\"/orgs/" + orgId + "/labels?key=app%26exists=false\",\"/orgs/" + orgId + "/labels?key=loc%26exists=false\",\"/orgs/" + orgId + "/labels?key=env%26exists=false\"]]"

	var newVen []Ven

	err := helper.GetJson(pce, orgId, path, "GET", apiUser, apiKey, client, &newVen, async)
	if err != nil {
		red.Printf("error getting ven data from pce: %s\n", err.Error())
		return nil
	}

	fmt.Println("* HTTP JSON URL:", url+path)

	if len(newVen) >= 500 {
		red.Println("- Returned the filtered VEN data from PCE is more than 500, fetching with async query...")
		err := helper.GetJson(pce, orgId, path, "GET", apiUser, apiKey, client, &newVen, true)
		if err != nil {
			red.Printf("error getting ven data from pce: %s\n", err.Error())
			return nil
		}
	}

	fmt.Println()
	green.Println("- Discovered new VENs on PCE without APP|ENV|LOC labels: ", len(newVen))

	for _, v := range newVen {
		fmt.Printf("	- %s : %s\n", v.Hostname, v.Href)
	}

	fmt.Println()
	return newVen
}

// update ven label based on csv input
func UpdateVenLabel(pce, orgId, apiUser, apiKey string, client *http.Client, venInfo [][]string) {
	baseUrl := pce + "/api/v2/orgs/" + orgId + "/workloads/bulk_update"
	fmt.Println()

	var param string
	fmt.Println(baseUrl)
	for i := range venInfo {
		param += "{\"href\": \"" + venInfo[i][0] + "\","
		param += "\"labels\": ["
		param += "{\"href\": \"" + venInfo[i][2] + "\"},"
		param += "{\"href\": \"" + venInfo[i][3] + "\"},"
		param += "{\"href\": \"" + venInfo[i][4] + "\"}]},"
	}
	body := []byte("[" + param + "{}]")

	fmt.Println("Labelling total VENs: ", len(venInfo))
	resp, _, err := helper.UpdateJson(baseUrl, "PUT", apiUser, apiKey, body, client)
	if err != nil {
		fmt.Printf("Error getting data from pce: %s\n", err.Error())
		return
	}

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Completed! -> Response: ", resp.StatusCode)
	} else {
		red.Println("Failed with error: ", resp.Status)
		red.Println("Verify this VEN is in the excelsheet or missing labels info in the excelsheet")
	}
}
