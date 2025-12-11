package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"
)

func (d *DB) RunMigrations(ctx context.Context, migrationsPath string) error {
	if err := d.createMigrationsTable(ctx); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	appliedMigrations, err := d.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var sqlFiles []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".sql" {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}

	sort.Strings(sqlFiles)

	migrationsRun := 0
	for _, file := range sqlFiles {
		if appliedMigrations[file] {
			fmt.Printf("Skipping already applied migration: %s\n", file)
			continue
		}

		filePath := filepath.Join(migrationsPath, file)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		tx, err := d.Pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", file, err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}

		if err := d.recordMigration(ctx, tx, file); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to record migration %s: %w", file, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", file, err)
		}

		fmt.Printf("Applied migration: %s\n", file)
		migrationsRun++
	}

	if migrationsRun == 0 {
		fmt.Println("No new migrations to apply")
	} else {
		fmt.Printf("Successfully applied %d migration(s)\n", migrationsRun)
	}

	return nil
}

func (d *DB) createMigrationsTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			migration_name VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_schema_migrations_name ON schema_migrations(migration_name);
	`
	_, err := d.Pool.Exec(ctx, query)
	return err
}

func (d *DB) getAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	rows, err := d.Pool.Query(ctx, "SELECT migration_name FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		applied[name] = true
	}

	return applied, rows.Err()
}

func (d *DB) recordMigration(ctx context.Context, tx pgx.Tx, migrationName string) error {
	_, err := tx.Exec(ctx,
		"INSERT INTO schema_migrations (migration_name, applied_at) VALUES ($1, $2)",
		migrationName, time.Now())
	return err
}
