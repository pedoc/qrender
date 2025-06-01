package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	Version = "1.0.7"
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

func evaluateCondition(key, operator, value string, env map[string]string) bool {
	envValue := env[key]

	// Debug output
	//fmt.Fprintf(os.Stderr, "Evaluating condition: %s %s %s (env value: %s)\n", key, operator, value, envValue)

	// Handle string special operators
	switch operator {
	case "startsWith":
		return strings.HasPrefix(envValue, value)
	case "endsWith":
		return strings.HasSuffix(envValue, value)
	}

	// Try to convert values to numbers for comparison
	envNum, envErr := strconv.ParseFloat(envValue, 64)
	valueNum, valueErr := strconv.ParseFloat(value, 64)

	// If both values can be converted to numbers, perform numeric comparison
	if envErr == nil && valueErr == nil {
		switch operator {
		case "==":
			return envNum == valueNum
		case "!=":
			return envNum != valueNum
		case ">":
			return envNum > valueNum
		case "<":
			return envNum < valueNum
		case ">=":
			return envNum >= valueNum
		case "<=":
			return envNum <= valueNum
		}
	}

	// Otherwise perform string comparison
	switch operator {
	case "==":
		return envValue == value
	case "!=":
		return envValue != value
	case ">":
		return envValue > value
	case "<":
		return envValue < value
	case ">=":
		return envValue >= value
	case "<=":
		return envValue <= value
	}

	return false
}

type ifBlock struct {
	skipUntilEndif bool
	lineNumber     int
	hasElse        bool
	hasMatched     bool
}

func loadEnvFile(filePath string) (map[string]string, error) {
	env := make(map[string]string)

	// Read environment variables file
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read env file: %v", err)
	}

	// Process line by line
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes
		value = strings.Trim(value, "\"'")

		env[key] = value
	}

	return env, nil
}

