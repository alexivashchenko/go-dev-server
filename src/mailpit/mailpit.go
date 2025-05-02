package mailpit

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/alexivashchenko/go-dev-server/helpers"
)

// Configuration holds all Mailpit-related settings
type Configuration struct {
	RootDir        string
	AppFolder      string
	ExecutableName string
	AppPath        string
	SMTPHost       string
	SMTPPort       string
	UIHost         string
	UIPort         string
}

// NewConfiguration creates a new Mailpit configuration
func NewConfiguration() (*Configuration, error) {
	rootDir := helpers.GetRootDirectory()

	// Get Mailpit app folder from environment
	mailpitAppFolder := os.Getenv("MAILPIT_APP_FOLDER")
	if mailpitAppFolder == "" {
		return nil, fmt.Errorf("mailpit_app_folder environment variable is not set")
	}

	// Get SMTP settings from environment
	smtpHost := os.Getenv("MAILPIT_SMTP_HOST")
	if smtpHost == "" {
		smtpHost = "127.0.0.1" // Default SMTP host
		log.Printf("MAILPIT_SMTP_HOST not set, using default: %s", smtpHost)
	}

	smtpPort := os.Getenv("MAILPIT_SMTP_PORT")
	if smtpPort == "" {
		smtpPort = "1025" // Default SMTP port
		log.Printf("MAILPIT_SMTP_PORT not set, using default: %s", smtpPort)
	}

	// Get UI settings from environment
	uiHost := os.Getenv("MAILPIT_UI_HOST")
	if uiHost == "" {
		uiHost = "127.0.0.1" // Default UI host
		log.Printf("MAILPIT_UI_HOST not set, using default: %s", uiHost)
	}

	uiPort := os.Getenv("MAILPIT_UI_PORT")
	if uiPort == "" {
		uiPort = "8025" // Default UI port
		log.Printf("MAILPIT_UI_PORT not set, using default: %s", uiPort)
	}

	// Determine executable name based on OS
	executableName := "mailpit"
	if runtime.GOOS == "windows" {
		executableName = "mailpit.exe"
	}

	// Create configuration
	config := &Configuration{
		RootDir:        rootDir,
		AppFolder:      mailpitAppFolder,
		ExecutableName: executableName,
		SMTPHost:       smtpHost,
		SMTPPort:       smtpPort,
		UIHost:         uiHost,
		UIPort:         uiPort,
	}

	// Set paths
	config.AppPath = filepath.Join(rootDir, "apps", "mailpit", mailpitAppFolder)

	return config, nil
}

// Start initializes and starts the Mailpit service
func Start() error {
	log.Println("Starting Mailpit service...")
	startTime := time.Now()

	// Initialize configuration
	config, err := NewConfiguration()
	if err != nil {
		return fmt.Errorf("failed to initialize mailpit configuration: %w", err)
	}

	// Start Mailpit
	if err := startMailpit(config); err != nil {
		return fmt.Errorf("failed to start mailpit: %w", err)
	}

	elapsed := time.Since(startTime)
	log.Printf("Mailpit service started successfully in %.2f seconds", elapsed.Seconds())
	return nil
}

// Stop stops the Mailpit service
func Stop() error {
	log.Println("Stopping Mailpit service...")
	startTime := time.Now()

	config, err := NewConfiguration()
	if err != nil {
		return fmt.Errorf("failed to initialize mailpit configuration: %w", err)
	}

	// Kill Mailpit process
	if err := helpers.KillProcess(config.ExecutableName); err != nil {
		return fmt.Errorf("failed to stop mailpit process: %w", err)
	}

	elapsed := time.Since(startTime)
	log.Printf("Mailpit service stopped successfully in %.2f seconds", elapsed.Seconds())
	return nil
}

// Restart restarts the Mailpit service
func Restart() error {
	log.Println("Restarting Mailpit service...")

	if err := Stop(); err != nil {
		log.Printf("Warning: Error stopping Mailpit service: %v", err)
		// Continue with start even if stop failed
	}

	// Small delay to ensure process has fully terminated
	time.Sleep(500 * time.Millisecond)

	if err := Start(); err != nil {
		return fmt.Errorf("failed to restart mailpit service: %w", err)
	}

	return nil
}

// GetStatus returns the current status of the Mailpit service
func GetStatus() string {
	config, err := NewConfiguration()
	if err != nil {
		return "Error: " + err.Error()
	}

	running, pid := helpers.IsProcessRunning(config.ExecutableName)
	if running {
		return fmt.Sprintf("Running (PID: %d, SMTP: %s:%s, UI: http://%s:%s)",
			pid, config.SMTPHost, config.SMTPPort, config.UIHost, config.UIPort)
	}
	return "Stopped"
}

// startMailpit starts the Mailpit server
func startMailpit(config *Configuration) error {
	log.Println("Starting Mailpit server...")

	// Build command with all necessary parameters
	command := fmt.Sprintf("%s --smtp=%s:%s --listen=%s:%s",
		filepath.Join(config.AppPath, config.ExecutableName),
		config.SMTPHost,
		config.SMTPPort,
		config.UIHost,
		config.UIPort,
	)

	log.Printf("Running command: %s", command)

	if err := helpers.RunCommand(command, true); err != nil {
		return fmt.Errorf("failed to start mailpit: %w", err)
	}

	// Verify process is running
	time.Sleep(500 * time.Millisecond) // Give process time to start
	running, _ := helpers.IsProcessRunning(config.ExecutableName)
	if !running {
		return fmt.Errorf("mailpit process failed to start")
	}

	log.Printf("Mailpit started with SMTP on %s:%s and UI on https://%s:%s",
		config.SMTPHost, config.SMTPPort, config.UIHost, config.UIPort)

	return nil
}
