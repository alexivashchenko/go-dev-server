package nginx

import (
	"fmt"
	"os"
	"strings"

	"github.com/alexivashchenko/go-dev-server/helpers"
)

func Stop() {
	// fmt.Println("NGINX stopping...")

	err := helpers.KillProcess("nginx.exe")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}

	// fmt.Println("NGINX stopped.")
}

func Restart() {
	Stop()
	Start()
}

func Start() {
	// fmt.Println("NGINX starting...")

	rootDir := helpers.GetRootDirectory()
	dirSeparator := string(os.PathSeparator)

	nginxAppFolder := rootDir + dirSeparator + "apps" + dirSeparator + "nginx" + dirSeparator + os.Getenv("NGINX_APP_FOLDER")
	// nginxAppConfigFolder := nginxAppFolder + dirSeparator + "conf"
	nginxDomainTail := os.Getenv("NGINX_DOMAIN_TAIL")
	nginxEtcFolder := rootDir + dirSeparator + "etc" + dirSeparator + "nginx"
	nginxSitesEnabledFolder := nginxEtcFolder + dirSeparator + "sites-enabled"

	hostsFileIdentifier := "#local server setting"
	hostsFilePath := "C:" + dirSeparator + "Windows" + dirSeparator + "System32" + dirSeparator + "drivers" + dirSeparator + "etc" + dirSeparator + "hosts"
	tmpHostsFilePath := rootDir + dirSeparator + "hosts.tmp"
	wwwDir := rootDir + dirSeparator + "www"
	templatesDir := rootDir + dirSeparator + "tpl"
	generalSiteConfTemplateFile := templatesDir + dirSeparator + "nginx" + dirSeparator + "general-site.conf.tpl"

	// fmt.Println("rootDir:", rootDir)
	// fmt.Println("nginxAppFolder:", nginxAppFolder)
	// fmt.Println("nginxDomainTail:", nginxDomainTail)
	// fmt.Println("nginxAppConfigFolder:", nginxAppConfigFolder)
	// fmt.Println("nginxEtcFolder:", nginxEtcFolder)
	// fmt.Println("nginxSitesEnabledFolder:", nginxSitesEnabledFolder)
	// fmt.Println("hostsFileIdentifier:", hostsFileIdentifier)
	// fmt.Println("hostsFilePath:", hostsFilePath)
	// fmt.Println("tmpHostsFilePath:", tmpHostsFilePath)
	// fmt.Println("wwwDir:", wwwDir)

	buildHostsFile(
		hostsFilePath,
		tmpHostsFilePath,
		hostsFileIdentifier,
		nginxDomainTail,
		wwwDir,
	)

	copyNginxConfFile(
		nginxAppFolder,
		nginxDomainTail,
		wwwDir,
		templatesDir,
		dirSeparator,
		rootDir,
	)

	copyDefaultConfFile(
		nginxSitesEnabledFolder,
		templatesDir,
		dirSeparator,
		rootDir,
	)

	nginxLogsFilesDirectory := rootDir + dirSeparator + "logs" + dirSeparator + "nginx"
	err := helpers.RemoveDirectoryAndContents(nginxLogsFilesDirectory)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	helpers.CreateDirectoryIfNotExists(nginxLogsFilesDirectory)

	dirs, err := helpers.ListDirectories(wwwDir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}

	for _, dir := range dirs {
		createFilesForEachSite(
			dir,
			nginxDomainTail,
			nginxSitesEnabledFolder,
			dirSeparator,
			generalSiteConfTemplateFile,
			rootDir,
			wwwDir,
		)
	}

	createNginxErrorLogFile(
		rootDir,
		dirSeparator,
	)

	checkNginxConfiguration(
		nginxAppFolder,
		dirSeparator,
	)

	runNginx(
		nginxAppFolder,
		dirSeparator,
	)

	// fmt.Println("NGINX started.")

}

func runNginx(
	nginxAppFolder string,
	dirSeparator string,
) {
	err := helpers.RunCommand(nginxAppFolder+dirSeparator+"nginx.exe -p "+nginxAppFolder+dirSeparator, true)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}
}

func checkNginxConfiguration(
	nginxAppFolder string,
	dirSeparator string,
) {
	err := helpers.RunCommand(nginxAppFolder+dirSeparator+"nginx.exe -p "+nginxAppFolder+dirSeparator+" -t", true)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}
}

func createNginxErrorLogFile(
	rootDir string,
	dirSeparator string,
) {
	nginxErrorLogFile := rootDir + dirSeparator + "logs" + dirSeparator + "nginx" + dirSeparator + "error.log"
	err := helpers.RemoveOldFileAndCreateNew(nginxErrorLogFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}
}

