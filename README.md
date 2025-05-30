# QRender

A simple and flexible template rendering tool that supports environment variable substitution and conditional statements.

## Features

- Environment variable substitution using `$VAR` or `${VAR}` syntax
- Conditional statements with support for various operators
- Support for both system environment variables and custom environment files
- Verbose mode for debugging
- Flexible output options (stdout or file)

## Installation

```bash
go install github.com/yourusername/qrender@latest
```

## Usage

```bash
qrender -template <template_file> [-output <output_file>] [-verbose] [-env <env_file>]
```

### Arguments

- `-template`: Template file path (required)
- `-output`: Output file path (default: stdout)
- `-verbose`: Print environment variables (default: false)
- `-env`: Environment variables file path (optional)

### Environment Variables

You can use environment variables in two ways:

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
{{ endif }}

{{ if DB_PORT > 5000 }}
Using a high port number
{{ endif }}

{{ if DB_NAME startsWith "test" }}
This is a test database
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

4. Debug mode:
```bash
qrender -template example.txt -verbose
```

## License

MIT 