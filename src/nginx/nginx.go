package nginx

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

// Configuration holds all Nginx-related settings
type Configuration struct {
	RootDir             string
	AppFolder           string
	AppPath             string
	DomainTail          string
	EtcFolder           string
	SitesEnabledFolder  string
	HostsFilePath       string
	TmpHostsFilePath    string
	HostsFileIdentifier string
	WWWDir              string
	TemplatesDir        string
	LogsDir             string
	ExecutableName      string
	DefaultConfTemplate string
	NginxConfTemplate   string
	GeneralSiteTemplate string
}

// NewConfiguration creates a new Nginx configuration
func NewConfiguration() (*Configuration, error) {
	rootDir := helpers.GetRootDirectory()

	// Get Nginx app folder from environment
	nginxAppFolder := os.Getenv("NGINX_APP_FOLDER")
	if nginxAppFolder == "" {
		return nil, fmt.Errorf("nginx_app_folder environment variable is not set")
	}

	// Get Nginx domain tail from environment
	nginxDomainTail := os.Getenv("NGINX_DOMAIN_TAIL")
	if nginxDomainTail == "" {
		return nil, fmt.Errorf("nginx_domain_tail environment variable is not set")
	}

	// Determine executable name based on OS
	executableName := "nginx"
	hostsFilePath := "/etc/hosts"

	if runtime.GOOS == "windows" {
		executableName = "nginx.exe"
		hostsFilePath = filepath.Join("C:\\", "Windows", "System32", "drivers", "etc", "hosts")
	} else if runtime.GOOS == "darwin" {
		hostsFilePath = "/private/etc/hosts"
	}

	// Create configuration
	config := &Configuration{
		RootDir:             rootDir,
		AppFolder:           nginxAppFolder,
		DomainTail:          nginxDomainTail,
		HostsFilePath:       hostsFilePath,
		HostsFileIdentifier: "#local server setting",
		ExecutableName:      executableName,
	}

	// Set paths
	config.AppPath = filepath.Join(rootDir, "apps", "nginx", nginxAppFolder)
	config.EtcFolder = filepath.Join(rootDir, "etc", "nginx")
	config.SitesEnabledFolder = filepath.Join(config.EtcFolder, "sites-enabled")
	config.TmpHostsFilePath = filepath.Join(rootDir, "hosts.tmp")
	config.WWWDir = filepath.Join(rootDir, "www")
	config.TemplatesDir = filepath.Join(rootDir, "tpl")
	config.LogsDir = filepath.Join(rootDir, "logs", "nginx")

	// Template files
	config.DefaultConfTemplate = filepath.Join(config.TemplatesDir, "nginx", "00-default.conf.tpl")
	config.NginxConfTemplate = filepath.Join(config.TemplatesDir, "nginx", "nginx.conf.tpl")
	config.GeneralSiteTemplate = filepath.Join(config.TemplatesDir, "nginx", "general-site.conf.tpl")

	// Validate template files exist
	templates := []string{
		config.DefaultConfTemplate,
		config.NginxConfTemplate,
		config.GeneralSiteTemplate,
	}

	for _, template := range templates {
		if _, err := os.Stat(template); os.IsNotExist(err) {
			return nil, fmt.Errorf("template file not found: %s", template)
		}
	}

	return config, nil
}

// Start initializes and starts the Nginx service
func Start() error {
	log.Println("Starting Nginx service...")
	startTime := time.Now()

	// Initialize configuration
	config, err := NewConfiguration()
	if err != nil {
		return fmt.Errorf("failed to initialize nginx configuration: %w", err)
	}

	// Create required directories
	if err := ensureDirectoriesExist(config); err != nil {
		return fmt.Errorf("failed to create required directories: %w", err)
	}

	// Update hosts file
	if err := updateHostsFile(config); err != nil {
		return fmt.Errorf("failed to update hosts file: %w", err)
	}

	// Configure Nginx
	if err := configureNginx(config); err != nil {
		return fmt.Errorf("failed to configure nginx: %w", err)
	}

	// Create site configurations
	if err := createSiteConfigurations(config); err != nil {
		return fmt.Errorf("failed to create site configurations: %w", err)
	}

	// Create log files
	if err := createLogFiles(config); err != nil {
		return fmt.Errorf("failed to create log files: %w", err)
	}

	// Check Nginx configuration
	if err := checkNginxConfiguration(config); err != nil {
		return fmt.Errorf("nginx configuration check failed: %w", err)
	}

	// Start Nginx
	if err := startNginx(config); err != nil {
		return fmt.Errorf("failed to start nginx: %w", err)
	}

	elapsed := time.Since(startTime)
	log.Printf("Nginx service started successfully in %.2f seconds", elapsed.Seconds())
	return nil
}

