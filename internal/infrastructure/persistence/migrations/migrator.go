package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
)

//go:embed sql/*.sql
var sqlFiles embed.FS

// Migration represents a single database migration
type Migration struct {
	Version   int
	Name      string
	SQL       string
	AppliedAt time.Time
}

// MigrationRecord tracks applied migrations in the database
type MigrationRecord struct {
	ID        uint      `gorm:"primaryKey"`
	Version   int       `gorm:"uniqueIndex;not null"`
	Name      string    `gorm:"not null"`
	AppliedAt time.Time `gorm:"not null"`
}

// Migrator handles database migrations
type Migrator struct {
	db     *gorm.DB
	rawDB  *sql.DB
	log    *slog.Logger
	dryRun bool
}

// NewMigrator creates a new migration manager
func NewMigrator(db *gorm.DB, rawDB *sql.DB, log *slog.Logger) *Migrator {
	return &Migrator{
		db:     db,
		rawDB:  rawDB,
		log:    log,
		dryRun: false,
	}
}

// SetDryRun enables/disables dry-run mode
func (m *Migrator) SetDryRun(enabled bool) {
	m.dryRun = enabled
}

// Init initializes the migrations tracking table
func (m *Migrator) Init() error {
	if m.db == nil {
		return fmt.Errorf("postgres database not initialized")
	}

	// Create migrations table if it doesn't exist
	if err := m.db.AutoMigrate(&MigrationRecord{}); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	m.log.Info("Migrations table initialized")
	return nil
}

// loadMigrations loads all migration files from the embedded filesystem
func (m *Migrator) loadMigrations() ([]Migration, error) {
	entries, err := sqlFiles.ReadDir("sql")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		// Parse version from filename (e.g., "001_initial_schema.sql")
		var version int
		var name string
		_, err := fmt.Sscanf(entry.Name(), "%d_%s", &version, &name)
		if err != nil {
			m.log.Warn("Skipping invalid migration file", "file", entry.Name(), "error", err)
			continue
		}

		// Remove .sql extension
		name = strings.TrimSuffix(name, ".sql")

		// Read SQL content
		content, err := sqlFiles.ReadFile("sql/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to read migration %s: %w", entry.Name(), err)
		}

		migrations = append(migrations, Migration{
			Version: version,
			Name:    name,
			SQL:     string(content),
		})
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// getAppliedMigrations retrieves list of applied migrations from database
func (m *Migrator) getAppliedMigrations() (map[int]bool, error) {
	if m.db == nil {
		return nil, fmt.Errorf("postgres database not initialized")
	}

	var records []MigrationRecord
	if err := m.db.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to query migrations: %w", err)
	}

	applied := make(map[int]bool)
	for _, record := range records {
		applied[record.Version] = true
	}

	return applied, nil
}

// recordMigration records a successful migration in the database
func (m *Migrator) recordMigration(migration Migration) error {
	if m.db == nil {
		return fmt.Errorf("postgres database not initialized")
	}

	record := MigrationRecord{
		Version:   migration.Version,
		Name:      migration.Name,
		AppliedAt: time.Now(),
	}

	return m.db.Create(&record).Error
}

// Migrate runs all pending migrations
func (m *Migrator) Migrate() error {
	// Initialize migrations table
	if err := m.Init(); err != nil {
		return err
	}

	// Load all migrations
	migrations, err := m.loadMigrations()
	if err != nil {
		return err
	}

	if len(migrations) == 0 {
		m.log.Info("No migrations found")
		return nil
	}

	// Get applied migrations
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return err
	}

	// Find pending migrations
	var pending []Migration
	for _, migration := range migrations {
		if !applied[migration.Version] {
			pending = append(pending, migration)
		}
	}

	if len(pending) == 0 {
		m.log.Info("No pending migrations")
		return nil
	}

	m.log.Info("Found pending migrations", "count", len(pending))

	// Apply each pending migration
	for _, migration := range pending {
		if err := m.applyMigration(migration); err != nil {
			return fmt.Errorf("migration %d (%s) failed: %w", migration.Version, migration.Name, err)
		}
	}

	m.log.Info("All migrations applied successfully")
	return nil
}

// applyMigration applies a single migration
func (m *Migrator) applyMigration(migration Migration) error {
	m.log.Info("Applying migration", "version", migration.Version, "name", migration.Name)

	if m.dryRun {
		m.log.Info("DRY RUN - Migration SQL:", "sql", migration.SQL)
		return nil
	}

	// Determine which database connection to use
	// If migration contains TimescaleDB-specific functions, use rawDB
	db := m.db
	useRawDB := strings.Contains(migration.SQL, "create_hypertable") ||
		strings.Contains(migration.SQL, "add_retention_policy") ||
		strings.Contains(migration.SQL, "add_compression_policy") ||
		strings.Contains(migration.SQL, "timescaledb")

	if useRawDB {
		if m.rawDB == nil {
			return fmt.Errorf("timescaledb connection required but not initialized")
		}

		// Execute with raw SQL connection
		if _, err := m.rawDB.Exec(migration.SQL); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	} else {
		// Execute with GORM
		if err := db.Exec(migration.SQL).Error; err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	// Record migration
	if err := m.recordMigration(migration); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	m.log.Info("Migration applied successfully", "version", migration.Version)
	return nil
}

// Status returns the current migration status
func (m *Migrator) Status() error {
	migrations, err := m.loadMigrations()
	if err != nil {
		return err
	}

	applied, err := m.getAppliedMigrations()
	if err != nil {
		return err
	}

	m.log.Info("Migration Status")
	m.log.Info("================")
	m.log.Info("Total migrations", "count", len(migrations))
	m.log.Info("Applied migrations", "count", len(applied))

	for _, migration := range migrations {
		status := "PENDING"
		if applied[migration.Version] {
			status = "APPLIED"
		}
		m.log.Info("Migration", "version", migration.Version, "name", migration.Name, "status", status)
	}

	return nil
}