func main() {
	// Parse command line arguments
	templateFile := flag.String("template", "example.txt", "template file path (required)")
	outputFile := flag.String("output", "", "output file path (default: stdout)")
	verbose := flag.Bool("verbose", false, "print environment variables (default: false)")
	envFile := flag.String("env", "", "environment variables file (optional)")
	vars := flag.String("vars", "", "comma-separated list of environment variables to substitute (optional)")
	showVersion := flag.Bool("version", false, "show version information")
	flag.Parse()

	// Show version and exit if requested
	if *showVersion {
		fmt.Printf("QRender version %s\n", Version)
		os.Exit(0)
	}

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

	// Load environment variables
	var env map[string]string
	if *envFile != "" {
		// Load from environment file
		env, err = loadEnvFile(*envFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Load from system environment
		env = make(map[string]string)
		for _, e := range os.Environ() {
			pair := strings.SplitN(e, "=", 2)
			if len(pair) == 2 {
				key := pair[0]
				value := pair[1]
				env[key] = value
			}
		}
	}

	// Filter environment variables if vars flag is set
	if *vars != "" {
		filteredEnv := make(map[string]string)
		for _, v := range strings.Split(*vars, ",") {
			v = strings.TrimSpace(v)
			if value, exists := env[v]; exists {
				filteredEnv[v] = value
			}
		}
		env = filteredEnv
	}

	// Check and clean environment variable values, removing quotes
	for key, value := range env {
		// Remove all quotes, including those that might appear in the middle
		value = strings.ReplaceAll(value, "\"", "")
		value = strings.ReplaceAll(value, "'", "")
		env[key] = value
	}

	// Print environment variables if verbose mode is enabled
	if *verbose {
		fmt.Println("Environment variables:")
		for key, value := range env {
			fmt.Printf("  %s=%s\n", key, value)
		}
		fmt.Println()
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
	ifStack := make([]ifBlock, 0)
	lineNumber := 0

	// Compile regex for if statements
	ifRegex := regexp.MustCompile(`(?:#@)?{{\s*if\s+(.+?)\s*}}`)
	elseIfRegex := regexp.MustCompile(`(?:#@)?{{\s*else\s+if\s+(.+?)\s*}}`)
	elseRegex := regexp.MustCompile(`(?:#@)?{{\s*else\s*}}`)
	endifRegex := regexp.MustCompile(`(?:#@)?{{\s*endif\s*}}`)

	// Supported operators
	operators := []string{
		"==", "!=", ">=", "<=", ">", "<",
		"startsWith", "endsWith",
	}

	for _, line := range lines {
		lineNumber++
		trimmedLine := strings.TrimSpace(line)

		// Process if statement
		if ifMatch := ifRegex.FindStringSubmatch(trimmedLine); ifMatch != nil {
			// Extract condition
			condition := strings.TrimSpace(ifMatch[1])

			// Find operator
			var foundOperator string
			var parts []string

			for _, op := range operators {
				if strings.Contains(condition, " "+op+" ") {
					foundOperator = op
					parts = strings.Split(condition, " "+op+" ")
					break
				}
			}

			if foundOperator == "" || len(parts) != 2 {
				fmt.Fprintf(os.Stderr, "Warning: Line %d: Invalid condition statement format. Supported operators: ==, !=, >, <, >=, <=, startsWith, endsWith\n", lineNumber)
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

			// Create new if block
			conditionResult := evaluateCondition(key, foundOperator, value, env)
			newBlock := ifBlock{
				skipUntilEndif: !conditionResult,
				lineNumber:     lineNumber,
				hasElse:        false,
				hasMatched:     conditionResult,
			}

			// If parent if block is skipped, this block should also be skipped
			if len(ifStack) > 0 && ifStack[len(ifStack)-1].skipUntilEndif {
				newBlock.skipUntilEndif = true
				newBlock.hasMatched = false
			}

			ifStack = append(ifStack, newBlock)
			continue
		}

		// Process else if statement
		if elseIfMatch := elseIfRegex.FindStringSubmatch(trimmedLine); elseIfMatch != nil {
			if len(ifStack) == 0 {
				fmt.Fprintf(os.Stderr, "Warning: Line %d: Unmatched else if statement\n", lineNumber)
				continue
			}

			// Extract condition
			condition := strings.TrimSpace(elseIfMatch[1])

			// Find operator
			var foundOperator string
			var parts []string

			for _, op := range operators {
				if strings.Contains(condition, " "+op+" ") {
					foundOperator = op
					parts = strings.Split(condition, " "+op+" ")
					break
				}
			}

			if foundOperator == "" || len(parts) != 2 {
				fmt.Fprintf(os.Stderr, "Warning: Line %d: Invalid condition statement format. Supported operators: ==, !=, >, <, >=, <=, startsWith, endsWith\n", lineNumber)
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

			// Get the current if block
			currentBlock := &ifStack[len(ifStack)-1]
			currentBlock.hasElse = true

			// Only evaluate the condition if no previous condition has matched
			if !currentBlock.hasMatched {
				conditionResult := evaluateCondition(key, foundOperator, value, env)
				currentBlock.skipUntilEndif = !conditionResult
				currentBlock.hasMatched = conditionResult
			} else {
				currentBlock.skipUntilEndif = true
			}
			continue
		}

		// Process else statement
		if elseRegex.MatchString(trimmedLine) {
			if len(ifStack) == 0 {
				fmt.Fprintf(os.Stderr, "Warning: Line %d: Unmatched else statement\n", lineNumber)
				continue
			}

			// Get the current if block
			currentBlock := &ifStack[len(ifStack)-1]
			currentBlock.hasElse = true

			// Only execute else if no previous condition has matched
			currentBlock.skipUntilEndif = currentBlock.hasMatched
			continue
		}

		// Process endif
		if endifRegex.MatchString(trimmedLine) {
			if len(ifStack) == 0 {
				fmt.Fprintf(os.Stderr, "Warning: Line %d: Unmatched endif statement\n", lineNumber)
			} else {
				// Pop the last if block
				ifStack = ifStack[:len(ifStack)-1]
			}
			continue
		}

		// Check if current line should be skipped
		shouldSkip := false
		for _, block := range ifStack {
			if block.skipUntilEndif {
				shouldSkip = true
				break
			}
		}

		if shouldSkip {
			continue
		}

		// Output normal line
		if _, err := writer.WriteString(line + "\n"); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write output: %v\n", err)
			os.Exit(1)
		}
	}

	// Check for unclosed if blocks
	if len(ifStack) > 0 {
		for _, block := range ifStack {
			fmt.Fprintf(os.Stderr, "Warning: Unclosed if block starting at line %d\n", block.lineNumber)
		}
	}

	// Ensure all content is written
	if err := writer.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to flush output buffer: %v\n", err)
		os.Exit(1)
	}
}
