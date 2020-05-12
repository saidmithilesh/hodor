package utils

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// ReadContentFromFile ...
func ReadContentFromFile(fp string) ([]byte, error) {
	return ioutil.ReadFile(fp)
}

// IsValidPath checks if a file or directory path is valid.
func IsValidPath(fp string) bool {
	_, err := os.Stat(fp)
	if err != nil {
		return false
	}
	return true
}

// GetFullPath resolves a relative path to an absolute path.
func GetFullPath(relativePath string) string {
	absPath, err := filepath.Abs(relativePath)
	if err != nil {
		log.Println(err)
		return ""
	}
	return absPath
}

// GetPathComponents breaks down a file path into its three main components.
// 1. Directory in which the file is present
// 2. Name of the file without the extension
// 3. Extension of the file without the leading '.'
func GetPathComponents(fp string) (string, string, string) {
	directory, filename := filepath.Split(fp)
	extension := filepath.Ext(fp)

	// filename comes with extension by default, trim the extension
	filename = strings.TrimSuffix(filename, extension)

	// remove the leading '.' from the extension
	extension = strings.TrimPrefix(extension, ".")
	return directory, filename, extension
}
