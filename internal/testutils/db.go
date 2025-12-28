package testutils

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var defaultDBConfig = struct {
	Image    string
	Port     string
	User     string
	Password string
	DBName   string
}{
	Image:    "postgres:13",
	Port:     "5432",
	User:     "test_user",
	Password: "test_password",
	DBName:   "test_db",
}

type DBSetup struct {
	DB        *pgxpool.Pool
	Container testcontainers.Container
}

func SetupTestDB() (*DBSetup, error) {
	cfg := defaultDBConfig
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        cfg.Image,
		ExposedPorts: []string{cfg.Port + "/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       cfg.DBName,
			"POSTGRES_USER":     cfg.User,
			"POSTGRES_PASSWORD": cfg.Password,
		},
		WaitingFor: wait.ForListeningPort(nat.Port(cfg.Port + "/tcp")),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, nat.Port(cfg.Port+"/tcp"))
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	pgxDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port.Port(), cfg.User, cfg.Password, cfg.DBName)

	migrateDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.User, cfg.Password, host, port.Port(), cfg.DBName)

	connectCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := pgxpool.New(connectCtx, pgxDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	m, err := migrate.New("file://../../../migrations", migrateDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return &DBSetup{
		DB:        db,
		Container: container,
	}, nil
}

func TruncateTables(db *pgxpool.Pool) error {
	_, err := db.Exec(context.Background(), "TRUNCATE TABLE users RESTART IDENTITY CASCADE;")
	if err != nil {
		return fmt.Errorf("failed to truncate tables: %w", err)
	}
	return nil
}

func CloseDBSetup(setup *DBSetup) {
	if setup.DB != nil {
		setup.DB.Close()
	}
	if setup.Container != nil {
		_ = setup.Container.Terminate(context.Background())
	}
}
