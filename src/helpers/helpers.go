package helpers

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// FileSystem operations

// ReplaceBackslashToSlash converts Windows backslashes to forward slashes
func ReplaceBackslashToSlash(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

// ReadLinesIntoSlice reads a file and returns its contents as a slice of strings
func ReadLinesIntoSlice(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	return lines, nil
}

// RemoveOldFileAndCreateNew removes a file if it exists and creates a new one
func RemoveOldFileAndCreateNew(file string) error {
	// Remove file if it exists
	if _, err := os.Stat(file); err == nil {
		if err := os.Remove(file); err != nil {
			return fmt.Errorf("failed to remove existing file %s: %w", file, err)
		}
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create new file
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", file, err)
	}
	defer f.Close()

	return nil
}

// CreateFile creates a new file, ensuring its directory exists
func CreateFile(filename string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	return nil
}

// RemoveFile removes a file if it exists
func RemoveFile(filename string) error {
	if _, err := os.Stat(filename); err == nil {
		if err := os.Remove(filename); err != nil {
			return fmt.Errorf("failed to remove file %s: %w", filename, err)
		}
		return nil
	} else if os.IsNotExist(err) {
		return nil // File doesn't exist, which is fine
	} else {
		return fmt.Errorf("failed to check if file %s exists: %w", filename, err)
	}
}

// ListDirectories returns a list of directory names in the given path
func ListDirectories(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", path, err)
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs, nil
}

// AppendLines appends lines to a file
func AppendLines(filename string, lines []string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer f.Close()

	writer := strings.Builder{}
	for _, line := range lines {
		writer.WriteString(line + "\n")
	}

	if _, err := f.WriteString(writer.String()); err != nil {
		return fmt.Errorf("failed to write lines to %s: %w", filename, err)
	}

	return nil
}

// ReplaceInFileByMap replaces multiple strings in a file based on a map
func ReplaceInFileByMap(filename string, replacements map[string]string) error {
	// Read content
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	newContent := string(content)
	for old, new := range replacements {
		newContent = strings.ReplaceAll(newContent, old, new)
	}

	// Write back
	err = os.WriteFile(filename, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	return nil
}

// ReplaceInFile replaces a string in a file
func ReplaceInFile(filename, old, new string) error {
	// Read content
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Replace content
	newContent := strings.ReplaceAll(string(content), old, new)

	// Write back
	err = os.WriteFile(filename, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	return nil
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy from %s to %s: %w", src, dst, err)
	}

	// Copy permissions from source file
	sourceInfo, err := os.Stat(src)
	if err == nil {
		err = os.Chmod(dst, sourceInfo.Mode())
		if err != nil {
			log.Printf("Warning: Failed to copy file permissions: %v", err)
		}
	}

	return nil
}

// CopyFileAsAdmin copies a file with administrative privileges
func CopyFileAsAdmin(source, destination string) error {
	if runtime.GOOS == "windows" {
		err := RunPowerShellAsAdmin(fmt.Sprintf("Copy-Item -Path \"%s\" -Destination \"%s\" -Force", source, destination))
		if err != nil {
			return fmt.Errorf("failed to copy file as admin from %s to %s: %w", source, destination, err)
		}
	} else {
		// For Unix-like systems, try to use sudo
		cmd := exec.Command("sudo", "cp", source, destination)
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to copy file as admin: %s, error: %w", string(output), err)
		}
	}

	return nil
}

// CreateDirectoryIfNotExists creates a directory if it doesn't exist
func CreateDirectoryIfNotExists(directory string) bool {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if err := os.MkdirAll(directory, 0755); err != nil {
			log.Printf("Failed to create directory %s: %v", directory, err)
			return false
		}
		log.Printf("Created directory: %s", directory)
		return true
	}

	return false
}

// RemoveDirectoryAndContents removes a directory and all its contents
func RemoveDirectoryAndContents(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, which is fine
	}

	err := os.RemoveAll(dir)
	if err != nil {
		return fmt.Errorf("failed to remove directory %s: %w", dir, err)
	}
	return nil
}

// Process management

// RunPowerShellAsAdmin runs a PowerShell command with administrative privileges
func RunPowerShellAsAdmin(command string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("RunPowerShellAsAdmin is only supported on Windows")
	}

	// log.Printf("Running PowerShell command as admin: %s", command)

	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf("Start-Process powershell -Verb RunAs -Wait -ArgumentList '-NoProfile -ExecutionPolicy Bypass -Command \"%s\"'", command))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run PowerShell command: %s, error: %w", string(output), err)
	}

	return nil
}

