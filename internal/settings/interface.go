package settings

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func PromptForConfirmation(prompt string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(prompt + " [Y/n]: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			return false, fmt.Errorf("error reading input: %w", err)
		}

		input = strings.TrimSpace(input)
		println(input)
		switch input {
		case "Y", "y":
			return true, nil
		case "N", "n":
			return false, nil
		default:
			fmt.Println("Invalid input. Please enter Y or n.")
		}
	}
}

func EnsureFileExists(path, initialContent string) error {
	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directories: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.WriteFile(path, []byte(initialContent), 0644); err != nil {
			return fmt.Errorf("failed to write file: %v", err)
		}
	}

	return nil
}

func ParseTime(timeStr string) (uint, uint, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid time format")
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid hour value")
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid minute value")
	}

	if hour < 0 || hour > 23 {
		return 0, 0, fmt.Errorf("hour out of range")
	}

	if minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("minute out of range")
	}

	return uint(hour), uint(minute), nil
}
