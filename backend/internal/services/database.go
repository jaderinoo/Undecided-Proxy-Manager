package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"upm-backend/internal/config"
	"upm-backend/internal/models"

	_ "modernc.org/sqlite"
)

type DatabaseService struct {
	db *sql.DB
}

func NewDatabaseService() (*DatabaseService, error) {
	cfg := config.Load()
	
	db, err := sql.Open("sqlite", cfg.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	service := &DatabaseService{db: db}
	
	// Initialize tables
	if err := service.initTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	log.Printf("Database initialized at: %s", cfg.DatabasePath)
	return service, nil
}

func (d *DatabaseService) initTables() error {
	// Create proxies table
	proxyTable := `
	CREATE TABLE IF NOT EXISTS proxies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		domain TEXT NOT NULL UNIQUE,
		target_url TEXT NOT NULL,
		ssl_enabled BOOLEAN DEFAULT FALSE,
		ssl_path TEXT,
		status TEXT DEFAULT 'active',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := d.db.Exec(proxyTable); err != nil {
		return fmt.Errorf("failed to create proxies table: %w", err)
	}

	// Create users table
	userTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password TEXT,
		is_active BOOLEAN DEFAULT TRUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := d.db.Exec(userTable); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create certificates table
	certTable := `
	CREATE TABLE IF NOT EXISTS certificates (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		domain TEXT NOT NULL UNIQUE,
		cert_path TEXT NOT NULL,
		key_path TEXT NOT NULL,
		expires_at DATETIME NOT NULL,
		is_valid BOOLEAN DEFAULT TRUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := d.db.Exec(certTable); err != nil {
		return fmt.Errorf("failed to create certificates table: %w", err)
	}

	return nil
}

// Proxy methods
func (d *DatabaseService) GetProxies() ([]models.Proxy, error) {
	query := `
		SELECT id, name, domain, target_url, ssl_enabled, ssl_path, status, created_at, updated_at
		FROM proxies
		ORDER BY created_at DESC`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query proxies: %w", err)
	}
	defer rows.Close()

	var proxies []models.Proxy
	for rows.Next() {
		var proxy models.Proxy
		err := rows.Scan(
			&proxy.ID,
			&proxy.Name,
			&proxy.Domain,
			&proxy.TargetURL,
			&proxy.SSLEnabled,
			&proxy.SSLPath,
			&proxy.Status,
			&proxy.CreatedAt,
			&proxy.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan proxy: %w", err)
		}
		proxies = append(proxies, proxy)
	}

	return proxies, nil
}

func (d *DatabaseService) GetProxy(id int) (*models.Proxy, error) {
	query := `
		SELECT id, name, domain, target_url, ssl_enabled, ssl_path, status, created_at, updated_at
		FROM proxies
		WHERE id = ?`

	var proxy models.Proxy
	err := d.db.QueryRow(query, id).Scan(
		&proxy.ID,
		&proxy.Name,
		&proxy.Domain,
		&proxy.TargetURL,
		&proxy.SSLEnabled,
		&proxy.SSLPath,
		&proxy.Status,
		&proxy.CreatedAt,
		&proxy.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("proxy not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query proxy: %w", err)
	}

	return &proxy, nil
}

func (d *DatabaseService) CreateProxy(proxy *models.Proxy) error {
	query := `
		INSERT INTO proxies (name, domain, target_url, ssl_enabled, ssl_path, status)
		VALUES (?, ?, ?, ?, ?, ?)`

	result, err := d.db.Exec(query, proxy.Name, proxy.Domain, proxy.TargetURL, proxy.SSLEnabled, proxy.SSLPath, proxy.Status)
	if err != nil {
		return fmt.Errorf("failed to insert proxy: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	proxy.ID = int(id)
	proxy.CreatedAt = time.Now()
	proxy.UpdatedAt = time.Now()

	return nil
}

func (d *DatabaseService) UpdateProxy(proxy *models.Proxy) error {
	query := `
		UPDATE proxies 
		SET name = ?, domain = ?, target_url = ?, ssl_enabled = ?, ssl_path = ?, status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	result, err := d.db.Exec(query, proxy.Name, proxy.Domain, proxy.TargetURL, proxy.SSLEnabled, proxy.SSLPath, proxy.Status, proxy.ID)
	if err != nil {
		return fmt.Errorf("failed to update proxy: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("proxy not found")
	}

	proxy.UpdatedAt = time.Now()
	return nil
}

func (d *DatabaseService) DeleteProxy(id int) error {
	query := `DELETE FROM proxies WHERE id = ?`

	result, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete proxy: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("proxy not found")
	}

	return nil
}

func (d *DatabaseService) Close() error {
	return d.db.Close()
}