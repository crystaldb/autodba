package main

import (
	"os"
	"testing"
	"log"

	"fmt"
	"net/http"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	if err := SetupTestContainer(); err != nil {
		fmt.Printf("Failed to set up container: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := TearDownTestContainer(); err != nil {
			fmt.Printf("Failed to tear down container: %v\n", err)
		}
	}()

	m.Run()
}

func TestAPIRequest(t *testing.T) {
	resp, err := http.Get("http://localhost:" + BFF_PORT + "/api/v1/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	log.Println("Response: ", resp)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	// body, err := ioutil.ReadAll(resp.Body)
	// assert.NoError(t, err)
	// assert.Equal(t, expectedBody, string(body))
}
