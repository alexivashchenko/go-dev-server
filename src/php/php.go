package php

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

// Configuration holds all PHP-related settings
type Configuration struct {
	RootDir         string
	AppFolder       string
	ErrorLog        string
	IncludePath     string
	ExtensionDir    string
	SessionSavePath string
	CurlCaInfo      string
	SendmailPath    string
	TemplatesDir    string
	AppDir          string
	IniTemplateFile string
	IniFile         string
	Host            string
	Port            int
	ProcessName     string
}

// NewConfiguration creates a new PHP configuration
func NewConfiguration() (*Configuration, error) {
	rootDir := helpers.GetRootDirectory()

	// Get PHP app folder from environment
	phpAppFolder := os.Getenv("PHP_APP_FOLDER")
	if phpAppFolder == "" {
		return nil, fmt.Errorf("PHP_APP_FOLDER environment variable is not set")
	}

	// Determine process name based on OS
	processName := "php-cgi"
	if runtime.GOOS == "windows" {
		processName = "php-cgi.exe"
	}

	// Create configuration
	config := &Configuration{
		RootDir:      rootDir,
		AppFolder:    phpAppFolder,
		Host:         "127.0.0.1",
		Port:         9003,
		ProcessName:  processName,
		TemplatesDir: filepath.Join(rootDir, "tpl"),
	}

	// Set paths
	config.AppDir = filepath.Join(rootDir, "apps", "php", phpAppFolder)
	config.IniTemplateFile = filepath.Join(config.TemplatesDir, "php", "php.ini.tpl")
	config.IniFile = filepath.Join(config.AppDir, "php.ini")

	// Process environment variables with placeholders
	envVars := map[string]*string{
		"PHP_ERROR_LOG":         &config.ErrorLog,
		"PHP_INCLUDE_PATH":      &config.IncludePath,
		"PHP_EXTENSION_DIR":     &config.ExtensionDir,
		"PHP_SESSION_SAVE_PATH": &config.SessionSavePath,
		"PHP_CURL_CAINFO":       &config.CurlCaInfo,
		"PHP_SENDMAIL_PATH":     &config.SendmailPath,
	}

	for envName, configVar := range envVars {
		value := os.Getenv(envName)
		if value == "" {
			log.Printf("Warning: %s environment variable is not set", envName)
			continue
		}

		// Replace placeholders
		value = strings.ReplaceAll(value, "{ROOT_DIR}", rootDir)
		value = strings.ReplaceAll(value, "{PHP_APP_FOLDER}", phpAppFolder)
		*configVar = value
	}

	return config, nil
}

// Start initializes and starts the PHP service
func Start() error {
	log.Println("Starting PHP service...")
	startTime := time.Now()

	// Initialize configuration
	config, err := NewConfiguration()
	if err != nil {
		return fmt.Errorf("failed to initialize PHP configuration: %w", err)
	}

	// Ensure required directories exist
	if err := ensureDirectoriesExist(config); err != nil {
		return fmt.Errorf("failed to create required directories: %w", err)
	}

	// Create PHP configuration file
	if err := createPHPConfig(config); err != nil {
		return fmt.Errorf("failed to create PHP configuration: %w", err)
	}

	// Start PHP-CGI process
	if err := startPHPProcess(config); err != nil {
		return fmt.Errorf("failed to start PHP process: %w", err)
	}

	elapsed := time.Since(startTime)
	log.Printf("PHP service started successfully in %.2f seconds", elapsed.Seconds())
	return nil
}

// Stop stops the PHP service
func Stop() error {
	log.Println("Stopping PHP service...")
	startTime := time.Now()

	config, err := NewConfiguration()
	if err != nil {
		return fmt.Errorf("failed to initialize PHP configuration: %w", err)
	}

	// Kill PHP-CGI process
	if err := helpers.KillProcess(config.ProcessName); err != nil {
		return fmt.Errorf("failed to stop PHP process: %w", err)
	}

	elapsed := time.Since(startTime)
	log.Printf("PHP service stopped successfully in %.2f seconds", elapsed.Seconds())
	return nil
}