func createFilesForEachSite(
	dir string,
	nginxDomainTail string,
	nginxSitesEnabledFolder string,
	dirSeparator string,
	generalSiteConfTemplateFile string,
	rootDir string,
	wwwDir string,
) {

	domainName := dir + "." + nginxDomainTail
	siteConfFile := nginxSitesEnabledFolder + dirSeparator + domainName + ".conf"

	err := helpers.CopyFile(generalSiteConfTemplateFile, siteConfFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}

	err = helpers.ReplaceInFile(siteConfFile, "{root_folder}", helpers.ReplaceBackslashToSlash(rootDir+dirSeparator))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}

	err = helpers.ReplaceInFile(siteConfFile, "{folder_name}", dir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}

	err = helpers.ReplaceInFile(siteConfFile, "{domain_name}", domainName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}

	siteErrorLogFile := rootDir + dirSeparator + "logs" + dirSeparator + "nginx" + dirSeparator + "error-" + domainName + ".log"
	err = helpers.RemoveOldFileAndCreateNew(siteErrorLogFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}

	siteAccessLogFile := rootDir + dirSeparator + "logs" + dirSeparator + "nginx" + dirSeparator + "access-" + domainName + ".log"
	err = helpers.RemoveOldFileAndCreateNew(siteAccessLogFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}
}

func copyDefaultConfFile(
	nginxSitesEnabledFolder string,
	templatesDir string,
	dirSeparator string,
	rootDir string,
) {
	err := helpers.RemoveDirectoryAndContents(nginxSitesEnabledFolder)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	helpers.CreateDirectoryIfNotExists(nginxSitesEnabledFolder)

	defaultConfTemplateFile := templatesDir + dirSeparator + "nginx" + dirSeparator + "00-default.conf.tpl"
	defaultConfFile := nginxSitesEnabledFolder + dirSeparator + "00-default.conf"
	err = helpers.CopyFile(defaultConfTemplateFile, defaultConfFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}
	// sed -i "s/{root_folder}/$root_folder/g" "$nginx_app_folder_path/conf/nginx.conf"
	err = helpers.ReplaceInFile(defaultConfFile, "{root_folder}", helpers.ReplaceBackslashToSlash(rootDir+dirSeparator))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}
}

func copyNginxConfFile(
	nginxAppFolder string,
	nginxDomainTail string,
	wwwDir string,
	templatesDir string,
	dirSeparator string,
	rootDir string,
) {
	nginxConfTemplateFile := templatesDir + dirSeparator + "nginx" + dirSeparator + "nginx.conf.tpl"
	nginxConfFile := nginxAppFolder + dirSeparator + "conf" + dirSeparator + "nginx.conf"
	// cp "${PWD}/lib/nginx.conf.example" "$nginx_app_folder_path/conf/nginx.conf"
	err := helpers.CopyFile(nginxConfTemplateFile, nginxConfFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}
	// sed -i "s/{root_folder}/$root_folder/g" "$nginx_app_folder_path/conf/nginx.conf"
	err = helpers.ReplaceInFile(nginxConfFile, "{root_folder}", helpers.ReplaceBackslashToSlash(rootDir+dirSeparator))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}
}

func buildHostsFile(
	hostsFilePath string,
	tmpHostsFilePath string,
	hostsFileIdentifier string,
	nginxDomainTail string,
	wwwDir string,
) {

	helpers.RemoveFile(tmpHostsFilePath)

	err := helpers.CreateFile(tmpHostsFilePath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}
	// defer helpers.RemoveFile(tmpHostsFilePath) // Cannot delete at the script end since it's brake copying to C:\Windows\System32\drivers\etc\hosts

	lines, err := helpers.ReadLinesIntoSlice(hostsFilePath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	newLines := []string{}

	for _, line := range lines {
		if !strings.Contains(line, hostsFileIdentifier) {
			newLines = append(newLines, line)
			continue
		}
	}

	dirs, err := helpers.ListDirectories(wwwDir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}

	for _, dir := range dirs {
		const hostFormat = "127.0.0.1\t%s\t%s"
		newLines = append(newLines, fmt.Sprintf(hostFormat, dir+"."+nginxDomainTail, hostsFileIdentifier))
	}

	// for _, line := range newLines {
	// 	fmt.Println(line)
	// }

	err = helpers.AppendLines(tmpHostsFilePath, newLines)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}

	err = helpers.CopyFileAsAdmin(tmpHostsFilePath, hostsFilePath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}
}
