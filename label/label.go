package label

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/boons1215/auto-label/helper"
	"github.com/fatih/color"
)

type Label struct {
	Href  string `json:"href"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type LabelBody struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var blue = color.New(color.FgBlue)
var red = color.New(color.FgRed)

// retrieve all the existing labels
func GetAllLabels(pce, orgId, apiUser, apiKey string, client *http.Client, async bool) []Label {
	baseURL := pce + "/api/v2/orgs/" + orgId
	path := "/labels"

	var label []Label

	err := helper.GetJson(pce, orgId, path, "GET", apiUser, apiKey, client, &label, async)
	if err != nil {
		fmt.Printf("error getting label data from pce: %s\n", err.Error())
		return nil
	}

	fmt.Println()

	// refetch if the return data is more than 500 from PCE
	if len(label) == 500 {
		blue.Println("Returned labels data from PCE is more than 500, fetching with async query...")
		err := helper.GetJson(pce, orgId, path, "GET", apiUser, apiKey, client, &label, true)
		if err != nil {
			fmt.Printf("error getting label data from pce: %s\n", err.Error())
			return nil
		}
	}

	fmt.Println("* HTTP JSON URL:", baseURL+path)
	blue.Printf("- Retrieved total number of %d labels from PCE\n", len(label))
	fmt.Println()
	return label
}

// returns a unique subset of the string slice provided
func UniqueSlice(input []string) []string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}

	return u
}

// return true when value not find in the target slice
func Contains(target []Label, str, ltype string) bool {
	for _, v := range target {
		if v.Key == ltype && v.Value == str {
			return false
		}
	}

	return true
}

// compare the label set, output the label name which is not found in the pce as slice
func UniqueLabel(labelSet []string, pceLabelSet []Label, labelType string) []string {
	newLabel := []string{}

	for _, value := range labelSet {
		if value != "" {
			t := Contains(pceLabelSet, value, labelType)
			if t {
				newLabel = append(newLabel, value)
			}
		}
	}

	return newLabel
}

// create new label on pce
func CreateNewLabels(pce, orgId, apiUser, apiKey, labelType string, labelSet []string, client *http.Client) {
	url := pce + "/api/v2/orgs/" + orgId + "/labels"
	fmt.Println()
	fmt.Println("* HTTP JSON URL:", url)

	if len(labelSet) != 0 {
		blue.Printf("Creating new %s labels ...\n", labelType)

		for _, value := range labelSet {
			newLabel := LabelBody{
				Key:   labelType,
				Value: value,
			}

			body, _ := json.Marshal(newLabel)

			resp, respBody, err := helper.UpdateJson(url, "POST", apiUser, apiKey, body, client)
			if err != nil {
				fmt.Printf("error getting label data from pce: %s\n", err.Error())
			}

			if resp.StatusCode == http.StatusCreated {
				fmt.Println("Response: ", string(respBody))
			} else {
				red.Println("Failed with error: ", resp.Status)
			}
		}
	} else {
		fmt.Println("No new labels created.")
	}
}

// map the label href by matching the label key and value
func MapLabelHref(mapHref map[string]string, labelSet []string, pceLabelInfo []Label, labelType string) {
	for _, v := range labelSet {
		for _, u := range pceLabelInfo {
			if v == u.Value && u.Key == labelType {
				mapHref[u.Value] = u.Href
			}
		}
	}
}
