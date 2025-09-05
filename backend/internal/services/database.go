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

	// Create DNS configurations table
	dnsConfigTable := `
	CREATE TABLE IF NOT EXISTS dns_configs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		provider TEXT NOT NULL,
		domain TEXT NOT NULL,
		username TEXT NOT NULL,
		password TEXT NOT NULL,
		is_active BOOLEAN DEFAULT TRUE,
		last_update DATETIME,
		last_ip TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := d.db.Exec(dnsConfigTable); err != nil {
		return fmt.Errorf("failed to create dns_configs table: %w", err)
	}

	// Create DNS records table
	dnsRecordsTable := `
	CREATE TABLE IF NOT EXISTS dns_records (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		config_id INTEGER NOT NULL,
		host TEXT NOT NULL,
		current_ip TEXT,
		last_update DATETIME,
		is_active BOOLEAN DEFAULT TRUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (config_id) REFERENCES dns_configs (id) ON DELETE CASCADE
	);`

	if _, err := d.db.Exec(dnsRecordsTable); err != nil {
		return fmt.Errorf("failed to create dns_records table: %w", err)
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

// DNS Config methods
func (d *DatabaseService) GetDNSConfigs() ([]models.DNSConfig, error) {
	query := `
		SELECT id, provider, domain, username, password, is_active, last_update, last_ip, created_at, updated_at
		FROM dns_configs
		ORDER BY created_at DESC`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query dns configs: %w", err)
	}
	defer rows.Close()

	var configs []models.DNSConfig
	for rows.Next() {
		var config models.DNSConfig
		err := rows.Scan(
			&config.ID,
			&config.Provider,
			&config.Domain,
			&config.Username,
			&config.Password,
			&config.IsActive,
			&config.LastUpdate,
			&config.LastIP,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan dns config: %w", err)
		}
		configs = append(configs, config)
	}

	return configs, nil
}

func (d *DatabaseService) GetDNSConfig(id int) (*models.DNSConfig, error) {
	query := `
		SELECT id, provider, domain, username, password, is_active, last_update, last_ip, created_at, updated_at
		FROM dns_configs
		WHERE id = ?`

	var config models.DNSConfig
	err := d.db.QueryRow(query, id).Scan(
		&config.ID,
		&config.Provider,
		&config.Domain,
		&config.Username,
		&config.Password,
		&config.IsActive,
		&config.LastUpdate,
		&config.LastIP,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("dns config not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query dns config: %w", err)
	}

	return &config, nil
}

func (d *DatabaseService) CreateDNSConfig(config *models.DNSConfig) error {
	query := `
		INSERT INTO dns_configs (provider, domain, username, password, is_active)
		VALUES (?, ?, ?, ?, ?)`

	result, err := d.db.Exec(query, config.Provider, config.Domain, config.Username, config.Password, config.IsActive)
	if err != nil {
		return fmt.Errorf("failed to insert dns config: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	config.ID = int(id)
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	return nil
}

func (d *DatabaseService) UpdateDNSConfig(config *models.DNSConfig) error {
	query := `
		UPDATE dns_configs 
		SET provider = ?, domain = ?, username = ?, password = ?, is_active = ?, last_update = ?, last_ip = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	result, err := d.db.Exec(query, config.Provider, config.Domain, config.Username, config.Password, config.IsActive, config.LastUpdate, config.LastIP, config.ID)
	if err != nil {
		return fmt.Errorf("failed to update dns config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("dns config not found")
	}

	config.UpdatedAt = time.Now()
	return nil
}

func (d *DatabaseService) DeleteDNSConfig(id int) error {
	query := `DELETE FROM dns_configs WHERE id = ?`

	result, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete dns config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("dns config not found")
	}

	return nil
}

// DNS Record methods
func (d *DatabaseService) GetDNSRecords(configID int) ([]models.DNSRecord, error) {
	query := `
		SELECT id, config_id, host, current_ip, last_update, is_active, created_at, updated_at
		FROM dns_records
		WHERE config_id = ?
		ORDER BY created_at DESC`

	rows, err := d.db.Query(query, configID)
	if err != nil {
		return nil, fmt.Errorf("failed to query dns records: %w", err)
	}
	defer rows.Close()

	var records []models.DNSRecord
	for rows.Next() {
		var record models.DNSRecord
		err := rows.Scan(
			&record.ID,
			&record.ConfigID,
			&record.Host,
			&record.CurrentIP,
			&record.LastUpdate,
			&record.IsActive,
			&record.CreatedAt,
			&record.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan dns record: %w", err)
		}
		records = append(records, record)
	}

	return records, nil
}

func (d *DatabaseService) GetDNSRecord(id int) (*models.DNSRecord, error) {
	query := `
		SELECT id, config_id, host, current_ip, last_update, is_active, created_at, updated_at
		FROM dns_records
		WHERE id = ?`

	var record models.DNSRecord
	err := d.db.QueryRow(query, id).Scan(
		&record.ID,
		&record.ConfigID,
		&record.Host,
		&record.CurrentIP,
		&record.LastUpdate,
		&record.IsActive,
		&record.CreatedAt,
		&record.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("dns record not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query dns record: %w", err)
	}

	return &record, nil
}

func (d *DatabaseService) CreateDNSRecord(record *models.DNSRecord) error {
	query := `
		INSERT INTO dns_records (config_id, host, current_ip, is_active)
		VALUES (?, ?, ?, ?)`

	result, err := d.db.Exec(query, record.ConfigID, record.Host, record.CurrentIP, record.IsActive)
	if err != nil {
		return fmt.Errorf("failed to insert dns record: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	record.ID = int(id)
	record.CreatedAt = time.Now()
	record.UpdatedAt = time.Now()

	return nil
}

func (d *DatabaseService) UpdateDNSRecord(record *models.DNSRecord) error {
	query := `
		UPDATE dns_records 
		SET host = ?, current_ip = ?, last_update = ?, is_active = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	result, err := d.db.Exec(query, record.Host, record.CurrentIP, record.LastUpdate, record.IsActive, record.ID)
	if err != nil {
		return fmt.Errorf("failed to update dns record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("dns record not found")
	}

	record.UpdatedAt = time.Now()
	return nil
}

func (d *DatabaseService) DeleteDNSRecord(id int) error {
	query := `DELETE FROM dns_records WHERE id = ?`

	result, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete dns record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("dns record not found")
	}

	return nil
}

func (d *DatabaseService) Close() error {
	return d.db.Close()
}