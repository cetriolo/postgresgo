package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context) (*DB, error) {
	connStr := getConnectionString()

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

func (d *DB) Close() {
	if d.Pool != nil {
		d.Pool.Close()
	}
}

func (d *DB) Ping(ctx context.Context) error {
	return d.Pool.Ping(ctx)
}

func getConnectionString() string {
	return "postgresql://postgres.ayflwtkdrsitdqkodsan:BZhFhWPjcJxeHYIt@aws-0-us-east-1.pooler.supabase.com:6543/postgres"
	//return "postgresql://postgres.ayflwtkdrsitdqkodsan:BZhFhWPjcJxeHYIt@ayflwtkdrsitdqkodsan.pooler.supabase.com:6543/postgres"
	//return `postgresql://postgres.ayflwtkdrsitdqkodsan:CJ5V6FgxnmQi5PZy@aws-0-us-east-1.pooler.supabase.com:6543/postgres`
	//return "postgresql://postgres.ayflwtkdrsitdqkodsan:CJ5V6FgxnmQi5PZy@aws-0-us-east-1.pooler.supabase.com:5432/postgres"
	//return "postgresql://postgres.ajrbwkcuthywddfihrmflo:[YOUR-PASSWORD]@aws-0-us-east-1.pooler.supabase.com:6543/postgres"
}

//	if connStr := os.Getenv("DATABASE_URL"); connStr != "" {
//		return connStr
//	}
//
//	host := getEnv("DB_HOST", "localhost")
//	port := getEnv("DB_PORT", "5432")
//	user := getEnv("DB_USER", "postgres")
//	password := getEnv("DB_PASSWORD", "postgres")
//	dbname := getEnv("DB_NAME", "postgres")
//	sslmode := getEnv("DB_SSLMODE", "disable")
//
//	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
//		host, port, user, password, dbname, sslmode)
//}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
