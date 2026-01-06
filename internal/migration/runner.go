package migration

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Runner executes SQL migrations stored on disk.
type Runner struct {
	db  *sql.DB
	dir string
}

// NewRunner constructs a migration runner for the given directory.
func NewRunner(db *sql.DB, dir string) *Runner {
	return &Runner{db: db, dir: dir}
}

// Run applies any pending migrations.
func (r *Runner) Run(ctx context.Context) error {
	if err := r.ensureTable(ctx); err != nil {
		return err
	}

	files, err := filepath.Glob(filepath.Join(r.dir, "*.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)

	for _, file := range files {
		version := filepath.Base(file)
		applied, err := r.isApplied(ctx, version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		statements, err := readStatements(file)
		if err != nil {
			return fmt.Errorf("migration %s: %w", version, err)
		}

		if err := r.applyMigration(ctx, version, statements); err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) ensureTable(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
        version VARCHAR(255) PRIMARY KEY,
        applied_at DATETIME NOT NULL
    )`)
	return err
}

func (r *Runner) isApplied(ctx context.Context, version string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM schema_migrations WHERE version = ?`, version).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Runner) applyMigration(ctx context.Context, version string, statements []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for _, stmt := range statements {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s failed: %w", version, err)
		}
	}

	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`, version, time.Now().UTC()); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func readStatements(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var (
		statements []string
		builder    strings.Builder
	)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}

		builder.WriteString(line)
		builder.WriteString("\n")

		if strings.HasSuffix(strings.TrimSpace(line), ";") {
			stmt := strings.TrimSpace(builder.String())
			stmt = strings.TrimSuffix(stmt, ";")
			stmt = strings.TrimSpace(stmt)
			if stmt != "" {
				statements = append(statements, stmt)
			}
			builder.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if strings.TrimSpace(builder.String()) != "" {
		statements = append(statements, strings.TrimSpace(builder.String()))
	}

	return statements, nil
}