// Stop stops the Nginx service
func Stop() error {
	log.Println("Stopping Nginx service...")
	startTime := time.Now()

	config, err := NewConfiguration()
	if err != nil {
		return fmt.Errorf("failed to initialize nginx configuration: %w", err)
	}

	// Kill Nginx process
	if err := helpers.KillProcess(config.ExecutableName); err != nil {
		return fmt.Errorf("failed to stop nginx process: %w", err)
	}

	elapsed := time.Since(startTime)
	log.Printf("Nginx service stopped successfully in %.2f seconds", elapsed.Seconds())
	return nil
}

// Restart restarts the Nginx service
func Restart() error {
	log.Println("Restarting Nginx service...")

	if err := Stop(); err != nil {
		log.Printf("Warning: Error stopping Nginx service: %v", err)
		// Continue with start even if stop failed
	}

	// Small delay to ensure process has fully terminated
	time.Sleep(500 * time.Millisecond)

	if err := Start(); err != nil {
		return fmt.Errorf("failed to restart nginx service: %w", err)
	}

	return nil
}

// GetStatus returns the current status of the Nginx service
func GetStatus() string {
	config, err := NewConfiguration()
	if err != nil {
		return "Error: " + err.Error()
	}

	running, pid := helpers.IsProcessRunning(config.ExecutableName)
	if running {
		return fmt.Sprintf("Running (PID: %d)", pid)
	}
	return "Stopped"
}

