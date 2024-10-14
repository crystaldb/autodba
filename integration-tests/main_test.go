package main

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"testing"
	"time"
)

var dbIdentifier string
var config Config

var testCases = []DbInfo{
	{
		Description:    "Test case 1",
		DbConnString:   "postgres://postgres:password@mohammad-dashti-rds-1.cvirkksghnig.us-west-2.rds.amazonaws.com:5432/postgres?sslmode=require",
		AwsRdsInstance: "mohammad-dashti-rds-1",
	},
	{
		Description:    "Postgres Version 13",
		DbConnString:   "postgres://postgres:rme49DKjpE4wwx16Bemu@radcliffe-1.c7mrowi2kiu4.us-east-1.rds.amazonaws.com:5432/postgres?sslmode=require",
		AwsRdsInstance: "radcliffe-1",
	},
}

func TestAPISuite(t *testing.T) {
	config, err := readConfig()
	if err != nil {
		t.Fatalf("Failed to read config: %v\n", err)
	}

	for _, testCase := range testCases {
		dbIdentifier = testCase.AwsRdsInstance

		if err := SetupTestContainer(config, testCase); err != nil {
			t.Fatalf("Failed to set up container for %s: %v\n", testCase.Description, err)
		}
		defer func() {
			if err := TearDownTestContainer(); err != nil {
				log.Printf("Failed to tear down container for %s: %v\n", testCase.Description, err)
			}
		}()

		t.Run(testCase.Description, TestAPIRequest)

		if err := TearDownTestContainer(); err != nil {
			log.Printf("Failed to tear down container for %s: %v\n", testCase.Description, err)
		}
	}
}

func TestAPIRequest(t *testing.T) {
	url := fmt.Sprintf("http://localhost:%s/api/v1/activity?why=cube&database_list=(postgres|rdsadmin)&start=now-900000ms&end=now&step=5000ms&limitdim=15&limitlegend=15&legend=wait_event_name&dim=time&filterdim=&filterdimselected=&dbidentifier=%s", config.BffPort, dbIdentifier)

	var responseData struct {
		Data []struct {
			Metric struct {
				WaitEventName string `json:"wait_event_name"`
			} `json:"metric"`
			Values []struct {
				Timestamp int64 `json:"timestamp"`
				Value     int   `json:"value"`
			} `json:"values"`
		} `json:"data"`
		ServerNow int64 `json:"server_now"`
	}

	timeout := time.Now().Add(2 * time.Minute)
	interval := 5 * time.Second

	for time.Now().Before(timeout) {
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", body)

		err = json.Unmarshal(body, &responseData)
		assert.NoError(t, err)

		if len(responseData.Data) > 0 && len(responseData.Data[0].Values) > 0 {
			break
		}

		time.Sleep(interval)
	}

	assert.Greater(t, len(responseData.Data), 0, "Expected at least one data point")
	assert.Greater(t, len(responseData.Data[0].Values), 0, "Expected at least one value for the first metric")
}
