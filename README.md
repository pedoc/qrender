# QRender

A simple and flexible template rendering tool that supports environment variable substitution and conditional statements.

## Features

- Environment variable substitution using `$VAR` or `${VAR}` syntax
- Conditional statements with support for various operators
- Support for both system environment variables and custom environment files
- Ability to specify which environment variables to substitute
- Verbose mode for debugging
- Flexible output options (stdout or file)
- Support for nested conditional statements
- Support for Unicode and special characters
- Support for empty and whitespace value checks
- Support for `#@` prefix in conditional statements (compatible with .conf files)

## Installation

```bash
go install github.com/yourusername/qrender@latest
```

## Usage

```bash
qrender -template <template_file> [-output <output_file>] [-verbose] [-env <env_file>] [-vars <var1,var2,...>]
```

### Arguments

- `-template`: Template file path (required)
- `-output`: Output file path (default: stdout)
- `-verbose`: Print environment variables (default: false)
- `-env`: Environment variables file path (optional)
- `-vars`: Comma-separated list of environment variables to substitute (optional)
- `-version`: Show version information

### Environment Variables

You can use environment variables in three ways:

1. System environment variables:
```bash
export DB_HOST=localhost
export DB_PORT=5432
qrender -template example.txt
```

2. Custom environment file:
```bash
# vars.env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=mydb
DB_USER=admin
DB_PASSWORD="secret123"

# Run with environment file
qrender -template example.txt -env vars.env
```

3. Specify variables to substitute:
```bash
# Only substitute specific variables
qrender -template example.txt -vars "DB_HOST,DB_PORT,DB_NAME"
```

The environment file supports:
- One variable per line in `KEY=VALUE` format
- Comments (lines starting with #)
- Quoted values (automatically unquoted)
- Empty lines are ignored

### Template Syntax

1. Environment Variable Substitution:
```
Database host: $DB_HOST
Database port: ${DB_PORT}
```

2. Conditional Statements:
```
{{ if DB_HOST == "localhost" }}
This is a local development environment
{{ else if DB_HOST == "staging" }}
This is a staging environment
{{ else }}
This is a production environment
{{ endif }}

# Support #@ prefix(Use # comment syntax to avoid damaging LSP (plugin) of .conf and similar files)
#@{{ if DB_HOST == "localhost" }}
This is a local development environment
#@{{ else if DB_HOST == "staging" }}
This is a staging environment
#@{{ else }}
This is a production environment
#@{{ endif }}

{{ if DB_PORT > 5000 }}
Using a high port number
{{ else if DB_PORT > 1000 }}
Using a medium port number
{{ else }}
Using a low port number
{{ endif }}

{{ if DB_NAME startsWith "test" }}
This is a test database
{{ else if DB_NAME startsWith "dev" }}
This is a development database
{{ else }}
This is a production database
{{ endif }}
```

3. Nested Conditional Statements:
```
{{ if ENV == "dev" }}
  {{ if DEBUG == "true" }}
    Debug mode is enabled in development environment
  {{ endif }}
{{ endif }}
```

4. Empty and Whitespace Value Checks:
```
{{ if EMPTY_VAR == "" }}
Variable is empty
{{ endif }}

{{ if WHITESPACE == "" }}
Variable contains only whitespace
{{ endif }}
```

5. Special Characters and Unicode:
```
{{ if SPECIAL_CHARS == "!@#$%^&*()" }}
Special characters match
{{ endif }}

{{ if UNICODE == "测试" }}
Unicode characters match
{{ endif }}
```

### Supported Operators

- Comparison: `==`, `!=`, `>`, `<`, `>=`, `<=`
- String: `startsWith`, `endsWith`

## Examples

1. Basic usage:
```bash
qrender -template example.txt
```

2. Save to file:
```bash
qrender -template example.txt -output config.txt
```

3. Use custom environment file:
```bash
qrender -template example.txt -env vars.env
```

4. Specify variables to substitute:
```bash
qrender -template example.txt -vars "DB_HOST,DB_PORT,DB_NAME"
```

5. Debug mode:
```bash
qrender -template example.txt -verbose
```

6. Nested conditions:
```bash
# template.txt
{{ if ENV == "dev" }}
  {{ if DEBUG == "true" }}
    Debug mode in dev
  {{ endif }}
{{ endif }}

# Run
qrender -template template.txt
```

7. Special characters and Unicode:
```bash
# template.txt
{{ if SPECIAL_CHARS == "!@#$%^&*()" }}
Special chars OK
{{ endif }}
{{ if UNICODE == "测试" }}
Unicode OK
{{ endif }}

# Run
qrender -template template.txt
```

8. Empty value checks:
```bash
# template.txt
{{ if EMPTY_VAR == "" }}
Empty var detected
{{ endif }}

# Run
qrender -template template.txt
```

## License

MIT 