// Restart restarts the PHP service
func Restart() error {
	log.Println("Restarting PHP service...")

	if err := Stop(); err != nil {
		log.Printf("Warning: Error stopping PHP service: %v", err)
		// Continue with start even if stop failed
	}

	// Small delay to ensure process has fully terminated
	time.Sleep(500 * time.Millisecond)

	if err := Start(); err != nil {
		return fmt.Errorf("failed to restart PHP service: %w", err)
	}

	return nil
}

// GetStatus returns the current status of the PHP service
func GetStatus() string {
	config, err := NewConfiguration()
	if err != nil {
		return "Error: " + err.Error()
	}

	running, pid := helpers.IsProcessRunning(config.ProcessName)
	if running {
		return fmt.Sprintf("Running (PID: %d, Port: %d)", pid, config.Port)
	}
	return "Stopped"
}

// ensureDirectoriesExist creates necessary directories
func ensureDirectoriesExist(config *Configuration) error {
	// Create logs directory
	if config.ErrorLog != "" {
		logDir := filepath.Dir(config.ErrorLog)
		// Check if directory exists first
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			log.Printf("Creating log directory: %s", logDir)
			if err := os.MkdirAll(logDir, 0755); err != nil {
				return fmt.Errorf("failed to create log directory %s: %w", logDir, err)
			}
		} else if err != nil {
			// Some other error occurred when checking the directory
			return fmt.Errorf("failed to check log directory %s: %w", logDir, err)
		}
	}

	// Create session directory
	if config.SessionSavePath != "" {
		// Check if directory exists first
		if _, err := os.Stat(config.SessionSavePath); os.IsNotExist(err) {
			log.Printf("Creating session directory: %s", config.SessionSavePath)
			if err := os.MkdirAll(config.SessionSavePath, 0755); err != nil {
				return fmt.Errorf("failed to create session directory %s: %w", config.SessionSavePath, err)
			}
		} else if err != nil {
			// Some other error occurred when checking the directory
			return fmt.Errorf("failed to check session directory %s: %w", config.SessionSavePath, err)
		}
	}

	return nil
}

// createPHPConfig creates the PHP configuration file
func createPHPConfig(config *Configuration) error {
	// Check if template exists
	if _, err := os.Stat(config.IniTemplateFile); os.IsNotExist(err) {
		return fmt.Errorf("PHP ini template file not found: %s", config.IniTemplateFile)
	}

	// Copy template to destination
	if err := helpers.CopyFile(config.IniTemplateFile, config.IniFile); err != nil {
		return fmt.Errorf("failed to copy PHP ini template: %w", err)
	}

	// Replace placeholders in the ini file
	replaceMap := map[string]string{
		"{PHP_APP_FOLDER}":        config.AppFolder,
		"{PHP_ERROR_LOG}":         config.ErrorLog,
		"{PHP_INCLUDE_PATH}":      config.IncludePath,
		"{PHP_EXTENSION_DIR}":     config.ExtensionDir,
		"{PHP_SESSION_SAVE_PATH}": config.SessionSavePath,
		"{PHP_CURL_CAINFO}":       config.CurlCaInfo,
		"{PHP_SENDMAIL_PATH}":     config.SendmailPath,
	}

	if err := helpers.ReplaceInFileByMap(config.IniFile, replaceMap); err != nil {
		return fmt.Errorf("failed to update PHP ini file: %w", err)
	}

	return nil
}

// startPHPProcess starts the PHP-CGI process
func startPHPProcess(config *Configuration) error {
	// Prepare command
	command := fmt.Sprintf("%s -b %s:%d", config.ProcessName, config.Host, config.Port)

	// Set working directory to PHP app directory for proper DLL loading
	workingDir := config.AppDir

	// Start PHP-CGI process
	if err := helpers.RunCommandInDirectory(command, workingDir, true); err != nil {
		return fmt.Errorf("failed to start PHP-CGI: %w", err)
	}

	// Verify process is running
	time.Sleep(500 * time.Millisecond) // Give process time to start
	running, _ := helpers.IsProcessRunning(config.ProcessName)
	if !running {
		return fmt.Errorf("PHP-CGI process failed to start")
	}

	return nil
}
