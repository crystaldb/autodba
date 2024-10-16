package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib" // Importing pgx driver
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"
)

var dbIdentifier string
var port string
var dbConfigStr = flag.String("dbconfig", "", "JSON string of database configuration map")

func TestAPISuite(t *testing.T) {
	var dbInfoMap DbInfoMap
	if err := json.Unmarshal([]byte(*dbConfigStr), &dbInfoMap); err != nil {
		log.Fatalf("Failed to unmarshal database configuration: %v\n", err)
	}

	fmt.Printf("config: %+v\n", defaultConfig)
	fmt.Printf("dbMap: %+v\n", dbInfoMap)

	for version, info := range dbInfoMap {
		dbIdentifier = info.AwsRdsInstance
		port = defaultConfig.BffPort

		dbVersion, err := getDatabaseVersion(info.DbConnString)
		if err != nil {
			t.Fatalf("Failed to get database version for %s: %v\n", info.Description, err)
			return
		}
		log.Println("Db version : ", dbVersion)

		versionPrefix := fmt.Sprintf("PostgreSQL %s", version)

		if !strings.HasPrefix(dbVersion, versionPrefix) {
			t.Fatalf("Database version %s does not match expected version %s for %s\n", dbVersion, version, info.Description)
		}

		if err := SetupTestContainer(&defaultConfig, info); err != nil {
			t.Fatalf("Failed to set up container for %s: %v\n", info.Description, err)
		}

		t.Run(info.Description, TestAPIRequest)

		if err := TearDownTestContainer(); err != nil {
			log.Printf("Failed to tear down container for %s: %v\n", info.Description, err)
		}
	}
}

func TestAPIRequest(t *testing.T) {
	url := fmt.Sprintf("http://localhost:%s/api/v1/activity?why=cube&database_list=(postgres|rdsadmin)&start=now-900000ms&end=now&step=5000ms&limitdim=15&limitlegend=15&legend=wait_event_name&dim=time&filterdim=&filterdimselected=&dbidentifier=%s", port, dbIdentifier)

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

		if resp.StatusCode != http.StatusOK {
			t.Logf("Received non-OK response: %s", resp.Status)
			time.Sleep(interval)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		fmt.Printf("%s\n", body)

		if err := json.Unmarshal(body, &responseData); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}

		break
	}

	assert.Greater(t, len(responseData.Data), 0, "Expected at least one data point")
	assert.Greater(t, len(responseData.Data[0].Values), 0, "Expected at least one value for the first metric")
}

func getDatabaseVersion(connectionString string) (string, error) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return "", err
	}
	defer db.Close()

	var version string
	err = db.QueryRow("SELECT version();").Scan(&version)
	if err != nil {
		return "", err
	}

	return version, nil
}
