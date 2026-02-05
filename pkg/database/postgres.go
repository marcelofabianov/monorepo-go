package database

import (
"context"
"database/sql"
"log/slog"
"time"

"github.com/marcelofabianov/fault"

_ "github.com/jackc/pgx/v5/stdlib"
)

var (
ErrConnectionFailed = fault.New(
"database connection failed after retries",
fault.WithCode(fault.InfraError),
)

ErrInvalidConfig = fault.New(
"invalid database configuration",
fault.WithCode(fault.Invalid),
)

ErrAlreadyConnected = fault.New(
"database already connected",
fault.WithCode(fault.Conflict),
)

ErrNotConnected = fault.New(
"database not connected",
fault.WithCode(fault.NotFound),
)

ErrOpenFailed = fault.New(
"failed to open database connection",
fault.WithCode(fault.InfraError),
)

ErrPingFailed = fault.New(
"failed to ping database",
fault.WithCode(fault.InfraError),
)

ErrCloseFailed = fault.New(
"failed to close database connection",
fault.WithCode(fault.Internal),
)

ErrExecFailed = fault.New(
"failed to execute query",
fault.WithCode(fault.Internal),
)

ErrQueryFailed = fault.New(
"failed to execute query",
fault.WithCode(fault.Internal),
)

ErrTransactionFailed = fault.New(
"failed to begin transaction",
fault.WithCode(fault.Internal),
)
)

type DB struct {
conn   *sql.DB
config *Config
logger *slog.Logger
}

func New(cfg *Config, logger *slog.Logger) (*DB, error) {
if cfg == nil {
return nil, ErrInvalidConfig
}

if logger == nil {
logger = slog.Default()
}

return &DB{
config: cfg,
logger: logger,
}, nil
}

func (db *DB) SetLogger(logger *slog.Logger) {
if logger != nil {
db.logger = logger
}
}

func (db *DB) Connect(ctx context.Context) error {
if db.conn != nil {
return ErrAlreadyConnected
}

db.logger.Info("Connecting to database",
"host", db.config.Database.Credentials.Host,
"database", db.config.Database.Credentials.Name,
)

if err := db.connect(ctx); err != nil {
db.logger.Error("Failed to connect to database",
"host", db.config.Database.Credentials.Host,
"database", db.config.Database.Credentials.Name,
"error", err.Error(),
)
return err
}

db.logger.Info("Database connected successfully",
"host", db.config.Database.Credentials.Host,
"database", db.config.Database.Credentials.Name,
"pool_max_open", db.config.Database.Pool.MaxOpenConns,
"pool_max_idle", db.config.Database.Pool.MaxIdleConns,
)

return nil
}

func (db *DB) connect(ctx context.Context) error {
dsn := db.config.GetDatabaseDSN()

conn, err := sql.Open("pgx", dsn)
if err != nil {
return fault.Wrap(ErrOpenFailed, "sql.Open failed",
fault.WithWrappedErr(err),
fault.WithContext("driver", "pgx"),
)
}

db.configurePool(conn)

pingCtx, cancel := context.WithTimeout(ctx, db.config.Database.Connect.QueryTimeout)
defer cancel()

if err := conn.PingContext(pingCtx); err != nil {
_ = conn.Close()
return fault.Wrap(ErrPingFailed, "ping failed",
fault.WithWrappedErr(err),
fault.WithContext("timeout", db.config.Database.Connect.QueryTimeout.String()),
)
}

db.conn = conn
return nil
}

func (db *DB) configurePool(conn *sql.DB) {
poolConfig := db.config.Database.Pool

conn.SetMaxOpenConns(poolConfig.MaxOpenConns)
conn.SetMaxIdleConns(poolConfig.MaxIdleConns)
conn.SetConnMaxLifetime(poolConfig.ConnMaxLifetime)
conn.SetConnMaxIdleTime(poolConfig.ConnMaxIdleTime)
}

func (db *DB) Close() error {
if db.conn == nil {
return ErrNotConnected
}

db.logger.Info("Closing database connection")

if err := db.conn.Close(); err != nil {
return fault.Wrap(ErrCloseFailed, "close failed",
fault.WithWrappedErr(err),
)
}

db.conn = nil
return nil
}

