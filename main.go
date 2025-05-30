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

	// 处理字符串特殊操作符
	switch operator {
	case "startsWith":
		return strings.HasPrefix(envValue, value)
	case "endsWith":
		return strings.HasSuffix(envValue, value)
	}

	// 尝试将值转换为数字进行比较
	envNum, envErr := strconv.ParseFloat(envValue, 64)
	valueNum, valueErr := strconv.ParseFloat(value, 64)

	// 如果两个值都可以转换为数字，则进行数值比较
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

	// 否则进行字符串比较
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
}

func main() {
	// Parse command line arguments
	templateFile := flag.String("template", "example.txt", "template file path (required)")
	outputFile := flag.String("output", "", "output file path (default: stdout)")
	verbose := flag.Bool("verbose", false, "print environment variables (default: false)")
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

	// Print environment variables if verbose mode is enabled
	if *verbose {
		fmt.Fprintln(os.Stderr, "Environment variables:")
		for key, value := range env {
			fmt.Fprintf(os.Stderr, "  %s=%s\n", key, value)
		}
		fmt.Fprintln(os.Stderr)
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
	ifRegex := regexp.MustCompile(`{{\s*if\s+(.+?)\s*}}`)
	endifRegex := regexp.MustCompile(`{{\s*endif\s*}}`)

	// 支持的操作符
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

			// 查找操作符
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

			// 创建新的if块
			newBlock := ifBlock{
				skipUntilEndif: !evaluateCondition(key, foundOperator, value, env),
				lineNumber:     lineNumber,
			}

			// 如果父级if块被跳过，这个块也应该被跳过
			if len(ifStack) > 0 && ifStack[len(ifStack)-1].skipUntilEndif {
				newBlock.skipUntilEndif = true
			}

			ifStack = append(ifStack, newBlock)
			continue
		}

		// Process endif
		if endifRegex.MatchString(trimmedLine) {
			if len(ifStack) == 0 {
				fmt.Fprintf(os.Stderr, "Warning: Line %d: Unmatched endif statement\n", lineNumber)
			} else {
				// 弹出最后一个if块
				ifStack = ifStack[:len(ifStack)-1]
			}
			continue
		}

		// 检查是否应该跳过当前行
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
