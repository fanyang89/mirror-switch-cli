package backup

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const backupSuffix = ".bak"

func CreateBackup(filePath string) (string, error) {
	backupPath := filePath + backupSuffix
	source, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer source.Close()

	destination, err := os.Create(backupPath)
	if err != nil {
		return "", err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	if err != nil {
		return "", err
	}

	return backupPath, nil
}

func RestoreBackups(dirPath string) ([]string, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	var restored []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), backupSuffix) {
			backupPath := filepath.Join(dirPath, file.Name())
			originalPath := strings.TrimSuffix(backupPath, backupSuffix)

			if err := restoreFile(backupPath, originalPath); err != nil {
				return restored, fmt.Errorf("failed to restore %s: %w", backupPath, err)
			}
			restored = append(restored, originalPath)
		}
	}

	return restored, nil
}

func restoreFile(backupPath, originalPath string) error {
	source, err := os.Open(backupPath)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(originalPath)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
