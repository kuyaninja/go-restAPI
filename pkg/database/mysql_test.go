package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"sync/atomic"
	"testing"

	"skeleton-go/pkg/config"
)

func TestNewMySQLBuildsDSNAndConfiguresPool(t *testing.T) {
	var (
		receivedDriver string
		receivedDSN    string
		pingCount      int32
	)

	originalOpen := sqlOpen
	sqlOpen = func(driverName, dataSourceName string) (*sql.DB, error) {
		receivedDriver = driverName
		receivedDSN = dataSourceName
		return sql.OpenDB(stubConnector{pingCount: &pingCount}), nil
	}
	defer func() { sqlOpen = originalOpen }()

	cfg := config.DatabaseConfig{
		Host:     "db.example.com",
		Port:     3307,
		User:     "user",
		Password: "pass",
		Name:     "appdb",
	}

	db, err := NewMySQL(cfg)
	if err != nil {
		t.Fatalf("NewMySQL returned error: %v", err)
	}
	defer db.Close()

	expectedDSN := "user:pass@tcp(db.example.com:3307)/appdb?parseTime=true"
	if receivedDSN != expectedDSN {
		t.Fatalf("dsn mismatch: got %q want %q", receivedDSN, expectedDSN)
	}
	if receivedDriver != mysqlDriverName {
		t.Fatalf("driver mismatch: got %q want %q", receivedDriver, mysqlDriverName)
	}
	if got := db.Stats().MaxOpenConnections; got != 25 {
		t.Fatalf("max open connections mismatch: got %d", got)
	}
	if atomic.LoadInt32(&pingCount) != 1 {
		t.Fatalf("expected ping to be called once, got %d", pingCount)
	}
}

func TestNewMySQLOpenError(t *testing.T) {
	originalOpen := sqlOpen
	sqlOpen = func(string, string) (*sql.DB, error) {
		return nil, errors.New("boom")
	}
	defer func() { sqlOpen = originalOpen }()

	if _, err := NewMySQL(config.DatabaseConfig{}); err == nil {
		t.Fatalf("expected error when open fails")
	}
}

func TestNewMySQLPingError(t *testing.T) {
	originalOpen := sqlOpen
	sqlOpen = func(string, string) (*sql.DB, error) {
		return sql.OpenDB(stubConnector{pingErr: errors.New("ping fail")}), nil
	}
	defer func() { sqlOpen = originalOpen }()

	if _, err := NewMySQL(config.DatabaseConfig{}); err == nil {
		t.Fatalf("expected ping error")
	}
}

type stubConnector struct {
	pingCount *int32
	pingErr   error
}

func (c stubConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return &stubConn{pingCount: c.pingCount, pingErr: c.pingErr}, nil
}

func (c stubConnector) Driver() driver.Driver {
	return stubDriver{pingCount: c.pingCount, pingErr: c.pingErr}
}

type stubDriver struct {
	pingCount *int32
	pingErr   error
}

func (d stubDriver) Open(name string) (driver.Conn, error) {
	return &stubConn{pingCount: d.pingCount, pingErr: d.pingErr}, nil
}

type stubConn struct {
	pingCount *int32
	pingErr   error
}

func (c *stubConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("not implemented") }
func (c *stubConn) Close() error                        { return nil }
func (c *stubConn) Begin() (driver.Tx, error)           { return nil, errors.New("not implemented") }
func (c *stubConn) Ping(context.Context) error {
	if c.pingCount != nil {
		atomic.AddInt32(c.pingCount, 1)
	}
	return c.pingErr
}
