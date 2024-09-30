package server

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"testing"
	"time"
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

	now := time.Now()

	for _, tt := range testCases {
		t.Run(tt.Name, func(t *testing.T) {
			var rawInput map[string]interface{}
			err := json.Unmarshal(tt.Input, &rawInput)
			if err != nil {
				t.Fatalf("Failed to unmarshal input: %v", err)
			}

			startStr, ok := rawInput["start"].(string)
			if !ok {
				t.Fatalf("Invalid type for start: %v", rawInput["start"])
			}
			endStr, ok := rawInput["end"].(string)
			if !ok {
				t.Fatalf("Invalid type for end: %v", rawInput["end"])
			}

			startTime, err := parseTimeParameter(startStr, now)
			if err != nil {
				if !tt.HasError {
					t.Errorf("unexpected error parsing start time %v", err)
				}
				return
			}

			endTime, err := parseTimeParameter(endStr, now)
			if err != nil {
				if !tt.HasError {
					t.Errorf("unexpected error parsing start time %v", err)
				}
				return
			}

			limitValue := 0
			offsetValue := 0

			if rawInput["limit"].(string) != "" {
				limitValue, err = strconv.Atoi(rawInput["limit"].(string))
				if err != nil {
					t.Errorf("unexpected error parsing limit %v", err)
					return
				}
			}

			if rawInput["offset"].(string) != "" {
				offsetValue, err = strconv.Atoi(rawInput["offset"].(string))
				if err != nil {
					t.Errorf("unexpected error parsing offset %v", err)
					return
				}
			}

			input := PromQLInput{
				DatabaseList:      rawInput["database_list"].(string),
				Start:             startTime,
				End:               endTime,
				Legend:            rawInput["legend"].(string),
				Dim:               rawInput["dim"].(string),
				FilterDim:         rawInput["filterdim"].(string),
				FilterDimSelected: rawInput["filterdimselected"].(string),
				Limit:             limitValue,
				Offset:            offsetValue,
			}

			query, err := GenerateActivityCubePromQLQuery(input)
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
