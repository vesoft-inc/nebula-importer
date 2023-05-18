package utils

import "path/filepath"

// RelativePathBaseOn changes relative path base on the basePath
func RelativePathBaseOn(basePath, filePath string) string {
	if filepath.IsAbs(filePath) {
		return filePath
	}
	return filepath.Join(basePath, filePath)
}
