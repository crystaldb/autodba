package server_test

import (
	"encoding/json"
	"fmt"
	"io"
	"local/bff/pkg/server" // Replace with the actual import path to your server package
	"os"
	"testing"
)

type TestCase struct {
	Name     string          `json:"name"`
	Input    json.RawMessage `json:"input"`
	Expected string          `json:"expected"`
	HasError bool            `json:"has_error"`
}

func loadTestCases(filePath string) ([]TestCase, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open test cases file: %w", err)
	}
	defer file.Close()

	var testCases []TestCase
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read test cases file: %w", err)
	}
	err = json.Unmarshal(data, &testCases)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal test cases: %w", err)
	}
	return testCases, nil
}

func saveTestCases(filePath string, testCases []TestCase) error {
	data, err := json.MarshalIndent(testCases, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal test cases: %w", err)
	}
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write test cases file: %w", err)
	}
	return nil
}

func TestGenerateActivityCubePromQLQuery(t *testing.T) {
	filePath := "promql_codegen_test_cases.json" // Path to your JSON test cases file
	testCases, err := loadTestCases(filePath)
	if err != nil {
		t.Fatalf("Error loading test cases: %v", err)
	}

	var updatedTestCases []TestCase

	// Track whether any changes were made
	anyMismatch := false

	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			var input server.PromQLInput
			err := json.Unmarshal(tt.Input, &input)
			if err != nil {
				t.Fatalf("Failed to unmarshal input: %v", err)
			}

			query, err := server.GenerateActivityCubePromQLQuery(input)
			if (err != nil) != tt.HasError {
				t.Errorf("expected error: %v, got: %v", tt.HasError, err)
				anyMismatch = true
			}
			if query != tt.Expected {
				t.Errorf("expected query: %s, got: %s", tt.Expected, query)
				anyMismatch = true
				tt.Expected = query
			}
		})
		updatedTestCases = append(updatedTestCases, tt)
	}

	// If there was any mismatch, save the updated test cases
	if anyMismatch {
		err = saveTestCases(filePath, updatedTestCases)
		if err != nil {
			t.Fatalf("Error saving test cases: %v", err)
		}
	}
}
