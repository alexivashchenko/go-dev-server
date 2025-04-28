package helpers

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func ReplaceBackslashToSlash(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

func ReadLinesIntoSlice(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return lines, nil
}

func RemoveOldFileAndCreateNew(file string) error {
	RemoveFile(file)
	err := CreateFile(file)
	if err != nil {
		return err
	}

	return nil
}

func CreateFile(filename string) error {
	if _, err := os.Stat(filename); err == nil {
		return fmt.Errorf("file already exists: %s", filename)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	file.Close()

	return nil
}

func RemoveFile(filename string) error {
	if _, err := os.Stat(filename); err == nil {
		if err := os.Remove(filename); err != nil {
			return fmt.Errorf("failed to remove file: %w", err)
		}
		return fmt.Errorf("file does not exist: %w", err)
	}
	return nil
}

func RunPowerShellAsAdmin(command string) error {
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf("Start-Process powershell -Verb RunAs '%s'", command))

	// fmt.Println("Running powershell command: " + command)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run '%s' command: %s, error: %w", command, string(output), err)
	}

	return nil
}

func ListDirectories(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs, nil
}

func AppendLines(filename string, lines []string) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	writer := strings.Builder{}
	for _, line := range lines {
		writer.WriteString(line + "\n")
	}

	if _, err := f.WriteString(writer.String()); err != nil {
		return fmt.Errorf("failed to write lines: %w", err)
	}

	return nil
}

func KillProcess(process string) error {
	// fmt.Println("Killing process: " + process)

	cmd := exec.Command("taskkill", "/F", "/IM", process)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run '%s' command: %w", process, err)
	}
	return nil
}

func RunCommand(command string, inBackground bool) error {
	// fmt.Println("Running command: " + command)

	if inBackground {
		runCommandInBackground(command)
	} else {
		runCommandAndWait(command)
	}

	return nil
}

func runCommandAndWait(command string) error {
	cmd := exec.Command("cmd", "/C", command)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run '%s' command: %w", command, err)
	}
	return nil
}

func runCommandInBackground(command string) error {
	cmd := exec.Command("cmd", "/C", command)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to run '%s' command: %w", command, err)
	}
	return nil
}

func ReplaceInFileByMap(filename string, replacements map[string]string) error {
	// Read content
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	newContent := string(content)
	for old, new := range replacements {
		newContent = strings.ReplaceAll(newContent, old, new)
	}

	// Write back
	err = os.WriteFile(filename, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil

}

func ReplaceInFile(filename, old, new string) error {
	// Read content
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Replace content
	newContent := strings.ReplaceAll(string(content), old, new)

	// Write back
	err = os.WriteFile(filename, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func CopyFileAsAdmin(source, destination string) error {

	err := RunPowerShellAsAdmin("Copy-Item -Path " + source + " -Destination " + destination + " -Force")
	if err != nil {
		return fmt.Errorf("failed to copy file: %s, error: %w", string(destination), err)
	}

	return nil
}

func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

func CreateDirectoryIfNotExists(directory string) bool {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if err := os.MkdirAll(directory, 0700); err != nil {
			fmt.Println("Failed to create directory: %w", err)
			os.Exit(0)
		}
		fmt.Printf("Created directory: %s\n", directory)
		return true
	}

	return false
}

func RemoveDirectoryAndContents(dir string) error {
	err := os.RemoveAll(dir)
	if err != nil {
		return fmt.Errorf("failed to remove directory: %w", err)
	}
	return nil
}

func GetCommand() (string, error) {

	command := "start"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	allowed_commands := []string{"start", "stop", "restart"}

	found := false
	for _, allowed_command := range allowed_commands {
		if allowed_command == command {
			found = true
			break
		}
	}

	if found {
		return command, nil
	}
	return "", errors.New("Command '" + command + "' not found")

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
