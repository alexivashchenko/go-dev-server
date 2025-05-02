package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/alexivashchenko/go-dev-server/env"
	"github.com/alexivashchenko/go-dev-server/helpers"
	"github.com/alexivashchenko/go-dev-server/mysql"
	"github.com/alexivashchenko/go-dev-server/nginx"
	"github.com/alexivashchenko/go-dev-server/php"
	"github.com/alexivashchenko/go-dev-server/ssl"
)

// Service represents a server component that can be started, stopped, and restarted
type Service struct {
	Name    string
	Start   func() error
	Stop    func() error
	Restart func() error
}

func main() {
	// Load environment variables
	env.Load()

	// Get command from arguments
	command, err := helpers.GetCommand()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		printUsage()
		os.Exit(1)
	}

	// Define services
	services := []Service{
		{
			Name:    "MySQL",
			Start:   wrapFunc(mysql.Start),
			Stop:    wrapFunc(mysql.Stop),
			Restart: wrapFunc(mysql.Restart),
		},
		{
			Name:    "SSL",
			Start:   wrapFunc(ssl.Start),
			Stop:    wrapFunc(ssl.Stop),
			Restart: wrapFunc(ssl.Restart),
		},
		{
			Name:    "PHP",
			Start:   wrapFunc(php.Start),
			Stop:    wrapFunc(php.Stop),
			Restart: wrapFunc(php.Restart),
		},
		{
			Name:    "Nginx",
			Start:   wrapFunc(nginx.Start),
			Stop:    wrapFunc(nginx.Stop),
			Restart: wrapFunc(nginx.Restart),
		},
	}

	// Process command
	switch command {
	case "start":
		startServices(services)
	case "stop":
		stopServices(services)
	case "restart":
		restartServices(services)
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

// wrapFunc wraps a void function to return an error
func wrapFunc(fn func()) func() error {
	return func() error {
		fn()
		return nil // Assuming the original functions handle their own errors
	}
}

// startServices starts all services in parallel
func startServices(services []Service) {
	fmt.Println("Starting all services...")

	var wg sync.WaitGroup
	errChan := make(chan error, len(services))

	for _, service := range services {
		wg.Add(1)
		go func(s Service) {
			defer wg.Done()
			fmt.Printf("Starting %s...\n", s.Name)
			start := time.Now()
			if err := s.Start(); err != nil {
				errChan <- fmt.Errorf("failed to start %s: %w", s.Name, err)
				return
			}
			elapsed := time.Since(start)
			fmt.Printf("%s started successfully in %.2f seconds\n", s.Name, elapsed.Seconds())
		}(service)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("All services started successfully")
}

// stopServices stops all services in parallel
func stopServices(services []Service) {
	fmt.Println("Stopping all services...")

	var wg sync.WaitGroup
	for _, service := range services {
		wg.Add(1)
		go func(s Service) {
			defer wg.Done()
			fmt.Printf("Stopping %s...\n", s.Name)
			if err := s.Stop(); err != nil {
				fmt.Printf("Warning: Failed to stop %s: %v\n", s.Name, err)
				return
			}
			fmt.Printf("%s stopped successfully\n", s.Name)
		}(service)
	}

	wg.Wait()
	fmt.Println("All services stopped")
}

// restartServices restarts all services in parallel
func restartServices(services []Service) {
	fmt.Println("Restarting all services...")

	var wg sync.WaitGroup
	errChan := make(chan error, len(services))

	for _, service := range services {
		wg.Add(1)
		go func(s Service) {
			defer wg.Done()
			fmt.Printf("Restarting %s...\n", s.Name)
			start := time.Now()
			if err := s.Restart(); err != nil {
				errChan <- fmt.Errorf("failed to restart %s: %w", s.Name, err)
				return
			}
			elapsed := time.Since(start)
			fmt.Printf("%s restarted successfully in %.2f seconds\n", s.Name, elapsed.Seconds())
		}(service)
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("All services restarted successfully")
}

// printUsage prints usage information
func printUsage() {
	fmt.Println("Usage: server <command>")
	fmt.Println("\nAvailable commands:")
	fmt.Println("  start    - Start all services")
	fmt.Println("  stop     - Stop all services")
	fmt.Println("  restart  - Restart all services")
	fmt.Println("  help     - Show this help message")
}
