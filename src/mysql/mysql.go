package mysql

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/alexivashchenko/go-dev-server/helpers"
)

// Configuration holds all MySQL-related settings
type Configuration struct {
	RootDir            string
	AppFolder          string
	DataFolder         string
	AppDir             string
	EtcDir             string
	DataDir            string
	TemplatesDir       string
	ConfigFile         string
	InitFile         string
	ConfigTemplateFile string
	InitTemplateFile string
	ExecutableName     string
	Port               int
	User               string
	Password           string
}

// NewConfiguration creates a new MySQL configuration
func NewConfiguration() (*Configuration, error) {
	rootDir := helpers.GetRootDirectory()

	// Get MySQL app folder from environment
	mysqlAppFolder := os.Getenv("MYSQL_APP_FOLDER")
	if mysqlAppFolder == "" {
		return nil, fmt.Errorf("MYSQL_APP_FOLDER environment variable is not set")
	}

	// Get MySQL data folder from environment
	mysqlDataFolder := os.Getenv("MYSQL_DATA_FOLDER")
	if mysqlDataFolder == "" {
		return nil, fmt.Errorf("MYSQL_DATA_FOLDER environment variable is not set")
	}

	// Get MySQL data folder from environment
	mysqlRootPassword := os.Getenv("MYSQL_ROOT_PASSWORD")

	// Determine executable name based on OS
	executableName := "mysqld"
	if runtime.GOOS == "windows" {
		executableName = "mysqld.exe"
	}

	// Create configuration
	config := &Configuration{
		RootDir:        rootDir,
		AppFolder:      mysqlAppFolder,
		DataFolder:     mysqlDataFolder,
		ExecutableName: executableName,
		Port:           3306,
		User:           "root",
		Password:       mysqlRootPassword,
	}

	// Set paths
	config.AppDir = filepath.Join(rootDir, "apps", "mysql", mysqlAppFolder)
	config.EtcDir = filepath.Join(rootDir, "etc", "mysql")
	config.DataDir = filepath.Join(rootDir, "data", mysqlDataFolder)
	config.TemplatesDir = filepath.Join(rootDir, "tpl")
	config.ConfigFile = filepath.Join(config.AppDir, "my.ini")
	config.InitFile = filepath.Join(config.EtcDir, "mysql-init.txt")
	config.ConfigTemplateFile = filepath.Join(config.TemplatesDir, "mysql", "my.ini.tpl")
	config.InitTemplateFile = filepath.Join(config.TemplatesDir, "mysql", "mysql-init.txt.tpl")

	// Validate template file exists
	if _, err := os.Stat(config.ConfigTemplateFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("MySQL configuration template not found: %s", config.ConfigTemplateFile)
	}

	return config, nil
}

// Start initializes and starts the MySQL service
func Start() error {
	log.Println("Starting MySQL service...")
	startTime := time.Now()

	// Initialize configuration
	config, err := NewConfiguration()
	if err != nil {
		return fmt.Errorf("failed to initialize MySQL configuration: %w", err)
	}

	// Check if MySQL needs initialization
	needsInit, err := checkIfNeedsInitialization(config)
	if err != nil {
		return fmt.Errorf("failed to check if MySQL needs initialization: %w", err)
	}

	// Initialize MySQL if needed
	if needsInit {
		log.Println("MySQL data directory not found, initializing MySQL...")
		if err := initializeMySQL(config); err != nil {
			return fmt.Errorf("failed to initialize MySQL: %w", err)
		}
		log.Println("MySQL initialization completed successfully")
	} else {
		log.Println("MySQL data directory exists, skipping initialization")
	}

	// Start MySQL server
	if err := startMySQLServer(config); err != nil {
		return fmt.Errorf("failed to start MySQL server: %w", err)
	}

	// Verify MySQL is running
	if err := verifyMySQLRunning(config); err != nil {
		return fmt.Errorf("MySQL verification failed: %w", err)
	}

	elapsed := time.Since(startTime)
	log.Printf("MySQL service started successfully in %.2f seconds", elapsed.Seconds())
	return nil
}

// Stop stops the MySQL service
func Stop() error {
	log.Println("Stopping MySQL service...")
	startTime := time.Now()

	config, err := NewConfiguration()
	if err != nil {
		return fmt.Errorf("failed to initialize MySQL configuration: %w", err)
	}

	// Try graceful shutdown first
	if err := gracefulShutdown(config); err != nil {
		log.Printf("Warning: Graceful shutdown failed: %v", err)
		log.Println("Attempting to kill MySQL process...")

		// Fall back to killing the process
		if err := helpers.KillProcess(config.ExecutableName); err != nil {
			return fmt.Errorf("failed to stop MySQL process: %w", err)
		}
	}

	elapsed := time.Since(startTime)
	log.Printf("MySQL service stopped successfully in %.2f seconds", elapsed.Seconds())
	return nil
}

// Restart restarts the MySQL service
func Restart() error {
	log.Println("Restarting MySQL service...")

	if err := Stop(); err != nil {
		log.Printf("Warning: Error stopping MySQL service: %v", err)
		// Continue with start even if stop failed
	}

	// Small delay to ensure process has fully terminated
	time.Sleep(1000 * time.Millisecond)

	if err := Start(); err != nil {
		return fmt.Errorf("failed to restart MySQL service: %w", err)
	}

	return nil
}

