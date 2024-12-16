package project

import (
	"fmt"
	"os"
	"path/filepath"
)

func WritePostgresRepository(projectPath string) error {
	repoContent := `package postgres

import (
    "context"
    "sync"
)

type Repository struct {
    conn *Connection
}

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
    if r.conn != nil {
        return r.conn.Close()
    }
    return nil
}`

	repoPath := filepath.Join(projectPath, "internal", "app", "repositories", "postgres", "repository.go")

	if err := os.WriteFile(repoPath, []byte(repoContent), 0644); err != nil {
		return fmt.Errorf("failed to write repository.go: %v", err)
	}

	return nil
}
