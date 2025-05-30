# QRender - Simple Template Rendering Tool

QRender is a lightweight template rendering tool that supports environment variable substitution and conditional statements. It helps you quickly generate configuration files or other files that require dynamic content.

## Features

- Environment Variable Substitution
  - `${VAR}` format
  - `$VAR` format
- Conditional Statements
  - Numeric comparisons: `>`, `<`, `>=`, `<=`, `==`, `!=`
  - String comparisons: `==`, `!=`
  - String operations: `startsWith`, `endsWith`
- Nested Conditional Statements
- Multi-line Content Support
- Empty Value Detection
- Special Character Support

## Usage

```bash
qrender.exe -template <template_file> [-output <output_file>] [-verbose]
```

### Parameters

- `-template`: Template file path (required)
- `-output`: Output file path (optional, defaults to stdout)
- `-verbose`: Show environment variables (optional, defaults to false)

## Examples

### Template File (example.txt)

```
# Environment Variable Test
Current User: ${USER}
Current Directory: ${PWD}
Simple Variable: $HOME

# Conditional Tests
{{ if VERSION > "1.0.0" }}
Version is greater than 1.0.0
{{ endif }}

{{ if STATUS == "running" }}
Status is running
{{ endif }}

{{ if PATH startsWith "/usr" }}
This is a system path
{{ endif }}
```

### Running the Command

```bash
# Set environment variables
set VERSION=1.5.0
set STATUS=running
set PATH=/usr/local/bin

# Run the program
qrender.exe -template example.txt -output result.txt -verbose
```

## Conditional Statement Syntax

1. Numeric Comparisons:
   ```
   {{ if VAR > "1.0.0" }}
   {{ if VAR < "2.0.0" }}
   {{ if VAR >= "5" }}
   {{ if VAR <= "10" }}
   ```

2. String Comparisons:
   ```
   {{ if VAR == "value" }}
   {{ if VAR != "value" }}
   ```

3. String Operations:
   ```
   {{ if VAR startsWith "prefix" }}
   {{ if VAR endsWith "suffix" }}
   ```

## Notes

1. Values in conditional statements must be quoted
2. Operators must have spaces before and after
3. Nested conditional statements are supported
4. Non-existent environment variables are replaced with empty strings

## Error Handling

The program will output error messages in the following cases:
- Template file does not exist
- Output file cannot be created
- Invalid conditional statement format
- Unclosed conditional statement blocks 