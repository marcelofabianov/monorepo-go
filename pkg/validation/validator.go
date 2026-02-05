package validation

import (
"context"
"fmt"
"log/slog"
"reflect"
"strings"
"sync"

"github.com/go-playground/validator/v10"
"github.com/marcelofabianov/fault"
)

type Validator interface {
Struct(ctx context.Context, s any) error
Field(ctx context.Context, field any, tag string) error
RegisterCustom(tag string, fn validator.Func) error
}

type validatorImpl struct {
validate         *validator.Validate
logger           *slog.Logger
config           *Config
mu               sync.RWMutex
sensitiveFields  map[string]bool
customValidators map[string]validator.Func
}

var (
defaultSensitiveFields = []string{
"password", "senha", "token", "secret", "apikey", "api_key",
"credit_card", "card_number", "cvv", "pin", "private_key",
}

ErrValidationFailed = fault.New(
"validation failed",
fault.WithCode(fault.Invalid),
)

ErrInvalidInput = fault.New(
"invalid input for validation",
fault.WithCode(fault.Invalid),
)
)

func New(cfg *Config, logger *slog.Logger) Validator {
if cfg == nil {
cfg = DefaultConfig()
}

if logger == nil {
logger = slog.Default()
}

v := validator.New()

v.RegisterTagNameFunc(func(fld reflect.StructField) string {
name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
if name == "-" {
return ""
}
if name == "" {
return fld.Name
}
return name
})

sensitiveMap := make(map[string]bool)
for _, field := range defaultSensitiveFields {
sensitiveMap[strings.ToLower(field)] = true
}
for _, field := range cfg.AdditionalSensitiveFields {
sensitiveMap[strings.ToLower(field)] = true
}

return &validatorImpl{
validate:         v,
logger:           logger,
config:           cfg,
sensitiveFields:  sensitiveMap,
customValidators: make(map[string]validator.Func),
}
}

func (vi *validatorImpl) Struct(ctx context.Context, s any) error {
if s == nil {
return fault.Wrap(ErrInvalidInput, "struct cannot be nil")
}

err := vi.validate.StructCtx(ctx, s)
if err == nil {
return nil
}

if valErrs, ok := err.(validator.ValidationErrors); ok {
sanitized := vi.sanitizeStruct(s)
faultErr := vi.buildValidationError(valErrs)

if vi.config.EnableLogging {
vi.logger.ErrorContext(ctx, "Struct validation failed",
"struct_type", fmt.Sprintf("%T", s),
"struct_data", sanitized,
"errors", len(valErrs),
)
}

return faultErr
}

faultErr := fault.Wrap(err, "unexpected validation error",
fault.WithCode(fault.Internal),
)

if vi.config.EnableLogging {
vi.logger.ErrorContext(ctx, "Unexpected validation error",
"struct_type", fmt.Sprintf("%T", s),
"error", err.Error(),
)
}

return faultErr
}

func (vi *validatorImpl) Field(ctx context.Context, field any, tag string) error {
if field == nil {
return fault.Wrap(ErrInvalidInput, "field cannot be nil")
}

if tag == "" {
return fault.Wrap(ErrInvalidInput, "validation tag cannot be empty")
}

err := vi.validate.VarCtx(ctx, field, tag)
if err == nil {
return nil
}

if valErrs, ok := err.(validator.ValidationErrors); ok {
sanitizedValue := vi.sanitizeValue(field, "field")

faultErr := fault.New(
fmt.Sprintf("field validation failed for tag '%s'", tag),
fault.WithCode(fault.Invalid),
fault.WithContext("tag", tag),
fault.WithContext("field_value", sanitizedValue),
)

if vi.config.EnableLogging {
vi.logger.ErrorContext(ctx, "Field validation failed",
"field_value", sanitizedValue,
"tag", tag,
"errors", len(valErrs),
)
}

return faultErr
}

faultErr := fault.Wrap(err, "unexpected field validation error",
fault.WithCode(fault.Internal),
)

if vi.config.EnableLogging {
vi.logger.ErrorContext(ctx, "Unexpected field validation error",
"tag", tag,
"error", err.Error(),
)
}

return faultErr
}

func (vi *validatorImpl) RegisterCustom(tag string, fn validator.Func) error {
vi.mu.Lock()
defer vi.mu.Unlock()

if tag == "" {
return fault.Wrap(ErrInvalidInput, "custom validator tag cannot be empty")
}

if fn == nil {
return fault.Wrap(ErrInvalidInput, "custom validator function cannot be nil")
}

if err := vi.validate.RegisterValidation(tag, fn); err != nil {
return fault.Wrap(err, "failed to register custom validator",
fault.WithContext("tag", tag),
)
}

vi.customValidators[tag] = fn
return nil
}

func (vi *validatorImpl) buildValidationError(valErrs validator.ValidationErrors) error {
var messages []string
contexts := make(map[string]interface{})

for i, fieldErr := range valErrs {
msg := fmt.Sprintf("field '%s' failed validation '%s'",
fieldErr.Field(),
fieldErr.Tag(),
)

if fieldErr.Param() != "" {
msg += fmt.Sprintf(" (param: %s)", fieldErr.Param())
}

messages = append(messages, msg)
contexts[fmt.Sprintf("error_%d", i)] = msg
}

return fault.Wrap(
ErrValidationFailed,
strings.Join(messages, "; "),
fault.WithContext("validation_errors", contexts),
fault.WithContext("error_count", len(valErrs)),
fault.WithCode(fault.Invalid),
)
}

func (vi *validatorImpl) sanitizeStruct(s any) map[string]interface{} {
result := make(map[string]interface{})

val := reflect.ValueOf(s)
if val.Kind() == reflect.Ptr {
val = val.Elem()
}

if val.Kind() != reflect.Struct {
return result
}

typ := val.Type()
for i := 0; i < val.NumField(); i++ {
field := typ.Field(i)
fieldValue := val.Field(i)

if !field.IsExported() {
continue
}

jsonTag := field.Tag.Get("json")
if jsonTag == "-" {
continue
}

fieldName := field.Name
if jsonTag != "" {
parts := strings.Split(jsonTag, ",")
if parts[0] != "" {
fieldName = parts[0]
}
}

result[fieldName] = vi.sanitizeValue(fieldValue.Interface(), fieldName)
}

return result
}

func (vi *validatorImpl) sanitizeValue(value any, fieldName string) interface{} {
if !vi.config.SanitizeSensitiveData {
return value
}

if vi.isSensitiveField(fieldName) {
return "***REDACTED***"
}

return value
}

func (vi *validatorImpl) isSensitiveField(fieldName string) bool {
vi.mu.RLock()
defer vi.mu.RUnlock()

lowerField := strings.ToLower(fieldName)
return vi.sensitiveFields[lowerField]
}
