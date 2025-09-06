package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"upm-backend/internal/config"
	"upm-backend/internal/models"
	"upm-backend/internal/utils"

	_ "modernc.org/sqlite"
)

type DatabaseService struct {
	db            *sql.DB
	encryptionSvc *utils.EncryptionService
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

	// Initialize encryption service
	encryptionSvc, err := utils.NewEncryptionService(cfg.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize encryption service: %w", err)
	}

	service := &DatabaseService{
		db:            db,
		encryptionSvc: encryptionSvc,
	}

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

	// Create ui_settings table
	uiSettingsTable := `
	CREATE TABLE IF NOT EXISTS ui_settings (
		id INTEGER PRIMARY KEY DEFAULT 1,
		display_name TEXT NOT NULL DEFAULT 'UPM Admin',
		theme TEXT NOT NULL DEFAULT 'auto',
		language TEXT NOT NULL DEFAULT 'en',
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if _, err := d.db.Exec(uiSettingsTable); err != nil {
		return fmt.Errorf("failed to create ui_settings table: %w", err)
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
		var lastUpdate sql.NullTime
		var lastIP sql.NullString
		var encryptedPassword string

		err := rows.Scan(
			&config.ID,
			&config.Provider,
			&config.Domain,
			&config.Username,
			&encryptedPassword,
			&config.IsActive,
			&lastUpdate,
			&lastIP,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan dns config: %w", err)
		}

		// Decrypt the password
		decryptedPassword, err := d.encryptionSvc.Decrypt(encryptedPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt password for config %d: %w", config.ID, err)
		}
		config.Password = decryptedPassword

		// Handle nullable fields
		if lastUpdate.Valid {
			config.LastUpdate = &lastUpdate.Time
		}
		if lastIP.Valid {
			config.LastIP = lastIP.String
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
	var lastUpdate sql.NullTime
	var lastIP sql.NullString
	var encryptedPassword string

	err := d.db.QueryRow(query, id).Scan(
		&config.ID,
		&config.Provider,
		&config.Domain,
		&config.Username,
		&encryptedPassword,
		&config.IsActive,
		&lastUpdate,
		&lastIP,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("dns config not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query dns config: %w", err)
	}

	// Decrypt the password
	decryptedPassword, err := d.encryptionSvc.Decrypt(encryptedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt password: %w", err)
	}
	config.Password = decryptedPassword

	// Handle nullable fields
	if lastUpdate.Valid {
		config.LastUpdate = &lastUpdate.Time
	}
	if lastIP.Valid {
		config.LastIP = lastIP.String
	}

	return &config, nil
}

func (d *DatabaseService) CreateDNSConfig(config *models.DNSConfig) error {
	// Encrypt the password before storing
	encryptedPassword, err := d.encryptionSvc.Encrypt(config.Password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	query := `
		INSERT INTO dns_configs (provider, domain, username, password, is_active)
		VALUES (?, ?, ?, ?, ?)`

	result, err := d.db.Exec(query, config.Provider, config.Domain, config.Username, encryptedPassword, config.IsActive)
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
	// Encrypt the password before storing
	encryptedPassword, err := d.encryptionSvc.Encrypt(config.Password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	query := `
		UPDATE dns_configs
		SET provider = ?, domain = ?, username = ?, password = ?, is_active = ?, last_update = ?, last_ip = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	result, err := d.db.Exec(query, config.Provider, config.Domain, config.Username, encryptedPassword, config.IsActive, config.LastUpdate, config.LastIP, config.ID)
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
		var currentIP sql.NullString
		var lastUpdate sql.NullTime

		err := rows.Scan(
			&record.ID,
			&record.ConfigID,
			&record.Host,
			&currentIP,
			&lastUpdate,
			&record.IsActive,
			&record.CreatedAt,
			&record.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan dns record: %w", err)
		}

		// Handle nullable fields
		if currentIP.Valid {
			record.CurrentIP = currentIP.String
		}
		if lastUpdate.Valid {
			record.LastUpdate = &lastUpdate.Time
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
	var currentIP sql.NullString
	var lastUpdate sql.NullTime

	err := d.db.QueryRow(query, id).Scan(
		&record.ID,
		&record.ConfigID,
		&record.Host,
		&currentIP,
		&lastUpdate,
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

	// Handle nullable fields
	if currentIP.Valid {
		record.CurrentIP = currentIP.String
	}
	if lastUpdate.Valid {
		record.LastUpdate = &lastUpdate.Time
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

// Certificate methods
func (d *DatabaseService) GetCertificates() ([]models.Certificate, error) {
	query := `
		SELECT id, domain, cert_path, key_path, expires_at, is_valid, created_at, updated_at
		FROM certificates
		ORDER BY created_at DESC`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query certificates: %w", err)
	}
	defer rows.Close()

	var certificates []models.Certificate
	for rows.Next() {
		var cert models.Certificate
		err := rows.Scan(
			&cert.ID,
			&cert.Domain,
			&cert.CertPath,
			&cert.KeyPath,
			&cert.ExpiresAt,
			&cert.IsValid,
			&cert.CreatedAt,
			&cert.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan certificate: %w", err)
		}
		certificates = append(certificates, cert)
	}

	return certificates, nil
}

func (d *DatabaseService) GetCertificate(id int) (*models.Certificate, error) {
	query := `
		SELECT id, domain, cert_path, key_path, expires_at, is_valid, created_at, updated_at
		FROM certificates
		WHERE id = ?`

	var cert models.Certificate
	err := d.db.QueryRow(query, id).Scan(
		&cert.ID,
		&cert.Domain,
		&cert.CertPath,
		&cert.KeyPath,
		&cert.ExpiresAt,
		&cert.IsValid,
		&cert.CreatedAt,
		&cert.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate: %w", err)
	}

	return &cert, nil
}

func (d *DatabaseService) CreateCertificate(cert *models.Certificate) error {
	query := `
		INSERT INTO certificates (domain, cert_path, key_path, expires_at, is_valid)
		VALUES (?, ?, ?, ?, ?)`

	result, err := d.db.Exec(query, cert.Domain, cert.CertPath, cert.KeyPath, cert.ExpiresAt, cert.IsValid)
	if err != nil {
		return fmt.Errorf("failed to insert certificate: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	cert.ID = int(id)
	cert.CreatedAt = time.Now()
	cert.UpdatedAt = time.Now()

	return nil
}

func (d *DatabaseService) UpdateCertificate(cert *models.Certificate) error {
	query := `
		UPDATE certificates
		SET domain = ?, cert_path = ?, key_path = ?, expires_at = ?, is_valid = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	result, err := d.db.Exec(query, cert.Domain, cert.CertPath, cert.KeyPath, cert.ExpiresAt, cert.IsValid, cert.ID)
	if err != nil {
		return fmt.Errorf("failed to update certificate: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("certificate not found")
	}

	cert.UpdatedAt = time.Now()
	return nil
}

func (d *DatabaseService) DeleteCertificate(id int) error {
	query := `DELETE FROM certificates WHERE id = ?`

	result, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete certificate: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("certificate not found")
	}

	return nil
}

func (d *DatabaseService) GetCertificateByDomain(domain string) (*models.Certificate, error) {
	query := `
		SELECT id, domain, cert_path, key_path, expires_at, is_valid, created_at, updated_at
		FROM certificates
		WHERE domain = ?`

	var cert models.Certificate
	err := d.db.QueryRow(query, domain).Scan(
		&cert.ID,
		&cert.Domain,
		&cert.CertPath,
		&cert.KeyPath,
		&cert.ExpiresAt,
		&cert.IsValid,
		&cert.CreatedAt,
		&cert.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("certificate not found for domain: %s", domain)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query certificate: %w", err)
	}

	return &cert, nil
}

func (d *DatabaseService) GetProxiesByDomain(domain string) ([]models.Proxy, error) {
	query := `
		SELECT id, name, domain, target_url, ssl_enabled, ssl_path, status, created_at, updated_at
		FROM proxies
		WHERE domain = ? OR domain LIKE ?`

	rows, err := d.db.Query(query, domain, "%."+domain)
	if err != nil {
		return nil, fmt.Errorf("failed to query proxies by domain: %w", err)
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

// UI Settings methods
func (d *DatabaseService) GetUISettings() (models.UISettings, error) {
	var settings models.UISettings
	query := `SELECT display_name, theme, language
			  FROM ui_settings WHERE id = 1`

	err := d.db.QueryRow(query).Scan(
		&settings.DisplayName,
		&settings.Theme,
		&settings.Language,
	)

	return settings, err
}

func (d *DatabaseService) SaveUISettings(settings models.UISettings) error {
	query := `INSERT OR REPLACE INTO ui_settings
			  (id, display_name, theme, language, updated_at)
			  VALUES (1, ?, ?, ?, ?)`

	_, err := d.db.Exec(query,
		settings.DisplayName,
		settings.Theme,
		settings.Language,
		time.Now(),
	)

	return err
}

// Admin user management methods
func (d *DatabaseService) AdminUserExists() (bool, error) {
	query := `SELECT COUNT(*) FROM users WHERE username = 'admin'`
	var count int
	err := d.db.QueryRow(query).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check admin user existence: %w", err)
	}
	return count > 0, nil
}

func (d *DatabaseService) CreateAdminUser(hashedPassword string) error {
	query := `INSERT INTO users (username, email, password, is_active, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?)`
	
	now := time.Now()
	_, err := d.db.Exec(query,
		"admin",
		"admin@upm.local",
		hashedPassword,
		true,
		now,
		now,
	)
	
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}
	
	log.Println("Admin user created successfully")
	return nil
}

func (d *DatabaseService) GetAdminUser() (*models.User, error) {
	query := `SELECT id, username, email, password, is_active, created_at, updated_at
			  FROM users WHERE username = 'admin'`
	
	var user models.User
	err := d.db.QueryRow(query).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get admin user: %w", err)
	}
	
	return &user, nil
}

func (d *DatabaseService) UpdateAdminUserPassword(hashedPassword string) error {
	query := `UPDATE users SET password = ?, updated_at = ? WHERE username = 'admin'`
	
	_, err := d.db.Exec(query, hashedPassword, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update admin user password: %w", err)
	}
	
	log.Println("Admin user password updated successfully")
	return nil
}

func (d *DatabaseService) DeleteAdminUser() error {
	query := `DELETE FROM users WHERE username = 'admin'`
	
	_, err := d.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete admin user: %w", err)
	}
	
	log.Println("Admin user deleted successfully")
	return nil
}

func (d *DatabaseService) Close() error {
	return d.db.Close()
}
