package ssl

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/alexivashchenko/go-dev-server/helpers"
)

// Configuration holds all SSL-related paths and settings
type Configuration struct {
	RootDir                   string
	TemplatesDir              string
	SSLDir                    string
	WWWDir                    string
	OpenSSLConfigTemplateFile string
	OpenSSLConfigFile         string
	PrivateKeyFile            string
	CSRFile                   string
	CertificateFile           string
	NginxDomainTail           string
	ValidityDays              int
}

// NewConfiguration creates a new SSL configuration
func NewConfiguration() (*Configuration, error) {
	rootDir := helpers.GetRootDirectory()
	nginxDomainTail := os.Getenv("NGINX_DOMAIN_TAIL")
	if nginxDomainTail == "" {
		return nil, fmt.Errorf("NGINX_DOMAIN_TAIL environment variable is not set")
	}

	sslDir := filepath.Join(rootDir, "etc", "ssl")

	// Ensure SSL directory exists
	if err := os.MkdirAll(sslDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create SSL directory: %w", err)
	}

	return &Configuration{
		RootDir:                   rootDir,
		TemplatesDir:              filepath.Join(rootDir, "tpl"),
		SSLDir:                    sslDir,
		WWWDir:                    filepath.Join(rootDir, "www"),
		OpenSSLConfigTemplateFile: filepath.Join(rootDir, "tpl", "ssl", "openssl.conf.tpl"),
		OpenSSLConfigFile:         filepath.Join(sslDir, "openssl.conf"),
		PrivateKeyFile:            filepath.Join(sslDir, "private.key"),
		CSRFile:                   filepath.Join(sslDir, "csr.csr"),
		CertificateFile:           filepath.Join(sslDir, "certificate.crt"),
		NginxDomainTail:           nginxDomainTail,
		ValidityDays:              365,
	}, nil
}

// Start initializes SSL certificates
func Start() error {
	log.Println("Starting SSL configuration...")
	startTime := time.Now()

	config, err := NewConfiguration()
	if err != nil {
		return fmt.Errorf("failed to initialize SSL configuration: %w", err)
	}

	// Create OpenSSL config file from template
	if err := helpers.CopyFile(config.OpenSSLConfigTemplateFile, config.OpenSSLConfigFile); err != nil {
		return fmt.Errorf("failed to copy OpenSSL config template: %w", err)
	}

	// Generate DNS entries for the certificate
	dnsLines, err := generateDNSEntries(config)
	if err != nil {
		return fmt.Errorf("failed to generate DNS entries: %w", err)
	}

	// Append DNS entries to OpenSSL config
	if err := helpers.AppendLines(config.OpenSSLConfigFile, dnsLines); err != nil {
		return fmt.Errorf("failed to append DNS entries to OpenSSL config: %w", err)
	}

	// Check if certificate needs to be regenerated
	regenerate, err := shouldRegenerateCertificate(config)
	if err != nil {
		log.Printf("Warning: Could not determine if certificate needs regeneration: %v", err)
		regenerate = true
	}

	if regenerate {
		log.Println("Generating new SSL certificate...")
		if err := createCertificate(config); err != nil {
			return fmt.Errorf("failed to create certificate: %w", err)
		}
	} else {
		log.Println("Using existing SSL certificate (still valid)")
	}

	// Install certificate in the system trust store
	if err := installCertificate(config); err != nil {
		return fmt.Errorf("failed to install certificate: %w", err)
	}

	elapsed := time.Since(startTime)
	log.Printf("SSL configuration completed in %.2f seconds", elapsed.Seconds())
	return nil
}

// Stop removes SSL certificates from the system
func Stop() error {
	log.Println("Removing SSL certificates...")

	// Remove certificates from system trust store
	if err := uninstallCertificate(); err != nil {
		return fmt.Errorf("failed to uninstall certificate: %w", err)
	}

	log.Println("SSL certificates removed successfully")
	return nil
}

// Restart restarts the SSL configuration
func Restart() error {
	log.Println("Restarting SSL configuration...")

	if err := Stop(); err != nil {
		return fmt.Errorf("failed to stop SSL: %w", err)
	}

	if err := Start(); err != nil {
		return fmt.Errorf("failed to start SSL: %w", err)
	}

	return nil
}

