package database_test

import (
"os"
"testing"

"github.com/marcelofabianov/database"
)

func TestLoadConfig(t *testing.T) {
origHost := os.Getenv("DATABASE_HOST")
origPort := os.Getenv("DATABASE_PORT")
defer func() {
os.Setenv("DATABASE_HOST", origHost)
os.Setenv("DATABASE_PORT", origPort)
}()

t.Run("loads defaults when no env vars set", func(t *testing.T) {
os.Unsetenv("DATABASE_HOST")
os.Unsetenv("DATABASE_PORT")

cfg, err := database.LoadConfig()
if err != nil {
t.Fatalf("LoadConfig() error = %v", err)
}

if cfg.Database.Credentials.Host != "localhost" {
t.Errorf("expected host localhost, got %s", cfg.Database.Credentials.Host)
}
if cfg.Database.Credentials.Port != 5432 {
t.Errorf("expected port 5432, got %d", cfg.Database.Credentials.Port)
}
if cfg.Database.Pool.MaxOpenConns != 25 {
t.Errorf("expected max open conns 25, got %d", cfg.Database.Pool.MaxOpenConns)
}
})

t.Run("loads from environment variables", func(t *testing.T) {
os.Setenv("DATABASE_HOST", "db-server")
os.Setenv("DATABASE_PORT", "5433")
os.Setenv("DATABASE_USER", "testuser")
os.Setenv("DATABASE_PASSWORD", "secret")
os.Setenv("DATABASE_NAME", "testdb")
defer func() {
os.Unsetenv("DATABASE_USER")
os.Unsetenv("DATABASE_PASSWORD")
os.Unsetenv("DATABASE_NAME")
}()

cfg, err := database.LoadConfig()
if err != nil {
t.Fatalf("LoadConfig() error = %v", err)
}

if cfg.Database.Credentials.Host != "db-server" {
t.Errorf("expected host db-server, got %s", cfg.Database.Credentials.Host)
}
if cfg.Database.Credentials.Port != 5433 {
t.Errorf("expected port 5433, got %d", cfg.Database.Credentials.Port)
}
if cfg.Database.Credentials.User != "testuser" {
t.Errorf("expected user testuser, got %s", cfg.Database.Credentials.User)
}
})

t.Run("validates invalid port", func(t *testing.T) {
os.Setenv("DATABASE_PORT", "99999")
defer os.Unsetenv("DATABASE_PORT")

_, err := database.LoadConfig()
if err == nil {
t.Error("expected error for invalid port")
}
})
}

func TestGetDatabaseDSN(t *testing.T) {
cfg := &database.Config{
Database: database.DatabaseConfig{
Credentials: database.DatabaseCredentialsConfig{
Host:     "localhost",
Port:     5432,
User:     "postgres",
Password: "secret",
Name:     "testdb",
SSLMode:  "disable",
},
},
}

dsn := cfg.GetDatabaseDSN()
expected := "host=localhost port=5432 user=postgres password=secret dbname=testdb sslmode=disable"

if dsn != expected {
t.Errorf("expected DSN %s, got %s", expected, dsn)
}
}

func TestConfigValidation(t *testing.T) {
tests := []struct {
name    string
config  *database.Config
wantErr bool
}{
{
name: "valid config",
config: &database.Config{
Database: database.DatabaseConfig{
Credentials: database.DatabaseCredentialsConfig{
Host: "localhost",
Port: 5432,
User: "postgres",
Name: "testdb",
},
Pool: database.DatabasePoolConfig{
MaxOpenConns: 10,
MaxIdleConns: 5,
},
Connect: database.DatabaseConnectConfig{
BackoffRetries: 3,
},
},
},
wantErr: false,
},
{
name: "empty host",
config: &database.Config{
Database: database.DatabaseConfig{
Credentials: database.DatabaseCredentialsConfig{
Host: "",
Port: 5432,
User: "postgres",
Name: "testdb",
},
Pool: database.DatabasePoolConfig{
MaxOpenConns: 10,
},
},
},
wantErr: true,
},
{
name: "invalid max open conns",
config: &database.Config{
Database: database.DatabaseConfig{
Credentials: database.DatabaseCredentialsConfig{
Host: "localhost",
Port: 5432,
User: "postgres",
Name: "testdb",
},
Pool: database.DatabasePoolConfig{
MaxOpenConns: 0,
},
},
},
wantErr: true,
},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
err := database.ValidateConfig(tt.config)
if (err != nil) != tt.wantErr {
t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
}
})
}
}
