package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const (
	dbFileName = "bagtracker.db"
)

var DB *sql.DB

func Init() error {
	// Create data directory if it doesn't exist
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	dbPath := filepath.Join(dataDir, dbFileName)
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	// Enable foreign keys and other SQLite settings
	if _, err := DB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return err
	}
	if _, err := DB.Exec("PRAGMA journal_mode = WAL"); err != nil {
		log.Printf("Warning: could not set WAL mode: %v", err)
	}

	// Create tables
	if err := createTables(); err != nil {
		return err
	}

	log.Printf("Database initialized at %s", dbPath)
	return nil
}

func createTables() error {
	// Bag rail tables
	bagRailTables := []string{
		"clean_bag_rail",
		"ironer_bag_rail",
		"sorting_bag_rail",
	}

	for _, table := range bagRailTables {
		query := `
		CREATE TABLE IF NOT EXISTS ` + table + ` (
			uuid TEXT PRIMARY KEY,
			bag_number INTEGER NOT NULL,
			fixed BOOLEAN NOT NULL DEFAULT FALSE,
			in_production BOOLEAN NOT NULL DEFAULT FALSE
		);
		`
		if _, err := DB.Exec(query); err != nil {
			return err
		}
	}

	// Work order tables
	workOrderTables := []string{
		"clean_bag_work_order",
		"ironer_bag_work_order",
		"sorting_bag_work_order",
	}

	for _, table := range workOrderTables {
		query := `
		CREATE TABLE IF NOT EXISTS ` + table + ` (
			uuid TEXT PRIMARY KEY,
			bag_number INTEGER NOT NULL,
			date DATETIME NOT NULL,
			work_request_order BOOLEAN NOT NULL DEFAULT FALSE,
			work_completion_order BOOLEAN NOT NULL DEFAULT FALSE,
			main_trolley BOOLEAN,
			secondary_trolley BOOLEAN,
			clutch BOOLEAN,
			large_rope BOOLEAN,
			small_rope BOOLEAN,
			bag_weight BOOLEAN,
			rope_weight BOOLEAN,
			frame BOOLEAN,
			carabina BOOLEAN,
			fabric BOOLEAN
		);
		`
		if _, err := DB.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
