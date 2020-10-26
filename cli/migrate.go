package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	// Migrate configuration
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/golang-migrate/migrate/v4"

	"github.com/figment-networks/polkadothub-indexer/config"
)

func startMigrations(cfg *config.Config, version uint) error {
	log.Println("getting current directory")
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	srcDir := filepath.Join(dir, "migrations")
	srcPath := fmt.Sprintf("file://%s", srcDir)

	log.Println("using migrations from", srcDir)
	migrations, err := migrate.New(srcPath, cfg.DatabaseDSN)
	if err != nil {
		return err
	}

	if version > 0 {
		log.Println("running migrations to version ", version)
		return migrations.Migrate(version)
	}

	log.Println("running migrations up")
	return migrations.Up()
}
