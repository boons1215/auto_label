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

var blue = color.New(color.FgBlue)
var red = color.New(color.FgRed)

// retrieve all new ven without app/env/loc labels
func GetNewVen(pce, orgId, apiUser, apiKey string, client *http.Client, wg *sync.WaitGroup, async bool) []Ven {
	defer wg.Done()

	url := pce + "/api/v2/orgs/" + orgId
	path := "/workloads?"
	path += "representation=workload_labels&managed=true&"
	path += "labels=[[\"/orgs/" + orgId + "/labels?key=app%26exists=false\",\"/orgs/" + orgId + "/labels?key=loc%26exists=false\",\"/orgs/" + orgId + "/labels?key=env%26exists=false\"]]"

	var newVen []Ven

	err := helper.GetJson(pce, orgId, path, "GET", apiUser, apiKey, client, &newVen, async)
	if err != nil {
		fmt.Printf("error getting ven data from pce: %s\n", err.Error())
		return nil
	}

	fmt.Println("* HTTP JSON URL:", url+path)
	blue.Printf("- Discovered %d new VENs on PCE without APP|ENV|LOC labels:\n", len(newVen))

	for _, v := range newVen {
		fmt.Printf("	- %s : %s\n", v.Hostname, v.Href)
	}

	fmt.Println()
	return newVen
}

// update ven label based on csv input
func UpdateVenLabel(pce, orgId, apiUser, apiKey string, client *http.Client, venInfo []Ven) {
	baseUrl := pce + "/api/v2/orgs/" + orgId + "/workloads/set_labels"
	fmt.Println()

	fmt.Println(baseUrl)
	for i := range venInfo {
		param := "{\"workloads\": [{\"href\": \"" + venInfo[i].Href + "\"}],"
		param += "\"labels\": ["
		param += "{\"href\": \"" + venInfo[i].App + "\"},"
		param += "{\"href\": \"" + venInfo[i].Env + "\"},"
		param += "{\"href\": \"" + venInfo[i].Loc + "\"}],"
		param += "\"delete_existing_keys\": []}"

		body := []byte(param)
		fmt.Println("> ", venInfo[i].Hostname, venInfo[i].Href)

		resp, _, err := helper.UpdateJson(baseUrl, "PUT", apiUser, apiKey, body, client)
		if err != nil {
			fmt.Printf("Error getting data from pce: %s\n", err.Error())
			return
		}

		if resp.StatusCode == http.StatusOK {
			fmt.Println("Response: ", resp.StatusCode)
		} else {
			red.Println("Failed with error: ", resp.Status)
			red.Println("Verify this VEN is in the excelsheet or missing labels info in the excelsheet")
		}
	}
}
