// Package main provides a command-line tool to generate Prometheus recording rules
// from predefined metrics configurations.
//
// The tool takes an optional output path flag and generates a YAML configuration
// file containing Prometheus recording rules. These rules are used to pre-compute
// and save frequently-used or computationally-expensive queries.
//
// Usage:
//
//	go run main.go [-output path/to/output.yml]
//
// Flags:
//
//	-output: Specifies the output file path for the recording rules YAML
//	        (default: "../prometheus/recording_rules.yml")
//
// The program will:
// 1. Extract recording rules from the server package
// 2. Generate a Prometheus-compatible YAML configuration
// 3. Create the output directory if it doesn't exist
// 4. Write the configuration to the specified file
//
// Exit codes:
//
//	0: Success
//	1: Error (directory creation or file writing failed)
package main

import (
	"flag"
	"fmt"
	"local/bff/pkg/server"
	"os"
	"path/filepath"
	"strings"
)

func writeRulesToFile(path string, yamlConfig string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %v", err)
	}

	if err := os.WriteFile(path, []byte(yamlConfig), 0644); err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	fmt.Printf("Recording rules written to %s\n", path)
	return nil
}

func main() {
	outputPaths := flag.String("output", "../prometheus/recording_rules.yml,../collector-api/recording_rules.yml",
		"Comma-separated list of output paths for the recording rules YAML")
	flag.Parse()

	// Generate rules
	rules := server.ExtractRecordingRules()
	yamlConfig := server.GeneratePrometheusConfig(rules)

	// Split paths and write to each location
	paths := strings.Split(*outputPaths, ",")
	for _, path := range paths {
		path = strings.TrimSpace(path) // Handle any spaces after commas
		if err := writeRulesToFile(path, yamlConfig); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
	}

	fmt.Printf("Generated %d rule groups\n", len(rules))
}
