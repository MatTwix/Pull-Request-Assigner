package app

import (
	"errors"
	"fmt"
	"time"

	"github.com/MatTwix/Pull-Request-Assigner/pkg/logger"
	"github.com/golang-migrate/migrate/v4"

	// migrate tools
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	defaultAttempts = 20
	defaultTimeout  = time.Second
)

func runMigrations(logger logger.Logger, url string) error {
	if len(url) == 0 {
		return fmt.Errorf("pg URL is not declareted")
	}

	url += "?sslmode=disable"

	var (
		attempts = defaultAttempts
		err      error
		m        *migrate.Migrate
	)

	for attempts > 0 {
		m, err = migrate.New("file://migrations", url)
		if err == nil {
			break
		}

		logger.Info("trying to connect to db", map[string]any{"attempts left": attempts})
		time.Sleep(defaultTimeout)
		attempts--
	}

	if err != nil {
		return fmt.Errorf("db connection error: %w", err)
	}

	err = m.Up()
	defer func() { _, _ = m.Close() }()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration up error: %w", err)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		logger.Info("Migrate: no change")
		return nil
	}

	logger.Info("Migrate: up success")
	return nil
}
