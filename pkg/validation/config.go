package validation

import (
"os"
"path/filepath"
"strings"

"github.com/spf13/viper"
)

type Config struct {
EnableLogging             bool
SanitizeSensitiveData     bool
AdditionalSensitiveFields []string
LogSuccessfulValidations  bool
}

func LoadConfig() (*Config, error) {
v := viper.New()
v.SetEnvPrefix("VALIDATION")
v.AutomaticEnv()
v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

if envFile := findEnvFile(); envFile != "" {
v.SetConfigFile(envFile)
_ = v.ReadInConfig()
}

setDefaults(v)

cfg := &Config{
EnableLogging:             v.GetBool("enable_logging"),
SanitizeSensitiveData:     v.GetBool("sanitize_sensitive_data"),
AdditionalSensitiveFields: v.GetStringSlice("additional_sensitive_fields"),
LogSuccessfulValidations:  v.GetBool("log_successful_validations"),
}

return cfg, nil
}

func setDefaults(v *viper.Viper) {
v.SetDefault("enable_logging", true)
v.SetDefault("sanitize_sensitive_data", true)
v.SetDefault("additional_sensitive_fields", []string{})
v.SetDefault("log_successful_validations", false)
}

func findEnvFile() string {
dir, err := os.Getwd()
if err != nil {
return ""
}

for i := 0; i < 5; i++ {
envPath := filepath.Join(dir, ".env")
if _, err := os.Stat(envPath); err == nil {
return envPath
}
parent := filepath.Dir(dir)
if parent == dir {
break
}
dir = parent
}

return ""
}

func DefaultConfig() *Config {
return &Config{
EnableLogging:             true,
SanitizeSensitiveData:     true,
AdditionalSensitiveFields: []string{},
LogSuccessfulValidations:  false,
}
}