// generateDNSEntries creates DNS entries for the certificate
func generateDNSEntries(config *Configuration) ([]string, error) {
	dnsLines := []string{
		"IP.1 = 127.0.0.1",
		"DNS.2 = localhost",
	}

	// Add local machine IP
	localIP, err := helpers.GetLocalIP()
	if err == nil && localIP != "" {
		dnsLines = append(dnsLines, fmt.Sprintf("IP.2 = %s", localIP))
	}

	// Get website directories
	dirs, err := helpers.ListDirectories(config.WWWDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list website directories: %w", err)
	}

	// Add DNS entries for each website directory
	dnsIndex := 3
	for _, dir := range dirs {
		baseName := filepath.Base(dir)
		domain := fmt.Sprintf("%s.%s", baseName, config.NginxDomainTail)

		dnsLines = append(dnsLines, fmt.Sprintf("DNS.%d = %s", dnsIndex, domain))
		dnsIndex++

		dnsLines = append(dnsLines, fmt.Sprintf("DNS.%d = *.%s", dnsIndex, domain))
		dnsIndex++
	}

	// Add wildcard entries
	dnsLines = append(dnsLines, fmt.Sprintf("DNS.%d = *.localhost", dnsIndex))
	dnsIndex++
	dnsLines = append(dnsLines, fmt.Sprintf("DNS.%d = *.%s", dnsIndex, config.NginxDomainTail))

	return dnsLines, nil
}

// shouldRegenerateCertificate checks if the certificate needs to be regenerated
func shouldRegenerateCertificate(config *Configuration) (bool, error) {
	// Check if certificate exists
	if _, err := os.Stat(config.CertificateFile); os.IsNotExist(err) {
		return true, nil
	}

	// TODO: Check certificate expiration date
	// TODO: Check if domains have changed since last generation

	return false, nil
}

// createCertificate generates a new SSL certificate
func createCertificate(config *Configuration) error {
	log.Println("Generating private key...")
	cmd := fmt.Sprintf("openssl genrsa -out %s 2048", config.PrivateKeyFile)
	if err := helpers.RunCommand(cmd, false); err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	log.Println("Generating certificate signing request...")
	cmd = fmt.Sprintf("openssl req -new -key %s -out %s -config %s",
		config.PrivateKeyFile, config.CSRFile, config.OpenSSLConfigFile)
	if err := helpers.RunCommand(cmd, false); err != nil {
		return fmt.Errorf("failed to generate CSR: %w", err)
	}

	log.Println("Generating self-signed certificate...")
	cmd = fmt.Sprintf("openssl x509 -req -days %d -in %s -signkey %s -out %s -extensions v3_req -extfile %s",
		config.ValidityDays, config.CSRFile, config.PrivateKeyFile, config.CertificateFile, config.OpenSSLConfigFile)
	if err := helpers.RunCommand(cmd, false); err != nil {
		return fmt.Errorf("failed to generate certificate: %w", err)
	}

	return nil
}