// ensureDirectoriesExist creates necessary directories
func ensureDirectoriesExist(config *Configuration) error {
	directories := []string{
		config.EtcFolder,
		config.SitesEnabledFolder,
		config.LogsDir,
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// updateHostsFile updates the hosts file with site entries
func updateHostsFile(config *Configuration) error {
	log.Println("Updating hosts file...")

	// Remove temporary hosts file if it exists
	if err := helpers.RemoveFile(config.TmpHostsFilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove temporary hosts file: %w", err)
	}

	// Create new temporary hosts file
	if err := helpers.CreateFile(config.TmpHostsFilePath); err != nil {
		return fmt.Errorf("failed to create temporary hosts file: %w", err)
	}

	// Read current hosts file
	lines, err := helpers.ReadLinesIntoSlice(config.HostsFilePath)
	if err != nil {
		return fmt.Errorf("failed to read hosts file: %w", err)
	}

	// Filter out existing server entries
	newLines := []string{}
	for _, line := range lines {
		if !strings.Contains(line, config.HostsFileIdentifier) {
			newLines = append(newLines, line)
		}
	}

	// Get website directories
	dirs, err := helpers.ListDirectories(config.WWWDir)
	if err != nil {
		return fmt.Errorf("failed to list website directories: %w", err)
	}

	// Add new entries for each website
	for _, dir := range dirs {
		baseName := filepath.Base(dir)
		hostEntry := fmt.Sprintf("127.0.0.1\t%s\t%s",
			baseName+"."+config.DomainTail,
			config.HostsFileIdentifier)
		newLines = append(newLines, hostEntry)
	}

	// Write to temporary file
	if err := helpers.AppendLines(config.TmpHostsFilePath, newLines); err != nil {
		return fmt.Errorf("failed to write to temporary hosts file: %w", err)
	}

	// Copy to actual hosts file with admin privileges
	if err := helpers.CopyFileAsAdmin(config.TmpHostsFilePath, config.HostsFilePath); err != nil {
		return fmt.Errorf("failed to update hosts file: %w", err)
	}

	return nil
}

// configureNginx configures the Nginx server
func configureNginx(config *Configuration) error {
	log.Println("Configuring Nginx...")

	// Copy and configure main nginx.conf
	nginxConfFile := filepath.Join(config.AppPath, "conf", "nginx.conf")

	if err := helpers.CopyFile(config.NginxConfTemplate, nginxConfFile); err != nil {
		return fmt.Errorf("failed to copy nginx configuration template: %w", err)
	}

	// Replace placeholders in nginx.conf
	rootDirFormatted := helpers.ReplaceBackslashToSlash(config.RootDir + string(os.PathSeparator))
	if err := helpers.ReplaceInFile(nginxConfFile, "{root_folder}", rootDirFormatted); err != nil {
		return fmt.Errorf("failed to update nginx configuration: %w", err)
	}

	// Clean and recreate sites-enabled directory
	if err := helpers.RemoveDirectoryAndContents(config.SitesEnabledFolder); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clean sites-enabled directory: %w", err)
	}

	if err := os.MkdirAll(config.SitesEnabledFolder, 0755); err != nil {
		return fmt.Errorf("failed to create sites-enabled directory: %w", err)
	}

	// Copy and configure default site
	defaultConfFile := filepath.Join(config.SitesEnabledFolder, "00-default.conf")

	if err := helpers.CopyFile(config.DefaultConfTemplate, defaultConfFile); err != nil {
		return fmt.Errorf("failed to copy default site template: %w", err)
	}

	if err := helpers.ReplaceInFile(defaultConfFile, "{root_folder}", rootDirFormatted); err != nil {
		return fmt.Errorf("failed to update default site configuration: %w", err)
	}

	return nil
}

// createSiteConfigurations creates configuration files for each site
func createSiteConfigurations(config *Configuration) error {
	log.Println("Creating site configurations...")

	// Get website directories
	dirs, err := helpers.ListDirectories(config.WWWDir)
	if err != nil {
		return fmt.Errorf("failed to list website directories: %w", err)
	}

	// Create configuration for each site
	for _, dir := range dirs {
		baseName := filepath.Base(dir)
		domainName := baseName + "." + config.DomainTail
		siteConfFile := filepath.Join(config.SitesEnabledFolder, domainName+".conf")

		// Copy site template
		if err := helpers.CopyFile(config.GeneralSiteTemplate, siteConfFile); err != nil {
			return fmt.Errorf("failed to copy site template for %s: %w", domainName, err)
		}

		// Replace placeholders
		replacements := map[string]string{
			"{root_folder}": helpers.ReplaceBackslashToSlash(config.RootDir + string(os.PathSeparator)),
			"{folder_name}": baseName,
			"{domain_name}": domainName,
		}

		for placeholder, value := range replacements {
			if err := helpers.ReplaceInFile(siteConfFile, placeholder, value); err != nil {
				return fmt.Errorf("failed to update site configuration for %s: %w", domainName, err)
			}
		}
	}

	return nil
}

// createLogFiles creates log files for Nginx and each site
func createLogFiles(config *Configuration) error {
	log.Println("Creating log files...")

	// Clean and recreate logs directory
	if err := helpers.RemoveDirectoryAndContents(config.LogsDir); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Failed to clean logs directory: %v", err)
	}

	if err := os.MkdirAll(config.LogsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Create main error log
	mainErrorLog := filepath.Join(config.LogsDir, "error.log")
	if err := helpers.RemoveOldFileAndCreateNew(mainErrorLog); err != nil {
		return fmt.Errorf("failed to create main error log: %w", err)
	}

	// Create logs for each site
	dirs, err := helpers.ListDirectories(config.WWWDir)
	if err != nil {
		return fmt.Errorf("failed to list website directories: %w", err)
	}

	for _, dir := range dirs {
		baseName := filepath.Base(dir)
		domainName := baseName + "." + config.DomainTail

		// Create error log
		siteErrorLog := filepath.Join(config.LogsDir, "error-"+domainName+".log")
		if err := helpers.RemoveOldFileAndCreateNew(siteErrorLog); err != nil {
			return fmt.Errorf("failed to create error log for %s: %w", domainName, err)
		}

		// Create access log
		siteAccessLog := filepath.Join(config.LogsDir, "access-"+domainName+".log")
		if err := helpers.RemoveOldFileAndCreateNew(siteAccessLog); err != nil {
			return fmt.Errorf("failed to create access log for %s: %w", domainName, err)
		}
	}

	return nil
}

// checkNginxConfiguration checks if the Nginx configuration is valid
func checkNginxConfiguration(config *Configuration) error {
	log.Println("Checking Nginx configuration...")

	command := fmt.Sprintf("%s -p %s -t",
		filepath.Join(config.AppPath, config.ExecutableName),
		config.AppPath)

	if err := helpers.RunCommand(command, true); err != nil {
		return fmt.Errorf("nginx configuration test failed: %w", err)
	}

	return nil
}

// startNginx starts the Nginx server
func startNginx(config *Configuration) error {
	log.Println("Starting Nginx server...")

	command := fmt.Sprintf("%s -p %s",
		filepath.Join(config.AppPath, config.ExecutableName),
		config.AppPath)

	if err := helpers.RunCommand(command, true); err != nil {
		return fmt.Errorf("failed to start nginx: %w", err)
	}

	// Verify process is running
	time.Sleep(500 * time.Millisecond) // Give process time to start
	running, _ := helpers.IsProcessRunning(config.ExecutableName)
	if !running {
		return fmt.Errorf("nginx process failed to start")
	}

	return nil
}
