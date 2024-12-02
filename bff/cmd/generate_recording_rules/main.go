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
)

func main() {
	outputPath := flag.String("output", "../prometheus/recording_rules.yml", "Output path for the recording rules YAML")
	flag.Parse()

	// Generate rules
	rules := server.ExtractRecordingRules()
	yamlConfig := server.GeneratePrometheusConfig(rules)

	// Ensure directory exists
	dir := filepath.Dir(*outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
		os.Exit(1)
	}

	// Write to file
	if err := os.WriteFile(*outputPath, []byte(yamlConfig), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Recording rules written to %s\n", *outputPath)
	fmt.Printf("Generated %d rule groups\n", len(rules))
}