func (db *DB) Ping(ctx context.Context) error {
if db.conn == nil {
return ErrNotConnected
}

pingCtx, cancel := context.WithTimeout(ctx, db.config.Database.Connect.QueryTimeout)
defer cancel()

if err := db.conn.PingContext(pingCtx); err != nil {
return fault.Wrap(ErrPingFailed, "ping failed",
fault.WithWrappedErr(err),
fault.WithContext("timeout", db.config.Database.Connect.QueryTimeout.String()),
)
}

return nil
}

func (db *DB) HealthCheck(ctx context.Context) error {
if db.conn == nil {
return ErrNotConnected
}

if err := db.Ping(ctx); err != nil {
return err
}

stats := db.conn.Stats()

if stats.InUse >= stats.MaxOpenConnections {
db.logger.Warn("All database connections are in use",
"in_use", stats.InUse,
"max_open", stats.MaxOpenConnections,
)
}

if stats.WaitCount > 0 {
db.logger.Warn("Database connections waiting",
"wait_count", stats.WaitCount,
"wait_duration", stats.WaitDuration,
)
}

return nil
}

func (db *DB) Stats() sql.DBStats {
if db.conn == nil {
return sql.DBStats{}
}
return db.conn.Stats()
}

func (db *DB) DB() *sql.DB {
return db.conn
}

func (db *DB) IsConnected() bool {
return db.conn != nil
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
if db.conn == nil {
return nil, ErrNotConnected
}

execCtx, cancel := context.WithTimeout(ctx, db.config.Database.Connect.ExecTimeout)
defer cancel()

result, err := db.conn.ExecContext(execCtx, query, args...)
if err != nil {
db.logger.Error("Query execution failed",
"query", query,
"timeout", db.config.Database.Connect.ExecTimeout.String(),
"error", err.Error(),
)
return nil, fault.Wrap(ErrExecFailed, "exec failed",
fault.WithWrappedErr(err),
fault.WithContext("query", query),
fault.WithContext("timeout", db.config.Database.Connect.ExecTimeout.String()),
)
}

return result, nil
}

func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
if db.conn == nil {
return nil, ErrNotConnected
}

queryCtx, cancel := context.WithTimeout(ctx, db.config.Database.Connect.QueryTimeout)
defer cancel()

rows, err := db.conn.QueryContext(queryCtx, query, args...)
if err != nil {
db.logger.Error("Query failed",
"query", query,
"timeout", db.config.Database.Connect.QueryTimeout.String(),
"error", err.Error(),
)
return nil, fault.Wrap(ErrQueryFailed, "query failed",
fault.WithWrappedErr(err),
fault.WithContext("query", query),
fault.WithContext("timeout", db.config.Database.Connect.QueryTimeout.String()),
)
}

return rows, nil
}

func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
if db.conn == nil {
return nil
}

queryCtx, cancel := context.WithTimeout(ctx, db.config.Database.Connect.QueryTimeout)
defer cancel()

return db.conn.QueryRowContext(queryCtx, query, args...)
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
if db.conn == nil {
return nil, ErrNotConnected
}

tx, err := db.conn.BeginTx(ctx, opts)
if err != nil {
db.logger.Error("Failed to begin transaction", "error", err.Error())
return nil, fault.Wrap(ErrTransactionFailed, "begin transaction failed",
fault.WithWrappedErr(err),
)
}

return tx, nil
}

func (db *DB) StartHealthCheckRoutine(ctx context.Context) {
if db.conn == nil {
db.logger.Error("Cannot start health check routine: database not connected")
return
}

period := db.config.Database.Pool.HealthCheckPeriod
ticker := time.NewTicker(period)

go func() {
defer ticker.Stop()

for {
select {
case <-ctx.Done():
db.logger.Info("Health check routine stopped")
return
case <-ticker.C:
if err := db.HealthCheck(context.Background()); err != nil {
db.logger.Error("Health check failed", "error", err)
} else {
stats := db.Stats()
db.logger.Debug("Database health check passed",
"open_connections", stats.OpenConnections,
"in_use", stats.InUse,
"idle", stats.Idle,
)
}
}
}
}()

db.logger.Info("Health check routine started", "period", period)
}
