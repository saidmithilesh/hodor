package helpers

import (
	"log"
	"os"
	"path/filepath"
)

// FPHelper is an exported type which provides helper methods for dealing
// with filepaths.
type FPHelper struct{}

// FilePathHelper is an instance of type FPHelper and is exported to be
// used to call helper methods on the FPHelper
var FilePathHelper FPHelper

// GetFullPath takes in a relative path string for a directory or a file
// and returns the absolute path string. If an error occurs while resolving
// the relative path to absolute path, the relative path is returned as is.
func (fph FPHelper) GetFullPath(relPath string) string {
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		log.Printf("Error while resolving absolute path from relative path :: %#v", err)
		return relPath
	}

	return absPath
}

// IsValidPath runs an os.Stat on the provided filepath to check if the
// path points to a valid directory/file or no. Returns a boolean.
func (fph FPHelper) IsValidPath(fp string) bool {
	_, err := os.Stat(fp)
	if err != nil {
		log.Printf("Error while checking path validity %#v", err)
		return false
	}

	return true
}
