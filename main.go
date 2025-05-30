package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func checkFileExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", path)
		}
		return fmt.Errorf("error checking file: %v", err)
	}
	if info.IsDir() {
		return fmt.Errorf("path is a directory, not a file: %s", path)
	}
	return nil
}

func checkOutputPath(path string) error {
	if path == "" {
		return nil
	}

	// Check if output directory exists
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}
	}

	// Check if file is writable
	if _, err := os.Stat(path); err == nil {
		// File exists, check if writable
		file, err := os.OpenFile(path, os.O_WRONLY, 0)
		if err != nil {
			return fmt.Errorf("file exists but is not writable: %v", err)
		}
		file.Close()
	}
	return nil
}

func main() {
	// Parse command line arguments
	templateFile := flag.String("template", "example.txt", "template file path (required)")
	outputFile := flag.String("output", "", "output file path (default: stdout)")
	flag.Parse()

	// Check template file
	if err := checkFileExists(*templateFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Check output file
	if err := checkOutputPath(*outputFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Read template file
	templateContent, err := ioutil.ReadFile(*templateFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read template file: %v\n", err)
		os.Exit(1)
	}

	// Read all environment variables
	env := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			env[pair[0]] = pair[1]
		}
	}

	// Prepare output
	var writer *bufio.Writer
	if *outputFile != "" {
		file, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create output file: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if err := file.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to close output file: %v\n", err)
			}
		}()
		writer = bufio.NewWriter(file)
	} else {
		writer = bufio.NewWriter(os.Stdout)
	}
	defer writer.Flush()

	// Simple template processing
	content := string(templateContent)

	// Replace environment variables
	// First replace ${VAR} format
	for key, value := range env {
		content = strings.ReplaceAll(content, "${"+key+"}", value)
	}

	// Then replace $VAR format
	// Create a regex to match $VAR but not ${VAR}
	re := regexp.MustCompile(`\$([a-zA-Z_][a-zA-Z0-9_]*)`)
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		// Remove the $ prefix
		varName := match[1:]
		if value, exists := env[varName]; exists {
			return value
		}
		return match
	})

	// Process if conditions
	lines := strings.Split(content, "\n")
	inIfBlock := false
	skipUntilEndif := false
	lineNumber := 0

	for _, line := range lines {
		lineNumber++
		trimmedLine := strings.TrimSpace(line)

		// Process if statement
		if strings.HasPrefix(trimmedLine, "{{if") {
			if inIfBlock {
				fmt.Fprintf(os.Stderr, "Warning: Line %d: Nested if statements may not work as expected\n", lineNumber)
			}
			inIfBlock = true

			// Extract condition
			condition := strings.TrimSpace(strings.TrimPrefix(trimmedLine, "{{if"))
			condition = strings.TrimSuffix(condition, "}}")
			condition = strings.TrimSpace(condition)

			// Check condition
			parts := strings.Split(condition, "==")
			if len(parts) != 2 {
				fmt.Fprintf(os.Stderr, "Warning: Line %d: Invalid condition statement format\n", lineNumber)
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

			if env[key] != value {
				skipUntilEndif = true
			}
			continue
		}

		// Process endif
		if trimmedLine == "{{endif}}" {
			if !inIfBlock {
				fmt.Fprintf(os.Stderr, "Warning: Line %d: Unmatched endif statement\n", lineNumber)
			}
			inIfBlock = false
			skipUntilEndif = false
			continue
		}

		// Skip lines in if block if condition is false
		if inIfBlock && skipUntilEndif {
			continue
		}

		// Output normal line
		if _, err := writer.WriteString(line + "\n"); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write output: %v\n", err)
			os.Exit(1)
		}
	}

	// Check for unclosed if block
	if inIfBlock {
		fmt.Fprintf(os.Stderr, "Warning: Unclosed if block at end of file\n")
	}

	// Ensure all content is written
	if err := writer.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to flush output buffer: %v\n", err)
		os.Exit(1)
	}
}