// GetStatus returns the current status of the MySQL service
func GetStatus() string {
	config, err := NewConfiguration()
	if err != nil {
		return "Error: " + err.Error()
	}
	running, pid := helpers.IsProcessRunning(config.ExecutableName)
	if running {
		return fmt.Sprintf("Running (PID: %d, Port: %d)", pid, config.Port)
	}
	return "Stopped"
}

// checkIfNeedsInitialization checks if MySQL needs to be initialized
func checkIfNeedsInitialization(config *Configuration) (bool, error) {
	// Check if data directory exists
	if _, err := os.Stat(config.DataDir); os.IsNotExist(err) {
		return true, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to check data directory: %w", err)
	}

	// Check if data directory is empty
	entries, err := os.ReadDir(config.DataDir)
	if err != nil {
		return false, fmt.Errorf("failed to read data directory: %w", err)
	}

	return len(entries) == 0, nil
}

// initializeMySQL initializes a new MySQL instance
func initializeMySQL(config *Configuration) error {
	// Create data directory
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Copy and update configuration template
	if err := copyAndUpdateMySQLConfig(config); err != nil {
		return fmt.Errorf("failed to copy and update MySQL configuration: %w", err)
	}

	// Run MySQL initialization
	log.Println("Running MySQL initialization...")

	mysqldPath := filepath.Join(config.AppDir, "bin", config.ExecutableName)
	command := fmt.Sprintf("%s --defaults-file=%s --initialize", mysqldPath, config.ConfigFile)

	output, err := helpers.RunCommandWithOutput(command)
	if err != nil {
		return fmt.Errorf("MySQL initialization failed: %w\nOutput: %s", err, output)
	}

	// Extract temporary root password from output
	tempPassword := extractTempPassword(output)
	if tempPassword != "" {
		log.Printf("Temporary root password: %s", tempPassword)
		// Store the password in the configuration
		config.Password = tempPassword
	}

	return nil
}

// copyAndUpdateMySQLConfig copies and updates the MySQL configuration file
func copyAndUpdateMySQLConfig(config *Configuration) error {
	// Copy configuration template
	if err := helpers.CopyFile(config.ConfigTemplateFile, config.ConfigFile); err != nil {
		return fmt.Errorf("failed to copy MySQL configuration template: %w", err)
	}

	// Replace placeholders in the configuration file
	rootDirFormatted := helpers.ReplaceBackslashToSlash(config.RootDir + string(os.PathSeparator))
	replacements := map[string]string{
		"{mysql_data_folder}": config.DataFolder,
		"{root_folder}": rootDirFormatted,
	}

	for placeholder, value := range replacements {
		if err := helpers.ReplaceInFile(config.ConfigFile, placeholder, value); err != nil {
			return fmt.Errorf("failed to update MySQL configuration file: %w", err)
		}
	}

	return nil
}

// copyAndUpdateMySQLInitFile copies and updates the MySQL init file
func copyAndUpdateMySQLInitFile(config *Configuration) error {
	// Copy configuration template
	if err := helpers.CopyFile(config.InitTemplateFile, config.InitFile); err != nil {
		return fmt.Errorf("failed to copy MySQL init template: %w", err)
	}

	if err := helpers.ReplaceInFile(config.InitFile, "{mysql_root_password}", config.Password); err != nil {
		return fmt.Errorf("failed to update MySQL init file: %w", err)
	}

	return nil
}


// startMySQLServer starts the MySQL server
func startMySQLServer(config *Configuration) error {
	log.Println("Starting MySQL server...")

	// Copy and update init template
	if err := copyAndUpdateMySQLInitFile(config); err != nil {
		return fmt.Errorf("failed to copy and update MySQL init: %w", err)
	}

	mysqldPath := filepath.Join(config.AppDir, "bin", config.ExecutableName)
	command := fmt.Sprintf("%s --defaults-file=%s --init-file=%s", mysqldPath, config.ConfigFile, config.InitFile)

	if err := helpers.RunCommand(command, true); err != nil {
		return fmt.Errorf("failed to start MySQL server: %w", err)
	}

	return nil
}

// verifyMySQLRunning verifies that MySQL is running
func verifyMySQLRunning(config *Configuration) error {
	// Wait for MySQL to start
	maxRetries := 10
	retryInterval := 500 * time.Millisecond

	log.Println("Verifying MySQL is running...")

	for i := 0; i < maxRetries; i++ {
		running, pid := helpers.IsProcessRunning(config.ExecutableName)
		if running {
			log.Printf("MySQL is running with PID %d", pid)
			return nil
		}

		log.Printf("Waiting for MySQL to start (attempt %d/%d)...", i+1, maxRetries)
		time.Sleep(retryInterval)
	}

	return fmt.Errorf("MySQL failed to start within the expected time")
}

// gracefulShutdown attempts to shut down MySQL gracefully
func gracefulShutdown(config *Configuration) error {
	// This is a placeholder for a more graceful shutdown
	// In a real implementation, you might use mysqladmin shutdown
	// or another method to gracefully stop MySQL

	// For now, we'll just use the kill process method
	return helpers.KillProcess(config.ExecutableName)
}

// extractTempPassword extracts the temporary root password from MySQL initialization output
func extractTempPassword(output string) string {
	// Look for the temporary password line in the output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "temporary password") {
			parts := strings.Split(line, "root@localhost: ")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}