// installCertificate installs the certificate in the system trust store
func installCertificate(config *Configuration) error {
	switch runtime.GOOS {
	case "windows":
		return installWindowsCertificate(config.CertificateFile)
	case "darwin":
		return installMacCertificate(config.CertificateFile)
	case "linux":
		return installLinuxCertificate(config.CertificateFile)
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// uninstallCertificate removes the certificate from the system trust store
func uninstallCertificate() error {
	switch runtime.GOOS {
	case "windows":
		return uninstallWindowsCertificate()
	case "darwin":
		return uninstallMacCertificate()
	case "linux":
		return uninstallLinuxCertificate()
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// Windows-specific certificate installation
func installWindowsCertificate(certificateFile string) error {
	log.Println("Installing certificate in Windows trust store...")
	err := helpers.RunPowerShellAsAdmin(fmt.Sprintf(
		"Import-Certificate -FilePath \"%s\" -CertStoreLocation Cert:\\LocalMachine\\Root",
		certificateFile))
	if err != nil {
		return fmt.Errorf("failed to install Windows certificate: %w", err)
	}
	return nil
}

// Windows-specific certificate removal
func uninstallWindowsCertificate() error {
	log.Println("Removing certificate from Windows trust store...")
	err := helpers.RunPowerShellAsAdmin(
		"Get-ChildItem Cert:\\LocalMachine\\Root | Where-Object Subject -Like \"*local_server*\" | Remove-Item")
	if err != nil {
		return fmt.Errorf("failed to remove Windows certificate: %w", err)
	}
	return nil
}

// macOS-specific certificate installation
func installMacCertificate(certificateFile string) error {
	log.Println("Installing certificate in macOS trust store...")
	// Add certificate to keychain
	cmd := fmt.Sprintf("security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain %s", certificateFile)
	if err := helpers.RunCommand(cmd, true); err != nil {
		return fmt.Errorf("failed to install macOS certificate: %w", err)
	}
	return nil
}

// macOS-specific certificate removal
func uninstallMacCertificate() error {
	log.Println("Removing certificate from macOS trust store...")
	// Find and remove certificate from keychain
	cmd := "security find-certificate -a -c local_server -Z | grep SHA-1 | awk '{print $NF}' | xargs -I {} security delete-certificate -Z {}"
	if err := helpers.RunCommand(cmd, true); err != nil {
		return fmt.Errorf("failed to remove macOS certificate: %w", err)
	}
	return nil
}

// Linux-specific certificate installation
func installLinuxCertificate(certificateFile string) error {
	log.Println("Installing certificate in Linux trust store...")

	// Detect distribution
	if _, err := os.Stat("/etc/debian_version"); err == nil {
		// Debian/Ubuntu
		destPath := "/usr/local/share/ca-certificates/local_server.crt"
		if err := helpers.CopyFile(certificateFile, destPath); err != nil {
			return fmt.Errorf("failed to copy certificate: %w", err)
		}

		if err := helpers.RunCommand("update-ca-certificates", true); err != nil {
			return fmt.Errorf("failed to update CA certificates: %w", err)
		}
	} else if _, err := os.Stat("/etc/redhat-release"); err == nil {
		// RHEL/CentOS/Fedora
		destPath := "/etc/pki/ca-trust/source/anchors/local_server.crt"
		if err := helpers.CopyFile(certificateFile, destPath); err != nil {
			return fmt.Errorf("failed to copy certificate: %w", err)
		}

		if err := helpers.RunCommand("update-ca-trust extract", true); err != nil {
			return fmt.Errorf("failed to update CA certificates: %w", err)
		}
	} else {
		return fmt.Errorf("unsupported Linux distribution")
	}

	return nil
}

// Linux-specific certificate removal
func uninstallLinuxCertificate() error {
	log.Println("Removing certificate from Linux trust store...")

	// Detect distribution
	if _, err := os.Stat("/etc/debian_version"); err == nil {
		// Debian/Ubuntu
		if err := os.Remove("/usr/local/share/ca-certificates/local_server.crt"); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove certificate: %w", err)
		}

		if err := helpers.RunCommand("update-ca-certificates --fresh", true); err != nil {
			return fmt.Errorf("failed to update CA certificates: %w", err)
		}
	} else if _, err := os.Stat("/etc/redhat-release"); err == nil {
		// RHEL/CentOS/Fedora
		if err := os.Remove("/etc/pki/ca-trust/source/anchors/local_server.crt"); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove certificate: %w", err)
		}

		if err := helpers.RunCommand("update-ca-trust extract", true); err != nil {
			return fmt.Errorf("failed to update CA certificates: %w", err)
		}
	} else {
		return fmt.Errorf("unsupported Linux distribution")
	}

	return nil
}

// GetStatus returns the current status of the SSL configuration
func GetStatus() string {
	config, err := NewConfiguration()
	if err != nil {
		return "Error: " + err.Error()
	}

	if _, err := os.Stat(config.CertificateFile); os.IsNotExist(err) {
		return "Not configured"
	}

	// TODO: Check certificate expiration and return more detailed status
	return "Active"
}
