package php

import (
	"fmt"
	"os"
	"strings"

	"github.com/alexivashchenko/go-dev-server/helpers"
)

func Start() {
	// fmt.Println("PHP starting...")

	rootDir := helpers.GetRootDirectory()
	dirSeparator := string(os.PathSeparator)

	// PHP_APP_FOLDER='php-8.3.16-Win32-vs16-x64'
	phpAppFolder := os.Getenv("PHP_APP_FOLDER")
	// PHP_ERROR_LOG="{ROOT_DIR}\logs\php\php_errors.log"
	phpErrorLog := strings.Replace(os.Getenv("PHP_ERROR_LOG"), "{ROOT_DIR}", rootDir, 1)
	// PHP_INCLUDE_PATH=".;{ROOT_DIR}\etc\php\pear"
	phpIncludePath := strings.Replace(os.Getenv("PHP_INCLUDE_PATH"), "{ROOT_DIR}", rootDir, 1)
	// PHP_EXTENSION_DIR="{ROOT_DIR}\bin\php\{PHP_APP_FOLDER}\ext"
	phpExtensionDir := strings.Replace(os.Getenv("PHP_EXTENSION_DIR"), "{ROOT_DIR}", rootDir, 1)
	phpExtensionDir = strings.Replace(phpExtensionDir, "{PHP_APP_FOLDER}", phpAppFolder, 1)
	// PHP_SESSION_SAVE_PATH="{ROOT_DIR}\tmp"
	phpSessionSavePath := strings.Replace(os.Getenv("PHP_SESSION_SAVE_PATH"), "{ROOT_DIR}", rootDir, 1)
	// PHP_CURL_CAINFO="{ROOT_DIR}\tmp\etc\ssl\cacert.pem"
	phpCurlCainfo := strings.Replace(os.Getenv("PHP_CURL_CAINFO"), "{ROOT_DIR}", rootDir, 1)
	// PHP_SENDMAIL_PATHPHP_SENDMAIL_PATH=""
	phpSendmailPath := strings.Replace(os.Getenv("PHP_SENDMAIL_PATH"), "{ROOT_DIR}", rootDir, 1)

	templatesDir := rootDir + dirSeparator + "tpl"
	phpAppDir := rootDir + dirSeparator + "apps" + dirSeparator + "php" + dirSeparator + phpAppFolder
	phpIniTemplateFile := templatesDir + dirSeparator + "php" + dirSeparator + "php.ini.tpl"
	phpIniFile := phpAppDir + dirSeparator + "php.ini"

	// fmt.Println("rootDir:", rootDir)
	// fmt.Println("phpAppFolder:", phpAppFolder)
	// fmt.Println("phpErrorLog:", phpErrorLog)
	// fmt.Println("phpIncludePath:", phpIncludePath)
	// fmt.Println("phpExtensionDir:", phpExtensionDir)
	// fmt.Println("phpSessionSavePath:", phpSessionSavePath)
	// fmt.Println("phpSendmailPath:", phpSendmailPath)
	// fmt.Println("phpCurlCainfo:", phpCurlCainfo)

	// fmt.Println("phpIniTemplateFile:", phpIniTemplateFile)
	// fmt.Println("phpIniFile:", phpIniFile)

	err := helpers.CopyFile(phpIniTemplateFile, phpIniFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}

	replaceMap := map[string]string{
		"{PHP_APP_FOLDER}":        phpAppFolder,
		"{PHP_ERROR_LOG}":         phpErrorLog,
		"{PHP_INCLUDE_PATH}":      phpIncludePath,
		"{PHP_EXTENSION_DIR}":     phpExtensionDir,
		"{PHP_SESSION_SAVE_PATH}": phpSessionSavePath,
		"{PHP_CURL_CAINFO}":       phpCurlCainfo, // TODO: Find out what is this and maybe generate the file
		"{PHP_SENDMAIL_PATH}":     phpSendmailPath,
	}

	err = helpers.ReplaceInFileByMap(phpIniFile, replaceMap)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}

	err = helpers.RunCommand("php-cgi.exe -b 127.0.0.1:9003", true)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}

	// fmt.Println("PHP started.")

}

func Stop() {
	// fmt.Println("PHP stopping...")

	err := helpers.KillProcess("php-cgi.exe")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}

	// fmt.Println("PHP stopped.")
}

func Restart() {
	Stop()
	Start()
}
