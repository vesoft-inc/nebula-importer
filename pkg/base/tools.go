package base

import (
	"os"
	"path"
	"strings"
)

func MustCreateFile(filePath string) *os.File {
	if err := os.MkdirAll(path.Dir(filePath), 0775); err != nil && !os.IsExist(err) {
		panic(err)
	}
	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	return file
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func IsValidType(t string) bool {
	switch strings.ToLower(t) {
	case "string":
	case "int":
	case "float":
	case "double":
	case "bool":
	case "timestamp":
	default:
		return false
	}
	return true
}

func HasHttpPrefix(path string) bool {
	return strings.HasPrefix(path, "https://") || strings.HasPrefix(path, "http://")
}
