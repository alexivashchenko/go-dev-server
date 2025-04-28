package main

import (
	"fmt"
	"os"

	"github.com/alexivashchenko/go-dev-server/env"
	"github.com/alexivashchenko/go-dev-server/helpers"
	"github.com/alexivashchenko/go-dev-server/mysql"
	"github.com/alexivashchenko/go-dev-server/nginx"
	"github.com/alexivashchenko/go-dev-server/php"
	"github.com/alexivashchenko/go-dev-server/ssl"
)

func main() {

	env.Load()

	command, err := helpers.GetCommand()

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	if command == "start" {
		mysql.Start()
		ssl.Start()
		php.Start()
		nginx.Start()
		os.Exit(0)
	}

	if command == "stop" {
		mysql.Stop()
		ssl.Stop()
		php.Stop()
		nginx.Stop()
		os.Exit(0)
	}

	if command == "restart" {
		mysql.Restart()
		ssl.Restart()
		php.Restart()
		nginx.Restart()
		os.Exit(0)
	}
}
