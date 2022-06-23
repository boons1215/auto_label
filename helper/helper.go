package helper

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gocarina/gocsv"
)

type Workload struct {
	Hostname string `csv:"HOSTNAME"`
	App      string `csv:"APPLICATION_CODE"`
	Env      string `csv:"ENVIRONMENT"`
	Category string `csv:"APPLICATION_CATEGORY"`
}

// Unexported struct for handling the asyncResults
type asyncResults struct {
	Href        string `json:"href"`
	JobType     string `json:"job_type"`
	Description string `json:"description"`
	Result      struct {
		Href string `json:"href"`
	} `json:"result"`
	Status       string `json:"status"`
	RequestedAt  string `json:"requested_at"`
	TerminatedAt string `json:"terminated_at"`
	RequestedBy  struct {
		Href string `json:"href"`
	} `json:"requested_by"`
}

var (
	req *http.Request
	err error
)

// formalise json data into struct interface
func GetJson(pce, orgId, path, method, apiUser, apiKey string, client *http.Client, target interface{}, async bool) error {
	var asyncResults asyncResults
	baseURL := pce + "/api/v2/orgs/" + orgId

	req, err = http.NewRequest(method, baseURL+path, nil)
	if err != nil {
		fmt.Printf("Error in url: %s\n", err.Error())
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")
	if async {
		req.Header.Set("Prefer", "respond-async")
	}

	req.SetBasicAuth(apiUser, apiKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %s", err)
	}

	defer resp.Body.Close()

	if async {
		for asyncResults.Status != "done" {
			asyncResults, err = polling(pce, apiUser, apiKey, resp)
			if err != nil {
				log.Fatal("Async query failed")
			}
		}

		finalReq, err := http.NewRequest("GET", pce+"/api/v2"+asyncResults.Result.Href, nil)
		if err != nil {
			fmt.Printf("Error in url: %s\n", err.Error())
		}

		fmt.Println("> async datafile: ", finalReq.URL)

		finalReq.SetBasicAuth(apiUser, apiKey)
		finalReq.Header.Set("User-Agent", "Mozilla/5.0")
		finalReq.Header.Set("Content-Type", "application/json")

		resp, err = client.Do(finalReq)
		if err != nil {
			fmt.Printf("Error: %s", err)
		}

		defer resp.Body.Close()
		return json.NewDecoder(resp.Body).Decode(target)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

// polling is used in async requests to check when the data is ready
func polling(url, apiUser, apiKey string, resp *http.Response) (asyncResults, error) {
	var asyncResults asyncResults
	url += "/api/v2"

	// Create HTTP client and request
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	pollReq, err := http.NewRequest("GET", url+resp.Header.Get("Location"), nil)
	if err != nil {
		return asyncResults, err
	}

	fmt.Println("> async job: ", pollReq.URL)

	// Set basic authentication and headers
	pollReq.SetBasicAuth(apiUser, apiKey)
	pollReq.Header.Set("Content-Type", "application/json")
	pollReq.Header.Set("User-Agent", "Mozilla/5.0")

	// Wait for recommended time from Retry-After
	wait, err := strconv.Atoi(resp.Header.Get("Retry-After"))
	if err != nil {
		return asyncResults, err
	}
	duration := time.Duration(wait) * time.Second
	time.Sleep(duration)

	// Check if the data is ready
	pollResp, err := client.Do(pollReq)
	if err != nil {
		return asyncResults, err
	}

	// Process Response
	data, err := ioutil.ReadAll(pollResp.Body)
	if err != nil {
		return asyncResults, err
	}

	// Put relevant response info into struct
	json.Unmarshal(data[:], &asyncResults)

	return asyncResults, err
}

// update api function for put/post
func UpdateJson(url, method, apiUser, apiKey string, body []byte, client *http.Client) (*http.Response, []byte, error) {
	req, err = http.NewRequest(method, url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.SetBasicAuth(apiUser, apiKey)

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)

	return resp, respBody, err
}

// read raw report and unmarshall for workload struct
func GetCsv(file string, wg *sync.WaitGroup) []Workload {
	defer wg.Done()

	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Printf("Unable to open csv: %s", err)
	}

	var workloads []Workload

	err = gocsv.UnmarshalBytes(bytes, &workloads)
	if err != nil {
		fmt.Printf("Unmarshal csv error: %s", err)
	}

	return workloads
}
