package base

import (
	"os"
	"path"
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