// KillProcess terminates a process by name
func KillProcess(process string) error {
	log.Printf("Killing process: %s", process)

	if runtime.GOOS == "windows" {
		cmd := exec.Command("taskkill", "/F", "/IM", process)
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Check if the error is because the process is not running
			if strings.Contains(string(output), "not found") {
				log.Printf("Process %s is not running", process)
				return nil
			}
			return fmt.Errorf("failed to kill process %s: %s, error: %w", process, string(output), err)
		}
	} else {
		// For Unix-like systems
		cmd := exec.Command("pkill", "-f", process)
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Exit code 1 means no processes matched, which is fine
			if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
				log.Printf("Process %s is not running", process)
				return nil
			}
			return fmt.Errorf("failed to kill process %s: %s, error: %w", process, string(output), err)
		}
	}

	return nil
}

// RunCommand executes a command
func RunCommand(command string, inBackground bool) error {
	// log.Printf("Running command: %s", command)

	var err error
	if inBackground {
		err = runCommandInBackground(command)
	} else {
		_, err = RunCommandWithOutput(command)
	}

	return err
}

// RunCommandWithOutput executes a command and returns its output
func RunCommandWithOutput(command string) (string, error) {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("command failed: %s, error: %w", string(output), err)
	}

	return string(output), nil
}

// RunCommandInDirectory executes a command in a specific directory
func RunCommandInDirectory(command string, directory string, inBackground bool) error {
	log.Printf("Running command in directory %s: %s", directory, command)

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	cmd.Dir = directory

	if inBackground {
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start command: %w", err)
		}

		// Detach the process
		if runtime.GOOS != "windows" {
			setProcessGroupID(cmd)
		}

		return nil
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("command failed: %s, error: %w", string(output), err)
	}

	return nil
}

// runCommandAndWait executes a command and waits for it to complete
func runCommandAndWait(command string) (string, error) {
	return RunCommandWithOutput(command)
}

// runCommandInBackground executes a command in the background
func runCommandInBackground(command string) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	// Detach the process on Unix systems
	if runtime.GOOS != "windows" {
		setProcessGroupID(cmd)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	return nil
}

// IsProcessRunning checks if a process is running and returns its PID
func IsProcessRunning(processName string) (bool, int) {
	if runtime.GOOS == "windows" {
		// For Windows, use PowerShell to check if process is running
		command := fmt.Sprintf("powershell -Command \"Get-Process -Name '%s' -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Id\"",
			strings.TrimSuffix(processName, ".exe"))
		output, err := RunCommandWithOutput(command)
		if err != nil || output == "" {
			return false, 0
		}

		// Parse the PID from PowerShell output
		output = strings.TrimSpace(output)
		if output != "" {
			var pid int
			// PowerShell may return multiple PIDs, take the first one
			firstLine := strings.Split(output, "\n")[0]
			fmt.Sscanf(firstLine, "%d", &pid)
			return true, pid
		}
	} else {
		// For Unix-like systems
		output, err := RunCommandWithOutput(fmt.Sprintf("pgrep -f %s", processName))
		if err == nil && output != "" {
			var pid int
			fmt.Sscanf(output, "%d", &pid)
			return true, pid
		}
	}

	return false, 0
}

// Application helpers

// GetCommand returns the command from command line arguments
func GetCommand() (string, error) {
	command := "start"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	allowedCommands := []string{"start", "stop", "restart", "status", "help"}

	for _, allowedCommand := range allowedCommands {
		if allowedCommand == command {
			return command, nil
		}
	}

	return "", fmt.Errorf("unknown command: %s", command)
}

func GetRootDirectory() string {

	currentDir, _ := os.Getwd()

	parentDir := filepath.Dir(currentDir)

	if checkBuild() {
		return parentDir

	}

	return currentDir
}

func checkBuild() bool {
	cmd := exec.Command("go", "build", "-v")
	err := cmd.Run()

	return err == nil
}

// // GetRootDirectory returns the root directory of the application
// func GetRootDirectory() string {
// 	execPath, err := os.Executable()
// 	if err == nil {
// 		// Use the directory of the executable
// 		return filepath.Dir(execPath)
// 	}

// 	// Fallback to current directory
// 	currentDir, err := os.Getwd()
// 	if err != nil {
// 		log.Printf("Warning: Failed to get current directory: %v", err)
// 		return "."
// 	}

// 	// Check if we're in development mode
// 	if isDevMode() {
// 		// In development mode, use the parent directory if we're in the src directory
// 		if filepath.Base(currentDir) == "src" {
// 			return filepath.Dir(currentDir)
// 		}
// 	}

// 	return currentDir
// }

// // isDevMode checks if the application is running in development mode
// func isDevMode() bool {
// 	// Check if go.mod exists in the current directory
// 	_, err := os.Stat("go.mod")
// 	return err == nil
// }

// // checkBuild checks if the go build command is available
// func checkBuild() bool {
// 	cmd := exec.Command("go", "version")
// 	return cmd.Run() == nil
// }

// Network helpers

// GetLocalIP returns the local IP address
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("failed to get interface addresses: %w", err)
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", errors.New("no local IP address found")
}

// WaitForPort checks if a port is open
func WaitForPort(host string, port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), time.Second)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for %s:%d", host, port)
}
