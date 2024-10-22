package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
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
var imageName = flag.String("imageName", "", "Name of docker image to test against")
var currentDbInfo DbInfo

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
		currentDbInfo = info

		dbVersion, err := getDatabaseVersion(constructDBConnString(info))
		if err != nil {
			t.Fatalf("Failed to get database version for %s: %v\n", info.Description, err)
			return
		}
		log.Println("Db version : ", dbVersion)

		versionPrefix := fmt.Sprintf("PostgreSQL %s", version)

		if !strings.HasPrefix(dbVersion, versionPrefix) {
			t.Fatalf("Database version %s does not match expected version %s for %s\n", dbVersion, version, info.Description)
		}
		t.Run(info.Description, func(t *testing.T) {
			if err := SetupTestContainer(&defaultConfig, info, *imageName); err != nil {
				t.Fatalf("Failed to set up container for %s: %v\n", info.Description, err)
			}
			defer func() {
				if err := TearDownTestContainer(); err != nil {
					log.Printf("Failed to tear down container for %s: %v\n", info.Description, err)
				}
			}()

			TestAPIRequest(t)
		})
	}
}

func TestAPIRequest(t *testing.T) {
	url := fmt.Sprintf("http://localhost:%s/api/v1/activity?why=cube&database_list=(postgres|rdsadmin)&start=now-900000ms&end=now&step=5000ms&limitdim=15&limitlegend=15&legend=wait_event_name&dim=time&filterdim=&filterdimselected=&dbidentifier=%s", port, fmt.Sprintf("%s/%s/%s", currentDbInfo.SystemType, currentDbInfo.AwsRdsInstance, currentDbInfo.SystemScope))

	fmt.Println("url: ", url)
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

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}
		fmt.Printf("%s\n", body)

		if resp.StatusCode != http.StatusOK {
			t.Logf("Received non-OK response: %s", resp.Status)
			time.Sleep(interval)
			continue
		}

		if err := json.Unmarshal(body, &responseData); err != nil {
			t.Fatalf("Failed to unmarshal response body: %v", err)
		}
		if len(responseData.Data) == 0 {
			t.Log("Received empty data, retrying...")
			time.Sleep(interval)
			continue
		}

		break
	}

	if len(responseData.Data) == 0 {
		t.Fatalf("No valid data received after retries")
	}

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
