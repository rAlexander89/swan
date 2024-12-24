package app

import (
	"fmt"
	"os"
	"path/filepath"
)

func WritePostgres(projectPath string) error {
	postgresCode := postgresGoFile()

	postgresPath := filepath.Join("internal", "app", "repositories", "postgres", "postgres.go")

	if err := os.WriteFile(postgresPath, []byte(postgresCode), 0644); err != nil {
		return fmt.Errorf("failed to write postgres.go: %v", err)
	}

	return nil
}

func postgresGoFile() string {
	genCode := `// internal/app/repositories/postgres/postgres.go
package postgres

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    
    _ "github.com/lib/pq"
)

type Config struct {
    URI                   string
    MaxOpenConnections    int
    MaxIdleConnections    int
    MaxConnectionIdleTime int
    MaxConnectionLifetime int
}

func New(ctx context.Context, cfg Config) (*Connection, error) {
    if cfg.URI == "" {
        return nil, fmt.Errorf("postgres uri cannot be empty")
    }

    db, err := sql.Open("postgres", cfg.URI)
    if err != nil {
        return nil, fmt.Errorf("failed to open postgres connection: %w", err)
    }

    // verify connection with a simple query
    var version string
    err = db.QueryRowContext(ctx, "SELECT version()").Scan(&version)
    if err != nil {
        db.Close()
        return nil, fmt.Errorf("failed to verify postgres connection: %w", err)
    }

    fmt.Printf("connected to postgres version %s",  version)

    // set connection pool settings
    db.SetMaxOpenConns(cfg.MaxOpenConnections)
    db.SetMaxIdleConns(cfg.MaxIdleConnections)
    db.SetConnMaxIdleTime(time.Duration(cfg.MaxConnectionIdleTime) * time.Second)
    db.SetConnMaxLifetime(time.Duration(cfg.MaxConnectionLifetime) * time.Second)

    return &Connection{
        db: db,
    }, nil
}
  `

	return genCode
}

func WritePostgresConnectionFile(projectPath string) error {
	connContent := genPgConnCode()

	connectionPath := filepath.Join(projectPath, "internal", "app", "repositories", "postgres", "connection.go")

	if err := os.WriteFile(connectionPath, []byte(connContent), 0644); err != nil {
		return fmt.Errorf("failed to write connection.go: %v", err)
	}

	return nil
}

func genPgConnCode() string {
	connCode := `
// internal/app/repositories/postgres/connection.go
package postgres

import (
    "context"
    "database/sql"
    "fmt"
)

// Connection wraps sql.DB to provide a consistent interface for database operations
type Connection struct {
    db *sql.DB
}

// ExecContext wraps sql.DB.ExecContext
func (c *Connection) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
    if c.db == nil {
        return nil, fmt.Errorf("database connection is nil")
    }
    
    return c.db.ExecContext(ctx, query, args...)
}

// QueryContext wraps sql.DB.QueryContext
func (c *Connection) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    if c.db == nil {
        return nil, fmt.Errorf("database connection is nil")
    }
    
    return c.db.QueryContext(ctx, query, args...)
}

// QueryRowContext wraps sql.DB.QueryRowContext
func (c *Connection) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
    if c.db == nil {
        return nil
    }
    
    return c.db.QueryRowContext(ctx, query, args...)
}

// BeginTx wraps sql.DB.BeginTx
func (c *Connection) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
    if c.db == nil {
        return nil, fmt.Errorf("database connection is nil")
    }
    
    return c.db.BeginTx(ctx, opts)
}

// Close closes the database connection
func (c *Connection) Close() error {
  if c.db == nil {
        return nil
    }
    
    return c.db.Close()
}

`
	return connCode
}

func WriteRepository(projPath string) error {
	repoContent := genRepositoryCode()

	repoPath := filepath.Join(projPath, "internal", "app", "repositories", "postgres", "repository.go")

	if err := os.WriteFile(repoPath, []byte(repoContent), 0644); err != nil {
		return fmt.Errorf("failed to write repository.go: %v", err)
	}

	return nil
}

func genRepositoryCode() string {
	repoCode := `
// internal/app/repositories/postgres/repository.go
package postgres

import (
    "context"
    "database/sql"
    "fmt"
    "sync"
)

// Repository provides a base implementation for all domain repositories
type Repository struct {
    conn *Connection
}

// ensures thread-safe singleton creation
var (
    repo *Repository
    once sync.Once
)

// NewRepository creates a singleton postgres repository
func NewRepository(ctx context.Context, cfg Config) (*Repository, error) {
    var initErr error

    once.Do(func() {
        conn, err := New(ctx, cfg)
        if err != nil {
            initErr = fmt.Errorf("failed to initialize postgres connection: %w", err)
            return
        }

        repo = &Repository{
            conn: conn,
        }
    })

    if initErr != nil {
        return nil, initErr
    }

    return repo, nil
}

// GetConnection returns the underlying database connection
func (r *Repository) GetConnection() *Connection {
    return r.conn
}

// Close closes the database connection
func (r *Repository) Close() error {
    if r.conn == nil {
        return nil
    }
    
    return r.conn.Close()
}

// Transaction executes the given function within a database transaction
func (r *Repository) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
    if r.conn == nil {
        return fmt.Errorf("repository connection is nil")
    }

    tx, err := r.conn.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }

    // if fn returns an error, rollback the transaction
    if err := fn(tx); err != nil {
        if rbErr := tx.Rollback(); rbErr != nil {
            return fmt.Errorf("error rolling back transaction: %v (original error: %w)", rbErr, err)
        }
        return err
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    return nil
}

  `

	return repoCode
}
