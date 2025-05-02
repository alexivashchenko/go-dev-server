package mysql

import (
	"fmt"
	"os"

	"github.com/alexivashchenko/go-dev-server/helpers"
)

func Start() {
	// fmt.Println("MySQL starting...")

	rootDir := helpers.GetRootDirectory()
	dirSeparator := string(os.PathSeparator)

	mysqlAppFolder := os.Getenv("MYSQL_APP_FOLDER")
	mysqlDataFolder := os.Getenv("MYSQL_DATA_FOLDER")
	mysqlAppDir := rootDir + dirSeparator + "apps" + dirSeparator + "mysql" + dirSeparator + mysqlAppFolder
	mysqlDataDir := rootDir + dirSeparator + "data" + dirSeparator + mysqlDataFolder
	templatesDir := rootDir + dirSeparator + "tpl"
	myIniFile := mysqlAppDir + dirSeparator + "my.ini"

	// fmt.Println("rootDir:", rootDir)
	// fmt.Println("mysqlAppFolder:", mysqlAppFolder)
	// fmt.Println("mysqlDataFolder:", mysqlDataFolder)
	// fmt.Println("mysqlAppDir:", mysqlAppDir)
	// fmt.Println("mysqlDataDir:", mysqlDataDir)

	shouldInitMySql := helpers.CreateDirectoryIfNotExists(mysqlDataDir)

	if shouldInitMySql {
		fmt.Println("Initialize MySQL")

		myIniTemplateFile := templatesDir + dirSeparator + "mysql" + dirSeparator + "my.ini.tpl"

		err := helpers.CopyFile(myIniTemplateFile, myIniFile)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(0)
		}

		err = helpers.ReplaceInFile(myIniFile, "{mysql_data_folder}", mysqlDataFolder)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(0)
		}

		// C:\server\apps\mysql\mysql-8.4.5-winx64\bin\mysqld --defaults-file=C:\server\apps\mysql\mysql-8.4.5-winx64\my.ini --initialize
		err = helpers.RunCommand(mysqlAppDir+dirSeparator+"bin"+dirSeparator+"mysqld --defaults-file="+myIniFile+" --initialize", false)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(0)
		}

	}

	err := helpers.RunCommand(mysqlAppDir+dirSeparator+"bin"+dirSeparator+"mysqld --defaults-file="+myIniFile, true)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}

	// fmt.Println("MySQL started.")

}

func Stop() {
	// fmt.Println("MySQL stopping...")

	err := helpers.KillProcess("mysqld.exe")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(0)
	}

	// fmt.Println("MySQL stopped.")
}

func Restart() {
	Stop()
	Start()
}
