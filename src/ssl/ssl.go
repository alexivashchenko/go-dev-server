package ssl

import (
	"fmt"
	"os"

	"github.com/alexivashchenko/go-dev-server/helpers"
)

func Start() {
	fmt.Println("SSL creating...")

	rootDir := helpers.GetRootDirectory()
	dirSeparator := string(os.PathSeparator)

	nginxDomainTail := os.Getenv("NGINX_DOMAIN_TAIL")

	templatesDir := rootDir + dirSeparator + "tpl"
	sslDir := rootDir + dirSeparator + "etc"
	wwwDir := rootDir + dirSeparator + "www"

	openSslConfigTemplateFile := templatesDir + dirSeparator + "ssl" + dirSeparator + "openssl.conf.tpl"
	openSslConfigFile := sslDir + dirSeparator + "ssl" + dirSeparator + "openssl.conf"
	privateKeyFile := sslDir + dirSeparator + "ssl" + dirSeparator + "private.key"
	csrFile := sslDir + dirSeparator + "ssl" + dirSeparator + "csr.csr"
	certificateFile := sslDir + dirSeparator + "ssl" + dirSeparator + "certificate.crt"

	// fmt.Println("nginxDomainTail:", nginxDomainTail)
	// fmt.Println("templatesDir:", templatesDir)
	// fmt.Println("openSslConfigTemplateFile:", openSslConfigTemplateFile)
	// fmt.Println("openSslConfigFile:", openSslConfigFile)
	// fmt.Println("wwwDir:", wwwDir)
	// fmt.Println("sslDir:", sslDir)

	err := helpers.CopyFile(openSslConfigTemplateFile, openSslConfigFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}

	dnsLines := []string{
		"IP.1 = 127.0.0.1",
		"DNS.2 = localhost",
		// TODO: Add current machine local IP
	}

	dirs, err := helpers.ListDirectories(wwwDir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}

	dnsIndex := 3
	for _, dir := range dirs {
		// fmt.Println(dir)
		const dnsFormat = "DNS.%v = %s"
		const dnsWildcardFormat = "DNS.%v = *.%s"
		dnsLines = append(dnsLines, fmt.Sprintf(dnsFormat, dnsIndex, dir+"."+nginxDomainTail))
		dnsIndex++
		dnsLines = append(dnsLines, fmt.Sprintf(dnsWildcardFormat, dnsIndex, dir+"."+nginxDomainTail))
		dnsIndex++
	}

	dnsLines = append(dnsLines, fmt.Sprintf("DNS.%v = *.localhost", dnsIndex)) // DNS.N = *.localhost
	dnsIndex++

	dnsLines = append(dnsLines, fmt.Sprintf("DNS.%v = *.%s", dnsIndex, nginxDomainTail)) // DNS.N = *.oo
	dnsIndex++

	// for _, dnsLine := range dnsLines {
	// 	fmt.Println("dnsLine:", dnsLine)
	// }

	err = helpers.AppendLines(openSslConfigFile, dnsLines)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}

	// TODO: Create/Update Certificate only if domains was changed
	createCertificate(privateKeyFile, csrFile, openSslConfigFile, certificateFile)

	deleteWindowsCertificate()

	addWindowsCertificate(certificateFile)

	fmt.Println("SSL created.")
}

func Stop() {
	fmt.Println("SSL deleting...")

	deleteWindowsCertificate()

	fmt.Println("SSL deleted.")
}

func Restart() {
	Stop()
	Start()
}

func createCertificate(privateKeyFile string, csrFile string, openSslConfigFile string, certificateFile string) {
	// https://arie-m-prasetyo.medium.com/local-secure-web-server-with-nginx-and-ssl-125256e7a2f5

	// create Private Key
	// openssl genrsa -verbose -out ${PWD}/etc/ssl/private.key 2048
	err := helpers.RunCommand("openssl genrsa -verbose -out "+privateKeyFile+" 2048", false)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}

	// generate CSR
	// openssl req -new -key ${PWD}/etc/ssl/private.key -out ${PWD}/etc/ssl/csr.csr -verbose -config "${PWD}/etc/ssl/openssl.conf"
	err = helpers.RunCommand("openssl req -new -key "+privateKeyFile+" -out "+csrFile+" -verbose -config "+openSslConfigFile, false)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}

	// generate Certificate
	// openssl x509 -req -days 365 -in ${PWD}/etc/ssl/csr.csr -signkey ${PWD}/etc/ssl/private.key -out ${PWD}/etc/ssl/certificate.crt -extensions v3_req -extfile "${PWD}/etc/ssl/openssl.conf"
	err = helpers.RunCommand("openssl x509 -req -days 365 -in "+csrFile+" -signkey "+privateKeyFile+" -out "+certificateFile+" -extensions v3_req -extfile "+openSslConfigFile, false)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}
}

func addWindowsCertificate(certificateFile string) {
	// add new certificate to Windows
	// powershell -Command "Start-Process -Verb RunAs powershell \"Import-Certificate -FilePath "C:\\server\\etc\\ssl\\certificate.crt" -CertStoreLocation Cert:\\LocalMachine\\Root\""

	err := helpers.RunPowerShellAsAdmin("Import-Certificate -FilePath \"" + certificateFile + "\" -CertStoreLocation Cert:\\LocalMachine\\Root")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}
}

func deleteWindowsCertificate() {
	// remove old Windows Certificates
	// powershell -Command "Start-Process -Verb RunAs powershell \"Get-ChildItem Cert:\\LocalMachine\\Root | Where-Object Subject -Like '*local_server*' | Remove-Item\""

	err := helpers.RunPowerShellAsAdmin("Get-ChildItem Cert:\\LocalMachine\\Root | Where-Object Subject -Like \"*local_server*\" | Remove-Item")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}
}
