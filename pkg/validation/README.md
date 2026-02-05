# Validation Package

A robust, self-contained validation wrapper for Go applications with sensitive data sanitization and structured logging.

## Features

- ✅ **Self-contained**: Zero dependencies on central config module
- ✅ **Environment-based configuration**: 12-factor app compliant
- ✅ **go-playground/validator**: Built on industry standard validator
- ✅ **Sensitive data redaction**: Automatic sanitization of passwords, tokens, etc
- ✅ **Structured logging**: slog integration
- ✅ **Custom validators**: Register your own validation functions
- ✅ **Brazilian validators**: CPF, CNPJ, phone, CEP validation
- ✅ **Comprehensive error handling**: Using fault package

## Installation

```bash
go get github.com/marcelofabianov/validation
```

## Quick Start

### Using Environment Variables

Create a `.env` file (see `.env.example`):

```env
VALIDATION_ENABLE_LOGGING=true
VALIDATION_SANITIZE_SENSITIVE_DATA=true
VALIDATION_LOG_SUCCESSFUL_VALIDATIONS=false
```

Use in your code:

```go
package main

import (
    "context"
    "log/slog"
    "github.com/marcelofabianov/validation"
)

type User struct {
    Name     string `json:"name" validate:"required,min=3"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}

func main() {
    cfg, _ := validation.LoadConfig()
    validator := validation.New(cfg, slog.Default())
    
    user := User{
        Name:     "Jo",
        Email:    "invalid",
        Password: "123",
    }
    
    ctx := context.Background()
    if err := validator.Struct(ctx, user); err != nil {
        // Handle validation errors
        // Password will be automatically redacted in logs
    }
}
```

## Configuration

### Environment Variables

All variables use the `VALIDATION_` prefix:

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `VALIDATION_ENABLE_LOGGING` | bool | true | Enable logging for validation |
| `VALIDATION_SANITIZE_SENSITIVE_DATA` | bool | true | Redact sensitive fields in logs |
| `VALIDATION_ADDITIONAL_SENSITIVE_FIELDS` | []string | [] | Additional fields to redact |
| `VALIDATION_LOG_SUCCESSFUL_VALIDATIONS` | bool | false | Log successful validations |

### Default Sensitive Fields

The following fields are automatically redacted in logs:
- password, senha
- token, secret
- apikey, api_key
- credit_card, card_number
- cvv, pin
- private_key

## Operations

### Struct Validation

```go
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=3,max=100"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"required,min=18"`
    Password string `json:"password" validate:"required,min=8"`
}

req := CreateUserRequest{
    Name:     "John Doe",
    Email:    "john@example.com",
    Age:      25,
    Password: "secret123",
}

err := validator.Struct(ctx, req)
```

### Field Validation

```go
email := "invalid-email"
err := validator.Field(ctx, email, "required,email")
```

### Custom Validators

```go
import "github.com/go-playground/validator/v10"

isEven := func(fl validator.FieldLevel) bool {
    value := fl.Field().Int()
    return value%2 == 0
}

validator.RegisterCustom("even", isEven)

type Data struct {
    Number int `json:"number" validate:"even"`
}
```

### Brazilian Validators

```go
import "github.com/marcelofabianov/validation"

// Validate CPF
cpf := "12345678901"
err := validation.ValidateCPF(cpf)

// Validate CNPJ
cnpj := "12345678000190"
err := validation.ValidateCNPJ(cnpj)

// Validate Brazilian phone
phone := "11987654321"
err := validation.ValidateBrazilianPhone(phone)

// Validate CEP
cep := "01310100"
err := validation.ValidateCEP(cep)
```

## Validation Tags

Common tags (from go-playground/validator):

- `required` - Field must be present
- `email` - Valid email address
- `min=N` - Minimum length/value
- `max=N` - Maximum length/value
- `len=N` - Exact length
- `eq=N` - Equal to value
- `ne=N` - Not equal to value
- `gt=N` - Greater than
- `gte=N` - Greater than or equal
- `lt=N` - Less than
- `lte=N` - Less than or equal
- `oneof=red green blue` - One of the values
- `url` - Valid URL
- `uuid` - Valid UUID
- `alpha` - Alphabetic only
- `alphanum` - Alphanumeric only
- `numeric` - Numeric only

## Sensitive Data Sanitization

When `VALIDATION_SANITIZE_SENSITIVE_DATA=true`, sensitive fields are automatically redacted:

```go
type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

// In logs, password will appear as: "***REDACTED***"
```

Add custom sensitive fields:

```env
VALIDATION_ADDITIONAL_SENSITIVE_FIELDS=api_token,access_key,secret_key
```

## Architecture

This package follows the **self-contained pattern** for microservices monorepos:

- ✅ No imports of central config module
- ✅ Independent configuration via environment variables
- ✅ Can be extracted to separate repository
- ✅ Zero coupling with other packages

## Testing

```bash
go test ./...
```

## License

MIT
